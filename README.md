# joel-go

I have made this same bot (parts of it) around three times by now in different languages, I felt like giving it a Go. For old versions check the [old repo](https://github.com/Kyagara/joel-old) (Elixir and Rust).

It was made to support just one guild and DMs, though it would probably take just some tweaking to support multiple servers properly, haven't tested that yet.

## Installation

Requires Lavalink, you might need to setup `oauth` for youtube. After setting up your bot, remember to enable `applications.commands` and the `Message Content` privilege. Run the bot once with the `-register` flag.

## Features

If in a DM channel you can just talk to the bot and it will respond using an OpenAI API compatible endpoint, I use [llama.cpp](https://github.com/ggerganov/llama.cpp) for this. In a server you can just reply to any message from the bot and type your prompt, don't forget to not unmark the "Ping the user" option, or, you can send a new message mentioning the bot. For models I generally use `llama-3.2-1b-instruct`, for llama.cpp you will need a `gguf` file.

Plays music using [Lavalink](https://github.com/lavalink-devs/Lavalink), play command supports search or direct links (http, youtube, etc.)

## Slash commands

`help`, `reset`, `joel`, `ttj`, `play`, `stop`, `pause`, `resume`, `skip`, `join`, `leave`, `queue`, `playing`

- reset: Resets the users chat history with the bot
- joel: Posts JOEL
- ttj: Posts Time to Joel (latency test)
- play: Plays a song
- stop: Stops the current song
- pause: Pauses the current song
- resume: Resumes the current song
- skip: Skips the current song
- join: Joins the voice channel
- leave: Leaves the voice channel
- queue: Shows the queue
- playing: Shows the current song
