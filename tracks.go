package main

import (
	"fmt"
	"sync"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

type Tracks struct {
	store []lavalink.Track
	mu    sync.Mutex
}

func (t *Tracks) First() lavalink.Track {
	t.mu.Lock()
	track := t.store[0]
	t.mu.Unlock()
	return track
}

func (t *Tracks) All() []lavalink.Track {
	t.mu.Lock()
	tracks := t.store
	t.mu.Unlock()
	return tracks
}

func (t *Tracks) Few(n int) []lavalink.Track {
	t.mu.Lock()
	if n > len(t.store) {
		n = len(t.store)
	}
	tracks := t.store[:n]
	t.mu.Unlock()
	return tracks
}

func (t *Tracks) Get(index int) lavalink.Track {
	t.mu.Lock()
	track := t.store[index]
	t.mu.Unlock()
	return track
}

func (t *Tracks) GetTrackEmbed(track lavalink.Track) discord.Embed {
	info := track.Info
	user := UserInfo{}

	err := track.UserData.Unmarshal(&user)
	if err != nil {
		return discord.NewEmbedBuilder().SetDescription(err.Error()).Build()
	}

	embed := discord.NewEmbedBuilder().
		SetTitle(info.Title).
		SetURL(*info.URI).
		SetThumbnail(*info.ArtworkURL).
		SetAuthor(info.Author, "", "").
		SetFooter(fmt.Sprintf("Requested by %s", user.Username), user.Avatar).
		Build()

	return embed
}

func (t *Tracks) Push(track lavalink.Track) {
	t.mu.Lock()
	t.store = append(t.store, track)
	t.mu.Unlock()
}

func (t *Tracks) Pop() lavalink.Track {
	t.mu.Lock()
	track := t.store[0]
	t.store = t.store[1:]
	t.mu.Unlock()
	return track
}

func (t *Tracks) Replace(index int, track lavalink.Track) {
	t.mu.Lock()
	t.store[index] = track
	t.mu.Unlock()
}

func (t *Tracks) Len() int {
	t.mu.Lock()
	length := len(t.store)
	t.mu.Unlock()
	return length
}

func (t *Tracks) Empty() bool {
	t.mu.Lock()
	empty := len(t.store) == 0
	t.mu.Unlock()
	return empty
}
