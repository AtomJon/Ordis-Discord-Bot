package userdata

import (
	"encoding/gob"
	"fmt"
	"os"
)

// UserData An type for storing user data
type UserData struct {
	MessagesSent 	int
	PreferedReferal string
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
func SaveUserData(dataFile string, data *(map[string]UserData)) {
	outputFile, err := os.Create(dataFile)
	if err != nil {
		fmt.Println(err)
	}

	gob.NewEncoder(outputFile).Encode(data)
}