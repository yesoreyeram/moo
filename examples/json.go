// Example demonstrates JSON-like lexing with moo.
package main

import (
	"fmt"
	"regexp"

	"github.com/yesoreyeram/moo"
)

func main() {
	// Create a lexer for a simple JSON-like language
	spec := map[string]interface{}{
		"WS": map[string]interface{}{
			"match":      regexp.MustCompile(`[ \t\n\r]+`),
			"lineBreaks": true,
		},
		"comment": regexp.MustCompile(`//[^\n]*`),
		"number":  regexp.MustCompile(`-?(?:0|[1-9][0-9]*)(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?`),
		"string": map[string]interface{}{
			"match": regexp.MustCompile(`"(?:\\["\\\/bfnrt]|\\u[0-9a-fA-F]{4}|[^"\\])*"`),
			"value": func(s string) string {
				// Remove quotes for the value
				return s[1 : len(s)-1]
			},
		},
		"true":   "true",
		"false":  "false",
		"null":   "null",
		"lbrace": "{",
		"rbrace": "}",
		"lbrack": "[",
		"rbrack": "]",
		"comma":  ",",
		"colon":  ":",
	}

	lexer, err := moo.Compile(spec)
	if err != nil {
		panic(err)
	}

	// JSON-like input
	input := `{
  "name": "John Doe",
  "age": 30,
  "active": true,
  "scores": [95, 87.5, 92],
  "metadata": null
}`

	fmt.Println("Input:")
	fmt.Println(input)
	fmt.Println("\nTokens:")

	lexer.Reset(input)

	tokenCount := 0
	for {
		token := lexer.Next()
		if token == nil {
			break
		}

		// Skip whitespace and comments for cleaner output
		if token.Type == "WS" || token.Type == "comment" {
			continue
		}

		tokenCount++
		fmt.Printf("%3d. %-10s | Value: %-20s | Line: %d, Col: %d\n",
			tokenCount,
			token.Type,
			fmt.Sprintf("%q", token.Value),
			token.Line,
			token.Col,
		)
	}

	fmt.Printf("\nTotal tokens (excluding whitespace): %d\n", tokenCount)
}
