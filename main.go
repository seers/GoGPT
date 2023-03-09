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
	"regexp"
	"time"

	"github.com/fatih/color"
)

type ChatReq struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
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
	APIURL        = "https://api.openai.com/v1/chat/completions"
	ServerTimeout = 15
	MaxToken      = 4096
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
		Timeout: ServerTimeout * time.Second,
	}

	if proxy != "" {
		validIP := regexp.MustCompile(`^(https?|socks5)://[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}:[0-9]{1,5}$`)
		if !validIP.MatchString(proxy) {
			boldRed.Println("Error: Proxy format")
			os.Exit(2)
		}
		u, _ := url.Parse(proxy)
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(u),
		}
	}

	chatReq := &ChatReq{
		Model:  Model,
		Stream: true,
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
		if scanner.Text() == "" {
			fmt.Print("> ")
			continue
		}

		d.Print("> ")

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
			boldRed.Println("Error: Server Timeout")
			os.Exit(2)
		}

		if resp.StatusCode == http.StatusOK {
			dec := json.NewDecoder(resp.Body)

			for dec.More() {
				chatResp := &ChatResp{}

				err := dec.Decode(chatResp)
				if err != nil {
					os.Exit(2)
				}

				d.Print(chatResp.Choices[0].Message.Content)

				if chatResp.Usage.TotalTokens >= MaxToken {
					boldRed.Print("We reach the end of conversation")
					os.Exit(2)
				}

				messages = append(messages, chatResp.Choices[0].Message)
			}

			resp.Body.Close()

			fmt.Println()
			fmt.Println()

			fmt.Print("> ")
		} else {
			errResp := &ErrResp{}
			err := json.NewDecoder(resp.Body).Decode(errResp)
			if err != nil {
				os.Exit(2)
			}
			resp.Body.Close()

			boldRed.Println("Server stop the conversation,", errResp.Error.Message)
			os.Exit(2)
		}
	}
}
