package main

import (
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func joel(event *events.ApplicationCommandInteractionCreate, joel string) {
	joel = strings.TrimSpace(joel)
	path, ok := ASSETS_PATHS[joel]
	if !ok || joel == "" {
		n := rand.IntN(len(ASSETS) - 1)
		joel = ASSETS[n]
		path = ASSETS_PATHS[joel]
	}

	image, err := os.Open(path)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer image.Close()

	message := discord.NewMessageCreateBuilder().SetContent(joel).AddFile(image.Name(), image.Name(), image).Build()
	sendMessage(event, message)
}

func ttj(event *events.ApplicationCommandInteractionCreate) {
	now := time.Now()
	err := event.DeferCreateMessage(false)
	if err != nil {
		fmt.Printf("Error deferring message: %v\n", err)
		reply(event, "An error occurred while deferring the message.")
		return
	}
	since := time.Until(now)

	ttj := fmt.Sprintf("TTJ took %dms", since.Abs().Milliseconds())
	fmt.Println(ttj)

	image, err := os.Open("./assets/Joel.webp")
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer image.Close()

	message := discord.NewMessageCreateBuilder().SetContent(ttj).AddFile(image.Name(), image.Name(), image).Build()
	_, err = CLIENT.Rest.CreateFollowupMessage(event.ApplicationID(), event.Token(), message)
	if err != nil {
		fmt.Printf("Error replying: %v\n", err)
	}
}
