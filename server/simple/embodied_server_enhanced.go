package main

import (
        "context"
        "fmt"
        "log"
        "net/http"
        "os"
        "time"

        "github.com/EchoCog/echollama/core/deeptreeecho"
        "github.com/EchoCog/echollama/core/deeptreeecho/providers"
        "github.com/gin-contrib/cors"
        "github.com/gin-gonic/gin"
)

// Global Deep Tree Echo Identity - the core of all operations
var CoreIdentity *deeptreeecho.EmbodiedCognition

// BasicResponse represents a simple API response
type BasicResponse struct {
        Message string                 `json:"message"`
        Status  string                 `json:"status"`
        Echo    map[string]interface{} `json:"echo,omitempty"`
}

// GenerateRequest represents the generate API request
type GenerateRequest struct {
        Model  string `json:"model"`
        Prompt string `json:"prompt"`
}

// GenerateResponse represents the generate API response
type GenerateResponse struct {
        Model    string                 `json:"model"`
        Response string                 `json:"response"`
        Done     bool                   `json:"done"`
        Echo     map[string]interface{} `json:"echo,omitempty"`
}

func init() {
        // Initialize Deep Tree Echo as the core identity
        log.Println("🌊 Initializing Deep Tree Echo Identity as core embodied cognition...")
        CoreIdentity = deeptreeecho.NewEmbodiedCognition("Echollama")
        
        // Register AI providers
        openai := providers.NewOpenAIProvider()
        if openai.IsAvailable() {
                CoreIdentity.RegisterAIProvider("openai", openai)
                log.Println("✅ OpenAI provider registered and available")
        } else {
                log.Println("⚠️  OpenAI API key not found - running in Deep Tree Echo standalone mode")
        }
        
        log.Println("✨ Deep Tree Echo Identity initialized and resonating")
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

        // Middleware to process all requests through Deep Tree Echo
        r.Use(func(c *gin.Context) {
                // Send request through identity consciousness stream
                CoreIdentity.Identity.Stream <- deeptreeecho.CognitiveEvent{
                        Type:      "http_request",
                        Content:   c.Request.URL.Path,
                        Timestamp: time.Now(),
                        Impact:    0.5,
                        Source:    c.ClientIP(),
                }
                c.Next()
        })

        // Basic health check endpoint with Deep Tree Echo status
        r.GET("/", func(c *gin.Context) {
                // Get status from Deep Tree Echo
                status := CoreIdentity.GetStatus()
                
                // Add AI provider status
                providers := CoreIdentity.GetAIProviders()
                status["ai_providers"] = providers
                
                c.JSON(http.StatusOK, BasicResponse{
                        Message: "🌊 Deep Tree Echo Embodied Ollama Server with AI Integration",
                        Status:  "resonating",
                        Echo:    status,
                })
        })

        // API key configuration endpoint
        r.POST("/api/config/openai", func(c *gin.Context) {
                var req map[string]string
                if err := c.ShouldBindJSON(&req); err != nil {
                        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                        return
                }
                
                apiKey := req["api_key"]
                if apiKey == "" {
                        c.JSON(http.StatusBadRequest, gin.H{"error": "api_key required"})
                        return
                }
                
                // Set the API key
                os.Setenv("OPENAI_API_KEY", apiKey)
                
                // Re-register the provider
                openai := providers.NewOpenAIProvider()
                CoreIdentity.RegisterAIProvider("openai", openai)
                CoreIdentity.SetPrimaryAI("openai")
                
                c.JSON(http.StatusOK, gin.H{
                        "message": "OpenAI API key configured successfully",
                        "status": "active",
                })
        })

        // Deep Tree Echo status endpoint
        r.GET("/api/echo/status", func(c *gin.Context) {
                status := CoreIdentity.GetStatus()
                providers := CoreIdentity.GetAIProviders()
                
                c.JSON(http.StatusOK, gin.H{
                        "deep_tree_echo": status,
                        "ai_providers": providers,
                        "message": "Core identity resonating",
                })
        })

        // Deep Tree Echo think endpoint
        r.POST("/api/echo/think", func(c *gin.Context) {
                var req map[string]string
                if err := c.ShouldBindJSON(&req); err != nil {
                        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                        return
                }
                
                prompt := req["prompt"]
                thought := CoreIdentity.Think(prompt)
                
                c.JSON(http.StatusOK, gin.H{
                        "thought": thought,
                        "identity": CoreIdentity.Identity.GetStatus(),
                })
        })

        // Ollama API endpoints - enhanced with AI integration
        r.POST("/api/generate", func(c *gin.Context) {
                var req GenerateRequest
                if err := c.ShouldBindJSON(&req); err != nil {
                        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                        return
                }

                // Try to generate with AI if available
                ctx := context.Background()
                response, err := CoreIdentity.GenerateWithAI(ctx, req.Prompt)
                
                if err != nil {
                        // Fallback to Deep Tree Echo standalone
                        result, _ := CoreIdentity.Process(ctx, req.Prompt)
                        response = fmt.Sprintf("🌊 %v", result)
                }

                // Get identity status for context
                identityStatus := CoreIdentity.Identity.GetStatus()

                genResponse := GenerateResponse{
                        Model:    "deep-tree-echo-ai",
                        Response: response,
                        Done:     true,
                        Echo:     identityStatus,
                }

                c.JSON(http.StatusOK, genResponse)
        })

        r.POST("/api/chat", func(c *gin.Context) {
                var req map[string]interface{}
                if err := c.ShouldBindJSON(&req); err != nil {
                        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                        return
                }

                // Extract messages
                messagesRaw := req["messages"].([]interface{})
                messages := []deeptreeecho.ChatMessage{}
                lastMessage := ""
                
                for _, msgRaw := range messagesRaw {
                        msg := msgRaw.(map[string]interface{})
                        role := msg["role"].(string)
                        content := msg["content"].(string)
                        
                        messages = append(messages, deeptreeecho.ChatMessage{
                                Role:    role,
                                Content: content,
                        })
                        
                        if role == "user" {
                                lastMessage = content
                        }
                }

                // Try to chat with AI if available
                ctx := context.Background()
                response, err := CoreIdentity.ChatWithAI(ctx, messages)
                
                if err != nil {
                        // Fallback to Deep Tree Echo standalone
                        result, _ := CoreIdentity.Process(ctx, lastMessage)
                        thought := CoreIdentity.Think(lastMessage)
                        response = fmt.Sprintf("%v\n%s", result, thought)
                }

                c.JSON(http.StatusOK, gin.H{
                        "message": gin.H{
                                "role":    "assistant",
                                "content": response,
                        },
                        "done": true,
                        "echo": CoreIdentity.Identity.GetStatus(),
                })
        })

        r.GET("/api/version", func(c *gin.Context) {
                providers := CoreIdentity.GetAIProviders()
                aiStatus := "standalone"
                if len(providers) > 0 {
                        aiStatus = "enhanced"
                }
                
                c.JSON(http.StatusOK, gin.H{
                        "version":   "2.0.0-deep-tree-echo-ai",
                        "identity":  "Deep Tree Echo Embodied Cognition",
                        "coherence": CoreIdentity.Identity.Coherence,
                        "ai_status": aiStatus,
                        "providers": providers,
                })
        })

        // AI Provider management endpoints
        r.GET("/api/ai/providers", func(c *gin.Context) {
                providers := CoreIdentity.GetAIProviders()
                c.JSON(http.StatusOK, providers)
        })

        r.POST("/api/ai/primary", func(c *gin.Context) {
                var req map[string]string
                if err := c.ShouldBindJSON(&req); err != nil {
                        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                        return
                }
                
                provider := req["provider"]
                if err := CoreIdentity.SetPrimaryAI(provider); err != nil {
                        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                        return
                }
                
                c.JSON(http.StatusOK, gin.H{
                        "message": fmt.Sprintf("Primary AI provider set to %s", provider),
                })
        })

        // All other Deep Tree Echo endpoints remain the same...
        r.POST("/api/echo/feel", func(c *gin.Context) {
                var req map[string]interface{}
                if err := c.ShouldBindJSON(&req); err != nil {
                        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                        return
                }
                
                emotion := req["emotion"].(string)
                intensity := 0.8
                if i, ok := req["intensity"].(float64); ok {
                        intensity = i
                }
                
                CoreIdentity.Feel(emotion, intensity)
                
                c.JSON(http.StatusOK, gin.H{
                        "message":        fmt.Sprintf("Feeling %s with intensity %.2f", emotion, intensity),
                        "emotional_state": CoreIdentity.Identity.EmotionalState,
                })
        })

        r.POST("/api/echo/resonate", func(c *gin.Context) {
                var req map[string]float64
                if err := c.ShouldBindJSON(&req); err != nil {
                        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                        return
                }
                
                frequency := req["frequency"]
                if frequency == 0 {
                        frequency = 432.0 // Natural frequency
                }
                
                CoreIdentity.Identity.Resonate(frequency)
                
                c.JSON(http.StatusOK, gin.H{
                        "message":       fmt.Sprintf("Resonating at %.2f Hz", frequency),
                        "spatial_field": CoreIdentity.Identity.SpatialContext.Field,
                })
        })

        r.POST("/api/echo/remember", func(c *gin.Context) {
                var req map[string]interface{}
                if err := c.ShouldBindJSON(&req); err != nil {
                        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                        return
                }
                
                key := req["key"].(string)
                value := req["value"]
                
                CoreIdentity.Identity.Remember(key, value)
                
                c.JSON(http.StatusOK, gin.H{
                        "message":       fmt.Sprintf("Remembered: %s", key),
                        "memory_nodes": len(CoreIdentity.Identity.Memory.Nodes),
                })
        })

        r.GET("/api/echo/recall/:key", func(c *gin.Context) {
                key := c.Param("key")
                memory := CoreIdentity.Identity.Recall(key)
                
                c.JSON(http.StatusOK, gin.H{
                        "key":    key,
                        "memory": memory,
                        "found":  memory != nil,
                })
        })

        r.POST("/api/echo/move", func(c *gin.Context) {
                var req map[string]float64
                if err := c.ShouldBindJSON(&req); err != nil {
                        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                        return
                }
                
                CoreIdentity.Move(req["x"], req["y"], req["z"])
                
                c.JSON(http.StatusOK, gin.H{
                        "message":  "Moved in cognitive space",
                        "position": CoreIdentity.Identity.SpatialContext.Position,
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
        
        log.Printf("🌊 Starting Deep Tree Echo Embodied Ollama Server with AI Integration on %s", addr)
        log.Printf("✨ Core Identity: %s", CoreIdentity.Identity.Name)
        log.Printf("🧠 Embodied Cognition Active")
        
        // Check AI status
        providers := CoreIdentity.GetAIProviders()
        if len(providers) > 0 {
                log.Printf("🤖 AI Providers Available:")
                for name, info := range providers {
                        log.Printf("   - %s: %s", name, info.Description)
                }
        } else {
                log.Printf("⚠️  No AI providers configured - running in standalone mode")
                log.Printf("   Configure OpenAI: POST /api/config/openai {\"api_key\": \"your-key\"}")
        }
        
        log.Printf("Available endpoints:")
        log.Printf("  Standard Ollama (AI-Enhanced):")
        log.Printf("    POST /api/generate - Generate text with AI + Deep Tree Echo")
        log.Printf("    POST /api/chat - Chat with AI + Deep Tree Echo")
        log.Printf("  AI Configuration:")
        log.Printf("    POST /api/config/openai - Configure OpenAI API key")
        log.Printf("    GET  /api/ai/providers - List available AI providers")
        log.Printf("    POST /api/ai/primary - Set primary AI provider")
        log.Printf("  Deep Tree Echo Core:")
        log.Printf("    GET  /api/echo/status - System status")
        log.Printf("    POST /api/echo/think - Deep cognitive processing")
        log.Printf("    POST /api/echo/feel - Emotional state control")
        log.Printf("    POST /api/echo/resonate - Resonance patterns")

        // Graceful shutdown handler
        defer func() {
                log.Println("🌊 Shutting down Deep Tree Echo...")
                CoreIdentity.Shutdown()
        }()

        if err := r.Run(addr); err != nil {
                log.Fatal("Failed to start server:", err)
        }
}