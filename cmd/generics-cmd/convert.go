package genericsCmd

import (
	"github.com/spf13/cobra"

	anbuGenerics "github.com/tanq16/anbu/internal/generics"
	u "github.com/tanq16/anbu/utils"
)

var ConvertCmd = &cobra.Command{
	Use:     "convert [converter] [data or file]",
	Aliases: []string{"c"},
	Short:   "Convert data between different formats and encodings",
	Long: `Convert data between different formats and encodings.

Examples:
  anbu convert docker-compose "docker run ..."  # Convert docker run to compose
  anbu convert compose-docker compose.yaml      # Convert compose to docker run
  anbu convert url "Hello World"                # URL encode text
  anbu convert urld "Hello%20World"             # URL decode text
  anbu convert jwtd "$TOKEN"                    # Decode JWT token`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		converterType := args[0]
		input := args[1]
		if err := anbuGenerics.ConvertData(converterType, input); err != nil {
			u.PrintFatal("conversion failed", err)
		}
	},
}
