package cli

import (
	"fmt"
	"os"

	gz "github.com/hayeah/godzkilla"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a previously cloned remote source",
	Long: `Update runs git fetch + fast-forward merge on a remote source that was
previously cloned by the install command.

Example:
  gozkilla update --source github.com/hayeah/skills
`,
	RunE: runUpdate,
}

var updateSource string

func init() {
	updateCmd.Flags().StringVar(&updateSource, "source", "", "remote source to update (GitHub path)")
	_ = updateCmd.MarkFlagRequired("source")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if !gz.IsRemote(updateSource) {
		return fmt.Errorf("update only applies to remote sources; use git pull for local repos")
	}

	resolved, err := gz.Resolve(updateSource)
	if err != nil {
		return err
	}

	if _, err := os.Stat(resolved.RepoDir); os.IsNotExist(err) {
		return fmt.Errorf("source not yet cloned; run install first: %s", resolved.RepoDir)
	}

	return resolved.Fetch()
}
