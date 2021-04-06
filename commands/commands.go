package commands

import (
	"fmt"

	"github.com/AtomJon/Ordis-Discord-Bot/constants"
	"github.com/AtomJon/Ordis-Discord-Bot/userdata"

	"github.com/bwmarrin/discordgo"
)

// Command Type for commands
type Command struct {
	TriggerExpression string

	Activate func(s *discordgo.Session, m *discordgo.MessageCreate) string
}

//Commands : List of commands 
var Commands = []Command{
	{"(is|these|they).*authorized", func(s *discordgo.Session, m *discordgo.MessageCreate) string {
		msg := ""

		for _, mentioned := range m.Mentions {			
			if mentioned.ID != s.State.User.ID {
				memberIsAuthorized := false

				member, err := s.GuildMember(m.GuildID, mentioned.ID)
				if err != nil {
					fmt.Println("Error while obtaining member info: ", err)
					return "Critical error :_"
				}

				for _, role := range member.Roles {
					if role == constants.AuthorizedRoleID {
						msg += member.Mention() + " is authorized\n"						
						memberIsAuthorized = true
					}
				}

				if !memberIsAuthorized {
					msg += member.Mention() + " is not authorized\n"
				}
			}
		}

		return msg + "Sir."
	}},
	{"(msg|messages|written)", func(s *discordgo.Session, m *discordgo.MessageCreate) string {
		msg := ""

		for _, mentioned := range m.Mentions {	
		
			member, err := s.GuildMember(m.GuildID, mentioned.ID)
			if err != nil {
				fmt.Println("Error while obtaining member info: ", err)
				return "Critical error :_"
			}

			data := userdata.LoadUserData(constants.DataFile)
			user := data[mentioned.ID]
					
			if mentioned.ID != s.State.User.ID {
				msg += fmt.Sprintf("%s has written %d messages\n", member.Mention(), user.MessagesSent)
			}
		}

		return msg + "Sir."
	}},
}