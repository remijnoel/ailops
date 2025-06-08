package cmd

import (
	"fmt"
	"os"

	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/remijnoel/ailops/llm"
	"github.com/remijnoel/ailops/models"
	"github.com/remijnoel/ailops/report"
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
	Short: "Diagnose an issue on a host",
	Run: func(cmd *cobra.Command, args []string) {

		interactive, _ := cmd.Flags().GetBool("interactive")
		remote, _ := cmd.Flags().GetString("remote")
		useSudo, _ := cmd.Flags().GetBool("sudo")
		generateReport, _ := cmd.Flags().GetBool("generate-report")

		description, _ := cmd.Flags().GetString("description")
		if description == "" {
			log.Warn("No description provided, using default")

		}

		whitelist := viper.GetStringSlice("cmd_whitelist")
		log.Debug("Command whitelist: ", whitelist)
		blacklist := viper.GetStringSlice("cmd_blacklist")
		log.Debug("Command blacklist: ", blacklist)

		// Define commands to run for debugging the host
		commands := viper.GetStringSlice("initial_commands")
		log.Debug("Initial commands from config: ", commands)

		session := workflow.DebugWorkflow(description, &models.DebugSessionConfig{
			FirstCommands:    commands,
			Remote:           remote,
			UseSudo:          useSudo,
			CommandWhitelist: whitelist,
			CommandBlacklist: blacklist,
		}, interactive, op)

		rendered := markdown.Render(session.Summary, 100, 2)
		fmt.Println(string(rendered))

		if generateReport {
			// For now always use markdown for reports
			reportConfig := report.ReportConfig{
				Format:                 report.Markdown,
				IncludeCommandOutput:   true,
				IncludeAnalysisHistory: true,
			}
			report := report.GenerateReport(session, reportConfig)

			// If it does not exist, create the reports directory named .ailops
			if _, err := os.Stat(".ailops"); os.IsNotExist(err) {
				err := os.Mkdir(".ailops", 0755)
				if err != nil {
					log.Fatalf("Failed to create .ailops directory: %v", err)
				}
			}

			reportFile := fmt.Sprintf(".ailops/debug_report_%s.md", session.ID)
			err := os.WriteFile(reportFile, []byte(report), 0644)
			if err != nil {
				log.Fatalf("Failed to write report file: %v", err)
			}

		}
	},
}

func init() {
	debugCmd.Flags().StringP("description", "d", "", "Description of the issue to debug")
	debugCmd.MarkFlagRequired("description")
	debugCmd.Flags().BoolP("interactive", "i", false, "Run in interactive mode (default: false)")
	debugCmd.Flags().StringP("remote", "r", "", "Execute commands on a remote host (ssh format 'user@host') instead of locally")
	debugCmd.Flags().BoolP("sudo", "s", false, "Run all commands with sudo (default: false)")
	debugCmd.Flags().BoolP("generate-report", "g", false, "Generate a report after debugging (default: false)")
}
