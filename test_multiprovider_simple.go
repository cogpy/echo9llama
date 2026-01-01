//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"

	"github.com/cogpy/echo9llama/core/deeptreeecho"
	"github.com/cogpy/echo9llama/core/llm"
)

func main() {
	fmt.Println("Testing MultiProviderLLM interface compatibility...")
	
	// Create MultiProviderLLM
	multiProvider := deeptreeecho.NewMultiProviderLLM()
	
	// Test that it implements llm.LLMProvider interface
	var provider llm.LLMProvider = multiProvider
	
	fmt.Printf("Provider name: %s\n", provider.Name())
	fmt.Printf("Available: %v\n", provider.Available())
	fmt.Printf("Max tokens: %d\n", provider.MaxTokens())
	
	if provider.Available() {
		ctx := context.Background()
		opts := llm.DefaultGenerateOptions()
		result, err := provider.Generate(ctx, "Test prompt", opts)
		if err != nil {
			fmt.Printf("Generate error: %v\n", err)
		} else {
			fmt.Printf("Generate result: %s\n", result)
		}
	}
	
	fmt.Println("✅ MultiProviderLLM implements llm.LLMProvider interface correctly")
}
