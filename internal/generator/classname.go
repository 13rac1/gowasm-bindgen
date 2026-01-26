package generator

import (
	"strings"
	"unicode"
)

// DeriveClassName generates a TypeScript class name from a directory name.
// It prepends "Go" and converts to TitleCase, unless the directory name
// is "go" or starts with "go-" or "go_" (to avoid "GoGoWasm" from "go-wasm").
func DeriveClassName(dirName string) string {
	if dirName == "" || dirName == "." {
		return "GoMain"
	}

	titleCased := toTitleCase(dirName)
	lower := strings.ToLower(dirName)

	// Don't duplicate "Go" prefix for directories like "go", "go-wasm", "go_utils"
	if lower == "go" || strings.HasPrefix(lower, "go-") || strings.HasPrefix(lower, "go_") {
		return titleCased
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

// LowerFirst converts first letter to lowercase.
// Used for converting Go function names to JavaScript conventions.
func LowerFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// ToKebabCase converts "GoMain" to "go-main".
// Used for generating TypeScript filenames from class names.
func ToKebabCase(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '-')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}
