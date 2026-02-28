package gozkilla

import "strings"

// SkillName returns the flat, filesystem-safe skill name derived from a source
// path and the relative subpath of the skill within that source.
// Both src and relPath have "/" replaced with "_".
//
// Examples:
//
//	src "github.com/hayeah/skills", relPath ""      → "github.com_hayeah_skills"
//	src "github.com/hayeah/skills", relPath "foo"   → "github.com_hayeah_skills_foo"
//	src "github.com/hayeah/skills", relPath "a/b"   → "github.com_hayeah_skills_a_b"
func SkillName(src, relPath string) string {
	base := strings.ReplaceAll(src, "/", "_")
	if relPath == "" {
		return base
	}
	sub := strings.ReplaceAll(relPath, "/", "_")
	return base + "_" + sub
}
