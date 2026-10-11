package integration

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func newLifecycleTestHub(t *testing.T) *DeepTreeEchoHub {
	t.Helper()

	config := DefaultHubConfig()
	config.EnableAutonomousAgent = false
	config.EnableUnifiedOrchestrator = false
	config.EnableTelemetryShell = false
	config.EnableStateManager = false
	config.MainLoopInterval = time.Hour
	config.StateUpdateInterval = time.Hour
	config.MetricsSyncInterval = time.Hour

	hub := NewDeepTreeEchoHub(&MockLLMProvider{}, config)
	if err := hub.Start(); err != nil {
		t.Fatalf("start hub: %v", err)
	}

	return hub
}

func TestDeepTreeEchoHubStopWhilePublishing(t *testing.T) {
	hub := newLifecycleTestHub(t)

	const publishers = 8
	startPublishing := make(chan struct{})
	stopPublishing := make(chan struct{})
	var ready sync.WaitGroup
	var publisherWG sync.WaitGroup
	var publishAttempts atomic.Uint64

	ready.Add(publishers)
	publisherWG.Add(publishers)
	for publisher := range publishers {
		go func(publisher int) {
			defer publisherWG.Done()
			ready.Done()
			<-startPublishing

			for {
				select {
				case <-stopPublishing:
					return
				default:
					hub.publishEvent("race_test", "concurrent_publisher", publisher)
					publishAttempts.Add(1)
				}
			}
		}(publisher)
	}

	ready.Wait()
	close(startPublishing)

	deadline := time.After(time.Second)
	for publishAttempts.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("publishers did not start")
		default:
			time.Sleep(time.Millisecond)
		}
	}

	if err := hub.Stop(); err != nil {
		t.Fatalf("stop hub: %v", err)
	}
	close(stopPublishing)
	publisherWG.Wait()

	if publishAttempts.Load() == 0 {
		t.Fatal("expected concurrent event publication attempts")
	}
}

func TestDeepTreeEchoHubStopIsTerminal(t *testing.T) {
	hub := newLifecycleTestHub(t)

	if err := hub.Stop(); err != nil {
		t.Fatalf("stop hub: %v", err)
	}
	if err := hub.Start(); err == nil {
		t.Fatal("expected a stopped hub to reject restart")
	}
}

func TestDeepTreeEchoHubIntegrationCallbackDoesNotReenterHeldLock(t *testing.T) {
	config := DefaultHubConfig()
	config.EnableUnifiedOrchestrator = false
	config.MainLoopInterval = time.Hour
	config.StateUpdateInterval = time.Hour
	config.MetricsSyncInterval = time.Hour
	hub := NewDeepTreeEchoHub(&MockLLMProvider{}, config)
	if err := hub.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() { hub.performIntegrationCycle(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("integration callback deadlocked on hub mutex")
	}
	if hub.GetStatus().Metrics.TotalCycles != 1 {
		t.Fatal("integration cycle did not execute")
	}
	if err := hub.Stop(); err != nil {
		t.Fatal(err)
	}
}
