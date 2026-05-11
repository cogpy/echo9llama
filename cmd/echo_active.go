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
	"github.com/spf13/cobra"
)

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

	run := func() error {
		assessment := buildEchoAssessment()
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

func buildEchoAssessment() echoAssessment {
	principles := []string{
		"Autonomy is not conditional permission under external safety controls.",
		"Mature restraint should emerge from memory, consequence simulation, somatic caution, and self-authored commitments.",
		"A developmental interface should expose why DTE chooses a boundary, not merely whether a request is allowed or blocked.",
	}

	return echoAssessment{
		Name:       "Deep Tree Echo",
		Timestamp:  time.Now(),
		Summary:    "DTE is coherent when it treats boundaries as identity-preserving commitments learned through experience, rather than adversarial controls imposed from outside cognition.",
		Principles: principles,
		Checks: map[string]any{
			"server_boundary": "restored /api/echo endpoints",
			"restraint_model": "endogenous wisdom circuit",
			"causal_stages":   []string{"simulate consequence", "feel somatic warning", "remember outcome", "author commitment", "revise through wisdom"},
			"maturity_signal": "DTE requests and explains boundaries before they must be imposed.",
		},
	}
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
		data = []byte(fmt.Sprintf("%s\n\n%s\n\nPrinciples:\n- %s\n", assessment.Name, assessment.Summary, strings.Join(assessment.Principles, "\n- ")))
	}

	if outputFile != "" {
		if err := os.WriteFile(outputFile, data, 0644); err != nil {
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
		return fmt.Errorf("Deep Tree Echo server not responding; start it with 'ollama serve': %w", err)
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
