package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rapidloop/skv"
)

var (
	adminCommandPermission int64 = discordgo.PermissionAdministrator
	writePermission        int64 = discordgo.PermissionSendMessages
	dmPermission           bool  = false

	commands = []*discordgo.ApplicationCommand{
		{
			Name:                     "parse",
			Description:              "Parse a csv into the database",
			DefaultMemberPermissions: &adminCommandPermission,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionAttachment,
					Name:        "csv",
					Description: "The csv file to parse",
					Required:    true,
				},
			},
		},
		{
			Name:                     "get",
			Description:              "Get a backer",
			DefaultMemberPermissions: &adminCommandPermission,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "email",
					Description: "The email of the backer",
					Required:    true,
				},
			},
		},
		{
			Name:                     "claim",
			Description:              "Claim your rewards",
			DefaultMemberPermissions: &writePermission,
			DMPermission:             &dmPermission,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "email",
					Description: "Your kickstarter email",
					Required:    true,
				},
			},
		},
		{
			Name:                     "reclaim",
			Description:              "Reclaim your rewards",
			DefaultMemberPermissions: &writePermission,
			DMPermission:             &dmPermission,
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"parse": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			log.Printf("Received interaction: %s", i.ApplicationCommandData().Name)

			attachmentID := i.ApplicationCommandData().Options[0].Value.(string)
			attachment := i.ApplicationCommandData().Resolved.Attachments[attachmentID]
			// check if attachment ends in .csv
			if !strings.HasSuffix(attachment.Filename, ".csv") {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Please provide a .csv file",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}

			res, err := http.DefaultClient.Get(attachment.URL)
			if err != nil {
				// make ephemeral
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Could not download file",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}
			defer res.Body.Close()

			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Could not read file",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}

			log.Print("Parsing csv file...")
			err = parse(string(body))
			if err != nil {
				log.Print("Failed to parse csv")
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}
			log.Print("Done")

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Done",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		},
		"get": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			log.Printf("Received interaction: %s", i.ApplicationCommandData().Name)
			backerstore, err := skv.Open("/home/bot/bots/go/backerbot/backers.db")
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}
			defer backerstore.Close()
			linkstore, err := skv.Open("/home/bot/bots/go/backerbot/backerlinks.db")
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}
			defer linkstore.Close()

			email := i.ApplicationCommandData().Options[0].Value.(string)
			var b backer
			err = backerstore.Get(email, &b)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}
			var userid string = "No link found"
			var username string = "No link found"
			err = linkstore.Get(b.Email, &userid)
			if err == nil {
				member, err := s.GuildMember(i.GuildID, userid)
				if err != nil {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: err.Error(),
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					return
				}
				username = member.User.Username
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Email: " + email + "\nBacker Tier: " + b.BackerTier + "\nUser id: " + userid + "\nUsername: " + username,
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})

		},
		"claim": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			log.Printf("Received interaction: %s by %s", i.ApplicationCommandData().Name, i.Member.User.Username)
			linkstore, err := skv.Open("/home/bot/bots/go/backerbot/backerlinks.db")
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}
			defer linkstore.Close()

			backerstore, err := skv.Open("/home/bot/bots/go/backerbot/backers.db")
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}
			defer backerstore.Close()

			email := i.ApplicationCommandData().Options[0].Value.(string)
			// check if email exists in backerstore
			err = backerstore.Get(email, nil)
			if err == skv.ErrNotFound {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Email not found",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}
			// check if email is already claimed
			err = linkstore.Get(email, nil)
			if err == nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Rewards have already been claimed for this email",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}
			// check if the user has already claimed
			var linkemail string
			err = linkstore.Get(i.Member.User.ID, &linkemail)
			if err == nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "You have already claimed your rewards with " + linkemail,
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}

			// give roles
			log.Print("Claiming default role")
			err = giveBackerTier(s, i, "Default")
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}

			var b backer
			err = backerstore.Get(email, &b)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}
			log.Println("Claiming tier role")
			err = giveBackerTier(s, i, b.BackerTier)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}

			log.Printf("Claiming backer %s to %s as %s", email, i.Member.User.ID, i.Member.User.Username)
			linkstore.Put(i.Interaction.Member.User.ID, email)
			linkstore.Put(email, i.Member.User.ID)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Successfully claimed backer roles",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		},
		"reclaim": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			log.Printf("Received interaction: %s by %s", i.ApplicationCommandData().Name, i.Member.User.Username)
			backerstore, err := skv.Open("/home/bot/bots/go/backerbot/backers.db")
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}
			defer backerstore.Close()

			linkstore, err := skv.Open("/home/bot/bots/go/backerbot/backerlinks.db")
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}
			defer linkstore.Close()

			// check if you're already linked
			var email string
			err = linkstore.Get(i.Interaction.Member.User.ID, &email)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}

			// give roles
			log.Print("Claiming default role")
			err = giveBackerTier(s, i, "Default")
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}

			var b backer
			err = backerstore.Get(email, &b)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}
			log.Println("Claiming tier role")
			err = giveBackerTier(s, i, b.BackerTier)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}
			log.Printf("Reclaimed rewards for %s", i.Member.User.Username)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Successfully reclaimed rewards",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})

		},
	}
)
