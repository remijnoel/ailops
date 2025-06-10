package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ailops",
	Short: "A sysadmin assistant powered by LLMs",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        configPath, _ := cmd.Flags().GetString("config")
        config(configPath)
        return nil
    },
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
    rootCmd.AddCommand(debugCmd)
    rootCmd.PersistentFlags().StringP("config", "c", "", "Path to configuration file")
    rootCmd.PersistentFlags().Bool("debug", false, "Enable verbose logging")
}