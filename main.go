package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	apiURL = "https://api.groq.com/openai/v1/chat/completions"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Response struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func main() {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		fmt.Println("Please set the API_KEY environment variable to your Groq API key, from here: https://console.groq.com/keys.")
		fmt.Println("Exiting program...")
		os.Exit(0)
	}

	model := os.Getenv("MODEL")
	if model == "" {
		fmt.Println("No MODEL variable set. Defaulting to llama-3.1-70b-versatile")
		model = "llama-3.1-70b-versatile"
	}

	fmt.Println("Chat with your chosen LLM through GroqCloud!")
	fmt.Println("Type 'exit_chat' to quit the chat.")

	messages := []Message{}
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		userInput := scanner.Text()
		if userInput == "exit_chat" {
			break
		}

		messages = append(messages, Message{Role: "user", Content: userInput})

		response, err := sendRequest(apiKey, messages, model)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		if len(response.Choices) > 0 {
			assistantMessage := response.Choices[0].Message
			messages = append(messages, assistantMessage)
			fmt.Printf("Assistant: %s\n", assistantMessage.Content)
		} else {
			fmt.Println("No response from the assistant.")
		}
	}

	fmt.Println("Goodbye!")
}

func sendRequest(apiKey string, messages []Message, model string) (*Response, error) {
	requestBody := Request{
		Model:    model,
		Messages: messages,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
