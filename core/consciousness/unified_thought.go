package consciousness

import (
	"fmt"
	"time"
)

// UnifiedThought represents a thought that can be used across all consciousness systems
// This unifies the previous Thought and LLMThought types
type UnifiedThought struct {
	ID        string
	Content   string
	Type      ThoughtType
	Timestamp time.Time

	// Relevance and salience
	Relevance float64
	Depth     float64

	// Emotional aspects
	EmotionalTone string
	Emotion       string

	// Relationships
	TriggeredBy string
	LeadsTo     []string
	Tags        []string
}

// NewUnifiedThought creates a new unified thought
func NewUnifiedThought(content string, thoughtType ThoughtType) *UnifiedThought {
	return &UnifiedThought{
		ID:        generateThoughtID(),
		Content:   content,
		Type:      thoughtType,
		Timestamp: time.Now(),
		Relevance: 0.5,
		Depth:     0.5,
		LeadsTo:   make([]string, 0),
		Tags:      make([]string, 0),
	}
}

// WithRelevance sets the relevance score
func (t *UnifiedThought) WithRelevance(relevance float64) *UnifiedThought {
	t.Relevance = relevance
	return t
}

// WithDepth sets the depth score
func (t *UnifiedThought) WithDepth(depth float64) *UnifiedThought {
	t.Depth = depth
	return t
}

// WithEmotion sets the emotional tone
func (t *UnifiedThought) WithEmotion(emotion string) *UnifiedThought {
	t.EmotionalTone = emotion
	t.Emotion = emotion
	return t
}

// WithTags sets the tags
func (t *UnifiedThought) WithTags(tags []string) *UnifiedThought {
	t.Tags = tags
	return t
}

// WithTriggeredBy sets what triggered this thought
func (t *UnifiedThought) WithTriggeredBy(trigger string) *UnifiedThought {
	t.TriggeredBy = trigger
	return t
}

// AddLeadsTo adds a thought that this one leads to
func (t *UnifiedThought) AddLeadsTo(thoughtID string) {
	t.LeadsTo = append(t.LeadsTo, thoughtID)
}

// AddTag adds a tag to the thought
func (t *UnifiedThought) AddTag(tag string) {
	for _, existing := range t.Tags {
		if existing == tag {
			return
		}
	}
	t.Tags = append(t.Tags, tag)
}

// ToLegacyThought converts to the old Thought type for backward compatibility
func (t *UnifiedThought) ToLegacyThought() *Thought {
	return &Thought{
		ID:            t.ID,
		Content:       t.Content,
		Type:          t.Type,
		Timestamp:     t.Timestamp,
		Relevance:     t.Relevance,
		EmotionalTone: t.EmotionalTone,
		TriggeredBy:   t.TriggeredBy,
		LeadsTo:       t.LeadsTo,
	}
}

// ToLegacyLLMThought converts to the old LLMThought type for backward compatibility
func (t *UnifiedThought) ToLegacyLLMThought() *LLMThought {
	return &LLMThought{
		ID:        t.ID,
		Type:      t.Type,
		Content:   t.Content,
		Timestamp: t.Timestamp,
		Emotion:   t.Emotion,
		Depth:     t.Depth,
		Tags:      t.Tags,
	}
}

// FromLegacyThought creates a UnifiedThought from old Thought type
func FromLegacyThought(t *Thought) *UnifiedThought {
	return &UnifiedThought{
		ID:            t.ID,
		Content:       t.Content,
		Type:          t.Type,
		Timestamp:     t.Timestamp,
		Relevance:     t.Relevance,
		EmotionalTone: t.EmotionalTone,
		Emotion:       t.EmotionalTone,
		TriggeredBy:   t.TriggeredBy,
		LeadsTo:       t.LeadsTo,
		Tags:          make([]string, 0),
		Depth:         t.Relevance,
	}
}

// FromLegacyLLMThought creates a UnifiedThought from old LLMThought type
func FromLegacyLLMThought(t *LLMThought) *UnifiedThought {
	return &UnifiedThought{
		ID:            t.ID,
		Content:       t.Content,
		Type:          t.Type,
		Timestamp:     t.Timestamp,
		Relevance:     t.Depth,
		Depth:         t.Depth,
		EmotionalTone: t.Emotion,
		Emotion:       t.Emotion,
		Tags:          t.Tags,
		LeadsTo:       make([]string, 0),
	}
}

func generateThoughtID() string {
	return fmt.Sprintf("thought_%d", time.Now().UnixNano())
}
