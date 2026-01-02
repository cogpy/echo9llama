package deeptreeecho

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/minio/selfupdate"
)

// SelfUpdateManager manages autonomous self-updating for the agent
type SelfUpdateManager struct {
	mu                  sync.RWMutex
	ctx                 context.Context
	cancel              context.CancelFunc
	
	// Configuration
	config              SelfUpdateConfig
	
	// State
	currentVersion      string
	latestVersion       string
	updateAvailable     bool
	lastCheckTime       time.Time
	lastUpdateTime      time.Time
	
	// Update history
	updateHistory       []UpdateRecord
	
	// Event bus for notifications
	eventBus            *CognitiveEventBus
	
	// Running state
	running             bool
	checkTicker         *time.Ticker
}

// SelfUpdateConfig configures the self-update system
type SelfUpdateConfig struct {
	Enabled           bool          // Enable/disable self-updates
	CheckInterval     time.Duration // How often to check for updates
	UpdateURL         string        // GitHub releases URL pattern
	CurrentVersion    string        // Current version
	AutoApply         bool          // Automatically apply updates
	RequireSignature  bool          // Require code signature verification
	PublicKeyPath     string        // Path to public key for verification
	Owner             string        // GitHub repository owner
	Repo              string        // GitHub repository name
}

// UpdateInfo contains information about an available update
type UpdateInfo struct {
	Version       string
	ReleaseURL    string
	DownloadURL   string
	Checksum      string
	ReleaseNotes  string
	PublishedAt   time.Time
	AssetName     string
}

// UpdateRecord tracks a completed update
type UpdateRecord struct {
	FromVersion   string
	ToVersion     string
	UpdatedAt     time.Time
	Success       bool
	Error         string
	RollbackUsed  bool
}

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int    `json:"size"`
	} `json:"assets"`
}

// NewSelfUpdateManager creates a new self-update manager
func NewSelfUpdateManager(config SelfUpdateConfig, eventBus *CognitiveEventBus) *SelfUpdateManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Set defaults
	if config.CheckInterval == 0 {
		config.CheckInterval = 24 * time.Hour
	}
	if config.Owner == "" {
		config.Owner = "cogpy"
	}
	if config.Repo == "" {
		config.Repo = "echo9llama"
	}
	if config.UpdateURL == "" {
		config.UpdateURL = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", config.Owner, config.Repo)
	}
	
	return &SelfUpdateManager{
		ctx:            ctx,
		cancel:         cancel,
		config:         config,
		currentVersion: config.CurrentVersion,
		updateHistory:  make([]UpdateRecord, 0),
		eventBus:       eventBus,
	}
}

// Start begins the self-update manager
func (sum *SelfUpdateManager) Start() error {
	sum.mu.Lock()
	defer sum.mu.Unlock()
	
	if !sum.config.Enabled {
		fmt.Println("🔄 Self-update system disabled")
		return nil
	}
	
	if sum.running {
		return fmt.Errorf("already running")
	}
	
	sum.running = true
	
	fmt.Printf("🔄 Starting Self-Update Manager (current version: %s)\n", sum.currentVersion)
	fmt.Printf("   Check interval: %v\n", sum.config.CheckInterval)
	fmt.Printf("   Auto-apply: %v\n", sum.config.AutoApply)
	
	// Start periodic update checks
	sum.checkTicker = time.NewTicker(sum.config.CheckInterval)
	go sum.runUpdateChecker()
	
	// Do an initial check after a short delay
	go func() {
		time.Sleep(30 * time.Second)
		sum.CheckForUpdates()
	}()
	
	return nil
}

// Stop gracefully stops the self-update manager
func (sum *SelfUpdateManager) Stop() error {
	sum.mu.Lock()
	defer sum.mu.Unlock()
	
	if !sum.running {
		return nil
	}
	
	fmt.Println("🔄 Stopping self-update manager...")
	
	sum.running = false
	if sum.checkTicker != nil {
		sum.checkTicker.Stop()
	}
	sum.cancel()
	
	return nil
}

// runUpdateChecker runs the periodic update checker
func (sum *SelfUpdateManager) runUpdateChecker() {
	for {
		select {
		case <-sum.ctx.Done():
			return
		case <-sum.checkTicker.C:
			sum.CheckForUpdates()
		}
	}
}

