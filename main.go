package main

import (
	"github.com/AtomJon/Ordis-Discord-Bot/userdata"
	"io/ioutil"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	//_TokenFile : Filename of file containing my private discord bot token
	_TokenFile = "token.txt"

	//_DataFile : Filename of data file
	_DataFile = "users.dat"

	_AuthorizeUserMessage = "Welcome to Department of Debauchery, please react :white_check_mark: to this message to be allow acces to the server."

	_AuthorizedRoleID = "<@&651861255438467083>"
)

var _UsersBeingAuthorized []userdata.AuthorizationData

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

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

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

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	data := userdata.LoadUserData(_DataFile)
	user, _ := data[m.Author.ID]

	// Ignore messages by bots
	if m.Author.Bot {
		fmt.Print("Message sent by bot")
		for _, user := range m.Mentions {		
			fmt.Print("Mentioned: ", user.Username)			
			member, err := s.GuildMember(m.GuildID, user.ID)
			if err != nil {
				fmt.Println("Error getting member info: ", err)
				return
			}
			
			userJoinedAt, err := member.JoinedAt.Parse()
			if err != nil {
				fmt.Println("Error parsing time: ", err)
			}
		
			stillNewDuration, err := time.ParseDuration("2h")
			if err != nil {
				fmt.Println("Error parsing timeout duration: ", err)
			}
		
			fmt.Println("He joined: ", time.Now().Sub(userJoinedAt))
			if time.Now().Sub(userJoinedAt) < stillNewDuration {
				for _, role := range member.Roles {					
					if role == _AuthorizedRoleID {
						return
					}
				}
					
				fmt.Println("User joined")
				authorizeUser(s, m)
			}
		}

		return
	}

	for _, v := range _UsersBeingAuthorized {
		users, err := s.MessageReactions(v.PrivateChannelID, v.AuthorizationMessageID, ":white_check_mark:", 1, "", "")
		if err != nil {
			fmt.Println("Error checking dm reactions: ", err)
		}

		if len(users) > 0 && users[0].ID == v.UserID {
			fmt.Printf("%s succesfully authorized", users[0].Username)

			err = s.GuildMemberRoleAdd(v.GuildID, users[0].ID, _AuthorizedRoleID)
			if err != nil {
				fmt.Println("Error giving user authorized role: ", err)
			}

			return
		}
	}
	
	// increment messages by one
	user.MessagesSent++

	data[m.Author.ID] = user

	fmt.Printf("%s has written %d messages\n", m.Author.Username, user.MessagesSent)

	userdata.SaveUserData(_DataFile, &data)
}

func authorizeUser(session *discordgo.Session, m *discordgo.MessageCreate) {
	
	channel, err := session.UserChannelCreate(m.Author.ID)
	if err != nil {
		fmt.Println("Error creating private channel: ", err)
	}

	msg, err := session.ChannelMessageSend(channel.ID, _AuthorizeUserMessage)
	if err != nil {
		fmt.Println("Error sending dm: ", err)
	}

	err = session.MessageReactionAdd(channel.ID, msg.ID, ":white_check_mark:")
	if err != nil {
		fmt.Println("Error adding reaction to auth dm: ", err)
	}

	authData := userdata.AuthorizationData {
		PrivateChannelID: channel.ID,
		AuthorizationMessageID: msg.ID,
		UserID: m.Member.User.ID,
		GuildID: m.GuildID,
	}

	_UsersBeingAuthorized = append(_UsersBeingAuthorized, authData)
	fmt.Printf("%s is being authorized\n", m.Member.Nick)
}