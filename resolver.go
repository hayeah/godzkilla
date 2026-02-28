package gozkilla

import (
	"os"
	"path/filepath"
	"strings"
)

// knownHosts are URL prefixes that identify remote git sources.
var knownHosts = []string{
	"github.com/",
	"gitlab.com/",
	"bitbucket.org/",
	"codeberg.org/",
}

// IsRemote reports whether src looks like a remote git host path
// (e.g. "github.com/hayeah/skills").
func IsRemote(src string) bool {
	for _, h := range knownHosts {
		if strings.HasPrefix(src, h) {
			return true
		}
	}
	return false
}

// toHTTPS converts a bare host path like "github.com/hayeah/skills"
// to a full HTTPS clone URL.
func toHTTPS(src string) string {
	return "https://" + src + ".git"
}

// storageName converts a source path to a flat filesystem-safe name
// by replacing "/" with "_".
//
//	github.com/hayeah/skills → github.com_hayeah_skills
func storageName(src string) string {
	return strings.ReplaceAll(src, "/", "_")
}

// storageDir returns the local directory where a remote source should
// be cloned. It respects the SKILLA_PATH env var; if unset it defaults
// to ~/.skilla.
func storageDir(src string) (string, error) {
	base := os.Getenv("SKILLA_PATH")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".skilla")
	} else {
		// Expand ~ manually since os.Getenv won't do it.
		if strings.HasPrefix(base, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			base = filepath.Join(home, base[2:])
		}
	}
	return filepath.Join(base, storageName(src)), nil
}

// ParseRemote splits a remote source path into the repo path and an
// optional subpath within that repo.
//
//	"github.com/user/repo"         → ("github.com/user/repo", "")
//	"github.com/user/repo/foo"     → ("github.com/user/repo", "foo")
//	"github.com/user/repo/foo/bar" → ("github.com/user/repo", "foo/bar")
func ParseRemote(src string) (repoPath, subPath string) {
	parts := strings.Split(src, "/")
	if len(parts) <= 3 {
		return src, ""
	}
	return strings.Join(parts[:3], "/"), strings.Join(parts[3:], "/")
}

// Resolved holds the result of resolving a source identifier.
type Resolved struct {
	// LocalDir is the directory to walk for skills.
	// For remote sources with a subpath this is RepoDir/SubPath.
	LocalDir string
	// Remote is true when src was a remote host path.
	Remote bool
	// Name is the flat, filesystem-safe base name for skill naming.
	Name string
	// RepoPath is the repo portion of the source (e.g. "github.com/user/repo").
	// Empty for local sources.
	RepoPath string
	// RepoDir is the local clone directory for the repo root.
	// Empty for local sources.
	RepoDir string
	// SubPath is the path within the repo (e.g. "foo/bar"). Empty if the
	// source points to the whole repo.
	SubPath string
}

// Resolve returns a Resolved for src.
//   - If src is remote, LocalDir is the storage dir (may not exist yet).
//   - If src is local, LocalDir is the absolute path.
func Resolve(src string) (Resolved, error) {
	if IsRemote(src) {
		repoPath, subPath := ParseRemote(src)
		repoDir, err := storageDir(repoPath)
		if err != nil {
			return Resolved{}, err
		}
		localDir := repoDir
		if subPath != "" {
			localDir = filepath.Join(repoDir, subPath)
		}
		return Resolved{
			LocalDir: localDir,
			Remote:   true,
			Name:     storageName(src),
			RepoPath: repoPath,
			RepoDir:  repoDir,
			SubPath:  subPath,
		}, nil
	}
	abs, err := filepath.Abs(src)
	if err != nil {
		return Resolved{}, err
	}
	return Resolved{LocalDir: abs, Remote: false, Name: filepath.Base(abs)}, nil
}
