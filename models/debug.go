package models

import (
	"github.com/remijnoel/ailops/internal"
	"time"
)

type Action struct {
	Name       string `json:"name"`        // e.g., could be a command or a description of the action taken
	ActionType string `json:"action_type"` // e.g., "command" (only type for now)
	Result     string `json:"result"`      // for a command this would be the output, when pulling data this would be the data pulled, etc.
	Status     string `json:"status"`      // e.g., "success", "failure", "in-progress"
	Timestamp  string `json:"timestamp"`   // Time when the action was taken
	Remote     string `json:"remote"`      // Remote host if applicable, e.g., "remote_host_1"
}

func (a *Action) IsCommand() bool {
	// Check if the action is a command by its type
	return a.ActionType == "command"
}

func (a *Action) IsRemote() bool {
	// Check if the action is remote by checking if Remote is set
	return a.Remote != ""
}

type Batch struct {
	Description string    `json:"description"` // The reason for the batch, e.g., "Debugging issue with X"
	Actions     []*Action `json:"actions"`     // List of actions in this batch
	Analysis    string    `json:"analysis"`    // Analysis of the batch actions
	NextSteps   []string  `json:"next_steps"`  // Suggested next steps after this batch
	Completed   bool      `json:"completed"`   // Indicates if the batch has been completed
}

func (b *Batch) AddAction(name string, actionType string) *Action {
	action := &Action{
		Name:       name,
		ActionType: actionType,
		Status:     "new", // Default status is "new"
		Result:     "",    // Initially empty, will be filled after execution
	}
	b.Actions = append(b.Actions, action)

	return action
}

type DebugSessionLog struct {
	ID               string   `json:"id"`
	IssueDescription string   `json:"issue_description"`
	Batches          []*Batch `json:"batches"`
	StartTime        string   `json:"start_time"`
	EndTime          string   `json:"end_time"`
	Summary          string   `json:"summary"`
	Diagnosed        bool     `json:"ended"`
}

func (d *DebugSessionLog) SetIssueDescription(description string) {
	d.IssueDescription = description
}

func (d *DebugSessionLog) StartSession() {
	d.StartTime = time.Now().Format(time.RFC3339)
}

func (d *DebugSessionLog) EndSession() {
	d.EndTime = time.Now().Format(time.RFC3339)
}

func (d *DebugSessionLog) AddBatch(batch *Batch) {
	if d.Batches == nil {
		d.Batches = []*Batch{}
	}
	d.Batches = append(d.Batches, batch)
}

func (d *DebugSessionLog) LastBatch() *Batch {
	if len(d.Batches) == 0 {
		return nil
	}
	return d.Batches[len(d.Batches)-1]
}

func NewDebugSessionLog(issueDescription string) *DebugSessionLog {
	session := &DebugSessionLog{
		ID:               internal.GenerateUniqueID(), // Assume this function generates a unique ID
		IssueDescription: issueDescription,
		Batches:          []*Batch{},
		StartTime:        "",
		EndTime:          "",
		Summary:          "",
		Diagnosed:        false,
	}

	return session
}
