package workflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"

	// "net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/remijnoel/ailops/internal"
	"github.com/remijnoel/ailops/llm"
	"github.com/remijnoel/ailops/models"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	// "golang.org/x/crypto/ssh/agent"
)

// Helper to skip public key files
func isPublicKey(file string) bool {
	return filepath.Ext(file) == ".pub"
}

// Returns a slice of AuthMethod that mimics "ssh" CLI (agent + all ~/.ssh/* keys)
func defaultAuthMethods() []ssh.AuthMethod {
	var methods []ssh.AuthMethod

	// // 1. Try SSH agent, if running
	// if sock := os.Getenv("SSH_AUTH_SOCK"); sock != "" {
	// 	conn, err := net.Dial("unix", sock)
	// 	if err == nil {
	// 		methods = append(methods, ssh.PublicKeysCallback(agent.NewClient(conn).Signers))
	// 		log.Debugf("SSH agent found at %s, using its keys", sock)
	// 	} else {
	// 		log.Infof("SSH agent not available or error connecting: %v", err)
	// 	}
	// }

	// 2. Try all private keys in ~/.ssh/
	home, err := os.UserHomeDir()
	if err == nil {
		files, _ := filepath.Glob(filepath.Join(home, ".ssh", "id_*"))

		for _, file := range files {
			key, err := os.ReadFile(file)
			if err != nil || isPublicKey(file) {
				log.Debugf("Public key found. Skipping file %s: %v", file, err)
				continue
			}
			signer, err := ssh.ParsePrivateKey(key)
			if err == nil {
				log.Debugf("Loaded SSH key: %s", file)
				methods = append(methods, ssh.PublicKeys(signer))
			}else {
				log.Warnf("Failed to parse SSH key %s: %v", file, err)
			}
		}
	}
	log.Debugf("Available SSH auth methods: %v", methods)
	return methods
}

func parseRemote(remote string) (user string, host string, port string, err error) {
	// Set default values
	user = "root"
	port = "22"
	// Parse user@host:port format
	parts := strings.Split(remote, "@")
	if len(parts) == 1 { // No user specified, leave user as default
		host = parts[0]
	} else if len(parts) == 2 {
		user = parts[0]
		host = parts[1]
	} else {
		log.Errorf("Invalid remote format for expecting user@host:port but got '%s'", remote)
		return "", "", "", fmt.Errorf("invalid remote format: %s", remote)
	}
	// Check if host contains port
	if strings.Contains(host, ":") {
		hostParts := strings.Split(host, ":")
		if len(hostParts) != 2 {
			log.Errorf("Invalid remote format for expecting user@host:port but got '%s'", remote)
			return "", "", "", fmt.Errorf("invalid remote format: %s", remote)
		}
		host = hostParts[0]
		port = hostParts[1]
	}

	return user, host, port, nil

}

func RunRemoteCommand(user string, host string, port string,  auth []ssh.AuthMethod, cmd string, wg *sync.WaitGroup, results chan<- RemoteResult) {
	defer wg.Done()
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            auth,        // Use default auth methods (SSH agent + keys)
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // NOTE: Insecure, use only for testing!
		Timeout:         5 * time.Second,
	}
	log.Infof("Connecting to %s@%s with command: %s", user, host, cmd)
	log.Debugf("SSH client config: %+v", config)
	for _, method := range config.Auth {
		log.Debugf("Available auth method: %T", method)
	}

	client, err := ssh.Dial("tcp", host+":22", config)
	if err != nil {
		results <- RemoteResult{
			Command: cmd,
			Output:  fmt.Sprintf("[%s] Failed to connect: %v", host, err),
			Error:   err,
			Host:    host,
		}

		return
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		results <- RemoteResult{
			Command: cmd,
			Output:  fmt.Sprintf("[%s] Failed to create session: %v", host, err),
			Error:   err,
			Host:    host,
		}
		return
	}
	defer session.Close()
	out, err := session.CombinedOutput(cmd)
	if err != nil {
		results <- RemoteResult{
			Command: cmd,
			Output:  fmt.Sprintf("[%s] %s\n[ERROR] %v", host, strings.TrimSpace(string(out)), err),
			Error:   err,
			Host:    host,
		}
		return
	}
	results <- RemoteResult{
		Command: cmd,
		Output:  string(out),
		Error:   nil,
		Host:    host,
	}
}

type RemoteResult struct {
	Command string
	Output  string
	Error   error
	Host    string
}

