package anbuGenerics

import (
	"fmt"
	"os"
)

type converterInfo struct {
	InputType  string
	OutputType string
	Handler    func(input string) error
}

var supportedConverters = map[string]converterInfo{
	"docker-compose": {
		InputType:  "string",
		OutputType: "file",
		Handler:    convertDockerToCompose,
	},
	"compose-docker": {
		InputType:  "file",
		OutputType: "string",
		Handler:    convertComposeToDocker,
	},
	"urld": {
		InputType:  "string",
		OutputType: "string",
		Handler:    urlToText,
	},
	"url": {
		InputType:  "string",
		OutputType: "string",
		Handler:    textToUrl,
	},
	"jwtd": {
		InputType:  "string",
		OutputType: "string",
		Handler:    jwtDecode,
	},
}

func ConvertData(converterType string, input string) error {
	converter, exists := supportedConverters[converterType]
	if !exists {
		return fmt.Errorf("unsupported converter type: %s", converterType)
	}
	switch converter.InputType {
	case "file":
		_, err := os.Stat(input)
		if os.IsNotExist(err) {
			return fmt.Errorf("input file does not exist: %s", input)
		}
	case "string":
		if input == "" {
			return fmt.Errorf("input string cannot be empty")
		}
	}
	return converter.Handler(input)
}
