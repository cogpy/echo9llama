//go:build cgo && !nollama
// +build cgo,!nollama

package llm

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cogpy/echo9llama/core/backendcap"
	llamacpp "github.com/cogpy/echo9llama/llama"
)

var llamaBackendInitOnce sync.Once

// LocalGGUFProvider owns one native model/context pair and serializes access to
// that mutable context without blocking status reads behind token generation.
type LocalGGUFProvider struct {
	stateMu sync.RWMutex
	slot    chan struct{}
	config  LocalGGUFProviderConfig

	model        *llamacpp.Model
	llamaContext *llamacpp.Context
	architecture string
	loaded       bool
	loading      bool
	closed       bool
	loadErr      error
	permanentErr bool
	retryAt      time.Time
	retryDelay   time.Duration
	inFlight     atomic.Int32
}

// NewLocalGGUFProvider creates a provider for a concrete GGUF path.
func NewLocalGGUFProvider(modelPath string) *LocalGGUFProvider {
	return NewLocalGGUFProviderWithConfig(defaultLocalGGUFConfig(modelPath))
}

// NewLocalGGUFProviderFromCapability creates a provider from discovered metadata.
func NewLocalGGUFProviderFromCapability(capability backendcap.Capability) *LocalGGUFProvider {
	config := defaultLocalGGUFConfig(capability.ModelPath)
	config.Capability = capability
	if capability.ContextLength > 0 && config.ContextSize > capability.ContextLength {
		config.ContextSize = capability.ContextLength
	}
	return NewLocalGGUFProviderWithConfig(config)
}

// NewLocalGGUFProviderWithConfig creates a provider with explicit runtime policy.
func NewLocalGGUFProviderWithConfig(config LocalGGUFProviderConfig) *LocalGGUFProvider {
	config = normalizeLocalGGUFConfig(config)
	return &LocalGGUFProvider{
		config:     config,
		slot:       make(chan struct{}, 1),
		retryDelay: time.Second,
	}
}

func (provider *LocalGGUFProvider) Name() string {
	provider.stateMu.RLock()
	defer provider.stateMu.RUnlock()
	return provider.config.Name
}

// Available performs only bounded file/metadata and host-envelope checks. It
// never loads native model memory and never waits for the generation slot.
func (provider *LocalGGUFProvider) Available() bool {
	provider.stateMu.RLock()
	if provider.closed {
		provider.stateMu.RUnlock()
		return false
	}
	capability := provider.config.Capability
	modelPath := provider.config.ModelPath
	options := provider.config.discoveryOptions()
	provider.stateMu.RUnlock()

	if capability.ModelPath != "" {
		if _, err := os.Stat(capability.ModelPath); err == nil {
			return capability.Available && capability.MemorySafe
		}
		return false
	}
	if strings.TrimSpace(modelPath) == "" {
		return false
	}
	probed, err := backendcap.ProbeModelFile(modelPath, options)
	if err != nil {
		provider.recordLoadError(err, true)
		return false
	}
	provider.stateMu.Lock()
	if !provider.closed {
		provider.config.Capability = probed
		if probed.ContextLength > 0 && provider.config.ContextSize > probed.ContextLength {
			provider.config.ContextSize = probed.ContextLength
		}
	}
	provider.stateMu.Unlock()
	return probed.Available && probed.MemorySafe
}

func (provider *LocalGGUFProvider) MaxTokens() int {
	provider.stateMu.RLock()
	defer provider.stateMu.RUnlock()
	return provider.config.ContextSize
}

// CapabilitySnapshot returns a path-scrubbed dynamic capability for routing.
func (provider *LocalGGUFProvider) CapabilitySnapshot() backendcap.Capability {
	provider.stateMu.RLock()
	needsProbe := provider.config.Capability.ModelPath == ""
	provider.stateMu.RUnlock()
	if needsProbe {
		_ = provider.Available()
	}
	provider.stateMu.RLock()
	capability := provider.config.Capability
	closed := provider.closed
	provider.stateMu.RUnlock()
	capability.ModelPath = ""
	capability.CurrentInFlight = int(provider.inFlight.Load())
	capability.MaxConcurrency = 1
	capability.Available = capability.Available && capability.MemorySafe && !closed
	return capability
}

