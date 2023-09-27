package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/rapidloop/skv"
)

type role struct {
	RoleName string
	RoleId   string
	Donation string
}

// function to give backer tier to a discord user in a server
func giveBackerRoles(s *discordgo.Session, i *discordgo.InteractionCreate, donation string) error {
	guildid := i.GuildID
	userid := i.Member.User.ID
	// get guild
	guild, err := s.Guild(guildid)
	if err != nil {
		return err
	}
	rolesList, err := getRoles(guild)
	if err != nil {
		return err
	}
	for i, role := range rolesList {
		if donation >= role.Donation {
			err := s.GuildMemberRoleAdd(guildid, userid, rolesList[i].RoleId)
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