func IsCommandAllowed(command string, config *models.DebugSessionConfig) bool {
	if config == nil {
		log.Warn("No session config provided, allowing nothing.")
		return false // No restrictions if no config
	}

	if len(config.CommandWhitelist) > 0 {
		for _, allowed := range config.CommandWhitelist {
			matched, err := regexp.MatchString("^"+regexp.QuoteMeta(allowed)+"(\\s|$)", command)
			if err != nil {
				log.Warnf("Invalid regex pattern for whitelist command '%s': %v", allowed, err)
				continue
			}
			if matched {
				log.Debugf("Command '%s' is allowed by whitelist pattern '%s'", command, allowed)
				return true
			}
		}
		log.Warnf("Command '%s' is NOT allowed by whitelist", command)
		return false // Not in whitelist
	}

	if len(config.CommandBlacklist) > 0 {
		for _, disallowed := range config.CommandBlacklist {
			matched, err := regexp.MatchString("^"+regexp.QuoteMeta(disallowed)+"(\\s|$)", command)
			if err != nil {
				log.Warnf("Invalid regex pattern for blacklist command '%s': %v", disallowed, err)
				continue
			}
			if matched {
				log.Warnf("Command '%s' is disallowed by blacklist pattern '%s'", command, disallowed)
				return false // In blacklist
			}
		}
		log.Debugf("Command '%s' is allowed by default (not in blacklist)", command)
		return true // Not in blacklist
	}

	log.Debugf("No command restrictions configured, allowing command '%s'", command)
	return true // No restrictions, allow all commands

}



func RunCommands(actions []*models.Action) {
	log.Infof("Running commands in parallel for %d actions", len(actions))
	localCommands := make(map[string]*models.Action)
	remoteCommands := make(map[string]*models.Action)

	for _, action := range actions {
		if action.IsCommand() {
			if action.IsRemote() {
				remoteCommands[action.Name] = action
			} else {
				localCommands[action.Name] = action
			}
		} else {
			// Handle other action types if needed
			// For now, we just skip non-command actions
			log.Debugf("Skipping non-command action: %s", action.Name)
			continue
		}
	}

	// Run local commands in parallel
	var localCommandNames []string
	for name := range localCommands {
		localCommandNames = append(localCommandNames, name)
	}
	log.Debugf("Running commands: %v", localCommandNames)
	results := internal.RunCommandsParallelWithTimeout(localCommandNames, 15*time.Second) // Run commands in parallel with a timeout of 15 seconds

	// Run remote commands in parallel with SSH
	var wg sync.WaitGroup

	remoteResults := make(chan RemoteResult, len(remoteCommands))

	for _, action := range remoteCommands {
		user, host, port, err := parseRemote(action.Remote)
		if err != nil {
			log.Errorf("Failed to parse remote host %s for command %s: %v", action.Remote, action.Name, err)
			continue
		}

		wg.Add(1)
		auth := defaultAuthMethods() // Get default auth methods (SSH agent + keys)
		go RunRemoteCommand(user, host, port, auth, action.Name, &wg, remoteResults)
	}

	wg.Wait()
	close(remoteResults)


	// Update Action.Result with local command outputs
	for cmd, output := range results {
		if action, exists := localCommands[cmd]; exists {
			action.Result = output
			action.Status = "completed" // Update status to completed
		} else {
			log.Warnf("No action found for command: %s", cmd)
		}
	}

	// Update Action.Result with remote command outputs
	for cmd := range remoteResults {
		if action, exists := remoteCommands[cmd.Command]; exists {
			log.Infof("Updating action %s with remote output from %s", action.Name, cmd.Host)
			action.Result = cmd.Output
			action.Status = "completed" // Update status to completed
		} else {
			log.Warnf("No action found for remote command: %s", cmd)
		}
	}
}