func (provider *LocalGGUFProvider) Generate(ctx context.Context, prompt string, options GenerateOptions) (string, error) {
	var builder strings.Builder
	emitter := newStopSequenceEmitter(options.Stop, func(content string) error {
		builder.WriteString(content)
		return nil
	})
	if err := provider.generate(ctx, prompt, options, emitter); err != nil {
		return "", err
	}
	return strings.TrimSpace(builder.String()), nil
}

func (provider *LocalGGUFProvider) StreamGenerate(ctx context.Context, prompt string, options GenerateOptions) (<-chan StreamChunk, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	output := make(chan StreamChunk, 32)
	go func() {
		defer close(output)
		emitter := newStopSequenceEmitter(options.Stop, func(content string) error {
			if content == "" {
				return nil
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case output <- StreamChunk{Content: content}:
				return nil
			}
		})
		if err := provider.generate(ctx, prompt, options, emitter); err != nil {
			select {
			case output <- StreamChunk{Error: err, Done: true}:
			case <-ctx.Done():
			}
			return
		}
		select {
		case output <- StreamChunk{Done: true}:
		case <-ctx.Done():
		}
	}()
	return output, nil
}

// Warmup loads the selected model without generating tokens.
func (provider *LocalGGUFProvider) Warmup(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	provider.stateMu.RLock()
	queueWait := provider.config.QueueWait
	provider.stateMu.RUnlock()
	if err := provider.acquireSlot(ctx, RoutingOptions{PolicyExplicit: true, AllowQueue: true, QueueWait: queueWait}); err != nil {
		return err
	}
	provider.inFlight.Add(1)
	defer func() {
		provider.inFlight.Add(-1)
		provider.releaseSlot()
	}()
	_, _, err := provider.ensureLoaded(ctx)
	return err
}

