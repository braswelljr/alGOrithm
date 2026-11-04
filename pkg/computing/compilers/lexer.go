package compilers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ErrInput is wrapped by every error about the source text.
var ErrInput = errors.New("invalid program")

// Error is a mistake in the source text at Line:Col (both from 1).
type Error struct {
	Line int
	Col  int
	Msg  string
}

func (e *Error) Error() string { return fmt.Sprintf("%d:%d: %s", e.Line, e.Col, e.Msg) }

// Unwrap makes every Error match ErrInput.
func (e *Error) Unwrap() error { return ErrInput }

// TokenKind is what a token is.
type TokenKind string

// The kinds of token.
const (
	NumberToken   TokenKind = "number"
	IdentToken    TokenKind = "ident"
	KeywordToken  TokenKind = "keyword"
	OperatorToken TokenKind = "operator"
	PunctToken    TokenKind = "punct"
	EOFToken      TokenKind = "eof"
)

// Token is one word of the source: its kind, its text, and where it starts
// (Pos is a byte offset; Line and Col count from 1).
type Token struct {
	Kind TokenKind `json:"kind"`
	Text string    `json:"text"`
	Pos  int       `json:"pos"`
	Line int       `json:"line"`
	Col  int       `json:"col"`
}

// End is the byte offset just after the token.
func (t Token) End() int { return t.Pos + len(t.Text) }

var keywords = map[string]bool{
	"let": true, "if": true, "else": true, "while": true, "fn": true,
	"return": true, "print": true, "true": true, "false": true,
}

var twoChar = map[string]bool{"==": true, "!=": true, "<=": true, ">=": true, "&&": true, "||": true}

// Lex splits src into tokens, ending with an EOF token. Spaces and comments
// (from // to the end of the line) separate tokens and are dropped. It reads
// each character once, keeping the line and column as it goes.
//
// Time: O(n) for n characters.
// Space: O(n) for the tokens.
func Lex(src string) ([]Token, error) {
	var out []Token
	line, col := 1, 1
	i := 0
	advance := func(n int) {
		for k := 0; k < n; k++ {
			if src[i] == '\n' {
				line, col = line+1, 1
			} else {
				col++
			}
			i++
		}
	}
	for i < len(src) {
		c := src[i]
		switch {
		case c == ' ' || c == '\t' || c == '\r' || c == '\n':
			advance(1)
			continue
		case strings.HasPrefix(src[i:], "//"):
			for i < len(src) && src[i] != '\n' {
				advance(1)
			}
			continue
		}
		start, l, cl := i, line, col
		var kind TokenKind
		n := 1
		switch {
		case isDigit(c):
			for i+n < len(src) && isDigit(src[i+n]) {
				n++
			}
			if _, err := strconv.Atoi(src[i : i+n]); err != nil {
				return nil, &Error{l, cl, "number too large: " + src[i:i+n]}
			}
			kind = NumberToken
		case isLetter(c):
			for i+n < len(src) && (isLetter(src[i+n]) || isDigit(src[i+n])) {
				n++
			}
			kind = IdentToken
			if keywords[src[i:i+n]] {
				kind = KeywordToken
			}
		case i+1 < len(src) && twoChar[src[i:i+2]]:
			n, kind = 2, OperatorToken
		case strings.IndexByte("+-*/%<>=!^", c) >= 0:
			kind = OperatorToken
		case strings.IndexByte("(){},;", c) >= 0:
			kind = PunctToken
		default:
			return nil, &Error{l, cl, fmt.Sprintf("unexpected character %q", c)}
		}
		out = append(out, Token{Kind: kind, Text: src[start : start+n], Pos: start, Line: l, Col: cl})
		advance(n)
	}
	return append(out, Token{Kind: EOFToken, Pos: len(src), Line: line, Col: col}), nil
}

func isDigit(c byte) bool  { return c >= '0' && c <= '9' }
func isLetter(c byte) bool { return c == '_' || (c|0x20 >= 'a' && c|0x20 <= 'z') }
