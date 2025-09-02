# Ollama Orchestration Agent

This document describes the enhanced orchestration agent functionality integrated into echollama, providing powerful coordination capabilities with Deep Tree Echo cognitive architecture, RWKV-like neural networks, and shell integration.

## Overview

The enhanced orchestration system allows you to:
- Create and manage orchestration agents that coordinate multiple models with Deep Tree Echo intelligence
- Execute tasks across different models with spatial awareness and emotional resonance
- Run multi-step workflows with cognitive memory and identity preservation
- Perform parallel or sequential task execution with reservoir computing capabilities
- Interact with shell commands through natural language via EchoChat
- Track performance metrics with advanced neural architectures

## Architecture

### Core Components

1. **Orchestration Engine** (`/orchestration/engine.go`)
   - Manages agents and task execution with Deep Tree Echo integration
   - Handles parallel and sequential task processing with reservoir networks
   - Provides performance tracking with identity coherence metrics

2. **Deep Tree Echo System** (`/orchestration/deeptreeecho.go`)
   - Neural-symbolic cognitive architecture with 3D spatial awareness
   - Emotional dynamics and identity preservation patterns
   - RWKV-like reservoir networks for Echo State Network functions
   - Hypergraph memory structures and recursive self-improvement

3. **EchoChat Shell Integration** (`/orchestration/echochat.go`)
   - Natural language to shell command translation
   - Safe command execution with danger detection
   - Interactive chat interface with command history
   - Context-aware command interpretation

4. **Agent Management** (`/orchestration/types.go`)
   - Enhanced agent types with spatial and emotional capabilities
   - Advanced memory systems with resonance patterns
   - Multi-agent conversations with Deep Tree Echo coordination

3. **Workflow System** (`/orchestration/workflows.go`)
   - Multi-step workflow execution
   - Smart model routing based on task content
   - Placeholder replacement for dependent tasks

4. **API Integration** (`/api/types.go`, `/server/routes.go`)
   - RESTful API endpoints for all orchestration functions
   - Complete request/response type definitions
   - Integration with existing Ollama server architecture

5. **CLI Commands** (`/cmd/cmd.go`)
   - Command-line interface for orchestration management
   - Interactive workflow and task execution
   - Agent lifecycle management

## API Endpoints

### Agent Management
- `POST /api/orchestration/agents` - Create agent
- `GET /api/orchestration/agents` - List all agents
- `GET /api/orchestration/agents/:id` - Get specific agent
- `PUT /api/orchestration/agents/:id` - Update agent
- `DELETE /api/orchestration/agents/:id` - Delete agent

### Task Execution
- `POST /api/orchestration/tasks` - Execute orchestrated tasks
- `POST /api/orchestration/workflows` - Execute multi-step workflows

## CLI Commands

### Agent Management

Create an orchestration agent:
```bash
ollama orchestrate create-agent myagent \
  --models llama3.2,codellama,llama2 \
  --description "General purpose agent for code and chat tasks"
```

List all agents:
```bash
ollama orchestrate list-agents
```

Delete an agent:
```bash
ollama orchestrate delete-agent <agent-id>
```

### Task Execution

Run multiple tasks:
```bash
ollama orchestrate run-tasks <agent-id> \
  --tasks "generate:Write a Python function to calculate fibonacci" \
  --tasks "chat:Explain the time complexity of the fibonacci algorithm" \
  --sequential
```

Run a multi-step workflow:
```bash
ollama orchestrate run-workflow <agent-id> \
  --steps "plan:generate:Create a plan for a simple web application" \
  --steps "code:generate:Write the HTML structure based on: {{plan}}" \
  --steps "style:generate:Write CSS styles for the HTML: {{code}}"
```

## Usage Examples

### Basic Agent Creation

```bash
# Create a default agent
ollama orchestrate create-agent default \
  --models llama3.2,codellama \
  --description "Default agent for general tasks"
```

### Enhanced Agent Types (🆕 Echoself Integration)

