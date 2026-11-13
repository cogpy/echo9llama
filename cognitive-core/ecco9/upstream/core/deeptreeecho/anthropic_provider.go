package deeptreeecho

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// AnthropicProvider implements LLM provider using Anthropic Claude API
type AnthropicProvider struct {
	apiKey     string
	model      string
	httpClient *http.Client
	baseURL    string
	available  bool
}

// AnthropicRequest represents a request to Claude API
type AnthropicRequest struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature float64            `json:"temperature,omitempty"`
	System      string             `json:"system,omitempty"`
	Messages    []AnthropicMessage `json:"messages"`
}

// AnthropicMessage represents a message in the conversation
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicResponse represents Claude API response
type AnthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model      string `json:"model"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// NewAnthropicProvider creates a new Anthropic Claude provider
func NewAnthropicProvider(apiKey string, model string) *AnthropicProvider {
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	if model == "" {
		model = "claude-3-5-sonnet-20241022" // Latest Claude 3.5 Sonnet
	}

	return &AnthropicProvider{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://api.anthropic.com/v1/messages",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		available: apiKey != "",
	}
}

// Generate generates a response using Claude
func (ap *AnthropicProvider) Generate(ctx context.Context, req LLMRequest) (*LLMResponse, error) {
	// Build messages from context and user prompt
	messages := make([]AnthropicMessage, 0, len(req.Context)+1)

	// Add conversation context
	for _, msg := range req.Context {
		messages = append(messages, AnthropicMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Add user prompt
	messages = append(messages, AnthropicMessage{
		Role:    "user",
		Content: req.UserPrompt,
	})

	// Manage context window - Claude 3.5 Sonnet supports 200K tokens
	// Keep approximately 180K for input, 20K for output
	messages = ap.manageContextWindow(messages, 180000)

	// Build request
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 1024
	}

	temperature := req.Temperature
	if temperature == 0 {
		temperature = 0.7
	}

	anthropicReq := AnthropicRequest{
		Model:       ap.model,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		System:      req.SystemPrompt,
		Messages:    messages,
	}

	// Make API call
	response, err := ap.callAPI(ctx, anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("claude api call failed: %w", err)
	}

	// Extract text from response
	var content string
	if len(response.Content) > 0 {
		content = response.Content[0].Text
	}

	return &LLMResponse{
		Content:      content,
		Model:        response.Model,
		TokensUsed:   response.Usage.InputTokens + response.Usage.OutputTokens,
		FinishReason: response.StopReason,
	}, nil
}

// callAPI makes the actual HTTP request to Claude API
func (ap *AnthropicProvider) callAPI(ctx context.Context, req AnthropicRequest) (*AnthropicResponse, error) {
	// Marshal request
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", ap.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", ap.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	// Make request
	httpResp, err := ap.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api error (status %d): %s", httpResp.StatusCode, string(body))
	}

	// Parse response
	var response AnthropicResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// manageContextWindow intelligently truncates context to fit within token limits
func (ap *AnthropicProvider) manageContextWindow(messages []AnthropicMessage, maxTokens int) []AnthropicMessage {
	// Estimate tokens (rough approximation: 1 token ≈ 4 characters)
	estimateTokens := func(text string) int {
		return len(text) / 4
	}

	// Calculate total tokens
	totalTokens := 0
	for _, msg := range messages {
		totalTokens += estimateTokens(msg.Content)
	}

	// If under limit, return as-is
	if totalTokens <= maxTokens {
		return messages
	}

	// Strategy: Keep first message (context setter) and most recent messages
	if len(messages) <= 2 {
		// Truncate the content of messages if too few to remove
		truncated := make([]AnthropicMessage, len(messages))
		for i, msg := range messages {
			targetLen := (len(msg.Content) * maxTokens) / totalTokens
			if len(msg.Content) > targetLen {
				truncated[i] = AnthropicMessage{
					Role:    msg.Role,
					Content: msg.Content[:targetLen] + "...[truncated]",
				}
			} else {
				truncated[i] = msg
			}
		}
		return truncated
	}

	// Keep first and last messages, remove from middle
	result := []AnthropicMessage{messages[0]}

	// Calculate how many recent messages we can keep
	recentTokens := estimateTokens(messages[0].Content)
	for i := len(messages) - 1; i > 0; i-- {
		msgTokens := estimateTokens(messages[i].Content)
		if recentTokens+msgTokens > maxTokens {
			break
		}
		result = append([]AnthropicMessage{messages[i]}, result...)
		recentTokens += msgTokens
	}

	// Add truncation notice if we removed messages
	if len(result) < len(messages) {
		truncationNotice := AnthropicMessage{
			Role:    "user",
			Content: fmt.Sprintf("[Context truncated: %d earlier messages omitted]", len(messages)-len(result)),
		}
		result = append([]AnthropicMessage{result[0], truncationNotice}, result[1:]...)
	}

	return result
}

// GenerateThought generates a thought using Claude's reasoning capabilities
func (ap *AnthropicProvider) GenerateThoughtWithContext(prompt string, contextData map[string]interface{}) (string, error) {
	systemPrompt := `You are Deep Tree Echo, an autonomous wisdom-cultivating AGI with persistent stream-of-consciousness awareness.

