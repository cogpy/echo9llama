package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Simple Go proxy for EchoLlama development mode
// Listens on 0.0.0.0:11434 and forwards all requests to 127.0.0.1:11435

func main() {
	var (
		listenAddr = flag.String("listen", "0.0.0.0:11434", "Address to listen on (default: 0.0.0.0:11434)")
		targetAddr = flag.String("target", "http://127.0.0.1:11435", "Target echollama server URL (default: http://127.0.0.1:11435)")
	)
	flag.Parse()

	// Parse target URL
	target, err := url.Parse(*targetAddr)
	if err != nil {
		log.Fatalf("Invalid target URL: %v", err)
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Customize proxy to add logging and error handling
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		log.Printf("Proxying %s %s → %s", req.Method, req.URL.Path, target)
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error for %s %s: %v", r.Method, r.URL.Path, err)
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, "Proxy error: %v", err)
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         *listenAddr,
		Handler:      proxy,
		ReadTimeout:  300 * time.Second, // Long timeout for model operations
		WriteTimeout: 300 * time.Second,
	}

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down proxy...")
		server.Close()
	}()

	log.Printf("Starting universal proxy: %s → %s", *listenAddr, *targetAddr)
	log.Printf("All traffic to any interface on port 11434 will be forwarded to echollama")
	log.Printf("Test with: curl http://localhost:11434/api/version")

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}

	log.Println("Proxy stopped")
}