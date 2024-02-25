package jsonparser

import (
	"fmt"
	"strconv"
)

type parser struct {
	source string
}

func NewParser(source string) *parser {
	return &parser{
		source: source,
	}
}

func hasCompletedTokens(t []*token, cursor int) bool {
	return cursor >= len(t)
}

func safeGetToken(t []*token, cursor int) *token {
	if !hasCompletedTokens(t, cursor) {
		return t[cursor]
	}
	return nil
}

func getTokenParseError(msg string, t *token) error {
	return fmt.Errorf("%s at [ %s ] (line %d, column %d)", msg, t.value, t.lineNumber, t.linePosition)
}

func parseInteger(t []*token, cursor int) (int, interface{}) {
	if t[cursor].kind == intToken {
		v, err := strconv.Atoi(t[cursor].value)
		if err != nil {
			return cursor, nil
		}
		return cursor + 1, v
	}
	return cursor, nil

}

func parseFloat(t []*token, cursor int) (int, interface{}) {
	if t[cursor].kind == floatToken {
		v, err := strconv.ParseFloat(t[cursor].value, 64)
		if err != nil {
			panic(fmt.Sprintf("Couldn't parse float at token [%d]. %s", cursor, err))
		}
		return cursor + 1, v
	}
	return cursor, nil
}

func parseBoolean(t []*token, cursor int) (int, interface{}) {
	if t[cursor].kind == booleanToken {
		var v bool
		switch t[cursor].value {
		case "true":
			v = true
		case "false":
			v = false
		default:
			return cursor, nil
		}
		return cursor + 1, v
	}
	return cursor, nil
}

func parseString(t []*token, cursor int) (int, interface{}) {
	tt := safeGetToken(t, cursor)
	if tt != nil && tt.kind == stringToken {
		return cursor + 1, t[cursor].value
	}
	return cursor, nil
}

func parseNil(t []*token, cursor int) (int, interface{}) {
	tt := safeGetToken(t, cursor)
	if tt != nil && tt.kind == nullToken {
		return cursor + 1, nil
	}
	return cursor, nil
}

func parseArray(t []*token, cursor int) (int, interface{}) {
	tt := safeGetToken(t, cursor)
	if tt != nil && tt.value == "[" {
		idx := cursor
		idx++
		s := make([]interface{}, 0)
		for {
			tt := safeGetToken(t, idx)
			if tt == nil {
				return cursor, getTokenParseError("failed to parse JSON", tt)
			}
			vc, val := parseJSONValue(t, idx)
			if vc != idx {
				idx = vc
				s = append(s, val)
				nt := safeGetToken(t, idx)
				if nt != nil && nt.value == "," {
					ntt := safeGetToken(t, idx+1)
					if ntt != nil && ntt.kind == symbolToken {
						msg := fmt.Sprintf("unexpected %s after array element in JSON", nt.value)
						return cursor, getTokenParseError(msg, tt)
					}
					idx++
				}
			} else {
				if tt.value == "]" {
					return idx + 1, s
				} else {
					return cursor, getTokenParseError("expected ',' or ']' before array element in JSON", tt)
				}
			}
		}
	}
	return cursor, nil

}

func parseObject(t []*token, cursor int) (int, interface{}) {
	tt := safeGetToken(t, cursor)
	if tt != nil && tt.value == "{" {
		idx := cursor
		idx++
		m := make(map[string]interface{})
		for {
			tt := safeGetToken(t, idx)
			if tt == nil {
				return cursor, getTokenParseError("failed to parse JSON", tt)
			}
			switch tt.kind {
			case stringToken:
				var key interface{}
				idx, key = parseJSONKey(t, idx)

				nt := safeGetToken(t, idx)
				if nt == nil {
					return cursor, getTokenParseError("expected ':' before property value in JSON", nt)
				}
				idx++

				vc, val := parseJSONValue(t, idx)
				if vc == idx {
					return cursor, val.(error)
				}
				idx = vc

				m[key.(string)] = val

				nt = safeGetToken(t, idx)
				if nt != nil && nt.value == "," {
					ntt := safeGetToken(t, idx+1)
					if ntt != nil && ntt.kind == symbolToken {
						msg := fmt.Sprintf("unexpected %s after array element in JSON", nt.value)
						return cursor, getTokenParseError(msg, tt)
					}
					idx++
				}
			case symbolToken:
				if tt.value == "}" {
					return idx + 1, m
				}
			default:
				return cursor, getTokenParseError("expected ',' or '}' before property value in JSON", tt)
			}
		}
	}
	return cursor, nil
}

func parseJSONKey(t []*token, cursor int) (int, interface{}) {
	k := safeGetToken(t, cursor)
	if k == nil {
		err := getTokenParseError("failed to parse JSON", t[cursor])
		return cursor, err
	}
	kk := k.value
	return cursor + 1, kk
}

func parseJSONValue(t []*token, cursor int) (int, interface{}) {
	if !hasCompletedTokens(t, cursor) {
		x := []func([]*token, int) (int, interface{}){
			parseArray,
			parseBoolean,
			parseFloat,
			parseInteger,
			parseNil,
			parseString,
			parseObject,
		}
		for _, f := range x {
			c, v := f(t, cursor)
			if c != cursor {
				return c, v
			}
			switch v := v.(type) {
			case error:
				return cursor, v
			}
		}
	}
	err := getTokenParseError("failed to parse JSON", t[cursor])
	return cursor, err
}

func (p *parser) Parse() (v interface{}, err error) {
	defer func() {
		if _err := recover(); _err != nil {
			switch _err := _err.(type) {
			case error:
				err = _err
			default:
				err = fmt.Errorf("failed to parse JSON")
			}
		}
	}()
	s := []rune(p.source)
	l := &lexer{
		source: s,
	}
	tokens := l.buildTokens()
	cursor := 0
	cursor, v = parseJSONValue(tokens, 0)
	if cursor != len(tokens) {
		getTokenParseError("failed to parse JSON", tokens[cursor])
	}
	switch v := v.(type) {
	case error:
		err = v
	}
	return v, err
}
