package llm

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/backendcap"
)

const (
	ModelLifecycleDiscovered   = "model_discovered"
	ModelLifecycleSelected     = "model_selected"
	ModelLifecycleLoaded       = "model_loaded"
	ModelLifecycleUnloaded     = "model_unloaded"
	ModelLifecycleLoadFailed   = "model_load_failed"
	ModelLifecyclePolicyScored = "model_policy_scored"
)

// ModelSelectionTask describes the runtime need used to select a GGUF candidate.
type ModelSelectionTask struct {
	Intent                string
	RequiredContextTokens int
	ExpectedOutputTokens  int
}

// ModelScoringPolicy scores one safe concrete model. Higher scores are preferred.
type ModelScoringPolicy func(capability backendcap.Capability, task ModelSelectionTask) int

// LocalModelEvent is a path-scrubbed lifecycle transition snapshot.
type LocalModelEvent struct {
	Type                  string                     `json:"type"`
	ModelID               string                     `json:"model_id,omitempty"`
	ModelName             string                     `json:"model_name,omitempty"`
	Capability            backendcap.Capability      `json:"capability"`
	EstimatedMemoryBytes  uint64                     `json:"estimated_memory_bytes,omitempty"`
	Loaded                bool                       `json:"loaded"`
	Reason                string                     `json:"reason,omitempty"`
	ErrorCategory         string                     `json:"error_category,omitempty"`
	Error                 string                     `json:"error,omitempty"`
	PolicyScore           int                        `json:"policy_score,omitempty"`
	PolicyIntent          string                     `json:"policy_intent,omitempty"`
	RequiredContextTokens int                        `json:"required_context_tokens,omitempty"`
	Timestamp             time.Time                  `json:"timestamp"`
	HostMemory            backendcap.HostMemoryProbe `json:"host_memory"`
}

// LocalModelRegistryOptions configures discovery, selection, and residency policy.
type LocalModelRegistryOptions struct {
	ModelPaths         []string
	ModelRoots         []string
	ProviderName       string
	MemorySafetyRatio  float64
	MemoryReserveBytes uint64
	AllowUnknownMemory bool
	IdleUnloadAfter    time.Duration
	ScoringPolicy      ModelScoringPolicy
	SelectionTask      ModelSelectionTask
	OnEvent            func(LocalModelEvent)
	Now                func() time.Time
	ProbeHostMemory    func() backendcap.HostMemoryProbe
	providerFactory    func(backendcap.Capability) localGGUFRuntime
	discoverModels     func([]string, backendcap.DiscoveryOptions) []backendcap.Capability
}

type localGGUFRuntime interface {
	LLMProvider
	CapabilitySnapshot() backendcap.Capability
	Warmup(context.Context) error
	State() LocalGGUFState
	Loaded() bool
	LoadError() error
	Close() error
	Reset() error
}

// LocalModelRegistry is the sole owner of concrete-model selection and residency.
type LocalModelRegistry struct {
	mu sync.RWMutex

	options  LocalModelRegistryOptions
	models   []backendcap.Capability
	selected backendcap.Capability
	provider localGGUFRuntime
	wrapper  *localModelRegistryProvider
	closed   bool

	loadErr      error
	lastUsed     time.Time
	lastLoaded   time.Time
	lastUnloaded time.Time
	unloadReason string
}

// LocalModelRegistryState is safe for public status and does not expose model paths.
type LocalModelRegistryState struct {
	ConfiguredPathCount  int                        `json:"configured_path_count"`
	DiscoveredModels     []backendcap.Capability    `json:"discovered_models"`
	SelectedModel        backendcap.Capability      `json:"selected_model"`
	ProviderName         string                     `json:"provider_name"`
	Loaded               bool                       `json:"loaded"`
	Loading              bool                       `json:"loading"`
	Closed               bool                       `json:"closed"`
	InFlight             int                        `json:"in_flight"`
	LoadError            string                     `json:"load_error,omitempty"`
	EstimatedMemoryBytes uint64                     `json:"estimated_memory_bytes,omitempty"`
	LastUsed             time.Time                  `json:"last_used,omitempty"`
	LastLoaded           time.Time                  `json:"last_loaded,omitempty"`
	LastUnloaded         time.Time                  `json:"last_unloaded,omitempty"`
	UnloadReason         string                     `json:"unload_reason,omitempty"`
	HostMemory           backendcap.HostMemoryProbe `json:"host_memory"`
	MemorySafe           bool                       `json:"memory_safe"`
	RuntimeReady         bool                       `json:"runtime_ready"`
}

