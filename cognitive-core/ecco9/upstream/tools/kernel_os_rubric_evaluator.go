//go:build tools
// +build tools

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// KernelPrimitive represents a core kernel functionality category
type KernelPrimitive struct {
	Name            string
	Weight          int
	Keywords        []string
	TargetFunctions int
	TargetSLOC      int
	Description     string
}

// OSService represents a platform-level OS service
type OSService struct {
	Name            string
	Weight          int
	Keywords        []string
	TargetFunctions int
	TargetSLOC      int
	Description     string
}

// Evidence represents found evidence for a category
type Evidence struct {
	Files             []string
	MatchedKeywords   []string
	EstimatedSLOC     int
	EstimatedFuncs    int
	PresenceScore     float64
	CompletenessScore float64
	WeightedScore     float64
}

// EvaluationResult contains the complete evaluation
type EvaluationResult struct {
	KernelPrimitives map[string]*Evidence
	OSServices       map[string]*Evidence
	KernelScore      float64
	OSScore          float64
	Classification   string
	Summary          string
}

var kernelPrimitives = []KernelPrimitive{
	{
		Name:            "Boot / Initialisation",
		Weight:          10,
		Keywords:        []string{"boot", "stage0", "bootstrap", "init_membranes", "startup", "initialization"},
		TargetFunctions: 12,
		TargetSLOC:      5000,
		Description:     "System boot and initialization code",
	},
	{
		Name:            "CPU Scheduling",
		Weight:          9,
		Keywords:        []string{"scheduler", "sched_", "context", "switch", "runqueue", "dispatch", "tick"},
		TargetFunctions: 18,
		TargetSLOC:      8000,
		Description:     "CPU scheduling and task management",
	},
	{
		Name:            "Process/Thread Management",
		Weight:          8,
		Keywords:        []string{"process", "thread", "fork", "spawn", "create", "destroy", "task"},
		TargetFunctions: 24,
		TargetSLOC:      10000,
		Description:     "Process and thread lifecycle management",
	},
	{
		Name:            "Memory Management",
		Weight:          8,
		Keywords:        []string{"malloc", "alloc", "free", "memory", "heap", "pool", "allocator"},
		TargetFunctions: 24,
		TargetSLOC:      12000,
		Description:     "Memory allocation and management",
	},
	{
		Name:            "Interrupt Handling & Traps",
		Weight:          6,
		Keywords:        []string{"interrupt", "irq", "handler", "vector", "trap", "exception"},
		TargetFunctions: 15,
		TargetSLOC:      6000,
		Description:     "Hardware interrupt and trap handling",
	},
	{
		Name:            "System Call Interface",
		Weight:          5,
		Keywords:        []string{"syscall", "system_call", "abi", "entry"},
		TargetFunctions: 32,
		TargetSLOC:      10000,
		Description:     "System call gateway and interface",
	},
	{
		Name:            "Basic I/O Primitives",
		Weight:          5,
		Keywords:        []string{"io.", "read", "write", "device", "register", "hal"},
		TargetFunctions: 20,
		TargetSLOC:      7000,
		Description:     "Low-level I/O operations",
	},
	{
		Name:            "Low-level Synchronisation",
		Weight:          4,
		Keywords:        []string{"spinlock", "mutex", "atomic", "barrier", "lock", "sync"},
		TargetFunctions: 16,
		TargetSLOC:      4000,
		Description:     "Synchronization primitives",
	},
	{
		Name:            "Timers and Clock",
		Weight:          3,
		Keywords:        []string{"timer", "clock", "tick", "time", "quantum"},
		TargetFunctions: 10,
		TargetSLOC:      3000,
		Description:     "Time management and timers",
	},
	{
		Name:            "Protection / Privilege Separation",
		Weight:          2,
		Keywords:        []string{"mmu", "protect", "privilege", "mode", "ring", "permission"},
		TargetFunctions: 14,
		TargetSLOC:      6000,
		Description:     "Memory protection and privilege levels",
	},
}

