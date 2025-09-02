# Live Interactive Testing with REAL Models - NO FALLBACKS

This directory contains testing tools for EchoLlama that work EXCLUSIVELY with real language models. There are NO FALLBACKS, NO MOCKS, and NO OFFLINE MODES. Either the real model works completely or everything fails - no exceptions!

## Critical Requirements

⚠️ **REAL MODEL MUST WORK OR TOTAL FAILURE** ⚠️

- Real language model must be available and functional
- Ollama server must be running and responsive  
- Model must successfully respond to all test queries
- Deep Tree Echo cognitive architecture must initialize with real model
- EchoChat must translate commands using real model inference
- All orchestration must work with actual model processing

## Files

### `live-interactive-test.go`
Comprehensive automated test that validates ALL EchoLlama functionality with REAL model:
- Basic model connectivity and communication - MUST WORK
- Deep Tree Echo cognitive system with real model inference - MUST WORK
- EchoChat integration with actual natural language processing - MUST WORK  
- Real shell command translation using model - MUST WORK
- Orchestration capabilities with live model processing - MUST WORK

**Usage:**
```bash
go run scripts/live-interactive-test.go <model-name>
```

**Features:**
- ZERO tolerance for failures - any error causes immediate exit
- Real model inference required for all functionality
- No safe mode - all commands use actual model processing
- Comprehensive error reporting on any failure
- Hard timeouts with fatal errors

### `simple-demo.go`
Simplified demonstration script with REAL model requirements:
- Basic model conversation - MUST RESPOND
- EchoChat system integration - MUST WORK WITH REAL MODEL
- Zero fallback behavior - fails immediately on any error

**Usage:**
```bash
go run scripts/simple-demo.go <model-name>
```

## GitHub Action Workflow

The `.github/workflows/live-interactive-test.yml` workflow provides REAL model testing in CI:

### Triggers
- Push to main/develop branches
- Pull requests to main/develop  
- Manual workflow dispatch (allows custom model selection)

### Features  
- **Mandatory Model Success**: Downloads model and FAILS if unsuccessful
- **Server Management**: Starts ollama server - FAILS if server doesn't start
- **REAL Model Testing**: All tests use actual model inference - FAILS on any error
- **Orchestration Testing**: Validates agent creation with real model - FAILS if unsuccessful
- **Performance Monitoring**: Health checks with real model - FAILS if unhealthy
- **Permanent Storage**: Commits model metadata to repository
- **Zero Tolerance**: Any failure causes immediate workflow failure

### Manual Execution
Manually trigger the workflow with custom parameters:
1. Go to the Actions tab in GitHub
2. Select "Live Interactive Test with REAL Model - NO FALLBACKS"  
3. Click "Run workflow"
4. Choose:
   - **Model**: Which model to test with (default: `llama3.2:1b`)
   - **Test Mode**: `quick` or `full` (affects timeouts only)

### Supported Models
The workflow works with any model supported by Ollama:
- `llama3.2:1b` (default, ~1.3GB)
- `qwen2.5:0.5b` (~0.8GB) 
- `gemma2:2b` (~1.6GB)
- Any other Ollama-supported model

## REAL Model Testing ONLY

### Requirements for ALL Tests
- Ollama server MUST be running and responsive
- Specified model MUST be available and loaded
- Model MUST respond to test queries within timeout
- Deep Tree Echo MUST initialize successfully with real model
- EchoChat MUST process commands using real model inference
- Orchestration MUST work with actual model processing

### What Gets Tested
- ✅ **Real Model Conversations**: Actual model responses required
- ✅ **Deep Tree Echo Integration**: Must initialize and work with real model  
- ✅ **EchoChat Processing**: Real natural language to command translation
- ✅ **Orchestration Engine**: Live agent creation and task execution
- ✅ **Performance Metrics**: Real model response times and health
- ✅ **System Integration**: All components working together with real model

### Failure Conditions
- ❌ Model not available or unresponsive
- ❌ Server not starting or unhealthy
- ❌ Deep Tree Echo initialization failure
- ❌ EchoChat unable to process commands
- ❌ Orchestration system not functional
- ❌ Any timeout or processing error

## Model Storage

Models are permanently stored in the repository:
- Models directory: `models/`
- Model metadata tracked in `models/test-model.txt`
- Last update timestamp in `models/last-updated.txt`
- Model files allowed in repository via `.gitignore` exceptions

## Test Coverage

The system validates REAL functionality:
- ✅ Actual model conversations and intelligent responses
- ✅ Deep Tree Echo cognitive system operational with real model
- ✅ EchoChat natural language processing using model inference
- ✅ Orchestration agent creation and execution with real model
- ✅ Performance monitoring of actual model operations
- ✅ System integration with live model processing

## Troubleshooting

**All issues result in test failure - no graceful degradation:**

1. **Server Won't Start**: FATAL - Check ollama installation
2. **Model Not Available**: FATAL - Check model name and network  
3. **Model Unresponsive**: FATAL - Check model compatibility and resources
4. **Deep Tree Echo Failure**: FATAL - Check model and system integration
5. **EchoChat Processing Failure**: FATAL - Check model inference capabilities
6. **Orchestration Failure**: FATAL - Check model and orchestration system

**No Recovery Actions - Fix the underlying issue or test fails.**
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