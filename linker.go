package gozkilla

import (
	"fmt"
	"os"
	"path/filepath"
)

// LinkResult reports what happened for a single skill symlink.
type LinkResult struct {
	Name   string
	Target string
	Action string // "created", "updated", "skipped", "error"
	Err    error
}

// InstallAll installs all found skills into destDir by creating symlinks.
func InstallAll(baseName string, skills []Found, destDir string) []LinkResult {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return []LinkResult{{Err: fmt.Errorf("create dest dir: %w", err)}}
	}

	results := make([]LinkResult, len(skills))
	for i, s := range skills {
		results[i] = installOne(baseName, s, destDir)
	}
	return results
}

func installOne(baseName string, s Found, destDir string) LinkResult {
	skillName := SkillName(baseName, s.RelPath)
	linkPath := filepath.Join(destDir, skillName)

	// Use absolute target so symlinks work from any cwd.
	absTarget, err := filepath.Abs(s.SkillDir)
	if err != nil {
		return LinkResult{Name: skillName, Err: err}
	}

	// Check existing symlink.
	existing, err := os.Readlink(linkPath)
	if err == nil {
		if existing == absTarget {
			return LinkResult{Name: skillName, Target: absTarget, Action: "skipped"}
		}
		// Wrong target — update.
		if err := os.Remove(linkPath); err != nil {
			return LinkResult{Name: skillName, Err: fmt.Errorf("remove old symlink: %w", err)}
		}
	} else if !os.IsNotExist(err) {
		// Something else at that path (regular file/dir).
		return LinkResult{Name: skillName, Err: fmt.Errorf("unexpected file at %s: %w", linkPath, err)}
	}

	if err := os.Symlink(absTarget, linkPath); err != nil {
		return LinkResult{Name: skillName, Err: fmt.Errorf("symlink: %w", err)}
	}

	action := "created"
	if existing != "" {
		action = "updated"
	}
	return LinkResult{Name: skillName, Target: absTarget, Action: action}
}

// PrintResults prints a summary of link results to stdout.
func PrintResults(results []LinkResult) {
	for _, r := range results {
		if r.Err != nil {
			fmt.Printf("  error    %s: %v\n", r.Name, r.Err)
		} else {
			fmt.Printf("  %-8s %s → %s\n", r.Action, r.Name, r.Target)
		}
	}
}
