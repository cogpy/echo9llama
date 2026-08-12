package llm

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/backendcap"
)

const (
	maxRouteAttempts = 64
	maxRouteResults  = 256
)

// Provider is the historical non-streaming provider surface retained for
// compatibility with EchoDream and older autonomous packages.
type Provider interface {
	Generate(context.Context, string, GenerateOptions) (string, error)
	Name() string
	Available() bool
	MaxTokens() int
}

// MultiProviderLLM is the single production capability-aware inference router.
type MultiProviderLLM struct {
	providers []Provider
	mu        sync.RWMutex
	stats     map[string]*ProviderStats

	mode           ProviderMode
	attemptTimeout time.Duration
	localRegistry  *LocalModelRegistry
	routingState   BackendRoutingState
	routeResults   map[string]BackendRoutingState
	routeOrder     []string
}

// ProviderStats tracks provider performance without retaining prompts or outputs.
type ProviderStats struct {
	TotalCalls        int64
	SuccessCalls      int64
	FailedCalls       int64
	TotalLatency      time.Duration
	LastUsed          time.Time
	Available         bool
	LastErrorCategory string
}

// NewMultiProviderLLM creates a provider router from environment configuration.
func NewMultiProviderLLM() *MultiProviderLLM {
	router := &MultiProviderLLM{
		providers:      make([]Provider, 0, 5),
		stats:          make(map[string]*ProviderStats),
		mode:           parseProviderMode(os.Getenv("ECHO_PROVIDER_MODE")),
		attemptTimeout: localEnvDuration("ECHO_INFERENCE_TIMEOUT", 90*time.Second),
		routeResults:   make(map[string]BackendRoutingState),
		routeOrder:     make([]string, 0, 64),
	}
	router.routingState.Mode = router.mode
	router.initializeProviders()
	return router
}

func (router *MultiProviderLLM) initializeProviders() {
	modelPaths := splitLocalPathList(os.Getenv("ECHO_MODEL_PATHS"))
	if legacyPath := strings.TrimSpace(os.Getenv("LOCAL_MODEL_PATH")); legacyPath != "" {
		modelPaths = appendUniqueString(modelPaths, legacyPath)
	}
	modelRoots := splitLocalPathList(os.Getenv("ECHO_MODEL_ROOTS"))
	if len(modelPaths) > 0 {
		registry := NewLocalModelRegistry(LocalModelRegistryOptions{
			ModelPaths:         modelPaths,
			ModelRoots:         modelRoots,
			ProviderName:       "local_gguf",
			MemorySafetyRatio:  localEnvFloat("ECHO_MODEL_MEMORY_RATIO", 0.80),
			MemoryReserveBytes: localEnvBytes("ECHO_MODEL_MEMORY_RESERVE", 1024*1024*1024),
			AllowUnknownMemory: localEnvBool("ECHO_MODEL_ALLOW_UNKNOWN_MEMORY", false),
			IdleUnloadAfter:    localEnvDuration("ECHO_LOCAL_IDLE_UNLOAD", 30*time.Minute),
			SelectionTask: ModelSelectionTask{
				Intent:                "autonomous.general",
				RequiredContextTokens: 1024,
				ExpectedOutputTokens:  256,
			},
		})
		router.localRegistry = registry
		if provider := registry.Provider(); provider != nil {
			router.AddProvider(provider)
			state := registry.State()
			fmt.Printf("✓ Native GGUF candidate selected: %s (%s, context %d)\n", state.SelectedModel.ModelID, state.SelectedModel.Quantization, state.SelectedModel.ContextLength)
		} else {
			state := registry.State()
			fmt.Printf("⚠ Local GGUF configured but unavailable: discovered=%d native_build=%v\n", len(state.DiscoveredModels), backendcap.NativeRuntimeEnabled())
		}
	}

	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		model := localEnvString("ECHO_ANTHROPIC_MODEL", "claude-sonnet-4-5")
		router.AddProvider(NewAnthropicProvider(model))
		fmt.Println("✓ Anthropic Claude provider initialized")
	}
	if os.Getenv("OPENROUTER_API_KEY") != "" {
		model := localEnvString("ECHO_OPENROUTER_MODEL", "anthropic/claude-sonnet-4.5")
		router.AddProvider(NewOpenRouterProvider(model))
		fmt.Println("✓ OpenRouter provider initialized")
	}
	if os.Getenv("OPENAI_API_KEY") != "" {
		model := localEnvString("ECHO_OPENAI_MODEL", "gpt-4o")
		router.AddProvider(NewOpenAIProvider(model))
		fmt.Println("✓ OpenAI provider initialized")
	}

	realProviders := len(router.providers)
	router.AddProvider(&SimpleFallbackProvider{})
	if realProviders == 0 {
		fmt.Println("⚠ Using deterministic fallback provider (no safe real provider available)")
	}
}

