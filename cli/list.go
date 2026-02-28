package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed skills in a destination directory",
	Long: `List shows all symlinks in the destination directory that point to a
directory containing a SKILL.md file.

Example:
  gozkilla list --destination ~/.claude/skills
`,
	RunE: runList,
}

var listDest string

func init() {
	listCmd.Flags().StringVar(&listDest, "destination", "", "destination directory to inspect")
	_ = listCmd.MarkFlagRequired("destination")
}

func runList(cmd *cobra.Command, args []string) error {
	entries, err := os.ReadDir(listDest)
	if os.IsNotExist(err) {
		fmt.Printf("destination %s does not exist\n", listDest)
		return nil
	}
	if err != nil {
		return err
	}

	count := 0
	for _, e := range entries {
		if e.Type()&os.ModeSymlink == 0 {
			continue
		}
		linkPath := filepath.Join(listDest, e.Name())
		target, err := os.Readlink(linkPath)
		if err != nil {
			continue
		}
		if _, err := os.Stat(filepath.Join(target, "SKILL.md")); err != nil {
			continue
		}
		fmt.Printf("  %-50s → %s\n", e.Name(), target)
		count++
	}

	if count == 0 {
		fmt.Println("no skills installed")
	} else {
		fmt.Printf("\n%d skill(s) installed\n", count)
	}
	return nil
}
