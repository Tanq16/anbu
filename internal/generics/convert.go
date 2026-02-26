package anbuGenerics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	u "github.com/tanq16/anbu/internal/utils"
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

func convertYAMLToJSON(inputFile string) error {
	data, err := converterReadInputFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}
	var parsedData any
	if err := yaml.Unmarshal(data, &parsedData); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}
	jsonData, err := json.MarshalIndent(parsedData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to generate JSON: %w", err)
	}
	outputFile := generateOutputFileName(inputFile, "json")
	if err := converterWriteOutputFile(outputFile, jsonData); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}
	u.PrintSuccess(fmt.Sprintf("Converted YAML to JSON: %s", outputFile))
	return nil
}

func convertJSONToYAML(inputFile string) error {
	data, err := converterReadInputFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}
	var parsedData any
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}
	yamlData, err := yaml.Marshal(parsedData)
	if err != nil {
		return fmt.Errorf("failed to generate YAML: %w", err)
	}
	outputFile := generateOutputFileName(inputFile, "yaml")
	if err := converterWriteOutputFile(outputFile, yamlData); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}
	u.PrintSuccess(fmt.Sprintf("Converted JSON to YAML: %s", outputFile))
	return nil
}

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
