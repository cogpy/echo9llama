package goals

import (
	"context"
	"errors"
	"testing"

	"github.com/EchoCog/echollama/core/llm"
)

type fakeGoalGenerator struct {
	reply string
	err   error
	calls int
}

func (f *fakeGoalGenerator) Generate(context.Context, string, llm.GenerateOptions) (string, error) {
	f.calls++
	if f.err != nil {
		return "", f.err
	}
	return f.reply, nil
}

func newTestGoalOrchestrator() *GoalOrchestrator {
	return NewGoalOrchestrator(map[string]interface{}{"identity": "Deep Tree Echo"}, "")
}

func TestGenerateGoalsFromIdentityRequiresModel(t *testing.T) {
	g := newTestGoalOrchestrator()
	defer g.cancel()

	g.generateGoalsFromIdentity()
	if len(g.GetActiveGoals()) != 0 {
		t.Fatal("no model must mean no goals")
	}

	failing := &fakeGoalGenerator{err: errors.New("offline")}
	g.SetGoalGenerator(failing)
	g.generateGoalsFromIdentity()
	if failing.calls != 1 || len(g.GetActiveGoals()) != 0 {
		t.Fatalf("model failure must create no goal: calls=%d goals=%d", failing.calls, len(g.GetActiveGoals()))
	}

	g.SetGoalGenerator(&fakeGoalGenerator{reply: "I am not following the format."})
	g.generateGoalsFromIdentity()
	if len(g.GetActiveGoals()) != 0 {
		t.Fatal("an unparseable reply must create no goal")
	}
}

func TestGenerateGoalsFromIdentityUsesModelGoal(t *testing.T) {
	g := newTestGoalOrchestrator()
	defer g.cancel()
	g.SetGoalGenerator(&fakeGoalGenerator{reply: `TITLE: Map my memory topology
DESCRIPTION: I want to understand how my memories connect.
CATEGORY: knowledge_growth
PRIORITY: 12
CRITERIA: Chart 10 memory clusters; Describe 3 cross-links`})

	g.generateGoalsFromIdentity()
	goals := g.GetActiveGoals()
	if len(goals) != 1 {
		t.Fatalf("expected exactly one goal, got %d", len(goals))
	}
	goal := goals[0]
	if goal.Title != "Map my memory topology" || goal.Description != "I want to understand how my memories connect." {
		t.Fatalf("goal text was not the model's: %#v", goal)
	}
	if goal.Category != CategoryKnowledgeGrowth || goal.Priority != 10 {
		t.Fatalf("unexpected category/priority: %s %d", goal.Category, goal.Priority)
	}
	if len(goal.SuccessCriteria) != 2 || len(goal.Milestones) != 2 {
		t.Fatalf("expected two criteria and milestones, got %#v", goal.SuccessCriteria)
	}

	// The same goal is not duplicated.
	g.generateGoalsFromIdentity()
	if len(g.GetActiveGoals()) != 1 {
		t.Fatalf("duplicate goal was added: %d", len(g.GetActiveGoals()))
	}
}

func TestGoalOrchestratorStopDoesNotDeadlock(t *testing.T) {
	g := newTestGoalOrchestrator()
	g.mu.Lock()
	g.running = true
	g.mu.Unlock()
	if err := g.Stop(); err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}
	if err := g.Stop(); err == nil {
		t.Fatal("second Stop should report not running")
	}
}
