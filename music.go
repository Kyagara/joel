package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

var (
	ErrNoTracksFound = errors.New("no tracks found")
	ErrLoadingTracks = errors.New("loading tracks failed")
)

func play(event *events.ApplicationCommandInteractionCreate, url string) {
	url = strings.TrimSpace(url)
	if !strings.HasPrefix(url, "http") {
		url = "ytsearch:" + url
	}

	guildID := *event.GuildID()

	avatar := ""
	if event.User().AvatarURL() != nil {
		avatar = *event.User().AvatarURL()
	}

	user := UserInfo{
		Username: event.User().EffectiveName(),
		Avatar:   avatar,
	}

	track, err := handleUserQuery(user, url)
	if err != nil {
		reply(event, err.Error())
		return
	}

	userVoice, err := CLIENT.Rest.GetUserVoiceState(guildID, event.User().ID)
	if userVoice == nil || err != nil {
		reply(event, "You are not in a voice channel.")
		return
	}

	updateVoiceChannel(event, userVoice.ChannelID)

	player := CLIENT.Lavalink.Player(guildID)

	if !TRACKS.Empty() {
		TRACKS.Push(track)
		fmt.Printf("Queued track: %s\n", track.Info.Title)
		reply(event, fmt.Sprintf("Queued track: %s\n", track.Info.Title))
		return
	}

	TRACKS.Push(track)

	err = player.Update(context.TODO(), lavalink.WithTrack(track))
	if err != nil {
		fmt.Printf("Error playing track: %v\n", err)
		reply(event, err.Error())
		return
	}

	embed := TRACKS.GetTrackEmbed(track)
	message := discord.NewMessageCreateBuilder().SetEmbeds(embed).Build()
	sendMessage(event, message)
}

func pause(event *events.ApplicationCommandInteractionCreate) {
	guildID := *event.GuildID()

	voice := getBotVoiceState(event)
	if voice == nil {
		reply(event, "The bot is not in a voice channel.")
		return
	}

	if TRACKS.Empty() {
		reply(event, "No tracks currently playing.")
		return
	}

	player := CLIENT.Lavalink.Player(guildID)

	err := player.Update(context.TODO(), lavalink.WithPaused(true))
	if err != nil {
		fmt.Printf("Error pausing: %v\n", err)
		reply(event, err.Error())
	}
}

func resume(event *events.ApplicationCommandInteractionCreate) {
	guildID := *event.GuildID()

	voice := getBotVoiceState(event)
	if voice == nil {
		reply(event, "The bot is not in a voice channel.")
		return
	}

	if TRACKS.Empty() {
		reply(event, "No tracks currently playing.")
		return
	}

	player := CLIENT.Lavalink.Player(guildID)

	err := player.Update(context.TODO(), lavalink.WithPaused(false))
	if err != nil {
		fmt.Printf("Error resuming: %v\n", err)
		reply(event, err.Error())
	}
}

func skip(event *events.ApplicationCommandInteractionCreate) {
	guildID := *event.GuildID()

	voice := getBotVoiceState(event)
	if voice == nil {
		reply(event, "The bot is not in a voice channel.")
		return
	}

	length := TRACKS.Len()

	if length == 0 {
		reply(event, "No tracks currently playing.")
		return
	}

	player := CLIENT.Lavalink.Player(guildID)

	if length == 1 {
		err := player.Update(context.TODO(), lavalink.WithNullTrack())
		if err != nil {
			fmt.Printf("Error playing track: %v\n", err)
			reply(event, err.Error())
			return
		}

		return
	}

	user := UserInfo{
		Skipped: true,
	}

	skippedTrack, err := TRACKS.First().WithUserData(user)
	if err != nil {
		fmt.Printf("Error adding skipped user data: %v\n", err)
		return
	}

	TRACKS.Replace(0, skippedTrack)
	track := TRACKS.Get(1)

	err = player.Update(context.TODO(), lavalink.WithNullTrack())
	if err != nil {
		fmt.Printf("Error playing track: %v\n", err)
		reply(event, err.Error())
		return
	}

	embed := TRACKS.GetTrackEmbed(track)
	message := discord.NewMessageCreateBuilder().SetEmbeds(embed).Build()
	sendMessage(event, message)
}

func stop(event *events.ApplicationCommandInteractionCreate) {
	guildID := *event.GuildID()

	voice := getBotVoiceState(event)
	if voice == nil {
		reply(event, "The bot is not in a voice channel.")
		return
	}

	if TRACKS.Empty() {
		reply(event, "No tracks currently playing.")
		return
	}

	player := CLIENT.Lavalink.Player(guildID)

	err := player.Update(context.TODO(), lavalink.WithNullTrack())
	if err != nil {
		fmt.Printf("Error stopping: %v\n", err)
		reply(event, err.Error())
	}
}

func join(event *events.ApplicationCommandInteractionCreate) {
	guildID := *event.GuildID()

	voice, err := CLIENT.Rest.GetCurrentUserVoiceState(guildID)
	if err != nil {
		fmt.Printf("Error getting bot voice state: %v\n", err)
		reply(event, err.Error())
		return
	}

	if voice != nil {
		reply(event, "The bot is already in a voice channel.")
		return
	}

	userID := event.User().ID

	userVoice, err := CLIENT.Rest.GetUserVoiceState(guildID, userID)
	if userVoice == nil || err != nil {
		reply(event, "You are not in a voice channel.")
		return
	}

	updateVoiceChannel(event, userVoice.ChannelID)
}

func leave(event *events.ApplicationCommandInteractionCreate) {
	guildID := *event.GuildID()

	voice := getBotVoiceState(event)
	if voice == nil {
		reply(event, "The bot is not in a voice channel.")
		return
	}

	if !TRACKS.Empty() {
		player := CLIENT.Lavalink.Player(guildID)

		err := player.Update(context.TODO(), lavalink.WithNullTrack())
		if err != nil {
			fmt.Printf("Error stopping track before leaving: %v\n", err)
			reply(event, err.Error())
			return
		}
	}

	updateVoiceChannel(event, nil)
}

func queue(event *events.ApplicationCommandInteractionCreate) {
	if TRACKS.Empty() {
		reply(event, "No tracks currently playing.")
		return
	}

	tracks := TRACKS.Few(5)

	var queue []string
	for i, track := range tracks {
		user := UserInfo{}
		err := track.UserData.Unmarshal(&user)
		if err != nil {
			fmt.Printf("Error scanning user: %v\n", err)
			reply(event, err.Error())
			return
		}

		queue = append(queue, fmt.Sprintf("%d. %s - %s", i+1, track.Info.Title, user.Username))
	}

	reply(event, strings.Join(queue, "\n"))
}

func playing(event *events.ApplicationCommandInteractionCreate) {
	if TRACKS.Empty() {
		reply(event, "No tracks currently playing.")
		return
	}

	embed := TRACKS.GetTrackEmbed(TRACKS.First())
	message := discord.NewMessageCreateBuilder().SetEmbeds(embed).Build()
	sendMessage(event, message)
}