var osServices = []OSService{
	{
		Name:            "Virtual Memory / Paging",
		Weight:          8,
		Keywords:        []string{"page_table", "virtual", "mmu", "paging", "address"},
		TargetFunctions: 28,
		TargetSLOC:      15000,
		Description:     "Virtual memory and address translation",
	},
	{
		Name:            "Device Driver Framework",
		Weight:          8,
		Keywords:        []string{"driver", "register_driver", "bus", "hal", "device"},
		TargetFunctions: 35,
		TargetSLOC:      20000,
		Description:     "Device driver infrastructure",
	},
	{
		Name:            "Filesystem / VFS",
		Weight:          7,
		Keywords:        []string{"fs.", "vfs", "file", "inode", "dentry", "mount"},
		TargetFunctions: 42,
		TargetSLOC:      25000,
		Description:     "Filesystem and virtual file system",
	},
	{
		Name:            "Networking Stack",
		Weight:          5,
		Keywords:        []string{"socket", "network", "tcp", "udp", "ip", "packet"},
		TargetFunctions: 58,
		TargetSLOC:      35000,
		Description:     "Network protocol stack",
	},
	{
		Name:            "Inter-process Communication (IPC)",
		Weight:          4,
		Keywords:        []string{"message", "queue", "pipe", "signal", "ipc", "psystem"},
		TargetFunctions: 18,
		TargetSLOC:      10000,
		Description:     "IPC mechanisms and message passing",
	},
	{
		Name:            "Security Subsystems",
		Weight:          3,
		Keywords:        []string{"crypto", "auth", "capability", "attestation", "security"},
		TargetFunctions: 30,
		TargetSLOC:      18000,
		Description:     "Security and cryptographic services",
	},
	{
		Name:            "Power Management",
		Weight:          3,
		Keywords:        []string{"power", "sleep", "wake", "energy", "suspend"},
		TargetFunctions: 22,
		TargetSLOC:      12000,
		Description:     "Power management and energy control",
	},
	{
		Name:            "Profiling & Debug",
		Weight:          2,
		Keywords:        []string{"profiler", "trace", "debug", "perf", "monitor"},
		TargetFunctions: 25,
		TargetSLOC:      15000,
		Description:     "Performance profiling and debugging",
	},
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run kernel_os_rubric_evaluator.go <repository_path>")
		os.Exit(1)
	}

	repoPath := os.Args[1]
	fmt.Printf("Evaluating repository: %s\n\n", repoPath)

	result := &EvaluationResult{
		KernelPrimitives: make(map[string]*Evidence),
		OSServices:       make(map[string]*Evidence),
	}

	// Evaluate kernel primitives
	fmt.Println("=== Evaluating Kernel Primitives ===")
	totalKernelWeight := 0
	for _, primitive := range kernelPrimitives {
		totalKernelWeight += primitive.Weight
		evidence := evaluateCategory(repoPath, primitive.Keywords, primitive.TargetFunctions, primitive.TargetSLOC)
		evidence.WeightedScore = evidence.PresenceScore * evidence.CompletenessScore * float64(primitive.Weight)
		result.KernelPrimitives[primitive.Name] = evidence

		fmt.Printf("\n%s (Weight: %d):\n", primitive.Name, primitive.Weight)
		fmt.Printf("  Files Found: %d\n", len(evidence.Files))
		if len(evidence.Files) > 0 {
			fmt.Printf("  Sample Files: %v\n", evidence.Files[:min(3, len(evidence.Files))])
		}
		fmt.Printf("  Matched Keywords: %v\n", evidence.MatchedKeywords)
		fmt.Printf("  Estimated SLOC: %d / %d\n", evidence.EstimatedSLOC, primitive.TargetSLOC)
		fmt.Printf("  Presence: %.2f, Completeness: %.2f\n", evidence.PresenceScore, evidence.CompletenessScore)
		fmt.Printf("  Weighted Score: %.2f\n", evidence.WeightedScore)
	}

	// Calculate kernel score
	kernelRawScore := 0.0
	for _, evidence := range result.KernelPrimitives {
		kernelRawScore += evidence.WeightedScore
	}
	result.KernelScore = (kernelRawScore / float64(totalKernelWeight)) * 100

	// Evaluate OS services
	fmt.Println("\n\n=== Evaluating OS Platform Services ===")
	totalOSWeight := 0
	for _, service := range osServices {
		totalOSWeight += service.Weight
		evidence := evaluateCategory(repoPath, service.Keywords, service.TargetFunctions, service.TargetSLOC)
		evidence.WeightedScore = evidence.PresenceScore * evidence.CompletenessScore * float64(service.Weight)
		result.OSServices[service.Name] = evidence

		fmt.Printf("\n%s (Weight: %d):\n", service.Name, service.Weight)
		fmt.Printf("  Files Found: %d\n", len(evidence.Files))
		if len(evidence.Files) > 0 {
			fmt.Printf("  Sample Files: %v\n", evidence.Files[:min(3, len(evidence.Files))])
		}
		fmt.Printf("  Matched Keywords: %v\n", evidence.MatchedKeywords)
		fmt.Printf("  Estimated SLOC: %d / %d\n", evidence.EstimatedSLOC, service.TargetSLOC)
		fmt.Printf("  Presence: %.2f, Completeness: %.2f\n", evidence.PresenceScore, evidence.CompletenessScore)
		fmt.Printf("  Weighted Score: %.2f\n", evidence.WeightedScore)
	}

	// Calculate OS score
	osRawScore := 0.0
	for _, evidence := range result.OSServices {
		osRawScore += evidence.WeightedScore
	}
	result.OSScore = (osRawScore / float64(totalOSWeight)) * 100

	// Classify repository
	result.Classification = classifyRepository(result.KernelScore, result.OSScore)
	result.Summary = generateSummary(result)

	// Print final results
	fmt.Println("\n\n" + strings.Repeat("=", 70))
	fmt.Println("FINAL EVALUATION RESULTS")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("\nKernel Score: %.2f / 100\n", result.KernelScore)
	fmt.Printf("OS Score: %.2f / 100\n", result.OSScore)
	fmt.Printf("\nClassification: %s\n", result.Classification)
	fmt.Printf("\n%s\n", result.Summary)

	// Save detailed report
	saveReport(result, repoPath)
}

