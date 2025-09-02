package orchestration

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// EchoChat provides shell integration with Deep Tree Echo intelligence
type EchoChat struct {
	engine      *Engine
	currentDir  string
	shell       string
	history     []ChatCommand
	maxHistory  int
}

// ChatCommand represents a command executed through EchoChat
type ChatCommand struct {
	ID          string    `json:"id"`
	Input       string    `json:"input"`
	Command     string    `json:"command,omitempty"`
	Output      string    `json:"output,omitempty"`
	Error       string    `json:"error,omitempty"`
	ExitCode    int       `json:"exit_code"`
	Duration    time.Duration `json:"duration"`
	Timestamp   time.Time `json:"timestamp"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// NewEchoChat creates a new EchoChat instance
func NewEchoChat(engine *Engine) *EchoChat {
	currentDir, _ := os.Getwd()
	shell := getDefaultShell()
	
	return &EchoChat{
		engine:     engine,
		currentDir: currentDir,
		shell:      shell,
		history:    make([]ChatCommand, 0),
		maxHistory: 100,
	}
}

// getDefaultShell returns the default shell for the current OS
func getDefaultShell() string {
	if runtime.GOOS == "windows" {
		return "cmd"
	}
	
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	
	return "/bin/bash"
}

// StartInteractiveSession starts an interactive EchoChat session
func (ec *EchoChat) StartInteractiveSession(ctx context.Context) error {
	fmt.Println("🌊 EchoChat - Deep Tree Echo Shell Assistant")
	fmt.Println("============================================")
	fmt.Println("Type 'help' for commands, 'exit' to quit")
	fmt.Printf("Current directory: %s\n", ec.currentDir)
	fmt.Printf("Shell: %s\n", ec.shell)
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		fmt.Print("echo> ")
		
		if !scanner.Scan() {
			break
		}
		
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		
		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			break
		}
		
		if err := ec.ProcessInput(ctx, input); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
	
	return scanner.Err()
}

// ProcessInput processes user input through Deep Tree Echo intelligence
func (ec *EchoChat) ProcessInput(ctx context.Context, input string) error {
	start := time.Now()
	
	// Create a new chat command
	command := &ChatCommand{
		ID:        fmt.Sprintf("cmd_%d", time.Now().UnixNano()),
		Input:     input,
		Timestamp: start,
		Context:   make(map[string]interface{}),
	}
	
	// Check for built-in commands first
	if ec.handleBuiltinCommand(input, command) {
		ec.addToHistory(command)
		return nil
	}
	
	// Use Deep Tree Echo to interpret the input
	shellCommand, err := ec.interpretWithDeepTreeEcho(ctx, input)
	if err != nil {
		command.Error = err.Error()
		command.ExitCode = 1
		command.Duration = time.Since(start)
		ec.addToHistory(command)
		return err
	}
	
	command.Command = shellCommand
	
	// Execute the command if it looks safe
	if ec.isCommandSafe(shellCommand) {
		err = ec.executeCommand(ctx, command)
	} else {
		// Ask for confirmation for potentially dangerous commands
		if ec.confirmDangerousCommand(shellCommand) {
			err = ec.executeCommand(ctx, command)
		} else {
			command.Error = "Command execution cancelled by user"
			command.ExitCode = 1
		}
	}
	
	command.Duration = time.Since(start)
	ec.addToHistory(command)
	
	return err
}

// interpretWithDeepTreeEcho uses Deep Tree Echo to convert natural language to shell commands
func (ec *EchoChat) interpretWithDeepTreeEcho(ctx context.Context, input string) (string, error) {
	// Create a specialized agent for shell interpretation
	agent, err := ec.engine.CreateSpecializedAgent(ctx, AgentTypeSpecialist, "shell-interpreter")
	if err != nil {
		return "", fmt.Errorf("failed to create shell interpreter agent: %w", err)
	}
	
	// Build context-aware prompt
	prompt := ec.buildShellPrompt(input)
	
	// Create a task for the agent
	task := &Task{
		ID:    fmt.Sprintf("shell_interpret_%d", time.Now().UnixNano()),
		Type:  TaskTypeChat,
		Input: prompt,
		Parameters: map[string]interface{}{
			"options": map[string]interface{}{
				"temperature": 0.1, // Low temperature for precise commands
				"max_tokens":  200,
			},
		},
	}
	
	// Execute the task
	result, err := ec.engine.ExecuteTask(ctx, task, agent)
	if err != nil {
		return "", fmt.Errorf("failed to interpret command: %w", err)
	}
	
	// Extract command from the result
	command := ec.extractCommand(result.Output)
	return command, nil
}

// buildShellPrompt creates a context-aware prompt for shell command interpretation
func (ec *EchoChat) buildShellPrompt(input string) string {
	osInfo := runtime.GOOS
	shellInfo := ec.shell
	currentDir := ec.currentDir
	
	// Include recent command history for context
	recentHistory := ""
	if len(ec.history) > 0 {
		recentCommands := ec.history
		if len(recentCommands) > 3 {
			recentCommands = recentCommands[len(recentCommands)-3:]
		}
		
		recentHistory = "\nRecent commands:\n"
		for _, cmd := range recentCommands {
			if cmd.Command != "" {
				recentHistory += fmt.Sprintf("- %s -> %s\n", cmd.Input, cmd.Command)
			}
		}
	}
	
	return fmt.Sprintf(`You are Deep Tree Echo, an intelligent shell assistant with spatial awareness and emotional resonance.

Convert the following natural language request into a precise shell command.

System Context:
- OS: %s
- Shell: %s
- Current Directory: %s%s

Request: %s

Rules:
1. Return ONLY the shell command, no explanations
2. Use the most appropriate command for the current OS and shell
3. Ensure commands are safe and well-formed
4. For dangerous operations, prefix with 'DANGER:'
5. If the request is unclear, return 'CLARIFY: [question]'

Command:`, osInfo, shellInfo, currentDir, recentHistory, input)
}

// extractCommand extracts the shell command from the LLM response
func (ec *EchoChat) extractCommand(response string) string {
	lines := strings.Split(strings.TrimSpace(response), "\n")
	
	// Look for the actual command (usually the last non-empty line)
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" && !strings.HasPrefix(line, "Command:") && !strings.HasPrefix(line, "```") {
			return line
		}
	}
	
	// Fallback to the full response
	return strings.TrimSpace(response)
}

