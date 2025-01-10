package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-json-experiment/json"
)

type Config struct {
	Token  string `json:"token"`
	Prompt string `json:"prompt"`
}

var (
	ErrConfigNotFound = errors.New("config.json did not exist so a new one was created, please edit it and restart the bot")
)

func NewConfig() error {
	_, err := os.Stat("config.json")

	if os.IsNotExist(err) {
		file, err := os.Create("config.json")
		if err != nil {
			return err
		}
		defer file.Close()

		config := Config{
			Token:  "DISCORD_TOKEN",
			Prompt: "You are a helpful assistant...",
		}

		err = json.MarshalWrite(file, config)
		if err != nil {
			return err
		}

		return ErrConfigNotFound
	}

	fmt.Println("Reading config file")
	file, err := os.ReadFile("config.json")
	if err != nil {
		return err
	}

	fmt.Println("Parsing config file")
	err = json.Unmarshal(file, &CONFIG)
	if err != nil {
		return err
	}

	return err
}
