package main

import (
	"Ordis-Discord-Bot/userdata"
	"io/ioutil"
	"encoding/gob"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

const (

	//TokenFile : Filename of file containing my private discord bot token
	TokenFile = "token.txt"

	//DataFile : Filename of data file
	DataFile = "users.dat"
)

func main() {

	discordToken, err := ioutil.ReadFile(TokenFile)
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
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

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

	// Ignore messages by bots
	if m.Author.Bot {
		return
	}

	data := map[string]userdata.UserData{}

	inputFile, err := os.Open(DataFile)
	if err != nil {
		fmt.Println(err)
	}

	gob.NewDecoder(inputFile).Decode(&data)

	user, exists := data[m.Author.ID]
	if !exists {
		// Is the user already in database. Else, set messages to one
		user.MessagesSent = 1
	} else {
		// If the user is in database, increment messages by one
		user.MessagesSent++
	}

	data[m.Author.ID] = user

	fmt.Printf("%s has written %d messages\n", m.Author.Username, user.MessagesSent)

	outputFile, err := os.Create(DataFile)
	if err != nil {
		fmt.Println(err)
	}

	gob.NewEncoder(outputFile).Encode(&data)
}
