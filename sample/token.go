package sample

// token represents information about a single token during sampling.
// Defined here (untagged) so that both cgo and non-cgo builds compile,
// since transforms.go and the samplers reference it in all build modes.
type token struct {
	id    int32   // The token's unique identifier
	value float32 // The raw logit or probability from the model
}