func evaluateCategory(repoPath string, keywords []string, targetFuncs, targetSLOC int) *Evidence {
	evidence := &Evidence{
		Files:           []string{},
		MatchedKeywords: []string{},
	}

	// Search for files containing keywords
	matchedFiles := make(map[string]bool)
	keywordMatches := make(map[string]bool)

	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories and non-source files
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only check source files
		ext := filepath.Ext(path)
		if ext != ".go" && ext != ".c" && ext != ".cpp" && ext != ".h" && ext != ".rs" && ext != ".asm" {
			return nil
		}

		// Read file and check for keywords
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		contentLower := strings.ToLower(string(content))
		filename := strings.ToLower(filepath.Base(path))

		for _, keyword := range keywords {
			keywordLower := strings.ToLower(keyword)
			if strings.Contains(filename, keywordLower) || strings.Contains(contentLower, keywordLower) {
				relPath, _ := filepath.Rel(repoPath, path)
				matchedFiles[relPath] = true
				keywordMatches[keyword] = true

				// Estimate SLOC for this file
				lines := strings.Count(string(content), "\n")
				evidence.EstimatedSLOC += lines
				break
			}
		}

		return nil
	})

	// Convert maps to slices
	for file := range matchedFiles {
		evidence.Files = append(evidence.Files, file)
	}
	for keyword := range keywordMatches {
		evidence.MatchedKeywords = append(evidence.MatchedKeywords, keyword)
	}

	// Calculate presence score
	if len(evidence.Files) == 0 {
		evidence.PresenceScore = 0.0
	} else if len(evidence.Files) < 3 {
		evidence.PresenceScore = 0.5
	} else {
		evidence.PresenceScore = 1.0
	}

	// Estimate functions (rough approximation: 1 function per 30 lines of code)
	evidence.EstimatedFuncs = evidence.EstimatedSLOC / 30

	// Calculate completeness score
	slocRatio := float64(evidence.EstimatedSLOC) / float64(targetSLOC)
	funcRatio := float64(evidence.EstimatedFuncs) / float64(targetFuncs)

	// Use the average of both ratios, capped at 1.0
	avgRatio := (slocRatio + funcRatio) / 2.0
	if avgRatio > 1.0 {
		evidence.CompletenessScore = 1.0
	} else {
		evidence.CompletenessScore = avgRatio
	}

	return evidence
}

