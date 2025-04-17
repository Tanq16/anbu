package anbuGenerics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tanq16/anbu/utils"
	"gopkg.in/yaml.v3"
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
		// Handler: convertDockerToCompose,
	},
	"compose-docker": {
		InputType:  "file",
		OutputType: "string",
		// Handler: convertComposeToDocker,
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

// Converter functions
func convertYAMLToJSON(inputFile string) error {
	logger := utils.GetLogger("converter")
	logger.Debug().Str("input", inputFile).Msg("Converting YAML to JSON")
	data, err := converterReadInputFile(inputFile)
	if err != nil {
		return err
	}
	// Parse YAML
	var parsedData any
	if err := yaml.Unmarshal(data, &parsedData); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}
	// Convert to JSON
	jsonData, err := json.MarshalIndent(parsedData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to generate JSON: %w", err)
	}

	outputFile := generateOutputFileName(inputFile, "json")
	if err := converterWriteOutputFile(outputFile, jsonData); err != nil {
		return err
	}
	fmt.Println(utils.OutSuccess(fmt.Sprintf("Converted YAML to JSON: %s", outputFile)))
	return nil
}

func convertJSONToYAML(inputFile string) error {
	logger := utils.GetLogger("converter")
	logger.Debug().Str("input", inputFile).Msg("Converting JSON to YAML")
	data, err := converterReadInputFile(inputFile)
	if err != nil {
		return err
	}
	// Parse JSON
	var parsedData any
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}
	// Convert to YAML
	yamlData, err := yaml.Marshal(parsedData)
	if err != nil {
		return fmt.Errorf("failed to generate YAML: %w", err)
	}

	outputFile := generateOutputFileName(inputFile, "yaml")
	if err := converterWriteOutputFile(outputFile, yamlData); err != nil {
		return err
	}
	fmt.Println(utils.OutSuccess(fmt.Sprintf("Converted JSON to YAML: %s", outputFile)))
	return nil
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
