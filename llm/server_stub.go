//go:build !cgo
// +build !cgo

package llm

import (
	"context"
	"errors"

	"github.com/cogpy/echo9llama/api"
	"github.com/cogpy/echo9llama/discover"
	"github.com/cogpy/echo9llama/fs/ggml"
)

// Stub LlamaServer interface when CGO is not available
// This allows the code to compile but local GGUF models won't work

type LlamaServer interface {
	Ping(ctx context.Context) error
	WaitUntilRunning(ctx context.Context) error
	Completion(ctx context.Context, req CompletionRequest, fn func(CompletionResponse)) error
	Embedding(ctx context.Context, input string) ([]float32, error)
	Tokenize(ctx context.Context, content string) ([]int, error)
	Detokenize(ctx context.Context, tokens []int) (string, error)
	Close() error
	EstimatedVRAM() uint64
	EstimatedTotal() uint64
	EstimatedVRAMByGPU(gpuID string) uint64
	Pid() int
}

// NewLlamaServer stub returns an error when CGO is disabled
func NewLlamaServer(gpus discover.GpuInfoList, modelPath string, f *ggml.GGML, adapters, projectors []string, opts api.Options, numParallel int) (LlamaServer, error) {
	return nil, errors.New("local GGUF model support not available (CGO disabled) - use API-based providers instead")
}

// CompletionRequest represents a completion request
type CompletionRequest struct {
	Prompt  string
	Format  string
	Images  []ImageData
	Options api.Options
}

// CompletionResponse represents a completion response
type CompletionResponse struct {
	Content            string
	Model              string
	CreatedAt          string
	Done               bool
	DoneReason         string
	TotalDuration      int64
	LoadDuration       int64
	PromptEvalCount    int
	PromptEvalDuration int64
	EvalCount          int
	EvalDuration       int64
}

// ImageData represents image data for multimodal models
type ImageData struct {
	Data []byte
	ID   int
}

// MemoryEstimate represents memory usage estimates
type MemoryEstimate struct {
	VRAMSize       uint64
	TotalSize      uint64
	Layers         int
	Graph          uint64
	Weights        uint64
	ContextSize    uint64
	ComputeSize    uint64
	TensorSplit    []uint64
}

// EstimateGPULayers stub function
func EstimateGPULayers(gpus discover.GpuInfoList, f *ggml.GGML, projectors []string, opts api.Options, numParallel int) MemoryEstimate {
	return MemoryEstimate{}
}
