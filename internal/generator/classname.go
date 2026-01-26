package generator

import (
	"strings"
	"unicode"
)

// DeriveClassName generates a TypeScript class name from a directory name.
// It prepends "Go" and converts to TitleCase, unless the name already starts with "Go".
func DeriveClassName(dirName string) string {
	if dirName == "" || dirName == "." {
		return "GoMain"
	}

	titleCased := toTitleCase(dirName)

	// Don't duplicate "Go" prefix
	if strings.HasPrefix(strings.ToLower(titleCased), "go") {
		return strings.ToUpper(titleCased[:1]) + titleCased[1:]
	}

	return "Go" + titleCased
}

// toTitleCase converts "merkle-wasm" to "MerkleWasm"
func toTitleCase(s string) string {
	var result strings.Builder
	capitalizeNext := true

	for _, r := range s {
		if r == '-' || r == '_' {
			capitalizeNext = true
			continue
		}
		if capitalizeNext {
			result.WriteRune(unicode.ToUpper(r))
			capitalizeNext = false
		} else {
			result.WriteRune(unicode.ToLower(r))
		}
	}

	return result.String()
}
