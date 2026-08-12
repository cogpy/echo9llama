//go:build !cgo || nollama
// +build !cgo nollama

package llm

import (
	"context"
	"fmt"
	"sync"

	"github.com/cogpy/echo9llama/core/backendcap"
)

// LocalGGUFProvider preserves API symmetry when native llama bindings are unavailable.
type LocalGGUFProvider struct {
	mu     sync.RWMutex
	config LocalGGUFProviderConfig
	closed bool
}

func NewLocalGGUFProvider(modelPath string) *LocalGGUFProvider {
	return NewLocalGGUFProviderWithConfig(defaultLocalGGUFConfig(modelPath))
}

func NewLocalGGUFProviderFromCapability(capability backendcap.Capability) *LocalGGUFProvider {
	config := defaultLocalGGUFConfig(capability.ModelPath)
	config.Capability = capability
	return NewLocalGGUFProviderWithConfig(config)
}

func NewLocalGGUFProviderWithConfig(config LocalGGUFProviderConfig) *LocalGGUFProvider {
	return &LocalGGUFProvider{config: normalizeLocalGGUFConfig(config)}
}

func (provider *LocalGGUFProvider) Name() string {
	provider.mu.RLock()
	defer provider.mu.RUnlock()
	return provider.config.Name
}

func (provider *LocalGGUFProvider) Available() bool {
	return false
}

func (provider *LocalGGUFProvider) MaxTokens() int {
	provider.mu.RLock()
	defer provider.mu.RUnlock()
	return provider.config.ContextSize
}

func (provider *LocalGGUFProvider) CapabilitySnapshot() backendcap.Capability {
	provider.mu.RLock()
	defer provider.mu.RUnlock()
	capability := provider.config.Capability
	concrete := capability.ModelPath != ""
	capability.ModelPath = ""
	capability.Available = false
	capability.Native = true
	capability.Offline = true
	capability.Concrete = concrete
	capability.MaxConcurrency = 1
	capability.Reason = "native GGUF runtime unavailable in this build"
	return capability
}

func (provider *LocalGGUFProvider) Generate(ctx context.Context, prompt string, options GenerateOptions) (string, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return "", err
		}
	}
	return "", fmt.Errorf("%w: cgo is disabled or the nollama build tag is active", ErrLocalGGUFUnavailable)
}

func (provider *LocalGGUFProvider) StreamGenerate(ctx context.Context, prompt string, options GenerateOptions) (<-chan StreamChunk, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	return nil, fmt.Errorf("%w: cgo is disabled or the nollama build tag is active", ErrLocalGGUFUnavailable)
}

func (provider *LocalGGUFProvider) Warmup(ctx context.Context) error {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	return fmt.Errorf("%w: cgo is disabled or the nollama build tag is active", ErrLocalGGUFUnavailable)
}

func (provider *LocalGGUFProvider) State() LocalGGUFState {
	provider.mu.RLock()
	defer provider.mu.RUnlock()
	capability := provider.config.Capability
	return LocalGGUFState{
		Name:                 provider.config.Name,
		ModelID:              capability.ModelID,
		ContextSize:          provider.config.ContextSize,
		Quantization:         capability.Quantization,
		EstimatedMemoryBytes: capability.EstimatedMemoryBytes,
		NativeBuild:          false,
		Available:            false,
		Closed:               provider.closed,
		LoadError:            "native GGUF runtime unavailable in this build",
	}
}

func (provider *LocalGGUFProvider) Loaded() bool {
	return false
}

func (provider *LocalGGUFProvider) LoadError() error {
	return fmt.Errorf("%w: cgo is disabled or the nollama build tag is active", ErrLocalGGUFUnavailable)
}

func (provider *LocalGGUFProvider) loadModelForRegistryWarmup() error {
	return provider.Warmup(context.Background())
}

func (provider *LocalGGUFProvider) Close() error {
	provider.mu.Lock()
	provider.closed = true
	provider.mu.Unlock()
	return nil
}

func (provider *LocalGGUFProvider) Reset() error {
	provider.mu.Lock()
	provider.closed = false
	provider.mu.Unlock()
	return nil
}

var (
	_ LLMProvider = (*LocalGGUFProvider)(nil)
	_ interface {
		CapabilitySnapshot() backendcap.Capability
	} = (*LocalGGUFProvider)(nil)
)
