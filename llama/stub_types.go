//go:build !cgo
// +build !cgo

package llama

import (
	"errors"
)

// Stub types when CGO is not available
// These allow the code to compile but will return errors if actually used

// Model represents a stub llama model
type Model struct{}

// Grammar represents a stub grammar
type Grammar struct{}

// TokenData represents token information
type TokenData struct {
	ID    int32
	Logit float32
}

// ModelParams contains model loading parameters
type ModelParams struct {
	NumGpuLayers int
	MainGpu      int
	UseMmap      bool
	TensorSplit  []float32
	Progress     func(float32)
}

// NewGrammar creates a stub grammar (returns error)
func NewGrammar(grammar string, vocabIds []uint32, vocabValues []string, eogTokens []int32) *Grammar {
	return nil
}

// LoadModelFromFile is a stub that returns an error
func LoadModelFromFile(path string, params ModelParams) (*Model, error) {
	return nil, errors.New("llama.cpp bindings not available (CGO disabled)")
}

// FreeModel is a stub
func FreeModel(model *Model) {
	// No-op
}

// BackendInit is a stub
func BackendInit() {
	// No-op
}

// GetModelArch is a stub
func GetModelArch(modelPath string) (string, error) {
	return "", errors.New("llama.cpp bindings not available (CGO disabled)")
}
