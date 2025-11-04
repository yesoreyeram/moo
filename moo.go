// Package moo provides a highly-optimized tokenizer/lexer generator.
// It's a Go port of the JavaScript moo library.
package moo

import (
	"fmt"
	"regexp"
	"strings"
)

// Token represents a lexed token with position information.
type Token struct {
	Type       string // Token type/name
	Value      string // Token value (may be transformed)
	Text       string // Original matched text
	Offset     int    // Byte offset from start of buffer
	LineBreaks int    // Number of line breaks in token
	Line       int    // Line number (1-indexed)
	Col        int    // Column number (1-indexed)
}

// String returns the token's value.
func (t *Token) String() string {
	return t.Value
}

// Rule represents a token matching rule.
type rule struct {
	defaultType string
	match       []*regexp.Regexp
	lineBreaks  bool
	value       func(string) string
	typeFunc    func(string) string
	shouldThrow bool
	error       bool
	fallback    bool
	next        string
	push        string
	pop         bool
}

// compiledRules holds the compiled regexp and associated rules.
type compiledRules struct {
	regexp *regexp.Regexp
	groups []*rule
	fast   map[byte]*rule
	error  *rule
}

// Lexer represents a stateful lexer.
type Lexer struct {
	startState   string
	states       map[string]*compiledRules
	buffer       string
	index        int
	line         int
	col          int
	state        string
	stack        []string
	groups       []*rule
	re           *regexp.Regexp
	fast         map[byte]*rule
	errorRule    *rule
	queuedGroup  *rule
	queuedText   string
}

// KeywordTransform creates a type transform function for keywords.
// It returns a function that maps keyword strings to their token types.
func KeywordTransform(keywords map[string]interface{}) func(string) string {
	keywordMap := make(map[string]string)
	
	for tokenType, v := range keywords {
		switch kw := v.(type) {
		case string:
			keywordMap[kw] = tokenType
		case []string:
			for _, k := range kw {
				keywordMap[k] = tokenType
			}
		}
	}
	
	return func(text string) string {
		if t, ok := keywordMap[text]; ok {
			return t
		}
		return ""
	}
}

// Error is a special rule marker for error tokens.
var Error = map[string]interface{}{"error": true}

// Fallback is a special rule marker for fallback tokens.
var Fallback = map[string]interface{}{"fallback": true}

// Compile compiles a set of token rules into a Lexer.
func Compile(spec map[string]interface{}) (*Lexer, error) {
	rules, err := objectToRules(spec)
	if err != nil {
		return nil, err
	}
	
	compiled, err := compileRules(rules, false)
	if err != nil {
		return nil, err
	}
	
	states := map[string]*compiledRules{
		"start": compiled,
	}
	
	return &Lexer{
		startState: "start",
		states:     states,
		buffer:     "",
		stack:      []string{},
		line:       1,
		col:        1,
	}, nil
}

// States compiles a set of states with their respective rules.
func States(stateSpec map[string]map[string]interface{}, start string) (*Lexer, error) {
	if start == "" {
		// Use first key as start state
		for k := range stateSpec {
			start = k
			break
		}
	}
	
	compiledStates := make(map[string]*compiledRules)
	
	for stateName, spec := range stateSpec {
		rules, err := objectToRules(spec)
		if err != nil {
			return nil, err
		}
		
		compiled, err := compileRules(rules, true)
		if err != nil {
			return nil, err
		}
		
		compiledStates[stateName] = compiled
	}
	
	return &Lexer{
		startState: start,
		states:     compiledStates,
		buffer:     "",
		stack:      []string{},
		line:       1,
		col:        1,
	}, nil
}

// Reset resets the lexer with new input.
func (l *Lexer) Reset(data string) {
	l.buffer = data
	l.index = 0
	l.line = 1
	l.col = 1
	l.queuedGroup = nil
	l.queuedText = ""
	l.setState(l.startState)
	l.stack = []string{}
}

// setState changes the current lexer state.
func (l *Lexer) setState(state string) {
	if state == "" || l.state == state {
		return
	}
	
	l.state = state
	if info, ok := l.states[state]; ok {
		l.groups = info.groups
		l.errorRule = info.error
		l.re = info.regexp
		l.fast = info.fast
	}
}

