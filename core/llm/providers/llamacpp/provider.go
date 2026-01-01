//go:build ignore
// +build ignore

// Package llamacpp provides a local LLM provider implementation using llama.cpp.
// This enables direct, on-device inference without external API dependencies.
package llamacpp

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
	llamacpp "github.com/cogpy/echo9llama/llama"
)

// LlamaCppProvider implements the llm.LLMProvider interface using llama.cpp.
// It provides local, hardware-accelerated LLM inference.
type LlamaCppProvider struct {
	mu sync.RWMutex
	
	config Config
	model  *llamacpp.Model
	ctx    *llamacpp.Context
	
	// Model metadata
	modelPath    string
	architecture string
	
	// State
	isLoaded bool
}

// Config holds the configuration for LlamaCppProvider.
type Config struct {
	// Name is the identifier for this provider instance
	Name string `json:"name"`
	
	// ModelPath is the path to the GGUF model file
	ModelPath string `json:"model_path"`
	
	// ContextSize is the context window size
	ContextSize int `json:"context_size"`
	
	// BatchSize for processing tokens
	BatchSize int `json:"batch_size"`
	
	// Threads for CPU inference
	Threads int `json:"threads"`
	
	// GPU layers to offload (-1 for auto, 0 for CPU only)
	GPULayers int `json:"gpu_layers"`
	
	// FlashAttention enables flash attention optimization
	FlashAttention bool `json:"flash_attention"`
	
	// KVCacheType specifies the KV cache quantization ("f16", "q8_0", "q4_0")
	KVCacheType string `json:"kv_cache_type"`
	
	// Seed for reproducible generation (-1 for random)
	Seed int `json:"seed"`
}

// DefaultConfig returns a sensible default configuration.
func DefaultConfig(modelPath string) Config {
	return Config{
		Name:           "llamacpp",
		ModelPath:      modelPath,
		ContextSize:    4096,
		BatchSize:      512,
		Threads:        8,
		GPULayers:      -1, // Auto-detect
		FlashAttention: true,
		KVCacheType:    "f16",
		Seed:           -1,
	}
}

// NewLlamaCppProvider creates a new llama.cpp-based LLM provider.
func NewLlamaCppProvider(config Config) (*LlamaCppProvider, error) {
	if config.ModelPath == "" {
		return nil, fmt.Errorf("model path cannot be empty")
	}
	if config.Name == "" {
		config.Name = "llamacpp"
	}
	if config.ContextSize <= 0 {
		config.ContextSize = 4096
	}
	if config.BatchSize <= 0 {
		config.BatchSize = 512
	}
	if config.Threads <= 0 {
		config.Threads = 8
	}

	provider := &LlamaCppProvider{
		config:    config,
		modelPath: config.ModelPath,
		isLoaded:  false,
	}

	// Initialize llama.cpp backend
	llamacpp.BackendInit()

	// Load the model
	if err := provider.loadModel(); err != nil {
		return nil, fmt.Errorf("failed to load model: %w", err)
	}

	return provider, nil
}

// loadModel loads the GGUF model and creates a context.
func (p *LlamaCppProvider) loadModel() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Get model architecture
	arch, err := llamacpp.GetModelArch(p.config.ModelPath)
	if err != nil {
		return fmt.Errorf("failed to get model architecture: %w", err)
	}
	p.architecture = arch

	// Load model with parameters
	modelParams := llamacpp.NewModelParams(p.config.GPULayers)
	model, err := llamacpp.LoadModelFromFile(p.config.ModelPath, modelParams)
	if err != nil {
		return fmt.Errorf("failed to load model file: %w", err)
	}
	p.model = model

	// Create context
	contextParams := llamacpp.NewContextParams(
		p.config.ContextSize,
		p.config.BatchSize,
		1, // numSeqMax
		p.config.Threads,
		p.config.FlashAttention,
		p.config.KVCacheType,
	)
	
	ctx, err := llamacpp.NewContext(model, contextParams)
	if err != nil {
		model.Free()
		return fmt.Errorf("failed to create context: %w", err)
	}
	p.ctx = ctx

	p.isLoaded = true
	return nil
}

