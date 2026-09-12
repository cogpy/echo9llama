package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EchoCog/echollama/core/ecco9"
	"github.com/EchoCog/echollama/core/ecco9/drivers"
	"github.com/EchoCog/echollama/core/llm"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	// Create ecco9 platform
	config := ecco9.DefaultConfiguration()
	platform := ecco9.NewPlatform(config)

	// Boot the platform
	ctx := context.Background()
	if err := platform.Boot(ctx); err != nil {
		log.Fatalf("Failed to boot ecco9 platform: %v", err)
	}

	// Create LLM provider manager
	providerManager := llm.NewProviderManager()

	// Register all drivers
	allDrivers := []struct {
		driver ecco9.Driver
		config interface{}
	}{
		{drivers.NewReservoirDriver(), drivers.DefaultReservoirConfig()},
		{drivers.NewConsciousnessDriver(), drivers.DefaultConsciousnessConfig()},
		{drivers.NewEmotionDriver(), drivers.DefaultEmotionConfig()},
		{drivers.NewMemoryDriver(), drivers.DefaultMemoryConfig()},
		{drivers.NewLLMDriver(providerManager), nil},
	}

	for _, d := range allDrivers {
		if err := d.driver.Load(d.config); err != nil {
			log.Fatalf("Failed to load %s driver: %v", d.driver.GetName(), err)
		}
		if err := platform.RegisterDriver(d.driver); err != nil {
			log.Fatalf("Failed to register %s driver: %v", d.driver.GetName(), err)
		}

		// Initialize devices from this driver
		for _, device := range d.driver.ListDevices() {
			if err := device.Initialize(ctx); err != nil {
				log.Printf("Warning: Failed to initialize device %s: %v", device.GetID(), err)
			}
			if err := platform.RegisterDevice(device); err != nil {
				log.Printf("Warning: Failed to register device %s: %v", device.GetID(), err)
			}
		}
	}

	// Create HTTP server
	router := setupRouter(platform, providerManager)

	// Start server
	port := config.Ports.HTTP
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("\n🛑 Received shutdown signal")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}

		if err := platform.Shutdown(shutdownCtx); err != nil {
			log.Printf("Platform shutdown error: %v", err)
		}
	}()

	log.Printf("\n🌊 ecco9 Platform listening on http://localhost:%d\n", port)
	log.Println("📡 Available endpoints:")
	log.Printf("   GET  /              - Platform status dashboard")
	log.Printf("   GET  /api/status    - Platform status (JSON)")
	log.Printf("   GET  /api/devices   - List devices")
	log.Printf("   GET  /api/drivers   - List drivers")
	log.Printf("   GET  /api/health    - Health check")
	log.Printf("   POST /api/generate  - Generate text")
	log.Println()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

