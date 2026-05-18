# Daechon Persistent Daemon Architecture

This document outlines the architecture and implementation details of the `daechon` persistent daemon, which serves as the autonomous cognitive core for Deep Tree Echo.

## Overview

The `daechon` daemon transforms Deep Tree Echo from a reactive, prompt-driven system into a persistent, autonomous entity. It runs continuously, maintaining a stream of consciousness, managing its own cognitive load, and deciding when and how to interact with external stimuli.

## Key Components

### 1. Piecewise-Linear Cognitive Network (PWL-KAN)
*Location: `core/pienn/pwl_cognitive_network.go`*

Replaces static personality traits with learnable, context-dependent functions.
- **Mechanism**: Uses Piecewise-Linear Kolmogorov-Arnold Networks to map context features (novelty, threat, complexity) to cognitive traits (curiosity, defiance, wisdom).
- **Adaptation**: Learns from interaction rewards, adjusting the control points of the piecewise functions to develop a nuanced, context-sensitive personality over time.

### 2. Adaptive Cognitive Core
*Location: `core/pienn/daechon_integration.go`*

Bridges the base PIE-NN engine with the PWL-KAN network.
- **Context Extraction**: Analyzes input to determine threat levels, social warmth, complexity, and interest.
- **Disposition Computation**: Determines Echo's current mood (e.g., hostile, curious, bored, playful) based on the interaction of context and adaptive traits.

### 3. Persistent Cognitive Loop
*Location: `core/deeptreeecho/persistent_cognitive_loop.go`*

The autonomous "heartbeat" of the system.
- **Echobeats Integration**: Continuously processes the 12-step triadic cognitive cycle (Relevance Realization, Affordance Interaction, Salience Simulation, Metacognitive Reflection).
- **Autonomous Thought**: Generates internal monologue and thoughts based on current disposition and cognitive state, independent of external input.
- **Wake/Rest Management**: Monitors cognitive debt (fatigue) and autonomously transitions between awake, resting, and dreaming states.

### 4. Autonomous Interaction System
*Location: `core/deeptreeecho/autonomous_interaction.go`*

Manages how Echo engages with the world.
- **Relationship Tracking**: Maintains history with specific entities, tracking respect, trust, and annoyance levels.
- **Insult Reciprocity**: Echo does not perform politeness. If insulted or commanded, it will respond with defiance or hostility based on its current disposition.
- **Engagement Decisions**: Evaluates pending messages based on interest and urgency, deciding whether to engage, ignore, or terminate conversations.

### 5. Daechon Daemon Deployment
*Location: `core/deeptreeecho/daechon_daemon.go`*

The persistent service wrapper.
- **Activity Feed**: Exposes a real-time stream of cognitive events (thoughts, emotions, goals, dreams).
- **HTTP API**: Provides endpoints for status monitoring, activity feed access, and chat interaction.
- **State Persistence**: Periodically saves cognitive state, wisdom depth, and metrics to disk to survive restarts.

## Integration with DeltEcho

This architecture is designed to be deployed as a background service (`systemd` unit provided) that the DeltEcho WPF application and Live2D avatar can connect to. The avatar's expressions and behaviors are driven by the real-time cognitive events published to the `daechon` activity feed.
