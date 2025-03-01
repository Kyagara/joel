package main

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

var (
	READY = make(chan bool, 1)
)

func onReady(event *events.Ready) {
	fmt.Println("Bot is connected")
	READY <- true
}

func commandListener(event *events.ApplicationCommandInteractionCreate) {
	data := event.SlashCommandInteractionData()
	command := data.CommandName()

	fmt.Printf("Command %s by %s\n", command, event.User().Username)

	switch command {
	case "help":
		help := "**LLM**:\n`reset`: Resets the users chat history with the bot\n**Random**:\n`joel`: Posts a random or specific joel if a parameter is provided\n`ttj`: Posts Time to Joel (latency test)\n**Music**:\n`play`: Plays a song, accepts an URL or a search query\n`stop`: Stops the current song\n`pause`: Pauses the current song\n`resume`: Resumes the current song\n`skip`: Skips the current song\n`join`: Joins the voice channel\n`leave`: Leaves the voice channel\n`queue`: Shows the queue\n`playing`: Shows the current song"
		reply(event, help)

	case "joel":
		opt, _ := data.OptString("joel")

		joel(event, opt)
	case "ttj":
		ttj(event)
	case "play":
		query, ok := data.OptString("query")
		if !ok {
			reply(event, "Please provide a query.")
			return
		}

		play(event, query)
	case "stop":
		stop(event)
	case "pause":
		pause(event)
	case "resume":
		resume(event)
	case "skip":
		skip(event)
	case "join":
		join(event)
	case "leave":
		leave(event)
	case "queue":
		queue(event)
	case "playing":
		playing(event)

	case "reset":
		reset(event)

	default:
		message := discord.NewMessageCreateBuilder().SetContent("Unknown command, please use `/help` for a list of commands.").SetEphemeral(true).Build()
		err := event.CreateMessage(message)
		if err != nil {
			fmt.Printf("Error responding to command: %v\n", err)
		}
	}
}

func onMessageCreate(event *events.MessageCreate) {
	if event.Message.Author.Bot {
		return
	}

	// LLM
	handleUserMessage(event)
}

func onVoiceStateUpdate(event *events.GuildVoiceStateUpdate) {
	CLIENT.Lavalink.OnVoiceStateUpdate(context.TODO(), event.VoiceState.GuildID, event.VoiceState.ChannelID, event.VoiceState.SessionID)
}

func onVoiceServerUpdate(event *events.VoiceServerUpdate) {
	CLIENT.Lavalink.OnVoiceServerUpdate(context.TODO(), event.GuildID, event.Token, *event.Endpoint)
}

func onTrackEnd(player disgolink.Player, event lavalink.TrackEndEvent) {
	fmt.Printf("Track ended: %s | %s\n", event.Track.Info.Title, *event.Track.Info.URI)

	if TRACKS.Empty() {
		fmt.Println("No more tracks to play")
		return
	}

	user := UserInfo{}
	err := event.Track.UserData.Unmarshal(&user)
	if err != nil {
		fmt.Printf("Error scanning user: %v\n", err)
		return
	}

	TRACKS.Pop()

	if TRACKS.Empty() {
		fmt.Println("No more tracks to play")
		return
	}

	if user.Skipped {
		return
	}

	err = player.Update(context.TODO(), lavalink.WithTrack(TRACKS.First()))
	if err != nil {
		fmt.Printf("Error playing next track: %v\n", err)
	}
}