// CheckForUpdates queries GitHub for available updates
func (sum *SelfUpdateManager) CheckForUpdates() (*UpdateInfo, error) {
	sum.mu.Lock()
	sum.lastCheckTime = time.Now()
	sum.mu.Unlock()
	
	fmt.Println("🔄 Checking for updates...")
	
	// Fetch latest release from GitHub
	updateInfo, err := sum.fetchLatestRelease()
	if err != nil {
		fmt.Printf("⚠️  Failed to check for updates: %v\n", err)
		return nil, err
	}
	
	if updateInfo == nil {
		fmt.Println("✅ Already running the latest version")
		return nil, nil
	}
	
	sum.mu.Lock()
	sum.latestVersion = updateInfo.Version
	sum.updateAvailable = true
	sum.mu.Unlock()
	
	fmt.Printf("🎉 Update available: %s → %s\n", sum.currentVersion, updateInfo.Version)
	fmt.Printf("   Release notes: %s\n", truncateUpdate(updateInfo.ReleaseNotes, 100))
	
	// Publish event
	if sum.eventBus != nil {
		sum.eventBus.Publish(CognitiveEvent{
			Type:      EventUpdateAvailable,
			Source:    "selfupdate_manager",
			Timestamp: time.Now(),
			Data:      updateInfo,
			Priority:  0.8,
		})
	}
	
	// Auto-apply if configured
	if sum.config.AutoApply {
		fmt.Println("🔄 Auto-applying update...")
		return updateInfo, sum.ApplyUpdate(updateInfo)
	}
	
	return updateInfo, nil
}

// fetchLatestRelease fetches the latest release from GitHub
func (sum *SelfUpdateManager) fetchLatestRelease() (*UpdateInfo, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	
	req, err := http.NewRequest("GET", sum.config.UpdateURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set User-Agent to avoid GitHub rate limiting
	req.Header.Set("User-Agent", fmt.Sprintf("echo9llama/%s", sum.currentVersion))
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release: %w", err)
	}
	
	// Skip drafts and prereleases
	if release.Draft || release.Prerelease {
		return nil, nil
	}
	
	// Check if this is a newer version
	if !sum.isNewerVersion(release.TagName) {
		return nil, nil
	}
	
	// Find the appropriate asset for this platform
	assetName := sum.getAssetName()
	var downloadURL string
	
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}
	
	if downloadURL == "" {
		return nil, fmt.Errorf("no asset found for platform: %s", assetName)
	}
	
	return &UpdateInfo{
		Version:      release.TagName,
		ReleaseURL:   release.HTMLURL,
		DownloadURL:  downloadURL,
		ReleaseNotes: release.Body,
		PublishedAt:  release.PublishedAt,
		AssetName:    assetName,
	}, nil
}

// isNewerVersion checks if the given version is newer than current
func (sum *SelfUpdateManager) isNewerVersion(version string) bool {
	// Simple string comparison for now
	// In production, use proper semantic versioning
	return version > sum.currentVersion
}

// getAssetName returns the asset name for the current platform
func (sum *SelfUpdateManager) getAssetName() string {
	// Format: echo-autonomous-{os}-{arch}[.exe]
	osName := runtime.GOOS
	archName := runtime.GOARCH
	
	assetName := fmt.Sprintf("echo-autonomous-%s-%s", osName, archName)
	
	if osName == "windows" {
		assetName += ".exe"
	}
	
	return assetName
}

