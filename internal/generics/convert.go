package anbuGenerics

import (
	"fmt"
	"os"

	"github.com/tanq16/anbu/utils"
)

type converterInfo struct {
	InputType  string
	OutputType string
	Handler    func(input string) error
}

var supportedConverters = map[string]converterInfo{
	"yaml-json": {
		InputType:  "file",
		OutputType: "file",
		Handler: func(input string) error {
			logger := utils.GetLogger("converter")
			logger.Debug().Str("input", input).Msg("Converting YAML to JSON")
			return nil
		},
	},
	"json-yaml": {
		InputType:  "file",
		OutputType: "file",
		Handler: func(input string) error {
			logger := utils.GetLogger("converter")
			logger.Debug().Str("input", input).Msg("Converting JSON to YAML")
			return nil
		},
	},
	"docker-compose": {
		InputType:  "string",
		OutputType: "file",
		Handler: func(input string) error {
			logger := utils.GetLogger("converter")
			logger.Debug().Str("input", input).Msg("Converting Docker Compose to YAML")
			return nil
		},
	},
	"compose-docker": {
		InputType:  "file",
		OutputType: "string",
		Handler: func(input string) error {
			logger := utils.GetLogger("converter")
			logger.Debug().Str("input", input).Msg("Converting Compose to Docker")
			return nil
		},
	},
}

// ConvertData validates the converter and calls the appropriate handler
func ConvertData(converterType string, input string) error {
	converter, exists := supportedConverters[converterType]
	if !exists {
		return fmt.Errorf("unsupported converter type: %s", converterType)
	}
	if converter.InputType == "file" {
		_, err := os.Stat(input)
		if os.IsNotExist(err) {
			return fmt.Errorf("input file does not exist: %s", input)
		}
	} else if converter.InputType == "string" {
		if input == "" {
			return fmt.Errorf("input string cannot be empty")
		}
	}
	return converter.Handler(input)
}
