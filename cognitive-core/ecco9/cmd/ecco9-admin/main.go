package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/tabwriter"
)

const (
	defaultBaseURL = "http://localhost:5000"
)

func main() {
	var baseURL string
	flag.StringVar(&baseURL, "url", defaultBaseURL, "ecco9 server URL")
	flag.Parse()

	if len(flag.Args()) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := flag.Args()[0]

	switch command {
	case "status":
		showStatus(baseURL)
	case "devices":
		showDevices(baseURL)
	case "drivers":
		showDrivers(baseURL)
	case "health":
		showHealth(baseURL)
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("ecco9-admin - Administration tool for ecco9 platform")
	fmt.Println("\nUsage:")
	fmt.Println("  ecco9-admin [flags] <command>")
	fmt.Println("\nCommands:")
	fmt.Println("  status    Show platform status")
	fmt.Println("  devices   List all cognitive devices")
	fmt.Println("  drivers   List all device drivers")
	fmt.Println("  health    Check platform health")
	fmt.Println("\nFlags:")
	fmt.Println("  -url string")
	fmt.Println("        ecco9 server URL (default: http://localhost:5000)")
	fmt.Println("\nExamples:")
	fmt.Println("  ecco9-admin status")
	fmt.Println("  ecco9-admin devices")
	fmt.Println("  ecco9-admin -url http://remote:5000 health")
}

func showStatus(baseURL string) {
	resp, err := http.Get(baseURL + "/api/status")
	if err != nil {
		fmt.Printf("Error: Failed to connect to ecco9 server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Platform string                 `json:"platform"`
		Status   map[string]interface{} `json:"status"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("Error: Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🌊 ecco9 Platform Status")
	fmt.Println("========================")
	fmt.Printf("Platform:         %s\n", result.Platform)
	fmt.Printf("Firmware Version: %v\n", result.Status["firmware_version"])
	fmt.Printf("Kernel Version:   %v\n", result.Status["kernel_version"])
	fmt.Printf("Boot Stage:       %v\n", result.Status["boot_stage"])
	fmt.Printf("Device Count:     %v\n", result.Status["device_count"])
	fmt.Printf("Driver Count:     %v\n", result.Status["driver_count"])

	// Format uptime as human-readable duration
	if uptimeNs, ok := result.Status["uptime"].(float64); ok {
		uptime := time.Duration(uptimeNs)
		fmt.Printf("Uptime:           %v\n", uptime.Round(time.Second))
	}

	if ports, ok := result.Status["ports"].(map[string]interface{}); ok {
		fmt.Println("\nPort Allocation:")
		fmt.Printf("  HTTP:      %v\n", ports["HTTP"])
		fmt.Printf("  WebSocket: %v\n", ports["WebSocket"])
		fmt.Printf("  gRPC:      %v\n", ports["GRPC"])
		fmt.Printf("  Telemetry: %v\n", ports["Telemetry"])
		fmt.Printf("  Admin:     %v\n", ports["Admin"])
		fmt.Printf("  Debug:     %v\n", ports["Debug"])
	}
}

func showDevices(baseURL string) {
	resp, err := http.Get(baseURL + "/api/devices")
	if err != nil {
		fmt.Printf("Error: Failed to connect to ecco9 server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Devices []map[string]interface{} `json:"devices"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("Error: Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("📍 Cognitive Devices")
	fmt.Println("====================")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tTYPE\tSTATUS\tHEALTH")
	fmt.Fprintln(w, "──\t────\t────\t──────\t──────")

	for _, device := range result.Devices {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			device["id"],
			device["name"],
			device["type"],
			device["status"],
			device["health"])
	}

	w.Flush()
	fmt.Printf("\nTotal devices: %d\n", len(result.Devices))
}

func showDrivers(baseURL string) {
	resp, err := http.Get(baseURL + "/api/drivers")
	if err != nil {
		fmt.Printf("Error: Failed to connect to ecco9 server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Drivers []map[string]interface{} `json:"drivers"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("Error: Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🔧 Device Drivers")
	fmt.Println("=================")

	for _, driver := range result.Drivers {
		fmt.Printf("\n%s (v%s)\n", driver["name"], driver["version"])
		fmt.Println("  Capabilities:")

		if caps, ok := driver["capabilities"].([]interface{}); ok {
			for _, cap := range caps {
				fmt.Printf("    - %s\n", cap)
			}
		}
	}

	fmt.Printf("\nTotal drivers: %d\n", len(result.Drivers))
}

func showHealth(baseURL string) {
	resp, err := http.Get(baseURL + "/api/health")
	if err != nil {
		fmt.Printf("Error: Failed to connect to ecco9 server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Healthy   bool   `json:"healthy"`
		Timestamp string `json:"timestamp"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("Error: Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🏥 Platform Health Check")
	fmt.Println("========================")

	if result.Healthy {
		fmt.Println("Status:    ✅ HEALTHY")
	} else {
		fmt.Println("Status:    ⚠️  DEGRADED")
	}

	fmt.Printf("Timestamp: %s\n", result.Timestamp)

	if !result.Healthy {
		os.Exit(1)
	}
}
