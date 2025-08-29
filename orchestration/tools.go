package orchestration

import (
	"context"
	"fmt"
	"time"
)

// ExampleWebSearchTool demonstrates a web search tool integration
type ExampleWebSearchTool struct{}

func (t *ExampleWebSearchTool) Name() string {
	return "web_search"
}

func (t *ExampleWebSearchTool) Description() string {
	return "Performs web search queries (example implementation)"
}

func (t *ExampleWebSearchTool) Call(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	query, ok := params["query"].(string)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "query parameter required",
		}, nil
	}

	// Simulate web search
	time.Sleep(100 * time.Millisecond)
	
	return &ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Search results for '%s': Found 5 relevant articles on AI and orchestration", query),
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

// ExampleDataAnalysisPlugin demonstrates a data analysis plugin
type ExampleDataAnalysisPlugin struct{}

func (p *ExampleDataAnalysisPlugin) Name() string {
	return "data_analysis"
}

func (p *ExampleDataAnalysisPlugin) Description() string {
	return "Performs basic data analysis and pattern recognition"
}

func (p *ExampleDataAnalysisPlugin) Execute(ctx context.Context, input string, params map[string]interface{}) (interface{}, error) {
	// Simulate data analysis
	time.Sleep(200 * time.Millisecond)
	
	analysisType, ok := params["type"].(string)
	if !ok {
		analysisType = "summary"
	}

	switch analysisType {
	case "summary":
		return map[string]interface{}{
			"type":    "summary",
			"input":   input,
			"length":  len(input),
			"words":   len(input)/5, // rough word estimate
			"insight": "Text appears to contain structured information suitable for further analysis",
		}, nil
	case "sentiment":
		return map[string]interface{}{
			"type":      "sentiment",
			"sentiment": "neutral",
			"confidence": 0.7,
			"factors":   []string{"balanced_tone", "technical_content"},
		}, nil
	default:
		return fmt.Sprintf("Unknown analysis type: %s", analysisType), nil
	}
}

// RegisterDefaultTools registers the default set of tools with an engine
func RegisterDefaultTools(engine *Engine) {
	engine.RegisterTool(&ExampleWebSearchTool{})
	engine.RegisterTool(&ExampleCalculatorTool{})
}

// RegisterDefaultPlugins registers the default set of plugins with an engine
func RegisterDefaultPlugins(engine *Engine) {
	engine.RegisterPlugin(&ExampleDataAnalysisPlugin{})
}