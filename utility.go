package main

import (
	"github.com/bwmarrin/discordgo"
)

type roles struct {
	roleName string
	roleId   string
	tier     string
}

var rolesList = []roles{
	{"Early Supporter", "1156325434351439912", "Default"},
	{"Sub T3", "1156325466236526642", "Tier 1"},
}

// function to give backer tier to a discord user in a server
func giveBackerTier(s *discordgo.Session, i *discordgo.InteractionCreate, tier string) error {
	guildid := i.GuildID
	userid := i.Member.User.ID
	for i, role := range rolesList {
		if role.tier == tier {
			err := s.GuildMemberRoleAdd(guildid, userid, rolesList[i].roleId)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
