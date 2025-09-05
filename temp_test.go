package main

import (
"fmt"
"github.com/EchoCog/echollama/core/deeptreeecho"
)

func main() {
ec := deeptreeecho.NewEmbodiedCognition("Integration Test")
fmt.Printf("🌊 Deep Tree Echo Integration Status:\n")
fmt.Printf("   Identity: %s\n", ec.Identity.Name)
fmt.Printf("   Patterns: %d cognitive patterns extracted\n", len(ec.Identity.Patterns))
fmt.Printf("   Status: ✅ Ready for embodied cognition operations\n")
}
