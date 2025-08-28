package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ollama/ollama/api"
	"github.com/ollama/ollama/orchestration"
)

// EchoselfIntegrationDemo demonstrates the enhanced orchestration capabilities
// that represent progress toward full echoself integration
func main() {
	fmt.Println("🤖 Echollama Enhanced Orchestration Demo")
	fmt.Println("========================================")
	fmt.Println("Demonstrating echoself integration progress...")
	fmt.Println()

	// Initialize orchestration engine
	client := api.Client{}
	engine := orchestration.NewEngine(client)
	ctx := context.Background()

	// Register default tools and plugins
	orchestration.RegisterDefaultTools(engine)
	orchestration.RegisterDefaultPlugins(engine)

	fmt.Printf("✅ Registered %d tools and %d plugins\n", 
		len(engine.GetAvailableTools()), 
		len(engine.GetAvailablePlugins()))

	// Demonstrate different agent types
	fmt.Println("\n🧠 Creating Specialized Agents...")
	
	// Create a reflective agent
	reflectiveAgent, err := engine.CreateSpecializedAgent(ctx, orchestration.AgentTypeReflective, "analysis")
	if err != nil {
		log.Fatalf("Failed to create reflective agent: %v", err)
	}
	fmt.Printf("✅ Created reflective agent: %s\n", reflectiveAgent.Name)

	// Create an orchestrator agent
	orchestratorAgent, err := engine.CreateSpecializedAgent(ctx, orchestration.AgentTypeOrchestrator, "coordination")
	if err != nil {
		log.Fatalf("Failed to create orchestrator agent: %v", err)
	}
	fmt.Printf("✅ Created orchestrator agent: %s\n", orchestratorAgent.Name)

	// Create a specialist agent
	specialistAgent, err := engine.CreateSpecializedAgent(ctx, orchestration.AgentTypeSpecialist, "coding")
	if err != nil {
		log.Fatalf("Failed to create specialist agent: %v", err)
	}
	fmt.Printf("✅ Created specialist agent: %s\n", specialistAgent.Name)

	// Demonstrate enhanced task execution
	fmt.Println("\n🔧 Demonstrating Enhanced Task Types...")

	// Tool task example
	toolTask := &orchestration.Task{
		ID:      "demo-tool-task",
		Type:    orchestration.TaskTypeTool,
		Input:   "Calculate 42 * 37",
		Status:  orchestration.TaskStatusPending,
		AgentID: specialistAgent.ID,
		Parameters: map[string]interface{}{
			"tool": map[string]interface{}{
				"name": "calculator",
				"parameters": map[string]interface{}{
					"operation": "multiply",
					"a":         42.0,
					"b":         37.0,
				},
			},
		},
	}

	result, err := engine.ExecuteTask(ctx, toolTask, specialistAgent)
	if err != nil {
		log.Printf("Tool task failed: %v", err)
	} else {
		fmt.Printf("✅ Tool task completed: %s\n", result.Output)
	}

	// Plugin task example
	pluginTask := &orchestration.Task{
		ID:      "demo-plugin-task",
		Type:    orchestration.TaskTypePlugin,
		Input:   "This is sample data for analysis with multiple data points and patterns",
		Status:  orchestration.TaskStatusPending,
		AgentID: reflectiveAgent.ID,
		Parameters: map[string]interface{}{
			"plugin_name": "data_analysis",
			"type":        "summary",
		},
	}

	result, err = engine.ExecuteTask(ctx, pluginTask, reflectiveAgent)
	if err != nil {
		log.Printf("Plugin task failed: %v", err)
	} else {
		fmt.Printf("✅ Plugin task completed: %s\n", result.Output)
	}

	// Reflection task example
	reflectTask := &orchestration.Task{
		ID:      "demo-reflect-task",
		Type:    orchestration.TaskTypeReflect,
		Input:   "Analyze my recent performance in data analysis and tool usage tasks",
		Status:  orchestration.TaskStatusPending,
		AgentID: reflectiveAgent.ID,
	}

	result, err = engine.ExecuteTask(ctx, reflectTask, reflectiveAgent)
	if err != nil {
		log.Printf("Reflection task failed: %v", err)
	} else {
		fmt.Printf("✅ Reflection task completed: %s\n", result.Output)
	}

	// Demonstrate agent state management
	fmt.Println("\n🧩 Agent State Management...")
	updatedAgent, err := engine.GetAgent(ctx, reflectiveAgent.ID)
	if err != nil {
		log.Printf("Failed to get agent: %v", err)
	} else {
		fmt.Printf("✅ Agent has %d context items in memory\n", len(updatedAgent.State.Context))
		fmt.Printf("✅ Agent memory contains %d entries\n", len(updatedAgent.State.Memory))
		fmt.Printf("✅ Last interaction: %s\n", updatedAgent.State.LastInteraction.Format("15:04:05"))
	}

	fmt.Println("\n🎯 Echoself Integration Progress Summary:")
	fmt.Println("   ✅ Advanced agent types with specialized behaviors")
	fmt.Println("   ✅ Persistent state management and memory")
	fmt.Println("   ✅ Tool calling capabilities")
	fmt.Println("   ✅ Plugin system for extensibility")
	fmt.Println("   ✅ Self-reflection and learning capabilities")
	fmt.Println("   ✅ Enhanced coordination patterns")
	fmt.Println()
	fmt.Println("🚀 Ready for deeper echoself integration!")
}