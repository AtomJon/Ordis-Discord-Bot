package commands

import (
	"fmt"

	"github.com/AtomJon/Ordis-Discord-Bot/constants"

	"github.com/bwmarrin/discordgo"
)

// Command Type for commands
type Command struct {
	TriggerExpression string

	Activate func(s *discordgo.Session, m *discordgo.MessageCreate) string
}

//Commands : List of commands 
var Commands = []Command{
	{"?.is|these.?.authorized.*", func(s *discordgo.Session, m *discordgo.MessageCreate) string {
		msg := ""

		for _, mentioned := range m.Mentions {
			member, err := s.GuildMember(m.GuildID, mentioned.ID)
			if err != nil {
				fmt.Println("Error while sending private message: %w", err)
				return "Critical error :_"
			}

			for _, role := range member.Roles {
				if role == constants.AuthorizedRoleID {
					msg += member.Nick + " is authorized\n";
				} else {
					msg += member.Nick + " is not authorized\n";
				}
			}
		}

		return msg + "Sir."
	}},
}