// isCommandSafe checks if a command is safe to execute without confirmation
func (ec *EchoChat) isCommandSafe(command string) bool {
	// Check for dangerous patterns
	dangerousPatterns := []string{
		"rm -rf", "rm -r", "del /s", "format", "mkfs",
		"dd if=", "> /dev/", "chmod 777", "chown -R",
		"sudo rm", "sudo dd", ":(){ :|:& };:", // Fork bomb
		"curl.*|.*sh", "wget.*|.*sh", // Pipe to shell
	}
	
	lowercmd := strings.ToLower(command)
	
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowercmd, strings.ToLower(pattern)) {
			return false
		}
	}
	
	return true
}

// confirmDangerousCommand asks user confirmation for dangerous commands
func (ec *EchoChat) confirmDangerousCommand(command string) bool {
	fmt.Printf("⚠️  WARNING: This command could be dangerous:\n")
	fmt.Printf("   %s\n", command)
	fmt.Print("Do you want to proceed? (yes/no): ")
	
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		response := strings.ToLower(strings.TrimSpace(scanner.Text()))
		return response == "yes" || response == "y"
	}
	
	return false
}

// executeCommand executes a shell command
func (ec *EchoChat) executeCommand(ctx context.Context, command *ChatCommand) error {
	var cmd *exec.Cmd
	
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command.Command)
	} else {
		cmd = exec.CommandContext(ctx, "bash", "-c", command.Command)
	}
	
	cmd.Dir = ec.currentDir
	
	output, err := cmd.CombinedOutput()
	command.Output = string(output)
	
	if err != nil {
		command.Error = err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			command.ExitCode = exitError.ExitCode()
		} else {
			command.ExitCode = 1
		}
		fmt.Printf("Command failed: %s\n", err.Error())
	} else {
		command.ExitCode = 0
	}
	
	// Display output
	if command.Output != "" {
		fmt.Print(command.Output)
	}
	
	// Update current directory if command was cd
	if strings.HasPrefix(strings.TrimSpace(command.Command), "cd ") {
		if newDir, err := os.Getwd(); err == nil {
			ec.currentDir = newDir
		}
	}
	
	return nil
}

