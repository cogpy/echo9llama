package orchestration

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// ExampleWebSearchTool is a web search tool slot. No search backend is
// bundled, so it reports failure instead of returning fabricated results.
// Set Search to plug in a real backend.
type ExampleWebSearchTool struct {
	// Search performs a real web search and returns its results.
	Search func(ctx context.Context, query string) (interface{}, error)
}

func (t *ExampleWebSearchTool) Name() string {
	return "web_search"
}

func (t *ExampleWebSearchTool) Description() string {
	return "Performs web search queries (requires a configured search backend)"
}

// ErrNoSearchBackend is returned when web_search is invoked without a backend.
var ErrNoSearchBackend = errors.New("web_search: no search backend configured")

func (t *ExampleWebSearchTool) Call(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	query, ok := params["query"].(string)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "query parameter required",
		}, nil
	}

	if t.Search == nil {
		return nil, ErrNoSearchBackend
	}

	results, err := t.Search(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("web_search: %w", err)
	}

	return &ToolResult{
		Success: true,
		Output:  results,
	}, nil
}

// ExampleCalculatorTool demonstrates a simple calculation tool
type ExampleCalculatorTool struct{}

func (t *ExampleCalculatorTool) Name() string {
	return "calculator"
}

func (t *ExampleCalculatorTool) Description() string {
	return "Performs basic mathematical calculations"
}

func (t *ExampleCalculatorTool) Call(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	operation, ok := params["operation"].(string)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "operation parameter required",
		}, nil
	}

	a, aOk := params["a"].(float64)
	b, bOk := params["b"].(float64)

	if !aOk || !bOk {
		return &ToolResult{
			Success: false,
			Error:   "numeric parameters 'a' and 'b' required",
		}, nil
	}

	var result float64
	switch operation {
	case "add":
		result = a + b
	case "subtract":
		result = a - b
	case "multiply":
		result = a * b
	case "divide":
		if b == 0 {
			return &ToolResult{
				Success: false,
				Error:   "division by zero",
			}, nil
		}
		result = a / b
	default:
		return &ToolResult{
			Success: false,
			Error:   "unsupported operation",
		}, nil
	}

	return &ToolResult{
		Success: true,
		Output:  result,
	}, nil
}

// ExampleDataAnalysisPlugin performs basic data analysis. "summary" returns
// measured text statistics; "sentiment" is answered by the model through the
// engine (params["model"] selects it) and fails when no model is available.
type ExampleDataAnalysisPlugin struct {
	engine *Engine
}

func (p *ExampleDataAnalysisPlugin) Name() string {
	return "data_analysis"
}

func (p *ExampleDataAnalysisPlugin) Description() string {
	return "Performs basic data analysis and model-based sentiment analysis"
}

func (p *ExampleDataAnalysisPlugin) Execute(ctx context.Context, input string, params map[string]interface{}) (interface{}, error) {
	analysisType, ok := params["type"].(string)
	if !ok {
		analysisType = "summary"
	}

	switch analysisType {
	case "summary":
		return map[string]interface{}{
			"type":   "summary",
			"input":  input,
			"length": len(input),
			"words":  len(strings.Fields(input)),
			"lines":  strings.Count(input, "\n") + 1,
		}, nil
	case "sentiment":
		if p.engine == nil {
			return nil, fmt.Errorf("data_analysis: sentiment analysis requires a model engine")
		}
		model, _ := params["model"].(string)
		if model == "" {
			return nil, fmt.Errorf("data_analysis: sentiment analysis requires params[\"model\"]")
		}
		prompt := "Classify the sentiment of the following text as positive, negative, or neutral, and briefly justify it.\n\nText:\n" + input
		analysis, err := p.engine.generateText(ctx, model, prompt, nil)
		if err != nil {
			return nil, fmt.Errorf("data_analysis: sentiment: %w", err)
		}
		return map[string]interface{}{
			"type":     "sentiment",
			"model":    model,
			"analysis": analysis,
		}, nil
	default:
		return nil, fmt.Errorf("data_analysis: unknown analysis type: %s", analysisType)
	}
}

// RegisterDefaultTools registers the default set of tools with an engine
func RegisterDefaultTools(engine *Engine) {
	engine.RegisterTool(&ExampleWebSearchTool{})
	engine.RegisterTool(&ExampleCalculatorTool{})
}

// RegisterDefaultPlugins registers the default set of plugins with an engine
func RegisterDefaultPlugins(engine *Engine) {
	engine.RegisterPlugin(&ExampleDataAnalysisPlugin{engine: engine})
}