// Next returns the next token, or nil if EOF is reached.
func (l *Lexer) Next() *Token {
	index := l.index
	
	// If a fallback token matched, we don't need to re-run the RegExp
	if l.queuedGroup != nil {
		token := l.makeToken(l.queuedGroup, l.queuedText, index)
		l.queuedGroup = nil
		l.queuedText = ""
		return token
	}
	
	buffer := l.buffer
	if index >= len(buffer) {
		return nil // EOF
	}
	
	// Fast matching for single characters
	if len(l.fast) > 0 && index < len(buffer) {
		ch := buffer[index]
		if group, ok := l.fast[ch]; ok {
			return l.makeToken(group, string(ch), index)
		}
	}
	
	// Execute RegExp
	re := l.re
	text := buffer[index:]
	match := re.FindStringSubmatchIndex(text)
	
	// Error tokens match the remaining buffer
	if match == nil || match[0] != 0 {
		return l.makeToken(l.errorRule, buffer[index:], index)
	}
	
	// Find which group matched
	group := l.getGroup(match)
	matchedText := text[match[0]:match[1]]
	
	// Handle fallback
	if l.errorRule.fallback && match[0] != 0 {
		l.queuedGroup = group
		l.queuedText = matchedText
		return l.makeToken(l.errorRule, buffer[index:index+match[0]], index)
	}
	
	return l.makeToken(group, matchedText, index)
}

// getGroup finds which rule group matched based on submatch indices.
func (l *Lexer) getGroup(match []int) *rule {
	// Skip first pair (full match)
	for i := 2; i < len(match); i += 2 {
		if match[i] != -1 {
			groupIndex := (i - 2) / 2
			if groupIndex < len(l.groups) {
				return l.groups[groupIndex]
			}
		}
	}
	return l.errorRule
}

// makeToken creates a token from a matched group.
func (l *Lexer) makeToken(group *rule, text string, offset int) *Token {
	// Count line breaks
	lineBreaks := 0
	if group.lineBreaks {
		lineBreaks = strings.Count(text, "\n")
	}
	
	// Determine token type
	tokenType := group.defaultType
	if group.typeFunc != nil {
		if t := group.typeFunc(text); t != "" {
			tokenType = t
		}
	}
	
	// Determine token value
	value := text
	if group.value != nil {
		value = group.value(text)
	}
	
	token := &Token{
		Type:       tokenType,
		Value:      value,
		Text:       text,
		Offset:     offset,
		LineBreaks: lineBreaks,
		Line:       l.line,
		Col:        l.col,
	}
	
	// Update position
	size := len(text)
	l.index += size
	l.line += lineBreaks
	
	if lineBreaks > 0 {
		// Find position after last newline
		lastNL := strings.LastIndex(text, "\n")
		l.col = size - lastNL
	} else {
		l.col += size
	}
	
	// Throw error if required
	if group.shouldThrow {
		panic(fmt.Sprintf("invalid syntax at line %d col %d", token.Line, token.Col))
	}
	
	// Handle state changes
	if group.pop {
		l.popState()
	} else if group.push != "" {
		l.pushState(group.push)
	} else if group.next != "" {
		l.setState(group.next)
	}
	
	return token
}

// popState pops a state from the stack.
func (l *Lexer) popState() {
	if len(l.stack) > 0 {
		state := l.stack[len(l.stack)-1]
		l.stack = l.stack[:len(l.stack)-1]
		l.setState(state)
	}
}

// pushState pushes current state and switches to new state.
func (l *Lexer) pushState(state string) {
	l.stack = append(l.stack, l.state)
	l.setState(state)
}

// Save saves the current lexer state.
func (l *Lexer) Save() map[string]interface{} {
	stackCopy := make([]string, len(l.stack))
	copy(stackCopy, l.stack)
	
	return map[string]interface{}{
		"line":  l.line,
		"col":   l.col,
		"state": l.state,
		"stack": stackCopy,
	}
}

// FormatError formats a token error message.
func (l *Lexer) FormatError(token *Token, message string) string {
	if token == nil {
		return message
	}
	
	return fmt.Sprintf("%s at line %d col %d", message, token.Line, token.Col)
}

// Helper functions

func objectToRules(spec map[string]interface{}) ([]*rule, error) {
	var rules []*rule
	
	for name, value := range spec {
		r, err := parseRule(name, value)
		if err != nil {
			return nil, err
		}
		rules = append(rules, r...)
	}
	
	return rules, nil
}

