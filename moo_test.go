package moo

import (
	"regexp"
	"testing"
)

func TestBasicLexing(t *testing.T) {
	spec := map[string]interface{}{
		"WS":     regexp.MustCompile(`[ \t]+`),
		"number": regexp.MustCompile(`[0-9]+`),
		"word":   regexp.MustCompile(`[a-z]+`),
		"lparen": "(",
		"rparen": ")",
	}
	
	lexer, err := Compile(spec)
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}
	
	lexer.Reset("hello (123) world")
	
	tests := []struct {
		expectedType  string
		expectedValue string
	}{
		{"word", "hello"},
		{"WS", " "},
		{"lparen", "("},
		{"number", "123"},
		{"rparen", ")"},
		{"WS", " "},
		{"word", "world"},
	}
	
	for i, test := range tests {
		token := lexer.Next()
		if token == nil {
			t.Fatalf("Expected token at position %d, got nil", i)
		}
		
		if token.Type != test.expectedType {
			t.Errorf("Token %d: expected type %q, got %q", i, test.expectedType, token.Type)
		}
		
		if token.Value != test.expectedValue {
			t.Errorf("Token %d: expected value %q, got %q", i, test.expectedValue, token.Value)
		}
	}
	
	// Should be EOF
	if token := lexer.Next(); token != nil {
		t.Errorf("Expected EOF, got token: %v", token)
	}
}

func TestLineNumbers(t *testing.T) {
	spec := map[string]interface{}{
		"word": regexp.MustCompile(`[a-z]+`),
		"NL": map[string]interface{}{
			"match":      regexp.MustCompile(`\n`),
			"lineBreaks": true,
		},
		"WS": regexp.MustCompile(`[ \t]+`),
	}
	
	lexer, err := Compile(spec)
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}
	
	lexer.Reset("hello\nworld")
	
	token := lexer.Next()
	if token.Line != 1 || token.Col != 1 {
		t.Errorf("First token: expected line 1 col 1, got line %d col %d", token.Line, token.Col)
	}
	
	lexer.Next() // newline
	
	token = lexer.Next()
	if token.Line != 2 {
		t.Errorf("Third token: expected line 2, got line %d", token.Line)
	}
}

func TestStringLiterals(t *testing.T) {
	spec := map[string]interface{}{
		"lparen": "(",
		"rparen": ")",
		"word":   regexp.MustCompile(`[a-z]+`),
	}
	
	lexer, err := Compile(spec)
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}
	
	lexer.Reset("(test)")
	
	tests := []string{"lparen", "word", "rparen"}
	
	for i, expectedType := range tests {
		token := lexer.Next()
		if token == nil {
			t.Fatalf("Expected token at position %d, got nil", i)
		}
		
		if token.Type != expectedType {
			t.Errorf("Token %d: expected type %q, got %q", i, expectedType, token.Type)
		}
	}
}

func TestKeywords(t *testing.T) {
	keywordMap := map[string]interface{}{
		"keyword": []string{"if", "while", "else"},
	}
	
	spec := map[string]interface{}{
		"identifier": map[string]interface{}{
			"match": regexp.MustCompile(`[a-z]+`),
			"type":  KeywordTransform(keywordMap),
		},
		"WS": regexp.MustCompile(`[ \t]+`),
	}
	
	lexer, err := Compile(spec)
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}
	
	lexer.Reset("if foo while")
	
	tests := []struct {
		expectedType  string
		expectedValue string
	}{
		{"keyword", "if"},
		{"WS", " "},
		{"identifier", "foo"},
		{"WS", " "},
		{"keyword", "while"},
	}
	
	for i, test := range tests {
		token := lexer.Next()
		if token == nil {
			t.Fatalf("Expected token at position %d, got nil", i)
		}
		
		if token.Type != test.expectedType {
			t.Errorf("Token %d: expected type %q, got %q", i, test.expectedType, token.Type)
		}
		
		if token.Value != test.expectedValue {
			t.Errorf("Token %d: expected value %q, got %q", i, test.expectedValue, token.Value)
		}
	}
}

func TestArrayOfKeywords(t *testing.T) {
	spec := map[string]interface{}{
		"keyword": []string{"if", "while", "else"},
		"word":    regexp.MustCompile(`[a-z]+`),
		"WS":      regexp.MustCompile(`[ \t]+`),
	}
	
	lexer, err := Compile(spec)
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}
	
	lexer.Reset("if foo")
	
	token := lexer.Next()
	if token.Type != "keyword" || token.Value != "if" {
		t.Errorf("Expected keyword 'if', got %s '%s'", token.Type, token.Value)
	}
	
	lexer.Next() // whitespace
	
	token = lexer.Next()
	if token.Type != "word" || token.Value != "foo" {
		t.Errorf("Expected word 'foo', got %s '%s'", token.Type, token.Value)
	}
}

func TestValueTransform(t *testing.T) {
	spec := map[string]interface{}{
		"string": map[string]interface{}{
			"match": regexp.MustCompile(`"[^"]*"`),
			"value": func(s string) string {
				return s[1 : len(s)-1] // Remove quotes
			},
		},
	}
	
	lexer, err := Compile(spec)
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}
	
	lexer.Reset(`"hello"`)
	
	token := lexer.Next()
	if token.Text != `"hello"` {
		t.Errorf("Expected text %q, got %q", `"hello"`, token.Text)
	}
	
	if token.Value != "hello" {
		t.Errorf("Expected value %q, got %q", "hello", token.Value)
	}
}

func TestStates(t *testing.T) {
	stateSpec := map[string]map[string]interface{}{
		"main": {
			"word":   regexp.MustCompile(`[a-z]+`),
			"lbrace": map[string]interface{}{
				"match": "{",
				"push":  "braced",
			},
			"WS": regexp.MustCompile(`[ \t]+`),
		},
		"braced": {
			"word":   regexp.MustCompile(`[a-z]+`),
			"rbrace": map[string]interface{}{
				"match": "}",
				"pop":   true,
			},
			"WS": regexp.MustCompile(`[ \t]+`),
		},
	}
	
	lexer, err := States(stateSpec, "main")
	if err != nil {
		t.Fatalf("Failed to compile states: %v", err)
	}
	
	lexer.Reset("hello {world}")
	
	tests := []struct {
		expectedType  string
		expectedValue string
	}{
		{"word", "hello"},
		{"WS", " "},
		{"lbrace", "{"},
		{"word", "world"},
		{"rbrace", "}"},
	}
	
	for i, test := range tests {
		token := lexer.Next()
		if token == nil {
			t.Fatalf("Expected token at position %d, got nil", i)
		}
		
		if token.Type != test.expectedType {
			t.Errorf("Token %d: expected type %q, got %q", i, test.expectedType, token.Type)
		}
		
		if token.Value != test.expectedValue {
			t.Errorf("Token %d: expected value %q, got %q", i, test.expectedValue, token.Value)
		}
	}
}
