package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/go-json-experiment/json"
)

var (
	CHATS = map[snowflake.ID]Chat{}

	LLM_QUEUE = make(chan LLMRequest, 1)
)

type LLMRequest struct {
	Prompt  string
	Message discord.Message
}

type LLMResult struct {
	Content string
	Usage   Usage
}

type Chat struct {
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Response struct {
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Message Message `json:"message"`
}

type Usage struct {
	// Will be replaced
	TotalTime        uint64 `json:"total_time"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
}

func handleUserMessage(event *events.MessageCreate) {
	messageID := event.MessageID
	channelID := event.Message.ChannelID
	content := event.Message.Content

	if event.Message.GuildID != nil {
		mentioned := isMentioned(event.Message.Mentions)

		if event.Message.Type == discord.MessageTypeReply && !mentioned {
			return
		}

		if !mentioned {
			return
		}

		content = strings.Replace(content, MENTION, "", 1)
	}

	status, err := pingLLMServer()
	if err != nil || status != 200 {
		fmt.Printf("Error pinging server: %v | %v\n", status, err)
		addReaction(channelID, messageID, ERR_EMOJI)
		return
	}

	addReaction(channelID, messageID, FISH_EMOJI)

	var prompt string

	if len(event.Message.Attachments) == 1 {
		url := event.Message.Attachments[0].URL
		res, err := CLIENT.Rest.HTTPClient().Get(url)
		if err != nil {
			fmt.Printf("Error downloading attachment: %v\n", err)
			addReaction(channelID, messageID, ERR_EMOJI)
			return
		}
		defer res.Body.Close()

		var body string

		ext := filepath.Ext(event.Message.Attachments[0].Filename)

		switch ext {
		case ".txt":
			buffer, err := io.ReadAll(res.Body)
			if err != nil {
				fmt.Printf("Error reading txt: %v\n", err)
				addReaction(channelID, messageID, ERR_EMOJI)
				return
			}

			body = string(buffer)

		case ".pdf":
			var buffer bytes.Buffer
			_, err = io.Copy(&buffer, res.Body)
			if err != nil {
				fmt.Printf("Error reading pdf: %v\n", err)
				addReaction(channelID, messageID, ERR_EMOJI)
				return
			}

			body, err = readPDF(buffer)
			if err != nil {
				fmt.Printf("Error parsing pdf: %v\n", err)
				addReaction(channelID, messageID, ERR_EMOJI)
				return
			}
		}

		prompt = fmt.Sprintf("%s\n\n%s", content, body)
	} else {
		prompt = event.Message.Content
	}

	go func() {
		LLM_QUEUE <- LLMRequest{
			Prompt:  prompt,
			Message: event.Message,
		}
	}()
}

func submitLLMChat(user snowflake.ID, prompt string) (LLMResult, error) {
	chat := CHATS[user]
	if chat.Messages == nil {
		CHATS[user] = Chat{
			Messages: []Message{
				{
					Role:    "system",
					Content: CONFIG.Prompt,
				},
			},
		}
		chat = CHATS[user]
	}

	chat.Messages = append(chat.Messages, Message{
		Role:    "user",
		Content: prompt,
	})

	now := time.Now()

	body, err := json.Marshal(chat)
	if err != nil {
		return LLMResult{}, err
	}

	buffer := bytes.NewBuffer(body)

	req, err := http.NewRequest("POST", "http://localhost:2444/v1/chat/completions", buffer)
	if err != nil {
		return LLMResult{}, err
	}

	res, err := HTTP.Do(req)
	if err != nil {
		return LLMResult{}, err
	}
	defer res.Body.Close()

	elapsed := time.Until(now)

	var response Response
	err = json.UnmarshalRead(res.Body, &response)
	if err != nil {
		return LLMResult{}, err
	}

	response.Usage.TotalTime = uint64(elapsed.Seconds())

	firstChoice := response.Choices[0]
	newMessage := firstChoice.Message
	newMessage.Role = "assistant"

	chat.Messages = append(chat.Messages, newMessage)
	CHATS[user] = chat

	result := LLMResult{
		Content: newMessage.Content,
		Usage:   response.Usage,
	}

	return result, nil
}

func processLLM(request LLMRequest) {
	err := CLIENT.Rest.SendTyping(request.Message.ChannelID)
	if err != nil {
		fmt.Printf("Error sending typing: %v\n", err)
	}

	result, err := submitLLMChat(request.Message.Author.ID, request.Prompt)
	if err != nil {
		fmt.Printf("Error submitting chat: %v\n", err)
		addReaction(request.Message.ChannelID, request.Message.ID, ERR_EMOJI)
		return
	}

	fmt.Printf("User %s | Time: %ds | Prompt Tokens: %d | Completion Tokens: %d\n", request.Message.Author.Username, result.Usage.TotalTime, result.Usage.PromptTokens, result.Usage.CompletionTokens)

	// If the message is too long, send it as a file
	if len(result.Content) > 1950 {
		message := discord.NewMessageCreateBuilder()
		file := bytes.NewBuffer([]byte(result.Content))
		message.AddFile("message.txt", "The message was too long to send normally, so it's been attached as a file.", file)
		message.SetMessageReferenceByID(request.Message.ID)

		_, err = CLIENT.Rest.CreateMessage(request.Message.ChannelID, message.Build())
		if err != nil {
			fmt.Printf("Error sending reply with attachment: %v\n", err)
			addReaction(request.Message.ChannelID, request.Message.ID, ERR_EMOJI)
			return
		}

		return
	}

	message := discord.NewMessageCreateBuilder().SetContent(result.Content).SetMessageReferenceByID(request.Message.ID).Build()

	_, err = CLIENT.Rest.CreateMessage(request.Message.ChannelID, message)
	if err != nil {
		fmt.Printf("Error sending reply: %v\n", err)
		addReaction(request.Message.ChannelID, request.Message.ID, ERR_EMOJI)
	}
}

func pingLLMServer() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:2444/health", nil)
	if err != nil {
		return 0, err
	}

	res, err := HTTP.Do(req)
	if err != nil {
		return 0, err
	}

	return res.StatusCode, nil
}

func reset(event *events.ApplicationCommandInteractionCreate) {
	user := event.User().ID
	length := len(CHATS[user].Messages)
	CHATS[user] = Chat{}

	message := fmt.Sprintf("Cleared %d messages from your chat.", length)
	reply(event, message)
}
