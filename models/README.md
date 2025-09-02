# Test Models Directory

This directory contains minimal test model configurations for CI/CD testing.

## tiny-test-model.modelfile

A minimal Modelfile that defines a test model based on qwen2.5:0.5b (the smallest available model).

- tiny-test-model.modelfile

Note: Due to size constraints, large model binary files (*.gguf) cannot be stored in git repositories.

Below are some small .gguf files kept locally for testing:

- stories260K.gguf
- tinyllamas-stories-260k-f32.gguf
- yolov3-tiny.gguf


For large models the workflow downloads the base model at runtime.

## Usage in Tests

The CI workflow will:
1. Download the base model (qwen2.5:0.5b)  
2. Create the test model using the Modelfile
3. Test direct API calls with curl
4. Verify JSON responses contain expected tokens
