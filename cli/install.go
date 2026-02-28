package cli

import (
	"fmt"
	"os"

	gz "github.com/hayeah/gozkilla"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install skills from a source into a destination directory",
	Long: `Install clones the source (if it is a remote GitHub path) and creates
symlinks for every discovered SKILL.md into the destination directory.

The operation is idempotent: running install again skips symlinks that
already point to the correct target.

Examples:
  gozkilla install --source github.com/hayeah/skills --destination ~/.claude/skills
  gozkilla install --source ./my-local-skills         --destination ~/.claude/skills
  gozkilla install --source ./my-local-skills --name myskills --destination ~/.claude/skills
`,
	RunE: runInstall,
}

var (
	installSource string
	installDest   string
	installName   string
)

func init() {
	installCmd.Flags().StringVar(&installSource, "source", "", "skill source (GitHub path or local directory)")
	installCmd.Flags().StringVar(&installDest, "destination", "", "destination directory for skill symlinks")
	installCmd.Flags().StringVar(&installName, "name", "", "override base name for skill naming (default: derived from source)")
	_ = installCmd.MarkFlagRequired("source")
	_ = installCmd.MarkFlagRequired("destination")
}

func runInstall(cmd *cobra.Command, args []string) error {
	resolved, err := gz.Resolve(installSource)
	if err != nil {
		return err
	}

	baseName := resolved.Name
	if installName != "" {
		baseName = installName
	}

	if resolved.Remote {
		if err := gz.EnsureCloned(installSource, resolved.LocalDir); err != nil {
			return err
		}
	}

	if _, err := os.Stat(resolved.LocalDir); err != nil {
		return fmt.Errorf("source directory not found: %s", resolved.LocalDir)
	}

	skills, err := gz.FindAll(resolved.LocalDir)
	if err != nil {
		return fmt.Errorf("finding skills: %w", err)
	}
	if len(skills) == 0 {
		fmt.Println("no SKILL.md files found in source")
		return nil
	}

	fmt.Printf("found %d skill(s) in %s (base name: %s)\n", len(skills), resolved.LocalDir, baseName)
	results := gz.InstallAll(baseName, skills, installDest)
	gz.PrintResults(results)

	errCount := 0
	for _, r := range results {
		if r.Err != nil {
			errCount++
		}
	}
	if errCount > 0 {
		return fmt.Errorf("%d error(s) during install", errCount)
	}
	return nil
}
