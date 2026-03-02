package handlers

import (
	"strings"
	"unicode"
)

func parseArguments(arg string) []string {
	return strings.FieldsFunc(arg, func(r rune) bool {
		return unicode.IsSpace(r) || r == ',' || r == ';'
	})
}