func classifyRepository(kernelScore, osScore float64) string {
	if kernelScore >= 60 && osScore > 40 {
		return "Kernel-grade"
	} else if kernelScore >= 30 && kernelScore < 60 {
		return "Kernel-prototype"
	} else if kernelScore < 30 && osScore >= 50 {
		return "OS-platform"
	} else {
		return "Application / other"
	}
}

func generateSummary(result *EvaluationResult) string {
	var summary strings.Builder

	summary.WriteString("Summary:\n")
	summary.WriteString("--------\n")

	if result.Classification == "Application / other" {
		summary.WriteString("This repository is primarily an application-level codebase and does not\n")
		summary.WriteString("implement significant kernel or OS primitives. It appears to be focused on\n")
		summary.WriteString("higher-level functionality rather than low-level system programming.\n\n")
	} else if result.Classification == "Kernel-prototype" {
		summary.WriteString("This repository implements some kernel primitives but is missing critical\n")
		summary.WriteString("components required for a full kernel implementation. It may be a research\n")
		summary.WriteString("prototype or early-stage kernel development.\n\n")
	} else if result.Classification == "OS-platform" {
		summary.WriteString("This repository provides platform-level OS services but lacks core kernel\n")
		summary.WriteString("primitives. It likely builds on top of an existing kernel.\n\n")
	} else {
		summary.WriteString("This repository contains substantial kernel primitives and OS platform code,\n")
		summary.WriteString("qualifying it as kernel-grade software.\n\n")
	}

	// List strongest categories
	summary.WriteString("Strongest Kernel Categories:\n")
	strongestKernel := findStrongestCategories(result.KernelPrimitives, 3)
	for _, name := range strongestKernel {
		evidence := result.KernelPrimitives[name]
		summary.WriteString(fmt.Sprintf("  - %s (score: %.2f)\n", name, evidence.WeightedScore))
	}

	summary.WriteString("\nStrongest OS Service Categories:\n")
	strongestOS := findStrongestCategories(result.OSServices, 3)
	for _, name := range strongestOS {
		evidence := result.OSServices[name]
		summary.WriteString(fmt.Sprintf("  - %s (score: %.2f)\n", name, evidence.WeightedScore))
	}

	// AGI-OS readiness assessment
	summary.WriteString("\nAGI-OS Readiness Assessment:\n")
	if result.KernelScore >= 70 && result.OSScore >= 70 {
		summary.WriteString("  ✓ Repository meets AGI-OS readiness thresholds\n")
	} else {
		summary.WriteString("  ✗ Repository does not meet AGI-OS readiness thresholds\n")
		summary.WriteString(fmt.Sprintf("    (requires kernel_score >= 70 and os_score >= 70)\n"))
	}

	return summary.String()
}

