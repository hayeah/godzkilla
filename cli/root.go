package cli

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "gozkilla",
	Short: "Skill management CLI — find, install, and update AI agent skills",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(updateCmd)
}
