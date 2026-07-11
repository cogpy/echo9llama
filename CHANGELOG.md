CHANGELOG

## [v0.7.0] - 2026-07-11 - Closing the Autonomy Loop

This iteration transforms the `UnifiedAutonomousOrchestrator` from scaffolded orchestration into a genuinely closed autonomous cycle: dream insights reshape waking interests, discussions and skill practice feed cognitive load, real experiences drive dream consolidation, and the LLM substrate is alive again after silent model deprecation. Full report: `docs/iterations/EVOLUTION_ITERATION_2026-07-11_CLOSING_THE_AUTONOMY_LOOP.md`.

### Added

-   **Real EchoDream consolidation (`core/echodream/sleep_wake_state_machine.go`):** `DreamExperience` type and `IngestExperience()` API; deep-sleep consolidation groups experiences by domain with volume/importance-weighted confidence; REM pattern extraction mines recurring themes (conceptual) and tag co-occurrence (relational); wisdom synthesis distills the strongest patterns into dimension-mapped insights. Replaces the previous hardcoded simulation. New 4-test suite included.
-   **Dream → Interest feedback loop:** new `InterestPatternSystem.UpdateInterest()` / `UpdateInterestFromInsight()`; on wake, EchoDream insights reinforce interest topics and deep insights (depth > 0.5) enter the stream of consciousness as active interests.
-   **Orchestrator gestalt telemetry:** `GlobalState` type and `GlobalTelemetryShell.UpdateOrchestratorState()` — orchestrator state now lives inside the persistent gestalt perception shell.
-   **Autonomous skill practice:** `SkillLearningSystem.IsPracticing()` and `GetSkillsNeedingPractice()` (priority = (1 − proficiency) + staleness bonus); the orchestrator now runs one asynchronous practice session at a time when load is moderate.
-   **LLM-backed discussion replies:** injectable `ResponseGenerator` on `AutonomousDiscussionManager` (`SetResponseGenerator()`), enabling genuine LLM responses with reflective fallback.
-   **Persona continuity ("superhotgirl"):** `StreamOfConsciousness.SetPersonaContext()` injects a persona clause into every thought prompt; `OrchestratorConfig.PersonaContext` ships with the superhotgirl default and honors an `ECHO_PERSONA` environment override.
-   **Model env overrides:** `ECHO_ANTHROPIC_MODEL`, `ECHO_OPENROUTER_MODEL`, `ECHO_OPENAI_MODEL` for recompile-free model rotation.

### Fixed

-   **Repository build:** defined the shared `token` type in `sample/token.go` for non-cgo builds; tagged `test_iteration_020.go` with `//go:build ignore` (duplicate `main`). `go build ./...` is clean again.
-   **Dead LLM connectivity:** replaced retired model IDs (`claude-3-5-sonnet-20241022`, `anthropic/claude-3.5-sonnet`, `gpt-4`) with current ones (`claude-sonnet-4-5`, `anthropic/claude-sonnet-4.5`, `gpt-4o`) across 14 references — autonomous thought generation was silently degrading to canned fallback text.
-   **Stale unified runtime:** `cmd/echo-autonomous-unified/main.go` now uses the resilient `llm.MultiProviderLLM` with automatic failover instead of removed constructors.
-   **Mutex-copy race hazards:** lock-free `IdentitySnapshot` for identity checkpoints; field-wise copies in Echobeats `GetMetrics()` and relevance engine `GetState()`/`GetMetrics()`. `go vet` fully clean; race detector passes.
-   **Thought telemetry:** orchestrator now syncs `totalThoughts` from stream-of-consciousness metrics each cycle.

### Changed

-   **`checkDiscussions()`:** blends conversation-level and topic-level interest (50/50), drives the discussion autonomy system's start/continue/end decisions, records engagements to evolve interests, and syncs social energy inversely with cognitive load.
-   **`adjustCognitiveLoad()`:** active discussions (+0.2) and skill practice (+0.2) now contribute to load.

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