func (provider *LocalGGUFProvider) generate(ctx context.Context, prompt string, options GenerateOptions, emitter *stopSequenceEmitter) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := provider.acquireSlot(ctx, options.Routing); err != nil {
		return err
	}
	provider.inFlight.Add(1)
	defer func() {
		provider.inFlight.Add(-1)
		provider.releaseSlot()
	}()

	model, llamaContext, err := provider.ensureLoaded(ctx)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	fullPrompt := prompt
	if options.SystemPrompt != "" {
		fullPrompt = fmt.Sprintf("System: %s\n\nUser: %s\n\nAssistant:", options.SystemPrompt, prompt)
	}
	maxTokens := options.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 256
	}
	temperature := options.Temperature
	if temperature <= 0 {
		temperature = 0.7
	}
	topP := options.TopP
	if topP <= 0 || topP > 1 {
		topP = 0.9
	}

	tokens, err := model.Tokenize(fullPrompt, true, true)
	if err != nil {
		return fmt.Errorf("tokenize local GGUF prompt: %w", err)
	}
	if len(tokens) == 0 {
		return fmt.Errorf("tokenize local GGUF prompt: no tokens")
	}
	contextLimit := provider.MaxTokens()
	if contextLimit < 2 {
		return fmt.Errorf("local GGUF context size %d is too small", contextLimit)
	}
	if maxTokens >= contextLimit {
		maxTokens = contextLimit - 1
	}
	promptBudget := contextLimit - maxTokens
	if promptBudget < 1 {
		promptBudget = 1
	}
	if len(tokens) > promptBudget {
		tokens = tokens[len(tokens)-promptBudget:]
	}
	if maxTokens > contextLimit-len(tokens) {
		maxTokens = contextLimit - len(tokens)
	}
	if maxTokens <= 0 {
		return fmt.Errorf("local GGUF prompt exhausted context window")
	}

	provider.stateMu.RLock()
	batchSize := provider.config.BatchSize
	seed := provider.config.Seed
	provider.stateMu.RUnlock()
	batchSize = maxLocalInt(1, minLocalInt(batchSize, len(tokens)))

	llamaContext.KvCacheClear()
	batch, err := llamacpp.NewBatch(batchSize, 1, 0)
	if err != nil {
		return fmt.Errorf("create local GGUF batch: %w", err)
	}
	defer batch.Free()
	for start := 0; start < len(tokens); start += batchSize {
		end := minLocalInt(start+batchSize, len(tokens))
		batch.Clear()
		for index := start; index < end; index++ {
			batch.Add(tokens[index], nil, index, index == len(tokens)-1, 0)
		}
		if err := llamaContext.Decode(batch); err != nil {
			return fmt.Errorf("decode local GGUF prompt batch %d-%d: %w", start, end, err)
		}
	}
	sampler, err := llamacpp.NewSamplingContext(model, llamacpp.SamplingParams{
		TopK:          40,
		TopP:          float32(topP),
		MinP:          0.05,
		TypicalP:      1.0,
		Temp:          float32(temperature),
		RepeatLastN:   minLocalInt(64, contextLimit),
		PenaltyRepeat: 1.1,
		Seed:          seed,
	})
	if err != nil {
		return fmt.Errorf("initialize local GGUF sampler: %w", err)
	}
	defer sampler.Free()
	for _, token := range tokens {
		sampler.Accept(token, false)
	}

	for index := range maxTokens {
		if err := ctx.Err(); err != nil {
			return err
		}
		token := sampler.Sample(llamaContext, -1)
		if model.TokenIsEog(token) {
			break
		}
		stopped, emitErr := emitter.Write(model.TokenToPiece(token))
		if emitErr != nil {
			return emitErr
		}
		if stopped {
			return nil
		}
		sampler.Accept(token, true)
		batch.Clear()
		batch.Add(token, nil, len(tokens)+index, true, 0)
		if err := llamaContext.Decode(batch); err != nil {
			return fmt.Errorf("decode local GGUF token: %w", err)
		}
	}
	return emitter.Flush()
}

