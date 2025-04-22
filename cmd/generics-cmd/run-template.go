package genericsCmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
)

var templateVars []string

var TemplateCmd = &cobra.Command{
	Use:   "exec [template-file]",
	Short: "Run a template file, executing defined commands in sequence",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		templateFile := args[0]
		if _, err := os.Stat(templateFile); os.IsNotExist(err) {
			logger.Fatal().Err(err).Msg("Template file not found")
		}
		// Resolve variable overrides
		overrideVars := make(map[string]string)
		for _, varStr := range templateVars {
			parts := strings.SplitN(varStr, "=", 2)
			if len(parts) != 2 {
				logger.Fatal().Str("variable", varStr).Msg("Invalid variable format, expected 'key=value'")
			}
			overrideVars[parts[0]] = parts[1]
		}
		// Execute template
		if err := anbuGenerics.RunTemplate(templateFile, overrideVars); err != nil {
			logger.Fatal().Err(err).Msg("Template execution failed")
		}
	},
}

func init() {
	TemplateCmd.Flags().StringSliceVarP(&templateVars, "var", "v", []string{}, "Variables to override in format 'key=value'")
}
