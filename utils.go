package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
	"github.com/dslipak/pdf"
)

const (
	FISH_EMOJI = "🐟"
	OK_EMOJI   = "👍"
	ERR_EMOJI  = "❌"
)

type UserInfo struct {
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Skipped  bool   `json:"skipped"`
}

func isMentioned(mentions []discord.User) bool {
	if mentions == nil {
		return false
	}

	for _, mention := range mentions {
		if CLIENT.Bot.ID() == mention.ID {
			return true
		}
	}

	return false
}

func addReaction(channelID snowflake.ID, messageID snowflake.ID, emoji string) {
	err := CLIENT.Rest.AddReaction(channelID, messageID, emoji)
	if err != nil {
		fmt.Printf("Error reacting: %v\n", err)
	}
}

func reply(event *events.ApplicationCommandInteractionCreate, content string) {
	message := discord.NewMessageCreateBuilder().SetContent(content).Build()

	err := event.CreateMessage(message)
	if err != nil {
		fmt.Printf("Error replying: %v\n", err)
	}
}

func sendMessage(event *events.ApplicationCommandInteractionCreate, message discord.MessageCreate) {
	err := event.CreateMessage(message)
	if err != nil {
		fmt.Printf("Error replying: %v\n", err)
	}
}

func getBotVoiceState(event *events.ApplicationCommandInteractionCreate) *discord.VoiceState {
	guildID := *event.GuildID()
	voiceState, err := CLIENT.Rest.GetCurrentUserVoiceState(guildID)
	if err != nil {
		fmt.Printf("Error getting bot voice state: %v\n", err)
		reply(event, err.Error())
		return nil
	}

	return voiceState
}

func updateVoiceChannel(event *events.ApplicationCommandInteractionCreate, channelID *snowflake.ID) {
	guildID := *event.GuildID()
	err := CLIENT.Bot.UpdateVoiceState(context.TODO(), guildID, channelID, false, true)
	if err != nil {
		fmt.Printf("Error updating voice channel: %v\n", err)
		reply(event, err.Error())
	}
}

func readPDF(buffer bytes.Buffer) (string, error) {
	reader := bytes.NewReader(buffer.Bytes())
	pdfReader, err := pdf.NewReader(reader, int64(buffer.Len()))
	if err != nil {
		return "", err
	}

	numPages := pdfReader.NumPage()
	pages := make([]string, numPages)

	for i := 0; i < numPages; i++ {
		page := pdfReader.Page(i + 1)

		text, err := page.GetPlainText(nil)
		if err != nil {
			fmt.Printf("Error extracting text from page %d: %v\n", i, err)
			continue
		}

		pages[i] = text
	}

	return strings.Join(pages, "\n\n"), nil
}

func handleUserQuery(user UserInfo, query string) (lavalink.Track, error) {
	var play lavalink.Track
	var err error

	CLIENT.Lavalink.BestNode().LoadTracksHandler(context.TODO(), query, disgolink.NewResultHandler(
		func(track lavalink.Track) {
			fmt.Printf("Found track: %s\n", track.Info.Title)
			play = track
		},
		func(playlist lavalink.Playlist) {
			fmt.Printf("Found playlist: %s\n", playlist.Info.Name)
			play = playlist.Tracks[0]
		},
		func(tracks []lavalink.Track) {
			length := len(tracks)
			if length == 0 {
				err = ErrNoTracksFound
				return
			}

			play = tracks[0]
			fmt.Printf("Found %d tracks, choosing first one: %s\n", length, play.Info.Title)
		},
		func() {
			fmt.Println("No tracks found")
			err = ErrNoTracksFound
		},
		func(lavalinkErr error) {
			fmt.Printf("Error loading tracks: %v\n", lavalinkErr)
			err = ErrLoadingTracks
		},
	))

	if err == nil {
		play, err = play.WithUserData(user)
		if err != nil {
			fmt.Printf("Error adding user data: %v\n", err)
			return play, err
		}
	}

	return play, err
}
