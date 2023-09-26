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
				response := "Please provide a .csv file"
				respond(s, i, response)
				return
			}

			res, err := http.DefaultClient.Get(attachment.URL)
			if err != nil {
				response := "Could not download file"
				respond(s, i, response)
				return
			}
			defer res.Body.Close()

			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				response := "Could not read file"
				respond(s, i, response)
				return
			}

			err = parse(string(body))
			if err != nil {
				response := err.Error()
				log.Fatal(response)
				respond(s, i, response)
				return
			}
			log.Print("Done")

			response := "Done"
			respond(s, i, response)
		},
		"get": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			log.Printf("Received interaction: %s", i.ApplicationCommandData().Name)
			backerstore, err := skv.Open("backers.db")
			if err != nil {
				response := err.Error()
				log.Fatal(response)
				respond(s, i, response)
				return
			}
			defer backerstore.Close()
			linkstore, err := skv.Open("backerlinks.db")
			if err != nil {
				response := err.Error()
				log.Fatal(response)
				respond(s, i, response)
				return
			}
			defer linkstore.Close()

			email := i.ApplicationCommandData().Options[0].Value.(string)
			var b backer
			err = backerstore.Get(email, &b)
			if err != nil {
				response := err.Error()
				log.Fatal(response)
				respond(s, i, response)
				return
			}
			var userid string = "No link found"
			var username string = "No link found"
			err = linkstore.Get(b.Email, &userid)
			if err == nil {
				member, err := s.GuildMember(i.GuildID, userid)
				if err != nil {
					response := err.Error()
					log.Fatal(response)
					respond(s, i, response)
					return
				}
				username = member.User.Username
			}

			response := "**Email**: " + email + "\n**Backer Tier**: " + b.BackerTier + "\n**User id**: " + userid + "\n**Username**: " + username
			respond(s, i, response)

		},
		"claim": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			log.Printf("Received interaction: %s by %s", i.ApplicationCommandData().Name, i.Member.User.Username)
			linkstore, err := skv.Open("backerlinks.db")
			if err != nil {
				response := err.Error()
				log.Fatal(response)
				respond(s, i, response)
				return
			}
			defer linkstore.Close()

			backerstore, err := skv.Open("backers.db")
			if err != nil {
				response := err.Error()
				log.Fatal(response)
				respond(s, i, response)
				return
			}
			defer backerstore.Close()

			email := i.ApplicationCommandData().Options[0].Value.(string)
			// check if email exists in backerstore
			err = backerstore.Get(email, nil)
			if err == skv.ErrNotFound {
				response := "Email does not exist"
				respond(s, i, response)
				return
			}
			if err != nil {
				response := err.Error()
				log.Fatal(response)
				respond(s, i, response)
				return
			}
			// check if email is already claimed
			err = linkstore.Get(email, nil)
			if err == nil {
				response := "Rewards have already been claimed for this email"
				respond(s, i, response)
				return
			}
			// check if the user has already claimed
			var linkemail string
			err = linkstore.Get(i.Member.User.ID, &linkemail)
			if err == nil {
				response := "You have already claimed your rewards with " + linkemail
				respond(s, i, response)
				return
			}

			// give roles
			log.Print("Claiming default role")
			err = giveBackerTier(s, i, "Default")
			if err != nil {
				response := err.Error()
				log.Fatal(response)
				respond(s, i, response)
				return
			}

			var b backer
			err = backerstore.Get(email, &b)
			if err != nil {
				response := err.Error()
				log.Fatal(response)
				respond(s, i, response)
				return
			}
			log.Println("Claiming tier role")
			err = giveBackerTier(s, i, b.BackerTier)
			if err != nil {
				response := err.Error()
				log.Fatal(response)
				respond(s, i, response)
				return
			}

			log.Printf("Claiming backer %s to %s as %s", email, i.Member.User.ID, i.Member.User.Username)
			linkstore.Put(i.Interaction.Member.User.ID, email)
			linkstore.Put(email, i.Member.User.ID)
			response := "Successfully claimed backer roles"
			respond(s, i, response)
		},
		"reclaim": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			log.Printf("Received interaction: %s by %s", i.ApplicationCommandData().Name, i.Member.User.Username)
			backerstore, err := skv.Open("backers.db")
			if err != nil {
				response := err.Error()
				log.Fatal(response)
				respond(s, i, response)
				return
			}
			defer backerstore.Close()

			linkstore, err := skv.Open("backerlinks.db")
			if err != nil {
				response := err.Error()
				log.Fatal(response)
				respond(s, i, response)
				return
			}
			defer linkstore.Close()

			// check if you're already linked
			var email string
			err = linkstore.Get(i.Interaction.Member.User.ID, &email)
			if err != nil {
				response := err.Error()
				log.Fatal(response)
				respond(s, i, response)
				return
			}

			// give roles
			log.Print("Claiming default role")
			err = giveBackerTier(s, i, "Default")
			if err != nil {
				response := err.Error()
				log.Fatal(response)
				respond(s, i, response)
				return
			}

			var b backer
			err = backerstore.Get(email, &b)
			if err != nil {
				response := err.Error()
				log.Fatal(response)
				respond(s, i, response)
				return
			}
			log.Println("Claiming tier role")
			err = giveBackerTier(s, i, b.BackerTier)
			if err != nil {
				response := err.Error()
				log.Fatal(response)
				respond(s, i, response)
				return
			}
			log.Printf("Reclaimed rewards for %s", i.Member.User.Username)
			response := "Successfully reclaimed rewards"
			respond(s, i, response)
		},
	}
)
