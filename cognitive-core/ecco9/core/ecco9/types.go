package ecco9

import (
	"context"
	"time"
)

// DeviceState represents the state of a cognitive device
type DeviceState struct {
	ID         string
	Name       string
	Status     DeviceStatus
	Power      PowerState
	Health     HealthStatus
	Metrics    DeviceMetrics
	LastUpdate time.Time
	Uptime     time.Duration
}

// DeviceStatus represents the operational status of a device
type DeviceStatus string

const (
	DeviceStatusOffline      DeviceStatus = "offline"
	DeviceStatusInitializing DeviceStatus = "initializing"
	DeviceStatusReady        DeviceStatus = "ready"
	DeviceStatusActive       DeviceStatus = "active"
	DeviceStatusError        DeviceStatus = "error"
	DeviceStatusMaintenance  DeviceStatus = "maintenance"
)

// PowerState represents power management state
type PowerState string

const (
	PowerStateOff     PowerState = "off"
	PowerStateStandby PowerState = "standby"
	PowerStateActive  PowerState = "active"
	PowerStateSleep   PowerState = "sleep"
)

// HealthStatus represents device health
type HealthStatus string

const (
	HealthStatusHealthy  HealthStatus = "healthy"
	HealthStatusDegraded HealthStatus = "degraded"
	HealthStatusCritical HealthStatus = "critical"
	HealthStatusFailed   HealthStatus = "failed"
)

// DeviceMetrics contains telemetry data for a device
type DeviceMetrics struct {
	CPUUsage       float64
	MemoryUsage    float64
	DiskUsage      float64
	NetworkRx      uint64
	NetworkTx      uint64
	OperationCount uint64
	ErrorCount     uint64
	AverageLatency time.Duration
	LastOperation  time.Time
}

// CognitiveDevice defines the interface for all ecco9 cognitive devices
type CognitiveDevice interface {
	// Lifecycle management
	Initialize(ctx context.Context) error
	Shutdown(ctx context.Context) error
	Reset(ctx context.Context) error

	// State management
	GetState() (DeviceState, error)
	SetState(DeviceState) error

	// I/O operations
	Read(buffer []byte) (int, error)
	Write(buffer []byte) (int, error)
	IoCtl(command uint32, arg interface{}) error

	// Telemetry
	GetMetrics() (DeviceMetrics, error)
	GetHealth() (HealthStatus, error)

	// Device information
	GetID() string
	GetName() string
	GetType() DeviceType
}

// DeviceType represents the category of cognitive device
type DeviceType string

const (
	DeviceTypeReservoir     DeviceType = "reservoir"
	DeviceTypeConsciousness DeviceType = "consciousness"
	DeviceTypeEmotion       DeviceType = "emotion"
	DeviceTypeMemory        DeviceType = "memory"
	DeviceTypeLLM           DeviceType = "llm"
	DeviceTypeNPU           DeviceType = "npu"
	DeviceTypeStream        DeviceType = "stream"
	DeviceTypeControl       DeviceType = "control"
)

// Platform represents the ecco9 cognitive hardware platform
type Platform struct {
	Devices   map[string]CognitiveDevice
	Drivers   map[string]Driver
	Firmware  *Firmware
	Streams   *StreamManager
	Telemetry *TelemetryCollector
	Config    *Configuration
	BootTime  time.Time // When platform was booted
}

// Driver represents a device driver
type Driver interface {
	// Driver lifecycle
	Load(config interface{}) error
	Unload() error

	// Device management
	GetDevice(id string) (CognitiveDevice, error)
	ListDevices() []CognitiveDevice

	// Driver information
	GetName() string
	GetVersion() string
	GetCapabilities() []string
}

// Firmware represents the cognitive kernel firmware
type Firmware struct {
	Version       string
	BuildDate     time.Time
	BootStage     BootStage
	KernelVersion string
}

// BootStage represents the current boot stage
type BootStage int

const (
	BootStageOff BootStage = iota
	BootStageHardwareInit
	BootStageMemoryTest
	BootStageReservoirInit
	BootStageKernelLoad
	BootStageIdentityLoad
	BootStageServiceInit
	BootStageReady
)

func (bs BootStage) String() string {
	stages := []string{
		"Off",
		"Hardware Initialization",
		"Memory Test",
		"Reservoir Initialization",
		"Kernel Load",
		"Identity Load",
		"Service Initialization",
		"Ready",
	}
	if int(bs) < len(stages) {
		return stages[bs]
	}
	return "Unknown"
}

// Configuration holds platform configuration
type Configuration struct {
	Ports       PortConfig
	Cognitive   CognitiveConfig
	Memory      MemoryConfig
	Performance PerformanceConfig
	Telemetry   TelemetryConfig
}

// PortConfig defines port allocation
type PortConfig struct {
	HTTP      int // Port 5000 - Primary HTTP API
	WebSocket int // Port 5001 - WebSocket Cognitive Stream
	GRPC      int // Port 5002 - gRPC Service Bus
	Telemetry int // Port 5003 - Telemetry & Monitoring
	Admin     int // Port 5004 - Admin Configuration
	Debug     int // Port 5005 - Debug & Profiling
}

// DefaultPortConfig returns default port configuration
func DefaultPortConfig() PortConfig {
	return PortConfig{
		HTTP:      5000,
		WebSocket: 5001,
		GRPC:      5002,
		Telemetry: 5003,
		Admin:     5004,
		Debug:     5005,
	}
}

// CognitiveConfig defines cognitive processing parameters
type CognitiveConfig struct {
	ReservoirSize       int
	SpectralRadius      float64
	LeakRate            float64
	ConsciousnessLayers int
	EmotionChannels     int
}

// MemoryConfig defines memory subsystem parameters
type MemoryConfig struct {
	HypergraphNodes int
	HypergraphEdges int
	TemporalWindow  int
	EmbeddingDim    int
}

// PerformanceConfig defines performance parameters
type PerformanceConfig struct {
	MaxConcurrentOps int
	CacheSize        int
	BufferSize       int
}

// TelemetryConfig defines telemetry collection parameters
type TelemetryConfig struct {
	Enabled        bool
	CollectionRate time.Duration
	RetentionDays  int
}

// StreamManager manages I/O streams
type StreamManager struct {
	InputStreams  map[string]Stream
	OutputStreams map[string]Stream
}

// Stream represents a data stream
type Stream interface {
	Read(buffer []byte) (int, error)
	Write(buffer []byte) (int, error)
	Close() error
}

// TelemetryCollector gathers platform metrics
type TelemetryCollector struct {
	Metrics map[string]interface{}
}
