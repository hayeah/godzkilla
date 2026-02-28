package gozkilla

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// EnsureCloned clones repoPath into repoDir using a treeless partial clone
// if the directory does not already exist. When subPath is non-empty, sparse
// checkout is used so only that subtree is materialised on disk.
func EnsureCloned(repoPath, repoDir, subPath string) error {
	gitDir := filepath.Join(repoDir, ".git")

	if _, err := os.Stat(gitDir); err != nil {
		// Fresh clone.
		if err := os.MkdirAll(filepath.Dir(repoDir), 0o755); err != nil {
			return fmt.Errorf("create parent dir: %w", err)
		}

		url := toHTTPS(repoPath)
		fmt.Printf("cloning %s → %s\n", url, repoDir)

		if subPath != "" {
			return cloneSparse(url, repoDir, subPath)
		}
		return cloneFull(url, repoDir)
	}

	// Already cloned — ensure subpath is checked out if needed.
	if subPath == "" {
		return nil
	}
	if _, err := os.Stat(filepath.Join(repoDir, subPath)); err == nil {
		return nil // subpath already present in working tree
	}
	// Add subpath to sparse checkout (works for repos already in sparse mode).
	fmt.Printf("adding sparse checkout path: %s\n", subPath)
	return gitRun(repoDir, "sparse-checkout", "add", subPath)
}

func cloneFull(url, dir string) error {
	cmd := exec.Command("git", "clone", "--filter=tree:0", url, dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone %s: %w", url, err)
	}
	return nil
}

func cloneSparse(url, dir, subPath string) error {
	cmd := exec.Command("git", "clone", "--filter=tree:0", "--sparse", url, dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone --sparse %s: %w", url, err)
	}
	if err := gitRun(dir, "sparse-checkout", "add", subPath); err != nil {
		return fmt.Errorf("sparse-checkout add %s: %w", subPath, err)
	}
	return nil
}

func gitRun(dir string, args ...string) error {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Fetch runs `git fetch --all` + fast-forward merge in localDir.
func Fetch(localDir string) error {
	fmt.Printf("fetching %s\n", localDir)
	cmd := exec.Command("git", "-C", localDir, "fetch", "--all")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git fetch in %s: %w", localDir, err)
	}

	cmd = exec.Command("git", "-C", localDir, "merge", "--ff-only", "@{u}")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git merge in %s: %w", localDir, err)
	}
	return nil
}
