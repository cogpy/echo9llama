# Test Models Directory

This directory contains minimal test model configurations for CI/CD testing.

## tiny-test-model.modelfile

A minimal Modelfile that defines a test model based on qwen2.5:0.5b (the smallest available model).

Note: Due to size constraints, actual model binary files (*.gguf) cannot be stored in git repositories. The workflow downloads the base model at runtime.

## Usage in Tests

The CI workflow will:
1. Download the base model (qwen2.5:0.5b)  
2. Create the test model using the Modelfile
3. Test direct API calls with curl
4. Verify JSON responses contain expected tokens