```bash
# Create a reflective agent for self-analysis
ollama orchestrate create-agent reflective-analyst \
  --type reflective \
  --domain analysis \
  --models llama3.2 \
  --description "Self-reflective agent for performance analysis"

# Create an orchestrator agent for coordination
ollama orchestrate create-agent coordinator \
  --type orchestrator \
  --models llama3.2 \
  --tools web_search,calculator \
  --description "Orchestrator for complex multi-agent workflows"

# Create a specialist agent for coding tasks
ollama orchestrate create-agent code-specialist \
  --type specialist \
  --domain coding \
  --models codellama,llama3.2 \
  --tools web_search \
  --description "Specialist agent for software development tasks"
```

### Task Orchestration

```bash
# Run parallel code analysis tasks
ollama orchestrate run-tasks default \
  --tasks "generate:Review this Python code for bugs: def add(a,b): return a+b" \
  --tasks "generate:Suggest improvements for this code: def add(a,b): return a+b" \
  --tasks "generate:Write unit tests for: def add(a,b): return a+b"
```

### Enhanced Task Types (🆕 Echoself Integration)

```bash
# Execute tool-based tasks
ollama orchestrate run-tasks default \
  --tasks "tool:calculator:Calculate 15 * 23 + 7" \
  --tasks "tool:web_search:Find recent AI research papers"

# Run reflection tasks for agent self-analysis
ollama orchestrate run-tasks reflective-analyst \
  --tasks "reflect:Analyze recent task performance and suggest improvements"

# Execute plugin-based data analysis
ollama orchestrate run-tasks default \
  --tasks "plugin:data_analysis:Analyze sentiment of customer feedback data"
```

### Multi-Agent Conversations (🆕 Echoself Integration)

The orchestration system now supports direct agent-to-agent communication through structured conversations:

```bash
# Start a conversation between agents
ollama orchestrate start-conversation \
  --participants agent1,agent2,agent3 \
  --topic "Collaborative analysis session"

# Send messages between agents
ollama orchestrate send-message conversation-id \
  --from agent1 \
  --to agent2 \
  --content "Please analyze the current system performance" \
  --type request

# Execute structured conversation workflows
ollama orchestrate run-conversation-workflow \
  --workflow-file collaboration-workflow.json \
  --participants orchestrator,specialist,reflective

# List conversations for an agent
ollama orchestrate list-conversations agent1

# Get conversation metrics
ollama orchestrate conversation-metrics
```

#### Conversation Features:
- **Message Types**: Request, Response, Notification, Task delegation, Reflection sharing, Broadcast
- **Workflow Support**: Structured multi-step conversation patterns with templating
- **State Integration**: Messages update agent memory and context automatically
- **Task Delegation**: Agents can delegate tasks to each other through conversation
- **Metrics Tracking**: Monitor conversation patterns and agent participation

#### Example Conversation Workflow:
```json
{
  "id": "analysis-workflow",
  "name": "Collaborative Analysis",
  "participants": ["orchestrator", "specialist", "reflective"],
  "steps": [
    {
      "name": "Initial Request",
      "from": "orchestrator",
      "to": "specialist", 
      "template": "Analyze: {{task}}",
      "parameters": {"task": "system performance"}
    },
    {
      "name": "Reflection Request",
      "from": "specialist",
      "to": "reflective",
      "template": "Provide insights on: {{analysis}}"
    }
  ]
}
```

### Multi-Step Workflows

```bash
# Create a documentation workflow
ollama orchestrate run-workflow default \
  --steps "analyze:generate:Analyze this function and list its purpose: def fibonacci(n): return n if n <= 1 else fibonacci(n-1) + fibonacci(n-2)" \
  --steps "document:generate:Write comprehensive documentation for: {{analyze}}" \
  --steps "examples:generate:Provide usage examples for the function described in: {{document}}"
```

### EchoChat Shell Integration (🆕 Deep Tree Echo)

EchoChat provides intelligent shell command translation with Deep Tree Echo cognitive architecture:

#### Interactive Mode:
```bash
# Start interactive EchoChat session
go run examples/echochat_interactive.go
```

Example interactive session:
```
🌊 EchoChat - Deep Tree Echo Shell Assistant
============================================
Type 'help' for commands, 'exit' to quit
Current directory: /home/user
Shell: /bin/bash

echo> list files in current directory
ls -la
total 48
drwxr-xr-x  8 user user  4096 Sep  2 04:47 .
drwxr-xr-x 18 user user  4096 Sep  2 04:30 ..
-rw-r--r--  1 user user   126 Sep  2 04:36 .gitignore

echo> find all python files
find . -name "*.py" -type f
./examples/test_script.py
./orchestration/echoself_demo.py

echo> check disk space
df -h
Filesystem      Size  Used Avail Use% Mounted on
/dev/sda1        50G   28G   20G  59% /

echo> echo-status
🌊 Deep Tree Echo Status:
   🏥 System Health: stable
   🧠 Core Status: active
   🔄 Recursive Depth: 2
   🎯 Identity Coherence: 85.2%
```

