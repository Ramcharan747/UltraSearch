package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// TokenType represents different lexical tokens in USQL
type TokenType int

const (
	TokError TokenType = iota
	TokEOF
	TokSearch
	TokFrom
	TokReturn
	TokWith
	TokCache
	TokFormat
	TokIdentifier
	TokString
	TokNumber
	TokDuration
	TokColon
	TokComma
	TokLeftBrace
	TokRightBrace
	TokGreater
	TokSemicolon
	TokLeftParen
	TokRightParen
)

func (t TokenType) String() string {
	switch t {
	case TokError:
		return "ERROR"
	case TokEOF:
		return "EOF"
	case TokSearch:
		return "SEARCH"
	case TokFrom:
		return "FROM"
	case TokReturn:
		return "RETURN"
	case TokWith:
		return "WITH"
	case TokCache:
		return "CACHE"
	case TokFormat:
		return "FORMAT"
	case TokIdentifier:
		return "IDENTIFIER"
	case TokString:
		return "STRING"
	case TokNumber:
		return "NUMBER"
	case TokDuration:
		return "DURATION"
	case TokColon:
		return ":"
	case TokComma:
		return ","
	case TokLeftBrace:
		return "{"
	case TokRightBrace:
		return "}"
	case TokGreater:
		return ">"
	case TokSemicolon:
		return ";"
	case TokLeftParen:
		return "("
	case TokRightParen:
		return ")"
	default:
		return "UNKNOWN"
	}
}

// Token holds a token's type, literal value, and source location details
type Token struct {
	Type  TokenType
	Val   string
	Pos   int
}

// Lexer implements a simple, hand-written lexical scanner for USQL
type Lexer struct {
	input string
	pos   int
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: input, pos: 0}
}

func (l *Lexer) peek() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) next() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	b := l.input[l.pos]
	l.pos++
	return b
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		b := l.input[l.pos]
		if unicode.IsSpace(rune(b)) {
			l.pos++
		} else {
			break
		}
	}
}

// NextToken scans the input string and returns the next Token
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()
	if l.pos >= len(l.input) {
		return Token{Type: TokEOF, Val: "", Pos: l.pos}
	}

	start := l.pos
	b := l.next()

	switch b {
	case ':':
		return Token{Type: TokColon, Val: ":", Pos: start}
	case ',':
		return Token{Type: TokComma, Val: ",", Pos: start}
	case '{':
		return Token{Type: TokLeftBrace, Val: "{", Pos: start}
	case '}':
		return Token{Type: TokRightBrace, Val: "}", Pos: start}
	case '>':
		return Token{Type: TokGreater, Val: ">", Pos: start}
	case ';':
		return Token{Type: TokSemicolon, Val: ";", Pos: start}
	case '(':
		return Token{Type: TokLeftParen, Val: "(", Pos: start}
	case ')':
		return Token{Type: TokRightParen, Val: ")", Pos: start}
	case '"', '\'':
		// Read quoted string
		quote := b
		var val []byte
		for l.pos < len(l.input) {
			curr := l.next()
			if curr == quote {
				return Token{Type: TokString, Val: string(val), Pos: start}
			}
			if curr == '\\' && l.peek() == quote {
				l.next() // consume escape
				val = append(val, quote)
			} else {
				val = append(val, curr)
			}
		}
		return Token{Type: TokError, Val: "unterminated string literal", Pos: start}
	}

	// Read number or duration
	if unicode.IsDigit(rune(b)) || b == '.' {
		var val []byte
		val = append(val, b)
		for l.pos < len(l.input) && (unicode.IsDigit(rune(l.peek())) || l.peek() == '.') {
			val = append(val, l.next())
		}
		
		// Check for duration prefix suffix (e.g. h, m, s, d)
		if l.pos < len(l.input) && (l.peek() == 'h' || l.peek() == 'm' || l.peek() == 's' || l.peek() == 'd') {
			suffix := l.next()
			return Token{Type: TokDuration, Val: string(val) + string(suffix), Pos: start}
		}

		return Token{Type: TokNumber, Val: string(val), Pos: start}
	}

	// Read identifier or keyword
	if unicode.IsLetter(rune(b)) || b == '_' {
		var val []byte
		val = append(val, b)
		// Allow alphanumeric, dot, or dash inside search entities (e.g. company:Databricks or paper:Q*)
		for l.pos < len(l.input) {
			p := l.peek()
			if unicode.IsLetter(rune(p)) || unicode.IsDigit(rune(p)) || p == '_' || p == '-' || p == '.' || p == '*' || p == '/' {
				val = append(val, l.next())
			} else {
				break
			}
		}

		valStr := string(val)
		upperVal := strings.ToUpper(valStr)

		switch upperVal {
		case "SEARCH":
			return Token{Type: TokSearch, Val: valStr, Pos: start}
		case "FROM":
			return Token{Type: TokFrom, Val: valStr, Pos: start}
		case "RETURN":
			return Token{Type: TokReturn, Val: valStr, Pos: start}
		case "WITH":
			return Token{Type: TokWith, Val: valStr, Pos: start}
		case "CACHE":
			return Token{Type: TokCache, Val: valStr, Pos: start}
		case "FORMAT":
			return Token{Type: TokFormat, Val: valStr, Pos: start}
		default:
			return Token{Type: TokIdentifier, Val: valStr, Pos: start}
		}
	}

	return Token{Type: TokError, Val: fmt.Sprintf("invalid character '%c'", b), Pos: start}
}

