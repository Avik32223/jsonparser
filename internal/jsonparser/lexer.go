package jsonparser

import (
	"fmt"
	"unicode"
)

type tokenKind uint

const (
	stringToken tokenKind = iota
	intToken
	floatToken
	booleanToken
	nullToken

	// leftBraceToken
	// rightBraceToken
	// leftCurlyToken
	// rightCurlyToken
	// commaToken
	// colonToken
	symbolToken
	endOfFileToken
)

type token struct {
	kind         tokenKind
	value        string
	lineNumber   int
	linePosition int
}

type lexer struct {
	source []rune
}

func hasConsumedAllRunes(s []rune, cursor int) bool {
	return cursor >= len(s)
}

func safeGetCurrentRune(s []rune, cursor int) rune {
	if hasConsumedAllRunes(s, cursor) {
		return rune(0)
	}
	return s[cursor]
}

func safeGetNextRune(s []rune, cursor int) rune {
	if cursor+1 >= len(s) {
		return rune(0)
	}
	return s[cursor+1]
}

func consumeLineReturn(s []rune, cursor int) (int, int) {
	switch safeGetCurrentRune(s, cursor) {
	case '\n':
		return 1, cursor + 1
	case '\r':
		return 1, cursor + 1
	}
	return 0, cursor
}

func consumeWhiteSpace(s []rune, cursor int) int {
	if unicode.IsSpace(s[cursor]) {
		return cursor + 1
	}
	return cursor
}

func makeSymbolToken(s []rune, cursor int) (int, *token) {
	c := s[cursor]
	switch c {
	case '[', ']', '{', '}', ':', ',':
		return cursor + 1, &token{
			kind:  symbolToken,
			value: string(c),
		}
	}
	return cursor, nil
}

func makeStringToken(s []rune, cursor int) (int, *token) {
	start, end := cursor, cursor
	if s[start] == '"' {
		end++
		for safeGetCurrentRune(s, end) != '"' && !hasConsumedAllRunes(s, end) {
			end++
		}
		if hasConsumedAllRunes(s, end) {
			panic("Unterminated string")
		}
		end++
		return end, &token{
			kind:  stringToken,
			value: string(s[start+1 : end-1]),
		}
	}
	return cursor, nil
}

func makeNumberToken(s []rune, cursor int) (int, *token) {
	isNegative := false
	isFloat := false
	start := cursor
	if s[start] == '-' {
		isNegative = true
		start++
	}
	if unicode.IsDigit(s[start]) {
		end := start
		for unicode.IsDigit(safeGetCurrentRune(s, end)) {
			end++
		}
		if safeGetCurrentRune(s, end) == '.' {
			if !unicode.IsDigit(safeGetNextRune(s, end)) {
				panic("couldn't parse float")
			}
			isFloat = true
			end++
			for unicode.IsDigit(safeGetCurrentRune(s, end)) {
				end++
			}
		}
		value := s[start:end]
		if isNegative {
			value = append([]rune{'-'}, value...)
		}
		k := intToken
		if isFloat {
			k = floatToken
		}
		return end, &token{
			kind:  k,
			value: string(value),
		}
	}
	if isNegative {
		panic("no number after -")
	}
	return cursor, nil
}

func makeBooleanToken(s []rune, cursor int) (int, *token) {
	start, end := cursor, cursor
	for unicode.IsLetter(safeGetCurrentRune(s, end)) {
		end++
	}
	value := string(s[start:end])
	switch value {
	case "true", "false":
		return end, &token{
			kind:  booleanToken,
			value: value,
		}
	}
	return cursor, nil
}

func makeNullToken(s []rune, cursor int) (int, *token) {
	start, end := cursor, cursor
	for unicode.IsLetter(safeGetCurrentRune(s, end)) {
		end++
	}
	value := string(s[start:end])
	switch value {
	case "null":
		return end, &token{
			kind:  nullToken,
			value: value,
		}
	}
	return cursor, nil
}

func (l *lexer) buildTokens() []*token {
	cursor := 0
	var tokens []*token
	var t *token
	lineNumber, linePosition := 1, 1
	for cursor < len(l.source) {
		ln, c := consumeLineReturn(l.source, cursor)
		if ln != 0 {
			lineNumber++
			linePosition = cursor
		}
		if c != cursor {
			cursor = c
			continue
		}
		c = consumeWhiteSpace(l.source, cursor)
		if c != cursor {
			cursor = c
			continue
		}
		cursor, t = makeSymbolToken(l.source, cursor)
		if t != nil {
			t.lineNumber = lineNumber
			t.linePosition = cursor - linePosition - len(t.value)
			tokens = append(tokens, t)
			continue
		}
		cursor, t = makeStringToken(l.source, cursor)
		if t != nil {
			t.lineNumber = lineNumber
			t.linePosition = cursor - linePosition - len(t.value)
			tokens = append(tokens, t)
			continue
		}
		cursor, t = makeNumberToken(l.source, cursor)
		if t != nil {
			t.lineNumber = lineNumber
			t.linePosition = cursor - linePosition - len(t.value)
			tokens = append(tokens, t)
			continue
		}
		cursor, t = makeBooleanToken(l.source, cursor)
		if t != nil {
			t.lineNumber = lineNumber
			t.linePosition = cursor - linePosition - len(t.value)
			tokens = append(tokens, t)
			continue
		}
		cursor, t = makeNullToken(l.source, cursor)
		if t != nil {
			t.lineNumber = lineNumber
			t.linePosition = cursor - linePosition - len(t.value)
			tokens = append(tokens, t)
			continue
		}
		panic(fmt.Sprintf("failed to parse JSON at (line %d, column %d)", lineNumber, linePosition))
	}
	return tokens
}
