package backendcap

import (
	"fmt"
	"sort"
	"strings"
)

// MemoryTier describes the host-memory class required by a backend path.
type MemoryTier string

const (
	MemoryUnknown     MemoryTier = "unknown"
	MemoryConstrained MemoryTier = "constrained"
	MemoryStandard    MemoryTier = "standard"
	MemoryStress      MemoryTier = "stress"
)

// BackendKind describes the substrate class used by an inference backend.
type BackendKind string

const (
	BackendNativeCPU BackendKind = "native_cpu"
	BackendNativeGPU BackendKind = "native_gpu"
	BackendRemoteAPI BackendKind = "remote_api"
	BackendFallback  BackendKind = "fallback"
)

// LatencyClass describes how long a workload can wait for provider capacity.
type LatencyClass string

const (
	LatencyInteractive LatencyClass = "interactive"
	LatencyNormal      LatencyClass = "normal"
	LatencyBatch       LatencyClass = "batch"
)

// Capability describes one concrete, schedulable inference substrate. ModelPath
// is internal runtime state and must not be exposed through public status APIs.
type Capability struct {
	Name                 string      `json:"name"`
	ProviderName         string      `json:"provider_name,omitempty"`
	ModelID              string      `json:"model_id,omitempty"`
	Kind                 BackendKind `json:"kind"`
	Available            bool        `json:"available"`
	Native               bool        `json:"native"`
	Offline              bool        `json:"offline"`
	Concrete             bool        `json:"concrete"`
	MemorySafe           bool        `json:"memory_safe"`
	StressGrade          bool        `json:"stress_grade"`
	MemoryTier           MemoryTier  `json:"memory_tier"`
	BuildTags            []string    `json:"build_tags,omitempty"`
	ModelPath            string      `json:"-"`
	ContextLength        int         `json:"context_length,omitempty"`
	Quantization         string      `json:"quantization,omitempty"`
	EstimatedMemoryBytes uint64      `json:"estimated_memory_bytes,omitempty"`
	MaxConcurrency       int         `json:"max_concurrency,omitempty"`
	CurrentInFlight      int         `json:"current_in_flight,omitempty"`
	Priority             int         `json:"priority,omitempty"`
	Reason               string      `json:"reason,omitempty"`
	Guidance             string      `json:"guidance,omitempty"`
}

// Workload describes an inference task from the scheduler's perspective.
type Workload struct {
	Intent                string
	NeedOffline           bool
	PreferNative          bool
	RequireStress         bool
	RequireRealModel      bool
	AllowRemote           bool
	AllowFallback         bool
	AllowQueue            bool
	MinMemoryTier         MemoryTier
	RequiredContextTokens int
	ExpectedOutputTokens  int
	LatencyClass          LatencyClass
}

// Rejection records why a capability could not satisfy a workload.
type Rejection struct {
	Capability string `json:"capability"`
	Reason     string `json:"reason"`
}

// Decision captures the selected capability, ordered route, and rejected options.
type Decision struct {
	Selected     Capability   `json:"selected"`
	Route        []string     `json:"route,omitempty"`
	Degraded     bool         `json:"degraded"`
	Reason       string       `json:"reason"`
	Alternatives []Capability `json:"alternatives,omitempty"`
	Rejections   []Rejection  `json:"rejections,omitempty"`
}

// NativeRuntimeEnabled reports whether this build may load native GGUF models.
func NativeRuntimeEnabled() bool {
	return cgoEnabled
}

// Snapshot returns the build-independent continuity capability. Concrete native
// models and remote providers are added by their runtime owners.
func Snapshot() []Capability {
	return []Capability{fallbackCapability()}
}

// SnapshotWithModelPaths enriches the continuity snapshot with concrete GGUF files.
func SnapshotWithModelPaths(paths []string) []Capability {
	caps := append([]Capability{}, Snapshot()...)
	caps = append(caps, DiscoverModelCapabilities(paths, DiscoveryOptions{})...)
	sort.SliceStable(caps, func(i, j int) bool { return caps[i].Name < caps[j].Name })
	return caps
}

// Select chooses the best capability from the build-independent snapshot.
func Select(workload Workload) Decision {
	return SelectFromCapabilities(workload, Snapshot())
}

// SelectWithModelPaths chooses a capability after discovering configured GGUF files.
func SelectWithModelPaths(workload Workload, paths []string) Decision {
	return SelectFromCapabilities(workload, SnapshotWithModelPaths(paths))
}