// NewLocalModelRegistry creates a registry and performs one bounded discovery pass.
func NewLocalModelRegistry(options LocalModelRegistryOptions) *LocalModelRegistry {
	if strings.TrimSpace(options.ProviderName) == "" {
		options.ProviderName = "local_gguf"
	}
	if options.MemorySafetyRatio <= 0 || options.MemorySafetyRatio > 1 {
		options.MemorySafetyRatio = 0.80
	}
	if options.MemoryReserveBytes == 0 {
		options.MemoryReserveBytes = 1024 * 1024 * 1024
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	if options.ProbeHostMemory == nil {
		options.ProbeHostMemory = backendcap.ProbeHostMemory
	}
	if options.discoverModels == nil {
		options.discoverModels = func(paths []string, option backendcap.DiscoveryOptions) []backendcap.Capability {
			return backendcap.DiscoverModelCapabilities(paths, option)
		}
	}
	if options.providerFactory == nil {
		options.providerFactory = func(capability backendcap.Capability) localGGUFRuntime {
			config := defaultLocalGGUFConfig(capability.ModelPath)
			config.Name = options.ProviderName
			config.ModelRoots = append([]string(nil), options.ModelRoots...)
			config.MemorySafetyRatio = options.MemorySafetyRatio
			config.MemoryReserveBytes = options.MemoryReserveBytes
			config.AllowUnknownMemory = options.AllowUnknownMemory
			config.Capability = capability
			return NewLocalGGUFProviderWithConfig(config)
		}
	}
	registry := &LocalModelRegistry{options: options}
	registry.wrapper = &localModelRegistryProvider{registry: registry}
	registry.Refresh()
	return registry
}

// Refresh reprobes configured paths and replaces the selected provider only when
// the concrete model changes. Native Close and callbacks occur after unlocking.
func (registry *LocalModelRegistry) Refresh() LocalModelRegistryState {
	registry.mu.RLock()
	closed := registry.closed
	registry.mu.RUnlock()
	if closed {
		return registry.State()
	}
	hostMemory := registry.options.ProbeHostMemory()
	models := registry.options.discoverModels(registry.options.ModelPaths, backendcap.DiscoveryOptions{
		Roots:              registry.options.ModelRoots,
		AllowUnknownMemory: registry.options.AllowUnknownMemory,
		MemorySafetyRatio:  registry.options.MemorySafetyRatio,
		MemoryReserveBytes: registry.options.MemoryReserveBytes,
		HostMemory:         &hostMemory,
	})
	selected, scoreEvents := registry.selectModel(models)

	registry.mu.Lock()
	if registry.closed {
		registry.mu.Unlock()
		return registry.State()
	}
	oldProvider := localGGUFRuntime(nil)
	changed := selected.ModelID != registry.selected.ModelID || selected.ModelPath != registry.selected.ModelPath
	if changed {
		oldProvider = registry.provider
		registry.provider = nil
		registry.loadErr = nil
		registry.unloadReason = "selected model changed"
	}
	registry.models = cloneCapabilities(models)
	registry.selected = selected
	registry.mu.Unlock()

	if oldProvider != nil {
		_ = oldProvider.Close()
	}
	for _, event := range scoreEvents {
		registry.emit(event)
	}
	registry.emit(LocalModelEvent{
		Type:       ModelLifecycleDiscovered,
		Reason:     fmt.Sprintf("discovered %d safe concrete GGUF candidate(s)", len(models)),
		Timestamp:  registry.options.Now(),
		HostMemory: registry.options.ProbeHostMemory(),
	})
	if selected.ModelID != "" {
		registry.emit(registry.eventFor(ModelLifecycleSelected, selected, false, nil, "selected highest-scoring safe concrete GGUF model"))
	}
	return registry.State()
}

func (registry *LocalModelRegistry) selectModel(models []backendcap.Capability) (backendcap.Capability, []LocalModelEvent) {
	candidates := make([]backendcap.Capability, 0, len(models))
	for _, model := range models {
		if model.ModelPath == "" || !model.Available || !model.MemorySafe {
			continue
		}
		required := registry.options.SelectionTask.RequiredContextTokens + registry.options.SelectionTask.ExpectedOutputTokens
		if required > 0 && model.ContextLength > 0 && model.ContextLength < required {
			continue
		}
		candidates = append(candidates, model)
	}

	events := make([]LocalModelEvent, 0, len(candidates))
	sort.SliceStable(candidates, func(left, right int) bool {
		leftScore := registry.scoreModel(candidates[left])
		rightScore := registry.scoreModel(candidates[right])
		if leftScore == rightScore {
			return candidates[left].ModelID < candidates[right].ModelID
		}
		return leftScore > rightScore
	})
	for _, candidate := range candidates {
		score := registry.scoreModel(candidate)
		event := registry.eventFor(ModelLifecyclePolicyScored, candidate, false, nil, "registry scoring policy evaluated concrete GGUF candidate")
		event.PolicyScore = score
		event.PolicyIntent = registry.options.SelectionTask.Intent
		event.RequiredContextTokens = registry.options.SelectionTask.RequiredContextTokens
		events = append(events, event)
	}
	if len(candidates) == 0 {
		return backendcap.Capability{}, events
	}
	return candidates[0], events
}

func (registry *LocalModelRegistry) scoreModel(capability backendcap.Capability) int {
	if registry.options.ScoringPolicy != nil {
		return registry.options.ScoringPolicy(capability, registry.options.SelectionTask)
	}
	score := capability.Priority
	required := registry.options.SelectionTask.RequiredContextTokens + registry.options.SelectionTask.ExpectedOutputTokens
	if required > 0 && capability.ContextLength >= required {
		score += 100
	}
	if capability.ContextLength > 0 {
		score += minLocalInt(capability.ContextLength/1024, 64)
	}
	// Prefer the smallest adequate resident footprint for repeated autonomous work.
	score -= int(capability.EstimatedMemoryBytes / (1024 * 1024 * 1024))
	return score
}

// Provider returns a stable registry wrapper, not the disposable native instance.
func (registry *LocalModelRegistry) Provider() LLMProvider {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	if registry.closed || registry.selected.ModelPath == "" {
		return nil
	}
	return registry.wrapper
}

func (registry *LocalModelRegistry) providerForUse() (localGGUFRuntime, error) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	if registry.closed {
		return nil, ErrLocalGGUFClosed
	}
	if registry.selected.ModelPath == "" {
		return nil, ErrLocalGGUFUnavailable
	}
	if registry.provider == nil {
		registry.provider = registry.options.providerFactory(registry.selected)
	}
	return registry.provider, nil
}

