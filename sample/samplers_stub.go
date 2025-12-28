//go:build !cgo
// +build !cgo

package sample

import (
	"errors"

	"github.com/cogpy/echo9llama/model"
)

// Stub implementations when CGO is not available

type GrammarSampler struct{}

func NewGrammarSampler(model model.TextProcessor, grammarStr string) (*GrammarSampler, error) {
	return nil, errors.New("grammar sampling not available (CGO disabled)")
}

func (g *GrammarSampler) Apply(tokens []token) {
	// No-op
}

func (g *GrammarSampler) Accept(tokenId int32) {
	// No-op
}

func (g *GrammarSampler) Reset() {
	// No-op
}
