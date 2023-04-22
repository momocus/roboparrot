package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	loadEnv()

	token := os.Getenv("DISCORD_TOKEN")
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
			fmt.Println("Error creating Discord session: ", err)
			return
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
			fmt.Println("Error opening Discord session: ", err)
			return
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	<-make(chan struct{})
	return
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}


func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
			return
	}

	if strings.HasPrefix(m.Content, "!gpt ") {
			response, _ := chat(m.Content[5:])
			s.ChannelMessageSend(m.ChannelID, response)
	}
}

// creates a message by GPT-3.5
func chat(message string) (string, error) {
	apiKey := os.Getenv("OPENAI_TOKEN")
	client := openai.NewClient(apiKey)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: message,
				},
			},
		},
	)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