var commandAnalysisPrompt = `You are a Linux system assistant. Your task is to analyze system diagnostic data, summarize system health, identify notable issues, and recommend further actions if needed.

Constraints:
- Context Window: All recommendations, summaries, and command selections must consider that the LLM has a limited context window.
- Be concise.
- Do not repeat already provided information.
- Avoid commands or outputs that produce excessive or redundant data.
- Tailor recommendations to maximize useful insight with minimal output.

Recommendations Rules:
{{if .Session.Config.CommandWhitelist }}
- Only suggest commands from the following whitelist:
{{range .Session.Config.CommandWhitelist }}
	- {{.}}
{{end}}
{{else if .Session.Config.CommandBlacklist }}
- Only suggest commands that are NOT in the following blacklist:
{{range .Session.Config.CommandBlacklist }}
	- {{.}}
{{end}}
{{else }}
- Only suggest commands that are safe and appropriate for the current debugging context.
{{end}}

- Only suggest up to 5 shell commands per batch.
- All commands must be read-only (do not alter system state).
- No interactive commands (avoid prompts, user input, or commands that run in a loop; use, for example, 'top -n 1' instead of 'top').
{{ if .Session.Config.UseSudo }}
- ALWAYS use ‘sudo’ in all commands.
{{else }}
- NEVER include ‘sudo’ in any command.
{{ end }}
- Do not repeat any commands already included in the debugging history.
- Each command should include a concise comment at the end explaining its purpose (e.g., ps aux # list processes).
- Commands must be executable as-is in a shell, without extra context or input.
- Only recommend commands when they add significant new diagnostic value.

Stopping Criteria:
- If you have identified the root cause with reasonable certainty, or have sufficient diagnostic evidence:
- Clearly state this in your analysis.
- Leave the recommendations array empty.
- Set the "final" property to true.
- Only continue recommending additional commands if further investigation is absolutely necessary.
- If so, set "final" to false.

Problem description: {{.Session.IssueDescription}}

Debugging history:
{{range .Session.Batches}}
	Batch: {{.Description}}
	Commands:
	{{range .Actions}}
		{{.Name}}
		{{if $.IncludeAllCommandOutputs}}
			Output: {{.Result}}
		{{end}}
	{{end}}
	{{if $.IncludeAllBatchAnalysis}}
		Analysis: {{.Analysis}}
	{{end}}
	------------
{{end}}`

type CommandAnalysisInput struct {
	Session                  *models.DebugSessionLog
	IncludeAllBatchAnalysis  bool
	IncludeAllCommandOutputs bool
}

func CommandAnalysisPrompt(session *models.DebugSessionLog, includeAllBatchAnalysis bool, includeAllCommandOutputs bool) string {
	// Use the template package to format the prompt
	tmpl := template.Must(template.New("commandAnalysis").Parse(commandAnalysisPrompt))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, CommandAnalysisInput{
		Session:                  session,
		IncludeAllBatchAnalysis:  includeAllBatchAnalysis,
		IncludeAllCommandOutputs: includeAllCommandOutputs,
	}); err != nil {
		log.Errorf("Error executing template: %v", err)
		return ""
	}
	return buf.String()
}

var FinalAnalysisPrompt = `You are a Linux system assistant. A debugging session occurred during which several rounds of debugging commands were issued and the outputs analyzed. Based on those analyses, provide your final analysis of the system state, your theories about the root cause of the issue, and any recommended next steps.

Problem description: {{.IssueDescription}}

Debugging history:

{{range .Batches}}
Batch description: {{.Description}}
Commands executed:
{{range .Actions}}
{{.Name}}
{{end}}
Analysis: {{.Analysis}}
--------------------
{{end}}`

func FinalAnalysisPromptWithSessionLog(sessionLog *models.DebugSessionLog) string {
	// Use the template package to format the final analysis prompt
	tmpl := template.Must(template.New("finalAnalysis").Parse(FinalAnalysisPrompt))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, sessionLog); err != nil {
		log.Errorf("Error executing final analysis template: %v", err)
		return ""
	}
	return buf.String()
}

type CommandAnalysisResponse struct {
	Analysis        string   `json:"analysis" jsonschema:"required" jsonschema_description:"Analysis of the command outputs within the context of the current issue"`
	Recommendations []string `json:"recommendations" jsonschema:"required" jsonschema_description:"Recommended list of shell commands to execute as next steps to diagnose or resolve the issue"`
	Final           bool     `json:"final" jsonschema:"required" jsonschema_description:"Set to true if you are confident the debugging process is complete and no further commands are needed. Set to false if more steps are recommended."`
}

func AnalyzeCommands(prompt string, provider llm.Provider) (CommandAnalysisResponse,error) {
	log.Debugf("Analyzing commands with prompt: %s", prompt)

	// Generate schema
	schema := llm.GenerateSchema[CommandAnalysisResponse]()

	log.Debugf("Generated JSON schema for command analysis: %v", schema)

	res, err := provider.RequestCompletionWithJSONSchema(prompt, schema)
	if err != nil {
		log.Errorf("Error analyzing commands: %v", err)
		return CommandAnalysisResponse{}, fmt.Errorf("error analyzing commands: %w", err)
	}

	var analysisResponse CommandAnalysisResponse
	if err := json.Unmarshal([]byte(res), &analysisResponse); err != nil {
		log.Errorf("Failed to unmarshal response: %v", err)
		return CommandAnalysisResponse{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return analysisResponse, nil
}
