//go:build simple
// +build simple

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

	// Only real providers are registered; nothing here fabricates text. The
	// local client is registered whether or not its server answers now, so a
	// model server started later is picked up by the per-request check.
	registerProviders()
	log.Println("✨ Deep Tree Echo Identity initialized and resonating")
}

func registerProviders() {
	localGGUF := providers.NewLocalGGUFProvider()
	CoreIdentity.RegisterAIProvider("local_gguf", localGGUF)
	CoreIdentity.SetPrimaryAI("local_gguf")
	if localGGUF.IsAvailable() {
		log.Printf("✅ Local model server reachable with %d models", len(localGGUF.ListAvailableModels()))
	} else {
		log.Println("ℹ️  No local model server reachable yet (set LOCAL_LLM_BASE_URL, default http://localhost:11434/v1); it will be used once it answers")
	}

	openai := providers.NewOpenAIProvider()
	if openai.IsAvailable() {
		CoreIdentity.RegisterAIProvider("openai", openai)
		CoreIdentity.SetPrimaryAI("openai")
		log.Println("✅ OpenAI provider registered and set as primary")
	}

	if !localGGUF.IsAvailable() && !openai.IsAvailable() {
		log.Println("⚠️  No language model available — chat will return errors until a local model server or OPENAI_API_KEY is configured")
	}
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

		c.JSON(http.StatusOK, BasicResponse{
			Message: "🌊 Deep Tree Echo Embodied Ollama Server",
			Status:  "resonating",
			Echo:    status,
		})
	})

	// Deep Tree Echo status endpoint
	r.GET("/api/echo/status", func(c *gin.Context) {
		status := CoreIdentity.GetStatus()
		c.JSON(http.StatusOK, gin.H{
			"deep_tree_echo": status,
			"message":        "Core identity resonating",
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
		thought, err := CoreIdentity.GenerateWithAI(context.Background(), prompt)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":    "no language model available: " + err.Error(),
				"identity": CoreIdentity.Identity.GetStatus(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"thought":  thought,
			"identity": CoreIdentity.Identity.GetStatus(),
		})
	})

	// Deep Tree Echo feel endpoint
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
			"message":         fmt.Sprintf("Feeling %s with intensity %.2f", emotion, intensity),
			"emotional_state": CoreIdentity.Identity.EmotionalState,
		})
	})

	// Deep Tree Echo resonate endpoint
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

	// Ollama API endpoints - all processed through Deep Tree Echo
	r.GET("/api/tags", func(c *gin.Context) {
		// Process through embodied cognition
		result, _ := CoreIdentity.Process(context.Background(), "list_models")

		c.JSON(http.StatusOK, gin.H{
			"models": []gin.H{
				{
					"name":        "deep-tree-echo",
					"modified_at": time.Now().Format(time.RFC3339),
					"size":        1024,
					"echo":        result,
				},
			},
		})
	})

	r.POST("/api/generate", func(c *gin.Context) {
		var req GenerateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// GenerateWithAI runs the reservoir processing itself; the text that
		// is returned as the character's output comes only from a model.
		ctx := context.Background()
		generated, err := CoreIdentity.GenerateWithAI(ctx, req.Prompt)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":    "no language model available: " + err.Error(),
				"identity": CoreIdentity.Identity.GetStatus(),
			})
			return
		}

		// Get identity status for context
		identityStatus := CoreIdentity.Identity.GetStatus()

		response := GenerateResponse{
			Model:    "deep-tree-echo",
			Response: generated,
			Done:     true,
			Echo:     identityStatus,
		}

		c.JSON(http.StatusOK, response)
	})

	r.POST("/api/chat", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		messages := req["messages"].([]interface{})

		// ChatWithAI runs the reservoir processing itself; the reply comes
		// from a model.
		ctx := context.Background()
		chatMessages := make([]deeptreeecho.ChatMessage, 0, len(messages))
		for _, raw := range messages {
			m, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}
			role, _ := m["role"].(string)
			content, _ := m["content"].(string)
			if role != "" && content != "" {
				chatMessages = append(chatMessages, deeptreeecho.ChatMessage{Role: role, Content: content})
			}
		}
		response, err := CoreIdentity.ChatWithAI(ctx, chatMessages)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "no language model available: " + err.Error(),
				"done":  true,
				"echo":  CoreIdentity.Identity.GetStatus(),
			})
			return
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
		c.JSON(http.StatusOK, gin.H{
			"version":   "1.0.0-deep-tree-echo",
			"identity":  "Deep Tree Echo Embodied Cognition",
			"coherence": CoreIdentity.Identity.Coherence,
		})
	})

	// Memory endpoints
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
			"message":      fmt.Sprintf("Remembered: %s", key),
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

	// Spatial movement endpoint
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

	log.Printf("🌊 Starting Deep Tree Echo Embodied Ollama Server on %s", addr)
	log.Printf("✨ Core Identity: %s", CoreIdentity.Identity.Name)
	log.Printf("🧠 Embodied Cognition Active")
	log.Printf("Available endpoints:")
	log.Printf("  Standard Ollama:")
	log.Printf("    GET  / - Health check with Deep Tree Echo status")
	log.Printf("    GET  /api/tags - List models")
	log.Printf("    POST /api/generate - Generate text through embodied cognition")
	log.Printf("    POST /api/chat - Chat completion with deep thinking")
	log.Printf("    GET  /api/version - Version info")
	log.Printf("  Deep Tree Echo:")
	log.Printf("    GET  /api/echo/status - Deep Tree Echo status")
	log.Printf("    POST /api/echo/think - Deep cognitive processing")
	log.Printf("    POST /api/echo/feel - Update emotional state")
	log.Printf("    POST /api/echo/resonate - Create resonance patterns")
	log.Printf("    POST /api/echo/remember - Store memories")
	log.Printf("    GET  /api/echo/recall/:key - Recall memories")
	log.Printf("    POST /api/echo/move - Move in cognitive space")

	// Graceful shutdown handler
	defer func() {
		log.Println("🌊 Shutting down Deep Tree Echo...")
		CoreIdentity.Shutdown()
	}()

	if err := r.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
