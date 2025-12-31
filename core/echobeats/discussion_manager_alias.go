package echobeats

// DiscussionManager is an alias for AutonomousDiscussionManager
// This provides backward compatibility with code expecting DiscussionManager type
type DiscussionManager = AutonomousDiscussionManager

// NewDiscussionManager creates a new DiscussionManager (AutonomousDiscussionManager)
func NewDiscussionManager(interestScorer InterestScorer) *DiscussionManager {
	return NewAutonomousDiscussionManager(interestScorer)
}