// Name returns the provider's name.
func (p *LlamaCppProvider) Name() string {
	return p.config.Name
}

// Available checks if the model is loaded and ready.
func (p *LlamaCppProvider) Available() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.isLoaded && p.model != nil && p.ctx != nil
}

// MaxTokens returns the context size.
func (p *LlamaCppProvider) MaxTokens() int {
	return p.config.ContextSize
}

// Generate produces a completion for the given prompt.
func (p *LlamaCppProvider) Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isLoaded {
		return "", fmt.Errorf("model not loaded")
	}

	// Build full prompt with system message if provided
	fullPrompt := prompt
	if opts.SystemPrompt != "" {
		fullPrompt = fmt.Sprintf("System: %s\n\nUser: %s\n\nAssistant:", opts.SystemPrompt, prompt)
	}

	// Tokenize the prompt
	tokens := p.model.Tokenize(fullPrompt, true, true)
	if len(tokens) == 0 {
		return "", fmt.Errorf("failed to tokenize prompt")
	}

	// Clear KV cache for fresh generation
	p.ctx.KvCacheClear()

	// Create batch for prompt processing
	batch := llamacpp.NewBatch(len(tokens), 0, 1)
	defer batch.Free()

	// Add tokens to batch
	for i, token := range tokens {
		batch.Add(token, i, []int{0}, false)
	}

	// Process the prompt
	if err := p.ctx.Decode(batch); err != nil {
		return "", fmt.Errorf("failed to decode prompt: %w", err)
	}

	// Generate tokens
	var generated []int
	maxTokens := opts.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 500
	}

	// Sampling parameters
	samplingParams := llamacpp.NewSamplingParams()
	samplingParams.Temperature = float32(opts.Temperature)
	samplingParams.TopP = float32(opts.TopP)
	if p.config.Seed >= 0 {
		samplingParams.Seed = uint32(p.config.Seed)
	}

	sampler := llamacpp.NewSampling(p.model, samplingParams)
	defer sampler.Free()

	// Generation loop
	for i := 0; i < maxTokens; i++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		// Sample next token
		token := sampler.Sample(p.ctx, -1)
		
		// Check for EOS or stop sequences
		if p.model.TokenIsEog(token) {
			break
		}

		// Check stop sequences
		piece := p.model.TokenToPiece(token)
		if p.shouldStop(piece, opts.Stop) {
			break
		}

		generated = append(generated, token)

		// Prepare next batch
		batch.Clear()
		batch.Add(token, len(tokens)+i, []int{0}, true)

		// Decode
		if err := p.ctx.Decode(batch); err != nil {
			return "", fmt.Errorf("failed to decode token: %w", err)
		}

		// Accept the token in sampler
		sampler.Accept(token, true)
	}

	// Detokenize the generated tokens
	result := p.model.Detokenize(generated)
	return strings.TrimSpace(result), nil
}

// shouldStop checks if generation should stop based on stop sequences.
func (p *LlamaCppProvider) shouldStop(piece string, stopSequences []string) bool {
	for _, stop := range stopSequences {
		if strings.Contains(piece, stop) {
			return true
		}
	}
	return false
}

