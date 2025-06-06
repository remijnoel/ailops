package cmd

import (
	"fmt"
	"os"

	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/remijnoel/ailops/llm"
	"github.com/remijnoel/ailops/models"
	"github.com/remijnoel/ailops/workflow"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var op = llm.NewOpenAIProvider(
	os.Getenv("OPENAI_API_KEY"),
	"You are a Linux system assistant. Analyze the following system diagnostics and provide a clear, concise summary of system health, notable issues, and recommended actions.",
	llm.OPENAI_GPT41_Mini, // Using a smaller model for faster response times
)


var debugCmd = &cobra.Command{
	Use:   "diagnose",
	Short: "Debug host with targeted diagnostics and LLM analysis",
	Run: func(cmd *cobra.Command, args []string) {

		interactive, _ := cmd.Flags().GetBool("interactive")
		remote, _ := cmd.Flags().GetString("remote")
		useSudo, _ := cmd.Flags().GetBool("sudo")

		description, _ := cmd.Flags().GetString("description")
		if description == "" {
			log.Warn("No description provided, using default")

		}

		log.Debug("Log env prefix: ", viper.GetEnvPrefix())

		whitelist := viper.GetStringSlice("cmd_whitelist")
		log.Debug("Command whitelist: ", whitelist)
		blacklist := viper.GetStringSlice("cmd_blacklist")
		log.Debug("Command blacklist: ", blacklist)

		// Define commands to run for debugging the host
		commands := []string{
			"top -b -n1 | head -20",
			"ps aux | head -10",
			"df -h",
			"free -h",
			"dmesg | tail -n 50",
		}

		log.Debug("Running commands: ", commands)

		session := workflow.DebugWorkflow(description, &models.DebugSessionConfig{
			FirstCommands:    commands,
			Remote:           remote,
			UseSudo:          useSudo,
			CommandWhitelist: whitelist,
			CommandBlacklist: blacklist,
		}, interactive, op)
		rendered := markdown.Render(session.Summary, 100, 2)
		fmt.Println(string(rendered))
	},
}

func init() {
	debugCmd.Flags().StringP("description", "d", "", "Description of the issue to debug")
	debugCmd.MarkFlagRequired("description")
	debugCmd.Flags().BoolP("interactive", "i", false, "Run in interactive mode (default: false)")
	debugCmd.Flags().StringP("remote", "r", "", "Execute commands on a remote host (ssh format 'user@host') instead of locally")
	debugCmd.Flags().BoolP("sudo", "s", false, "Run all commands with sudo (default: false)")
}
