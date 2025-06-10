package report

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"text/template"

	"github.com/remijnoel/ailops/models"
)

//go:embed template.md
var templateMD string

type Format string

const (
	Markdown Format = "markdown"
	Json     Format = "json"
)

type ReportConfig struct {
	Format                 Format
	IncludeCommandOutput   bool
	IncludeAnalysisHistory bool
}

func GenerateReport(session *models.DebugSessionLog, config ReportConfig) string {
	if config.Format == Markdown {
		return generateMarkdownReport(session, config)
	} else if config.Format == Json {
		return generateJsonReport(session)
	}
	return ""
}

func generateJsonReport(session *models.DebugSessionLog) string {
	// Convert the session to JSON format
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return "Error generating JSON report: " + err.Error()
	}
	return string(data)
}

func generateMarkdownReport(session *models.DebugSessionLog, config ReportConfig) string {
	tmpl, err := template.New("report").Parse(templateMD)
	if err != nil {
		return "Error parsing markdown template: " + err.Error()
	}

	var buf bytes.Buffer
	data := struct {
		*models.DebugSessionLog
		Config ReportConfig
	}{
		DebugSessionLog: session,
		Config:          config,
	}

	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "Error executing markdown template: " + err.Error()
	}

	return buf.String()
}
