package gozkilla

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// EnsureCloned clones the remote source using a treeless partial clone if the
// repo directory does not already exist. When SubPath is non-empty, sparse
// checkout is used so only that subtree is materialised on disk.
// No-op for local sources.
func (r *Resolved) EnsureCloned() error {
	if !r.Remote {
		return nil
	}

	gitDir := filepath.Join(r.RepoDir, ".git")

	if _, err := os.Stat(gitDir); err != nil {
		// Fresh clone.
		if err := os.MkdirAll(filepath.Dir(r.RepoDir), 0o755); err != nil {
			return fmt.Errorf("create parent dir: %w", err)
		}

		url := toHTTPS(r.RepoPath)
		fmt.Printf("cloning %s → %s\n", url, r.RepoDir)

		if r.SubPath != "" {
			return cloneSparse(url, r.RepoDir, r.SubPath)
		}
		return cloneFull(url, r.RepoDir)
	}

	// Already cloned — ensure subpath is checked out if needed.
	if r.SubPath == "" {
		return nil
	}
	if _, err := os.Stat(filepath.Join(r.RepoDir, r.SubPath)); err == nil {
		return nil // subpath already present in working tree
	}
	// Add subpath to sparse checkout (works for repos already in sparse mode).
	fmt.Printf("adding sparse checkout path: %s\n", r.SubPath)
	return gitRun(r.RepoDir, "sparse-checkout", "add", r.SubPath)
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

// Fetch runs `git fetch --all` + fast-forward merge on the cloned repo.
func (r *Resolved) Fetch() error {
	fmt.Printf("fetching %s\n", r.RepoDir)
	cmd := exec.Command("git", "-C", r.RepoDir, "fetch", "--all")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git fetch in %s: %w", r.RepoDir, err)
	}

	cmd = exec.Command("git", "-C", r.RepoDir, "merge", "--ff-only", "@{u}")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git merge in %s: %w", r.RepoDir, err)
	}
	return nil
}
