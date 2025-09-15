package main

import (
        "bytes"
        "encoding/json"
        "fmt"
        "io"
        "log"
        "net/http"
        "os"
        "os/exec"
        "strings"
        "time"
)

// WorkflowOrchestrator manages complex workflow orchestration
type WorkflowOrchestrator struct {
        ServerURL    string
        ServerPort   string
        ServerCmd    string
        TestResults  []TestResult
        Capabilities map[string]bool
}

// TestResult captures test execution results
type TestResult struct {
        TestName    string
        Success     bool
        Duration    time.Duration
        Response    string
        Error       error
        Timestamp   time.Time
}

// ServerHealth represents server health status
type ServerHealth struct {
        Status     string
        Coherence  float64
        Patterns   int
        Memories   int
        Active     bool
}

// NewOrchestrator creates a new workflow orchestrator
func NewOrchestrator() *WorkflowOrchestrator {
        return &WorkflowOrchestrator{
                ServerURL:    "http://localhost",
                ServerPort:   "5000",
                ServerCmd:    "server/simple/embodied_server",
                TestResults:  make([]TestResult, 0),
                Capabilities: make(map[string]bool),
        }
}

// CheckServerHealth verifies server is responding
func (w *WorkflowOrchestrator) CheckServerHealth() (*ServerHealth, error) {
        startTime := time.Now()
        
        resp, err := http.Get(fmt.Sprintf("%s:%s/", w.ServerURL, w.ServerPort))
        if err != nil {
                w.recordTest("Server Health Check", false, time.Since(startTime), "", err)
                return nil, err
        }
        defer resp.Body.Close()
        
        var health ServerHealth
        body, _ := io.ReadAll(resp.Body)
        
        if err := json.Unmarshal(body, &health); err != nil {
                w.recordTest("Server Health Check", false, time.Since(startTime), string(body), err)
                return nil, err
        }
        
        health.Active = true
        w.recordTest("Server Health Check", true, time.Since(startTime), fmt.Sprintf("Status: %s", health.Status), nil)
        
        return &health, nil
}

// StartServer starts the Deep Tree Echo server
func (w *WorkflowOrchestrator) StartServer() error {
        log.Println("🚀 Starting Deep Tree Echo server...")
        
        // Check if already running
        if health, err := w.CheckServerHealth(); err == nil && health.Active {
                log.Println("✅ Server already running")
                return nil
        }
        
        // Start server in background
        cmd := exec.Command(w.ServerCmd)
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        
        if err := cmd.Start(); err != nil {
                return fmt.Errorf("failed to start server: %v", err)
        }
        
        // Wait for server to be ready
        for i := 0; i < 30; i++ {
                time.Sleep(1 * time.Second)
                if health, err := w.CheckServerHealth(); err == nil && health.Active {
                        log.Println("✅ Server started successfully")
                        return nil
                }
        }
        
        return fmt.Errorf("server failed to start within 30 seconds")
}

// TestEchoThink tests the Deep Tree Echo thinking endpoint
func (w *WorkflowOrchestrator) TestEchoThink(prompt string) (string, error) {
        startTime := time.Now()
        
        payload := map[string]string{"prompt": prompt}
        jsonData, _ := json.Marshal(payload)
        
        resp, err := http.Post(
                fmt.Sprintf("%s:%s/api/echo/think", w.ServerURL, w.ServerPort),
                "application/json",
                bytes.NewBuffer(jsonData),
        )
        if err != nil {
                w.recordTest("Echo Think", false, time.Since(startTime), "", err)
                return "", err
        }
        defer resp.Body.Close()
        
        body, _ := io.ReadAll(resp.Body)
        var result map[string]interface{}
        json.Unmarshal(body, &result)
        
        thought := fmt.Sprintf("%v", result["thought"])
        w.recordTest("Echo Think", true, time.Since(startTime), thought, nil)
        
        return thought, nil
}

