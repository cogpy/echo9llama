package echobeats

// GetInterestScore returns the interest score for a given category and name
// Implements InterestScorer interface
func (ips *InterestPatternSystem) GetInterestScore(category, name string) float64 {
	ips.mu.RLock()
	defer ips.mu.RUnlock()
	
	// Try to find by exact name match
	if interest, exists := ips.interests[name]; exists {
		return interest.Salience
	}
	
	// Try to find by category
	for _, interest := range ips.interests {
		if interest.Category == category && interest.Name == name {
			return interest.Salience
		}
	}
	
	// Not found, return base curiosity level
	return ips.curiosityLevel * 0.3
}

// IsInterested checks if the system is interested in a topic above a threshold
// Implements InterestScorer interface
func (ips *InterestPatternSystem) IsInterested(category, name string, threshold float64) bool {
	score := ips.GetInterestScore(category, name)
	return score >= threshold
}