// USQLQuery represents the Abstract Syntax Tree (AST) of a compiled USQL query
type USQLQuery struct {
	SearchEntity string                 // e.g. "company:Databricks" or "Q* Reasoning"
	Sources      []string               // e.g. ["crunchbase", "pitchbook"]
	ReturnFields map[string]interface{} // Field Name -> Type/Func/Nested Schema
	Confidence   float64                // e.g. 0.8
	CacheTTL     time.Duration          // e.g. 24h
	Format       string                 // e.g. "json"
	Language     string                 // e.g. "en"
	Country      string                 // e.g. "us"
	SafeSearch   string                 // e.g. "off"
}

// FuncExpr represents a parsed function call in the USQL query AST
type FuncExpr struct {
	Name string
	Args []interface{} // Can be strings, numbers, identifiers, or other FuncExpr
}

// NestedSchema represents a nested JSON block e.g. { school: string }
type NestedSchema struct {
	Fields map[string]interface{}
}

// ArraySchema represents an array of primitive types or nested schemas
type ArraySchema struct {
	ValueType interface{} // Can be a string identifier ("string", "number") or NestedSchema
}

// USQLParser maps scanned tokens into a validated AST
type USQLParser struct {
	lexer *Lexer
	curr  Token
}

func NewUSQLParser(input string) *USQLParser {
	p := &USQLParser{lexer: NewLexer(input)}
	p.nextToken()
	return p
}

func (p *USQLParser) nextToken() {
	p.curr = p.lexer.NextToken()
}

func (p *USQLParser) match(t TokenType) bool {
	return p.curr.Type == t
}

func (p *USQLParser) expect(t TokenType) error {
	if p.curr.Type != t {
		return fmt.Errorf("line position %d: expected token %s, got %s ('%s')", p.curr.Pos, t, p.curr.Type, p.curr.Val)
	}
	p.nextToken()
	return nil
}

