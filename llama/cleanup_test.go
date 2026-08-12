//go:build cgo && !nollama

package llama

import "testing"

func TestNativeCleanupIsNilSafeAndIdempotent(t *testing.T) {
	var nilContext *Context
	nilContext.Free()
	context := &Context{}
	context.Free()
	context.Free()

	var nilSampler *SamplingContext
	nilSampler.Free()
	sampler := &SamplingContext{}
	sampler.Free()
	sampler.Free()
}
