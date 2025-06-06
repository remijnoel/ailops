package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ailops",
	Short: "A sysadmin assistant powered by LLMs",
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.AddCommand(debugHostCmd)
	rootCmd.AddCommand(debugCmd)
}
