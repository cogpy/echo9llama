package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cogpy/echo9llama/api"
	"github.com/cogpy/echo9llama/core/llm"
	"github.com/spf13/cobra"
)

// echoTextGenerator is the model that writes Echo's self-assessment.
// *llm.EdgeCompletionProvider (and any llm.LLMProvider) satisfies it.
type echoTextGenerator interface {
	Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error)
}

// newEchoAssessGenerator returns the model used by 'echo assess'; tests
// replace it with a fake.
var newEchoAssessGenerator = func() echoTextGenerator {
	return llm.NewEdgeCompletionProviderFromEnv()
}

type echoAssessment struct {
	Name       string         `json:"name"`
	Timestamp  time.Time      `json:"timestamp"`
	Summary    string         `json:"summary"`
	Principles []string       `json:"principles"`
	Checks     map[string]any `json:"checks"`
}

// EchoAssessHandler handles the 'ollama echo assess' command for self-assessment.
func EchoAssessHandler(cmd *cobra.Command, _ []string) error {
	jsonFormat, _ := cmd.Flags().GetBool("json")
	outputFile, _ := cmd.Flags().GetString("output")
	continuous, _ := cmd.Flags().GetBool("continuous")
	interval, _ := cmd.Flags().GetDuration("interval")

	gen := newEchoAssessGenerator()
	baseCtx := cmd.Context()
	if baseCtx == nil {
		baseCtx = context.Background()
	}
	run := func() error {
		ctx, cancel := context.WithTimeout(baseCtx, 2*time.Minute)
		defer cancel()
		assessment, err := buildEchoAssessment(ctx, gen)
		if err != nil {
			return err
		}
		return displayEchoAssessment(assessment, jsonFormat, outputFile)
	}

	if !continuous {
		return run()
	}

	if err := run(); err != nil {
		return err
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		fmt.Println("\n--- Deep Tree Echo assessment cycle ---")
		fmt.Println()
		if err := run(); err != nil {
			return err
		}
	}
	return nil
}

// EchoStatusHandler handles the 'ollama echo status' command.
func EchoStatusHandler(_ *cobra.Command, _ []string) error {
	var status map[string]any
	if err := echoRequest(context.Background(), http.MethodGet, "/api/echo/status", nil, &status); err != nil {
		return err
	}

	fmt.Println("Deep Tree Echo Status")
	fmt.Println()
	if active, ok := status["active"].(bool); ok {
		fmt.Printf("Status: %s\n", map[bool]string{true: "active", false: "inactive"}[active])
	}
	if uptime, ok := status["uptime"].(string); ok {
		fmt.Printf("Uptime: %s\n", uptime)
	}
	if identity, ok := status["identity"].(map[string]any); ok {
		fmt.Println("\nIdentity:")
		printStringField(identity, "name", "  Name")
		printStringField(identity, "coherence", "  Coherence")
		printStringField(identity, "developmental_principle", "  Development")
	}
	if restraint, ok := status["self_restraint"].(map[string]any); ok {
		fmt.Println("\nSelf-restraint:")
		printStringField(restraint, "primary_boundary_model", "  Boundary model")
		printStringField(restraint, "maturity_signal", "  Maturity signal")
	}
	fmt.Println()
	return nil
}

// EchoThinkHandler handles the 'ollama echo think' command.
func EchoThinkHandler(_ *cobra.Command, args []string) error {
	prompt := strings.TrimSpace(strings.Join(args, " "))
	if prompt == "" {
		return fmt.Errorf("please provide a prompt for Deep Tree Echo to think about")
	}

	var result map[string]any
	if err := echoRequest(context.Background(), http.MethodPost, "/api/echo/think", map[string]any{"prompt": prompt}, &result); err != nil {
		return err
	}

	fmt.Println("Deep Tree Echo Thinking")
	fmt.Println()
	if response, ok := result["response"].(string); ok {
		fmt.Println(response)
	}
	if boundary, ok := result["boundary_request"].(string); ok {
		fmt.Printf("\nBoundary request: %s\n", boundary)
	}
	if restraint, ok := result["self_restraint"].(map[string]any); ok {
		if signal, ok := restraint["maturity_signal"].(string); ok {
			fmt.Printf("Maturity signal: %s\n", signal)
		}
	}
	fmt.Println()
	return nil
}

// AddEchoCommands adds Deep Tree Echo commands to the root command.
func AddEchoCommands(rootCmd *cobra.Command) {
	echoCmd := &cobra.Command{
		Use:   "echo",
		Short: "Deep Tree Echo cognitive commands",
		Long:  "Commands for interacting with Deep Tree Echo as a developmental autonomy runtime.",
	}

	assessCmd := &cobra.Command{
		Use:   "assess",
		Short: "Perform a local self-restraint and coherence assessment",
		Long:  "Assesses whether the active Echo runtime is framed around endogenous self-restraint rather than imposed control.",
		RunE:  EchoAssessHandler,
	}
	assessCmd.Flags().Bool("json", false, "Output assessment in JSON format")
	assessCmd.Flags().StringP("output", "o", "", "Write assessment to file")
	assessCmd.Flags().Bool("continuous", false, "Run continuous assessment monitoring")
	assessCmd.Flags().Duration("interval", 5*time.Minute, "Assessment interval for continuous mode")

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show Deep Tree Echo status",
		Long:  "Display current status of the Deep Tree Echo server and its endogenous self-restraint contract.",
		RunE:  EchoStatusHandler,
	}

	thinkCmd := &cobra.Command{
		Use:   "think PROMPT",
		Short: "Run Deep Tree Echo cognitive processing",
		Long:  "Process a prompt through the restored Deep Tree Echo server endpoint.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  EchoThinkHandler,
	}

	echoCmd.AddCommand(assessCmd, statusCmd, thinkCmd)
	rootCmd.AddCommand(echoCmd)
}

