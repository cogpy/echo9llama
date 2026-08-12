package llm

import (
	"errors"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/cogpy/echo9llama/core/backendcap"
)

const (
	defaultLocalGGUFContext   = 4096
	defaultLocalGGUFQueueWait = 5 * time.Second
)

var (
	// ErrLocalGGUFUnavailable reports that native GGUF support or a safe model is absent.
	ErrLocalGGUFUnavailable = errors.New("local GGUF provider unavailable")
	// ErrLocalGGUFQueueSaturated reports that the single native generation slot was not acquired in time.
	ErrLocalGGUFQueueSaturated = errors.New("local GGUF generation queue saturated")
	// ErrLocalGGUFClosed reports use after terminal provider closure.
	ErrLocalGGUFClosed = errors.New("local GGUF provider closed")
)

// LocalGGUFProviderConfig configures one registry-owned GGUF model.
type LocalGGUFProviderConfig struct {
	Name               string
	ModelPath          string
	ModelRoots         []string
	ContextSize        int
	BatchSize          int
	Threads            int
	GPULayers          int
	FlashAttention     bool
	KVCacheType        string
	Seed               uint32
	QueueWait          time.Duration
	AllowUnknownMemory bool
	MemorySafetyRatio  float64
	MemoryReserveBytes uint64
	Capability         backendcap.Capability
}

// LocalGGUFState is a copyable, path-scrubbed provider status surface.
type LocalGGUFState struct {
	Name                 string    `json:"name"`
	ModelID              string    `json:"model_id,omitempty"`
	Architecture         string    `json:"architecture,omitempty"`
	ContextSize          int       `json:"context_size"`
	Quantization         string    `json:"quantization,omitempty"`
	EstimatedMemoryBytes uint64    `json:"estimated_memory_bytes,omitempty"`
	NativeBuild          bool      `json:"native_build"`
	Available            bool      `json:"available"`
	Loaded               bool      `json:"loaded"`
	Loading              bool      `json:"loading"`
	Closed               bool      `json:"closed"`
	InFlight             int       `json:"in_flight"`
	LoadError            string    `json:"load_error,omitempty"`
	RetryAt              time.Time `json:"retry_at,omitempty"`
}

func defaultLocalGGUFConfig(modelPath string) LocalGGUFProviderConfig {
	return normalizeLocalGGUFConfig(LocalGGUFProviderConfig{ModelPath: modelPath})
}

func normalizeLocalGGUFConfig(config LocalGGUFProviderConfig) LocalGGUFProviderConfig {
	if strings.TrimSpace(config.Name) == "" {
		config.Name = "local_gguf"
	}
	if strings.TrimSpace(config.ModelPath) == "" {
		config.ModelPath = strings.TrimSpace(os.Getenv("LOCAL_MODEL_PATH"))
	}
	if config.ContextSize <= 0 {
		config.ContextSize = localEnvInt("LOCAL_MODEL_CONTEXT", localEnvInt("ECHO_MODEL_CONTEXT", defaultLocalGGUFContext))
	}
	if config.Capability.ContextLength > 0 && config.ContextSize > config.Capability.ContextLength {
		config.ContextSize = config.Capability.ContextLength
	}
	if config.ContextSize <= 1 {
		config.ContextSize = defaultLocalGGUFContext
	}
	if config.BatchSize <= 0 {
		config.BatchSize = localEnvInt("LOCAL_MODEL_BATCH", minLocalInt(config.ContextSize, 512))
	}
	if config.BatchSize > config.ContextSize {
		config.BatchSize = config.ContextSize
	}
	if config.Threads <= 0 {
		config.Threads = localEnvInt("LOCAL_MODEL_THREADS", maxLocalInt(1, runtime.NumCPU()/2))
	}
	if config.GPULayers == 0 {
		config.GPULayers = localEnvInt("LOCAL_MODEL_GPU_LAYERS", 0)
	}
	if strings.TrimSpace(config.KVCacheType) == "" {
		config.KVCacheType = localEnvString("LOCAL_MODEL_KV_CACHE", "f16")
	}
	if config.QueueWait <= 0 {
		config.QueueWait = localEnvDuration("ECHO_LOCAL_QUEUE_WAIT", defaultLocalGGUFQueueWait)
	}
	if !config.FlashAttention {
		config.FlashAttention = localEnvBool("LOCAL_MODEL_FLASH_ATTENTION", false)
	}
	if config.MemorySafetyRatio <= 0 || config.MemorySafetyRatio > 1 {
		config.MemorySafetyRatio = localEnvFloat("ECHO_MODEL_MEMORY_RATIO", 0.80)
	}
	if config.MemoryReserveBytes == 0 {
		config.MemoryReserveBytes = localEnvBytes("ECHO_MODEL_MEMORY_RESERVE", 1024*1024*1024)
	}
	if !config.AllowUnknownMemory {
		config.AllowUnknownMemory = localEnvBool("ECHO_MODEL_ALLOW_UNKNOWN_MEMORY", false)
	}
	return config
}

func (config LocalGGUFProviderConfig) discoveryOptions() backendcap.DiscoveryOptions {
	return backendcap.DiscoveryOptions{
		Roots:              append([]string(nil), config.ModelRoots...),
		AllowUnknownMemory: config.AllowUnknownMemory,
		MemorySafetyRatio:  config.MemorySafetyRatio,
		MemoryReserveBytes: config.MemoryReserveBytes,
	}
}

func localEnvString(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func localEnvInt(name string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv(name)))
	if err == nil {
		return value
	}
	return fallback
}

func localEnvFloat(name string, fallback float64) float64 {
	value, err := strconv.ParseFloat(strings.TrimSpace(os.Getenv(name)), 64)
	if err == nil {
		return value
	}
	return fallback
}

func localEnvBool(name string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(name))) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func localEnvDuration(name string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(strings.TrimSpace(os.Getenv(name)))
	if err == nil && value > 0 {
		return value
	}
	return fallback
}

func localEnvBytes(name string, fallback uint64) uint64 {
	raw := strings.ToUpper(strings.TrimSpace(os.Getenv(name)))
	if raw == "" {
		return fallback
	}
	multipliers := map[string]uint64{"B": 1, "KB": 1024, "MB": 1024 * 1024, "GB": 1024 * 1024 * 1024, "KIB": 1024, "MIB": 1024 * 1024, "GIB": 1024 * 1024 * 1024}
	for _, suffix := range []string{"GIB", "GB", "MIB", "MB", "KIB", "KB", "B"} {
		if strings.HasSuffix(raw, suffix) {
			value, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(raw, suffix)), 64)
			if err == nil && value >= 0 {
				return uint64(value * float64(multipliers[suffix]))
			}
			return fallback
		}
	}
	value, err := strconv.ParseUint(raw, 10, 64)
	if err == nil {
		return value
	}
	return fallback
}

func minLocalInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func maxLocalInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
