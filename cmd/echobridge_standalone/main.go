package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	pb "github.com/cogpy/echo9llama/core/echobridge"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	grpcPort = ":50051"
	httpPort = ":50052"
)

// SimpleScheduler is a basic implementation of the Scheduler interface
type SimpleScheduler struct {
	mu     sync.Mutex
	events []*pb.CognitiveEvent
}

func NewSimpleScheduler() *SimpleScheduler {
	return &SimpleScheduler{
		events: make([]*pb.CognitiveEvent, 0),
	}
}

func (s *SimpleScheduler) ScheduleEvent(event *pb.CognitiveEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
	return nil
}

func (s *SimpleScheduler) GetNextEvent() (*pb.CognitiveEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.events) == 0 {
		return nil, fmt.Errorf("no events available")
	}
	event := s.events[0]
	s.events = s.events[1:]
	return event, nil
}

func (s *SimpleScheduler) CancelEvent(eventID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, event := range s.events {
		if event.Id == eventID {
			s.events = append(s.events[:i], s.events[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("event not found")
}

// EchoBridgeServerWrapper wraps the generated server
type EchoBridgeServerWrapper struct {
	pb.UnimplementedEchoBridgeServer
	scheduler *SimpleScheduler
	mu        sync.RWMutex
	state     *pb.CognitiveState
	goals     map[string]*pb.Goal
}

func NewEchoBridgeServerWrapper(scheduler *SimpleScheduler) *EchoBridgeServerWrapper {
	now := time.Now().UnixMilli()
	return &EchoBridgeServerWrapper{
		scheduler: scheduler,
		state: &pb.CognitiveState{
			Energy:            1.0,
			Fatigue:           0.0,
			Coherence:         1.0,
			Curiosity:         0.7,
			CurrentState:      pb.State_STATE_INITIALIZING,
			LastRestTimestamp: now,
			CyclesSinceRest:   0,
			CurrentStep:       0,
		},
		goals: make(map[string]*pb.Goal),
	}
}

func (s *EchoBridgeServerWrapper) ScheduleEvent(ctx context.Context, event *pb.CognitiveEvent) (*pb.EventResponse, error) {
	err := s.scheduler.ScheduleEvent(event)
	if err != nil {
		return &pb.EventResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}
	return &pb.EventResponse{
		Success: true,
		Message: "Event scheduled successfully",
		EventId: event.Id,
	}, nil
}

func (s *EchoBridgeServerWrapper) GetState(ctx context.Context, req *pb.StateRequest) (*pb.CognitiveState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state, nil
}

func (s *EchoBridgeServerWrapper) UpdateState(ctx context.Context, state *pb.CognitiveState) (*pb.StateResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = state
	return &pb.StateResponse{
		Success: true,
		Message: "State updated successfully",
	}, nil
}

func (s *EchoBridgeServerWrapper) RegisterGoal(ctx context.Context, goal *pb.Goal) (*pb.GoalResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.goals[goal.Id] = goal
	return &pb.GoalResponse{
		Success: true,
		Message: "Goal registered successfully",
		GoalId:  goal.Id,
	}, nil
}

func (s *EchoBridgeServerWrapper) GetActiveGoals(ctx context.Context, req *pb.GoalRequest) (*pb.GoalList, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	goals := make([]*pb.Goal, 0)
	for _, goal := range s.goals {
		if req.StatusFilter == pb.GoalStatus_GOAL_UNKNOWN || goal.Status == req.StatusFilter {
			goals = append(goals, goal)
		}
	}

	return &pb.GoalList{Goals: goals}, nil
}

func (s *EchoBridgeServerWrapper) UpdateGoalProgress(ctx context.Context, progress *pb.GoalProgress) (*pb.GoalResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	goal, exists := s.goals[progress.GoalId]
	if !exists {
		return &pb.GoalResponse{
			Success: false,
			Message: "Goal not found",
		}, fmt.Errorf("goal not found")
	}

	goal.Progress = progress.Progress
	return &pb.GoalResponse{
		Success: true,
		Message: "Goal progress updated",
		GoalId:  progress.GoalId,
	}, nil
}

func (s *EchoBridgeServerWrapper) GetMetrics() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"energy":            s.state.Energy,
		"fatigue":           s.state.Fatigue,
		"coherence":         s.state.Coherence,
		"curiosity":         s.state.Curiosity,
		"current_state":     s.state.CurrentState.String(),
		"current_step":      s.state.CurrentStep,
		"cycles_since_rest": s.state.CyclesSinceRest,
		"active_goals":      len(s.goals),
	}
}

func main() {
	log.Println("🌳 Starting EchoBridge gRPC Server...")

	// Create scheduler
	scheduler := NewSimpleScheduler()

	// Create gRPC server
	grpcServer := grpc.NewServer()
	bridgeServer := NewEchoBridgeServerWrapper(scheduler)
	pb.RegisterEchoBridgeServer(grpcServer, bridgeServer)

	// Enable reflection for debugging
	reflection.Register(grpcServer)

	// Start gRPC server
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", grpcPort, err)
	}

	go func() {
		log.Printf("✅ gRPC server listening on %s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Start HTTP status server
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK\n")
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics := bridgeServer.GetMetrics()
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\n")
		first := true
		for k, v := range metrics {
			if !first {
				fmt.Fprintf(w, ",\n")
			}
			fmt.Fprintf(w, "  \"%s\": ", k)
			switch v := v.(type) {
			case string:
				fmt.Fprintf(w, "\"%s\"", v)
			case int64:
				fmt.Fprintf(w, "%d", v)
			case float64:
				fmt.Fprintf(w, "%.2f", v)
			case int32:
				fmt.Fprintf(w, "%d", v)
			case int:
				fmt.Fprintf(w, "%d", v)
			default:
				fmt.Fprintf(w, "\"%v\"", v)
			}
			first = false
		}
		fmt.Fprintf(w, "\n}\n")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>EchoBridge Status</title>
    <style>
        body { font-family: monospace; background: #1a1a1a; color: #00ff00; padding: 20px; }
        h1 { color: #00ff00; }
        .status { background: #2a2a2a; padding: 15px; border-radius: 5px; margin: 10px 0; }
        .metric { margin: 5px 0; }
        .label { color: #888; }
        .value { color: #00ff00; font-weight: bold; }
    </style>
</head>
<body>
    <h1>🌳 EchoBridge gRPC Server</h1>
    <div class="status">
        <div class="metric"><span class="label">Status:</span> <span class="value">Running</span></div>
        <div class="metric"><span class="label">gRPC Port:</span> <span class="value">%s</span></div>
        <div class="metric"><span class="label">HTTP Port:</span> <span class="value">%s</span></div>
    </div>
    <div class="status">
        <h2>Endpoints</h2>
        <div class="metric">• <a href="/health" style="color: #00ff00;">/health</a> - Health check</div>
        <div class="metric">• <a href="/metrics" style="color: #00ff00;">/metrics</a> - Server metrics (JSON)</div>
    </div>
</body>
</html>`, grpcPort, httpPort)
	})

	go func() {
		log.Printf("✅ HTTP status server listening on %s", httpPort)
		if err := http.ListenAndServe(httpPort, nil); err != nil {
			log.Fatalf("Failed to serve HTTP: %v", err)
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	log.Println("🛑 Shutting down EchoBridge server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grpcServer.GracefulStop()

	<-ctx.Done()
	log.Println("✅ EchoBridge server stopped")
}
