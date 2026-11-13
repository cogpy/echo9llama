package hgql

import "github.com/EchoCog/echollama/core/deeptreeecho"

// NewHGQLParser returns a parser instance with initialized collections.
func NewHGQLParser() *HGQLParser {
	return &HGQLParser{
		Rules:           make(map[string]*ParseRule),
		HyperExtensions: make(map[string]*HyperExtension),
	}
}

// NewHyperGraphExecutor wires basic executor dependencies around the provided identity.
func NewHyperGraphExecutor(identity *deeptreeecho.Identity) *HyperGraphExecutor {
	return &HyperGraphExecutor{
		Identity:        identity,
		Resolvers:       make(map[string]*HyperResolver),
		TraversalEngine: &TraversalEngine{},
		PatternMatcher:  &PatternMatcher{},
	}
}

// NewQueryOptimizer constructs a no-op optimizer with empty analytics.
func NewQueryOptimizer() *QueryOptimizer {
	return &QueryOptimizer{
		OptimizationRules: []OptimizationRule{},
		CostModel:         &CostModel{},
		Statistics:        &QueryStatistics{},
	}
}

// NewPatternRecognition returns an initialized pattern recognition pipeline.
func NewPatternRecognition(identity *deeptreeecho.Identity) *PatternRecognition {
	return &PatternRecognition{
		Identity:          identity,
		PatternLibrary:    make(map[string]*CognitivePattern),
		MatchingAlgorithm: "default",
		Confidence:        0.0,
	}
}

// NewMultiScaleProcessor constructs a processor with default scale tracking.
func NewMultiScaleProcessor() *MultiScaleProcessor {
	return &MultiScaleProcessor{
		Scales:       []ProcessingScale{},
		Aggregators:  make(map[string]*ScaleAggregator),
		CurrentScale: 0,
	}
}

// NewAuthenticationManager returns an authentication manager with no providers configured.
func NewAuthenticationManager() *AuthenticationManager {
	return &AuthenticationManager{
		Providers: make(map[string]*AuthProvider),
		Sessions:  make(map[string]*AuthSession),
		Config:    &AuthManagerConfig{},
	}
}

// NewRateLimiter returns a limiter with empty counters and limits enabled.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		Limits:   make(map[string]*RateLimit),
		Counters: make(map[string]*RateCounter),
		Enabled:  true,
	}
}

// NewTransformationPipeline returns an empty transformation pipeline.
func NewTransformationPipeline() *TransformationPipeline {
	return &TransformationPipeline{
		Stages:     []TransformStage{},
		Config:     &PipelineConfig{},
		Metrics:    &PipelineMetrics{},
		Processors: make(map[string]*Processor),
	}
}

// NewConnectionMonitor returns a monitor with default collections.
func NewConnectionMonitor() *ConnectionMonitor {
	return &ConnectionMonitor{
		Connections: make(map[string]*ConnectionStatus),
		Alerts:      []*MonitoringAlert{},
		Thresholds:  &MonitoringThresholds{},
	}
}

// NewConnectionPool constructs a pool with empty connection buckets.
func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		Connections: make(map[string][]interface{}),
		MaxSize:     0,
		MinSize:     0,
	}
}
