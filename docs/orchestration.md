# Ollama Orchestration Agent

This document describes the orchestration agent functionality integrated into echollama, providing powerful coordination capabilities for multiple models and complex task workflows.

## Overview

The orchestration system allows you to:
- Create and manage orchestration agents that coordinate multiple models
- Execute tasks across different models with intelligent routing
- Run multi-step workflows with dependency management
- Perform parallel or sequential task execution
- Track performance metrics and execution status

## Architecture

### Core Components

1. **Orchestration Engine** (`/orchestration/engine.go`)
   - Manages agents and task execution
   - Handles parallel and sequential task processing
   - Provides performance tracking and error handling

2. **Agent Management** (`/orchestration/types.go`)
   - Create, read, update, delete orchestration agents
   - Configure agent behavior and model preferences
   - Track agent metadata and capabilities

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

### Task Orchestration

```bash
# Run parallel code analysis tasks
ollama orchestrate run-tasks default \
  --tasks "generate:Review this Python code for bugs: def add(a,b): return a+b" \
  --tasks "generate:Suggest improvements for this code: def add(a,b): return a+b" \
  --tasks "generate:Write unit tests for: def add(a,b): return a+b"
```

### Multi-Step Workflows

```bash
# Create a documentation workflow
ollama orchestrate run-workflow default \
  --steps "analyze:generate:Analyze this function and list its purpose: def fibonacci(n): return n if n <= 1 else fibonacci(n-1) + fibonacci(n-2)" \
  --steps "document:generate:Write comprehensive documentation for: {{analyze}}" \
  --steps "examples:generate:Provide usage examples for the function described in: {{document}}"
```

## Features

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
Supported task types:
- `generate` - Text generation tasks
- `chat` - Conversational tasks
- `embed` - Embedding generation
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

## Future Enhancements

This implementation provides a foundation for integration with the actual echoself system. Planned enhancements include:

1. **Advanced Agent Types**: Support for specialized agent behaviors and learning patterns
2. **State Management**: Persistent agent memory and context across sessions  
3. **Tool Integration**: Ability to call external tools and APIs within workflows
4. **Monitoring Dashboard**: Web interface for orchestration monitoring and management
5. **Performance Optimization**: Advanced scheduling and resource management
6. **Plugin System**: Extensible architecture for custom task types and behaviors

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