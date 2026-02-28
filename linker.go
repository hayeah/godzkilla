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

// SyncAll makes destDir match desired exactly: creates missing links,
// updates changed ones, and removes links not present in desired.
// When dry is true it reports what would happen without touching the filesystem.
func SyncAll(desired map[string]string, destDir string, dry bool) []LinkResult {
	if !dry {
		if err := os.MkdirAll(destDir, 0o755); err != nil {
			return []LinkResult{{Err: fmt.Errorf("create dest dir: %w", err)}}
		}
	}

	// Scan current symlinks in destDir.
	current := map[string]string{} // name → readlink target
	entries, err := os.ReadDir(destDir)
	if err != nil && !os.IsNotExist(err) {
		return []LinkResult{{Err: fmt.Errorf("read dest dir: %w", err)}}
	}
	for _, e := range entries {
		if e.Type()&os.ModeSymlink == 0 {
			continue
		}
		target, err := os.Readlink(filepath.Join(destDir, e.Name()))
		if err != nil {
			continue
		}
		current[e.Name()] = target
	}

	var results []LinkResult

	// Creates and updates: iterate desired.
	for name, target := range desired {
		linkPath := filepath.Join(destDir, name)
		existing, exists := current[name]

		if exists && existing == target {
			results = append(results, LinkResult{Name: name, Target: target, Action: actionName("skip", dry)})
			continue
		}

		verb := "create"
		if exists {
			verb = "update"
		}

		if dry {
			results = append(results, LinkResult{Name: name, Target: target, Action: actionName(verb, dry)})
			continue
		}

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
		results = append(results, LinkResult{Name: name, Target: target, Action: actionName(verb, dry)})
	}

	// Removals: symlinks in current but not in desired.
	for name, target := range current {
		if _, want := desired[name]; want {
			continue
		}
		linkPath := filepath.Join(destDir, name)
		if dry {
			results = append(results, LinkResult{Name: name, Target: target, Action: actionName("remove", dry)})
			continue
		}
		if err := os.Remove(linkPath); err != nil {
			results = append(results, LinkResult{Name: name, Err: fmt.Errorf("remove: %w", err)})
			continue
		}
		results = append(results, LinkResult{Name: name, Target: target, Action: actionName("remove", dry)})
	}

	return results
}

// actionName returns a display string for the given verb.
// In dry mode: "would create"; otherwise: "created".
func actionName(verb string, dry bool) string {
	if dry {
		return "would " + verb
	}
	switch verb {
	case "skip":
		return "skipped"
	case "create":
		return "created"
	case "update":
		return "updated"
	case "remove":
		return "removed"
	}
	return verb + "d"
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
