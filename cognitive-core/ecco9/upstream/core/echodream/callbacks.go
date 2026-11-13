package echodream

// SetWisdomCallback sets the callback for when wisdom is extracted
func (dci *DreamCycleIntegration) SetWisdomCallback(callback func(wisdom Wisdom)) {
	dci.mu.Lock()
	defer dci.mu.Unlock()
	dci.onWisdomExtracted = callback
}

// SetDreamCompleteCallback sets the callback for when a dream completes
func (dci *DreamCycleIntegration) SetDreamCompleteCallback(callback func(dream *Dream)) {
	dci.mu.Lock()
	defer dci.mu.Unlock()
	dci.onDreamComplete = callback
}
