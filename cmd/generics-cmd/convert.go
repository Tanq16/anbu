package genericsCmd

import (
	"github.com/spf13/cobra"

	anbuGenerics "github.com/tanq16/anbu/internal/generics"
	u "github.com/tanq16/anbu/internal/utils"
)

var ConvertCmd = &cobra.Command{
	Use:     "convert [converter] [data or file]",
	Aliases: []string{"c"},
	Short:   "Convert data between different formats and encodings",
	Long: `Convert data between different formats and encodings.

Examples:
  anbu convert yaml-json config.yaml          # Convert YAML file to JSON
  anbu convert json-yaml data.json            # Convert JSON file to YAML
  anbu convert b64 "Hello World"              # Convert text to base64
  anbu convert b64d "SGVsbG8gV29ybGQ="        # Decode base64 to text
  anbu convert hex "Hello World"              # Convert text to hex
  anbu convert hexd "48656c6c6f20576f726c64"  # Decode hex to text
  anbu convert url "Hello World"              # URL encode text
  anbu convert urld "Hello%20World"           # URL decode text
  anbu convert jwtd "$TOKEN"                  # Decode JWT token`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		converterType := args[0]
		input := args[1]
		if err := anbuGenerics.ConvertData(converterType, input); err != nil {
			u.PrintFatal("conversion failed", err)
		}
	},
}