func findStrongestCategories(categories map[string]*Evidence, limit int) []string {
	type categoryScore struct {
		name  string
		score float64
	}

	scores := []categoryScore{}
	for name, evidence := range categories {
		scores = append(scores, categoryScore{name, evidence.WeightedScore})
	}

	// Sort by score descending
	for i := 0; i < len(scores); i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].score > scores[i].score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	result := []string{}
	for i := 0; i < min(limit, len(scores)); i++ {
		result = append(result, scores[i].name)
	}
	return result
}

func saveReport(result *EvaluationResult, repoPath string) {
	// Save JSON report
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	reportPath := filepath.Join(repoPath, "KERNEL_OS_EVALUATION.json")
	err = os.WriteFile(reportPath, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing JSON report: %v\n", err)
		return
	}
	fmt.Printf("\nDetailed JSON report saved to: %s\n", reportPath)

	// Save markdown report
	mdReport := generateMarkdownReport(result)
	mdPath := filepath.Join(repoPath, "KERNEL_OS_EVALUATION.md")
	err = os.WriteFile(mdPath, []byte(mdReport), 0644)
	if err != nil {
		fmt.Printf("Error writing Markdown report: %v\n", err)
		return
	}
	fmt.Printf("Markdown report saved to: %s\n", mdPath)
}