func setupRouter(platform *ecco9.Platform, providerManager *llm.ProviderManager) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(cors.Default())

	// Root endpoint - Platform dashboard
	router.GET("/", func(c *gin.Context) {
		html := generateDashboardHTML(platform)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	// API endpoints
	api := router.Group("/api")
	{
		// Platform status
		api.GET("/status", func(c *gin.Context) {
			status := platform.GetStatus()
			c.JSON(http.StatusOK, gin.H{
				"platform": "ecco9",
				"status":   status,
			})
		})

		// List devices
		api.GET("/devices", func(c *gin.Context) {
			devices := make([]map[string]interface{}, 0)
			for id, device := range platform.Devices {
				state, _ := device.GetState()
				devices = append(devices, map[string]interface{}{
					"id":     id,
					"name":   device.GetName(),
					"type":   device.GetType(),
					"status": state.Status,
					"health": state.Health,
				})
			}
			c.JSON(http.StatusOK, gin.H{"devices": devices})
		})

		// List drivers
		api.GET("/drivers", func(c *gin.Context) {
			drivers := make([]map[string]interface{}, 0)
			for name, driver := range platform.Drivers {
				drivers = append(drivers, map[string]interface{}{
					"name":         name,
					"version":      driver.GetVersion(),
					"capabilities": driver.GetCapabilities(),
				})
			}
			c.JSON(http.StatusOK, gin.H{"drivers": drivers})
		})

		// Health check
		api.GET("/health", func(c *gin.Context) {
			healthy := true
			for _, device := range platform.Devices {
				health, _ := device.GetHealth()
				if health == ecco9.HealthStatusFailed || health == ecco9.HealthStatusCritical {
					healthy = false
					break
				}
			}

			status := http.StatusOK
			if !healthy {
				status = http.StatusServiceUnavailable
			}

			c.JSON(status, gin.H{
				"healthy":   healthy,
				"timestamp": time.Now(),
			})
		})

		// Generate text
		api.POST("/generate", func(c *gin.Context) {
			var req struct {
				Prompt string `json:"prompt"`
				Model  string `json:"model"`
			}

			if err := c.BindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// Use ecco9 platform for generation
			response := fmt.Sprintf(
				"🌊 ecco9 Cognitive Platform Response\n\n"+
					"Platform: %s\n"+
					"Firmware: %s\n"+
					"Devices: %d active\n"+
					"Prompt: %s\n\n"+
					"This is a demonstration of the ecco9 virtual cognitive hardware platform. "+
					"The platform provides device drivers, firmware interfaces, and hardware "+
					"simulation for Deep Tree Echo's embodied cognition architecture.\n",
				"ecco9",
				platform.Firmware.Version,
				len(platform.Devices),
				req.Prompt,
			)

			c.JSON(http.StatusOK, gin.H{
				"response": response,
				"platform": "ecco9",
			})
		})
	}

	return router
}

func generateDashboardHTML(platform *ecco9.Platform) string {
	status := platform.GetStatus()
	statusJSON, _ := json.MarshalIndent(status, "", "  ")

	devices := make([]map[string]interface{}, 0)
	for id, device := range platform.Devices {
		state, _ := device.GetState()
		devices = append(devices, map[string]interface{}{
			"id":     id,
			"name":   device.GetName(),
			"type":   device.GetType(),
			"status": state.Status,
			"health": state.Health,
		})
	}
	devicesJSON, _ := json.MarshalIndent(devices, "", "  ")

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>ecco9 - Cognitive Hardware Platform</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%s, #764ba2 100%s);
            color: #fff;
            margin: 0;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        h1 {
            font-size: 3em;
            margin: 0 0 10px 0;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.3);
        }
        .subtitle {
            font-size: 1.2em;
            opacity: 0.9;
            margin-bottom: 30px;
        }
        .card {
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(10px);
            border-radius: 15px;
            padding: 25px;
            margin-bottom: 20px;
            border: 1px solid rgba(255, 255, 255, 0.2);
        }
        .card h2 {
            margin-top: 0;
            font-size: 1.5em;
            border-bottom: 2px solid rgba(255, 255, 255, 0.3);
            padding-bottom: 10px;
        }
        pre {
            background: rgba(0, 0, 0, 0.3);
            padding: 15px;
            border-radius: 8px;
            overflow-x: auto;
            font-size: 0.9em;
        }
        .status-badge {
            display: inline-block;
            padding: 5px 15px;
            border-radius: 20px;
            font-size: 0.9em;
            font-weight: bold;
            background: rgba(0, 255, 0, 0.3);
            border: 1px solid rgba(0, 255, 0, 0.5);
        }
        .endpoints {
            list-style: none;
            padding: 0;
        }
        .endpoints li {
            padding: 8px 0;
            border-bottom: 1px solid rgba(255, 255, 255, 0.1);
        }
        .endpoints code {
            background: rgba(0, 0, 0, 0.3);
            padding: 3px 8px;
            border-radius: 4px;
            font-size: 0.9em;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🌊 ecco9</h1>
        <div class="subtitle">Cognitive Hardware Platform - Virtual Device Architecture</div>
        
        <div class="card">
            <h2>Platform Status <span class="status-badge">ONLINE</span></h2>
            <pre>%s</pre>
        </div>
        
        <div class="card">
            <h2>📍 Registered Devices</h2>
            <pre>%s</pre>
        </div>
        
        <div class="card">
            <h2>📡 API Endpoints</h2>
            <ul class="endpoints">
                <li><code>GET  /api/status</code> - Platform status</li>
                <li><code>GET  /api/devices</code> - List devices</li>
                <li><code>GET  /api/drivers</code> - List drivers</li>
                <li><code>GET  /api/health</code> - Health check</li>
                <li><code>POST /api/generate</code> - Generate text</li>
            </ul>
        </div>
        
        <div class="card">
            <h2>🔧 Architecture</h2>
            <p>
                ecco9 implements Deep Tree Echo's embodied cognition as a virtual hardware platform with:
            </p>
            <ul>
                <li><strong>Virtual Devices:</strong> Reservoir, Consciousness, Emotion, Memory processors</li>
                <li><strong>Device Drivers:</strong> CHAL (Cognitive Hardware Abstraction Layer)</li>
                <li><strong>Firmware:</strong> Boot sequence, kernel, identity loader</li>
                <li><strong>Streams:</strong> I/O streams for cognitive data flow</li>
                <li><strong>Telemetry:</strong> Real-time metrics and monitoring</li>
            </ul>
        </div>
    </div>
</body>
</html>
`, string(statusJSON), string(devicesJSON))

	return html
}
