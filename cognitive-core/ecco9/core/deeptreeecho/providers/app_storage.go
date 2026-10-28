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

	"github.com/EchoCog/echollama/core/deeptreeecho"
)

// AppStorageProvider implements ModelProvider for Replit App Storage
type AppStorageProvider struct {
	bucketID     string
	localCache   string
	cachedModels map[string]string // model name -> local path
	loadedModel  string
	modelInfo    map[string]interface{}
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

// ListStorageModels lists models available in App Storage.
// Listing the bucket is not implemented; it reports an error instead of a
// fabricated model list.
func (p *AppStorageProvider) ListStorageModels() ([]string, error) {
	return nil, fmt.Errorf("listing App Storage bucket %s is not implemented", p.bucketID)
}

// DownloadModel downloads a model from App Storage to local cache
func (p *AppStorageProvider) DownloadModel(modelName string) (string, error) {
	// Check if already cached
	if cachedPath, exists := p.cachedModels[modelName]; exists {
		if _, err := os.Stat(cachedPath); err == nil {
			return cachedPath, nil
		}
	}

	// Create cache directory
	if err := os.MkdirAll(p.localCache, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %v", err)
	}

	// Build local path
	localPath := filepath.Join(p.localCache, modelName)

	_ = localPath
	return "", fmt.Errorf("downloading %s from App Storage is not implemented", modelName)
}

// LoadModel loads a model from App Storage
func (p *AppStorageProvider) LoadModel(modelName string) error {
	// Download if needed
	localPath, err := p.DownloadModel(modelName)
	if err != nil {
		return err
	}

	p.loadedModel = modelName
	p.modelInfo["name"] = modelName
	p.modelInfo["path"] = localPath
	p.modelInfo["bucket"] = p.bucketID
	p.modelInfo["cached"] = true

	// Get file size
	if info, err := os.Stat(localPath); err == nil {
		p.modelInfo["size_mb"] = info.Size() / (1024 * 1024)
	}

	p.modelInfo["status"] = "loaded from App Storage"
	return nil
}

// Generate implements ModelProvider.Generate
func (p *AppStorageProvider) Generate(ctx context.Context, prompt string, options deeptreeecho.GenerateOptions) (string, error) {
	if p.loadedModel == "" {
		// Try to load a default model
		models, err := p.ListStorageModels()
		if err != nil || len(models) == 0 {
			return "", fmt.Errorf("no models available in App Storage")
		}

		if err := p.LoadModel(models[0]); err != nil {
			return "", fmt.Errorf("failed to load model: %v", err)
		}
	}

	// No inference backend is linked into this provider; it never fabricates
	// text in place of a model.
	return "", fmt.Errorf("App Storage inference is not available in this build (model %s)", p.loadedModel)
}

// GenerateStream implements ModelProvider.GenerateStream
func (p *AppStorageProvider) GenerateStream(ctx context.Context, prompt string, options deeptreeecho.GenerateOptions) (<-chan string, error) {
	ch := make(chan string, 100)

	go func() {
		defer close(ch)

		response, err := p.Generate(ctx, prompt, options)
		if err != nil {
			return
		}

		words := strings.Fields(response)
		for _, word := range words {
			select {
			case <-ctx.Done():
				return
			case ch <- word + " ":
			}
		}
	}()

	return ch, nil
}

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

// Embeddings implements ModelProvider.Embeddings
func (p *AppStorageProvider) Embeddings(ctx context.Context, text string) ([]float64, error) {
	return nil, fmt.Errorf("App Storage embeddings are not available in this build")
}

// GetInfo implements ModelProvider.GetInfo
func (p *AppStorageProvider) GetInfo() deeptreeecho.ProviderInfo {
	models, _ := p.ListStorageModels()

	return deeptreeecho.ProviderInfo{
		Name:         "App Storage",
		Description:  fmt.Sprintf("Large models from Replit App Storage (bucket: %s)", p.bucketID),
		Models:       models,
		Capabilities: []string{"cloud-storage"},
	}
}

// IsAvailable implements ModelProvider.IsAvailable
// Inference is not implemented, so the provider never reports itself as able
// to answer.
func (p *AppStorageProvider) IsAvailable() bool {
	return false
}

// GetLoadedModel returns the currently loaded model
func (p *AppStorageProvider) GetLoadedModel() string {
	return p.loadedModel
}

// GetModelInfo returns information about the loaded model
func (p *AppStorageProvider) GetModelInfo() map[string]interface{} {
	return p.modelInfo
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
