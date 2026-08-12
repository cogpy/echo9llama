package llm

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/cogpy/echo9llama/core/backendcap"
)

// ProviderMode controls how capability scoring biases local and remote substrates.
type ProviderMode string

const (
	ProviderModeBalanced    ProviderMode = "balanced"
	ProviderModeLocalFirst  ProviderMode = "local_first"
	ProviderModeRemoteFirst ProviderMode = "remote_first"
	ProviderModeOffline     ProviderMode = "offline"
)

// RoutingOptions lets a cognitive task express substrate requirements without
// coupling itself to concrete provider names.
type RoutingOptions struct {
	TraceID               string
	Intent                string
	NeedOffline           bool
	PreferNative          bool
	RequireRealModel      bool
	AllowRemote           bool
	AllowFallback         bool
	PolicyExplicit        bool
	AllowQueue            bool
	RequiredContextTokens int
	LatencyClass          backendcap.LatencyClass
	QueueWait             time.Duration
}

// RouteAttempt records one provider attempt without prompts, outputs, keys, or paths.
type RouteAttempt struct {
	Provider      string                 `json:"provider"`
	BackendKind   backendcap.BackendKind `json:"backend_kind"`
	ModelID       string                 `json:"model_id,omitempty"`
	Success       bool                   `json:"success"`
	Skipped       bool                   `json:"skipped"`
	ErrorCategory string                 `json:"error_category,omitempty"`
	Latency       time.Duration          `json:"latency"`
	Timestamp     time.Time              `json:"timestamp"`
}

// BackendRoutingState is a bounded, path-scrubbed production status snapshot.
type BackendRoutingState struct {
	TraceID          string                  `json:"trace_id,omitempty"`
	Mode             ProviderMode            `json:"mode"`
	Decision         backendcap.Decision     `json:"decision"`
	Attempts         []RouteAttempt          `json:"attempts,omitempty"`
	SelectedProvider string                  `json:"selected_provider,omitempty"`
	SelectedKind     backendcap.BackendKind  `json:"selected_kind,omitempty"`
	SelectedModelID  string                  `json:"selected_model_id,omitempty"`
	Degraded         bool                    `json:"degraded"`
	FallbackCount    uint64                  `json:"fallback_count"`
	LocalModel       LocalModelRegistryState `json:"local_model"`
	UpdatedAt        time.Time               `json:"updated_at"`
}

// CapabilityProvider supplies dynamic provider capacity to the router.
type CapabilityProvider interface {
	CapabilitySnapshot() backendcap.Capability
}

// LocalRuntimeController is owned by MultiProviderLLM and coordinated by UAO wake/rest.
type LocalRuntimeController interface {
	Warmup(context.Context) error
	Cooldown(string) bool
	Refresh() LocalModelRegistryState
	State() LocalModelRegistryState
	UnloadIdle(string) bool
	MaybeUnloadForMemoryPressure(string) bool
	Close() error
}

func parseProviderMode(value string) ProviderMode {
	switch ProviderMode(strings.ToLower(strings.TrimSpace(value))) {
	case ProviderModeLocalFirst:
		return ProviderModeLocalFirst
	case ProviderModeRemoteFirst:
		return ProviderModeRemoteFirst
	case ProviderModeOffline:
		return ProviderModeOffline
	default:
		return ProviderModeBalanced
	}
}

func normalizeRouting(options RoutingOptions, maxOutput int, mode ProviderMode) backendcap.Workload {
	allowRemote := options.AllowRemote
	allowFallback := options.AllowFallback
	if !options.PolicyExplicit {
		allowRemote = true
		allowFallback = true
	}
	needOffline := options.NeedOffline || mode == ProviderModeOffline
	if needOffline {
		allowRemote = false
	}
	preferNative := options.PreferNative
	if mode == ProviderModeLocalFirst || mode == ProviderModeOffline {
		preferNative = true
	}
	intent := strings.TrimSpace(options.Intent)
	if intent == "" {
		intent = "general"
	}
	latency := options.LatencyClass
	if latency == "" {
		latency = backendcap.LatencyNormal
	}
	return backendcap.Workload{
		Intent:                intent,
		NeedOffline:           needOffline,
		PreferNative:          preferNative,
		RequireRealModel:      options.RequireRealModel,
		AllowRemote:           allowRemote,
		AllowFallback:         allowFallback,
		AllowQueue:            options.AllowQueue,
		MinMemoryTier:         backendcap.MemoryConstrained,
		RequiredContextTokens: options.RequiredContextTokens,
		ExpectedOutputTokens:  maxOutput,
		LatencyClass:          latency,
	}
}

func capabilityForProvider(provider Provider, index int, mode ProviderMode) backendcap.Capability {
	if provider == nil {
		return backendcap.Capability{Name: "nil-provider", Available: false, MemorySafe: true, Reason: "provider is nil"}
	}
	if capabilityProvider, ok := provider.(CapabilityProvider); ok {
		capability := capabilityProvider.CapabilitySnapshot()
		if capability.ProviderName == "" {
			capability.ProviderName = provider.Name()
		}
		capability.ModelPath = ""
		capability.Priority += providerModePriority(capability, index, mode)
		return capability
	}

	name := provider.Name()
	lowerName := strings.ToLower(name)
	kind := backendcap.BackendRemoteAPI
	offline := false
	concrete := true
	if strings.Contains(lowerName, "fallback") {
		kind = backendcap.BackendFallback
		offline = true
		concrete = false
	}
	capability := backendcap.Capability{
		Name:          name,
		ProviderName:  name,
		Kind:          kind,
		Available:     provider.Available(),
		Offline:       offline,
		Concrete:      concrete,
		MemorySafe:    true,
		MemoryTier:    backendcap.MemoryConstrained,
		ContextLength: provider.MaxTokens(),
		Priority:      providerModePriority(backendcap.Capability{Kind: kind}, index, mode),
	}
	return capability
}

func providerModePriority(capability backendcap.Capability, index int, mode ProviderMode) int {
	// Preserve configured order within each substrate class.
	priority := 100 - index
	switch mode {
	case ProviderModeLocalFirst, ProviderModeOffline:
		if capability.Native {
			priority += 500
		}
		if capability.Kind == backendcap.BackendRemoteAPI {
			priority -= 500
		}
	case ProviderModeRemoteFirst:
		if capability.Kind == backendcap.BackendRemoteAPI {
			priority += 500
		}
		if capability.Native {
			priority -= 100
		}
	case ProviderModeBalanced:
		if capability.Native {
			priority += 120
		}
		if capability.Kind == backendcap.BackendRemoteAPI {
			priority += 100
		}
	}
	if capability.Kind == backendcap.BackendFallback {
		priority -= 1000
	}
	return priority
}

func routeErrorCategory(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, context.Canceled):
		return "canceled"
	case errors.Is(err, context.DeadlineExceeded):
		return "deadline"
	case errors.Is(err, ErrLocalGGUFQueueSaturated):
		return "queue_saturated"
	case errors.Is(err, ErrLocalGGUFUnavailable):
		return "unavailable"
	case errors.Is(err, ErrLocalGGUFClosed):
		return "closed"
	default:
		return "provider"
	}
}

func cloneAttempts(values []RouteAttempt) []RouteAttempt {
	return append([]RouteAttempt(nil), values...)
}