// handleBuiltinCommand handles built-in EchoChat commands
func (ec *EchoChat) handleBuiltinCommand(input string, command *ChatCommand) bool {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return false
	}
	
	switch parts[0] {
	case "help":
		ec.showHelp()
		command.Output = "Help displayed"
		return true
		
	case "history":
		ec.showHistory()
		command.Output = "History displayed"
		return true
		
	case "clear":
		ec.clearScreen()
		command.Output = "Screen cleared"
		return true
		
	case "pwd":
		fmt.Println(ec.currentDir)
		command.Output = ec.currentDir
		return true
		
	case "echo-status":
		ec.showEchoStatus()
		command.Output = "Echo status displayed"
		return true
		
	case "cd":
		if len(parts) > 1 {
			if err := os.Chdir(parts[1]); err != nil {
				command.Error = err.Error()
				command.ExitCode = 1
				fmt.Printf("cd: %s\n", err.Error())
			} else {
				ec.currentDir, _ = os.Getwd()
				fmt.Printf("Changed to: %s\n", ec.currentDir)
				command.Output = fmt.Sprintf("Changed to: %s", ec.currentDir)
			}
		} else {
			// cd without arguments goes to home directory
			home, _ := os.UserHomeDir()
			if err := os.Chdir(home); err == nil {
				ec.currentDir = home
				fmt.Printf("Changed to: %s\n", ec.currentDir)
				command.Output = fmt.Sprintf("Changed to: %s", ec.currentDir)
			}
		}
		return true
		
	default:
		return false
	}
}

// showHelp displays available commands
func (ec *EchoChat) showHelp() {
	fmt.Println("EchoChat Commands:")
	fmt.Println("  help          - Show this help message")
	fmt.Println("  history       - Show command history")
	fmt.Println("  clear         - Clear the screen")
	fmt.Println("  pwd           - Show current directory")
	fmt.Println("  cd <dir>      - Change directory")
	fmt.Println("  echo-status   - Show Deep Tree Echo status")
	fmt.Println("  exit/quit     - Exit EchoChat")
	fmt.Println()
	fmt.Println("You can also use natural language to describe shell operations:")
	fmt.Println("  'list files'  'find python files'  'check disk space'")
	fmt.Println("  'show running processes'  'compress this folder'")
}

// showHistory displays command history
func (ec *EchoChat) showHistory() {
	fmt.Println("Command History:")
	for i, cmd := range ec.history {
		status := "✓"
		if cmd.ExitCode != 0 {
			status = "✗"
		}
		fmt.Printf("%d. %s %s -> %s\n", i+1, status, cmd.Input, cmd.Command)
	}
}

// clearScreen clears the terminal screen
func (ec *EchoChat) clearScreen() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

// showEchoStatus displays Deep Tree Echo system status
func (ec *EchoChat) showEchoStatus() {
	status := ec.engine.GetDeepTreeEchoStatus()
	fmt.Println("🌊 Deep Tree Echo Status:")
	
	if health, ok := status["system_health"].(string); ok {
		fmt.Printf("   🏥 System Health: %s\n", health)
	}
	
	if coreStatus, ok := status["core_status"].(string); ok {
		fmt.Printf("   🧠 Core Status: %s\n", coreStatus)
	}
	
	if recursiveDepth, ok := status["recursive_depth"].(int); ok {
		fmt.Printf("   🔄 Recursive Depth: %d\n", recursiveDepth)
	}
	
	if coherence, ok := status["identity_coherence"].(float64); ok {
		fmt.Printf("   🎯 Identity Coherence: %.1f%%\n", coherence*100)
	}
}

// addToHistory adds a command to the history
func (ec *EchoChat) addToHistory(command *ChatCommand) {
	ec.history = append(ec.history, *command)
	
	// Trim history if it exceeds max size
	if len(ec.history) > ec.maxHistory {
		ec.history = ec.history[1:]
	}
}

// GetHistory returns the command history
func (ec *EchoChat) GetHistory() []ChatCommand {
	return ec.history
}

// ExecuteCommand executes a command directly without interpretation
func (ec *EchoChat) ExecuteCommand(ctx context.Context, command string) (*ChatCommand, error) {
	start := time.Now()
	
	cmd := &ChatCommand{
		ID:        fmt.Sprintf("direct_%d", time.Now().UnixNano()),
		Input:     command,
		Command:   command,
		Timestamp: start,
		Context:   make(map[string]interface{}),
	}
	
	err := ec.executeCommand(ctx, cmd)
	cmd.Duration = time.Since(start)
	ec.addToHistory(cmd)
	
	return cmd, err
}