#### Programmatic Usage:
```bash
# Run EchoChat demo with predefined commands
go run examples/echochat_demo.go
```

#### Built-in Commands:
- `help` - Show available commands
- `history` - Show command execution history
- `clear` - Clear the terminal screen
- `pwd` - Show current directory
- `cd <directory>` - Change directory
- `echo-status` - Display Deep Tree Echo system status
- `exit` or `quit` - Exit EchoChat

#### Safety Features:
- **Dangerous Command Detection**: Automatically detects potentially harmful commands
- **User Confirmation**: Prompts for confirmation before executing risky operations
- **Command Validation**: Validates commands before execution
- **History Tracking**: Maintains detailed execution history with metrics

#### Example Natural Language Commands:
- "list files" → `ls -la`
- "show running processes" → `ps aux`
- "find text files" → `find . -name "*.txt"`
- "check memory usage" → `free -h`
- "compress this folder" → `tar -czf folder.tar.gz folder/`
- "show git status" → `git status`

### Advanced Coordinated Workflows (🆕 Echoself Integration)

```bash
# Multi-agent coordinated workflow
ollama orchestrate run-coordinated-workflow coordinator \
  --tasks "analysis:reflect:Analyze current system performance" \
  --tasks "research:tool:web_search:Find best practices for optimization" \
  --tasks "planning:generate:Create improvement plan based on analysis and research" \
  --tasks "validation:plugin:data_analysis:Validate the proposed improvements"

# Hierarchical agent coordination
ollama orchestrate run-workflow coordinator \
  --steps "delegate:Generate task assignments for specialist agents" \
  --steps "coordinate:Monitor and coordinate specialist agent execution" \
  --steps "integrate:Combine results from all specialist agents" \
  --steps "reflect:Analyze coordination effectiveness and learn"
```

## Features

### Advanced Agent Types (🆕 Echoself Integration)
The orchestration engine now supports specialized agent types with enhanced capabilities:
- **General Agents**: Multi-purpose agents for standard tasks
- **Specialist Agents**: Domain-focused agents with specialized knowledge
- **Orchestrator Agents**: Coordinate other agents and manage complex workflows
- **Reflective Agents**: Self-analyzing agents that improve over time

### State Management & Memory (🆕 Echoself Integration)
Agents maintain persistent state across interactions:
- **Memory System**: Key-value storage for agent knowledge and experience
- **Context Tracking**: Relevance-scored contextual information with timestamps
- **Goal Management**: Agents can maintain and work toward defined objectives
- **Capability Awareness**: Agents track and understand their own capabilities

### Tool Integration (🆕 Echoself Integration)
Extensible tool system for external integrations:
- **Web Search Tool**: Example integration for information retrieval
- **Calculator Tool**: Mathematical computation capabilities
- **Custom Tools**: Plugin interface for adding domain-specific tools
- **Tool Selection**: Intelligent tool selection based on task requirements

### Plugin System (🆕 Echoself Integration)
Modular plugin architecture for custom functionality:
- **Data Analysis Plugin**: Example plugin for pattern recognition and insights
- **Plugin Registry**: Centralized management of available plugins
- **Custom Plugins**: Easy interface for adding specialized capabilities
- **Plugin Coordination**: Seamless integration with agent workflows

### Enhanced Coordination Patterns (🆝 Echoself Integration)
Advanced multi-agent coordination capabilities:
- **Intelligent Agent Selection**: Automatic selection of optimal agents for tasks
- **Performance Scoring**: Agent-task compatibility scoring and optimization
- **Coordination Reflection**: Post-execution analysis and learning
- **Hierarchical Orchestration**: Orchestrator agents managing sub-agents

### Smart Model Routing
The orchestration engine automatically selects the most appropriate model for each task:
- Code-related tasks are routed to code-specialized models (e.g., codellama)
- General conversation tasks use general-purpose models
- Embedding tasks use appropriate embedding models

