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
	// ボット自身からのメッセージは無視する
	if m.Author.ID == s.State.User.ID {
		return
	}

	// メッセージにボットへのメンションが含まれていなければ無視する
	if !isBotMentioned(s.State.User.ID, m.Message.Mentions) {
		return
	}

	response, _ := callGPT4(m.Content)
	s.ChannelMessageSend(m.ChannelID, response)
}

func isBotMentioned(botID string, mentions []*discordgo.User) bool {
	for _, user := range mentions {
		if user.ID == botID {
			return true
		}
	}
	return false
}

// creates a message by GPT-4
func callGPT4(message string) (string, error) {
	message = strings.TrimSpace(message)
	// 履歴を維持するため、現在の会話履歴に新しいプロンプトを追加します。
	conversationHistory += "User: " + message + "\nAI: "

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
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
