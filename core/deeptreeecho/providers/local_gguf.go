//go:build simple
// +build simple

package providers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cogpy/echo9llama/core/deeptreeecho"
	"github.com/cogpy/echo9llama/core/llm"
)

// LocalGGUFProvider implements ModelProvider for local GGUF models
type LocalGGUFProvider struct {
	modelsPath  string
	loadedModel string
	modelInfo   map[string]interface{}
	available   bool
	native      *llm.LocalGGUFProvider
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
	p.loadedModel = modelName
	p.native = nil
	p.modelInfo["name"] = modelName
	p.modelInfo["path"] = modelPath

	// Get file size
	if info, err := os.Stat(modelPath); err == nil {
		p.modelInfo["size_mb"] = info.Size() / (1024 * 1024)
	}

	p.modelInfo["status"] = "selected"
	return nil
}

// Generate implements ModelProvider.Generate by running the loaded GGUF
// model in-process through llama.cpp (core/llm.LocalGGUFProvider).
func (p *LocalGGUFProvider) Generate(ctx context.Context, prompt string, options deeptreeecho.GenerateOptions) (string, error) {
	model, err := p.runtime()
	if err != nil {
		return "", err
	}
	return model.Generate(ctx, prompt, llmOptions(options))
}

// runtime loads (once) the native model for the selected GGUF file.
func (p *LocalGGUFProvider) runtime() (*llm.LocalGGUFProvider, error) {
	if p.loadedModel == "" {
		models := p.ListAvailableModels()
		if len(models) == 0 {
			return nil, fmt.Errorf("no GGUF models available in %s", p.modelsPath)
		}
		if err := p.LoadModel(models[0]); err != nil {
			return nil, fmt.Errorf("no model loaded and failed to load default: %v", err)
		}
	}
	if p.native == nil {
		path, _ := p.modelInfo["path"].(string)
		p.native = llm.NewLocalGGUFProvider(path)
	}
	if !p.native.Available() {
		return nil, fmt.Errorf("local GGUF model %s is not served: native llama.cpp bindings unavailable", p.loadedModel)
	}
	return p.native, nil
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
// own token stream.
func (p *LocalGGUFProvider) GenerateStream(ctx context.Context, prompt string, options deeptreeecho.GenerateOptions) (<-chan string, error) {
	model, err := p.runtime()
	if err != nil {
		return nil, err
	}
	chunks, err := model.StreamGenerate(ctx, prompt, llmOptions(options))
	if err != nil {
		return nil, err
	}
	ch := make(chan string, 16)
	go func() {
		defer close(ch)
		for chunk := range chunks {
			if chunk.Error != nil {
				return
			}
			if chunk.Content != "" {
				select {
				case <-ctx.Done():
					return
				case ch <- chunk.Content:
				}
			}
		}
	}()
	return ch, nil
}

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
	return p.loadedModel
}

// GetModelInfo returns information about the loaded model
func (p *LocalGGUFProvider) GetModelInfo() map[string]interface{} {
	return p.modelInfo
}
