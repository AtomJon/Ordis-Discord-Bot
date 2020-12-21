package userdata

import (
	"os"
	"fmt"
	"encoding/gob"
)

// UserData An type for storing user data
type UserData struct {
	MessagesSent 	int
	PreferedReferal string
}

// AuthorizationData An struct for holding reference to users private channel
type AuthorizationData struct {
	PrivateChannelID 		string
	AuthorizationMessageID 	string
	UserID			 		string
	GuildID					string
}

//LoadUserData Returns the userdata saved in dataFile
func LoadUserData(dataFile string) (map[string]UserData) {
	data := map[string]UserData{}

	inputFile, err := os.Open(dataFile)
	if err != nil {
		fmt.Println(err)
	}

	gob.NewDecoder(inputFile).Decode(&data)

	return data
}

//SaveUserData Saves the data in dataFile
func SaveUserData(dataFile string, data *interface{}) {
	outputFile, err := os.Create(dataFile)
	if err != nil {
		fmt.Println(err)
	}

	gob.NewEncoder(outputFile).Encode(data)
}