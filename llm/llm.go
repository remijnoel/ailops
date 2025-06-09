package llm

import (
	"fmt"
)

type Model struct {
	Name       string
	ContextSize int // in tokens
	Options   map[string]string
}

// Define the interface
type Provider interface {
    RequestCompletion(string) (string, error)
	RequestCompletionWithJSONSchema(string,  interface{}) (string, error)
}

func AnalyzeCommands(results map[string]string, provider Provider) string {
    cmdOutput := ""
    for cmd, out := range results {
        cmdOutput += fmt.Sprintf("Command: %s\nOutput:\n%s\n\n", cmd, out)
    }
    
	res, err := provider.RequestCompletion(cmdOutput)
	if err != nil {
		return fmt.Sprintf("Error analyzing commands: %v", err)
	}
	
	return res
}