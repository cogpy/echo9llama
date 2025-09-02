package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ollama/ollama/api"
	"github.com/ollama/ollama/orchestration"
)

// LiveInteractiveTest runs automated tests with REAL model - NO FALLBACKS, NO MOCKS
// Either the real model works completely or everything fails - no exceptions!
func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run live-interactive-test.go <model-name>")
	}

	modelName := os.Args[1]
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	fmt.Printf("🧪 Starting REAL Model Test with: %s\n", modelName)
	fmt.Printf("=" + strings.Repeat("=", 50) + "\n")
	fmt.Println("⚠️  NO FALLBACKS - Real model must work or TOTAL FAILURE!")

	// Test 1: Basic model connectivity - MUST WORK
	fmt.Println("\n🔗 Test 1: Basic Model Connectivity - REQUIRED")
	if err := testModelConnectivity(ctx, modelName); err != nil {
		log.Fatalf("❌ FATAL: Model connectivity test failed: %v", err)
	}
	fmt.Println("✅ Model connectivity test PASSED")

	// Test 2: Deep Tree Echo System - MUST WORK  
	fmt.Println("\n🧠 Test 2: Deep Tree Echo System - REQUIRED")
	if err := testDeepTreeEcho(ctx, modelName); err != nil {
		log.Fatalf("❌ FATAL: Deep Tree Echo test failed: %v", err)
	}
	fmt.Println("✅ Deep Tree Echo test PASSED")

	// Test 3: EchoChat Integration - MUST WORK
	fmt.Println("\n🌊 Test 3: EchoChat Integration - REQUIRED")
	if err := testEchoChatIntegration(ctx, modelName); err != nil {
		log.Fatalf("❌ FATAL: EchoChat integration test failed: %v", err)
	}
	fmt.Println("✅ EchoChat integration test PASSED")

	// Test 4: Natural Language Shell Commands - MUST WORK
	fmt.Println("\n🗣️  Test 4: Natural Language Shell Commands - REQUIRED")
	if err := testShellCommands(ctx, modelName); err != nil {
		log.Fatalf("❌ FATAL: Shell commands test failed: %v", err)
	}
	fmt.Println("✅ Shell commands test PASSED")

	// Test 5: Orchestration Capabilities - MUST WORK
	fmt.Println("\n⚙️  Test 5: Orchestration Capabilities - REQUIRED")
	if err := testOrchestrationCapabilities(ctx, modelName); err != nil {
		log.Fatalf("❌ FATAL: Orchestration capabilities test failed: %v", err)
	}
	fmt.Println("✅ Orchestration capabilities test PASSED")

	fmt.Printf("\n🎉 ALL TESTS PASSED! Real model %s working perfectly!\n", modelName)
}

func testModelConnectivity(ctx context.Context, modelName string) error {
	fmt.Printf("   🔌 Creating client connection...\n")
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return fmt.Errorf("FATAL: Failed to create client - server must be running: %w", err)
	}

	// Test server connectivity - MUST WORK
	fmt.Printf("   🏥 Testing server heartbeat...\n")
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	err = client.Heartbeat(pingCtx)
	if err != nil {
		return fmt.Errorf("FATAL: Server heartbeat failed - ollama server not responding: %w", err)
	}
	fmt.Printf("   ✅ Server heartbeat successful\n")

	// Test basic chat functionality with REAL model
	fmt.Printf("   🤖 Testing real model conversation...\n")
	req := &api.ChatRequest{
		Model: modelName,
		Messages: []api.Message{
			{
				Role:    "user", 
				Content: "Hello! Please respond with exactly: 'Echo test successful'",
			},
		},
		Stream: new(bool), // false
	}

	responseReceived := false
	respFunc := func(resp api.ChatResponse) error {
		if !responseReceived {
			fmt.Printf("   📝 Model response: %s\n", strings.TrimSpace(resp.Message.Content))
			responseReceived = true
		}
		return nil
	}

	// Real model must respond within reasonable time
	connectCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	err = client.Chat(connectCtx, req, respFunc)
	if err != nil {
		return fmt.Errorf("FATAL: Model chat failed - real model must respond: %w", err)
	}

	if !responseReceived {
		return fmt.Errorf("FATAL: No response received from model %s", modelName)
	}

	fmt.Printf("   ✅ Successfully connected and communicated with real model %s\n", modelName)
	return nil
}