func (provider *LocalGGUFProvider) acquireSlot(ctx context.Context, routing RoutingOptions) error {
	provider.stateMu.RLock()
	closed := provider.closed
	wait := provider.config.QueueWait
	provider.stateMu.RUnlock()
	if closed {
		return ErrLocalGGUFClosed
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if routing.PolicyExplicit {
		if !routing.AllowQueue {
			wait = 0
		} else if routing.QueueWait > 0 {
			wait = routing.QueueWait
		}
	} else if routing.QueueWait > 0 {
		wait = routing.QueueWait
	}
	if wait <= 0 {
		select {
		case provider.slot <- struct{}{}:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		default:
			return ErrLocalGGUFQueueSaturated
		}
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case provider.slot <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return ErrLocalGGUFQueueSaturated
	}
}

func (provider *LocalGGUFProvider) releaseSlot() {
	<-provider.slot
}

func (provider *LocalGGUFProvider) ensureLoaded(ctx context.Context) (*llamacpp.Model, *llamacpp.Context, error) {
	provider.stateMu.Lock()
	if provider.closed {
		provider.stateMu.Unlock()
		return nil, nil, ErrLocalGGUFClosed
	}
	if provider.loaded && provider.model != nil && provider.llamaContext != nil {
		model, llamaContext := provider.model, provider.llamaContext
		provider.stateMu.Unlock()
		return model, llamaContext, nil
	}
	if provider.loadErr != nil && (provider.permanentErr || time.Now().Before(provider.retryAt)) {
		err := provider.loadErr
		provider.stateMu.Unlock()
		return nil, nil, err
	}
	provider.loading = true
	config := provider.config
	provider.stateMu.Unlock()

	if err := ctx.Err(); err != nil {
		provider.finishLoading(nil, nil, "", err, false)
		return nil, nil, err
	}
	capability := config.Capability
	if capability.ModelPath == "" {
		probed, err := backendcap.ProbeModelFile(config.ModelPath, config.discoveryOptions())
		if err != nil {
			provider.finishLoading(nil, nil, "", err, true)
			return nil, nil, err
		}
		capability = probed
	}
	if !capability.Available || !capability.MemorySafe {
		err := fmt.Errorf("%w: %s", ErrLocalGGUFUnavailable, capability.Reason)
		provider.finishLoading(nil, nil, "", err, true)
		return nil, nil, err
	}
	config.Capability = capability
	config.ModelPath = capability.ModelPath
	config = normalizeLocalGGUFConfig(config)

	llamaBackendInitOnce.Do(llamacpp.BackendInit)
	architecture, _ := llamacpp.GetModelArch(config.ModelPath)
	model, err := llamacpp.LoadModelFromFile(config.ModelPath, llamacpp.ModelParams{NumGpuLayers: config.GPULayers, UseMmap: true})
	if err != nil {
		err = fmt.Errorf("load local GGUF model %s: %w", capability.ModelID, err)
		provider.finishLoading(nil, nil, architecture, err, false)
		return nil, nil, err
	}
	llamaContext, err := llamacpp.NewContextWithModel(model, llamacpp.NewContextParams(config.ContextSize, config.BatchSize, 1, config.Threads, config.FlashAttention, config.KVCacheType))
	if err != nil {
		llamacpp.FreeModel(model)
		err = fmt.Errorf("create local GGUF context: %w", err)
		provider.finishLoading(nil, nil, architecture, err, false)
		return nil, nil, err
	}
	if err := ctx.Err(); err != nil {
		llamaContext.Free()
		llamacpp.FreeModel(model)
		provider.finishLoading(nil, nil, architecture, err, false)
		return nil, nil, err
	}
	provider.finishLoading(model, llamaContext, architecture, nil, false)
	return model, llamaContext, nil
}

func (provider *LocalGGUFProvider) finishLoading(model *llamacpp.Model, llamaContext *llamacpp.Context, architecture string, err error, permanent bool) {
	provider.stateMu.Lock()
	defer provider.stateMu.Unlock()
	provider.loading = false
	if provider.closed && model != nil {
		llamaContext.Free()
		llamacpp.FreeModel(model)
		return
	}
	if err != nil {
		provider.loadErr = err
		provider.permanentErr = permanent
		if !permanent {
			if provider.retryDelay <= 0 {
				provider.retryDelay = time.Second
			}
			provider.retryAt = time.Now().Add(provider.retryDelay)
			provider.retryDelay *= 2
			if provider.retryDelay > time.Minute {
				provider.retryDelay = time.Minute
			}
		}
		return
	}
	provider.model = model
	provider.llamaContext = llamaContext
	provider.architecture = architecture
	provider.loaded = true
	provider.loadErr = nil
	provider.permanentErr = false
	provider.retryAt = time.Time{}
	provider.retryDelay = time.Second
}

func (provider *LocalGGUFProvider) recordLoadError(err error, permanent bool) {
	provider.stateMu.Lock()
	defer provider.stateMu.Unlock()
	if provider.loadErr == nil || permanent {
		provider.loadErr = err
		provider.permanentErr = permanent
	}
}

// State returns a scrubbed snapshot without probing or loading native state.
func (provider *LocalGGUFProvider) State() LocalGGUFState {
	provider.stateMu.RLock()
	defer provider.stateMu.RUnlock()
	capability := provider.config.Capability
	state := LocalGGUFState{
		Name:                 provider.config.Name,
		ModelID:              capability.ModelID,
		Architecture:         provider.architecture,
		ContextSize:          provider.config.ContextSize,
		Quantization:         capability.Quantization,
		EstimatedMemoryBytes: capability.EstimatedMemoryBytes,
		NativeBuild:          true,
		Available:            capability.Available && capability.MemorySafe && !provider.closed,
		Loaded:               provider.loaded,
		Loading:              provider.loading,
		Closed:               provider.closed,
		InFlight:             int(provider.inFlight.Load()),
		RetryAt:              provider.retryAt,
	}
	if provider.loadErr != nil {
		state.LoadError = provider.loadErr.Error()
	}
	return state
}

func (provider *LocalGGUFProvider) Loaded() bool {
	return provider.State().Loaded
}

func (provider *LocalGGUFProvider) LoadError() error {
	provider.stateMu.RLock()
	defer provider.stateMu.RUnlock()
	return provider.loadErr
}

// Close drains the active operation and releases all native state idempotently.
func (provider *LocalGGUFProvider) Close() error {
	provider.slot <- struct{}{}
	defer provider.releaseSlot()
	provider.stateMu.Lock()
	defer provider.stateMu.Unlock()
	if provider.closed && provider.model == nil && provider.llamaContext == nil {
		return nil
	}
	provider.closed = true
	if provider.llamaContext != nil {
		provider.llamaContext.Free()
		provider.llamaContext = nil
	}
	if provider.model != nil {
		llamacpp.FreeModel(provider.model)
		provider.model = nil
	}
	provider.loaded = false
	provider.loading = false
	return nil
}

// Reset unloads the provider and permits a later explicit retry after refresh.
func (provider *LocalGGUFProvider) Reset() error {
	if err := provider.Close(); err != nil {
		return err
	}
	provider.stateMu.Lock()
	provider.closed = false
	provider.loadErr = nil
	provider.permanentErr = false
	provider.retryAt = time.Time{}
	provider.retryDelay = time.Second
	provider.stateMu.Unlock()
	return nil
}

type stopSequenceEmitter struct {
	stops   []string
	pending string
	keep    int
	emit    func(string) error
	stopped bool
}

func newStopSequenceEmitter(stops []string, emit func(string) error) *stopSequenceEmitter {
	filtered := make([]string, 0, len(stops))
	longest := 0
	for _, stop := range stops {
		if stop == "" {
			continue
		}
		filtered = append(filtered, stop)
		if len(stop) > longest {
			longest = len(stop)
		}
	}
	return &stopSequenceEmitter{stops: filtered, keep: maxLocalInt(0, longest-1), emit: emit}
}

func (emitter *stopSequenceEmitter) Write(content string) (bool, error) {
	if emitter.stopped {
		return true, nil
	}
	emitter.pending += content
	if index := earliestStopIndex(emitter.pending, emitter.stops); index >= 0 {
		if index > 0 && emitter.emit != nil {
			if err := emitter.emit(emitter.pending[:index]); err != nil {
				return false, err
			}
		}
		emitter.pending = ""
		emitter.stopped = true
		return true, nil
	}
	if emitter.keep == 0 {
		if emitter.pending != "" && emitter.emit != nil {
			if err := emitter.emit(emitter.pending); err != nil {
				return false, err
			}
		}
		emitter.pending = ""
		return false, nil
	}
	if len(emitter.pending) > emitter.keep {
		flushLength := len(emitter.pending) - emitter.keep
		if emitter.emit != nil {
			if err := emitter.emit(emitter.pending[:flushLength]); err != nil {
				return false, err
			}
		}
		emitter.pending = emitter.pending[flushLength:]
	}
	return false, nil
}

func (emitter *stopSequenceEmitter) Flush() error {
	if emitter.stopped || emitter.pending == "" || emitter.emit == nil {
		emitter.pending = ""
		return nil
	}
	err := emitter.emit(emitter.pending)
	emitter.pending = ""
	return err
}

func earliestStopIndex(content string, stops []string) int {
	earliest := -1
	for _, stop := range stops {
		if index := strings.Index(content, stop); index >= 0 && (earliest < 0 || index < earliest) {
			earliest = index
		}
	}
	return earliest
}

var (
	_ LLMProvider = (*LocalGGUFProvider)(nil)
	_ interface {
		CapabilitySnapshot() backendcap.Capability
	} = (*LocalGGUFProvider)(nil)
)
