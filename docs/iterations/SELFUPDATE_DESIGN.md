# Self-Update System Design for Echo9llama

## Overview

This document outlines the integration of the `minio/selfupdate` library into echo9llama to enable autonomous self-updating capabilities. The agent will be able to detect, download, and apply updates to itself, ensuring it stays current with the latest improvements.

## Architecture

### Components

1. **Self-Update Manager** (`core/deeptreeecho/selfupdate_manager.go`)
   - Manages the self-update lifecycle
   - Checks for updates periodically
   - Downloads and applies updates
   - Handles rollback on failure

2. **Update Source**
   - GitHub Releases (primary)
   - Custom update server (future)

3. **Integration Points**
   - Autonomous Agent: Triggers update checks
   - Global Telemetry Shell: Reports update status
   - Event Bus: Publishes update events

### Update Flow

```
1. Periodic Check (every 24 hours or on-demand)
   ↓
2. Query GitHub Releases API for latest version
   ↓
3. Compare with current version
   ↓
4. If newer version available:
   ↓
5. Download binary from release assets
   ↓
6. Verify checksum/signature (if available)
   ↓
7. Apply update using selfupdate.Apply()
   ↓
8. Restart agent with new binary
   ↓
9. Verify successful update
   ↓
10. Remove old binary (.old file)
```

### Configuration

```go
type SelfUpdateConfig struct {
    Enabled           bool          // Enable/disable self-updates
    CheckInterval     time.Duration // How often to check for updates
    UpdateURL         string        // URL pattern for updates
    CurrentVersion    string        // Current version
    AutoApply         bool          // Automatically apply updates
    RequireSignature  bool          // Require code signature verification
    PublicKeyPath     string        // Path to public key for verification
}
```

### Safety Mechanisms

1. **Backup**: Old binary is kept as `.old` file
2. **Rollback**: If new binary fails to start, restore from `.old`
3. **Verification**: Checksum and signature verification before applying
4. **Graceful Shutdown**: Ensure all cognitive loops complete before update
5. **State Persistence**: Save cognitive state before update, restore after

## Implementation Plan

### Phase 1: Core Self-Update Manager
- Create `SelfUpdateManager` struct
- Implement version checking against GitHub Releases
- Implement download and apply logic
- Add basic error handling and logging

### Phase 2: Integration with Autonomous Agent
- Add self-update manager to `AutonomousAgent`
- Integrate with event bus for update notifications
- Add update check to periodic tasks
- Implement graceful shutdown before update

### Phase 3: Safety and Verification
- Add checksum verification
- Implement rollback mechanism
- Add signature verification (minisign)
- Implement state persistence across updates

### Phase 4: Advanced Features
- Add update scheduling (update during low activity)
- Implement update channels (stable, beta, nightly)
- Add update history tracking
- Implement update notifications via discussion system

## API Design

```go
// SelfUpdateManager manages autonomous self-updating
type SelfUpdateManager struct {
    config            SelfUpdateConfig
    currentVersion    string
    latestVersion     string
    updateAvailable   bool
    lastCheckTime     time.Time
    eventBus          *EventBus
}

// CheckForUpdates queries for available updates
func (sum *SelfUpdateManager) CheckForUpdates() (*UpdateInfo, error)

// ApplyUpdate downloads and applies an update
func (sum *SelfUpdateManager) ApplyUpdate(updateInfo *UpdateInfo) error

// ScheduleUpdate schedules an update for a specific time
func (sum *SelfUpdateManager) ScheduleUpdate(updateInfo *UpdateInfo, when time.Time) error

// RollbackUpdate restores the previous version
func (sum *SelfUpdateManager) RollbackUpdate() error

// GetUpdateHistory returns the history of updates
func (sum *SelfUpdateManager) GetUpdateHistory() []UpdateRecord
```

## Security Considerations

1. **HTTPS Only**: All downloads over HTTPS
2. **Signature Verification**: Use minisign for code signing
3. **Checksum Verification**: SHA256 checksums for all binaries
4. **Trusted Sources**: Only update from official GitHub releases
5. **User Confirmation**: Optionally require confirmation before update

## Testing Strategy

1. **Unit Tests**: Test version comparison, download, apply logic
2. **Integration Tests**: Test full update flow with mock server
3. **Manual Tests**: Test actual updates from GitHub releases
4. **Rollback Tests**: Verify rollback works correctly
5. **State Persistence Tests**: Ensure cognitive state survives updates

## Deployment

### GitHub Release Process

1. Tag new version: `git tag v0.x.x`
2. Build binaries for all platforms
3. Generate checksums: `sha256sum echo-autonomous-* > checksums.txt`
4. Sign binaries: `minisign -S -m echo-autonomous-linux-amd64`
5. Create GitHub release with assets
6. Agent will auto-detect and update

### Release Assets Structure

```
v0.x.x/
├── echo-autonomous-linux-amd64
├── echo-autonomous-linux-amd64.minisig
├── echo-autonomous-darwin-amd64
├── echo-autonomous-darwin-amd64.minisig
├── echo-autonomous-windows-amd64.exe
├── echo-autonomous-windows-amd64.exe.minisig
└── checksums.txt
```

## Future Enhancements

1. **Delta Updates**: Binary patching for smaller downloads
2. **P2P Updates**: Distribute updates via peer network
3. **A/B Testing**: Run multiple versions for testing
4. **Update Analytics**: Track update success rates
5. **Custom Update Server**: Self-hosted update infrastructure