// RunModel generates text using the model API
func (w *WorkflowOrchestrator) RunModel(prompt string) (string, error) {
        startTime := time.Now()
        
        payload := map[string]string{
                "model":  "deep-tree-echo",
                "prompt": prompt,
        }
        jsonData, _ := json.Marshal(payload)
        
        resp, err := http.Post(
                fmt.Sprintf("%s:%s/api/generate", w.ServerURL, w.ServerPort),
                "application/json",
                bytes.NewBuffer(jsonData),
        )
        if err != nil {
                w.recordTest("Model Generation", false, time.Since(startTime), "", err)
                return "", err
        }
        defer resp.Body.Close()
        
        body, _ := io.ReadAll(resp.Body)
        var result map[string]interface{}
        json.Unmarshal(body, &result)
        
        response := fmt.Sprintf("%v", result["response"])
        w.recordTest("Model Generation", true, time.Since(startTime), response, nil)
        
        return response, nil
}

// ExecuteChatSession runs an automated chat session
func (w *WorkflowOrchestrator) ExecuteChatSession(messages []string) ([]string, error) {
        responses := make([]string, 0)
        
        for i, msg := range messages {
                startTime := time.Now()
                
                // Build chat messages
                chatMessages := []map[string]string{
                        {"role": "user", "content": msg},
                }
                
                payload := map[string]interface{}{
                        "messages": chatMessages,
                }
                jsonData, _ := json.Marshal(payload)
                
                resp, err := http.Post(
                        fmt.Sprintf("%s:%s/api/chat", w.ServerURL, w.ServerPort),
                        "application/json",
                        bytes.NewBuffer(jsonData),
                )
                if err != nil {
                        w.recordTest(fmt.Sprintf("Chat Message %d", i+1), false, time.Since(startTime), "", err)
                        return responses, err
                }
                defer resp.Body.Close()
                
                body, _ := io.ReadAll(resp.Body)
                var result map[string]interface{}
                json.Unmarshal(body, &result)
                
                if message, ok := result["message"].(map[string]interface{}); ok {
                        content := fmt.Sprintf("%v", message["content"])
                        responses = append(responses, content)
                        w.recordTest(fmt.Sprintf("Chat Message %d", i+1), true, time.Since(startTime), content, nil)
                }
                
                // Small delay between messages
                time.Sleep(500 * time.Millisecond)
        }
        
        return responses, nil
}

// TestMemoryOperations tests memory store and recall
func (w *WorkflowOrchestrator) TestMemoryOperations() error {
        // Store memory
        startTime := time.Now()
        
        payload := map[string]interface{}{
                "key":   "orchestration_test",
                "value": "Testing memory at " + time.Now().Format(time.RFC3339),
        }
        jsonData, _ := json.Marshal(payload)
        
        resp, err := http.Post(
                fmt.Sprintf("%s:%s/api/echo/remember", w.ServerURL, w.ServerPort),
                "application/json",
                bytes.NewBuffer(jsonData),
        )
        if err != nil {
                w.recordTest("Memory Store", false, time.Since(startTime), "", err)
                return err
        }
        resp.Body.Close()
        w.recordTest("Memory Store", true, time.Since(startTime), "Memory stored", nil)
        
        // Recall memory
        startTime = time.Now()
        resp, err = http.Get(fmt.Sprintf("%s:%s/api/echo/recall/orchestration_test", w.ServerURL, w.ServerPort))
        if err != nil {
                w.recordTest("Memory Recall", false, time.Since(startTime), "", err)
                return err
        }
        defer resp.Body.Close()
        
        body, _ := io.ReadAll(resp.Body)
        w.recordTest("Memory Recall", true, time.Since(startTime), string(body), nil)
        
        return nil
}

