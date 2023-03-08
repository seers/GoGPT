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

type ErrResp struct {
	Error struct {
		Message string      `json:"message"`
		Type    string      `json:"type"`
		Param   interface{} `json:"param"`
		Code    interface{} `json:"code"`
	} `json:"error"`
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
	d := color.New(color.FgCyan, color.Bold)
	red := color.New(color.FgRed)
	boldRed := red.Add(color.Bold)

	flag.Parse()

	if APIKey == "" {
		APIKey = os.Getenv("API_KEY")
	}

	if APIKey == "" {
		fmt.Println("Error: APIKey is required")
		flag.Usage()
		os.Exit(2)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}

	if proxy != "" {
		u, err := url.Parse(proxy)
		if err != nil {
			fmt.Println("Error: Proxy format error")
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

	fmt.Print("> ")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Print("> ")
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

		req, err := http.NewRequest("POST", APIURL, bf)
		if err != nil {
			os.Exit(2)
		}

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+APIKey)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error: connection lost")
			os.Exit(2)
		}

		if resp.StatusCode == 200 {
			chatResp := &ChatResp{}
			err = json.NewDecoder(resp.Body).Decode(chatResp)
			if err != nil {
				os.Exit(2)
			}
			resp.Body.Close()

			for _, v := range chatResp.Choices {
				for _, r := range v.Message.Content {
					d.Print(string(r))
					time.Sleep(10 * time.Millisecond)
				}
			}

			fmt.Println()
			fmt.Println()

			fmt.Print("> ")
			if chatResp.Usage.TotalTokens == 4096 {
				boldRed.Print("We reach the end of conversation")
				os.Exit(2)
			}

			messages = append(messages, chatResp.Choices[0].Message)
		} else {
			errResp := &ErrResp{}
			err := json.NewDecoder(resp.Body).Decode(errResp)
			if err != nil {
				os.Exit(2)
			}
			resp.Body.Close()

			boldRed.Print("Server stop the conversation, ", errResp.Error.Message)
			os.Exit(2)
		}
	}
}
