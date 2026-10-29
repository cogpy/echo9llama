//go:build simple
// +build simple

package providers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cogpy/echo9llama/core/deeptreeecho"
)

// AppStorageProvider implements ModelProvider for Replit App Storage
type AppStorageProvider struct {
	bucketID     string
	localCache   string
	cachedModels map[string]string // model name -> local path
	loadedModel  string
	modelInfo    map[string]interface{}
	sel          sync.Mutex // guards loadedModel and modelInfo
	rt           nativeRuntime
}

// NewAppStorageProvider creates a new App Storage provider
func NewAppStorageProvider() *AppStorageProvider {
	bucketID := os.Getenv("REPLIT_OBJSTORE_BUCKET")
	if bucketID == "" {
		bucketID = "replit-objstore-16fee67f-aa23-4195-8eac-85a4289c2e1a"
	}

	return &AppStorageProvider{
		bucketID:     bucketID,
		localCache:   "/tmp/model_cache",
		cachedModels: make(map[string]string),
		modelInfo:    make(map[string]interface{}),
	}
}

// ListStorageModels lists the GGUF models present in the local cache
// (downloaded with DownloadFromURL). Only files that exist are listed.
func (p *AppStorageProvider) ListStorageModels() ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(p.localCache, "*.gguf"))
	if err != nil {
		return nil, err
	}
	models := make([]string, 0, len(matches))
	for _, m := range matches {
		models = append(models, filepath.Base(m))
	}
	return models, nil
}

// listStorageModelsOrNil lists cached models, or none when the cache
// cannot be read.
func (p *AppStorageProvider) listStorageModelsOrNil() []string {
	models, err := p.ListStorageModels()
	if err != nil {
		return nil
	}
	return models
}

// DownloadModel returns the cached path of a model. A model that has not
// been downloaded (DownloadFromURL) is an error; no placeholder is created.
func (p *AppStorageProvider) DownloadModel(modelName string) (string, error) {
	if cachedPath, exists := p.cachedModels[modelName]; exists {
		if _, err := os.Stat(cachedPath); err == nil {
			return cachedPath, nil
		}
	}
	localPath := filepath.Join(p.localCache, modelName)
	if _, err := os.Stat(localPath); err != nil {
		return "", fmt.Errorf("model %s is not in the local cache %s: download it with DownloadFromURL", modelName, p.localCache)
	}
	p.cachedModels[modelName] = localPath
	return localPath, nil
}

// LoadModel loads a model from App Storage
func (p *AppStorageProvider) LoadModel(modelName string) error {
	// Download if needed
	localPath, err := p.DownloadModel(modelName)
	if err != nil {
		return err
	}

	p.rt.reset() // closes the previous model once no request uses it
	p.sel.Lock()
	defer p.sel.Unlock()
	p.loadedModel = modelName
	p.modelInfo["name"] = modelName
	p.modelInfo["path"] = localPath
	p.modelInfo["bucket"] = p.bucketID
	p.modelInfo["cached"] = true

	// Get file size
	if info, err := os.Stat(localPath); err == nil {
		p.modelInfo["size_mb"] = info.Size() / (1024 * 1024)
	}

	p.modelInfo["status"] = "selected from the App Storage cache"
	return nil
}

// Generate implements ModelProvider.Generate by running the selected GGUF
// model in-process through llama.cpp (core/llm.LocalGGUFProvider).
func (p *AppStorageProvider) Generate(ctx context.Context, prompt string, options deeptreeecho.GenerateOptions) (string, error) {
	path, err := p.selectedPath()
	if err != nil {
		return "", err
	}
	return p.rt.generate(ctx, path, prompt, llmOptions(options))
}

