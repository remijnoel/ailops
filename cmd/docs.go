package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"os"
	log "github.com/sirupsen/logrus"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Docs related commands",
}

var genDocCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate documentation for the AILOps CLI",
	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")
		output, _ := cmd.Flags().GetString("output")
		if format != "markdown" {
			log.Fatalf("Unsupported format: %s. Only 'markdown' is supported for now.", format)
		}

		// Check if output directory exists
		if _, err := os.Stat(output); os.IsNotExist(err) {
			err := os.MkdirAll(output, 0755)
			if err != nil {
				log.Fatalf("Failed to create output directory: %s", err)
			}
		}

		log.Infof("Generating documentation in %s format to %s", format, output)
		err := doc.GenMarkdownTree(RootCmd, output)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("Documentation generated successfully in %s", output)
	},
}


func init() {
	docsCmd.AddCommand(genDocCmd)
	genDocCmd.Flags().StringP("output", "o", "docs", "Output directory for generated documentation")
	genDocCmd.Flags().StringP("format", "f", "markdown", "Format of the generated documentation (markdown, html, etc.)")

	RootCmd.AddCommand(docsCmd)
}