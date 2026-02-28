package skill

import "strings"

// Name returns the flat, filesystem-safe skill name derived from a base name
// (already sanitized) and the relative path within the source.
//
// Examples (baseName = "github.com_hayeah_skills"):
//
//	relPath ""    → "github.com_hayeah_skills"
//	relPath "foo" → "github.com_hayeah_skills_foo"
//	relPath "a/b" → "github.com_hayeah_skills_a_b"
func Name(baseName, relPath string) string {
	if relPath == "" {
		return baseName
	}
	sub := strings.ReplaceAll(relPath, "/", "_")
	return baseName + "_" + sub
}
