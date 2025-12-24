# LangchainGo and Ergo Integration Plan for Unified Cognitive Loop

**Author:** Manus AI
**Date:** December 24, 2025

## 1. Introduction

This document outlines a detailed implementation plan for integrating the `langchaingo` and `ergo` libraries into the `echo9llama` project, specifically targeting the `UnifiedCognitiveLoopV2`. The goal is to enhance the cognitive architecture with advanced reasoning capabilities and a more robust concurrency model, paving the way for a more sophisticated and scalable Deep Tree Echo AGI.

This plan provides a functional-level guide for developers, detailing the necessary code modifications, new components, and the step-by-step process for a successful integration.

## 2. High-Level Architecture

The integration will follow the architecture defined in the `INTEGRATION_ARCHITECTURE.png` diagram. The core idea is to introduce two new layers of abstraction within the `UnifiedCognitiveLoopV2`:

*   A **Reasoning Layer** powered by `langchaingo` to handle complex, multi-step thought processes and goal-oriented planning.
*   An **Actor-based Concurrency Model** powered by `ergo` to manage the concurrent execution of the three cognitive streams, replacing the current Go channel-based implementation.

These integrations will directly enhance the `UnifiedCognitiveLoopV2` and the `ThreeStreamCognitiveLoop`.

## 3. `langchaingo` Integration: The Reasoning & Planning Engine

The integration of `langchaingo` will provide the `UnifiedCognitiveLoopV2` with a powerful reasoning and planning engine. This will be achieved by creating a `langchaingo` agent that can use the existing `deeptreeecho` subsystems as tools.

### 3.1. New Component: `ReasoningManager`

A new component, `ReasoningManager`, will be created to encapsulate the `langchaingo` agent and its executor.

**File:** `core/deeptreeecho/reasoning_manager.go`

```go
package deeptreeecho

import (
	"context"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/tools"
)

// ReasoningManager orchestrates langchaingo agents for complex reasoning.
type ReasoningManager struct {
	agentExecutor *agents.Executor
}

// NewReasoningManager creates a new reasoning manager.
func NewReasoningManager(llm *openai.Chat, availableTools []tools.Tool) (*ReasoningManager, error) {
	// Initialize the langchaingo agent (e.g., conversational agent)
	agent := agents.NewConversationalAgent(llm, availableTools)
	executor := agents.NewExecutor(agent, availableTools...)

	return &ReasoningManager{
		agentExecutor: &executor,
	}, nil
}

// Reason performs a reasoning task using the langchaingo agent.
func (rm *ReasoningManager) Reason(ctx context.Context, input string) (string, error) {
	response, err := rm.agentExecutor.Call(ctx, map[string]any{"input": input})
	if err != nil {
		return "", err
	}

	return response["output"].(string), nil
}
```

### 3.2. Defining Tools for the Agent

We will create a set of `tools` that the `langchaingo` agent can use. These tools will be wrappers around the existing `deeptreeecho` subsystems.

**File:** `core/deeptreeecho/langchain_tools.go`

```go
package deeptreeecho

import (
	"context"
	"github.com/tmc/langchaingo/tools"
)

// NewSkillLearningTool creates a tool for interacting with the SkillLearningSystem.
func NewSkillLearningTool(sls *SkillLearningSystem) tools.Tool {
	return tools.NewSimpleTool(
		"SkillLearner",
		"Use this tool to learn new skills or query existing ones.",
		func(ctx context.Context, input string) (string, error) {
			// Logic to interact with sls based on input
			return sls.GetSkillLevel(input), nil
		},
	)
}

// NewDiscussionTool creates a tool for the DiscussionAutonomySystem.
func NewDiscussionTool(das *DiscussionAutonomySystem) tools.Tool {
	return tools.NewSimpleTool(
		"DiscussionManager",
		"Use this tool to start or participate in discussions.",
		func(ctx context.Context, input string) (string, error) {
			// Logic to interact with das based on input
			return das.ConsiderStartingDiscussion(input, 1.0)
		},
	)
}

// ... other tools for other subsystems ...
```

### 3.3. Integrating `ReasoningManager` into `UnifiedCognitiveLoopV2`

The `UnifiedCognitiveLoopV2` will be updated to include and use the `ReasoningManager`.

**File:** `core/deeptreeecho/unified_cognitive_loop_v2.go`

1.  **Add `ReasoningManager` to the struct:**

    ```go
    type UnifiedCognitiveLoopV2 struct {
        // ... existing fields ...
        reasoningManager *ReasoningManager
    }
    ```