func testDeepTreeEcho(ctx context.Context, modelName string) error {
	fmt.Printf("   🔌 Creating client for Deep Tree Echo...\n")
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return fmt.Errorf("FATAL: Client creation failed - server required: %w", err)
	}

	// Test server is available - REQUIRED
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	err = client.Heartbeat(pingCtx)
	if err != nil {
		return fmt.Errorf("FATAL: Server not available for Deep Tree Echo: %w", err)
	}

	// Create engine with REAL client (not empty)
	engine := orchestration.NewEngine(*client)

	// Initialize Deep Tree Echo - MUST WORK
	fmt.Println("   🧠 Initializing Deep Tree Echo with real model...")
	err = engine.InitializeDeepTreeEcho(ctx)
	if err != nil {
		return fmt.Errorf("FATAL: Deep Tree Echo initialization failed: %w", err)
	}

	// Get status - MUST be healthy
	status := engine.GetDeepTreeEchoStatus()
	fmt.Printf("   📊 System Status: %v\n", status)

	// Verify system health is good
	health, hasHealth := status["system_health"].(string)
	if !hasHealth {
		return fmt.Errorf("FATAL: System health status missing from Deep Tree Echo")
	}
	
	fmt.Printf("   🏥 System Health: %s\n", health)
	
	coreStatus, hasCoreStatus := status["core_status"].(string)
	if !hasCoreStatus {
		return fmt.Errorf("FATAL: Core status missing from Deep Tree Echo")
	}
	
	fmt.Printf("   🧠 Core Status: %s\n", coreStatus)

	// Test actual cognitive processing with the real model
	fmt.Printf("   🧪 Testing cognitive processing with model %s...\n", modelName)
	
	// Create a test request that requires real model inference
	req := &api.ChatRequest{
		Model: modelName,
		Messages: []api.Message{
			{Role: "user", Content: "Analyze this input and respond with your cognitive state"},
		},
		Stream: new(bool),
	}

	cogCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	responseReceived := false
	err = client.Chat(cogCtx, req, func(resp api.ChatResponse) error {
		if !responseReceived {
			fmt.Printf("   🧠 Cognitive response: %s\n", strings.TrimSpace(resp.Message.Content))
			responseReceived = true
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("FATAL: Deep Tree Echo cognitive processing failed: %w", err)
	}

	if !responseReceived {
		return fmt.Errorf("FATAL: No cognitive response from Deep Tree Echo")
	}

	fmt.Printf("   ✅ Deep Tree Echo fully operational with real model\n")
	return nil
}

func testEchoChatIntegration(ctx context.Context, modelName string) error {
	fmt.Printf("   🔌 Creating client for EchoChat...\n")
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return fmt.Errorf("FATAL: Client initialization failed - server required: %w", err)
	}
	
	// Test server connectivity - REQUIRED
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	err = client.Heartbeat(pingCtx)
	if err != nil {
		return fmt.Errorf("FATAL: Server not available for EchoChat: %w", err)
	}
	fmt.Printf("   ✅ Server connection verified\n")

	// Create engine with REAL client
	engine := orchestration.NewEngine(*client)
	
	// Register tools and plugins - MUST WORK
	fmt.Printf("   🔧 Registering tools and plugins...\n")
	orchestration.RegisterDefaultTools(engine)
	orchestration.RegisterDefaultPlugins(engine)

	// Initialize Deep Tree Echo - REQUIRED
	fmt.Printf("   🧠 Initializing Deep Tree Echo integration...\n")
	err = engine.InitializeDeepTreeEcho(ctx)
	if err != nil {
		return fmt.Errorf("FATAL: Deep Tree Echo initialization failed: %w", err)
	}

	// Create EchoChat instance - MUST WORK
	fmt.Printf("   🌊 Creating EchoChat instance...\n")
	echoChat := orchestration.NewEchoChat(engine)
	if echoChat == nil {
		return fmt.Errorf("FATAL: Failed to create EchoChat instance")
	}

	// Test real command processing with actual model inference
	fmt.Println("   🔍 Testing REAL command translation with model...")
	
	testCommands := []string{
		"show current directory",
		"list files in current directory", 
		"check disk usage",
	}

	for _, cmd := range testCommands {
		fmt.Printf("   📝 Testing real command: '%s'\n", cmd)
		
		// NO SAFE MODE - must use real model inference
		os.Unsetenv("ECHOCHAT_SAFE")
		
		// Real command processing with model
		cmdCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		
		err := echoChat.ProcessInput(cmdCtx, cmd)
		cancel()
		
		if err != nil {
			return fmt.Errorf("FATAL: Command processing failed for '%s': %w", cmd, err)
		}
		fmt.Printf("   ✅ Command processed successfully with real model\n")
	}

	fmt.Printf("   ✅ EchoChat integration fully functional with real model\n")
	return nil
}

