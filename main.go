package main

//DelayedUserReminder

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AtomJon/Ordis-Discord-Bot/userdata"

	"github.com/bwmarrin/discordgo"
)

const (
	//_TokenFile : Filename of file containing my private discord bot token
	_TokenFile = "token.txt"

	//_DataFile : Filename of data file
	_DataFile = "users.dat"

	_AuthorizeUserMessage = "Welcome to Department of Debauchery, please react :white_check_mark: to this message to be allow acces to the server."
	_AuthorizeCheckEmoji = "✅"

	_AuthorizedRoleID = "651861255438467083"
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

	// Register the handlers
	dg.AddHandler(messageCreate)
	dg.AddHandler(guildUpdate)
	dg.AddHandler(reactionAdded)

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

func reactionAdded(s *discordgo.Session, m *discordgo.MessageReactionAdd ) {
	if m.UserID == s.State.User.ID {
		return
	}

	if m.Emoji.Name != _AuthorizeCheckEmoji {
		return
	}

	for _, v := range _UsersBeingAuthorized {
		if m.MessageID == v.AuthorizationMessageID {
			fmt.Printf("Authorizing user: %s, guild: %s\n", v.UserID, v.GuildID)

			err := s.GuildMemberRoleAdd(v.GuildID, v.UserID, _AuthorizedRoleID) // TODO: FIX
			if err != nil {
				fmt.Println("Error giving user authorized role: ", err)
			} else {
				fmt.Println("User has been succesfully authorized")
			}
			
			return
		}
	}
}

func guildUpdate(s *discordgo.Session, m *discordgo.PresenceUpdate) {

	if m.User.ID == s.State.User.ID {
		return
	}

	if m.GuildID == "" || m.User == nil {
		return
	}

	// If user joins, the status will be online, so skip the rest
	if m.Status != discordgo.StatusOnline {
		return
	}

	member, err := s.GuildMember(m.GuildID, m.User.ID)
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
	
	if time.Now().Sub(userJoinedAt) < stillNewDuration {
		for _, role := range member.Roles {					
			if role == _AuthorizedRoleID {
				return
			}
		}

		authorizeUser(s, m.User.ID, m.GuildID)
	}
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Author.Bot {
		return
	}

	data := userdata.LoadUserData(_DataFile)
	user, _ := data[m.Author.ID]
	
	// increment messages by one
	user.MessagesSent++

	data[m.Author.ID] = user

	fmt.Printf("%s has written %d messages\n", m.Author.Username, user.MessagesSent)

	userdata.SaveUserData(_DataFile, &data)
}

func authorizeUser(session *discordgo.Session, userID string, guildID string) {

	channel, err := session.UserChannelCreate(userID)
	if err != nil {
		fmt.Println("Error creating private channel: ", err)
		return
	}

	if channel == nil {
		fmt.Println("Error private channel is nil")
		return
	}

	msg, err := session.ChannelMessageSend(channel.ID, _AuthorizeUserMessage)
	if err != nil {
		fmt.Println("Error sending dm: ", err)
		return
	}

	err = session.MessageReactionAdd(channel.ID, msg.ID, _AuthorizeCheckEmoji)
	if err != nil {
		fmt.Println("Error adding reaction to auth dm: ", err)
		return
	}

	authData := userdata.AuthorizationData {
		PrivateChannelID: channel.ID,
		AuthorizationMessageID: msg.ID,
		UserID: userID,
		GuildID: guildID,
	}

	_UsersBeingAuthorized = append(_UsersBeingAuthorized, authData)
	fmt.Printf("%s is being authorized\n", userID)
}