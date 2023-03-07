package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/fatih/color"
)

type ChatReq struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResp struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

var proxy string
var APIKey string
var Model string

const (
	APIURL = "https://api.openai.com/v1/chat/completions"
)

func init() {
	flag.StringVar(&proxy, "p", "", "Proxy address, eg http://127.0.0.1:7890 or sock5://127.0.0.1:7890")
	flag.StringVar(&APIKey, "k", "", "Your API Key")
	flag.StringVar(&Model, "m", "gpt-3.5-turbo", "The Model to chat with")
}

func main() {
	flag.Parse()

	if APIKey == "" {
		fmt.Println("Error: APIKey is required")
		flag.Usage()
		os.Exit(2)
	}

	client := &http.Client{}

	if proxy != "" {
		u, err := url.Parse(proxy)
		if err != nil {
			fmt.Println("Error: Proxy format")
			flag.Usage()
			os.Exit(2)
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(u),
		}
	}

	chatReq := &ChatReq{
		Model: Model,
	}

	messages := []Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant.",
		},
	}

	fmt.Println("Start ask your question!")
	fmt.Print("> ")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}

		messages = append(messages, Message{
			Role:    "user",
			Content: scanner.Text(),
		})

		chatReq.Messages = messages

		bf := &bytes.Buffer{}
		err := json.NewEncoder(bf).Encode(chatReq)
		if err != nil {
			os.Exit(2)
		}

		// log.Println("send:", bf.String())

		req, err := http.NewRequest("POST", APIURL, bf)
		if err != nil {
			os.Exit(2)
		}

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+APIKey)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error: can't connect to server, try proxy")
			flag.Usage()
			os.Exit(2)
		}

		chatResp := &ChatResp{}
		// log.Printf("%+v\n", chatResp)
		err = json.NewDecoder(resp.Body).Decode(chatResp)
		if err != nil {
			os.Exit(2)
		}
		resp.Body.Close()

		fmt.Print("> ")
		for _, v := range chatResp.Choices {
			for _, r := range v.Message.Content {
				d := color.New(color.FgCyan, color.Bold)
				d.Print(string(r))
				time.Sleep(20 * time.Millisecond)
			}
		}

		fmt.Println()
		fmt.Println()

		fmt.Print("> ")
		if chatResp.Usage.TotalTokens == 4096 {
			red := color.New(color.FgRed)
			boldRed := red.Add(color.Bold)
			boldRed.Print("We reach the end of conversation")
			os.Exit(0)
		}

		if len(chatResp.Choices) == 0 {
			red := color.New(color.FgRed)
			boldRed := red.Add(color.Bold)
			boldRed.Print("Server stop the conversation")
			os.Exit(0)
		}

		messages = append(messages, chatResp.Choices[0].Message)

	}
}