// State returns a path-scrubbed snapshot without holding registry locks while
// querying the provider or the host.
func (registry *LocalModelRegistry) State() LocalModelRegistryState {
	registry.mu.RLock()
	models := cloneCapabilities(registry.models)
	selectedInternal := registry.selected
	selected := scrubCapability(selectedInternal)
	provider := registry.provider
	closed := registry.closed
	loadErr := registry.loadErr
	lastUsed := registry.lastUsed
	lastLoaded := registry.lastLoaded
	lastUnloaded := registry.lastUnloaded
	unloadReason := registry.unloadReason
	pathCount := len(registry.options.ModelPaths)
	providerName := registry.options.ProviderName
	registry.mu.RUnlock()

	host := registry.options.ProbeHostMemory()
	state := LocalModelRegistryState{
		ConfiguredPathCount:  pathCount,
		DiscoveredModels:     models,
		SelectedModel:        selected,
		ProviderName:         providerName,
		Closed:               closed,
		EstimatedMemoryBytes: selected.EstimatedMemoryBytes,
		LastUsed:             lastUsed,
		LastLoaded:           lastLoaded,
		LastUnloaded:         lastUnloaded,
		UnloadReason:         unloadReason,
		HostMemory:           host,
		MemorySafe:           selectedInternal.ModelPath != "" && selectedInternal.MemorySafe,
	}
	if provider != nil {
		providerState := provider.State()
		state.Loaded = providerState.Loaded
		state.Loading = providerState.Loading
		state.InFlight = providerState.InFlight
		if providerState.LoadError != "" {
			state.LoadError = providerState.LoadError
		}
	}
	if loadErr != nil {
		state.LoadError = loadErr.Error()
	}
	state.RuntimeReady = state.SelectedModel.ModelID != "" && state.Loaded && state.MemorySafe && state.LoadError == "" && !state.Closed
	return state
}