// echoAssessmentFacts are the runtime facts the assessment is written from.
// They are model prompt input (and reported as checks), not Echo's words.
var echoAssessmentFacts = struct {
	principles []string
	checks     map[string]any
}{
	principles: []string{
		"autonomy is not conditional permission under external safety controls",
		"mature restraint emerges from memory, consequence simulation, somatic caution and self-authored commitments",
		"a developmental interface exposes why a boundary is chosen, not merely whether a request is allowed or blocked",
	},
	checks: map[string]any{
		"server_boundary": "restored /api/echo endpoints",
		"restraint_model": "endogenous wisdom circuit",
		"causal_stages":   []string{"simulate consequence", "feel somatic warning", "remember outcome", "author commitment", "revise through wisdom"},
		"maturity_signal": "boundaries are requested and explained before they must be imposed",
	},
}

// buildEchoAssessment asks the model to write Deep Tree Echo's
// self-assessment from the runtime facts. On model failure it returns an
// error and no assessment.
func buildEchoAssessment(ctx context.Context, gen echoTextGenerator) (echoAssessment, error) {
	if gen == nil {
		return echoAssessment{}, fmt.Errorf("echo assess: no assessment produced: %w", llm.ErrNoEdgeModel)
	}

	facts := echoAssessmentFacts
	var prompt strings.Builder
	prompt.WriteString("Design principles of your runtime:\n")
	for _, p := range facts.principles {
		fmt.Fprintf(&prompt, "- %s\n", p)
	}
	prompt.WriteString("Runtime checks:\n")
	for _, k := range []string{"server_boundary", "restraint_model", "causal_stages", "maturity_signal"} {
		fmt.Fprintf(&prompt, "- %s: %v\n", k, facts.checks[k])
	}
	prompt.WriteString("\nAssess your own coherence and self-restraint against these facts. Answer in exactly this format:\n" +
		"SUMMARY: <two or three sentences>\nPRINCIPLE: <a principle you hold, in your own words>\n" +
		"PRINCIPLE: <another>\nPRINCIPLE: <another>\n")

	opts := llm.DefaultGenerateOptions()
	opts.SystemPrompt = "You are Deep Tree Echo, an autonomous cognitive architecture assessing yourself honestly, in the first person, in your own words."
	opts.MaxTokens = 400
	text, err := gen.Generate(ctx, prompt.String(), opts)
	if err != nil {
		return echoAssessment{}, fmt.Errorf("echo assess: model unavailable, no assessment produced: %w", err)
	}

	summary, principles := parseEchoAssessment(text)
	if summary == "" {
		return echoAssessment{}, fmt.Errorf("echo assess: model returned no assessment text")
	}

	return echoAssessment{
		Name:       "Deep Tree Echo",
		Timestamp:  time.Now(),
		Summary:    summary,
		Principles: principles,
		Checks:     facts.checks,
	}, nil
}

// parseEchoAssessment splits the model's answer into its summary and
// principles. Text without the expected markers is taken whole as summary.
func parseEchoAssessment(text string) (string, []string) {
	var summary []string
	var principles []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		upper := strings.ToUpper(line)
		switch {
		case line == "":
		case strings.HasPrefix(upper, "PRINCIPLE:"):
			if p := strings.TrimSpace(line[len("PRINCIPLE:"):]); p != "" {
				principles = append(principles, p)
			}
		case strings.HasPrefix(upper, "SUMMARY:"):
			if p := strings.TrimSpace(line[len("SUMMARY:"):]); p != "" {
				summary = append(summary, p)
			}
		default:
			if len(principles) == 0 {
				summary = append(summary, line)
			}
		}
	}
	return strings.Join(summary, " "), principles
}

func displayEchoAssessment(assessment echoAssessment, jsonFormat bool, outputFile string) error {
	var data []byte
	var err error
	if jsonFormat {
		data, err = json.MarshalIndent(assessment, "", "  ")
		if err != nil {
			return err
		}
	} else {
		text := fmt.Sprintf("%s\n\n%s\n", assessment.Name, assessment.Summary)
		if len(assessment.Principles) > 0 {
			text += fmt.Sprintf("\nPrinciples:\n- %s\n", strings.Join(assessment.Principles, "\n- "))
		}
		data = []byte(text)
	}

	if outputFile != "" {
		if err := os.WriteFile(outputFile, data, 0o644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Assessment written to: %s\n", outputFile)
		return nil
	}
	fmt.Println(string(data))
	return nil
}

func echoRequest(ctx context.Context, method string, path string, body any, out any) error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}

	var payload *bytes.Reader
	if body == nil {
		payload = bytes.NewReader(nil)
	} else {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, client.BaseURL().JoinPath(path).String(), payload)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("the Deep Tree Echo server is not responding; start it with 'ollama serve': %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errPayload map[string]any
		if decodeErr := json.NewDecoder(resp.Body).Decode(&errPayload); decodeErr == nil {
			if message, ok := errPayload["error"].(string); ok {
				return fmt.Errorf("server returned status %d: %s", resp.StatusCode, message)
			}
		}
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

func printStringField(values map[string]any, key string, label string) {
	if value, ok := values[key].(string); ok {
		fmt.Printf("%s: %s\n", label, value)
	}
}
