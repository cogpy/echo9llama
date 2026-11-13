package llm

import "os"

// Environment variables shared with the deeptreeecho providers so one
// configuration serves every entry point.
const (
	envLocalBaseURL = "LOCAL_LLM_BASE_URL"
	envLocalModel   = "LOCAL_LLM_MODEL"
	envLocalAPIKey  = "LOCAL_LLM_API_KEY"

	defaultLocalBaseURL = "http://localhost:11434/v1"
	defaultLocalModel   = "llama3.2"
)

// NewLocalProvider returns the provider for a locally served model: any
// OpenAI-compatible server such as Ollama (default, http://localhost:11434/v1)
// or llama-server. It is always registered; whether it is available is
// decided per request by asking the server, never assumed.
func NewLocalProvider() *OpenAIProvider {
	base := os.Getenv(envLocalBaseURL)
	if base == "" {
		base = defaultLocalBaseURL
	}
	model := os.Getenv(envLocalModel)
	if model == "" {
		model = defaultLocalModel
	}
	return NewOpenAICompatibleProvider("local", base, os.Getenv(envLocalAPIKey), model)
}
