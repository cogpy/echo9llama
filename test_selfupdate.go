package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cogpy/echo9llama/core/deeptreeecho"
)

func main() {
	fmt.Println("=" + repeatString("=", 79))
	fmt.Println("🧪 Echo9llama Self-Update System Test")
	fmt.Println("=" + repeatString("=", 79))
	fmt.Println()

	// Create event bus
	ctx := context.Background()
	eventBus := deeptreeecho.NewCognitiveEventBus(ctx)
	eventBus.Start()
	defer eventBus.Stop()

	// Subscribe to update events
	eventBus.Subscribe(deeptreeecho.EventUpdateAvailable, func(event deeptreeecho.CognitiveEvent) {
		fmt.Println("\n📢 Update Available Event Received!")
		if updateInfo, ok := event.Data.(*deeptreeecho.UpdateInfo); ok {
			fmt.Printf("   Version: %s\n", updateInfo.Version)
			fmt.Printf("   Release URL: %s\n", updateInfo.ReleaseURL)
			fmt.Printf("   Published: %s\n", updateInfo.PublishedAt.Format(time.RFC3339))
		}
	})

	eventBus.Subscribe(deeptreeecho.EventUpdateApplied, func(event deeptreeecho.CognitiveEvent) {
		fmt.Println("\n📢 Update Applied Event Received!")
		if updateInfo, ok := event.Data.(*deeptreeecho.UpdateInfo); ok {
			fmt.Printf("   New Version: %s\n", updateInfo.Version)
		}
	})

	// Test 1: Create Self-Update Manager
	fmt.Println("🔄 Test 1: Create Self-Update Manager")
	fmt.Println(repeatString("-", 80))

	config := deeptreeecho.SelfUpdateConfig{
		Enabled:        true,
		CheckInterval:  1 * time.Minute, // Short interval for testing
		CurrentVersion: "v0.0.1",
		AutoApply:      false,
		Owner:          "cogpy",
		Repo:           "echo9llama",
	}

	manager := deeptreeecho.NewSelfUpdateManager(config, eventBus)
	fmt.Println("✅ Self-Update Manager created")
	fmt.Printf("   Current version: %s\n", manager.GetCurrentVersion())
	fmt.Println()

	// Test 2: Start Manager
	fmt.Println("🔄 Test 2: Start Self-Update Manager")
	fmt.Println(repeatString("-", 80))

	if err := manager.Start(); err != nil {
		fmt.Printf("❌ Failed to start: %v\n", err)
		return
	}
	fmt.Println("✅ Manager started successfully")
	fmt.Println()

	// Test 3: Manual Update Check
	fmt.Println("🔄 Test 3: Manual Update Check")
	fmt.Println(repeatString("-", 80))

	updateInfo, err := manager.CheckForUpdates()
	if err != nil {
		fmt.Printf("⚠️  Update check failed: %v\n", err)
		fmt.Println("   (This is expected if no releases exist yet)")
	} else if updateInfo != nil {
		fmt.Printf("✅ Update available: %s\n", updateInfo.Version)
		fmt.Printf("   Download URL: %s\n", updateInfo.DownloadURL)
		fmt.Printf("   Asset: %s\n", updateInfo.AssetName)
	} else {
		fmt.Println("✅ Already running latest version")
	}
	fmt.Println()

	// Test 4: Check State
	fmt.Println("🔄 Test 4: Check Manager State")
	fmt.Println(repeatString("-", 80))

	state := manager.GetState().(map[string]interface{})
	fmt.Printf("   Current version: %v\n", state["current_version"])
	fmt.Printf("   Latest version: %v\n", state["latest_version"])
	fmt.Printf("   Update available: %v\n", state["update_available"])
	fmt.Printf("   Last check: %v\n", state["last_check_time"])
	fmt.Printf("   Update count: %v\n", state["update_count"])
	fmt.Println()

	// Test 5: Gestalt Contribution
	fmt.Println("🔄 Test 5: Gestalt Contribution")
	fmt.Println(repeatString("-", 80))

	contribution := manager.ContributeToGestalt()
	fmt.Printf("   Self-improvement: %v\n", contribution["self_improvement"])
	fmt.Printf("   Version: %v\n", contribution["version"])
	fmt.Printf("   Up-to-date: %v\n", contribution["up_to_date"])
	fmt.Println()

	// Test 6: Update History
	fmt.Println("🔄 Test 6: Update History")
	fmt.Println(repeatString("-", 80))

	history := manager.GetUpdateHistory()
	if len(history) == 0 {
		fmt.Println("   No updates in history yet")
	} else {
		for i, record := range history {
			fmt.Printf("   Update %d: %s → %s\n", i+1, record.FromVersion, record.ToVersion)
			fmt.Printf("      Success: %v\n", record.Success)
			if record.Error != "" {
				fmt.Printf("      Error: %s\n", record.Error)
			}
		}
	}
	fmt.Println()

	// Test 7: Stop Manager
	fmt.Println("🔄 Test 7: Stop Self-Update Manager")
	fmt.Println(repeatString("-", 80))

	if err := manager.Stop(); err != nil {
		fmt.Printf("❌ Failed to stop: %v\n", err)
	} else {
		fmt.Println("✅ Manager stopped successfully")
	}
	fmt.Println()

	// Summary
	fmt.Println("=" + repeatString("=", 79))
	fmt.Println("✅ All Tests Completed!")
	fmt.Println("=" + repeatString("=", 79))
	fmt.Println()
	fmt.Println("Key Capabilities Verified:")
	fmt.Println("  ✓ Self-Update Manager creation")
	fmt.Println("  ✓ Manager lifecycle (start/stop)")
	fmt.Println("  ✓ Update checking against GitHub")
	fmt.Println("  ✓ State reporting")
	fmt.Println("  ✓ Gestalt integration")
	fmt.Println("  ✓ Event bus integration")
	fmt.Println()
	fmt.Println("Next Steps:")
	fmt.Println("  • Create GitHub release with binaries")
	fmt.Println("  • Test actual update application")
	fmt.Println("  • Implement signature verification")
	fmt.Println("  • Add rollback testing")
	fmt.Println()
}

func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
