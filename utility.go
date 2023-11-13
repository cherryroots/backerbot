package main

import (
	"backerbot/skv"
	"log"

	"github.com/bwmarrin/discordgo"
)

type role struct {
	RoleID   string
	Donation float64
}

func giveBackerRoles(s *discordgo.Session, i *discordgo.InteractionCreate, donation float64) error {
	guildID := i.GuildID
	userID := i.Member.User.ID

	guild, err := s.Guild(guildID)
	if err != nil {
		return err
	}
	rolesList, err := getRoles(guild)
	if err != nil {
		return err
	}
	for _, role := range rolesList {
		if donation >= role.Donation {
			err := s.GuildMemberRoleAdd(guildID, userID, role.RoleID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getRoleName(s *discordgo.Session, guildID string, roleID string) (string, error) {
	guild, err := s.Guild(guildID)
	if err != nil {
		return "", err
	}
	for _, guildRole := range guild.Roles {
		if guildRole.ID == roleID {
			return guildRole.Name, nil
		}
	}
	return "", nil
}

func getRoles(guild *discordgo.Guild) ([]role, error) {
	rolestore, err := skv.Open("roles.db")
	if err != nil {
		return nil, err
	}
	defer rolestore.Close()

	var rolesList = []role{}
	var newrole role
	for _, guildRole := range guild.Roles {
		err := rolestore.Get(guildRole.ID, &newrole)
		if err == nil {
			rolesList = append(rolesList, newrole)
		}
	}
	return rolesList, nil
}

func logRespond(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	log.Print(content)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

func respond(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func sendModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "modal_claim-" + i.Interaction.Member.User.ID,
			Title:    "Claim rewards!",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "email",
							Label:       "Kickstarter email",
							Style:       discordgo.TextInputParagraph,
							Placeholder: "Please provide the same email used in your kickstarter account...",
							Required:    true,
							MaxLength:   100,
							MinLength:   5,
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Print(err)
	}
}

func sendClaimButton(s *discordgo.Session, i *discordgo.InteractionCreate, title string, description string, buttontext string, emojiid string) {
	embed := discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       0x00ff00,
		Title:       title,
		Description: description,
		Footer:      &discordgo.MessageEmbedFooter{Text: "Made with ❤️ by @cherrywoods", IconURL: s.State.User.AvatarURL("512x512")},
	}
	button := discordgo.Button{
		Label:    buttontext,
		Emoji:    discordgo.ComponentEmoji{ID: emojiid},
		Style:    discordgo.PrimaryButton,
		CustomID: "button_claim",
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{&embed},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						button,
					},
				},
			},
		},
	})
	if err != nil {
		log.Print(err)
	}
}
