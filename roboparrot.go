package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

var (
	client              *openai.Client
	conversationHistory string
)

func init() {
	loadEnv()
	apiKey := os.Getenv("OPENAI_API_KEY")
	client = openai.NewClient(apiKey)
}

func main() {
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

// messageCreate is called when a message is sent on discord
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	response, _ := callGPT3Dot5TurboAPI(m.Content)
	s.ChannelMessageSend(m.ChannelID, response)
}

// creates a message by GPT-3.5
func callGPT3Dot5TurboAPI(message string) (string, error) {
	message = strings.TrimSpace(message)
	// 履歴を維持するため、現在の会話履歴に新しいプロンプトを追加します。
	conversationHistory += "User: " + message + "\nAI: "

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: conversationHistory,
				},
			},
		},
	)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return "", err
	}
	// 生成されたレスポンスを履歴に追加
	response := resp.Choices[0].Message.Content
	conversationHistory += response + "\n"

	// 履歴をJSONファイルに書き出す
	err = writeHistoryToJSONFile(conversationHistory)
	if err != nil {
		return "", err
	}

	return response, nil
}

func writeHistoryToJSONFile(history string) error {
	// 履歴をJSON形式で保存するために、文字列をマップに変換します。
	historyMap := map[string]string{
		"history": history,
	}

	// マップをJSONに変換します。
	jsonHistory, err := json.Marshal(historyMap)
	if err != nil {
		return err
	}

	// JSONデータをファイルに書き込みます。
	err = ioutil.WriteFile("conversation_history.json", jsonHistory, 0644)
	if err != nil {
		return err
	}
	return nil
}
