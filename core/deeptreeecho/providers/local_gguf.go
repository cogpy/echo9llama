//go:build simple
// +build simple

package providers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cogpy/echo9llama/core/deeptreeecho"
	"github.com/cogpy/echo9llama/core/llm"
)

// LocalGGUFProvider implements ModelProvider for local GGUF models
type LocalGGUFProvider struct {
	modelsPath  string
	loadedModel string
	modelInfo   map[string]interface{}
	available   bool
	sel         sync.Mutex // guards loadedModel and modelInfo
	rt          nativeRuntime
}

// NewLocalGGUFProvider creates a new local GGUF provider
func NewLocalGGUFProvider() *LocalGGUFProvider {
	return &LocalGGUFProvider{
		modelsPath: "models",
		modelInfo:  make(map[string]interface{}),
		available:  true,
	}
}

// LoadModel loads a GGUF model from disk
func (p *LocalGGUFProvider) LoadModel(modelName string) error {
	modelPath := filepath.Join(p.modelsPath, modelName)

	// Check if file exists
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return fmt.Errorf("model file not found: %s", modelPath)
	}

	// For now, just mark the model as loaded without trying to read GGUF metadata
	// (GGUF reading seems to hang in this environment)
	p.rt.reset() // closes the previous model once no request uses it
	p.sel.Lock()
	defer p.sel.Unlock()
	p.loadedModel = modelName
	p.modelInfo["name"] = modelName
	p.modelInfo["path"] = modelPath

	// Get file size
	if info, err := os.Stat(modelPath); err == nil {
		p.modelInfo["size_mb"] = info.Size() / (1024 * 1024)
	}

	p.modelInfo["status"] = "selected"
	return nil
}

// Generate implements ModelProvider.Generate by running the selected GGUF
// model in-process through llama.cpp (core/llm.LocalGGUFProvider).
func (p *LocalGGUFProvider) Generate(ctx context.Context, prompt string, options deeptreeecho.GenerateOptions) (string, error) {
	path, err := p.selectedPath()
	if err != nil {
		return "", err
	}
	return p.rt.generate(ctx, path, prompt, llmOptions(options))
}

// selectedPath returns the selected model file, selecting the first
// available one when none is.
func (p *LocalGGUFProvider) selectedPath() (string, error) {
	p.sel.Lock()
	selected := p.loadedModel != ""
	p.sel.Unlock()
	if !selected {
		models := p.ListAvailableModels()
		if len(models) == 0 {
			return "", fmt.Errorf("no GGUF models available in %s", p.modelsPath)
		}
		if err := p.LoadModel(models[0]); err != nil {
			return "", fmt.Errorf("no model loaded and failed to load default: %v", err)
		}
	}
	p.sel.Lock()
	defer p.sel.Unlock()
	path, _ := p.modelInfo["path"].(string)
	return path, nil
}

func llmOptions(o deeptreeecho.GenerateOptions) llm.GenerateOptions {
	opts := llm.DefaultGenerateOptions()
	if o.Temperature > 0 {
		opts.Temperature = o.Temperature
	}
	if o.MaxTokens > 0 {
		opts.MaxTokens = o.MaxTokens
	}
	return opts
}

// GenerateStream implements ModelProvider.GenerateStream with the model's
// own token stream. If the stream ends early, StreamErr says why.
func (p *LocalGGUFProvider) GenerateStream(ctx context.Context, prompt string, options deeptreeecho.GenerateOptions) (<-chan string, error) {
	path, err := p.selectedPath()
	if err != nil {
		return nil, err
	}
	return p.rt.stream(ctx, path, prompt, llmOptions(options))
}

// StreamErr reports why the last GenerateStream ended early, or nil.
func (p *LocalGGUFProvider) StreamErr() error { return p.rt.StreamErr() }

// Chat implements ModelProvider.Chat
func (p *LocalGGUFProvider) Chat(ctx context.Context, messages []deeptreeecho.ChatMessage, options deeptreeecho.ChatOptions) (string, error) {
	// Build context from messages
	var prompt strings.Builder
	for _, msg := range messages {
		prompt.WriteString(fmt.Sprintf("[%s]: %s\n", msg.Role, msg.Content))
	}

	return p.Generate(ctx, prompt.String(), options.GenerateOptions)
}

// ChatStream implements ModelProvider.ChatStream
func (p *LocalGGUFProvider) ChatStream(ctx context.Context, messages []deeptreeecho.ChatMessage, options deeptreeecho.ChatOptions) (<-chan string, error) {
	var prompt strings.Builder
	for _, msg := range messages {
		prompt.WriteString(fmt.Sprintf("[%s]: %s\n", msg.Role, msg.Content))
	}

	return p.GenerateStream(ctx, prompt.String(), options.GenerateOptions)
}

// Embeddings implements ModelProvider.Embeddings. This provider does not
// extract embeddings from the model, so it reports that instead of
// returning pseudo-vectors.
func (p *LocalGGUFProvider) Embeddings(ctx context.Context, text string) ([]float64, error) {
	return nil, fmt.Errorf("embeddings are not supported by the local GGUF provider")
}

// GetInfo implements ModelProvider.GetInfo
func (p *LocalGGUFProvider) GetInfo() deeptreeecho.ProviderInfo {
	models := p.ListAvailableModels()

	return deeptreeecho.ProviderInfo{
		Name:        "Local GGUF",
		Description: "Local GGUF model files (llama.cpp format)",
		Models:      models,
		Capabilities: []string{
			"generation",
			"streaming",
			"chat",
			"embeddings",
			"offline",
		},
	}
}

// IsAvailable implements ModelProvider.IsAvailable
func (p *LocalGGUFProvider) IsAvailable() bool {
	// Check if we have any GGUF models
	models := p.ListAvailableModels()
	return len(models) > 0
}

// ListAvailableModels returns list of available GGUF models
func (p *LocalGGUFProvider) ListAvailableModels() []string {
	var models []string

	files, err := os.ReadDir(p.modelsPath)
	if err != nil {
		return models
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".gguf") {
			models = append(models, file.Name())
		}
	}

	return models
}

// GetLoadedModel returns the currently loaded model name
func (p *LocalGGUFProvider) GetLoadedModel() string {
	p.sel.Lock()
	defer p.sel.Unlock()
	return p.loadedModel
}

// GetModelInfo returns information about the loaded model
func (p *LocalGGUFProvider) GetModelInfo() map[string]interface{} {
	p.sel.Lock()
	defer p.sel.Unlock()
	info := make(map[string]interface{}, len(p.modelInfo))
	for k, v := range p.modelInfo {
		info[k] = v
	}
	return info
}