2.  **Initialize `ReasoningManager` in the constructor:**

    ```go
    func NewUnifiedCognitiveLoopV2(llmProvider llm.LLMProvider) *UnifiedCognitiveLoopV2 {
        // ... existing initializations ...

        // Create tools
        availableTools := []tools.Tool{
            NewSkillLearningTool(ucl.skillLearning),
            NewDiscussionTool(ucl.discussionAutonomy),
            // ... add other tools
        }

        // Create reasoning manager
        rm, err := NewReasoningManager(llmProvider, availableTools)
        if err != nil {
            // handle error
        }
        ucl.reasoningManager = rm

        return ucl
    }
    ```

3.  **Use `ReasoningManager` in the `cognitiveStep`:**

    The `cognitiveStep` function will be modified to delegate complex tasks to the `ReasoningManager`.

    ```go
    func (ucl *UnifiedCognitiveLoopV2) cognitiveStep() {
        // ... existing logic ...

        // Example: If a high-priority thought requires complex planning
        if thought.Importance > 0.8 && requiresPlanning(thought) {
            go func() {
                plan, err := ucl.reasoningManager.Reason(ucl.ctx, "Create a plan for: "+thought.Content)
                if err == nil {
                    // Execute the plan
                }
            }()
        }
    }
    ```

## 4. `ergo` Integration: The Actor-based Concurrency Model

The integration of `ergo` will refactor the `ThreeStreamCognitiveLoop` to use the actor model, making the concurrency more robust and manageable.

### 4.1. Defining the Cognitive Stream Actor

We will define an `ergo` actor to represent a single cognitive stream.

**File:** `core/deeptreeecho/cognitive_stream_actor.go`

```go
package deeptreeecho

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type CognitiveStreamActor struct {
	act.Actor
	streamType string // e.g., "perception", "action", "simulation"
}

func (csa *CognitiveStreamActor) Init(args ...any) error {
	csa.streamType = args[0].(string)
	return nil
}

func (csa *CognitiveStreamActor) HandleMessage(from gen.PID, message any) error {
	// Process cognitive tasks for this stream
	// ...
	return nil
}

// ... other actor callbacks ...
```

### 4.2. Refactoring `ThreeStreamCognitiveLoop`

The `ThreeStreamCognitiveLoop` will be refactored to spawn and manage these actors.

**File:** `core/deeptreeecho/three_stream_cognitive_loop.go`

```go
package deeptreeecho

import (
	"ergo.services/ergo/node"
)

type ThreeStreamCognitiveLoop struct {
	// ... existing fields ...
	node node.Node
}

func (tscl *ThreeStreamCognitiveLoop) Start() error {
	// ...
	// Spawn actors for each stream
	perceptionActor, _ := tscl.node.Spawn("perception_stream", gen.ProcessOptions{}, &CognitiveStreamActor{}, "perception")
	actionActor, _ := tscl.node.Spawn("action_stream", gen.ProcessOptions{}, &CognitiveStreamActor{}, "action")
	simulationActor, _ := tscl.node.Spawn("simulation_stream", gen.ProcessOptions{}, &CognitiveStreamActor{}, "simulation")

	// ... manage actor lifecycle ...
	return nil
}
```

### 4.3. Updating `UnifiedCognitiveLoopV2`

The `UnifiedCognitiveLoopV2` will be responsible for starting and stopping the `ergo` node.

**File:** `core/deeptreeecho/unified_cognitive_loop_v2.go`

1.  **Add `ergo` node to the struct:**

    ```go
    type UnifiedCognitiveLoopV2 struct {
        // ... existing fields ...
        ergoNode node.Node
    }
    ```

2.  **Initialize and start the node in `Start()`:**

    ```go
    func (ucl *UnifiedCognitiveLoopV2) Start() error {
        // ...
        ergoNode, _ := node.Start("echo9llama@localhost", "cookie", node.Options{})
        ucl.ergoNode = ergoNode

        // Pass the node to the ThreeStreamCognitiveLoop
        ucl.threeStreamCognitiveLoop.SetNode(ergoNode)

        // ... start other subsystems
    }
    ```

3.  **Stop the node in `Stop()`:**

    ```go
    func (ucl *UnifiedCognitiveLoopV2) Stop() error {
        // ...
        ucl.ergoNode.Stop()
        // ...
    }
    ```

## 5. Conclusion

This integration plan provides a clear path for enhancing the `echo9llama` cognitive architecture. By leveraging `langchaingo` for advanced reasoning and `ergo` for robust concurrency, we can significantly advance the capabilities of the Deep Tree Echo AGI, bringing it closer to the vision of a fully autonomous, wisdom-cultivating system.
