package gozkilla

import (
	"io/fs"
	"path/filepath"
)

// Found represents a discovered SKILL.md location.
type Found struct {
	// RootDir is the source root being searched.
	RootDir string
	// SkillDir is the directory that contains SKILL.md (may equal RootDir).
	SkillDir string
	// RelPath is the path of SkillDir relative to RootDir ("" for root skill).
	RelPath string
}

// FindSkills discovers all skills under the resolved source directory.
func (r *Resolved) FindSkills() ([]Found, error) {
	return FindAll(r.LocalDir)
}

// FindAll walks rootDir recursively and returns every directory that directly
// contains a SKILL.md file.
func FindAll(rootDir string) ([]Found, error) {
	var found []Found

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() != "SKILL.md" {
			return nil
		}
		skillDir := filepath.Dir(path)
		rel, err := filepath.Rel(rootDir, skillDir)
		if err != nil {
			return err
		}
		if rel == "." {
			rel = ""
		}
		found = append(found, Found{
			RootDir:  rootDir,
			SkillDir: skillDir,
			RelPath:  rel,
		})
		return nil
	})
	return found, err
}
