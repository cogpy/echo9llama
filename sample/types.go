package sample

// token represents a token with its ID and logit value
type token struct {
	id    int32   // The token's unique identifier
	value float32 // The raw logit or probability from the model
}

// Sampler represents a sampling configuration
type Sampler struct {
	// Configuration fields (stub for non-CGO builds)
}

// NewSampler creates a new sampler (stub)
func NewSampler(topK int, topP, minP, temperature float32) *Sampler {
	return &Sampler{}
}
