//go:build cgo && !nollama

package llm

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cogpy/echo9llama/core/backendcap"
)

func TestLocalGGUFProviderRealModel(t *testing.T) {
	modelPath := strings.TrimSpace(os.Getenv("ECHO_TEST_REAL_GGUF"))
	if modelPath == "" {
		t.Skip("set ECHO_TEST_REAL_GGUF to run the native model integration")
	}
	absolute, err := filepath.Abs(modelPath)
	if err != nil {
		t.Fatal(err)
	}
	capability, err := backendcap.ProbeModelFile(absolute, backendcap.DiscoveryOptions{
		Roots:              []string{filepath.Dir(absolute)},
		AllowUnknownMemory: true,
		MemorySafetyRatio:  0.90,
	})
	if err != nil {
		t.Fatalf("probe real GGUF: %v", err)
	}
	contextSize := capability.ContextLength
	if contextSize <= 0 || contextSize > 512 {
		contextSize = 512
	}
	provider := NewLocalGGUFProviderWithConfig(LocalGGUFProviderConfig{
		Name:               "real-local-gguf",
		ModelPath:          absolute,
		ModelRoots:         []string{filepath.Dir(absolute)},
		ContextSize:        contextSize,
		BatchSize:          64,
		Threads:            2,
		QueueWait:          2 * time.Second,
		AllowUnknownMemory: true,
		MemorySafetyRatio:  0.90,
		MemoryReserveBytes: 0,
		Capability:         capability,
	})
	t.Cleanup(func() {
		if err := provider.Close(); err != nil {
			t.Errorf("close native provider: %v", err)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	if err := provider.Warmup(ctx); err != nil {
		t.Fatalf("warm real GGUF: %v", err)
	}
	if !provider.Loaded() || !provider.Available() {
		t.Fatalf("provider not ready after warmup: %#v", provider.State())
	}
	if snapshot := provider.CapabilitySnapshot(); snapshot.ModelPath != "" || snapshot.ModelID == "" {
		t.Fatalf("real-model capability leaked path or lost identity: %#v", snapshot)
	}

	longPrompt := strings.Repeat("Once upon a time the wise explorer reflected carefully. ", 80)
	result, err := provider.Generate(ctx, longPrompt, GenerateOptions{
		MaxTokens:   8,
		Temperature: 0.7,
		Routing: RoutingOptions{
			PolicyExplicit: true,
			AllowQueue:     true,
			QueueWait:      2 * time.Second,
		},
	})
	if err != nil {
		t.Fatalf("real GGUF generation: %v", err)
	}
	if strings.TrimSpace(result) == "" {
		t.Fatal("real GGUF returned empty nonstreaming output")
	}

	stream, err := provider.StreamGenerate(ctx, "The wise explorer", GenerateOptions{
		MaxTokens:   8,
		Temperature: 0.7,
		Routing: RoutingOptions{
			PolicyExplicit: true,
			AllowQueue:     true,
			QueueWait:      2 * time.Second,
		},
	})
	if err != nil {
		t.Fatalf("real GGUF stream start: %v", err)
	}
	var streamed strings.Builder
	seenDone := false
	for chunk := range stream {
		if chunk.Error != nil {
			t.Fatalf("real GGUF stream: %v", chunk.Error)
		}
		streamed.WriteString(chunk.Content)
		seenDone = seenDone || chunk.Done
	}
	if strings.TrimSpace(streamed.String()) == "" || !seenDone {
		t.Fatalf("real GGUF stream incomplete: output=%q done=%v", streamed.String(), seenDone)
	}
}