// Warmup eagerly loads the selected model through the registry-owned provider.
func (registry *LocalModelRegistry) Warmup(ctx context.Context) error {
	provider, err := registry.providerForUse()
	if err != nil {
		return err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	err = provider.Warmup(ctx)
	registry.recordUse(provider, err, "explicit registry warmup")
	return err
}

// Cooldown unloads the resident provider but keeps selection available for lazy reload.
func (registry *LocalModelRegistry) Cooldown(reason string) bool {
	if strings.TrimSpace(reason) == "" {
		reason = "runtime cooldown requested"
	}
	return registry.Unload(reason)
}

func (registry *LocalModelRegistry) RuntimeReadiness() bool {
	return registry.State().RuntimeReady
}

// MaybeUnloadForMemoryPressure unloads when the selected model no longer fits
// the current effective host envelope.
func (registry *LocalModelRegistry) MaybeUnloadForMemoryPressure(reason string) bool {
	state := registry.State()
	if !state.Loaded || state.SelectedModel.EstimatedMemoryBytes == 0 || !state.HostMemory.Known {
		return false
	}
	limit := uint64(float64(state.HostMemory.AvailableBytes) * registry.options.MemorySafetyRatio)
	reserveLimit := uint64(0)
	if state.HostMemory.AvailableBytes > registry.options.MemoryReserveBytes {
		reserveLimit = state.HostMemory.AvailableBytes - registry.options.MemoryReserveBytes
	}
	if reserveLimit < limit {
		limit = reserveLimit
	}
	if state.SelectedModel.EstimatedMemoryBytes <= limit {
		return false
	}
	if strings.TrimSpace(reason) == "" {
		reason = "selected local model exceeds the current safe host-memory envelope"
	}
	return registry.Unload(reason)
}

// UnloadIdle applies the configured idle residency limit.
func (registry *LocalModelRegistry) UnloadIdle(reason string) bool {
	state := registry.State()
	if !state.Loaded || registry.options.IdleUnloadAfter <= 0 || state.LastUsed.IsZero() {
		return false
	}
	if registry.options.Now().Sub(state.LastUsed) < registry.options.IdleUnloadAfter {
		return false
	}
	if strings.TrimSpace(reason) == "" {
		reason = "idle unload policy"
	}
	return registry.Unload(reason)
}

// Unload drains and closes the active native provider, then permits lazy recreation.
func (registry *LocalModelRegistry) Unload(reason string) bool {
	registry.mu.Lock()
	provider := registry.provider
	if provider == nil {
		registry.mu.Unlock()
		return false
	}
	registry.provider = nil
	registry.mu.Unlock()

	wasLoaded := provider.Loaded()
	_ = provider.Close()
	now := registry.options.Now()
	registry.mu.Lock()
	registry.lastUnloaded = now
	registry.unloadReason = reason
	registry.mu.Unlock()
	if wasLoaded {
		registry.emit(registry.eventFor(ModelLifecycleUnloaded, registry.selectedSnapshot(), false, nil, reason))
	}
	return wasLoaded
}

// Close terminally closes the registry and its provider.
func (registry *LocalModelRegistry) Close() error {
	registry.mu.Lock()
	if registry.closed {
		registry.mu.Unlock()
		return nil
	}
	registry.closed = true
	provider := registry.provider
	registry.provider = nil
	registry.mu.Unlock()
	if provider != nil {
		return provider.Close()
	}
	return nil
}

func (registry *LocalModelRegistry) recordUse(provider localGGUFRuntime, err error, reason string) {
	now := registry.options.Now()
	loaded := provider != nil && provider.Loaded()
	registry.mu.Lock()
	wasLoaded := !registry.lastLoaded.IsZero() && registry.lastLoaded.After(registry.lastUnloaded)
	registry.lastUsed = now
	if err != nil {
		registry.loadErr = err
	} else {
		registry.loadErr = nil
		if loaded && !wasLoaded {
			registry.lastLoaded = now
		}
	}
	selected := registry.selected
	registry.mu.Unlock()
	if err != nil {
		registry.emit(registry.eventFor(ModelLifecycleLoadFailed, selected, false, err, reason))
		return
	}
	if loaded && !wasLoaded {
		registry.emit(registry.eventFor(ModelLifecycleLoaded, selected, true, nil, reason))
	}
}

func (registry *LocalModelRegistry) selectedSnapshot() backendcap.Capability {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	return registry.selected
}

func (registry *LocalModelRegistry) eventFor(eventType string, capability backendcap.Capability, loaded bool, err error, reason string) LocalModelEvent {
	capability = scrubCapability(capability)
	event := LocalModelEvent{
		Type:                 eventType,
		ModelID:              capability.ModelID,
		ModelName:            strings.TrimPrefix(capability.Name, "model:"),
		Capability:           capability,
		EstimatedMemoryBytes: capability.EstimatedMemoryBytes,
		Loaded:               loaded,
		Reason:               reason,
		Timestamp:            registry.options.Now(),
		HostMemory:           registry.options.ProbeHostMemory(),
	}
	if err != nil {
		event.Error = err.Error()
		event.ErrorCategory = localModelErrorCategory(err)
	}
	return event
}

func (registry *LocalModelRegistry) emit(event LocalModelEvent) {
	if registry.options.OnEvent != nil {
		registry.options.OnEvent(event)
	}
}

func localModelErrorCategory(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, context.Canceled):
		return "canceled"
	case errors.Is(err, context.DeadlineExceeded):
		return "deadline"
	case errors.Is(err, ErrLocalGGUFQueueSaturated):
		return "queue_saturated"
	case errors.Is(err, ErrLocalGGUFClosed):
		return "closed"
	case errors.Is(err, ErrLocalGGUFUnavailable):
		return "unavailable"
	default:
		return "runtime"
	}
}

