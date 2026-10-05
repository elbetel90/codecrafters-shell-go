package utils

import (
	"unicode"
)

func LongestCommandPrefix(matches []string) string {
	if len(matches) == 0 {
		return ""
	}

	prefix := matches[0]
	for _, m := range matches {
		for len(prefix) > 0 && !hasPrefix(m, prefix) {
			prefix = prefix[:len(prefix)-1]
		}
	}

	return prefix
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func IsValidIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if r != '_' && !unicode.IsLetter(r) {
				return false
			}
			continue
		}
		if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
