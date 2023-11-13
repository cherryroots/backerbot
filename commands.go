package main

import (
	"backerbot/skv"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
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
			DMPermission:             &dmPermission,
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
			Name:                     "addrole",
			Description:              "Maps a role to a tier",
			DefaultMemberPermissions: &adminCommandPermission,
			DMPermission:             &dmPermission,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionRole,
					Name:        "role",
					Description: "The guild role to map",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "donation",
					Description: "At least this much to get the role",
					Required:    true,
				},
			},
		},
		{
			Name:                     "removerole",
			Description:              "Removes a role from the database",
			DefaultMemberPermissions: &adminCommandPermission,
			DMPermission:             &dmPermission,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "id",
					Description: "The id of the role",
					Required:    true,
				},
			},
		},
		{
			Name:                     "listroles",
			Description:              "Lists all roles",
			DefaultMemberPermissions: &adminCommandPermission,
			DMPermission:             &dmPermission,
		},
		{
			Name:                     "get-email",
			Description:              "Get a backer",
			DefaultMemberPermissions: &adminCommandPermission,
			DMPermission:             &dmPermission,
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
			Name:                     "get-user",
			Description:              "Get a backer",
			DefaultMemberPermissions: &adminCommandPermission,
			DMPermission:             &dmPermission,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "The user to look up",
					Required:    true,
				},
			},
		},
		{
			Name:                     "get-backer-info",
			Type:                     discordgo.UserApplicationCommand,
			DefaultMemberPermissions: &adminCommandPermission,
			DMPermission:             &dmPermission,
		},
		{
			Name:                     "unlink",
			Description:              "Unlink a backer from a role and remove roles",
			DefaultMemberPermissions: &adminCommandPermission,
			DMPermission:             &dmPermission,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "email",
					Description: "The email of the backer",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "userid",
					Description: "the user to remove",
					Required:    true,
				},
			},
		},
		{
			Name:                     "button",
			Description:              "Create a button to claim rewards",
			DefaultMemberPermissions: &adminCommandPermission,
			DMPermission:             &dmPermission,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "title",
					Description: "Title of the post",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "description",
					Description: "description of the post",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "buttontext",
					Description: "Text of the button",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "emojiid",
					Description: "Emoji id of the button, get it by adding a backslash before the emoji in discord",
					Required:    false,
				},
			},
		},
		{
			Name:                     "claim",
			Description:              "Claim your rewards",
			DefaultMemberPermissions: &writePermission,
			DMPermission:             &dmPermission,
		},
		{
			Name:                     "reclaim",
			Description:              "Reclaim your rewards",
			DefaultMemberPermissions: &writePermission,
			DMPermission:             &dmPermission,
		},
		{
			Name:                     "fix-emails",
			Description:              "Fix email",
			DefaultMemberPermissions: &adminCommandPermission,
			DMPermission:             &dmPermission,
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"parse": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			log.Printf("Received interaction: %s by %s", i.ApplicationCommandData().Name, i.Interaction.Member.User.Username)

			attachmentID := i.ApplicationCommandData().Options[0].Value.(string)
			attachment := i.ApplicationCommandData().Resolved.Attachments[attachmentID]
			// check if attachment ends in .csv
			if !strings.HasSuffix(attachment.Filename, ".csv") {
				response := "Please provide a .csv file"
				respond(s, i, response)
				return
			}

			// download file
			res, err := http.DefaultClient.Get(attachment.URL)
			if err != nil {
				response := "Could not download file"
				respond(s, i, response)
				return
			}
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				response := "Could not read file"
				respond(s, i, response)
				return
			}

			// parse csv and put it into the backers db
			err = parse(string(body))
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}
			log.Print("Done")

			response := "Done"
			respond(s, i, response)
		},
		"addrole": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			log.Printf("Received interaction: %s by %s", i.ApplicationCommandData().Name, i.Interaction.Member.User.Username)
			rolestore, err := skv.Open("roles.db")
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}
			defer rolestore.Close()

			discordRole := i.ApplicationCommandData().Options[0].Value.(string)
			roleDonation := i.ApplicationCommandData().Options[1].Value.(float64)

			var newrole role = role{
				RoleID:   discordRole,
				Donation: roleDonation,
			}

			// Put a new role into the rolestore
			err = rolestore.Put(newrole.RoleID, newrole)
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}

			roleName, _ := getRoleName(s, i.GuildID, discordRole)
			response := "Added role " + roleName
			respond(s, i, response)

		},
		"removerole": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			log.Printf("Received interaction: %s by %s", i.ApplicationCommandData().Name, i.Interaction.Member.User.Username)
			rolestore, err := skv.Open("roles.db")
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}
			defer rolestore.Close()

			roleID := i.ApplicationCommandData().Options[0].Value.(string)
			err = rolestore.Delete(roleID)
			if err != nil {
				response := "There's no role with that id"
				logRespond(s, i, response)
				return
			}
			log.Print("Done")
			response := "Done"
			respond(s, i, response)

		},
		"listroles": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			log.Printf("Received interaction: %s by %s", i.ApplicationCommandData().Name, i.Interaction.Member.User.Username)

			guild, err := s.Guild(i.GuildID)
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}

			roles, err := getRoles(guild)
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}
			var response string
			for _, role := range roles {
				roleName, _ := getRoleName(s, i.GuildID, role.RoleID)
				response += "**Role Name**: " + roleName + "\n**Role ID**: " + role.RoleID + "\n**Donation**: " + fmt.Sprintf("%.2f", role.Donation) + "\n\n"
			}
			if response == "" {
				response = "No roles have been added"
			}
			respond(s, i, response)

		},
		"get-email": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			log.Printf("Received interaction: %s by %s", i.ApplicationCommandData().Name, i.Interaction.Member.User.Username)
			backerstore, err := skv.Open("backers.db")
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}
			defer backerstore.Close()
			linkstore, err := skv.Open("backerlinks.db")
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}
			defer linkstore.Close()

			email := i.ApplicationCommandData().Options[0].Value.(string)
			var b backer
			err = backerstore.Get(email, &b)
			if err != nil {
				response := "No backer found with that email"
				logRespond(s, i, response)
				return
			}
			var userID string = "No link found"
			var username string = "No link found"
			err = linkstore.Get(b.Email, &userID)
			if err == nil {
				member, err := s.GuildMember(i.GuildID, userID)
				if err != nil {
					response := err.Error()
					logRespond(s, i, response)
					return
				}
				username = member.User.Username
			}

			response := "**Email**: " + email + "\n**Backer Reward Title**: " + b.RewardTitle + "\n**Backer Donation**: " + fmt.Sprintf("%.2f", b.Donation) + "\n**Status**: " + b.Status + "\n**User id**: " + userID + "\n**Username**: " + username
			respond(s, i, response)

		},
		"get-user": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			log.Printf("Received interaction: %s by %s", i.ApplicationCommandData().Name, i.Interaction.Member.User.Username)
			linkstore, err := skv.Open("backerlinks.db")
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}
			defer linkstore.Close()

			backerstore, err := skv.Open("backers.db")
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}
			defer backerstore.Close()

			userID := i.ApplicationCommandData().Options[0].Value.(string)

			var email string = ""
			err = linkstore.Get(userID, &email)
			if err != nil {
				response := "This user hasn't linked to an email"
				logRespond(s, i, response)
				return
			}

			var b backer
			err = backerstore.Get(email, &b)
			if err != nil {
				response := "No backer found with that email"
				logRespond(s, i, response)
				return
			}

			response := "**Email**: " + email + "\n**Backer Reward Title**: " + b.RewardTitle + "\n**Backer Donation**: " + fmt.Sprintf("%.2f", b.Donation) + "\n**Status**: " + b.Status

			respond(s, i, response)
		},
		"get-backer-info": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			log.Printf("Received interaction: %s by %s", i.ApplicationCommandData().Name, i.Interaction.Member.User.Username)
			linkstore, err := skv.Open("backerlinks.db")
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}
			defer linkstore.Close()

			backerstore, err := skv.Open("backers.db")
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}
			defer backerstore.Close()

			userID := i.ApplicationCommandData().TargetID

			var email string = ""
			err = linkstore.Get(userID, &email)
			if err != nil {
				response := "This user hasn't linked to an email"
				logRespond(s, i, response)
				return
			}

			var b backer
			err = backerstore.Get(email, &b)
			if err != nil {
				response := "No backer found with that email"
				logRespond(s, i, response)
				return
			}

			response := "**Email**: " + email + "\n**Backer Reward Title**: " + b.RewardTitle + "\n**Backer Donation**: " + fmt.Sprintf("%.2f", b.Donation) + "\n**Status**: " + b.Status

			respond(s, i, response)
		},
		"unlink": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			log.Printf("Received interaction: %s by %s", i.ApplicationCommandData().Name, i.Interaction.Member.User.Username)

			linkstore, err := skv.Open("backerlinks.db")
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}
			defer linkstore.Close()

			email := i.ApplicationCommandData().Options[0].Value.(string)
			userID := i.ApplicationCommandData().Options[1].Value.(string)

			linkstore.Delete(email)
			linkstore.Delete(userID)

			// get roles and delete them from the unlinked user
			guild, err := s.Guild(i.GuildID)
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}

			roles, err := getRoles(guild)
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}

			for _, role := range roles {
				err = s.GuildMemberRoleRemove(i.GuildID, userID, role.RoleID)
				if err != nil {
					response := err.Error()
					logRespond(s, i, response)
					return
				}
			}

			response := "Done"
			respond(s, i, response)
		},
		"button": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			log.Printf("Received interaction: %s by %s", i.ApplicationCommandData().Name, i.Interaction.Member.User.Username)
			var title, description, buttontext, emijiid string = "", "", "", ""
			for _, option := range i.ApplicationCommandData().Options {
				switch option.Name {
				case "title":
					title = option.Value.(string)
				}
				switch option.Name {
				case "description":
					description = option.Value.(string)
				}
				switch option.Name {
				case "buttontext":
					buttontext = option.Value.(string)
				}
				switch option.Name {
				case "emojiid":
					emijiid = option.Value.(string)
				}

			}

			sendClaimButton(s, i, title, description, buttontext, emijiid)
		},
		"claim": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			log.Printf("Received interaction: %s by %s", i.ApplicationCommandData().Name, i.Interaction.Member.User.Username)
			// Open modal and which will run "modal_claim" afterwards
			sendModal(s, i)
		},
		"button_claim": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			log.Printf("Received interaction: %s by %s", i.MessageComponentData().CustomID, i.Interaction.Member.User.Username)
			backerstore, err := skv.Open("backers.db")
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}
			defer backerstore.Close()

			linkstore, err := skv.Open("backerlinks.db")
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}
			defer linkstore.Close()

			// Check if the user has claimed before
			var email string
			err = linkstore.Get(i.Interaction.Member.User.ID, &email)
			if err == nil {
				// Reclaim rewards
				var b backer
				err = backerstore.Get(email, &b)
				if err != nil {
					response := err.Error()
					logRespond(s, i, response)
					return
				}
				err = giveBackerRoles(s, i, b.Donation)
				if err != nil {
					response := err.Error()
					logRespond(s, i, response)
					return
				}
				log.Printf("Reclaimed rewards for %s", i.Member.User.Username)
				response := "Successfully reclaimed rewards"
				respond(s, i, response)
				return
			}

			// Open modal and which will run "modal_claim" afterwards
			sendModal(s, i)
		},
		"modal_claim": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			log.Printf("Received interaction: %s by %s", i.ModalSubmitData().CustomID, i.Member.User.Username)
			linkstore, err := skv.Open("backerlinks.db")
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}
			defer linkstore.Close()

			backerstore, err := skv.Open("backers.db")
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}
			defer backerstore.Close()

			email := i.ModalSubmitData().Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			email = strings.ToLower(email)

			// Check if email exists and get backer
			var b backer
			err = backerstore.Get(email, &b)
			if err == skv.ErrNotFound {
				response := "Email does not exist, please check that you wrote it correctly."
				respond(s, i, response)
				return
			}
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}

			if b.Status != "collected" {
				response := "Your kickstarter pledge has not been received yet, please message Raffle.\n\nThank you!"
				respond(s, i, response)
				return
			}

			// Check if the email has already been claimed
			err = linkstore.Get(email, nil)
			if err == nil {
				response := "Rewards have already been claimed for this email. If it wasn't you please message Raffle.\n\nThank you!"
				respond(s, i, response)
				return
			}

			// Check if the user has already claimed
			var linkemail string
			err = linkstore.Get(i.Member.User.ID, &linkemail)
			if err == nil {
				response := "You have already claimed your rewards with " + linkemail
				respond(s, i, response)
				return
			}

			// claim rewards
			err = giveBackerRoles(s, i, b.Donation)
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}

			log.Printf("Claimied backer %s to %s as %s", email, i.Member.User.ID, i.Member.User.Username)
			linkstore.Put(i.Interaction.Member.User.ID, email)
			linkstore.Put(email, i.Member.User.ID)
			response := "Successfully claimed backer roles"
			respond(s, i, response)
		},
		"fix-emails": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// iterate over all backers in the linkstore and lowercase the email
			linkstore, err := skv.Open("backerlinks.db")
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}
			defer linkstore.Close()

			backerstore, err := skv.Open("backers.db")
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}
			defer backerstore.Close()

			keys, err := linkstore.GetKeys()
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}

			// email regex
			reEmail := regexp.MustCompile(`[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}`)
			// only numbers regex
			reNum := regexp.MustCompile(`[0-9]+`)

			log.Printf("Found %d backers", len(keys))

			for count, key := range keys {
				log.Printf("Processing %d of %d", count+1, len(keys))
				// check if email or number
				if reEmail.MatchString(key) {
					// if the key is an email get the user
					var userID string
					err = linkstore.Get(key, &userID)
					if err != nil {
						response := err.Error()
						logRespond(s, i, response)
						return
					}
					email := strings.ToLower(key)
					linkstore.Delete(key)
					linkstore.Put(email, userID)

				} else if reNum.MatchString(key) {
					// if the key is a number get the email linked to it
					var email string
					err = linkstore.Get(key, &email)
					if err != nil {
						response := err.Error()
						logRespond(s, i, response)
						return
					}
					lowerEmail := strings.ToLower(email)
					linkstore.Delete(key)
					linkstore.Put(key, lowerEmail)

				} else {
					continue
				}
			}

			response := "Successfully fixed emails"
			respond(s, i, response)

		},
		"reclaim": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			log.Printf("Received interaction: %s by %s", i.ApplicationCommandData().Name, i.Member.User.Username)
			backerstore, err := skv.Open("backers.db")
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}
			defer backerstore.Close()

			linkstore, err := skv.Open("backerlinks.db")
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}
			defer linkstore.Close()

			// Get email from user id
			var email string
			err = linkstore.Get(i.Interaction.Member.User.ID, &email)
			if err != nil {
				response := "You haven't claimed your rewards yet"
				logRespond(s, i, response)
				return
			}

			// Reclaim rewards
			var b backer
			err = backerstore.Get(email, &b)
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}
			log.Println("Claiming donation roles")
			err = giveBackerRoles(s, i, b.Donation)
			if err != nil {
				response := err.Error()
				logRespond(s, i, response)
				return
			}
			log.Printf("Reclaimed rewards for %s", i.Member.User.Username)
			response := "Successfully reclaimed rewards"
			respond(s, i, response)
		},
	}
)