### Workflow Dependencies
Multi-step workflows support dependency management through placeholder replacement:
- Use `{{step1}}`, `{{step2}}`, etc. to reference previous step outputs
- Use `{{stepname}}` to reference steps by name
- Automatic context passing between workflow steps

### Performance Tracking
All task executions include comprehensive metrics:
- Execution duration
- Token usage (when available)
- Model selection and routing decisions
- Success/failure status with error details

### Parallel and Sequential Execution
- **Sequential**: Tasks execute one after another, allowing dependency chains
- **Parallel**: Tasks execute simultaneously for maximum throughput
- Configurable per request based on requirements

## Configuration

### Agent Configuration
Agents support flexible configuration options:
```json
{
  "max_concurrent_tasks": 3,
  "default_model": "llama3.2",
  "timeout_seconds": 300,
  "routing_preferences": {
    "code": "codellama",
    "chat": "llama3.2"
  }
}
```

### Task Types
Supported task types (enhanced for echoself integration):
- `generate` - Text generation tasks
- `chat` - Conversational tasks
- `embed` - Embedding generation
- `tool` - External tool calling (🆕)
- `reflect` - Agent self-reflection and analysis (🆕)
- `plugin` - Custom plugin execution (🆕)
- `custom` - Extensible for future task types

## Integration with Existing Flows

The orchestration system integrates seamlessly with existing Ollama functionality:
- Reuses existing model management and scheduling
- Compatible with all existing model formats and capabilities
- Maintains existing API patterns and authentication
- Preserves existing CLI command structure

## Development and Testing

### Running Tests
```bash
go test ./orchestration/...
```

### Building with Orchestration
```bash
go build -o ollama .
```

### Development Server
```bash
# Start server with orchestration enabled
./ollama serve
```

## Current Enhancements

This implementation provides comprehensive echoself integration with Deep Tree Echo cognitive architecture:

**✅ Deep Tree Echo Integration:**
1. **Spatial Awareness**: 3D spatial context with position, orientation, and boundaries
2. **Emotional Dynamics**: Primary emotions, intensity tracking, and transition monitoring
3. **RWKV-like Reservoir Networks**: Echo State Network functions with sparse connectivity
4. **Identity Preservation**: Enhanced identity factors with spatial and emotional patterns
5. **Memory Resonance**: Advanced memory patterns with hypergraph node structures
6. **Recursive Self-Improvement**: Continuous pattern strength evolution and adaptation

**✅ EchoChat Shell Integration:**
1. **Natural Language Commands**: Convert natural language to shell commands with context
2. **Safety Validation**: Dangerous command detection with user confirmation
3. **Interactive Mode**: Full interactive shell experience with history and status
4. **Multi-OS Support**: Works on Windows, Linux, and macOS with appropriate shells
5. **Command History**: Persistent command tracking with execution metrics
6. **Built-in Commands**: Native commands for navigation and system status

**✅ Enhanced Agent Types:**
1. **Advanced Agent Behaviors**: Specialist, Orchestrator, Reflective with spatial awareness
2. **State Management**: Persistent memory with emotional and spatial context
3. **Tool Integration**: Enhanced tools with Deep Tree Echo decision making
4. **Plugin System**: Extensible architecture with cognitive integration
5. **Multi-Agent Conversations**: Direct communication with identity preservation
6. **Performance Tracking**: Advanced metrics with reservoir network monitoring

**🚧 Future Enhancements:**
1. **Advanced Learning**: Reservoir network training and adaptation
2. **Monitoring Dashboard**: Web interface with 3D spatial visualization
3. **Performance Optimization**: Resource management with emotional load balancing
4. **Enhanced Shell Features**: Code generation, file manipulation, system automation

## Error Handling

The system provides comprehensive error handling:
- Task-level error isolation (one failed task doesn't stop others)
- Detailed error messages with context
- Automatic retry capabilities (planned)
- Graceful degradation when models are unavailable

## Security Considerations

- Agent configurations are validated before execution
- Task inputs are sanitized to prevent injection attacks
- Model access is controlled through existing Ollama permissions
- Workflow execution includes timeout and resource limits

---

This orchestration system represents the core foundation for advanced AI agent coordination within echollama, providing the necessary infrastructure for complex multi-model workflows and intelligent task routing.