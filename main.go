package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/AtomJon/Ordis-Discord-Bot/commands"
	"github.com/AtomJon/Ordis-Discord-Bot/constants"
	"github.com/AtomJon/Ordis-Discord-Bot/userdata"

	"github.com/bwmarrin/discordgo"
)

const (
	//_TokenFile : Filename of file containing my private discord bot token
	_TokenFile = "token.txt"

	_RemindUserMessage = "Could you please go to the door-sign channel and pick the member role, so that you can access the rest of the server, sir?"

	_RemindDelay = time.Second * 30
)

func main() {

	discordToken, err := ioutil.ReadFile(_TokenFile)
	if err != nil {
		fmt.Println("error reading discord token,", err)
		return
	}

	fmt.Println("Discord token: " + string(discordToken))

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + string(discordToken))
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the handlers
	dg.AddHandler(messageCreate)
	dg.AddHandler(guildUpdate)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)//discordgo.IntentsGuildMessages | discordgo.IntentsGuildMembers | discordgo.IntentsDirectMessages | discordgo.IntentsDirectMessageReactions)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func guildUpdate(s *discordgo.Session, m *discordgo.PresenceUpdate) {

	if m.User.ID == s.State.User.ID {
		return
	}

	if m.GuildID == "" || m.User == nil {
		return
	}

	member, err := s.GuildMember(m.GuildID, m.User.ID)
	if err != nil {
		fmt.Println("Error while getting member info: %w", err)
		return
	}

	if userIsAuthorized(member) {
		return
	}

	time.AfterFunc(_RemindDelay, func() {
		if (!userIsAuthorized(member)) {
			channel, err := s.UserChannelCreate(m.User.ID)
			if err != nil {
				fmt.Println("Error while creating user channel: %w", err)
				return
			}

			_, err = s.ChannelMessageSend(channel.ID, _RemindUserMessage)
			if err != nil {
				fmt.Println("Error while sending private message: %w", err)
				return
			}
		}
	});
}

func userIsAuthorized(user *discordgo.Member) bool {
	for _, role := range user.Roles {
		if role == constants.AuthorizedRoleID {
			return true
		}
	}

	return false
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Author.Bot {
		return
	}

	data := userdata.LoadUserData(constants.DataFile)
	user, _ := data[m.Author.ID]
	
	// increment messages by one
	user.MessagesSent++

	data[m.Author.ID] = user

	fmt.Printf("%s has written %d messages\n", m.Author.Username, user.MessagesSent)

	for _, mention := range m.Mentions {
		if mention.ID != s.State.User.ID {
			for _, command := range commands.Commands {
				match, err := regexp.Match(command.TriggerExpression, []byte(m.Content))
				if err != nil {
					fmt.Println("Error parsing regexp: ", err)
				}
		
				if (match) {
					msg := command.Activate(s, m)
					_, err = s.ChannelMessageSend(m.ChannelID, msg)
					if err != nil {
						fmt.Println("Error while sending response message: ", err)
					}
				}
			}
		}
	}	

	userdata.SaveUserData(constants.DataFile, &data)
}