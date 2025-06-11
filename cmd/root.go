package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "ailops",
	Short: "A sysadmin assistant powered by LLMs",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		configPath, _ := cmd.Flags().GetString("config")
		config(configPath)
		return nil
	},
}


func Execute() {
	cobra.CheckErr(RootCmd.Execute())
}

func init() {
	RootCmd.PersistentFlags().StringP("config", "c", "", "Path to configuration file")
	RootCmd.PersistentFlags().Bool("debug", false, "Enable verbose logging")
}
