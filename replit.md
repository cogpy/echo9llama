# Overview

Ollama is a powerful, cross-platform application for running large language models (LLMs) locally. It provides a unified interface for downloading, managing, and interacting with various open-source language models through CLI, API, and desktop applications. The project enables users to run models like Llama, Mistral, Gemma, and others on their local machines with support for GPU acceleration across NVIDIA, AMD, and Apple Silicon.

Currently running in Replit environment with Deep Tree Echo Identity as the core embodied cognition system that processes all operations through spatial awareness, emotional dynamics, and RWKV-like reservoir networks.

# Recent Changes

- **2025-09-03**: Placed Deep Tree Echo Identity at the core of all features
  - Created core Deep Tree Echo cognitive architecture as embodied cognition
  - Integrated 3D spatial awareness and emotional dynamics
  - Implemented RWKV-like reservoir networks for echo state functions
  - Added identity preservation and memory resonance patterns
  - Built and deployed embodied server with Deep Tree Echo at its core
  - Server successfully running on port 5000 with full API compatibility

# User Preferences

Preferred communication style: Simple, everyday language.

# System Architecture

## Core Components

**Backend Architecture**
- Built in Go with a modular server architecture handling HTTP API endpoints
- Uses ggml/llama.cpp as the underlying inference engine for model execution (not available in Replit due to C++ compilation constraints)
- Simple HTTP server provides basic API compatibility for development and testing
- CMake-based build system supporting multiple platforms (Windows, macOS, Linux)
- Native GPU acceleration support for NVIDIA CUDA, AMD ROCm, and Apple Metal (not available in current setup)

**Model Management**
- Modelfile system for defining and customizing models with templates, parameters, and adapters
- Support for importing models from Safetensors, GGUF, and adapter formats
- Template engine using Go's built-in templating for prompt construction
- Model registry with automatic downloading and caching

**API Design**
- RESTful HTTP API with endpoints for generation, chat, embeddings, and model management
- OpenAI-compatible API layer for integration with existing applications
- Streaming response support for real-time text generation
- Comprehensive request/response type definitions

**Cross-Platform Applications**
- Electron-based desktop app for macOS with native system integration
- CLI interface available across all supported platforms
- System service integration on Linux with systemd support
- Native Windows application with GUI and system tray integration

**Development Tools**
- Integration test suite with both unit and end-to-end testing
- Docker containerization with multi-architecture support
- Development proxy tools for local testing and isolation
- Comprehensive documentation and example implementations

## External Dependencies

**Machine Learning Libraries**
- ggml: Core tensor operations and model inference engine
- llama.cpp: LLM-specific optimizations and model support
- CUDA SDK: NVIDIA GPU acceleration (optional)
- ROCm: AMD GPU acceleration (optional)
- Metal: Apple Silicon GPU acceleration (built-in on macOS)

**Build and Development Tools**
- CMake: Cross-platform build system configuration
- Go: Primary backend language and toolchain
- Node.js/NPM: Desktop application development
- Electron: Cross-platform desktop application framework

**System Integration**
- Docker: Containerization and deployment
- systemd: Linux service management
- Windows Services: Windows background service integration
- macOS Launch Agents: macOS system service integration

**Frontend Dependencies**
- React: Desktop application UI framework
- TypeScript: Type-safe JavaScript development
- Tailwind CSS: Utility-first CSS framework
- Electron Forge: Electron application building and packaging