// AddProvider appends a provider while preserving configured order within its class.
func (router *MultiProviderLLM) AddProvider(provider Provider) {
	if provider == nil {
		return
	}
	available := provider.Available()
	router.mu.Lock()
	defer router.mu.Unlock()
	router.providers = append(router.providers, provider)
	if router.stats == nil {
		router.stats = make(map[string]*ProviderStats)
	}
	router.stats[provider.Name()] = &ProviderStats{Available: available, LastUsed: time.Now()}
}

// Generate routes a completion through capabilities and preserves caller cancellation.
func (router *MultiProviderLLM) Generate(ctx context.Context, prompt string, options GenerateOptions) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	providers, decision, ordered := router.route(options)
	traceID := options.Routing.TraceID
	if len(ordered) == 0 {
		router.recordDecision(traceID, decision)
		return "", fmt.Errorf("no provider satisfies workload: %s", decision.Reason)
	}
	router.recordDecision(traceID, decision)

	var lastErr error
	for _, name := range ordered {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		provider := providers[name]
		if provider == nil || !provider.Available() {
			router.recordAttempt(traceID, RouteAttempt{Provider: name, Skipped: true, ErrorCategory: "unavailable", Timestamp: time.Now()})
			continue
		}
		capability := capabilityForProvider(provider, 0, router.currentMode())
		attemptCtx, cancel := router.attemptContext(ctx)
		start := time.Now()
		result, err := provider.Generate(attemptCtx, prompt, options)
		cancel()
		latency := time.Since(start)
		router.updateStats(provider.Name(), err == nil, latency, routeErrorCategory(err))
		router.recordAttempt(traceID, RouteAttempt{
			Provider:      provider.Name(),
			BackendKind:   capability.Kind,
			ModelID:       capability.ModelID,
			Success:       err == nil,
			ErrorCategory: routeErrorCategory(err),
			Latency:       latency,
			Timestamp:     time.Now(),
		})
		if err == nil {
			router.recordSuccess(traceID, provider.Name(), capability)
			return result, nil
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return "", ctxErr
		}
		lastErr = err
	}
	if lastErr != nil {
		return "", fmt.Errorf("all routed providers failed: %w", lastErr)
	}
	return "", fmt.Errorf("no routed provider is currently available")
}

// StreamGenerate routes streaming output. Failover is safe only before content is emitted.
func (router *MultiProviderLLM) StreamGenerate(ctx context.Context, prompt string, options GenerateOptions) (<-chan StreamChunk, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	providers, decision, ordered := router.route(options)
	traceID := options.Routing.TraceID
	if len(ordered) == 0 {
		router.recordDecision(traceID, decision)
		return nil, fmt.Errorf("no provider satisfies workload: %s", decision.Reason)
	}
	router.recordDecision(traceID, decision)

	output := make(chan StreamChunk, 32)
	go router.runStreamRoute(ctx, prompt, options, providers, ordered, output)
	return output, nil
}

