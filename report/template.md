# Debug Session Report

## Issue Description

{{.IssueDescription}}

## Batches

{{range .Batches}}
### {{.Description}}

{{range .Actions}}
**Command:** `{{.Name}}`
{{if $.Config.IncludeCommandOutput}}

```shell
{{.Result}}
```

{{end}}
{{end}}
{{if $.Config.IncludeAnalysisHistory}}
**Analysis:**

{{.Analysis}}
{{end}}
{{end}}

## Summary

{{if .Summary}}{{.Summary}}{{else}}No summary available.{{end}}
