package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/rapidloop/skv"
)

type role struct {
	RoleName string
	RoleID   string
	Donation string
}

func giveBackerRoles(s *discordgo.Session, i *discordgo.InteractionCreate, donation string) error {
	guildid := i.GuildID
	userid := i.Member.User.ID

	guild, err := s.Guild(guildid)
	if err != nil {
		return err
	}
	rolesList, err := getRoles(guild)
	if err != nil {
		return err
	}
	for _, role := range rolesList {
		if donation >= role.Donation {
			err := s.GuildMemberRoleAdd(guildid, userid, role.RoleID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getRoles(guild *discordgo.Guild) ([]role, error) {
	rolestore, err := skv.Open("roles.db")
	if err != nil {
		return nil, err
	}
	defer rolestore.Close()

	var rolesList = []role{}
	var newrole role
	for _, grole := range guild.Roles {
		err := rolestore.Get(grole.ID, &newrole)
		if err == nil {
			rolesList = append(rolesList, newrole)
		}
	}
	return rolesList, nil
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
		CustomID: "claim_button",
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