// StreamGenerate produces a streaming completion.
func (p *LlamaCppProvider) StreamGenerate(ctx context.Context, prompt string, opts llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	outChan := make(chan llm.StreamChunk, 10)

	go func() {
		defer close(outChan)

		p.mu.Lock()
		defer p.mu.Unlock()

		if !p.isLoaded {
			outChan <- llm.StreamChunk{Error: fmt.Errorf("model not loaded")}
			return
		}

		// Build full prompt
		fullPrompt := prompt
		if opts.SystemPrompt != "" {
			fullPrompt = fmt.Sprintf("System: %s\n\nUser: %s\n\nAssistant:", opts.SystemPrompt, prompt)
		}

		// Tokenize
		tokens := p.model.Tokenize(fullPrompt, true, true)
		if len(tokens) == 0 {
			outChan <- llm.StreamChunk{Error: fmt.Errorf("failed to tokenize prompt")}
			return
		}

		// Clear KV cache
		p.ctx.KvCacheClear()

		// Process prompt
		batch := llamacpp.NewBatch(len(tokens), 0, 1)
		defer batch.Free()

		for i, token := range tokens {
			batch.Add(token, i, []int{0}, false)
		}

		if err := p.ctx.Decode(batch); err != nil {
			outChan <- llm.StreamChunk{Error: fmt.Errorf("failed to decode prompt: %w", err)}
			return
		}

		// Sampling setup
		maxTokens := opts.MaxTokens
		if maxTokens <= 0 {
			maxTokens = 500
		}

		samplingParams := llamacpp.NewSamplingParams()
		samplingParams.Temperature = float32(opts.Temperature)
		samplingParams.TopP = float32(opts.TopP)
		if p.config.Seed >= 0 {
			samplingParams.Seed = uint32(p.config.Seed)
		}

		sampler := llamacpp.NewSampling(p.model, samplingParams)
		defer sampler.Free()

		// Generation loop with streaming
		for i := 0; i < maxTokens; i++ {
			select {
			case <-ctx.Done():
				outChan <- llm.StreamChunk{Error: ctx.Err()}
				return
			default:
			}

			token := sampler.Sample(p.ctx, -1)
			
			if p.model.TokenIsEog(token) {
				outChan <- llm.StreamChunk{Done: true}
				return
			}

			piece := p.model.TokenToPiece(token)
			if p.shouldStop(piece, opts.Stop) {
				outChan <- llm.StreamChunk{Done: true}
				return
			}

			// Stream the token
			outChan <- llm.StreamChunk{Content: piece, Done: false}

			// Continue generation
			batch.Clear()
			batch.Add(token, len(tokens)+i, []int{0}, true)

			if err := p.ctx.Decode(batch); err != nil {
				outChan <- llm.StreamChunk{Error: fmt.Errorf("failed to decode token: %w", err)}
				return
			}

			sampler.Accept(token, true)
		}

		outChan <- llm.StreamChunk{Done: true}
	}()

	return outChan, nil
}

// GetEmbedding generates an embedding for the given text.
// This is useful for the memory system's EmbeddingProvider interface.
func (p *LlamaCppProvider) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isLoaded {
		return nil, fmt.Errorf("model not loaded")
	}

	// Tokenize
	tokens := p.model.Tokenize(text, true, false)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("failed to tokenize text")
	}

	// Clear KV cache
	p.ctx.KvCacheClear()

	// Create batch
	batch := llamacpp.NewBatch(len(tokens), 0, 1)
	defer batch.Free()

	for i, token := range tokens {
		batch.Add(token, i, []int{0}, false)
	}

	// Decode
	if err := p.ctx.Decode(batch); err != nil {
		return nil, fmt.Errorf("failed to decode: %w", err)
	}

	// Get embeddings
	embeddings := p.ctx.GetEmbeddingsSeq(0)
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return embeddings, nil
}

// Close frees the model and context resources.
func (p *LlamaCppProvider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ctx != nil {
		p.ctx.Free()
		p.ctx = nil
	}

	if p.model != nil {
		p.model.Free()
		p.model = nil
	}

	p.isLoaded = false
	return nil
}

// ModelInfo returns information about the loaded model.
func (p *LlamaCppProvider) ModelInfo() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	info := map[string]interface{}{
		"name":            p.config.Name,
		"model_path":      p.modelPath,
		"architecture":    p.architecture,
		"context_size":    p.config.ContextSize,
		"batch_size":      p.config.BatchSize,
		"gpu_layers":      p.config.GPULayers,
		"flash_attention": p.config.FlashAttention,
		"loaded":          p.isLoaded,
	}

	if p.model != nil {
		info["vocab_size"] = p.model.VocabSize()
		info["n_ctx_train"] = p.model.NCtxTrain()
	}

	return info
}
