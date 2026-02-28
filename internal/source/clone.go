package source

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// EnsureCloned clones src into localDir using a treeless partial clone if the
// directory does not already exist. If localDir already contains a git repo,
// it is left untouched (idempotent).
func EnsureCloned(src, localDir string) error {
	// Check if already cloned (presence of .git dir is sufficient).
	gitDir := filepath.Join(localDir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return nil // already present, nothing to do
	}

	if err := os.MkdirAll(filepath.Dir(localDir), 0o755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	url := ToHTTPS(src)
	fmt.Printf("cloning %s → %s\n", url, localDir)

	cmd := exec.Command("git", "clone", "--filter=tree:0", url, localDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone %s: %w", url, err)
	}
	return nil
}

// Fetch runs `git fetch --all` in localDir to update a previously cloned repo.
func Fetch(localDir string) error {
	fmt.Printf("fetching %s\n", localDir)
	cmd := exec.Command("git", "-C", localDir, "fetch", "--all")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git fetch in %s: %w", localDir, err)
	}

	// Fast-forward the current branch.
	cmd = exec.Command("git", "-C", localDir, "merge", "--ff-only", "@{u}")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git merge in %s: %w", localDir, err)
	}
	return nil
}
