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
	Action string // "create", "update", "skip", "remove"
	Err    error
}

// Linker manages symlinks in a destination directory.
type Linker struct {
	DestDir string
	Dry     bool
}

// Install creates symlinks for all found skills (additive only).
func (l *Linker) Install(baseName string, skills []Found) []LinkResult {
	if err := os.MkdirAll(l.DestDir, 0o755); err != nil {
		return []LinkResult{{Err: fmt.Errorf("create dest dir: %w", err)}}
	}

	results := make([]LinkResult, len(skills))
	for i, s := range skills {
		results[i] = l.installOne(baseName, s)
	}
	return results
}

func (l *Linker) installOne(baseName string, s Found) LinkResult {
	skillName := SkillName(baseName, s.RelPath)
	linkPath := filepath.Join(l.DestDir, skillName)

	// Use absolute target so symlinks work from any cwd.
	absTarget, err := filepath.Abs(s.SkillDir)
	if err != nil {
		return LinkResult{Name: skillName, Err: err}
	}

	// Check existing symlink.
	existing, err := os.Readlink(linkPath)
	if err == nil {
		if existing == absTarget {
			return LinkResult{Name: skillName, Target: absTarget, Action: "skip"}
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

	action := "create"
	if existing != "" {
		action = "update"
	}
	return LinkResult{Name: skillName, Target: absTarget, Action: action}
}

// Sync makes DestDir match desired exactly: creates missing links,
// updates changed ones, and removes links not present in desired.
// Respects l.Dry to report without touching the filesystem.
func (l *Linker) Sync(desired map[string]string) []LinkResult {
	if !l.Dry {
		if err := os.MkdirAll(l.DestDir, 0o755); err != nil {
			return []LinkResult{{Err: fmt.Errorf("create dest dir: %w", err)}}
		}
	}

	// Scan current symlinks in DestDir.
	current := map[string]string{} // name → readlink target
	entries, err := os.ReadDir(l.DestDir)
	if err != nil && !os.IsNotExist(err) {
		return []LinkResult{{Err: fmt.Errorf("read dest dir: %w", err)}}
	}
	for _, e := range entries {
		if e.Type()&os.ModeSymlink == 0 {
			continue
		}
		target, err := os.Readlink(filepath.Join(l.DestDir, e.Name()))
		if err != nil {
			continue
		}
		current[e.Name()] = target
	}

	var results []LinkResult

	// Creates and updates: iterate desired.
	for name, target := range desired {
		linkPath := filepath.Join(l.DestDir, name)
		existing, exists := current[name]

		if exists && existing == target {
			results = append(results, LinkResult{Name: name, Target: target, Action: "skip"})
			continue
		}

		verb := "create"
		if exists {
			verb = "update"
		}

		if !l.Dry {
			if exists {
				if err := os.Remove(linkPath); err != nil {
					results = append(results, LinkResult{Name: name, Err: fmt.Errorf("remove old symlink: %w", err)})
					continue
				}
			}
			if err := os.Symlink(target, linkPath); err != nil {
				results = append(results, LinkResult{Name: name, Err: fmt.Errorf("symlink: %w", err)})
				continue
			}
		}
		results = append(results, LinkResult{Name: name, Target: target, Action: verb})
	}

	// Removals: symlinks in current but not in desired.
	for name, target := range current {
		if _, want := desired[name]; want {
			continue
		}
		linkPath := filepath.Join(l.DestDir, name)
		if !l.Dry {
			if err := os.Remove(linkPath); err != nil {
				results = append(results, LinkResult{Name: name, Err: fmt.Errorf("remove: %w", err)})
				continue
			}
		}
		results = append(results, LinkResult{Name: name, Target: target, Action: "remove"})
	}

	return results
}

// PrintResults prints a summary of link results to stdout.
func PrintResults(results []LinkResult) {
	for _, r := range results {
		if r.Err != nil {
			fmt.Printf("  error    %s: %v\n", r.Name, r.Err)
		} else {
			fmt.Printf("  %-12s %s → %s\n", r.Action, r.Name, r.Target)
		}
	}
}
