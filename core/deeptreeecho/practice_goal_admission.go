package deeptreeecho

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// AdmitPracticeGoal admits a caller-journaled task with a stable ID. It does not
// perform inference, grant tool permissions, or invent completed goal evidence.
func (sched *EchobeatsScheduler) AdmitPracticeGoal(id, description string, priority float64, created time.Time) error {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(description) == "" || math.IsNaN(priority) || math.IsInf(priority, 0) || priority < 0 || priority > 1 || created.IsZero() {
		return fmt.Errorf("invalid practice goal")
	}
	sched.mu.Lock()
	defer sched.mu.Unlock()
	for _, goal := range sched.goalQueue {
		if goal.ID == id {
			if goal.Description != description || goal.Priority != priority {
				return fmt.Errorf("practice goal conflict: %s", id)
			}
			return nil
		}
	}
	sched.goalQueue = append(sched.goalQueue, ScheduledGoal{ID: id, Description: description, Priority: priority, Status: GoalPending, CreatedAt: created, UpdatedAt: created})
	sched.priorityMatrix[id] = priority
	sched.sortGoalsByPriority()
	return nil
}

// HasEligibleOrdinaryGoal mirrors the scheduler's pending/active selection rule.
// A retired practice trial must yield one durable ordinary context turn whenever
// ordinary work exists, including when the trial ended in a provider outage.
func (sched *EchobeatsScheduler) HasEligibleOrdinaryGoal() bool {
	sched.mu.RLock()
	defer sched.mu.RUnlock()
	for _, goal := range sched.goalQueue {
		if !strings.HasPrefix(goal.ID, "practice-goal:") && (goal.Status == GoalPending || goal.Status == GoalActive) {
			return true
		}
	}
	return false
}

// FinishPracticeGoal separates trial retirement from arithmetic achievement.
// Only an independently verified pass earns completed progress. Failure/unknown
// preserves progress and retires an existing task without synthesizing a goal.
func (sched *EchobeatsScheduler) FinishPracticeGoal(goalID, evidenceID string, passed bool) bool {
	if !strings.HasPrefix(goalID, "practice-goal:") || strings.TrimSpace(evidenceID) == "" {
		return false
	}
	sched.mu.Lock()
	defer sched.mu.Unlock()
	if sched.appliedGoalEvidence == nil {
		sched.appliedGoalEvidence = make(map[string]struct{})
	}
	if _, exists := sched.appliedGoalEvidence[evidenceID]; exists {
		return false
	}
	for i := range sched.goalQueue {
		goal := &sched.goalQueue[i]
		if goal.ID != goalID {
			continue
		}
		if passed {
			goal.Progress = 1
			if goal.Status != GoalCompleted {
				goal.Status = GoalCompleted
				sched.goalsCompleted++
			}
		} else {
			goal.Status = GoalAbandoned
		}
		goal.UpdatedAt = time.Now()
		sched.appliedGoalEvidence[evidenceID] = struct{}{}
		return true
	}
	return false
}
