package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	gz "github.com/hayeah/godzkilla"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync skills from sources into a destination directory",
	Long: `Sync makes the destination directory match the combined set of sources
exactly. It adds missing symlinks, updates changed ones, and removes
symlinks that no longer belong to any source.

Unlike install (which is additive-only), sync removes stale links.

Examples:
  gozkilla sync --destination ~/.claude/skills --source github.com/hayeah/skills
  gozkilla sync --destination ~/.claude/skills --source github.com/hayeah/skills --source ./local-skills
  gozkilla sync --destination ~/.claude/skills --source github.com/hayeah/skills --dry
`,
	RunE: runSync,
}

var (
	syncSources []string
	syncDest    string
	syncDry     bool
)

func init() {
	syncCmd.Flags().StringArrayVar(&syncSources, "source", nil, "skill source (repeatable; GitHub path or local directory)")
	syncCmd.Flags().StringVar(&syncDest, "destination", "", "destination directory for skill symlinks")
	syncCmd.Flags().BoolVar(&syncDry, "dry", false, "dry run — show what would happen without making changes")
	_ = syncCmd.MarkFlagRequired("source")
	_ = syncCmd.MarkFlagRequired("destination")
}

func runSync(cmd *cobra.Command, args []string) error {
	// desired maps skill name → absolute target path.
	desired := map[string]string{}

	for _, src := range syncSources {
		resolved, err := gz.Resolve(src)
		if err != nil {
			return fmt.Errorf("resolving %s: %w", src, err)
		}

		if err := resolved.EnsureCloned(); err != nil {
			return fmt.Errorf("cloning %s: %w", src, err)
		}

		if _, err := os.Stat(resolved.LocalDir); err != nil {
			return fmt.Errorf("source directory not found: %s", resolved.LocalDir)
		}

		skills, err := resolved.FindSkills()
		if err != nil {
			return fmt.Errorf("finding skills in %s: %w", src, err)
		}

		baseName := resolved.Name
		for _, s := range skills {
			name := gz.SkillName(baseName, s.RelPath)
			absTarget, err := filepath.Abs(s.SkillDir)
			if err != nil {
				return fmt.Errorf("abs path for %s: %w", s.SkillDir, err)
			}
			desired[name] = absTarget
		}

		fmt.Printf("found %d skill(s) in %s (base name: %s)\n", len(skills), resolved.LocalDir, baseName)
	}

	if syncDry {
		fmt.Println("\ndry run:")
	}

	linker := gz.Linker{DestDir: syncDest, Dry: syncDry}
	results := linker.Sync(desired)

	// Sort results for stable output.
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	gz.PrintResults(results)

	errCount := 0
	for _, r := range results {
		if r.Err != nil {
			errCount++
		}
	}
	if errCount > 0 {
		return fmt.Errorf("%d error(s) during sync", errCount)
	}
	return nil
}
