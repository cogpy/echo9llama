package main

import (
        "fmt"
        "log"
        "net/http"
        "os"

        "github.com/gin-contrib/cors"
        "github.com/gin-gonic/gin"
)

// BasicResponse represents a simple API response
type BasicResponse struct {
        Message string `json:"message"`
        Status  string `json:"status"`
}

// GenerateRequest represents the generate API request
type GenerateRequest struct {
        Model  string `json:"model"`
        Prompt string `json:"prompt"`
}

// GenerateResponse represents the generate API response
type GenerateResponse struct {
        Model    string `json:"model"`
        Response string `json:"response"`
        Done     bool   `json:"done"`
}

func main() {
        // Set Gin mode
        gin.SetMode(gin.ReleaseMode)
        
        // Create Gin router
        r := gin.Default()

        // Configure CORS to allow all origins (required for Replit)
        config := cors.DefaultConfig()
        config.AllowAllOrigins = true
        config.AllowHeaders = []string{"*"}
        config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
        r.Use(cors.New(config))

        // Basic health check endpoint
        r.GET("/", func(c *gin.Context) {
                c.JSON(http.StatusOK, BasicResponse{
                        Message: "Ollama-compatible server is running",
                        Status:  "ready",
                })
        })

        // Ollama API endpoints
        r.GET("/api/tags", func(c *gin.Context) {
                c.JSON(http.StatusOK, gin.H{
                        "models": []gin.H{},
                })
        })

        r.POST("/api/generate", func(c *gin.Context) {
                var req GenerateRequest
                if err := c.ShouldBindJSON(&req); err != nil {
                        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                        return
                }

                // Simple echo response for now
                response := GenerateResponse{
                        Model:    req.Model,
                        Response: fmt.Sprintf("Echo: %s", req.Prompt),
                        Done:     true,
                }

                c.JSON(http.StatusOK, response)
        })

        r.POST("/api/chat", func(c *gin.Context) {
                var req map[string]interface{}
                if err := c.ShouldBindJSON(&req); err != nil {
                        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                        return
                }

                c.JSON(http.StatusOK, gin.H{
                        "message": gin.H{
                                "role":    "assistant",
                                "content": "This is a basic Ollama-compatible server response.",
                        },
                        "done": true,
                })
        })

        r.GET("/api/version", func(c *gin.Context) {
                c.JSON(http.StatusOK, gin.H{
                        "version": "0.1.0-simple",
                })
        })

        // Get port from environment or default to 5000
        port := os.Getenv("PORT")
        if port == "" {
                port = "5000"
        }

        // Get host - use 0.0.0.0 for Replit
        host := "0.0.0.0"
        if envHost := os.Getenv("HOST"); envHost != "" {
                host = envHost
        }

        addr := fmt.Sprintf("%s:%s", host, port)
        
        log.Printf("Starting simple Ollama-compatible server on %s", addr)
        log.Printf("Available endpoints:")
        log.Printf("  GET  / - Health check")
        log.Printf("  GET  /api/tags - List models")
        log.Printf("  POST /api/generate - Generate text")
        log.Printf("  POST /api/chat - Chat completion")
        log.Printf("  GET  /api/version - Version info")

        if err := r.Run(addr); err != nil {
                log.Fatal("Failed to start server:", err)
        }
}