func (router *MultiProviderLLM) runStreamRoute(ctx context.Context, prompt string, options GenerateOptions, providers map[string]Provider, ordered []string, output chan<- StreamChunk) {
	defer close(output)
	traceID := options.Routing.TraceID
	var lastErr error
	for _, name := range ordered {
		if err := ctx.Err(); err != nil {
			router.sendTerminal(ctx, output, err)
			return
		}
		provider := providers[name]
		if provider == nil || !provider.Available() {
			router.recordAttempt(traceID, RouteAttempt{Provider: name, Skipped: true, ErrorCategory: "unavailable", Timestamp: time.Now()})
			continue
		}
		capability := capabilityForProvider(provider, 0, router.currentMode())
		start := time.Now()
		if streamer, ok := provider.(interface {
			StreamGenerate(context.Context, string, GenerateOptions) (<-chan StreamChunk, error)
		}); ok {
			attemptCtx, cancel := router.attemptContext(ctx)
			stream, err := streamer.StreamGenerate(attemptCtx, prompt, options)
			if err != nil {
				cancel()
				lastErr = err
				router.updateStats(provider.Name(), false, time.Since(start), routeErrorCategory(err))
				router.recordAttempt(traceID, RouteAttempt{Provider: provider.Name(), BackendKind: capability.Kind, ModelID: capability.ModelID, ErrorCategory: routeErrorCategory(err), Latency: time.Since(start), Timestamp: time.Now()})
				if ctx.Err() != nil {
					router.sendTerminal(ctx, output, ctx.Err())
					return
				}
				continue
			}
			emitted, success, streamErr := router.forwardProviderStream(ctx, stream, output)
			cancel()
			latency := time.Since(start)
			router.updateStats(provider.Name(), success, latency, routeErrorCategory(streamErr))
			router.recordAttempt(traceID, RouteAttempt{Provider: provider.Name(), BackendKind: capability.Kind, ModelID: capability.ModelID, Success: success, ErrorCategory: routeErrorCategory(streamErr), Latency: latency, Timestamp: time.Now()})
			if success {
				router.recordSuccess(traceID, provider.Name(), capability)
				return
			}
			lastErr = streamErr
			if emitted {
				router.sendTerminal(ctx, output, streamErr)
				return
			}
			continue
		}

		attemptCtx, cancel := router.attemptContext(ctx)
		result, err := provider.Generate(attemptCtx, prompt, options)
		cancel()
		latency := time.Since(start)
		router.updateStats(provider.Name(), err == nil, latency, routeErrorCategory(err))
		router.recordAttempt(traceID, RouteAttempt{Provider: provider.Name(), BackendKind: capability.Kind, ModelID: capability.ModelID, Success: err == nil, ErrorCategory: routeErrorCategory(err), Latency: latency, Timestamp: time.Now()})
		if err == nil {
			router.recordSuccess(traceID, provider.Name(), capability)
			select {
			case output <- StreamChunk{Content: result}:
			case <-ctx.Done():
				return
			}
			select {
			case output <- StreamChunk{Done: true}:
			case <-ctx.Done():
			}
			return
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no routed provider is currently available")
	}
	router.sendTerminal(ctx, output, fmt.Errorf("all routed providers failed: %w", lastErr))
}

func (router *MultiProviderLLM) forwardProviderStream(ctx context.Context, input <-chan StreamChunk, output chan<- StreamChunk) (bool, bool, error) {
	emitted := false
	for {
		select {
		case <-ctx.Done():
			return emitted, false, ctx.Err()
		case chunk, ok := <-input:
			if !ok {
				if emitted {
					select {
					case output <- StreamChunk{Done: true}:
					case <-ctx.Done():
						return emitted, false, ctx.Err()
					}
					return emitted, true, nil
				}
				return false, false, fmt.Errorf("provider stream closed without output")
			}
			if chunk.Error != nil {
				return emitted, false, chunk.Error
			}
			if chunk.Content != "" {
				emitted = true
			}
			select {
			case output <- chunk:
			case <-ctx.Done():
				return emitted, false, ctx.Err()
			}
			if chunk.Done {
				return emitted, true, nil
			}
		}
	}
}

func (router *MultiProviderLLM) sendTerminal(ctx context.Context, output chan<- StreamChunk, err error) {
	if err == nil {
		return
	}
	select {
	case output <- StreamChunk{Error: err, Done: true}:
	case <-ctx.Done():
	}
}

func (router *MultiProviderLLM) route(options GenerateOptions) (map[string]Provider, backendcap.Decision, []string) {
	router.mu.RLock()
	providers := append([]Provider(nil), router.providers...)
	mode := router.mode
	router.mu.RUnlock()
	if mode == "" {
		mode = ProviderModeBalanced
	}

	providerMap := make(map[string]Provider, len(providers))
	capabilities := make([]backendcap.Capability, 0, len(providers))
	for index, provider := range providers {
		if provider == nil {
			continue
		}
		providerMap[provider.Name()] = provider
		capabilities = append(capabilities, capabilityForProvider(provider, index, mode))
	}
	workload := normalizeRouting(options.Routing, options.MaxTokens, mode)
	decision := backendcap.SelectFromCapabilities(workload, capabilities)
	ordered := make([]string, 0, len(decision.Route))
	for _, name := range decision.Route {
		if _, exists := providerMap[name]; exists {
			ordered = append(ordered, name)
		}
	}
	return providerMap, decision, ordered
}

func (router *MultiProviderLLM) attemptContext(ctx context.Context) (context.Context, context.CancelFunc) {
	timeout := router.attemptTimeout
	if timeout <= 0 {
		timeout = 90 * time.Second
	}
	return context.WithTimeout(ctx, timeout)
}

func (router *MultiProviderLLM) currentMode() ProviderMode {
	router.mu.RLock()
	defer router.mu.RUnlock()
	if router.mode == "" {
		return ProviderModeBalanced
	}
	return router.mode
}

func (router *MultiProviderLLM) recordDecision(traceID string, decision backendcap.Decision) {
	router.mu.Lock()
	defer router.mu.Unlock()
	mode := router.mode
	if mode == "" {
		mode = ProviderModeBalanced
	}
	now := time.Now()
	router.routingState.TraceID = traceID
	router.routingState.Mode = mode
	router.routingState.Decision = cloneDecision(decision)
	router.routingState.Degraded = decision.Degraded
	router.routingState.UpdatedAt = now
	if traceID == "" {
		return
	}
	state := BackendRoutingState{TraceID: traceID, Mode: mode, Decision: cloneDecision(decision), Degraded: decision.Degraded, UpdatedAt: now}
	router.setTraceStateLocked(traceID, state)
}

func (router *MultiProviderLLM) recordAttempt(traceID string, attempt RouteAttempt) {
	router.mu.Lock()
	defer router.mu.Unlock()
	router.routingState.Attempts = append(router.routingState.Attempts, attempt)
	if len(router.routingState.Attempts) > maxRouteAttempts {
		overflow := len(router.routingState.Attempts) - maxRouteAttempts
		router.routingState.Attempts = append([]RouteAttempt(nil), router.routingState.Attempts[overflow:]...)
	}
	now := time.Now()
	router.routingState.UpdatedAt = now
	if traceID == "" {
		return
	}
	state := router.routeResults[traceID]
	state.TraceID = traceID
	state.Attempts = append(state.Attempts, attempt)
	if len(state.Attempts) > maxRouteAttempts {
		overflow := len(state.Attempts) - maxRouteAttempts
		state.Attempts = append([]RouteAttempt(nil), state.Attempts[overflow:]...)
	}
	state.UpdatedAt = now
	router.setTraceStateLocked(traceID, state)
}

func (router *MultiProviderLLM) recordSuccess(traceID, providerName string, capability backendcap.Capability) {
	router.mu.Lock()
	defer router.mu.Unlock()
	now := time.Now()
	router.routingState.SelectedProvider = providerName
	router.routingState.SelectedKind = capability.Kind
	router.routingState.SelectedModelID = capability.ModelID
	if capability.Kind == backendcap.BackendFallback {
		router.routingState.FallbackCount++
		router.routingState.Degraded = true
	}
	router.routingState.UpdatedAt = now
	if traceID == "" {
		return
	}
	state := router.routeResults[traceID]
	state.TraceID = traceID
	state.SelectedProvider = providerName
	state.SelectedKind = capability.Kind
	state.SelectedModelID = capability.ModelID
	if capability.Kind == backendcap.BackendFallback {
		state.FallbackCount++
		state.Degraded = true
	}
	state.UpdatedAt = now
	router.setTraceStateLocked(traceID, state)
}

func (router *MultiProviderLLM) setTraceStateLocked(traceID string, state BackendRoutingState) {
	if router.routeResults == nil {
		router.routeResults = make(map[string]BackendRoutingState)
	}
	if _, exists := router.routeResults[traceID]; !exists {
		router.routeOrder = append(router.routeOrder, traceID)
	}
	router.routeResults[traceID] = state
	if len(router.routeOrder) > maxRouteResults {
		overflow := len(router.routeOrder) - maxRouteResults
		for _, expired := range router.routeOrder[:overflow] {
			delete(router.routeResults, expired)
		}
		router.routeOrder = append([]string(nil), router.routeOrder[overflow:]...)
	}
}

func (router *MultiProviderLLM) updateStats(providerName string, success bool, latency time.Duration, errorCategory string) {
	router.mu.Lock()
	defer router.mu.Unlock()
	if router.stats == nil {
		router.stats = make(map[string]*ProviderStats)
	}
	stats := router.stats[providerName]
	if stats == nil {
		stats = &ProviderStats{}
		router.stats[providerName] = stats
	}
	stats.TotalCalls++
	stats.LastUsed = time.Now()
	stats.TotalLatency += latency
	stats.LastErrorCategory = errorCategory
	if success {
		stats.SuccessCalls++
		stats.Available = true
		stats.LastErrorCategory = ""
	} else {
		stats.FailedCalls++
	}
}

func (router *MultiProviderLLM) GetStats() map[string]*ProviderStats {
	router.mu.RLock()
	defer router.mu.RUnlock()
	result := make(map[string]*ProviderStats, len(router.stats))
	for name, stats := range router.stats {
		copyValue := *stats
		result[name] = &copyValue
	}
	return result
}

// GetBackendState returns path-scrubbed route and local runtime status.
func (router *MultiProviderLLM) GetBackendState() BackendRoutingState {
	router.mu.RLock()
	state := router.routingState
	registry := router.localRegistry
	state.Attempts = cloneAttempts(state.Attempts)
	state.Decision = cloneDecision(state.Decision)
	router.mu.RUnlock()
	if registry != nil {
		state.LocalModel = registry.State()
	}
	return state
}

// GetRouteResult returns the exact path-scrubbed decision for a traced request.
func (router *MultiProviderLLM) GetRouteResult(traceID string) (BackendRoutingState, bool) {
	router.mu.RLock()
	state, exists := router.routeResults[traceID]
	registry := router.localRegistry
	state.Attempts = cloneAttempts(state.Attempts)
	state.Decision = cloneDecision(state.Decision)
	router.mu.RUnlock()
	if !exists {
		return BackendRoutingState{}, false
	}
	if registry != nil {
		state.LocalModel = registry.State()
	}
	return state, true
}

// LocalRuntime returns the optional registry lifecycle controller.
func (router *MultiProviderLLM) LocalRuntime() LocalRuntimeController {
	router.mu.RLock()
	defer router.mu.RUnlock()
	if router.localRegistry == nil {
		return nil
	}
	return router.localRegistry
}

// Close releases the registry-owned native runtime.
func (router *MultiProviderLLM) Close() error {
	router.mu.RLock()
	registry := router.localRegistry
	router.mu.RUnlock()
	if registry != nil {
		return registry.Close()
	}
	return nil
}

func (router *MultiProviderLLM) Name() string { return "MultiProvider" }

func (router *MultiProviderLLM) Available() bool {
	router.mu.RLock()
	providers := append([]Provider(nil), router.providers...)
	router.mu.RUnlock()
	for _, provider := range providers {
		if provider != nil && provider.Available() {
			return true
		}
	}
	return false
}

func (router *MultiProviderLLM) MaxTokens() int {
	router.mu.RLock()
	providers := append([]Provider(nil), router.providers...)
	router.mu.RUnlock()
	maximum := 0
	for _, provider := range providers {
		if provider != nil && provider.Available() && provider.MaxTokens() > maximum {
			maximum = provider.MaxTokens()
		}
	}
	if maximum == 0 {
		return 4096
	}
	return maximum
}

func cloneDecision(decision backendcap.Decision) backendcap.Decision {
	decision.Route = append([]string(nil), decision.Route...)
	decision.Alternatives = append([]backendcap.Capability(nil), decision.Alternatives...)
	decision.Rejections = append([]backendcap.Rejection(nil), decision.Rejections...)
	return decision
}

func splitLocalPathList(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == os.PathListSeparator || r == ',' || r == '\n' || r == '\r'
	})
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = appendUniqueString(result, part)
		}
	}
	return result
}

func appendUniqueString(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

var _ LLMProvider = (*MultiProviderLLM)(nil)
