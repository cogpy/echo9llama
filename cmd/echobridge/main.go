//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cogpy/echo9llama/core/echobridge"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	grpcPort = ":50051"
	httpPort = ":50052"
)

// SimpleScheduler is a basic implementation of the Scheduler interface
type SimpleScheduler struct {
	events chan *echobridge.CognitiveEvent
}

func NewSimpleScheduler() *SimpleScheduler {
	return &SimpleScheduler{
		events: make(chan *echobridge.CognitiveEvent, 1000),
	}
}

func (s *SimpleScheduler) ScheduleEvent(event *echobridge.CognitiveEvent) error {
	select {
	case s.events <- event:
		return nil
	default:
		return fmt.Errorf("scheduler queue full")
	}
}

func (s *SimpleScheduler) GetNextEvent() (*echobridge.CognitiveEvent, error) {
	select {
	case event := <-s.events:
		return event, nil
	case <-time.After(100 * time.Millisecond):
		return nil, fmt.Errorf("no events available")
	}
}

func (s *SimpleScheduler) CancelEvent(eventID string) error {
	// Simple implementation - would need more sophisticated logic for production
	return nil
}

func main() {
	log.Println("🌳 Starting EchoBridge gRPC Server...")
	
	// Create scheduler
	scheduler := NewSimpleScheduler()
	
	// Create gRPC server
	grpcServer := grpc.NewServer()
	bridgeServer := echobridge.NewEchoBridgeServer(scheduler)
	echobridge.RegisterEchoBridgeServer(grpcServer, bridgeServer)
	
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
