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

// SimpleDemo demonstrates basic echollama functionality with REAL model - NO FALLBACKS
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run simple-demo.go <model-name>")
		os.Exit(1)
	}

	modelName := os.Args[1]
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Printf("🌊 EchoLlama REAL Model Demo with %s\n", modelName)
	fmt.Printf("=" + strings.Repeat("=", 40) + "\n")
	fmt.Println("⚠️  NO FALLBACKS - Real model must work or TOTAL FAILURE!")

	// Test 1: Basic model chat - MUST WORK
	fmt.Println("\n💬 Testing basic real model conversation...")
	if err := testBasicChat(ctx, modelName); err != nil {
		fmt.Printf("❌ FATAL: Basic chat test failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Basic chat test successful")

	// Test 2: EchoChat system - MUST WORK
	fmt.Println("\n🧠 Testing EchoChat integration...")
	if err := testEchoChatDemo(ctx, modelName); err != nil {
		fmt.Printf("❌ FATAL: EchoChat test failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ EchoChat test successful")

	fmt.Printf("\n🎉 Demo completed successfully for real model: %s\n", modelName)
}

func testBasicChat(ctx context.Context, modelName string) error {
	fmt.Printf("   🔌 Creating client connection...\n")
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return fmt.Errorf("FATAL: Client initialization failed - server must be running: %w", err)
	}

	// Test server connectivity - MUST WORK
	fmt.Printf("   🏥 Testing server connection...\n")
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	err = client.Heartbeat(pingCtx)
	if err != nil {
		return fmt.Errorf("FATAL: Server not responding: %w", err)
	}

	// Real model conversation test - MUST WORK
	req := &api.ChatRequest{
		Model: modelName,
		Messages: []api.Message{
			{Role: "user", Content: "Say hello and tell me you're working correctly!"},
		},
		Stream: new(bool),
	}

	fmt.Printf("   🤖 Asking %s to respond...\n", modelName)
	
	connectCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
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
		return fmt.Errorf("FATAL: Chat with real model failed: %w", err)
	}

	if !responseReceived {
		return fmt.Errorf("FATAL: No response received from real model %s", modelName)
	}

	return nil
}

func testEchoChatDemo(ctx context.Context, modelName string) error {
	fmt.Printf("   🔌 Creating client for EchoChat demo...\n")
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return fmt.Errorf("FATAL: Client initialization failed - server required: %w", err)
	}

	// Test server connectivity - REQUIRED
	testCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	err = client.Heartbeat(testCtx)
	if err != nil {
		return fmt.Errorf("FATAL: Server not responding for EchoChat: %w", err)
	}
	fmt.Printf("   ✅ Server connection verified\n")

	// Create REAL EchoChat with real model
	fmt.Printf("   🌊 Creating EchoChat with Deep Tree Echo...\n")
	
	engine := orchestration.NewEngine(*client)
	orchestration.RegisterDefaultTools(engine)
	orchestration.RegisterDefaultPlugins(engine)

	// Initialize Deep Tree Echo - MUST WORK
	err = engine.InitializeDeepTreeEcho(ctx)
	if err != nil {
		return fmt.Errorf("FATAL: Deep Tree Echo initialization failed: %w", err)
	}

	echoChat := orchestration.NewEchoChat(engine)
	if echoChat == nil {
		return fmt.Errorf("FATAL: Failed to create EchoChat instance")
	}

	// Test REAL command interpretation with model
	testCommand := "show current directory"
	fmt.Printf("   🗣️  Testing real command: '%s'\n", testCommand)
	
	// NO SAFE MODE - use real model inference
	os.Unsetenv("ECHOCHAT_SAFE")
	
	cmdCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	err = echoChat.ProcessInput(cmdCtx, testCommand)
	if err != nil {
		return fmt.Errorf("FATAL: Command processing failed: %w", err)
	}
	fmt.Printf("   ✅ Command processed successfully with real model\n")

	// Verify system status is healthy
	status := engine.GetDeepTreeEchoStatus()
	if health, ok := status["system_health"].(string); ok {
		fmt.Printf("   🏥 System Health: %s\n", health)
		if health == "" {
			return fmt.Errorf("FATAL: System health is empty")
		}
	} else {
		return fmt.Errorf("FATAL: System health status not available")
	}

	return nil
}

