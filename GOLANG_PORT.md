# Go Port of Moo - Implementation Summary

This document summarizes the Go port of the moo tokenizer/lexer generator.

## Overview

The Go port (`moo.go`) is a complete implementation of the moo lexer with the same core functionality as the JavaScript version. It provides a fast, efficient tokenizer for parsing text based on regular expressions.

## Implementation Details

### Core Components

1. **Token Struct** - Represents a lexed token with:
   - Type, Value, Text
   - Position tracking (Offset, Line, Col)
   - LineBreaks count

2. **Lexer** - Main lexer struct with:
   - State management
   - Position tracking
   - Token matching and generation

3. **Rule Compilation** - Converts user-defined rules into optimized regular expressions

### API

- `Compile(spec)` - Creates a stateless lexer
- `States(stateSpec, start)` - Creates a stateful lexer with multiple states
- `KeywordTransform(keywords)` - Helper for keyword matching
- `Lexer.Next()` - Returns next token
- `Lexer.Reset(data)` - Resets lexer with new input
- `Lexer.Save()` - Saves current state
- `Lexer.FormatError(token, message)` - Formats error messages

## Features

✅ Regular expression-based token matching
✅ String literal tokens
✅ Keyword support (arrays and transform functions)
✅ Line and column tracking
✅ Stateful lexing (push/pop/next)
✅ Value transforms
✅ Type transforms
✅ Error handling
✅ Fast single-character matching

## Testing

- **12 Go tests** with 82.1% code coverage
- **92 JavaScript tests** continue to pass
- **2 working examples** (basic and JSON)

## Examples

### Basic Usage

```go
spec := map[string]interface{}{
    "number": regexp.MustCompile(`[0-9]+`),
    "word":   regexp.MustCompile(`[a-z]+`),
    "lparen": "(",
}

lexer, _ := moo.Compile(spec)
lexer.Reset("hello (123)")

for {
    token := lexer.Next()
    if token == nil {
        break
    }
    fmt.Printf("%s: %s\n", token.Type, token.Value)
}
```

### Stateful Lexing

```go
stateSpec := map[string]map[string]interface{}{
    "main": {
        "word": regexp.MustCompile(`[a-z]+`),
        "lbrace": map[string]interface{}{
            "match": "{",
            "push":  "braced",
        },
    },
    "braced": {
        "word": regexp.MustCompile(`[a-z]+`),
        "rbrace": map[string]interface{}{
            "match": "}",
            "pop":   true,
        },
    },
}

lexer, _ := moo.States(stateSpec, "main")
```

## Differences from JavaScript

1. **Type System** - Uses Go's static types vs JavaScript's dynamic types
2. **Regular Expressions** - Uses Go's `regexp` package
3. **Error Handling** - Returns `error` values where appropriate
4. **Iteration** - Uses `Next()` method instead of JavaScript iterators
5. **Rule Specification** - Uses `map[string]interface{}` for flexibility

## Files

- `moo.go` - Main implementation
- `moo_test.go` - Comprehensive tests
- `README_GO.md` - Go-specific documentation
- `examples/basic.go` - Basic usage examples
- `examples/json.go` - JSON tokenization example
- `go.mod` - Go module definition

## Performance

The Go implementation maintains the same optimization strategies as the JavaScript version:

- Single compiled regexp for all rules
- Fast path for single-character tokens
- Efficient state management
- Minimal allocations in hot paths

## Security

✅ CodeQL analysis: No vulnerabilities found
✅ Code review: No issues found
✅ All tests passing

## Compatibility

The Go port maintains API compatibility with the JavaScript version where appropriate, making it easy for developers familiar with the JavaScript version to use the Go version.