func generateMarkdownReport(result *EvaluationResult) string {
	var md strings.Builder

	md.WriteString("# Kernel/OS Functionality Evaluation Report\n\n")
	md.WriteString("This report evaluates the repository against a comprehensive kernel and operating system functionality rubric.\n\n")

	md.WriteString("## Overall Scores\n\n")
	md.WriteString(fmt.Sprintf("- **Kernel Score**: %.2f / 100\n", result.KernelScore))
	md.WriteString(fmt.Sprintf("- **OS Score**: %.2f / 100\n", result.OSScore))
	md.WriteString(fmt.Sprintf("- **Classification**: %s\n\n", result.Classification))

	md.WriteString("## Classification Criteria\n\n")
	md.WriteString("| Classification | Criteria | Interpretation |\n")
	md.WriteString("|---------------|----------|----------------|\n")
	md.WriteString("| Kernel-grade | kernel_score ≥ 60; os_score > 40 | Contains substantial core primitives and platform code |\n")
	md.WriteString("| Kernel-prototype | 30 ≤ kernel_score < 60 | Implements some core primitives but missing critical parts |\n")
	md.WriteString("| OS-platform | kernel_score < 30 and os_score ≥ 50 | Provides platform services on top of existing kernel |\n")
	md.WriteString("| Application / other | kernel_score < 30 and os_score < 50 | Primarily user-space code or unrelated library |\n\n")

	md.WriteString("## Kernel Primitives Evaluation\n\n")
	md.WriteString("| Primitive | Weight | Presence | Completeness | Weighted Score | Files Found |\n")
	md.WriteString("|-----------|--------|----------|--------------|----------------|-------------|\n")

	for _, primitive := range kernelPrimitives {
		evidence := result.KernelPrimitives[primitive.Name]
		md.WriteString(fmt.Sprintf("| %s | %d | %.2f | %.2f | %.2f | %d |\n",
			primitive.Name, primitive.Weight, evidence.PresenceScore,
			evidence.CompletenessScore, evidence.WeightedScore, len(evidence.Files)))
	}

	md.WriteString("\n## OS Platform Services Evaluation\n\n")
	md.WriteString("| Service | Weight | Presence | Completeness | Weighted Score | Files Found |\n")
	md.WriteString("|---------|--------|----------|--------------|----------------|-------------|\n")

	for _, service := range osServices {
		evidence := result.OSServices[service.Name]
		md.WriteString(fmt.Sprintf("| %s | %d | %.2f | %.2f | %.2f | %d |\n",
			service.Name, service.Weight, evidence.PresenceScore,
			evidence.CompletenessScore, evidence.WeightedScore, len(evidence.Files)))
	}

	md.WriteString("\n## Detailed Findings\n\n")
	md.WriteString("### Kernel Primitives\n\n")
	for _, primitive := range kernelPrimitives {
		evidence := result.KernelPrimitives[primitive.Name]
		md.WriteString(fmt.Sprintf("#### %s\n\n", primitive.Name))
		md.WriteString(fmt.Sprintf("- **Description**: %s\n", primitive.Description))
		md.WriteString(fmt.Sprintf("- **Weight**: %d\n", primitive.Weight))
		md.WriteString(fmt.Sprintf("- **Target**: %d functions, %d SLOC\n", primitive.TargetFunctions, primitive.TargetSLOC))
		md.WriteString(fmt.Sprintf("- **Found**: ~%d functions, %d SLOC\n", evidence.EstimatedFuncs, evidence.EstimatedSLOC))
		md.WriteString(fmt.Sprintf("- **Matched Keywords**: %v\n", evidence.MatchedKeywords))
		md.WriteString(fmt.Sprintf("- **Files Found**: %d\n", len(evidence.Files)))
		if len(evidence.Files) > 0 {
			md.WriteString("- **Sample Files**:\n")
			for i, file := range evidence.Files {
				if i >= 5 {
					md.WriteString(fmt.Sprintf("  - ... and %d more\n", len(evidence.Files)-5))
					break
				}
				md.WriteString(fmt.Sprintf("  - `%s`\n", file))
			}
		}
		md.WriteString("\n")
	}

	md.WriteString("### OS Platform Services\n\n")
	for _, service := range osServices {
		evidence := result.OSServices[service.Name]
		md.WriteString(fmt.Sprintf("#### %s\n\n", service.Name))
		md.WriteString(fmt.Sprintf("- **Description**: %s\n", service.Description))
		md.WriteString(fmt.Sprintf("- **Weight**: %d\n", service.Weight))
		md.WriteString(fmt.Sprintf("- **Target**: %d functions, %d SLOC\n", service.TargetFunctions, service.TargetSLOC))
		md.WriteString(fmt.Sprintf("- **Found**: ~%d functions, %d SLOC\n", evidence.EstimatedFuncs, evidence.EstimatedSLOC))
		md.WriteString(fmt.Sprintf("- **Matched Keywords**: %v\n", evidence.MatchedKeywords))
		md.WriteString(fmt.Sprintf("- **Files Found**: %d\n", len(evidence.Files)))
		if len(evidence.Files) > 0 {
			md.WriteString("- **Sample Files**:\n")
			for i, file := range evidence.Files {
				if i >= 5 {
					md.WriteString(fmt.Sprintf("  - ... and %d more\n", len(evidence.Files)-5))
					break
				}
				md.WriteString(fmt.Sprintf("  - `%s`\n", file))
			}
		}
		md.WriteString("\n")
	}

	md.WriteString("## Summary\n\n")
	md.WriteString(result.Summary)
	md.WriteString("\n")

	md.WriteString("## Recommendations\n\n")
	if result.Classification == "Application / other" {
		md.WriteString("This repository is not suitable as an AGI-OS kernel. To develop kernel functionality:\n\n")
		md.WriteString("1. Implement core boot and initialization code\n")
		md.WriteString("2. Develop a proper CPU scheduler with preemption\n")
		md.WriteString("3. Add process/thread management primitives\n")
		md.WriteString("4. Implement memory management with virtual memory support\n")
		md.WriteString("5. Add interrupt handling and system call interface\n")
	} else if result.Classification == "Kernel-prototype" {
		md.WriteString("This repository shows promise as a kernel prototype. To improve:\n\n")
		md.WriteString("1. Complete missing kernel primitives (identify gaps in evaluation above)\n")
		md.WriteString("2. Enhance existing implementations to meet target SLOC/function counts\n")
		md.WriteString("3. Add comprehensive testing for kernel components\n")
		md.WriteString("4. Implement missing OS platform services\n")
	}

	return md.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Unused but kept for potential future regex-based analysis
var _ = regexp.MustCompile