func (p *USQLParser) parseValue() (interface{}, error) {
	if p.match(TokLeftBrace) {
		p.nextToken() // consume {
		fields := make(map[string]interface{})
		for !p.match(TokRightBrace) && !p.match(TokEOF) {
			if !p.match(TokIdentifier) {
				return nil, fmt.Errorf("position %d: expected identifier for schema field name, got %s ('%s')", p.curr.Pos, p.curr.Type, p.curr.Val)
			}
			fieldName := p.curr.Val
			p.nextToken()

			if err := p.expect(TokColon); err != nil {
				return nil, err
			}

			val, err := p.parseValue()
			if err != nil {
				return nil, err
			}
			fields[fieldName] = val

			if p.match(TokComma) {
				p.nextToken()
				continue
			}
			break
		}
		if err := p.expect(TokRightBrace); err != nil {
			return nil, err
		}
		return NestedSchema{Fields: fields}, nil
	}

	if p.match(TokString) {
		val := p.curr.Val
		p.nextToken()
		return val, nil
	}

	if p.match(TokNumber) {
		val, err := strconv.ParseFloat(p.curr.Val, 64)
		if err != nil {
			valStr := p.curr.Val
			p.nextToken()
			return valStr, nil
		}
		p.nextToken()
		return val, nil
	}

	if p.match(TokIdentifier) {
		id := p.curr.Val
		p.nextToken()

		// If followed by '(', it's a function call OR array specification
		if p.match(TokLeftParen) {
			p.nextToken() // consume (

			if strings.ToLower(id) == "array" {
				arrVal, err := p.parseValue()
				if err != nil {
					return nil, err
				}
				if err := p.expect(TokRightParen); err != nil {
					return nil, err
				}
				return ArraySchema{ValueType: arrVal}, nil
			}

			var args []interface{}
			for !p.match(TokRightParen) && !p.match(TokEOF) {
				argVal, err := p.parseValue()
				if err != nil {
					return nil, err
				}
				args = append(args, argVal)

				if p.match(TokComma) {
					p.nextToken()
					continue
				}
				break
			}
			if err := p.expect(TokRightParen); err != nil {
				return nil, err
			}
			return FuncExpr{Name: id, Args: args}, nil
		}

		if strings.ToLower(id) == "array" {
			return ArraySchema{ValueType: "string"}, nil
		}

		return id, nil
	}

	return nil, fmt.Errorf("position %d: unexpected token '%s'", p.curr.Pos, p.curr.Val)
}

