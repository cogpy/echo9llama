package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ollama/ollama/api"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test.go <model-name>")
		os.Exit(1)
	}

	modelName := os.Args[1]
	fmt.Printf("Testing model: %s\n", modelName)

	client, err := api.ClientFromEnvironment()
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req := &api.ChatRequest{
		Model:    modelName,
		Messages: []api.Message{{Role: "user", Content: "Hello"}},
		Stream:   new(bool),
	}

	err = client.Chat(ctx, req, func(resp api.ChatResponse) error {
		fmt.Printf("Response: %s\n", resp.Message.Content)
		return nil
	})

	if err != nil {
		fmt.Printf("Test failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Test passed")
}