package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rapidloop/skv"
)

type role struct {
	RoleName string
	RoleId   string
	Tier     string
}

//var rolesList = []role{
//	{"Early Supporter", "1156325434351439912", "Default"},
//	{"Sub T3", "1156325466236526642", "Tier 1"},
//}

// function to give backer tier to a discord user in a server
func giveBackerTier(s *discordgo.Session, i *discordgo.InteractionCreate, tier string) error {
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
		if role.Tier == tier {
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
