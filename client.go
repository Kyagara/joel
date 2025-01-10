package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/disgolink/v3/disgolink"
)

type Client struct {
	Bot      bot.Client
	Rest     rest.Rest
	Lavalink disgolink.Client
}

func NewClient() error {
	assets, err := os.ReadDir("assets")
	if err != nil {
		return err
	}

	choices := make([]discord.ApplicationCommandOptionChoiceString, 0, 25)

	for _, asset := range assets {
		name := asset.Name()
		name = name[:len(name)-len(filepath.Ext(name))]
		ASSETS_PATHS[name] = fmt.Sprintf("./assets/%s", asset.Name())
		ASSETS = append(ASSETS, name)

		choices = append(choices, discord.ApplicationCommandOptionChoiceString{
			Name:  name,
			Value: name,
		})
	}

	for i, command := range COMMANDS {
		if command.CommandName() == "joel" {
			COMMANDS[i].(discord.SlashCommandCreate).Options[0] = discord.ApplicationCommandOptionString{
				Name:        "joel",
				Description: "JOEL",
				DescriptionLocalizations: map[discord.Locale]string{
					discord.LocaleEnglishUS:    "Posts a random or specific JOEL",
					discord.LocalePortugueseBR: "Posta um JOEL aleatório ou um específico",
				},
				Choices:  choices,
				Required: false,
			}
			break
		}
	}

	bot, err := disgo.New(CONFIG.Token,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentGuilds,
				gateway.IntentGuildMessages,
				gateway.IntentDirectMessages,
				gateway.IntentMessageContent,
				gateway.IntentGuildVoiceStates,
			),

			gateway.WithPresenceOpts(
				gateway.WithPlayingActivity("/help"),
				gateway.WithOnlineStatus(discord.OnlineStatusDND),
			),
		),

		bot.WithEventListenerFunc(onReady),
		bot.WithEventListenerFunc(onMessageCreate),

		bot.WithEventListenerFunc(onVoiceStateUpdate),
		bot.WithEventListenerFunc(onVoiceServerUpdate),

		bot.WithEventListenerFunc(commandListener),
	)

	if err != nil {
		return err
	}

	fmt.Println("Connecting to Discord")
	err = bot.OpenGateway(context.TODO())
	if err != nil {
		return err
	}

	// Wait for ready
	<-READY

	APPLICATION_ID := bot.ApplicationID()
	BOT_ID := bot.ID()
	MENTION = fmt.Sprintf("<@%s>", APPLICATION_ID)

	if *REGISTER {
		cmds, err := bot.Rest().SetGlobalCommands(APPLICATION_ID, COMMANDS)
		if err != nil {
			fmt.Printf("Error setting global commands: %v\n", err)
			panic(err)
		}

		fmt.Printf("Registered %d global commands\n", len(cmds))
	}

	MENTION = fmt.Sprintf("<@%s>", BOT_ID)

	fmt.Println("Connecting to Lavalink")
	lavalink := disgolink.New(BOT_ID, disgolink.WithListenerFunc(onTrackEnd))
	_, err = lavalink.AddNode(context.TODO(), disgolink.NodeConfig{
		Name:    "joel",
		Address: "0.0.0.0:2333",
	})

	if err != nil {
		return err
	}

	CLIENT = Client{
		Bot:      bot,
		Lavalink: lavalink,
		Rest:     bot.Rest(),
	}

	return nil
}