func parseRule(name string, value interface{}) ([]*rule, error) {
	var rules []*rule
	
	switch v := value.(type) {
	case string:
		// String literal
		rules = append(rules, &rule{
			defaultType: name,
			match:       []*regexp.Regexp{regexp.MustCompile(regexp.QuoteMeta(v))},
			lineBreaks:  false,
		})
		
	case *regexp.Regexp:
		// Regular expression
		rules = append(rules, &rule{
			defaultType: name,
			match:       []*regexp.Regexp{v},
			lineBreaks:  false,
		})
		
	case []string:
		// Array of string literals (e.g., keywords)
		for _, s := range v {
			rules = append(rules, &rule{
				defaultType: name,
				match:       []*regexp.Regexp{regexp.MustCompile(regexp.QuoteMeta(s))},
				lineBreaks:  false,
			})
		}
		
	case map[string]interface{}:
		// Rule object with options
		r := &rule{
			defaultType: name,
			lineBreaks:  false,
		}
		
		// Check for special markers
		if _, isError := v["error"]; isError {
			r.error = true
			r.lineBreaks = true
			r.shouldThrow = true
		}
		
		if _, isFallback := v["fallback"]; isFallback {
			r.fallback = true
			r.lineBreaks = true
		}
		
		// Parse match
		if matchVal, ok := v["match"]; ok {
			switch m := matchVal.(type) {
			case string:
				r.match = []*regexp.Regexp{regexp.MustCompile(regexp.QuoteMeta(m))}
			case *regexp.Regexp:
				r.match = []*regexp.Regexp{m}
			}
		}
		
		// Parse lineBreaks
		if lb, ok := v["lineBreaks"].(bool); ok {
			r.lineBreaks = lb
		}
		
		// Parse state changes
		if next, ok := v["next"].(string); ok {
			r.next = next
		}
		if push, ok := v["push"].(string); ok {
			r.push = push
		}
		if pop, ok := v["pop"].(bool); ok {
			r.pop = pop
		}
		
		// Parse value transform
		if valFunc, ok := v["value"].(func(string) string); ok {
			r.value = valFunc
		}
		
		// Parse type transform
		if typeFunc, ok := v["type"].(func(string) string); ok {
			r.typeFunc = typeFunc
		}
		
		rules = append(rules, r)
	}
	
	return rules, nil
}

func compileRules(rules []*rule, hasStates bool) (*compiledRules, error) {
	var errorRule *rule
	fast := make(map[byte]*rule)
	var groups []*rule
	var parts []string
	
	// Default error rule
	defaultErrorRule := &rule{
		defaultType: "error",
		lineBreaks:  true,
		shouldThrow: true,
	}
	
	for _, r := range rules {
		if r.error || r.fallback {
			if errorRule != nil {
				return nil, fmt.Errorf("multiple error/fallback rules not allowed")
			}
			errorRule = r
		}
		
		// Fast path for single-character strings
		if len(r.match) > 0 && len(fast) < 256 {
			// Check if this is a single-character match
			for _, re := range r.match {
				pattern := re.String()
				// Simple check for single escaped char
				if len(pattern) == 1 {
					fast[pattern[0]] = r
				}
			}
		}
		
		// Add to groups if has match
		if len(r.match) > 0 {
			groups = append(groups, r)
			
			// Combine all match patterns for this rule
			var ruleParts []string
			for _, re := range r.match {
				ruleParts = append(ruleParts, re.String())
			}
			
			// Wrap in capturing group
			if len(ruleParts) == 1 {
				parts = append(parts, "("+ruleParts[0]+")")
			} else {
				parts = append(parts, "((?:"+strings.Join(ruleParts, "|")+"))")
			}
		}
	}
	
	if errorRule == nil {
		errorRule = defaultErrorRule
	}
	
	// Combine all parts into single regexp
	combined := strings.Join(parts, "|")
	if combined == "" {
		combined = "(?!)" // Never matches
	}
	
	re, err := regexp.Compile(combined)
	if err != nil {
		return nil, err
	}
	
	return &compiledRules{
		regexp: re,
		groups: groups,
		fast:   fast,
		error:  errorRule,
	}, nil
}