// Parse USQL statement
func (p *USQLParser) Parse() (*USQLQuery, error) {
	query := &USQLQuery{
		ReturnFields: make(map[string]interface{}),
		CacheTTL:     24 * time.Hour, // Default TTL
		Format:       "json",         // Default format
	}

	// 1. Parse SEARCH
	if err := p.expect(TokSearch); err != nil {
		return nil, err
	}

	// Parse search query target (identifier or string)
	if p.match(TokString) {
		query.SearchEntity = p.curr.Val
		p.nextToken()
	} else if p.match(TokIdentifier) {
		query.SearchEntity = p.curr.Val
		p.nextToken()
		// Allow parsed key:value pairs like company:Databricks
		if p.match(TokColon) {
			p.nextToken() // consume :
			if p.match(TokIdentifier) || p.match(TokString) {
				query.SearchEntity = query.SearchEntity + ":" + p.curr.Val
				p.nextToken()
			} else {
				return nil, fmt.Errorf("position %d: expected identifier/string value after colon", p.curr.Pos)
			}
		}
	} else {
		return nil, fmt.Errorf("position %d: expected search entity text or string", p.curr.Pos)
	}

	// 2. Parse FROM (optional)
	if p.match(TokFrom) {
		p.nextToken() // consume FROM
		for {
			if p.match(TokIdentifier) {
				query.Sources = append(query.Sources, p.curr.Val)
				p.nextToken()
			} else {
				return nil, fmt.Errorf("position %d: expected source identifier in FROM block", p.curr.Pos)
			}

			if p.match(TokComma) {
				p.nextToken() // consume ,
				continue
			}
			break
		}
	}

	// 3. Parse RETURN (optional)
	if p.match(TokReturn) {
		p.nextToken() // consume RETURN
		if err := p.expect(TokLeftBrace); err != nil {
			return nil, err
		}

		for !p.match(TokRightBrace) && !p.match(TokEOF) {
			if !p.match(TokIdentifier) {
				return nil, fmt.Errorf("position %d: expected identifier for schema field name", p.curr.Pos)
			}
			fieldName := p.curr.Val
			p.nextToken()

			if err := p.expect(TokColon); err != nil {
				return nil, err
			}

			val, err := p.parseValue()
			if err != nil {
				return nil, err
			}

			query.ReturnFields[fieldName] = val

			if p.match(TokComma) {
				p.nextToken()
				continue
			}
			break
		}

		if err := p.expect(TokRightBrace); err != nil {
			return nil, err
		}
	}

	// 4. Parse WITH (optional)
	if p.match(TokWith) {
		p.nextToken() // consume WITH
		for {
			if p.match(TokIdentifier) {
				paramName := strings.ToLower(p.curr.Val)
				p.nextToken()

				if paramName == "confidence" {
					if err := p.expect(TokGreater); err != nil {
						return nil, err
					}
					if !p.match(TokNumber) {
						return nil, fmt.Errorf("position %d: expected numeric threshold after confidence >", p.curr.Pos)
					}
					val, err := strconv.ParseFloat(p.curr.Val, 64)
					if err != nil {
						return nil, fmt.Errorf("position %d: invalid float number: %v", p.curr.Pos, err)
					}
					query.Confidence = val
					p.nextToken()
				} else {
					if err := p.expect(TokColon); err != nil {
						return nil, err
					}
					if !p.match(TokIdentifier) && !p.match(TokString) {
						return nil, fmt.Errorf("position %d: expected parameter value in WITH statement", p.curr.Pos)
					}
					paramVal := p.curr.Val
					p.nextToken()

					switch paramName {
					case "language":
						query.Language = paramVal
					case "country":
						query.Country = paramVal
					case "safe_search", "safe":
						query.SafeSearch = paramVal
					}
				}
			} else {
				return nil, fmt.Errorf("position %d: expected parameter identifier in WITH block", p.curr.Pos)
			}

			if p.match(TokComma) {
				p.nextToken() // consume ,
				continue
			}
			break
		}
	}

	// 5. Parse CACHE (optional)
	if p.match(TokCache) {
		p.nextToken() // consume CACHE
		if p.match(TokIdentifier) && strings.ToLower(p.curr.Val) == "ttl" {
			p.nextToken() // consume ttl
			if err := p.expect(TokColon); err != nil {
				return nil, err
			}
			if !p.match(TokDuration) {
				return nil, fmt.Errorf("position %d: expected duration (e.g. 24h, 30m) for cache TTL", p.curr.Pos)
			}
			dVal := p.curr.Val
			p.nextToken()

			// Parse custom duration
			duration, err := parseUSQLDuration(dVal)
			if err != nil {
				return nil, fmt.Errorf("position %d: invalid duration format: %v", p.curr.Pos, err)
			}
			query.CacheTTL = duration
		} else {
			return nil, fmt.Errorf("position %d: expected 'ttl' parameter in CACHE block", p.curr.Pos)
		}
	}

	// 6. Parse FORMAT (optional)
	if p.match(TokFormat) {
		p.nextToken() // consume FORMAT
		if !p.match(TokIdentifier) {
			return nil, fmt.Errorf("position %d: expected output format (e.g. json, text) in FORMAT statement", p.curr.Pos)
		}
		query.Format = strings.ToLower(p.curr.Val)
		p.nextToken()
	}

	// Consume optional trailing semicolon
	if p.match(TokSemicolon) {
		p.nextToken()
	}

	if !p.match(TokEOF) {
		return nil, fmt.Errorf("position %d: unexpected trailing token '%s'", p.curr.Pos, p.curr.Val)
	}

	return query, nil
}

func parseUSQLDuration(val string) (time.Duration, error) {
	if len(val) < 2 {
		return 0, errors.New("duration string too short")
	}
	suffix := val[len(val)-1]
	numberStr := val[:len(val)-1]
	amount, err := strconv.ParseInt(numberStr, 10, 64)
	if err != nil {
		return 0, err
	}

	switch suffix {
	case 's':
		return time.Duration(amount) * time.Second, nil
	case 'm':
		return time.Duration(amount) * time.Minute, nil
	case 'h':
		return time.Duration(amount) * time.Hour, nil
	case 'd':
		return time.Duration(amount) * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown duration suffix '%c'", suffix)
	}
}

func flattenFields(fields map[string]interface{}) []string {
	var list []string
	for k, v := range fields {
		list = append(list, strings.ReplaceAll(k, "_", " "))
		switch val := v.(type) {
		case NestedSchema:
			list = append(list, flattenFields(val.Fields)...)
		case ArraySchema:
			if subNS, ok := val.ValueType.(NestedSchema); ok {
				list = append(list, flattenFields(subNS.Fields)...)
			} else if subId, ok := val.ValueType.(string); ok {
				list = append(list, subId)
			} else if subFunc, ok := val.ValueType.(FuncExpr); ok {
				list = append(list, subFunc.Name)
			}
		case FuncExpr:
			list = append(list, val.Name)
		}
	}
	return list
}

