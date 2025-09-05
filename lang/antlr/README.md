
# ANTLR Grammars for EchoLlama

This directory contains ANTLR 4 grammar files that define the syntax for various components of the EchoLlama system.

## Grammar Files

### Modelfile.g4
Defines the grammar for Modelfile format parsing, supporting:
- FROM commands for model specification
- LICENSE, TEMPLATE, SYSTEM commands
- PARAMETER commands with key-value pairs
- MESSAGE commands with role specifications
- Comments and whitespace handling

### ThinkingParser.g4
Grammar for parsing thinking tags in model responses:
- Opening and closing `<thinking>` tags
- Content extraction between tags
- Whitespace preservation

### Grammar.g4
LLAMA grammar format parser supporting:
- Rule definitions with `::=` syntax
- Alternation with `|`
- Character classes and ranges
- Quantifiers (`?`, `*`, `+`)
- String literals and escape sequences

### JSONSchema.g4
JSON Schema parsing grammar for:
- Object and array structures
- String, number, boolean, null values
- Escape sequence handling

### API.g4
API request/response format grammar:
- Model, prompt, messages fields
- Parameter objects
- Stream boolean flags
- Template and system fields

## Usage

These grammars can be used with ANTLR 4 to generate parsers in various target languages. The grammars are designed to match the existing Go parser implementations in the codebase.

## Building

To generate parsers from these grammars:

```bash
# Install ANTLR 4
# Generate Go code
antlr4 -Dlanguage=Go -o generated/ *.g4
```

## Compatibility

These grammars are designed to be compatible with the existing parser implementations in:
- `parser/parser.go` - Modelfile parsing
- `thinking/parser.go` - Thinking tag parsing
- LLAMA grammar parsing in the C++ components
