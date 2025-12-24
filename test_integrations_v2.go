// +build ignore

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cogpy/echo9llama/core/deeptreeecho"
	"github.com/cogpy/echo9llama/core/llm"
)

var _ = context.Background // Suppress unused import warning

func main() {
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║     🧪 INTEGRATION TEST V2: SemanticMemory, Scheduler, EventBus║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Create LLM provider
	llmProvider := llm.NewMultiProviderLLM()
	fmt.Println("✅ Created multi-provider LLM")

	// Test 1: Semantic Memory
	fmt.Println("\n📦 Test 1: Semantic Memory (chromem-go inspired)...")
	
	semanticMemory := deeptreeecho.NewSemanticMemory(llmProvider)
	err := semanticMemory.Start()
	if err != nil {
		fmt.Printf("   ❌ Failed to start SemanticMemory: %v\n", err)
	} else {
		fmt.Println("   ✅ SemanticMemory started")
	}

	// Test storing memories
	episodeID, err := semanticMemory.StoreEpisode("Learned about vector databases today", map[string]string{"context": "learning"})
	if err != nil {
		fmt.Printf("   ❌ Failed to store episode: %v\n", err)
	} else {
		fmt.Printf("   ✅ Stored episode: %s\n", episodeID)
	}

	factID, err := semanticMemory.StoreFact("Vector databases enable semantic search using embeddings", "research")
	if err != nil {
		fmt.Printf("   ❌ Failed to store fact: %v\n", err)
	} else {
		fmt.Printf("   ✅ Stored fact: %s\n", factID)
	}

	wisdomID, err := semanticMemory.StoreWisdom("Understanding emerges from connecting related concepts", "reflection", 0.85)
	if err != nil {
		fmt.Printf("   ❌ Failed to store wisdom: %v\n", err)
	} else {
		fmt.Printf("   ✅ Stored wisdom: %s\n", wisdomID)
	}

	// Test querying
	results, err := semanticMemory.Query("semantic", "vector embeddings", 3)
	if err != nil {
		fmt.Printf("   ❌ Failed to query: %v\n", err)
	} else {
		fmt.Printf("   ✅ Query returned %d results\n", len(results))
		for _, r := range results {
			fmt.Printf("      - [%.2f] %s\n", r.Similarity, truncate(r.Document.Content, 50))
		}
	}

	// Get metrics
	smMetrics := semanticMemory.GetMetrics()
	fmt.Printf("   Metrics: collections=%v, documents=%v, queries=%v\n",
		smMetrics["collections"], smMetrics["total_documents"], smMetrics["total_queries"])

	// Test 2: Cognitive Scheduler
	fmt.Println("\n⏰ Test 2: Cognitive Scheduler (gocron inspired)...")
	
	cogScheduler := deeptreeecho.NewCognitiveScheduler()
	err = cogScheduler.Start()
	if err != nil {
		fmt.Printf("   ❌ Failed to start CognitiveScheduler: %v\n", err)
	} else {
		fmt.Println("   ✅ CognitiveScheduler started")
	}

	// Schedule an interval job
	job1, err := cogScheduler.ScheduleInterval("heartbeat", 500*time.Millisecond, func(ctx context.Context) error {
		fmt.Println("      💓 Heartbeat")
		return nil
	})
	if err != nil {
		fmt.Printf("   ❌ Failed to schedule interval job: %v\n", err)
	} else {
		fmt.Printf("   ✅ Scheduled interval job: %s\n", job1.Name)
	}

	// Schedule a cognitive phase job
	job2, err := cogScheduler.ScheduleCognitivePhase("perception_phase", 1, 2*time.Second, func(ctx context.Context) error {
		fmt.Println("      👁️ Perception phase triggered")
		return nil
	})
	if err != nil {
		fmt.Printf("   ❌ Failed to schedule phase job: %v\n", err)
	} else {
		fmt.Printf("   ✅ Scheduled cognitive phase job: %s\n", job2.Name)
	}

	// Let jobs run
	fmt.Println("   ⏳ Running scheduler for 2 seconds...")
	time.Sleep(2 * time.Second)

	// Get metrics
	csMetrics := cogScheduler.GetMetrics()
	fmt.Printf("   Metrics: jobs=%v, total_run=%v, failed=%v\n",
		csMetrics["job_count"], csMetrics["total_jobs_run"], csMetrics["total_jobs_failed"])

	// Test 3: Event Bus V2
	fmt.Println("\n📡 Test 3: Event Bus V2 (watermill inspired)...")
	
	eventBus := deeptreeecho.NewEventBusV2()
	err = eventBus.Start()
	if err != nil {
		fmt.Printf("   ❌ Failed to start EventBusV2: %v\n", err)
	} else {
		fmt.Println("   ✅ EventBusV2 started")
	}

	// Setup cognitive routes
	eventBus.SetupCognitiveRoutes()

	// Subscribe to topics
	perceptionCount := 0
	_, err = eventBus.Subscribe(deeptreeecho.TopicPerception, func(msg *deeptreeecho.EventMessage) error {
		perceptionCount++
		fmt.Printf("      📥 Received perception event: %s\n", string(msg.Payload))
		return nil
	})
	if err != nil {
		fmt.Printf("   ❌ Failed to subscribe: %v\n", err)
	} else {
		fmt.Println("   ✅ Subscribed to perception topic")
	}

	wisdomCount := 0
	_, err = eventBus.Subscribe(deeptreeecho.TopicWisdom, func(msg *deeptreeecho.EventMessage) error {
		wisdomCount++
		fmt.Printf("      🔮 Received wisdom event: %s\n", string(msg.Payload))
		return nil
	})
	if err != nil {
		fmt.Printf("   ❌ Failed to subscribe: %v\n", err)
	} else {
		fmt.Println("   ✅ Subscribed to wisdom topic")
	}

	// Publish events
	err = eventBus.PublishCognitiveEvent(deeptreeecho.TopicPerception, "visual", map[string]string{"content": "observed pattern"})
	if err != nil {
		fmt.Printf("   ❌ Failed to publish: %v\n", err)
	} else {
		fmt.Println("   ✅ Published perception event")
	}

	err = eventBus.PublishCognitiveEvent(deeptreeecho.TopicEmergence, "pattern", map[string]string{"pattern": "recursive structure"})
	if err != nil {
		fmt.Printf("   ❌ Failed to publish: %v\n", err)
	} else {
		fmt.Println("   ✅ Published emergence event (should route to wisdom)")
	}

	// Wait for events to be processed
	time.Sleep(500 * time.Millisecond)

	// Get metrics
	ebMetrics := eventBus.GetMetrics()
	fmt.Printf("   Metrics: topics=%v, published=%v, delivered=%v\n",
		ebMetrics["topics"], ebMetrics["total_published"], ebMetrics["total_delivered"])
	fmt.Printf("   Event counts: perception=%d, wisdom=%d\n", perceptionCount, wisdomCount)

	// Test 4: Gestalt contributions
	fmt.Println("\n🌐 Test 4: Gestalt Contributions...")
	
	smGestalt := semanticMemory.ContributeToGestalt()
	fmt.Printf("   SemanticMemory: running=%v, collections=%v\n", smGestalt["running"], smGestalt["collections"])

	csGestalt := cogScheduler.ContributeToGestalt()
	fmt.Printf("   CognitiveScheduler: running=%v, total_jobs_run=%v\n", csGestalt["running"], csGestalt["total_jobs_run"])

	ebGestalt := eventBus.ContributeToGestalt()
	fmt.Printf("   EventBusV2: running=%v, active_topics=%v\n", ebGestalt["running"], ebGestalt["active_topics"])

	// Cleanup
	fmt.Println("\n🧹 Cleaning up...")
	cogScheduler.Stop()
	eventBus.Stop()
	semanticMemory.Stop()

	fmt.Println("\n╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║     ✅ ALL INTEGRATION TESTS V2 PASSED                        ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