// TestEmotionalProcessing tests emotional state changes
func (w *WorkflowOrchestrator) TestEmotionalProcessing() error {
        emotions := []string{"curious", "excited", "calm", "focused"}
        
        for _, emotion := range emotions {
                startTime := time.Now()
                
                payload := map[string]interface{}{
                        "emotion":   emotion,
                        "intensity": 0.8,
                }
                jsonData, _ := json.Marshal(payload)
                
                resp, err := http.Post(
                        fmt.Sprintf("%s:%s/api/echo/feel", w.ServerURL, w.ServerPort),
                        "application/json",
                        bytes.NewBuffer(jsonData),
                )
                if err != nil {
                        w.recordTest(fmt.Sprintf("Emotion: %s", emotion), false, time.Since(startTime), "", err)
                        continue
                }
                resp.Body.Close()
                
                w.recordTest(fmt.Sprintf("Emotion: %s", emotion), true, time.Since(startTime), emotion, nil)
                time.Sleep(200 * time.Millisecond)
        }
        
        return nil
}

// TestSpatialMovement tests spatial awareness
func (w *WorkflowOrchestrator) TestSpatialMovement() error {
        positions := [][]float64{
                {10, 5, 3},
                {-5, 15, 7},
                {0, 0, 0},
        }
        
        for i, pos := range positions {
                startTime := time.Now()
                
                payload := map[string]float64{
                        "x": pos[0],
                        "y": pos[1],
                        "z": pos[2],
                }
                jsonData, _ := json.Marshal(payload)
                
                resp, err := http.Post(
                        fmt.Sprintf("%s:%s/api/echo/move", w.ServerURL, w.ServerPort),
                        "application/json",
                        bytes.NewBuffer(jsonData),
                )
                if err != nil {
                        w.recordTest(fmt.Sprintf("Move %d", i+1), false, time.Since(startTime), "", err)
                        continue
                }
                resp.Body.Close()
                
                w.recordTest(fmt.Sprintf("Move %d", i+1), true, time.Since(startTime), 
                        fmt.Sprintf("Position: (%.1f, %.1f, %.1f)", pos[0], pos[1], pos[2]), nil)
                time.Sleep(200 * time.Millisecond)
        }
        
        return nil
}

// RunComprehensiveTest executes all tests
func (w *WorkflowOrchestrator) RunComprehensiveTest() {
        log.Println("🔧 Starting Comprehensive Workflow Orchestration Test")
        log.Println("=" + strings.Repeat("=", 50))
        
        // Phase 1: Server Health
        log.Println("\n📊 Phase 1: Server Health Check")
        health, err := w.CheckServerHealth()
        if err != nil {
                log.Printf("❌ Server not responding: %v", err)
                log.Println("Attempting to start server...")
                if err := w.StartServer(); err != nil {
                        log.Printf("❌ Failed to start server: %v", err)
                        return
                }
                health, _ = w.CheckServerHealth()
        }
        
        if health != nil {
                log.Printf("✅ Server Status: %s", health.Status)
                log.Printf("   Coherence: %.2f", health.Coherence)
                log.Printf("   Patterns: %d", health.Patterns)
                log.Printf("   Memories: %d", health.Memories)
        }
        
        // Phase 2: Cognitive Processing
        log.Println("\n🧠 Phase 2: Cognitive Processing Tests")
        
        prompts := []string{
                "What is consciousness?",
                "How do patterns emerge?",
                "Describe resonance",
        }
        
        for _, prompt := range prompts {
                thought, err := w.TestEchoThink(prompt)
                if err == nil {
                        log.Printf("✅ Think: '%s' -> %s", prompt, truncate(thought, 50))
                }
        }
        
        // Phase 3: Model Generation
        log.Println("\n🤖 Phase 3: Model Generation Tests")
        
        for _, prompt := range prompts {
                response, err := w.RunModel(prompt)
                if err == nil {
                        log.Printf("✅ Generate: '%s' -> %s", prompt, truncate(response, 50))
                }
        }
        
        // Phase 4: Chat Session
        log.Println("\n💬 Phase 4: Automated Chat Session")
        
        chatMessages := []string{
                "Hello Deep Tree Echo",
                "Tell me about your cognitive architecture",
                "What patterns do you observe?",
        }
        
        responses, err := w.ExecuteChatSession(chatMessages)
        if err == nil {
                for i, resp := range responses {
                        log.Printf("✅ Chat %d: %s", i+1, truncate(resp, 60))
                }
        }
        
        // Phase 5: Memory Operations
        log.Println("\n💾 Phase 5: Memory Operations")
        w.TestMemoryOperations()
        
        // Phase 6: Emotional Processing
        log.Println("\n❤️ Phase 6: Emotional Processing")
        w.TestEmotionalProcessing()
        
        // Phase 7: Spatial Movement
        log.Println("\n📍 Phase 7: Spatial Awareness")
        w.TestSpatialMovement()
        
        // Generate Report
        w.GenerateReport()
}