// SelectFromCapabilities filters hard constraints before applying stable scoring.
func SelectFromCapabilities(workload Workload, caps []Capability) Decision {
	workload = normalizeWorkload(workload)
	candidates := make([]Capability, 0, len(caps))
	rejections := make([]Rejection, 0, len(caps))

	for _, capability := range caps {
		if reason := rejectionReason(capability, workload); reason != "" {
			rejections = append(rejections, Rejection{Capability: capability.Name, Reason: reason})
			continue
		}
		candidates = append(candidates, capability)
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return score(candidates[i], workload) > score(candidates[j], workload)
	})

	if len(candidates) == 0 {
		unavailable := Capability{
			Name:       "no_backend_available",
			Kind:       BackendFallback,
			Available:  false,
			MemorySafe: true,
			MemoryTier: MemoryConstrained,
			Reason:     "no backend capability satisfies the workload constraints",
			Guidance:   "Configure a safe GGUF model, enable an allowed remote provider, or permit deterministic continuity fallback.",
		}
		return Decision{Selected: unavailable, Degraded: true, Reason: unavailable.Reason, Rejections: rejections}
	}

	selected := candidates[0]
	route := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		name := strings.TrimSpace(candidate.ProviderName)
		if name == "" {
			name = candidate.Name
		}
		if name != "" && !contains(route, name) {
			route = append(route, name)
		}
	}

	degraded := selected.Kind == BackendFallback || (workload.PreferNative && !selected.Native)
	reason := fmt.Sprintf("selected %s as the highest-scoring capability for %s", selected.Name, workload.Intent)
	if degraded {
		reason = fmt.Sprintf("selected degraded capability %s to preserve cognitive continuity for %s", selected.Name, workload.Intent)
	}

	return Decision{
		Selected:     selected,
		Route:        route,
		Degraded:     degraded,
		Reason:       reason,
		Alternatives: append([]Capability(nil), candidates[1:]...),
		Rejections:   rejections,
	}
}

func normalizeWorkload(workload Workload) Workload {
	if strings.TrimSpace(workload.Intent) == "" {
		workload.Intent = "general"
	}
	if workload.LatencyClass == "" {
		workload.LatencyClass = LatencyNormal
	}
	if workload.MinMemoryTier == "" || workload.MinMemoryTier == MemoryUnknown {
		workload.MinMemoryTier = MemoryConstrained
	}
	// A zero-value Workload preserves existing behavior by allowing remote and
	// deterministic fallback providers. Explicit offline work overrides remote use.
	if !workload.NeedOffline && !workload.AllowRemote && !workload.AllowFallback &&
		!workload.PreferNative && !workload.RequireRealModel && workload.RequiredContextTokens == 0 &&
		workload.ExpectedOutputTokens == 0 {
		workload.AllowRemote = true
		workload.AllowFallback = true
	}
	if workload.NeedOffline {
		workload.AllowRemote = false
	}
	return workload
}

func rejectionReason(capability Capability, workload Workload) string {
	if !capability.Available {
		return "unavailable"
	}
	if !capability.MemorySafe {
		return "memory safety policy rejected capability"
	}
	if workload.RequireRealModel && (!capability.Concrete || capability.Kind == BackendFallback) {
		return "workload requires a concrete real model"
	}
	if workload.NeedOffline && !capability.Offline {
		return "workload requires offline execution"
	}
	if capability.Kind == BackendRemoteAPI && !workload.AllowRemote {
		return "remote providers are not allowed"
	}
	if capability.Kind == BackendFallback && !workload.AllowFallback {
		return "deterministic fallback is not allowed"
	}
	if workload.RequireStress && !capability.StressGrade {
		return "stress-grade capability required"
	}
	if !satisfiesMemoryTier(capability.MemoryTier, workload.MinMemoryTier) {
		return fmt.Sprintf("memory tier %s is below required tier %s", capability.MemoryTier, workload.MinMemoryTier)
	}
	requiredContext := workload.RequiredContextTokens + workload.ExpectedOutputTokens
	if requiredContext > 0 && capability.ContextLength > 0 && capability.ContextLength < requiredContext {
		return fmt.Sprintf("context length %d is below required %d", capability.ContextLength, requiredContext)
	}
	if capability.MaxConcurrency > 0 && capability.CurrentInFlight >= capability.MaxConcurrency && !workload.AllowQueue {
		return "provider concurrency capacity is saturated"
	}
	return ""
}

func score(capability Capability, workload Workload) int {
	value := capability.Priority
	if capability.Native && workload.PreferNative {
		value += 60
	}
	if capability.Offline && workload.NeedOffline {
		value += 50
	}
	if capability.Concrete {
		value += 30
	}
	if capability.ModelPath != "" {
		value += 10
	}
	if workload.RequiredContextTokens > 0 && capability.ContextLength >= workload.RequiredContextTokens+workload.ExpectedOutputTokens {
		value += 10
	}
	if capability.MaxConcurrency <= 0 || capability.CurrentInFlight < capability.MaxConcurrency {
		value += 8
	}
	switch capability.Kind {
	case BackendNativeCPU, BackendNativeGPU:
		value += 20
	case BackendRemoteAPI:
		value += 15
	case BackendFallback:
		value -= 1000
	}
	if capability.StressGrade && workload.RequireStress {
		value += 20
	}
	return value
}

func fallbackCapability() Capability {
	return Capability{
		Name:          "simple_fallback",
		ProviderName:  "simple_fallback",
		Kind:          BackendFallback,
		Available:     true,
		Offline:       true,
		MemorySafe:    true,
		MemoryTier:    MemoryConstrained,
		ContextLength: 4096,
		Priority:      -100,
		Reason:        "always available as a degraded cognitive continuity surface",
		Guidance:      "Use only when no real inference substrate can satisfy the workload.",
	}
}

func satisfiesMemoryTier(actual, required MemoryTier) bool {
	return tierRank(actual) >= tierRank(required)
}

func tierRank(tier MemoryTier) int {
	switch tier {
	case MemoryStress:
		return 3
	case MemoryStandard:
		return 2
	case MemoryConstrained:
		return 1
	default:
		return 0
	}
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
