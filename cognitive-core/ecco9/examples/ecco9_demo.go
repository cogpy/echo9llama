package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/EchoCog/echollama/core/ecco9"
	"github.com/EchoCog/echollama/core/ecco9/drivers"
	"github.com/EchoCog/echollama/core/llm"
)

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)

	fmt.Println("🌊 ecco9 Platform Demo - Cognitive Hardware Interaction")
	fmt.Println("=========================================================\n")

	// Create and boot platform
	platform := ecco9.NewPlatform(ecco9.DefaultConfiguration())
	ctx := context.Background()

	if err := platform.Boot(ctx); err != nil {
		log.Fatalf("Boot failed: %v", err)
	}

	// Register all drivers
	registerDrivers(platform, ctx)

	// Demonstrate device interactions
	demoReservoirProcessing(platform)
	demoConsciousnessLayers(platform)
	demoEmotionProcessing(platform)
	demoMemoryStorage(platform)

	// Show platform status
	showPlatformStatus(platform)

	// Shutdown
	fmt.Println("\n🌙 Shutting down platform...")
	if err := platform.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}

	fmt.Println("✨ Demo complete!")
}

func registerDrivers(platform *ecco9.Platform, ctx context.Context) {
	fmt.Println("📦 Registering Cognitive Drivers...")

	drivers := []struct {
		driver ecco9.Driver
		config interface{}
	}{
		{drivers.NewReservoirDriver(), drivers.DefaultReservoirConfig()},
		{drivers.NewConsciousnessDriver(), drivers.DefaultConsciousnessConfig()},
		{drivers.NewEmotionDriver(), drivers.DefaultEmotionConfig()},
		{drivers.NewMemoryDriver(), drivers.DefaultMemoryConfig()},
		{drivers.NewLLMDriver(llm.NewProviderManager()), nil},
	}

	for _, d := range drivers {
		if err := d.driver.Load(d.config); err != nil {
			log.Fatalf("Failed to load %s driver: %v", d.driver.GetName(), err)
		}
		if err := platform.RegisterDriver(d.driver); err != nil {
			log.Fatalf("Failed to register %s driver: %v", d.driver.GetName(), err)
		}

		for _, device := range d.driver.ListDevices() {
			if err := device.Initialize(ctx); err != nil {
				log.Printf("Warning: Failed to initialize device %s: %v", device.GetID(), err)
			}
			if err := platform.RegisterDevice(device); err != nil {
				log.Printf("Warning: Failed to register device %s: %v", device.GetID(), err)
			}
		}
	}

	fmt.Printf("   Registered %d drivers with %d devices\n\n",
		len(platform.Drivers), len(platform.Devices))
}

func demoReservoirProcessing(platform *ecco9.Platform) {
	fmt.Println("🌀 Demo: Reservoir Processing")
	fmt.Println("--------------------------------")

	device, err := platform.GetDevice("reservoir0")
	if err != nil {
		log.Printf("Failed to get reservoir device: %v", err)
		return
	}

	// Write some input to the reservoir
	input := []byte("Temporal pattern recognition through echo states")
	n, err := device.Write(input)
	if err != nil {
		log.Printf("Write failed: %v", err)
		return
	}
	fmt.Printf("   Wrote %d bytes to reservoir\n", n)

	// Read reservoir state
	buffer := make([]byte, 256)
	n, err = device.Read(buffer)
	if err != nil {
		log.Printf("Read failed: %v", err)
		return
	}
	fmt.Printf("   Read %d bytes from reservoir: %s\n", n, buffer[:min(n, 50)])

	// Get metrics
	metrics, _ := device.GetMetrics()
	fmt.Printf("   Operations: %d, Avg Latency: %v\n\n",
		metrics.OperationCount, metrics.AverageLatency)
}

func demoConsciousnessLayers(platform *ecco9.Platform) {
	fmt.Println("🧠 Demo: Consciousness Layers")
	fmt.Println("--------------------------------")

	device, err := platform.GetDevice("consciousness0")
	if err != nil {
		log.Printf("Failed to get consciousness device: %v", err)
		return
	}

	// Process through consciousness layers
	input := []byte("Multi-layer awareness processing")
	n, err := device.Write(input)
	if err != nil {
		log.Printf("Write failed: %v", err)
		return
	}
	fmt.Printf("   Processed %d bytes through consciousness layers\n", n)

	// Read layer activations
	buffer := make([]byte, 256)
	n, err = device.Read(buffer)
	if err != nil {
		log.Printf("Read failed: %v", err)
		return
	}
	fmt.Printf("   Layer states: %s\n\n", buffer[:n])
}

func demoEmotionProcessing(platform *ecco9.Platform) {
	fmt.Println("❤️  Demo: Emotion Processing")
	fmt.Println("--------------------------------")

	device, err := platform.GetDevice("emotion0")
	if err != nil {
		log.Printf("Failed to get emotion device: %v", err)
		return
	}

	// Trigger emotions
	inputs := []string{
		"Discovering new patterns!",
		"Complex challenges ahead",
		"Breakthrough achieved!",
	}

	for _, input := range inputs {
		device.Write([]byte(input))
		time.Sleep(200 * time.Millisecond)

		buffer := make([]byte, 256)
		n, _ := device.Read(buffer)
		fmt.Printf("   Input: %s\n", input)
		fmt.Printf("   State: %s\n", buffer[:n])
	}
	fmt.Println()
}

func demoMemoryStorage(platform *ecco9.Platform) {
	fmt.Println("💾 Demo: Memory Storage")
	fmt.Println("--------------------------------")

	device, err := platform.GetDevice("memory0")
	if err != nil {
		log.Printf("Failed to get memory device: %v", err)
		return
	}

	// Store some memories
	memories := []string{
		"ecco9 is a virtual cognitive hardware platform",
		"Deep Tree Echo provides embodied cognition",
		"Hypergraph memory supports multi-relational storage",
	}

	for _, memory := range memories {
		n, err := device.Write([]byte(memory))
		if err != nil {
			log.Printf("Store failed: %v", err)
			continue
		}
		fmt.Printf("   Stored %d bytes\n", n)
	}

	// Read memory statistics
	buffer := make([]byte, 256)
	n, _ := device.Read(buffer)
	fmt.Printf("   Memory stats: %s\n\n", buffer[:n])
}

func showPlatformStatus(platform *ecco9.Platform) {
	fmt.Println("📊 Platform Status Summary")
	fmt.Println("==========================")

	status := platform.GetStatus()
	fmt.Printf("Firmware: %v\n", status["firmware_version"])
	fmt.Printf("Kernel: %v\n", status["kernel_version"])
	fmt.Printf("Boot Stage: %v\n", status["boot_stage"])
	fmt.Printf("Devices: %v\n", status["device_count"])
	fmt.Printf("Drivers: %v\n", status["driver_count"])

	fmt.Println("\n📍 Device Health:")
	for id, device := range platform.Devices {
		health, _ := device.GetHealth()
		state, _ := device.GetState()
		fmt.Printf("   %s: %s (health: %s)\n", id, state.Status, health)
	}

	fmt.Println("\n🔧 Driver Capabilities:")
	for name, driver := range platform.Drivers {
		fmt.Printf("   %s v%s:\n", name, driver.GetVersion())
		for _, cap := range driver.GetCapabilities() {
			fmt.Printf("      - %s\n", cap)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
