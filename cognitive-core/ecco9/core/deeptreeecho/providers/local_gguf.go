//go:build simple
// +build simple

package providers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/EchoCog/echollama/core/deeptreeecho"
)

// LocalGGUFProvider implements ModelProvider for local GGUF models
type LocalGGUFProvider struct {
	modelsPath  string
	loadedModel string
	modelInfo   map[string]interface{}
	available   bool
}

// NewLocalGGUFProvider creates a new local GGUF provider
func NewLocalGGUFProvider() *LocalGGUFProvider {
	return &LocalGGUFProvider{
		modelsPath: "models",
		modelInfo:  make(map[string]interface{}),
		available:  false,
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
	p.modelInfo["name"] = modelName
	p.modelInfo["path"] = modelPath

	// Get file size
	if info, err := os.Stat(modelPath); err == nil {
		p.modelInfo["size_mb"] = info.Size() / (1024 * 1024)
	}

	p.modelInfo["status"] = "loaded successfully"
	return nil
}

// Generate implements ModelProvider.Generate
func (p *LocalGGUFProvider) Generate(ctx context.Context, prompt string, options deeptreeecho.GenerateOptions) (string, error) {
	if p.loadedModel == "" {
		// Try to load a default model
		models := p.ListAvailableModels()
		if len(models) > 0 {
			if err := p.LoadModel(models[0]); err != nil {
				return "", fmt.Errorf("no model loaded and failed to load default: %v", err)
			}
		} else {
			return "", fmt.Errorf("no GGUF models available")
		}
	}

	// No inference backend is linked into this provider; it never fabricates
	// text in place of a model.
	return "", fmt.Errorf("local GGUF inference is not available in this build (model %s is loaded but cannot be run)", p.loadedModel)
}

// getModelSize returns the size of the loaded model in MB
func (p *LocalGGUFProvider) getModelSize() int {
	if path, ok := p.modelInfo["path"].(string); ok {
		if info, err := os.Stat(path); err == nil {
			return int(info.Size() / (1024 * 1024))
		}
	}
	return 0
}

// GenerateStream implements ModelProvider.GenerateStream. There is no
// inference backend to stream from, so it reports Generate's error rather
// than splitting a finished reply into words.
func (p *LocalGGUFProvider) GenerateStream(ctx context.Context, prompt string, options deeptreeecho.GenerateOptions) (<-chan string, error) {
	if _, err := p.Generate(ctx, prompt, options); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("streaming is not available in this build")
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

// Embeddings implements ModelProvider.Embeddings
func (p *LocalGGUFProvider) Embeddings(ctx context.Context, text string) ([]float64, error) {
	return nil, fmt.Errorf("local GGUF embeddings are not available in this build")
}

// GetInfo implements ModelProvider.GetInfo
func (p *LocalGGUFProvider) GetInfo() deeptreeecho.ProviderInfo {
	models := p.ListAvailableModels()

	return deeptreeecho.ProviderInfo{
		Name:         "Local GGUF",
		Description:  "Local GGUF model files (llama.cpp format)",
		Models:       models,
		Capabilities: []string{},
	}
}

// IsAvailable implements ModelProvider.IsAvailable
// Inference is not linked into this provider, so it never reports itself as
// able to answer, even when GGUF files are present.
func (p *LocalGGUFProvider) IsAvailable() bool {
	return false
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