// GenerateReport creates a summary of all tests
func (w *WorkflowOrchestrator) GenerateReport() {
        log.Println("\n" + strings.Repeat("=", 60))
        log.Println("📋 ORCHESTRATION TEST REPORT")
        log.Println(strings.Repeat("=", 60))
        
        totalTests := len(w.TestResults)
        successCount := 0
        totalDuration := time.Duration(0)
        
        for _, result := range w.TestResults {
                if result.Success {
                        successCount++
                }
                totalDuration += result.Duration
        }
        
        successRate := float64(successCount) / float64(totalTests) * 100
        
        log.Printf("Total Tests: %d", totalTests)
        log.Printf("Successful: %d", successCount)
        log.Printf("Failed: %d", totalTests-successCount)
        log.Printf("Success Rate: %.1f%%", successRate)
        log.Printf("Total Duration: %v", totalDuration)
        log.Printf("Average Response Time: %v", totalDuration/time.Duration(totalTests))
        
        // Determine capabilities
        w.Capabilities["server_health"] = w.hasSuccessfulTest("Server Health Check")
        w.Capabilities["cognitive_processing"] = w.hasSuccessfulTest("Echo Think")
        w.Capabilities["model_generation"] = w.hasSuccessfulTest("Model Generation")
        w.Capabilities["chat_session"] = w.hasSuccessfulTest("Chat Message")
        w.Capabilities["memory_operations"] = w.hasSuccessfulTest("Memory")
        w.Capabilities["emotional_processing"] = w.hasSuccessfulTest("Emotion")
        w.Capabilities["spatial_awareness"] = w.hasSuccessfulTest("Move")
        
        log.Println("\n🎯 Detected Capabilities:")
        for capability, enabled := range w.Capabilities {
                status := "❌"
                if enabled {
                        status = "✅"
                }
                log.Printf("%s %s", status, capability)
        }
        
        if successRate >= 80 {
                log.Println("\n🎉 ORCHESTRATION TEST PASSED!")
                log.Println("System is fully operational and can orchestrate complex workflows")
        } else if successRate >= 50 {
                log.Println("\n⚠️ ORCHESTRATION TEST PARTIALLY PASSED")
                log.Println("Some capabilities need attention")
        } else {
                log.Println("\n❌ ORCHESTRATION TEST FAILED")
                log.Println("System requires troubleshooting")
        }
}

// Helper functions

func (w *WorkflowOrchestrator) recordTest(name string, success bool, duration time.Duration, response string, err error) {
        result := TestResult{
                TestName:  name,
                Success:   success,
                Duration:  duration,
                Response:  response,
                Error:     err,
                Timestamp: time.Now(),
        }
        w.TestResults = append(w.TestResults, result)
}

func (w *WorkflowOrchestrator) hasSuccessfulTest(prefix string) bool {
        for _, result := range w.TestResults {
                if strings.HasPrefix(result.TestName, prefix) && result.Success {
                        return true
                }
        }
        return false
}

func truncate(s string, maxLen int) string {
        if len(s) <= maxLen {
                return s
        }
        return s[:maxLen] + "..."
}

func main() {
        if len(os.Args) > 1 && os.Args[1] == "tasks" {
                RunTaskAutomation()
        } else {
                orchestrator := NewOrchestrator()
                orchestrator.RunComprehensiveTest()
        }
}