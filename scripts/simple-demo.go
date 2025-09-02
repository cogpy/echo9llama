package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ollama/ollama/api"
	"github.com/ollama/ollama/orchestration"
)

// SimpleDemo demonstrates basic echollama functionality with a real model
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run simple-demo.go <model-name>")
		os.Exit(1)
	}

	modelName := os.Args[1]
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Printf("🌊 EchoLlama Simple Demo with %s\n", modelName)
	fmt.Printf("=" + strings.Repeat("=", 40) + "\n")

	// Test 1: Basic model chat
	fmt.Println("\n💬 Testing basic model conversation...")
	if err := testBasicChat(ctx, modelName); err != nil {
		fmt.Printf("⚠️  Basic chat test: %v (server may not be running)\n", err)
	} else {
		fmt.Println("✅ Basic chat test successful")
	}

	// Test 2: EchoChat system
	fmt.Println("\n🧠 Testing EchoChat integration...")
	if err := testEchoChatDemo(ctx, modelName); err != nil {
		fmt.Printf("⚠️  EchoChat test: %v\n", err)
	} else {
		fmt.Println("✅ EchoChat test successful")
	}

	fmt.Printf("\n🎉 Demo completed for model: %s\n", modelName)
}

func testBasicChat(ctx context.Context, modelName string) error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return fmt.Errorf("client initialization failed: %w", err)
	}

	// Quick connectivity test
	req := &api.ChatRequest{
		Model: modelName,
		Messages: []api.Message{
			{Role: "user", Content: "Say hello and tell me you're working correctly!"},
		},
		Stream: new(bool),
	}

	fmt.Printf("   🤖 Asking %s to respond...\n", modelName)
	
	connectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	responseReceived := false
	err = client.Chat(connectCtx, req, func(resp api.ChatResponse) error {
		if !responseReceived {
			fmt.Printf("   📝 Model response: %s\n", strings.TrimSpace(resp.Message.Content))
			responseReceived = true
		}
		return nil
	})

	if err != nil {
		return err
	}

	if !responseReceived {
		return fmt.Errorf("no response received from model")
	}

	return nil
}

func testEchoChatDemo(ctx context.Context, modelName string) error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		// Try offline demo
		fmt.Printf("   ⚠️  Running offline demo (server not available)\n")
		return testOfflineDemo()
	}

	// Test server connectivity
	testCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	err = client.Heartbeat(testCtx)
	if err != nil {
		fmt.Printf("   ⚠️  Running offline demo (server not responding)\n")
		return testOfflineDemo()
	}

	// Run online demo with real model
	fmt.Printf("   🌊 Creating EchoChat with Deep Tree Echo...\n")
	
	engine := orchestration.NewEngine(*client)
	orchestration.RegisterDefaultTools(engine)
	orchestration.RegisterDefaultPlugins(engine)

	// Initialize Deep Tree Echo
	err = engine.InitializeDeepTreeEcho(ctx)
	if err != nil {
		fmt.Printf("   ⚠️  Deep Tree Echo warning: %v\n", err)
	}

	echoChat := orchestration.NewEchoChat(engine)

	// Test command interpretation
	testCommand := "show current directory"
	fmt.Printf("   🗣️  Testing command: '%s'\n", testCommand)
	
	// Set safe mode for demonstration
	os.Setenv("ECHOCHAT_SAFE", "true")
	
	cmdCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err = echoChat.ProcessInput(cmdCtx, testCommand)
	if err != nil {
		fmt.Printf("   ⚠️  Command processing: %v\n", err)
	} else {
		fmt.Printf("   ✅ Command processed successfully\n")
	}

	// Show status
	status := engine.GetDeepTreeEchoStatus()
	if health, ok := status["system_health"].(string); ok {
		fmt.Printf("   🏥 System Health: %s\n", health)
	}

	return nil
}

func testOfflineDemo() error {
	fmt.Printf("   🏗️  Testing EchoChat structure (offline)...\n")
	
	engine := orchestration.NewEngine(api.Client{})
	orchestration.RegisterDefaultTools(engine)
	orchestration.RegisterDefaultPlugins(engine)

	// Initialize Deep Tree Echo
	ctx := context.Background()
	err := engine.InitializeDeepTreeEcho(ctx)
	if err != nil {
		fmt.Printf("   ⚠️  Deep Tree Echo initialization: %v\n", err)
	}

	echoChat := orchestration.NewEchoChat(engine)
	if echoChat == nil {
		return fmt.Errorf("failed to create EchoChat")
	}

	status := engine.GetDeepTreeEchoStatus()
	fmt.Printf("   📊 System initialized with status: %s\n", status["system_health"])

	fmt.Printf("   ✅ Offline structure validation completed\n")
	return nil
}