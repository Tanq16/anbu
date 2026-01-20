package anbuGenerics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	u "github.com/tanq16/anbu/utils"
)

type converterInfo struct {
	InputType  string
	OutputType string
	Handler    func(input string)
}

var supportedConverters = map[string]converterInfo{
	"yaml-json": {
		InputType:  "file",
		OutputType: "file",
		Handler:    convertYAMLToJSON,
	},
	"json-yaml": {
		InputType:  "file",
		OutputType: "file",
		Handler:    convertJSONToYAML,
	},
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
	// Text to Text Conversion Handlers (present in convert-more.go)
	"b64": {
		InputType:  "string",
		OutputType: "string",
		Handler:    textToBase64,
	},
	"b64d": {
		InputType:  "string",
		OutputType: "string",
		Handler:    base64ToText,
	},
	"hex": {
		InputType:  "string",
		OutputType: "string",
		Handler:    textToHex,
	},
	"hexd": {
		InputType:  "string",
		OutputType: "string",
		Handler:    hexToText,
	},
	"b64-hex": {
		InputType:  "string",
		OutputType: "string",
		Handler:    base64ToHex,
	},
	"hex-b64": {
		InputType:  "string",
		OutputType: "string",
		Handler:    hexToBase64,
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

// Primary Handler
func ConvertData(converterType string, input string) {
	converter, exists := supportedConverters[converterType]
	if !exists {
		u.PrintError(fmt.Sprintf("unsupported converter type: %s", converterType), nil)
	}
	switch converter.InputType {
	case "file":
		_, err := os.Stat(input)
		if os.IsNotExist(err) {
			u.PrintError(fmt.Sprintf("input file does not exist: %s", input), err)
		}
	case "string":
		if input == "" {
			u.PrintError("input string cannot be empty", nil)
		}
	}
	converter.Handler(input)
}

// Converter functions for YAML <-> JSON

func convertYAMLToJSON(inputFile string) {
	data, err := converterReadInputFile(inputFile)
	if err != nil {
		u.PrintFatal("failed to read input file", err)
	}
	var parsedData any
	if err := yaml.Unmarshal(data, &parsedData); err != nil {
		u.PrintFatal("failed to parse YAML", err)
	}
	jsonData, err := json.MarshalIndent(parsedData, "", "  ")
	if err != nil {
		u.PrintFatal("failed to generate JSON", err)
	}
	outputFile := generateOutputFileName(inputFile, "json")
	if err := converterWriteOutputFile(outputFile, jsonData); err != nil {
		u.PrintFatal("failed to write output file", err)
	}
	u.PrintSuccess(fmt.Sprintf("Converted YAML to JSON: %s", outputFile))
}

func convertJSONToYAML(inputFile string) {
	data, err := converterReadInputFile(inputFile)
	if err != nil {
		u.PrintFatal("failed to read input file", err)
	}
	var parsedData any
	if err := json.Unmarshal(data, &parsedData); err != nil {
		u.PrintFatal("failed to parse JSON", err)
	}
	yamlData, err := yaml.Marshal(parsedData)
	if err != nil {
		u.PrintFatal("failed to generate YAML", err)
	}
	outputFile := generateOutputFileName(inputFile, "yaml")
	if err := converterWriteOutputFile(outputFile, yamlData); err != nil {
		u.PrintFatal("failed to write output file", err)
	}
	u.PrintSuccess(fmt.Sprintf("Converted JSON to YAML: %s", outputFile))
}

// Helper functions

func converterReadInputFile(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}

func converterWriteOutputFile(filePath string, data []byte) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

func generateOutputFileName(inputPath string, toExt string) string {
	baseName := strings.TrimSuffix(inputPath, filepath.Ext(inputPath))
	return fmt.Sprintf("%s.%s", baseName, toExt)
}
