package main

import (
	"errors"
	"fmt"
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
	SearchEntity string            // e.g. "company:Databricks" or "Q* Reasoning"
	Sources      []string          // e.g. ["crunchbase", "pitchbook"]
	ReturnFields map[string]string // e.g. {"ceo": "string", "latest_valuation": "number"}
	Confidence   float64           // e.g. 0.8
	CacheTTL     time.Duration     // e.g. 24h
	Format       string            // e.g. "json"
	Language     string            // e.g. "en"
	Country      string            // e.g. "us"
	SafeSearch   string            // e.g. "off"
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

// Parse USQL statement
func (p *USQLParser) Parse() (*USQLQuery, error) {
	query := &USQLQuery{
		ReturnFields: make(map[string]string),
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

			if !p.match(TokIdentifier) {
				return nil, fmt.Errorf("position %d: expected type identifier (e.g. string, number, array)", p.curr.Pos)
			}
			fieldType := p.curr.Val
			p.nextToken()

			query.ReturnFields[fieldName] = fieldType

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

// CompileToDorkQuery translates the AST properties into highly efficient Google search dork templates
func (q *USQLQuery) CompileToDorkQuery() string {
	// Strip type modifiers e.g., company:Databricks -> Databricks
	searchTarget := q.SearchEntity
	if idx := strings.Index(searchTarget, ":"); idx >= 0 {
		searchTarget = searchTarget[idx+1:]
	}

	var builders []string
	builders = append(builders, searchTarget)

	// Inject target fields to guide SGE focus
	for field := range q.ReturnFields {
		builders = append(builders, strings.ReplaceAll(field, "_", " "))
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

// RunUSQLDiagnostics parses and compiles mock queries to verify tokenizing and AST mapping
func RunUSQLDiagnostics() {
	fmt.Println("\n[*] Running USQL Language Compiler Diagnostics...")

	queryStr := `SEARCH company:"Databricks"
FROM crunchbase, pitchbook
RETURN {
  funding_rounds: array,
  latest_valuation: number,
  ceo: string
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
}
