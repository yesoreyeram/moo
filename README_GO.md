# Moo - Go Port

![Moo](cow.png)

Moo is a highly-optimized tokenizer/lexer generator. This is a Go port of the [JavaScript moo library](https://github.com/tjvr/moo).

## Features

* Fast - optimized regular expression compilation
* Convenient - simple API
* Uses Regular Expressions for token matching
* Tracks Line Numbers and Column positions
* Handles Keywords elegantly
* Supports Stateful lexing
* Custom Error handling
* No external dependencies

## Installation

```bash
go get github.com/yesoreyeram/moo
```

## Quick Start

```go
package main

import (
    "fmt"
    "regexp"
    
    "github.com/yesoreyeram/moo"
)

func main() {
    // Define your token rules
    spec := map[string]interface{}{
        "WS":      regexp.MustCompile(`[ \t]+`),
        "comment": regexp.MustCompile(`//.*`),
        "number":  regexp.MustCompile(`[0-9]+`),
        "string":  regexp.MustCompile(`"[^"]*"`),
        "lparen":  "(",
        "rparen":  ")",
        "keyword": []string{"while", "if", "else"},
        "NL": map[string]interface{}{
            "match":      regexp.MustCompile(`\n`),
            "lineBreaks": true,
        },
    }
    
    // Compile the lexer
    lexer, err := moo.Compile(spec)
    if err != nil {
        panic(err)
    }
    
    // Tokenize some input
    lexer.Reset("while (10)\n// comment")
    
    for {
        token := lexer.Next()
        if token == nil {
            break
        }
        fmt.Printf("Type: %s, Value: %s, Line: %d, Col: %d\n", 
            token.Type, token.Value, token.Line, token.Col)
    }
}
```

## Usage

### Basic Lexer

Define your tokens using regular expressions or string literals:

```go
spec := map[string]interface{}{
    "word":   regexp.MustCompile(`[a-z]+`),
    "number": regexp.MustCompile(`[0-9]+`),
    "lparen": "(",
    "rparen": ")",
}

lexer, _ := moo.Compile(spec)
lexer.Reset("hello (123)")

token := lexer.Next() // Type: "word", Value: "hello"
```

### Keywords

Handle keywords using arrays or the keyword transform function:

```go
// Simple array of keywords
spec := map[string]interface{}{
    "keyword": []string{"if", "while", "else"},
    "word":    regexp.MustCompile(`[a-z]+`),
}

// Or use keyword transform for more control
keywordMap := map[string]interface{}{
    "KW_IF":    "if",
    "KW_WHILE": "while",
}

spec := map[string]interface{}{
    "identifier": map[string]interface{}{
        "match": regexp.MustCompile(`[a-z]+`),
        "type":  moo.KeywordTransform(keywordMap),
    },
}
```

### Line Numbers

Track line numbers by marking rules that can contain newlines:

```go
spec := map[string]interface{}{
    "word": regexp.MustCompile(`[a-z]+`),
    "NL": map[string]interface{}{
        "match":      regexp.MustCompile(`\n`),
        "lineBreaks": true,
    },
}

lexer, _ := moo.Compile(spec)
lexer.Reset("hello\nworld")

token := lexer.Next() // Line: 1, Col: 1
token = lexer.Next()  // newline
token = lexer.Next()  // Line: 2, Col: 1
```

### Value Transforms

Transform matched values:

```go
spec := map[string]interface{}{
    "string": map[string]interface{}{
        "match": regexp.MustCompile(`"[^"]*"`),
        "value": func(s string) string {
            return s[1 : len(s)-1] // Remove quotes
        },
    },
}

lexer, _ := moo.Compile(spec)
lexer.Reset(`"hello"`)

token := lexer.Next()
// token.Text == `"hello"`
// token.Value == "hello"
```

### Stateful Lexing

Support different lexer states for complex grammars:

```go
stateSpec := map[string]map[string]interface{}{
    "main": {
        "word":   regexp.MustCompile(`[a-z]+`),
        "lbrace": map[string]interface{}{
            "match": "{",
            "push":  "braced",
        },
    },
    "braced": {
        "word":   regexp.MustCompile(`[a-z]+`),
        "rbrace": map[string]interface{}{
            "match": "}",
            "pop":   true,
        },
    },
}

lexer, _ := moo.States(stateSpec, "main")
lexer.Reset("hello {world}")
```

## Token Object

Each token returned by `Next()` contains:

* `Type` - The token type name
* `Value` - The token value (transformed if a value function was provided)
* `Text` - The original matched text
* `Offset` - Byte offset from the start of the buffer
* `LineBreaks` - Number of line breaks in the token
* `Line` - Line number (1-indexed)
* `Col` - Column number (1-indexed)

## Differences from JavaScript Version

The Go port maintains the same core functionality but with some adaptations for Go:

1. **Type System**: Uses Go's type system instead of JavaScript's dynamic typing
2. **Regular Expressions**: Uses Go's `regexp` package instead of JavaScript RegExp
3. **Error Handling**: Returns `error` values instead of throwing exceptions (except for shouldThrow errors)
4. **Iterator**: Uses `Next()` method instead of JavaScript iterators
5. **Map Types**: Uses `map[string]interface{}` for flexible rule specifications

## Testing

```bash
go test -v
```

## License

BSD-3-Clause (same as the original JavaScript version)

## Credits

This is a Go port of [moo](https://github.com/tjvr/moo) by Tim Radvan.
