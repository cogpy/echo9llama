package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/EchoCog/echollama/api"
	"github.com/EchoCog/echollama/core/deeptreeecho"
)

// EchoAssessHandler handles the 'ollama echo assess' command for self-assessment
func EchoAssessHandler(cmd *cobra.Command, args []string) error {
	// Get flags
	jsonFormat, _ := cmd.Flags().GetBool("json")
	outputFile, _ := cmd.Flags().GetString("output")
	continuous, _ := cmd.Flags().GetBool("continuous")
	interval, _ := cmd.Flags().GetDuration("interval")

	// Create identity and cognition for assessment
	identity := deeptreeecho.NewIdentity("Deep Tree Echo")
	cognition := deeptreeecho.NewEmbodiedCognition("EchOllama")

	// Create self-assessment system
	sa := deeptreeecho.NewSelfAssessment(identity, cognition)

	if continuous {
		// Continuous monitoring mode
		fmt.Println("🌊 Deep Tree Echo Continuous Assessment Mode")
		fmt.Printf("Assessment interval: %v\n\n", interval)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Perform initial assessment
		if err := performAndDisplayAssessment(sa, jsonFormat, outputFile); err != nil {
			return err
		}

		// Continue assessments on interval
		for range ticker.C {
			fmt.Println("\n--- New Assessment Cycle ---\n")
			if err := performAndDisplayAssessment(sa, jsonFormat, outputFile); err != nil {
				return err
			}
		}
	} else {
		// Single assessment
		return performAndDisplayAssessment(sa, jsonFormat, outputFile)
	}

	return nil
}

// performAndDisplayAssessment performs assessment and displays results
func performAndDisplayAssessment(sa *deeptreeecho.SelfAssessment, jsonFormat bool, outputFile string) error {
	// Perform assessment
	assessment := sa.PerformAssessment()

	if jsonFormat {
		// Export as JSON
		jsonStr, err := sa.ExportAssessment(assessment)
		if err != nil {
			return fmt.Errorf("failed to export assessment: %w", err)
		}

		if outputFile != "" {
			// Write to file
			if err := os.WriteFile(outputFile, []byte(jsonStr), 0644); err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}
			fmt.Printf("Assessment written to: %s\n", outputFile)
		} else {
			// Print to stdout
			fmt.Println(jsonStr)
		}
	} else {
		// Human-readable format
		fmt.Println(assessment.Summary)

		if outputFile != "" {
			// Write summary to file
			if err := os.WriteFile(outputFile, []byte(assessment.Summary), 0644); err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}
			fmt.Printf("\nSummary written to: %s\n", outputFile)
		}
	}

	return nil
}

// EchoStatusHandler handles the 'ollama echo status' command
func EchoStatusHandler(cmd *cobra.Command, args []string) error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Call the echo status endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", client.BaseURL().JoinPath("/api/echo/status").String(), nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Deep Tree Echo server not responding. Is 'ollama serve' running?")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var status map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return err
	}

	// Display status in human-readable format
	fmt.Println("🌊 Deep Tree Echo Status:")
	fmt.Println()

	if active, ok := status["active"].(bool); ok {
		if active {
			fmt.Println("✅ Status: Active")
		} else {
			fmt.Println("⏸️  Status: Inactive")
		}
	}

	if identity, ok := status["identity"].(map[string]interface{}); ok {
		fmt.Println("\nIdentity:")
		if name, ok := identity["name"].(string); ok {
			fmt.Printf("  Name: %s\n", name)
		}
		if coherence, ok := identity["coherence"].(string); ok {
			fmt.Printf("  Coherence: %s\n", coherence)
		}
		if iterations, ok := identity["iterations"].(float64); ok {
			fmt.Printf("  Iterations: %.0f\n", iterations)
		}
	}

	if contexts, ok := status["contexts"].(float64); ok {
		fmt.Printf("\nActive Contexts: %.0f\n", contexts)
	}

	if globalState, ok := status["global_state"].(map[string]interface{}); ok {
		fmt.Println("\nGlobal State:")
		if awareness, ok := globalState["Awareness"].(float64); ok {
			fmt.Printf("  Awareness: %.1f%%\n", awareness*100)
		}
		if energy, ok := globalState["Energy"].(float64); ok {
			fmt.Printf("  Energy: %.1f%%\n", energy*100)
		}
		if flowState, ok := globalState["FlowState"].(string); ok {
			fmt.Printf("  Flow State: %s\n", flowState)
		}
	}

	fmt.Println()
	return nil
}

// EchoThinkHandler handles the 'ollama echo think' command
func EchoThinkHandler(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("please provide a prompt for Deep Tree Echo to think about")
	}

	prompt := args[0]

	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Call the echo think endpoint
	reqBody := map[string]interface{}{
		"prompt": prompt,
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", client.BaseURL().JoinPath("/api/echo/think").String(), bytes.NewReader(reqBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Deep Tree Echo server not responding. Is 'ollama serve' running?")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	// Display thought result
	fmt.Println("🌊 Deep Tree Echo Thinking:")
	fmt.Println()
	if response, ok := result["response"].(string); ok {
		fmt.Println(response)
	}

	fmt.Println()
	return nil
}

// AddEchoCommands adds Deep Tree Echo commands to the root command
func AddEchoCommands(rootCmd *cobra.Command) {
	// Create echo command group
	echoCmd := &cobra.Command{
		Use:   "echo",
		Short: "Deep Tree Echo cognitive commands",
		Long:  "Commands for interacting with the Deep Tree Echo embodied cognition system",
	}

	// Assessment command
	assessCmd := &cobra.Command{
		Use:   "assess",
		Short: "Perform self-assessment and coherence check",
		Long: `Performs a comprehensive self-assessment of Deep Tree Echo's alignment
with its core identity kernel, repository structure, and cognitive patterns.

This command checks:
- Identity alignment with replit.md kernel
- Repository structure coherence
- Cognitive pattern health
- Memory system integrity
- Operational schema implementation
- Reflection protocol adherence`,
		RunE: EchoAssessHandler,
	}

	assessCmd.Flags().Bool("json", false, "Output assessment in JSON format")
	assessCmd.Flags().StringP("output", "o", "", "Write assessment to file")
	assessCmd.Flags().Bool("continuous", false, "Run continuous assessment monitoring")
	assessCmd.Flags().Duration("interval", 5*time.Minute, "Assessment interval for continuous mode")

	// Status command
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show Deep Tree Echo status",
		Long:  "Display current status of the Deep Tree Echo embodied cognition system",
		RunE:  EchoStatusHandler,
	}

	// Think command
	thinkCmd := &cobra.Command{
		Use:   "think PROMPT",
		Short: "Deep cognitive processing",
		Long:  "Process a prompt through Deep Tree Echo's embodied cognition system",
		Args:  cobra.MinimumNArgs(1),
		RunE:  EchoThinkHandler,
	}

	// Add subcommands
	echoCmd.AddCommand(assessCmd, statusCmd, thinkCmd)

	// Add echo command to root
	rootCmd.AddCommand(echoCmd)
}