// ApplyUpdate downloads and applies an update
func (sum *SelfUpdateManager) ApplyUpdate(updateInfo *UpdateInfo) error {
	fmt.Printf("🔄 Applying update: %s → %s\n", sum.currentVersion, updateInfo.Version)
	
	// Download the update
	resp, err := http.Get(updateInfo.DownloadURL)
	if err != nil {
		return sum.recordUpdate(updateInfo.Version, false, fmt.Sprintf("download failed: %v", err))
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return sum.recordUpdate(updateInfo.Version, false, fmt.Sprintf("download failed with status: %d", resp.StatusCode))
	}
	
	// Apply the update
	opts := selfupdate.Options{}
	
	// TODO: Add checksum verification
	// if updateInfo.Checksum != "" {
	//     opts.Checksum = []byte(updateInfo.Checksum)
	// }
	
	// TODO: Add signature verification
	// if sum.config.RequireSignature && sum.config.PublicKeyPath != "" {
	//     // Load public key and verify signature
	// }
	
	err = selfupdate.Apply(resp.Body, opts)
	if err != nil {
		// Rollback if apply fails
		if rollbackErr := selfupdate.RollbackError(err); rollbackErr != nil {
			return sum.recordUpdate(updateInfo.Version, false, fmt.Sprintf("update failed and rollback failed: %v, %v", err, rollbackErr))
		}
		return sum.recordUpdate(updateInfo.Version, false, fmt.Sprintf("update failed (rolled back): %v", err))
	}
	
	// Record successful update
	sum.recordUpdate(updateInfo.Version, true, "")
	
	fmt.Printf("✅ Update applied successfully: %s\n", updateInfo.Version)
	fmt.Println("🔄 Restart required to use new version")
	
	// Publish event
	if sum.eventBus != nil {
		sum.eventBus.Publish(CognitiveEvent{
			Type:      EventUpdateApplied,
			Source:    "selfupdate_manager",
			Timestamp: time.Now(),
			Data:      updateInfo,
			Priority:  0.9,
		})
	}
	
	return nil
}

// recordUpdate records an update attempt
func (sum *SelfUpdateManager) recordUpdate(toVersion string, success bool, errorMsg string) error {
	sum.mu.Lock()
	defer sum.mu.Unlock()
	
	record := UpdateRecord{
		FromVersion: sum.currentVersion,
		ToVersion:   toVersion,
		UpdatedAt:   time.Now(),
		Success:     success,
		Error:       errorMsg,
	}
	
	sum.updateHistory = append(sum.updateHistory, record)
	
	if success {
		sum.currentVersion = toVersion
		sum.lastUpdateTime = time.Now()
		sum.updateAvailable = false
	}
	
	if errorMsg != "" {
		return fmt.Errorf("%s", errorMsg)
	}
	
	return nil
}

// GetUpdateHistory returns the update history
func (sum *SelfUpdateManager) GetUpdateHistory() []UpdateRecord {
	sum.mu.RLock()
	defer sum.mu.RUnlock()
	
	history := make([]UpdateRecord, len(sum.updateHistory))
	copy(history, sum.updateHistory)
	return history
}

// GetCurrentVersion returns the current version
func (sum *SelfUpdateManager) GetCurrentVersion() string {
	sum.mu.RLock()
	defer sum.mu.RUnlock()
	return sum.currentVersion
}

// GetLatestVersion returns the latest available version
func (sum *SelfUpdateManager) GetLatestVersion() string {
	sum.mu.RLock()
	defer sum.mu.RUnlock()
	return sum.latestVersion
}

// IsUpdateAvailable returns whether an update is available
func (sum *SelfUpdateManager) IsUpdateAvailable() bool {
	sum.mu.RLock()
	defer sum.mu.RUnlock()
	return sum.updateAvailable
}

// GetID returns the subsystem ID for telemetry shell integration
func (sum *SelfUpdateManager) GetID() string {
	return "selfupdate_manager"
}

// GetState returns the current state for telemetry shell integration
func (sum *SelfUpdateManager) GetState() interface{} {
	sum.mu.RLock()
	defer sum.mu.RUnlock()
	
	return map[string]interface{}{
		"current_version":   sum.currentVersion,
		"latest_version":    sum.latestVersion,
		"update_available":  sum.updateAvailable,
		"last_check_time":   sum.lastCheckTime,
		"last_update_time":  sum.lastUpdateTime,
		"update_count":      len(sum.updateHistory),
	}
}

// UpdateFromGestalt updates from the gestalt state
func (sum *SelfUpdateManager) UpdateFromGestalt(gestalt *GestaltState) error {
	// Self-update manager doesn't need gestalt updates
	return nil
}

// ContributeToGestalt contributes to the gestalt state
func (sum *SelfUpdateManager) ContributeToGestalt() map[string]interface{} {
	return map[string]interface{}{
		"self_improvement": sum.updateAvailable,
		"version":          sum.currentVersion,
		"up_to_date":       !sum.updateAvailable,
	}
}

// truncateUpdate truncates a string to maxLen
func truncateUpdate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Event types for self-update are defined in cognitive_event_bus.go
// EventUpdateAvailable: "update_available"
// EventUpdateApplied: "update_applied"