// CompileToDorkQuery translates the AST properties into highly efficient Google search dork templates
func (q *USQLQuery) CompileToDorkQuery() string {
	searchTarget := q.SearchEntity
	if idx := strings.Index(searchTarget, ":"); idx >= 0 {
		searchTarget = searchTarget[idx+1:]
	}

	var builders []string
	builders = append(builders, searchTarget)

	// Flatten nested fields to guide SGE focus
	flatFields := flattenFields(q.ReturnFields)
	for _, field := range flatFields {
		builders = append(builders, field)
	}

	// Bind target source dorks
	var siteFilters []string
	for _, source := range q.Sources {
		switch strings.ToLower(source) {
		case "crunchbase":
			siteFilters = append(siteFilters, "site:crunchbase.com/organization")
		case "pitchbook":
			siteFilters = append(siteFilters, "site:pitchbook.com")
		case "dealroom":
			siteFilters = append(siteFilters, "site:dealroom.co")
		case "arxiv":
			siteFilters = append(siteFilters, "site:arxiv.org")
		case "wikipedia":
			siteFilters = append(siteFilters, "site:wikipedia.org")
		case "sec_edgar", "sec":
			siteFilters = append(siteFilters, "site:sec.gov")
		}
	}

	if len(siteFilters) > 0 {
		builders = append(builders, fmt.Sprintf("(%s)", strings.Join(siteFilters, " OR ")))
	}

	return strings.Join(builders, " ")
}

// USQLFunction defines the signature for local Go standard functions
type USQLFunction func(args []interface{}) (interface{}, error)

// GlobalFunctionRegistry maps standard enterprise keywords to Go handlers
var GlobalFunctionRegistry = map[string]USQLFunction{
	"UPPER": func(args []interface{}) (interface{}, error) {
		if len(args) == 0 {
			return "", nil
		}
		return strings.ToUpper(fmt.Sprintf("%v", args[0])), nil
	},
	"LOWER": func(args []interface{}) (interface{}, error) {
		if len(args) == 0 {
			return "", nil
		}
		return strings.ToLower(fmt.Sprintf("%v", args[0])), nil
	},
	"TITLE": func(args []interface{}) (interface{}, error) {
		if len(args) == 0 {
			return "", nil
		}
		return strings.Title(fmt.Sprintf("%v", args[0])), nil
	},
	"TRIM": func(args []interface{}) (interface{}, error) {
		if len(args) == 0 {
			return "", nil
		}
		return strings.TrimSpace(fmt.Sprintf("%v", args[0])), nil
	},
	"CLEAN_PII": func(args []interface{}) (interface{}, error) {
		if len(args) == 0 {
			return "", nil
		}
		val := fmt.Sprintf("%v", args[0])
		emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
		val = emailRegex.ReplaceAllString(val, "[EMAIL_REDACTED]")
		phoneRegex := regexp.MustCompile(`\+?\d{1,4}[-.\s]?\(?\d{1,3}\)?[-.\s]?\d{1,4}[-.\s]?\d{1,4}[-.\s]?\d{1,9}`)
		val = phoneRegex.ReplaceAllString(val, "[PHONE_REDACTED]")
		return val, nil
	},
	"CONVERT_CURRENCY": func(args []interface{}) (interface{}, error) {
		if len(args) == 0 {
			return 0.0, nil
		}
		valStr := fmt.Sprintf("%v", args[0])
		num, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			return valStr, nil
		}
		target := "USD"
		if len(args) > 1 {
			target = strings.ToUpper(fmt.Sprintf("%v", args[1]))
		}
		switch target {
		case "EUR":
			return num * 0.92, nil
		case "GBP":
			return num * 0.79, nil
		default:
			return num, nil
		}
	},
	"ESTIMATE_ARR": func(args []interface{}) (interface{}, error) {
		if len(args) == 0 {
			return 0.0, nil
		}
		valStr := fmt.Sprintf("%v", args[0])
		num, err := strconv.ParseFloat(valStr, 64)
		if err == nil {
			return num * 4.0, nil // Annualized ARR
		}
		return valStr + " (Annualized Estimate)", nil
	},
}