// selectedPath returns the selected model file, selecting the first
// available one when none is.
func (p *AppStorageProvider) selectedPath() (string, error) {
	p.sel.Lock()
	selected := p.loadedModel != ""
	p.sel.Unlock()
	if !selected {
		models := p.listStorageModelsOrNil()
		if len(models) == 0 {
			return "", fmt.Errorf("no models in the App Storage cache %s", p.localCache)
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

// GenerateStream implements ModelProvider.GenerateStream with the model's
// own token stream. If the stream ends early, StreamErr says why.
func (p *AppStorageProvider) GenerateStream(ctx context.Context, prompt string, options deeptreeecho.GenerateOptions) (<-chan string, error) {
	path, err := p.selectedPath()
	if err != nil {
		return nil, err
	}
	return p.rt.stream(ctx, path, prompt, llmOptions(options))
}

// StreamErr reports why the last GenerateStream ended early, or nil.
func (p *AppStorageProvider) StreamErr() error { return p.rt.StreamErr() }

// Chat implements ModelProvider.Chat
func (p *AppStorageProvider) Chat(ctx context.Context, messages []deeptreeecho.ChatMessage, options deeptreeecho.ChatOptions) (string, error) {
	var prompt strings.Builder
	for _, msg := range messages {
		prompt.WriteString(fmt.Sprintf("[%s]: %s\n", msg.Role, msg.Content))
	}

	return p.Generate(ctx, prompt.String(), options.GenerateOptions)
}

// ChatStream implements ModelProvider.ChatStream
func (p *AppStorageProvider) ChatStream(ctx context.Context, messages []deeptreeecho.ChatMessage, options deeptreeecho.ChatOptions) (<-chan string, error) {
	var prompt strings.Builder
	for _, msg := range messages {
		prompt.WriteString(fmt.Sprintf("[%s]: %s\n", msg.Role, msg.Content))
	}

	return p.GenerateStream(ctx, prompt.String(), options.GenerateOptions)
}

// Embeddings implements ModelProvider.Embeddings. This provider does not
// extract embeddings from the model, so it reports that instead of
// returning pseudo-vectors.
func (p *AppStorageProvider) Embeddings(ctx context.Context, text string) ([]float64, error) {
	return nil, fmt.Errorf("embeddings are not supported by the App Storage provider")
}

// GetInfo implements ModelProvider.GetInfo
func (p *AppStorageProvider) GetInfo() deeptreeecho.ProviderInfo {
	models, _ := p.ListStorageModels()

	return deeptreeecho.ProviderInfo{
		Name:        "App Storage",
		Description: fmt.Sprintf("Large models from Replit App Storage (bucket: %s)", p.bucketID),
		Models:      models,
		Capabilities: []string{
			"generation",
			"streaming",
			"chat",
			"embeddings",
			"cloud-storage",
			"large-models",
		},
	}
}

// IsAvailable implements ModelProvider.IsAvailable
func (p *AppStorageProvider) IsAvailable() bool {
	return p.bucketID != ""
}

// GetLoadedModel returns the currently loaded model
func (p *AppStorageProvider) GetLoadedModel() string {
	p.sel.Lock()
	defer p.sel.Unlock()
	return p.loadedModel
}

// GetModelInfo returns information about the loaded model
func (p *AppStorageProvider) GetModelInfo() map[string]interface{} {
	p.sel.Lock()
	defer p.sel.Unlock()
	info := make(map[string]interface{}, len(p.modelInfo))
	for k, v := range p.modelInfo {
		info[k] = v
	}
	return info
}

// GetCachedModels returns list of cached models
func (p *AppStorageProvider) GetCachedModels() map[string]string {
	return p.cachedModels
}

// ClearCache removes cached models to free space
func (p *AppStorageProvider) ClearCache() error {
	for modelName, cachePath := range p.cachedModels {
		if err := os.Remove(cachePath); err != nil {
			// Continue even if one fails
			continue
		}
		delete(p.cachedModels, modelName)
	}
	return nil
}

// UploadModel uploads a model to App Storage (for admin use)
func (p *AppStorageProvider) UploadModel(localPath, modelName string) error {
	return fmt.Errorf("upload to App Storage is not implemented")
}

// GetStorageURL returns the GCS URL for direct access
func (p *AppStorageProvider) GetStorageURL(modelName string) string {
	return fmt.Sprintf("gs://%s/%s", p.bucketID, modelName)
}

// DownloadFromURL downloads a model from a public URL to App Storage
func (p *AppStorageProvider) DownloadFromURL(url, modelName string) error {
	// This would download from URL and upload to GCS
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download from URL: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download from URL: %s", resp.Status)
	}
	if err := os.MkdirAll(p.localCache, 0o755); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	// Create local file
	localPath := filepath.Join(p.localCache, modelName)
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Copy content
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save model: %v", err)
	}

	// Update cache
	p.cachedModels[modelName] = localPath

	return nil
}
