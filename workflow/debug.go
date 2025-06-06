package workflow

import (
	"fmt"
	"strings"
	"time"

	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/remijnoel/ailops/internal"
	"github.com/remijnoel/ailops/llm"
	"github.com/remijnoel/ailops/models"
	"github.com/remijnoel/ailops/ui"
	log "github.com/sirupsen/logrus"
)

func RunLastBatch(session *models.DebugSessionLog, llmProvider llm.Provider) {
	batch := session.LastBatch()
	log.Infof("Running batch: %s", batch.Description)
	// For now, let's consider all Actions as commands
	actions := batch.Actions
	// Run all commands in parallel and update the actions with the results

	RunCommands(actions)

	// Include all analysis history and the commands output in the prompt
	prompt := CommandAnalysisPrompt(session, true, true)

	// Analyze the results using the LLM provider
	commandAnalysis := AnalyzeCommands(prompt, llmProvider)
	batch.Analysis = commandAnalysis.Analysis
	batch.NextSteps = commandAnalysis.Recommendations
	batch.Completed = true // Mark the batch as completed after analysis

	if commandAnalysis.Final {
		log.Infof("Final analysis completed for batch: %s", batch.Description)
		session.Diagnosed = true // Mark the session as diagnosed
		session.EndTime = time.Now().Format(time.RFC3339)
	} else {
		log.Infof("Batch %s analysis completed, but more steps are needed.", batch.Description)
	}
}

func Init(conf WorkflowConfig) *models.DebugSessionLog {
	log.Infof("Initializing debug session with issue: %s", conf.IssueDescription)
	log.Infof("First commands to run: %v", conf.FirstCommands)

	actions := make([]*models.Action, 0, len(conf.FirstCommands))
	for _, cmd := range conf.FirstCommands {
		action := &models.Action{
			Name:       cmd,
			ActionType: "command",
			Status:     "new",
			Result:     "",
			Remote:     conf.Remote, // Set the remote host if applicable
		}
		actions = append(actions, action)
	}

	batch := &models.Batch{
		Description: "Initial commands",
		Actions:     actions,
		NextSteps:   []string{},
		Completed:   false,
	}

	sessionLog := &models.DebugSessionLog{
		ID:               internal.GenerateUniqueID(),
		StartTime:        time.Now().Format(time.RFC3339),
		Batches:          []*models.Batch{batch},
		IssueDescription: conf.IssueDescription,
	}

	return sessionLog
}

func PrepareNextBatch(sessionLog *models.DebugSessionLog, nextCmds []string) {
	log.Infof("Preparing next batch with commands: %v", nextCmds)
	nextActions := make([]*models.Action, 0, len(nextCmds))
	for _, cmd := range nextCmds {
		action := &models.Action{
			Name:       cmd,
			ActionType: "command",
			Status:     "new",
			Result:     "",
		}
		nextActions = append(nextActions, action)
	}

	sessionLog.AddBatch(&models.Batch{
		Description: "Follow-up commands",
		Actions:     nextActions,
		NextSteps:   []string{},
		Completed:   false,
	})
}

func FinalAnalysis(sessionLog *models.DebugSessionLog, llmProvider llm.Provider) {
	log.Infof("Performing final analysis of the session log with ID: %s", sessionLog.ID)
	// Get the analysis from each batch
	analysis := ""
	for _, batch := range sessionLog.Batches {
		if batch.Completed {
			analysis += batch.Analysis + "\n"
		}
	}

	log.Debugf("Final analysis of batches: %s", analysis)
	// Use the LLM provider to analyze the overall session log
	response, err := llmProvider.RequestCompletion(FinalAnalysisPromptWithSessionLog(sessionLog))
	if err != nil {
		log.Errorf("Error during final analysis: %v", err)
		response = "Error during final analysis: " + err.Error()
	}
	log.Infof("Final analysis response: %s", response)
	sessionLog.Summary = response
}

type WorkflowConfig struct {
	IssueDescription string   `json:"issue_description"`
	FirstCommands    []string `json:"first_commands"`
	Remote           string   `json:"remote"`
	UseSudo          bool     `json:"use_sudo"`
}

func DebugWorkflow(conf WorkflowConfig, interactive bool, llmProvider llm.Provider) *models.DebugSessionLog {
	sessionLog := Init(conf)

	// For now, loop 5 times to simulate multiple batches
	for range 5 {
		// Get the last batch to run commands and analyze
		currentBatch := sessionLog.LastBatch()

		ui.RunWithSpinner(interactive, "Running commands and analyzing output", func() {
			RunLastBatch(sessionLog, llmProvider)
		})

		if interactive {
			var content strings.Builder
			content.WriteString("**Commands:**\n")
			for _, action := range currentBatch.Actions {
				content.WriteString("- " + action.Name + "\n")
			}
			content.WriteString("\n**Analysis:**\n")
			content.WriteString(currentBatch.Analysis + "\n\n")
			content.WriteString("**Next Steps:**\n")
			for _, cmd := range currentBatch.NextSteps {
				content.WriteString("- " + cmd + "\n")
			}

			rendered := string(markdown.Render(content.String(), 100, 2))
			fmt.Print(rendered)
			fmt.Println()
			fmt.Printf("Do you want to continue with the next batch of commands? (yes/no): ")
			var response string
			fmt.Scanln(&response)
			if response != "yes" {
				fmt.Println("Ending debug session.")
				sessionLog.EndSession()
				return sessionLog
			}
		}

		if len(currentBatch.NextSteps) == 0 {
			if interactive {
				fmt.Println("No next steps provided, ending debug session...")
			}
			break // No next steps, jump to the final analysis directly
		}

		ui.RunWithSpinner(interactive, "Preparing next batch of commands", func() {
			PrepareNextBatch(sessionLog, currentBatch.NextSteps)
			time.Sleep(2 * time.Second)
		})
	}

	ui.RunWithSpinner(interactive, "Performing final analysis", func() {
		FinalAnalysis(sessionLog, llmProvider)
	})

	return sessionLog
}
