package skkdic

import (
	"unicode"
	"unicode/utf8"
)

func isOkuriAri(s string) bool {
	r, n := utf8.DecodeRuneInString(s)
	if n == 0 {
		return false
	}
	if r == '>' || r == '#' {
		s = s[n:]
	}

	r, n = utf8.DecodeRuneInString(s)
	if n == 0 {
		return false
	}

	if r <= unicode.MaxASCII {
		return false
	}

	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			return true
		}
	}

	return false
}