func cloneCapabilities(values []backendcap.Capability) []backendcap.Capability {
	result := make([]backendcap.Capability, len(values))
	for index, capability := range values {
		result[index] = scrubCapability(capability)
		result[index].BuildTags = append([]string(nil), capability.BuildTags...)
	}
	return result
}

func scrubCapability(capability backendcap.Capability) backendcap.Capability {
	capability.ModelPath = ""
	return capability
}

type localModelRegistryProvider struct {
	registry *LocalModelRegistry
}

func (provider *localModelRegistryProvider) Generate(ctx context.Context, prompt string, options GenerateOptions) (string, error) {
	runtimeProvider, err := provider.registry.providerForUse()
	if err != nil {
		return "", err
	}
	result, err := runtimeProvider.Generate(ctx, prompt, options)
	provider.registry.recordUse(runtimeProvider, err, "generation completed")
	return result, err
}

func (provider *localModelRegistryProvider) StreamGenerate(ctx context.Context, prompt string, options GenerateOptions) (<-chan StreamChunk, error) {
	runtimeProvider, err := provider.registry.providerForUse()
	if err != nil {
		return nil, err
	}
	stream, err := runtimeProvider.StreamGenerate(ctx, prompt, options)
	if err != nil {
		provider.registry.recordUse(runtimeProvider, err, "stream start failed")
		return nil, err
	}
	output := make(chan StreamChunk, 16)
	go func() {
		defer close(output)
		var streamErr error
		for chunk := range stream {
			if chunk.Error != nil {
				streamErr = chunk.Error
			}
			select {
			case output <- chunk:
			case <-ctx.Done():
				streamErr = ctx.Err()
				provider.registry.recordUse(runtimeProvider, streamErr, "stream canceled")
				return
			}
		}
		provider.registry.recordUse(runtimeProvider, streamErr, "stream completed")
	}()
	return output, nil
}

func (provider *localModelRegistryProvider) Name() string {
	return provider.registry.options.ProviderName
}

func (provider *localModelRegistryProvider) Available() bool {
	provider.registry.MaybeUnloadForMemoryPressure("memory pressure before local availability check")
	state := provider.registry.State()
	return !state.Closed && state.SelectedModel.ModelID != "" && state.MemorySafe
}

func (provider *localModelRegistryProvider) MaxTokens() int {
	state := provider.registry.State()
	if state.SelectedModel.ContextLength > 0 {
		return minLocalInt(state.SelectedModel.ContextLength, defaultLocalGGUFContext)
	}
	return defaultLocalGGUFContext
}

func (provider *localModelRegistryProvider) CapabilitySnapshot() backendcap.Capability {
	state := provider.registry.State()
	capability := state.SelectedModel
	capability.Available = !state.Closed && state.SelectedModel.ModelID != "" && state.MemorySafe
	capability.CurrentInFlight = state.InFlight
	capability.MaxConcurrency = 1
	return capability
}

var _ LLMProvider = (*localModelRegistryProvider)(nil)
