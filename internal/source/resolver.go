package source

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

// ToHTTPS converts a bare host path like "github.com/hayeah/skills"
// to a full HTTPS clone URL.
func ToHTTPS(src string) string {
	return "https://" + src + ".git"
}

// StorageName converts a source path to a flat filesystem-safe name
// by replacing "/" with "_".
//
//	github.com/hayeah/skills → github.com_hayeah_skills
func StorageName(src string) string {
	return strings.ReplaceAll(src, "/", "_")
}

// StorageDir returns the local directory where a remote source should
// be cloned. It respects the SKILLA_PATH env var; if unset it defaults
// to ~/.skilla.
func StorageDir(src string) (string, error) {
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
	return filepath.Join(base, StorageName(src)), nil
}

// Resolved holds the result of resolving a source identifier.
type Resolved struct {
	// LocalDir is the local filesystem path (cloned or as-is for local sources).
	LocalDir string
	// Remote is true when src was a remote host path.
	Remote bool
	// Name is the flat, filesystem-safe base name for skill naming.
	// For remote sources this is derived from the full path
	// (e.g. "github.com_hayeah_skills"). For local sources it is the
	// basename of the directory.
	Name string
}

// Resolve returns a Resolved for src.
//   - If src is remote, localDir is the storage dir (may not exist yet).
//   - If src is local, localDir is the absolute path.
func Resolve(src string) (Resolved, error) {
	if IsRemote(src) {
		dir, err := StorageDir(src)
		if err != nil {
			return Resolved{}, err
		}
		return Resolved{LocalDir: dir, Remote: true, Name: StorageName(src)}, nil
	}
	abs, err := filepath.Abs(src)
	if err != nil {
		return Resolved{}, err
	}
	return Resolved{LocalDir: abs, Remote: false, Name: filepath.Base(abs)}, nil
}
