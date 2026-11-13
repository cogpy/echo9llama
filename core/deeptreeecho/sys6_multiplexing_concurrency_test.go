package deeptreeecho

import (
	"context"
	"runtime"
	"sync"
	"testing"
)

func TestSys6MultiplexingConcurrentCyclesMetricsAndStop(t *testing.T) {
	tm := NewSys6ThreadMultiplexer(context.Background())
	if err := tm.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Cleanup(func() {
		if err := tm.Stop(); err != nil {
			t.Errorf("Stop() cleanup error = %v", err)
		}
	})

	startCycles := make(chan struct{})
	var cycles sync.WaitGroup
	cycles.Add(2)

	go func() {
		defer cycles.Done()
		<-startCycles
		tm.ExecuteDyadicCycle(1)
	}()
	go func() {
		defer cycles.Done()
		<-startCycles
		tm.ExecuteTriadicCycle(1)
	}()

	stopMetrics := make(chan struct{})
	var metricsReaders sync.WaitGroup
	metricsReaders.Add(1)
	go func() {
		defer metricsReaders.Done()
		for {
			select {
			case <-stopMetrics:
				return
			default:
				_ = tm.GetMetrics()
				runtime.Gosched()
			}
		}
	}()

	stopSignal := make(chan struct{})
	stopResult := make(chan error, 1)
	go func() {
		<-stopSignal
		stopResult <- tm.Stop()
	}()

	close(startCycles)
	cycles.Wait()

	// A dyadic cycle completes two workers. A triadic cycle must complete all
	// six workers: three MP1 workers and three simultaneous MP2 workers.
	assertSys6WorkerCounts(t, tm.GetMetrics(), [4]uint64{3, 2, 2, 1}, 3)

	close(stopSignal)
	if err := <-stopResult; err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	close(stopMetrics)
	metricsReaders.Wait()

	metricsAfterStop := tm.GetMetrics()
	assertSys6WorkerCounts(t, metricsAfterStop, [4]uint64{3, 2, 2, 1}, 3)
	if running, ok := metricsAfterStop["running"].(bool); !ok || running {
		t.Fatalf("running after Stop() = %#v, want false", metricsAfterStop["running"])
	}
	for _, worker := range metricsAfterStop["workers"].([]map[string]interface{}) {
		if active, ok := worker["active"].(bool); !ok || active {
			t.Fatalf("worker %v active after Stop() = %#v, want false", worker["id"], worker["active"])
		}
	}

	// The cycle calls have returned and Stop has completed, so further yields
	// must not reveal a late MP2 worker completion.
	for range 100 {
		runtime.Gosched()
	}
	assertSys6WorkerCounts(t, tm.GetMetrics(), [4]uint64{3, 2, 2, 1}, 3)
}

func assertSys6WorkerCounts(t *testing.T, metrics map[string]interface{}, want [4]uint64, wantEntanglements uint64) {
	t.Helper()

	if entanglements, ok := metrics["entanglements"].(uint64); !ok || entanglements != wantEntanglements {
		t.Fatalf("entanglements = %#v, want %d", metrics["entanglements"], wantEntanglements)
	}

	workers, ok := metrics["workers"].([]map[string]interface{})
	if !ok {
		t.Fatalf("workers = %#v, want []map[string]interface{}", metrics["workers"])
	}
	if len(workers) != len(want) {
		t.Fatalf("workers length = %d, want %d", len(workers), len(want))
	}

	var total uint64
	for i, worker := range workers {
		processed, ok := worker["processed"].(uint64)
		if !ok {
			t.Fatalf("worker %d processed = %#v, want uint64", i+1, worker["processed"])
		}
		if processed != want[i] {
			t.Fatalf("worker %d processed = %d, want %d", i+1, processed, want[i])
		}
		total += processed
	}
	if total != 8 {
		t.Fatalf("total completed workers = %d, want 8 (2 dyadic + 6 triadic)", total)
	}
}