// EvaluateUSQLFunctions maps and processes standard library calls on parsed SGE search responses
func EvaluateUSQLFunctions(schema map[string]interface{}, data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, schemaVal := range schema {
		dataVal, exists := data[k]
		if !exists {
			continue
		}
		result[k] = evaluateSingleField(schemaVal, dataVal)
	}
	return result
}

func evaluateSingleField(schemaVal interface{}, dataVal interface{}) interface{} {
	switch s := schemaVal.(type) {
	case FuncExpr:
		var resolvedArgs []interface{}
		for _, arg := range s.Args {
			if strArg, ok := arg.(string); ok && (strArg == "string" || strArg == "number") {
				resolvedArgs = append(resolvedArgs, dataVal)
			} else if subFunc, ok := arg.(FuncExpr); ok {
				resolvedArgs = append(resolvedArgs, evaluateSingleField(subFunc, dataVal))
			} else {
				resolvedArgs = append(resolvedArgs, arg)
			}
		}

		fn, ok := GlobalFunctionRegistry[strings.ToUpper(s.Name)]
		if ok {
			res, err := fn(resolvedArgs)
			if err == nil {
				return res
			}
		}
		// Cognitive AI Fallback: SGE handles dynamic unverified community functions
		if len(resolvedArgs) > 0 {
			return resolvedArgs[0]
		}
		return dataVal

	case NestedSchema:
		if m, ok := dataVal.(map[string]interface{}); ok {
			nestedRes := make(map[string]interface{})
			for subK, subSchema := range s.Fields {
				if subData, ok := m[subK]; ok {
					nestedRes[subK] = evaluateSingleField(subSchema, subData)
				}
			}
			return nestedRes
		}
		return dataVal

	case ArraySchema:
		if arr, ok := dataVal.([]interface{}); ok {
			var nestedArr []interface{}
			for _, item := range arr {
				nestedArr = append(nestedArr, evaluateSingleField(s.ValueType, item))
			}
			return nestedArr
		}
		return dataVal
	}
	return dataVal
}

// RunUSQLDiagnostics parses and compiles mock queries to verify tokenizing and AST mapping
func RunUSQLDiagnostics() {
	fmt.Println("\n[*] Running USQL Language Compiler Diagnostics...")

	queryStr := `SEARCH company:"TLC Capital"
FROM crunchbase, pitchbook
RETURN {
  personnel: array({
    name: UPPER(string),
    title: string,
    education: { school: string }
  }),
  estimated_arr: ESTIMATE_ARR(number)
}
WITH confidence > 0.8, language:en, country:us, safe:off
CACHE ttl:24h
FORMAT json;`

	fmt.Printf("Parsing raw statement:\n%s\n\n", queryStr)
	parser := NewUSQLParser(queryStr)
	ast, err := parser.Parse()
	if err != nil {
		fmt.Printf("❌ Parsing error: %v\n", err)
		return
	}

	fmt.Println("🎉 AST parsed successfully!")
	fmt.Printf("Search Entity:   %s\n", ast.SearchEntity)
	fmt.Printf("Sources Chosen:  %v\n", ast.Sources)
	fmt.Printf("Return Schema:   %+v\n", ast.ReturnFields)
	fmt.Printf("Confidence:      %v\n", ast.Confidence)
	fmt.Printf("Language filter: %s\n", ast.Language)
	fmt.Printf("Country filter:  %s\n", ast.Country)
	fmt.Printf("SafeSearch flag: %s\n", ast.SafeSearch)
	fmt.Printf("Cache TTL:       %v\n", ast.CacheTTL)
	fmt.Printf("Format Output:   %s\n", ast.Format)

	compiledDork := ast.CompileToDorkQuery()
	fmt.Printf("\nGenerated Google Dork Template:\n%s\n", compiledDork)

	// Mock SGE output validation
	mockSGEData := map[string]interface{}{
		"personnel": []interface{}{
			map[string]interface{}{
				"name":  "glen lindenstaedt",
				"title": "Managing Director",
				"education": map[string]interface{}{
					"school": "Stanford University",
				},
			},
		},
		"estimated_arr": "2500000",
	}
	fmt.Printf("\nMocking SGE Raw Output:\n%+v\n", mockSGEData)
	evaluated := EvaluateUSQLFunctions(ast.ReturnFields, mockSGEData)
	evaluatedJSON, _ := json.MarshalIndent(evaluated, "", "  ")
	fmt.Printf("\nAfter Global Function Registry local evaluation:\n%s\n", string(evaluatedJSON))
}