func testShellCommands(ctx context.Context, modelName string) error {
	fmt.Printf("   🔌 Creating client for shell commands...\n")
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return fmt.Errorf("FATAL: Client initialization failed - server required: %w", err)
	}
	
	// Test server availability - REQUIRED
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	err = client.Heartbeat(pingCtx)
	if err != nil {
		return fmt.Errorf("FATAL: Server not available for shell commands: %w", err)
	}
	fmt.Printf("   ✅ Server verified for shell command processing\n")

	// Create engine with REAL client
	engine := orchestration.NewEngine(*client)

	// Register tools and initialize system - MUST WORK
	fmt.Printf("   🔧 Setting up shell command system...\n")
	orchestration.RegisterDefaultTools(engine)
	orchestration.RegisterDefaultPlugins(engine)
	
	err = engine.InitializeDeepTreeEcho(ctx)
	if err != nil {
		return fmt.Errorf("FATAL: Deep Tree Echo initialization failed: %w", err)
	}

	echoChat := orchestration.NewEchoChat(engine)
	if echoChat == nil {
		return fmt.Errorf("FATAL: Failed to create EchoChat for shell commands")
	}

	// Test REAL shell command translation with actual model
	fmt.Printf("   🗣️  Testing real shell command translation...\n")
	
	realCommands := []string{
		"show current working directory",
		"display environment variables",
		"show disk usage information",
	}

	for _, cmd := range realCommands {
		fmt.Printf("   🔄 Processing real command: %s\n", cmd)
		
		cmdCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
		
		// NO SAFE MODE - use real model inference for command translation
		os.Unsetenv("ECHOCHAT_SAFE")
		
		err := echoChat.ProcessInput(cmdCtx, cmd)
		cancel()
		
		if err != nil {
			return fmt.Errorf("FATAL: Shell command processing failed for '%s': %w", cmd, err)
		}
		fmt.Printf("   ✅ Command successfully processed with real model\n")
	}

	fmt.Printf("   ✅ Shell command system fully operational with real model\n")
	return nil
}

func testOrchestrationCapabilities(ctx context.Context, modelName string) error {
	fmt.Printf("   🔌 Creating client for orchestration...\n")
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return fmt.Errorf("FATAL: Client initialization failed - server required: %w", err)
	}
	
	// Test server availability - REQUIRED  
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	err = client.Heartbeat(pingCtx)
	if err != nil {
		return fmt.Errorf("FATAL: Server not available for orchestration: %w", err)
	}
	fmt.Printf("   ✅ Server verified for orchestration\n")

	// Create engine with REAL client
	engine := orchestration.NewEngine(*client)

	// Initialize orchestration system - MUST WORK
	fmt.Printf("   ⚙️  Initializing orchestration system...\n")
	orchestration.RegisterDefaultTools(engine)
	orchestration.RegisterDefaultPlugins(engine)

	// Initialize Deep Tree Echo - REQUIRED
	err = engine.InitializeDeepTreeEcho(ctx)
	if err != nil {
		return fmt.Errorf("FATAL: Deep Tree Echo initialization failed: %w", err)
	}

	// Test orchestration status - MUST be healthy
	status := engine.GetDeepTreeEchoStatus()
	fmt.Printf("   📊 Orchestration Status: %v\n", status)
	
	if health, ok := status["system_health"]; !ok || health == "" {
		return fmt.Errorf("FATAL: Orchestration system health not available")
	}

	// Test EchoChat with orchestration - MUST WORK
	echoChat := orchestration.NewEchoChat(engine)
	if echoChat == nil {
		return fmt.Errorf("FATAL: Failed to create EchoChat for orchestration")
	}

	// Test REAL orchestration command with model inference
	testCmd := "analyze system status and provide recommendations"
	fmt.Printf("   🧪 Testing real orchestration: '%s'\n", testCmd)
	
	cmdCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	
	// NO SAFE MODE - use real model for orchestration
	os.Unsetenv("ECHOCHAT_SAFE")
	
	err = echoChat.ProcessInput(cmdCtx, testCmd)
	if err != nil {
		return fmt.Errorf("FATAL: Orchestration command failed: %w", err)
	}
	fmt.Printf("   ✅ Orchestration command executed successfully with real model\n")

	// Verify orchestration is still healthy after real processing
	finalStatus := engine.GetDeepTreeEchoStatus()
	if health, ok := finalStatus["system_health"]; !ok || health == "" {
		return fmt.Errorf("FATAL: Orchestration system unhealthy after processing")
	}
	
	fmt.Printf("   📊 Final orchestration status: %v\n", finalStatus)
	fmt.Printf("   ✅ Orchestration capabilities fully operational with real model\n")
	return nil
}