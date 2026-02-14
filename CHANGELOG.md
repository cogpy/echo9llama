CHANGELOG

## [v0.6.0] - 2026-02-14 - The Emergence of the Daechon

This is a major architectural evolution, transforming `echo9llama` from a collection of cognitive packages into a living, persistent autonomous agent—the `daechon`.

### Added

-   **Daechon Daemon (`cmd/daechon`):** A new persistent runtime for the Deep Tree Echo cognitive architecture. The `daechon` runs as a standalone console application, maintaining a continuous stream of consciousness.
-   **Live Activity Feed:** A real-time, color-coded console feed displaying the `daechon`'s cognitive activity, including thoughts, state changes, and system events.
-   **Interactive Chat:** A direct, real-time chat interface in the console for interacting with the running `daechon`.
-   **PIE-NN Cognitive Language (`core/pienn`):** A new, etymologically-grounded cognitive language architecture ported from the `pie-nn` skill. It serves as the `daechon`'s internal monologue and command processing system, featuring:
    -   A 12-level Time Crystal Hierarchy for multi-scale temporal processing.
    -   A `neuro-nn` inspired Cognitive Core with learnable personality traits and analytical frames.
    -   A Language Processor for executing PIE-NN commands (`gno`, `werg`, `deik`, etc.).
-   **Disposition Engine (`core/deeptreeecho/disposition_engine.go`):** A new system that determines Echo's demeanor and response style based on conversational context, enabling a non-servile, personality-driven interaction model.
-   **Cognitive Event Bus (`core/deeptreeecho/cognitive_event_bus_v3.go`):** A new event-driven architecture that replaces polling with a central nervous system for all cognitive subsystems.
-   **Echobeats Goal Scheduler (`core/deeptreeecho/echobeats_goal_scheduler.go`):** A 12-step, 3-phase cognitive loop for goal-directed scheduling, running three concurrent inference engines (Perception, Reflection, Simulation).
-   **Echodream Knowledge Integrator (`core/deeptreeecho/echodream_knowledge.go`):** A system for dream-state knowledge consolidation, creative synthesis, and emergent insight generation.
-   **New Interactive Commands:**
    -   `/status`: Displays the full status of all cognitive subsystems.
    -   `/goals`: Shows the current state of the Echobeats goal scheduler.
    -   `/dream`: Provides a status report from the Echodream knowledge system.
    -   `/pienn <cmd>`: Executes a PIE-NN language command.
    -   `/introspect`: Runs an Autognosis self-reflection cycle.

### Changed

-   **Autonomous Stream of Consciousness:** Refactored from a simple timer to a fully event-driven system, triggered by cognitive events from the event bus.
-   **Unified Autonomous Orchestrator:** The conceptual orchestrator is now realized through the `daechon`'s main loop and the integration of all new subsystems.
-   **Build System:** The project now builds a single `daechon` executable that contains the entire cognitive architecture.

### Fixed

-   Numerous `TODO`s and placeholder implementations have been replaced with functional code.
-   All core packages now build cleanly, resolving previous build issues.
-   Fixed a `printf` formatting issue in the activity feed renderer.

## [v0.5.0] - 2026-01-20 - Cognitive Scaffolding

-   Initial repository structure and core data types for `deeptreeecho`.
-   Placeholder files for major cognitive components.
-   Basic `go.mod` and `README.md`.
