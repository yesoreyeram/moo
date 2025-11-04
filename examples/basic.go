// Example demonstrates basic usage of the moo lexer in Go.
package main

import (
	"fmt"
	"regexp"

	"github.com/yesoreyeram/moo"
)

func main() {
	// Example 1: Basic tokenization
	fmt.Println("=== Example 1: Basic Tokenization ===")
	basicExample()

	// Example 2: Keywords
	fmt.Println("\n=== Example 2: Keywords ===")
	keywordsExample()

	// Example 3: Line numbers
	fmt.Println("\n=== Example 3: Line Numbers ===")
	lineNumbersExample()

	// Example 4: Value transforms
	fmt.Println("\n=== Example 4: Value Transforms ===")
	valueTransformExample()

	// Example 5: Stateful lexing
	fmt.Println("\n=== Example 5: Stateful Lexing ===")
	statefulExample()
}

func basicExample() {
	spec := map[string]interface{}{
		"WS":     regexp.MustCompile(`[ \t]+`),
		"number": regexp.MustCompile(`[0-9]+`),
		"word":   regexp.MustCompile(`[a-z]+`),
		"lparen": "(",
		"rparen": ")",
	}

	lexer, err := moo.Compile(spec)
	if err != nil {
		panic(err)
	}

	lexer.Reset("hello (123) world")

	for {
		token := lexer.Next()
		if token == nil {
			break
		}
		fmt.Printf("  Type: %-8s Value: %s\n", token.Type, token.Value)
	}
}

func keywordsExample() {
	// Define keywords
	keywordMap := map[string]interface{}{
		"KW": []string{"if", "while", "else", "for"},
	}

	spec := map[string]interface{}{
		"identifier": map[string]interface{}{
			"match": regexp.MustCompile(`[a-z]+`),
			"type":  moo.KeywordTransform(keywordMap),
		},
		"WS": regexp.MustCompile(`[ \t]+`),
	}

	lexer, _ := moo.Compile(spec)
	lexer.Reset("if foo while bar")

	for {
		token := lexer.Next()
		if token == nil {
			break
		}
		if token.Type != "WS" {
			fmt.Printf("  Type: %-12s Value: %s\n", token.Type, token.Value)
		}
	}
}

func lineNumbersExample() {
	spec := map[string]interface{}{
		"word": regexp.MustCompile(`[a-z]+`),
		"NL": map[string]interface{}{
			"match":      regexp.MustCompile(`\n`),
			"lineBreaks": true,
		},
		"WS": regexp.MustCompile(`[ \t]+`),
	}

	lexer, _ := moo.Compile(spec)
	lexer.Reset("hello\nworld\nfoo")

	for {
		token := lexer.Next()
		if token == nil {
			break
		}
		if token.Type != "WS" && token.Type != "NL" {
			fmt.Printf("  Line %d, Col %d: %s\n", token.Line, token.Col, token.Value)
		}
	}
}

func valueTransformExample() {
	spec := map[string]interface{}{
		"string": map[string]interface{}{
			"match": regexp.MustCompile(`"[^"]*"`),
			"value": func(s string) string {
				// Remove quotes
				return s[1 : len(s)-1]
			},
		},
		"number": map[string]interface{}{
			"match": regexp.MustCompile(`[0-9]+`),
			"value": func(s string) string {
				return "NUM:" + s
			},
		},
		"WS": regexp.MustCompile(`[ \t]+`),
	}

	lexer, _ := moo.Compile(spec)
	lexer.Reset(`"hello" 123 "world"`)

	for {
		token := lexer.Next()
		if token == nil {
			break
		}
		if token.Type != "WS" {
			fmt.Printf("  Text: %-10s -> Value: %s\n", token.Text, token.Value)
		}
	}
}

func statefulExample() {
	// Tokenize nested braces
	stateSpec := map[string]map[string]interface{}{
		"main": {
			"word": regexp.MustCompile(`[a-z]+`),
			"lbrace": map[string]interface{}{
				"match": "{",
				"push":  "braced",
			},
			"WS": regexp.MustCompile(`[ \t]+`),
		},
		"braced": {
			"word":  regexp.MustCompile(`[a-z]+`),
			"colon": ":",
			"rbrace": map[string]interface{}{
				"match": "}",
				"pop":   true,
			},
			"WS": regexp.MustCompile(`[ \t]+`),
		},
	}

	lexer, _ := moo.States(stateSpec, "main")
	lexer.Reset("hello {foo: bar} world")

	for {
		token := lexer.Next()
		if token == nil {
			break
		}
		if token.Type != "WS" {
			fmt.Printf("  Type: %-10s Value: %s\n", token.Type, token.Value)
		}
	}
}
