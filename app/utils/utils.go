package utils

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
