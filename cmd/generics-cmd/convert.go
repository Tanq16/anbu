package genericsCmd

import (
	"github.com/spf13/cobra"

	anbuGenerics "github.com/tanq16/anbu/internal/generics"
)

var ConvertCmd = &cobra.Command{
	Use:     "convert [converter] [data or file]",
	Aliases: []string{"c"},
	Short:   "Convert data between different formats",
	Long: `Convert data between different formats. These are the supported converters:

File Formats:
- yaml-json:      Convert YAML to JSON, requires a file path as input
- json-yaml:      Convert JSON to YAML, requires a file path as input
- docker-compose: Convert Docker Compose to YAML, requires a string as input
- compose-docker: Convert Compose to Docker, requires a file path as input

Encoding Formats:
- b64:     Convert plain text to base64 encoded string
- b64d:    Convert base64 encoded string to plain text
- hex:     Convert plain text to hex encoded string
- hexd:    Convert hex encoded string to plain text
- b64-hex: Convert base64 encoded string to hex encoded string
- hex-b64: Convert hex encoded string to base64 encoded string
- url:     Convert plain text to URL encoded string
- urld:    Convert URL encoded string to plain text
- jwtd:    Convert JWT to decoded struct
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		converterType := args[0]
		input := args[1]
		anbuGenerics.ConvertData(converterType, input)
	},
}
