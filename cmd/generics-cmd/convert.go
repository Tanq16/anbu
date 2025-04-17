package genericsCmd

import (
	"fmt"

	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
	"github.com/tanq16/anbu/utils"
)

var ConvertCmd = &cobra.Command{
	Use:   "convert [converter] [data or file]",
	Short: "Convert data between different formats",
	Long: `Convert data between different formats.

Converters:

- yaml-json: Convert YAML to JSON, requires a file path as input
- json-yaml: Convert JSON to YAML, requires a file path as input
- docker-compose: Convert Docker Compose to YAML, requires a string as input
- compose-docker: Convert Compose to Docker, requires a file path as input
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		logger := utils.GetLogger("convert")
		converterType := args[0]
		input := args[1]
		err := anbuGenerics.ConvertData(converterType, input)
		if err != nil {
			logger.Fatal().Err(err).Msg("Conversion failed")
		}
		fmt.Println(utils.OutDetail("Data converted successfully"))
	},
}
