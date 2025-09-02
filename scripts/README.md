# Live Interactive Testing with Real Models

This directory contains testing tools for EchoLlama that work with real language models to validate the full interactive capabilities including Deep Tree Echo cognitive architecture and EchoChat shell integration.

## Files

### `live-interactive-test.go`
Comprehensive automated test that validates all EchoLlama functionality:
- Basic model connectivity and communication
- Deep Tree Echo cognitive system initialization and status
- EchoChat integration with natural language to shell command translation
- Orchestration capabilities and agent management
- Graceful fallback for offline/CI environments

**Usage:**
```bash
go run scripts/live-interactive-test.go <model-name>
```

**Features:**
- Works both online (with running ollama server) and offline (structure validation only)
- Comprehensive error handling and reporting
- Timeout management for CI environments
- Detailed logging and status reporting

### `simple-demo.go`
Simplified demonstration script that shows core functionality:
- Basic model conversation
- EchoChat system integration
- Quick validation of core components

**Usage:**
```bash
go run scripts/simple-demo.go <model-name>
```

## GitHub Action Workflow

The `.github/workflows/live-interactive-test.yml` workflow provides automated testing in CI:

### Triggers
- Push to main/develop branches
- Pull requests to main/develop
- Manual workflow dispatch (allows custom model selection)

### Features
- **Automatic Model Pulling**: Downloads and validates the specified model (default: `llama3.2:1b`)
- **Server Management**: Starts/stops ollama server with proper lifecycle management
- **Comprehensive Testing**: Runs both simple demo and full test suite
- **Orchestration Testing**: Validates agent creation and task execution
- **Performance Monitoring**: Health checks and responsiveness testing
- **Error Handling**: Detailed logging and artifact upload on failure
- **Cleanup**: Proper resource cleanup regardless of test outcome

### Manual Execution
You can manually trigger the workflow with custom parameters:
1. Go to the Actions tab in GitHub
2. Select "Live Interactive Test with Real Model"
3. Click "Run workflow"
4. Choose:
   - **Model**: Which model to test with (default: `llama3.2:1b`)
   - **Test Mode**: `quick` or `full` (affects timeout and verbosity)

### Supported Models
The workflow is designed to work with small, fast models suitable for CI:
- `llama3.2:1b` (default, ~1.3GB)
- `qwen2.5:0.5b` (~0.8GB)
- `gemma2:2b` (~1.6GB)
- Any other model supported by Ollama

## Testing Scenarios

### 1. Online Testing (with running ollama server)
When an ollama server is running:
- Tests actual model responses and conversations
- Validates real command interpretation and execution (in safe mode)
- Exercises full Deep Tree Echo cognitive processing
- Tests actual orchestration with model inference

### 2. Offline Testing (without server)
When no server is available:
- Validates code structure and initialization
- Tests Deep Tree Echo system status and configuration
- Verifies EchoChat component creation and basic functionality
- Validates orchestration engine initialization

## Deep Tree Echo Features Tested

### Core Cognitive Architecture
- ✅ System initialization and health monitoring
- ✅ Identity coherence tracking
- ✅ Memory resonance pattern analysis
- ✅ Recursive self-improvement capabilities
- ✅ Cross-system synthesis integration

### EchoChat Integration
- ✅ Natural language to shell command translation
- ✅ Command safety validation and user confirmation
- ✅ Execution history and performance tracking
- ✅ Built-in command handling (help, status, navigation)
- ✅ Multi-OS shell support (Windows cmd, Unix bash)

### Orchestration Capabilities
- ✅ Agent creation and management
- ✅ Task delegation and execution
- ✅ Multi-agent conversations
- ✅ Workflow orchestration
- ✅ Performance optimization and load balancing

## Example Output

```
🧪 Starting Live Interactive Test with model: llama3.2:1b
===================================================

🔗 Test 1: Basic Model Connectivity
   📝 Model response: Hello! I'm working correctly and ready to help.
✅ Model connectivity test passed

🧠 Test 2: Deep Tree Echo System
   🧠 Initializing Deep Tree Echo...
   🏥 System Health: stable
   🧠 Core Status: active
✅ Deep Tree Echo test passed

🌊 Test 3: EchoChat Integration
   🔍 Testing command translation...
   📝 Testing: 'show current directory'
   ✅ Command processed successfully
✅ EchoChat integration test passed

🗣️  Test 4: Natural Language Shell Commands
   🔄 Processing: show current working directory
   ✅ Command processed successfully
✅ Shell commands test passed

⚙️  Test 5: Orchestration Capabilities
   🧪 Testing orchestration command: 'get system status'
   ✅ Orchestration command processed successfully
✅ Orchestration capabilities test passed

🎉 All tests completed! Live interactive test with llama3.2:1b finished.
```

## Contributing

When adding new features to EchoLlama's interactive capabilities:

1. Add corresponding tests to `live-interactive-test.go`
2. Update the simple demo if the feature is user-facing
3. Consider both online and offline test scenarios
4. Update this documentation

## Troubleshooting

### Common Issues

**"Connection refused" errors**: The ollama server isn't running. Start it with:
```bash
./ollama serve
```

**"Model not found" errors**: Pull the model first:
```bash
./ollama pull llama3.2:1b
```

**Timeout errors in CI**: Consider using a smaller/faster model or adjusting timeouts in the workflow.

**Deep Tree Echo initialization warnings**: These are usually non-fatal and the system can continue with reduced functionality.