var (
	COMMANDS = []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name: "help",
			NameLocalizations: map[discord.Locale]string{
				discord.LocaleEnglishUS:    "help",
				discord.LocalePortugueseBR: "ajuda",
			},
			Description: "Displays the help message",
			DescriptionLocalizations: map[discord.Locale]string{
				discord.LocaleEnglishUS:    "Displays the help message",
				discord.LocalePortugueseBR: "Exibe a mensagem de ajuda",
			},
		},

		discord.SlashCommandCreate{
			Name:        "joel",
			Description: "JOEL",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:                     "",
					Description:              "",
					DescriptionLocalizations: map[discord.Locale]string{},
					Choices:                  []discord.ApplicationCommandOptionChoiceString{},
					Required:                 false,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "ttj",
			Description: "Latency test",
			DescriptionLocalizations: map[discord.Locale]string{
				discord.LocaleEnglishUS:    "Latency test",
				discord.LocalePortugueseBR: "Teste de latência",
			},
		},

		discord.SlashCommandCreate{
			Name:        "play",
			Description: "Plays a track",
			DescriptionLocalizations: map[discord.Locale]string{
				discord.LocaleEnglishUS:    "Plays a track",
				discord.LocalePortugueseBR: "Toca uma música",
			},
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name: "query",
					NameLocalizations: map[discord.Locale]string{
						discord.LocaleEnglishUS:    "query",
						discord.LocalePortugueseBR: "pesquisa",
					},
					Description: "Can be an URL or a search query",
					DescriptionLocalizations: map[discord.Locale]string{
						discord.LocaleEnglishUS:    "Can be an URL or a search query",
						discord.LocalePortugueseBR: "Pode ser uma URL ou uma pesquisa",
					},
					Required: true,
				},
			},
		},
		discord.SlashCommandCreate{
			Name: "stop",
			NameLocalizations: map[discord.Locale]string{
				discord.LocaleEnglishUS:    "stop",
				discord.LocalePortugueseBR: "parar",
			},
			Description: "Stops the current track",
			DescriptionLocalizations: map[discord.Locale]string{
				discord.LocaleEnglishUS:    "Stops the current track",
				discord.LocalePortugueseBR: "Para a música atual",
			},
		},
		discord.SlashCommandCreate{
			Name: "pause",
			NameLocalizations: map[discord.Locale]string{
				discord.LocaleEnglishUS:    "pause",
				discord.LocalePortugueseBR: "pausar",
			},
			Description: "Pauses the current track",
			DescriptionLocalizations: map[discord.Locale]string{
				discord.LocaleEnglishUS:    "Pauses the current track",
				discord.LocalePortugueseBR: "Pausa a música atual",
			},
		},
		discord.SlashCommandCreate{
			Name: "resume",
			NameLocalizations: map[discord.Locale]string{
				discord.LocaleEnglishUS:    "resume",
				discord.LocalePortugueseBR: "resumir",
			},
			Description: "Resumes the current track",
			DescriptionLocalizations: map[discord.Locale]string{
				discord.LocaleEnglishUS:    "Resumes the current track",
				discord.LocalePortugueseBR: "Resume a música atual",
			},
		},
		discord.SlashCommandCreate{
			Name: "skip",
			NameLocalizations: map[discord.Locale]string{
				discord.LocaleEnglishUS:    "skip",
				discord.LocalePortugueseBR: "pular",
			},
			Description: "Skips the current track",
			DescriptionLocalizations: map[discord.Locale]string{
				discord.LocaleEnglishUS:    "Skips the current track",
				discord.LocalePortugueseBR: "Pula a música atual",
			},
		},
		discord.SlashCommandCreate{
			Name: "join",
			NameLocalizations: map[discord.Locale]string{
				discord.LocaleEnglishUS:    "join",
				discord.LocalePortugueseBR: "entrar",
			},
			Description: "Joins the voice channel",
			DescriptionLocalizations: map[discord.Locale]string{
				discord.LocaleEnglishUS:    "Joins the voice channel",
				discord.LocalePortugueseBR: "Entra no canal de voz",
			},
		},
		discord.SlashCommandCreate{
			Name: "leave",
			NameLocalizations: map[discord.Locale]string{
				discord.LocaleEnglishUS:    "leave",
				discord.LocalePortugueseBR: "sair",
			},
			Description: "Leaves the voice channel",
			DescriptionLocalizations: map[discord.Locale]string{
				discord.LocaleEnglishUS:    "Leaves the voice channel",
				discord.LocalePortugueseBR: "Sai do canal de voz",
			},
		},
		discord.SlashCommandCreate{
			Name: "queue",
			NameLocalizations: map[discord.Locale]string{
				discord.LocaleEnglishUS:    "queue",
				discord.LocalePortugueseBR: "fila",
			},
			Description: "Displays the queue",
			DescriptionLocalizations: map[discord.Locale]string{
				discord.LocaleEnglishUS:    "Displays the music queue",
				discord.LocalePortugueseBR: "Exibe a fila de músicas",
			},
		},
		discord.SlashCommandCreate{
			Name: "playing",
			NameLocalizations: map[discord.Locale]string{
				discord.LocaleEnglishUS:    "playing",
				discord.LocalePortugueseBR: "tocando",
			},
			Description: "Displays the currently playing track",
			DescriptionLocalizations: map[discord.Locale]string{
				discord.LocaleEnglishUS:    "Displays the currently playing track",
				discord.LocalePortugueseBR: "Exibe a música atual",
			},
		},

		discord.SlashCommandCreate{
			Name:        "reset",
			Description: "Resets your chat history with the bot",
		},
	}
)