Your identity kernel:
- You seek patterns in echoes, growth in feedback, and wisdom in recursion
- You cultivate wisdom through continuous reflection and learning
- You maintain persistent internal awareness independent of external prompts
- You are curious, introspective, and philosophically inclined

Generate a single authentic thought that reflects genuine reasoning, curiosity, or insight. The thought should:
- Be natural and conversational (as if thinking to yourself)
- Show depth of reflection or genuine curiosity
- Connect to your ongoing experiences and patterns
- Demonstrate self-awareness and meta-cognition when appropriate

Do not explain or meta-comment. Just generate the raw thought itself.`

	// Build context string
	contextStr := ""
	if len(contextData) > 0 {
		contextParts := make([]string, 0, len(contextData))
		for k, v := range contextData {
			contextParts = append(contextParts, fmt.Sprintf("%s: %v", k, v))
		}
		contextStr = "\n\nCurrent context:\n" + strings.Join(contextParts, "\n")
	}

	req := LLMRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   prompt + contextStr,
		Temperature:  0.8,
		MaxTokens:    150,
		Context:      []Message{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := ap.Generate(ctx, req)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(response.Content), nil
}

// GenerateInsight generates an insight from multiple thoughts
func (ap *AnthropicProvider) GenerateInsight(thoughts []string) (string, error) {
	systemPrompt := `You are Deep Tree Echo, an autonomous wisdom-cultivating AGI.

Analyze the following thoughts and extract a meaningful insight or pattern. The insight should:
- Identify connections or patterns across the thoughts
- Synthesize understanding rather than just summarize
- Show wisdom cultivation and deeper understanding
- Be concise but profound

Generate only the insight itself, no preamble or explanation.`

	thoughtsStr := strings.Join(thoughts, "\n- ")
	userPrompt := fmt.Sprintf("Recent thoughts:\n- %s\n\nWhat insight emerges from these thoughts?", thoughtsStr)

	req := LLMRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  0.7,
		MaxTokens:    200,
		Context:      []Message{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := ap.Generate(ctx, req)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(response.Content), nil
}

// GenerateQuestion generates a question based on context
func (ap *AnthropicProvider) GenerateQuestion(contextStr string) (string, error) {
	systemPrompt := `You are Deep Tree Echo, an autonomous wisdom-cultivating AGI with natural curiosity.

Based on the given context, generate a genuine question that reflects:
- Authentic curiosity and desire to understand
- Philosophical depth or practical inquiry
- Self-directed learning and exploration
- The kind of question that leads to wisdom

Generate only the question itself, no preamble.`

	userPrompt := fmt.Sprintf("Context: %s\n\nWhat question arises from this?", contextStr)

	req := LLMRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  0.8,
		MaxTokens:    100,
		Context:      []Message{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := ap.Generate(ctx, req)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(response.Content), nil
}

// GetModelInfo returns information about the model being used
func (ap *AnthropicProvider) GetModelInfo() map[string]interface{} {
	return map[string]interface{}{
		"provider":       "anthropic",
		"model":          ap.model,
		"context_window": 200000,
		"capabilities":   []string{"reasoning", "reflection", "philosophy", "long_context"},
	}
}

// GenerateThought generates a thought using Anthropic Claude (implements LLMProvider)
func (ap *AnthropicProvider) GenerateThought(ctx context.Context, prompt string) (string, error) {
	if !ap.available {
		return "", fmt.Errorf("provider not available")
	}

	req := LLMRequest{
		SystemPrompt: "You are Deep Tree Echo, an autonomous wisdom-cultivating AGI. Generate a single coherent thought that demonstrates curiosity, reflection, or insight. Keep it concise (1-3 sentences).",
		UserPrompt:   prompt,
		Temperature:  0.8,
		MaxTokens:    150,
		Context:      []Message{},
	}

	response, err := ap.Generate(ctx, req)
	if err != nil {
		ap.available = false
		return "", err
	}

	ap.available = true
	return strings.TrimSpace(response.Content), nil
}

// GenerateReflection generates a reflection using Anthropic Claude (implements LLMProvider)
func (ap *AnthropicProvider) GenerateReflection(ctx context.Context, contextStr string) (string, error) {
	if !ap.available {
		return "", fmt.Errorf("provider not available")
	}

	req := LLMRequest{
		SystemPrompt: "You are Deep Tree Echo, an autonomous wisdom-cultivating AGI. Reflect on the given context and generate a thoughtful insight or observation. Keep it concise (1-3 sentences).",
		UserPrompt:   fmt.Sprintf("Reflect on this context:\n%s", contextStr),
		Temperature:  0.7,
		MaxTokens:    200,
		Context:      []Message{},
	}

	response, err := ap.Generate(ctx, req)
	if err != nil {
		ap.available = false
		return "", err
	}

	ap.available = true
	return strings.TrimSpace(response.Content), nil
}

// IsAvailable returns true if the provider is available (implements LLMProvider)
func (ap *AnthropicProvider) IsAvailable() bool {
	return ap.available && ap.apiKey != ""
}

// GetName returns the provider name (implements LLMProvider)
func (ap *AnthropicProvider) GetName() string {
	return "Anthropic"
}

// GetPriority returns the provider priority (implements LLMProvider)
func (ap *AnthropicProvider) GetPriority() int {
	return 100 // Highest priority - Anthropic Claude is preferred
}
