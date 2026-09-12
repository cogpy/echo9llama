package echobeats

// GetQueueSize returns the current event queue size
func (eb *EchoBeats) GetQueueSize() int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return eb.eventQueue.Len()
}

// GetState returns the current scheduler state
func (eb *EchoBeats) GetState() SchedulerState {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return eb.state
}

// GetMetrics returns scheduler metrics
func (eb *EchoBeats) GetMetrics() map[string]interface{} {
	eb.metrics.mu.RLock()
	defer eb.metrics.mu.RUnlock()

	return map[string]interface{}{
		"events_processed":    eb.metrics.EventsProcessed,
		"events_scheduled":    eb.metrics.EventsScheduled,
		"average_latency":     eb.metrics.AverageLatency.String(),
		"cycles_completed":    eb.metrics.CyclesCompleted,
		"current_load":        eb.metrics.CurrentLoad,
		"autonomous_thoughts": eb.metrics.AutonomousThoughts,
		"last_heartbeat":      eb.metrics.LastHeartbeat.Format("15:04:05"),
	}
}

// SetState sets the scheduler state (for orchestrator control)
func (eb *EchoBeats) SetState(state SchedulerState) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.state = state
}