// ParseHybridQuery checks if the query is a formal USQL statement. If not, it parses the query semantically.
func ParseHybridQuery(input string) (*USQLQuery, error) {
	trimmed := strings.TrimSpace(input)
	if len(trimmed) >= 6 && strings.EqualFold(trimmed[:6], "SEARCH") {
		parser := NewUSQLParser(trimmed)
		return parser.Parse()
	}

	q := &USQLQuery{
		ReturnFields: make(map[string]interface{}),
		CacheTTL:     24 * time.Hour,
		Format:       "json",
	}

	// 1. Semantic router mapping: check if query has high similarity to a staged Skill Book
	if len(GlobalRegistry) > 0 {
		if book, _, found := SemanticRouteQuery(trimmed); found {
			q.Sources = book.Domains
		}
	}

	// 2. Identify the target entity
	q.SearchEntity = extractTargetEntity(trimmed)

	// 3. Extract requested fields using keyword matching
	lowerInput := strings.ToLower(trimmed)
	fieldKeywords := map[string][]string{
		"ceo":              {"ceo", "chief executive", "leader"},
		"founder":          {"founder", "co-founder", "started by"},
		"valuation":        {"valuation", "worth", "market cap"},
		"funding":          {"funding", "raised", "capital", "series"},
		"revenue":          {"revenue", "arr", "annual recurring", "sales"},
		"key_personnel":    {"executives", "team", "personnel", "directors", "board"},
		"arxiv_id":         {"arxiv", "paper id", "identifier"},
		"abstract_summary": {"abstract", "summary", "tldr"},
		"authors":          {"authors", "written by", "investigators"},
	}

	for field, keywords := range fieldKeywords {
		for _, kw := range keywords {
			if strings.Contains(lowerInput, kw) {
				if field == "name" || field == "ceo" || field == "founder" {
					q.ReturnFields[field] = FuncExpr{Name: "UPPER", Args: []interface{}{"string"}}
				} else {
					q.ReturnFields[field] = "string"
				}
				break
			}
		}
	}

	if len(q.ReturnFields) == 0 {
		q.ReturnFields["summary"] = "string"
		q.ReturnFields["key_details"] = "string"
	}

	if q.SearchEntity == "" {
		q.SearchEntity = trimmed
	}

	return q, nil
}

func extractTargetEntity(input string) string {
	lower := strings.ToLower(input)
	words := strings.Fields(input)
	
	if idx := strings.Index(lower, "company:"); idx != -1 {
		rest := input[idx+8:]
		fields := strings.Fields(rest)
		if len(fields) > 0 {
			return strings.Trim(fields[0], `"'.,`)
		}
	}

	patterns := []string{"ceo of ", "founder of ", "valuation of ", "funding for ", "funding of ", "about ", "investors in "}
	for _, p := range patterns {
		if idx := strings.Index(lower, p); idx != -1 {
			target := input[idx+len(p):]
			targetWords := strings.Fields(target)
			if len(targetWords) > 0 {
				var entityParts []string
				for _, w := range targetWords {
					wLower := strings.ToLower(w)
					if wLower == "and" || wLower == "with" || wLower == "on" || wLower == "at" || wLower == "in" || wLower == "from" {
						break
					}
					entityParts = append(entityParts, strings.Trim(w, `"'.,`))
					if len(entityParts) >= 2 {
						break
					}
				}
				return strings.Join(entityParts, " ")
			}
		}
	}

	var properNouns []string
	for i, w := range words {
		if i == 0 {
			continue
		}
		if len(w) > 0 && unicode.IsUpper(rune(w[0])) {
			properNouns = append(properNouns, strings.Trim(w, `"'.,`))
			if len(properNouns) >= 2 {
				break
			}
		}
	}
	if len(properNouns) > 0 {
		return strings.Join(properNouns, " ")
	}

	if len(words) > 0 {
		return words[len(words)-1]
	}
	return input
}
