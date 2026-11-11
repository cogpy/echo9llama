package consciousness

import (
	"fmt"
	"time"
)

// AddExternalThought adds a thought from an external source
func (soc *StreamOfConsciousnessLLM) AddExternalThought(content string) {
	thought := &ThoughtLLM{
		ID:            fmt.Sprintf("external-%d", time.Now().UnixNano()),
		Timestamp:     time.Now(),
		Type:          ThoughtTypePerception,
		Content:       content,
		Source:        "external",
		Confidence:    0.9,
		EmotionalTone: soc.copyEmotionalState(),
		Context:       map[string]interface{}{"external": true},
		RelatedTo:     []string{},
		GeneratedBy:   "external",
	}

	soc.mu.Lock()
	soc.currentThought = thought
	soc.thoughtHistory = append(soc.thoughtHistory, thought)

	// Trim history if needed
	if len(soc.thoughtHistory) > soc.maxHistorySize {
		soc.thoughtHistory = soc.thoughtHistory[len(soc.thoughtHistory)-soc.maxHistorySize:]
	}

	soc.thoughtsGenerated++
	soc.mu.Unlock()

	// Add to recent experiences
	soc.AddExperience(content)
}
