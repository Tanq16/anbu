package anbuGenerics

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"text/template"

	"maps"

	"github.com/tanq16/anbu/utils"
	"gopkg.in/yaml.v3"
)

type TemplateStep struct {
	Name        string `yaml:"name"`
	Command     string `yaml:"command"`
	IgnoreError bool   `yaml:"ignore_errors,omitempty"`
}

type TemplateConfig struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Variables   map[string]string `yaml:"variables,omitempty"`
	Steps       []TemplateStep    `yaml:"steps"`
}

func RunTemplate(filePath string, overrideVars map[string]string) error {
	logger := utils.GetLogger("template")

	// Load template
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}
	var config TemplateConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse template YAML: %w", err)
	}
	if config.Variables == nil {
		config.Variables = make(map[string]string)
	}
	maps.Copy(config.Variables, overrideVars)
	fmt.Println(utils.OutDetail(fmt.Sprintf("Running template: %s", config.Name)))
	if config.Description != "" {
		fmt.Println(utils.OutInfo(config.Description))
	}

	// Run template steps
	for i, step := range config.Steps {
		cmdWithVars, err := processTemplateVariables(step.Command, config.Variables)
		if err != nil {
			return fmt.Errorf("failed to process variables in step %d: %w", i+1, err)
		}
		fmt.Printf("\n%s %s\n", utils.OutSuccess(fmt.Sprintf("[Step %d]", i+1)), utils.OutDetail(step.Name))
		fmt.Printf("%s %s\n", utils.OutDebug("Command:"), cmdWithVars)

		// Execute command
		cmd := exec.Command("sh", "-c", cmdWithVars)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			logger.Error().Err(err).Str("command", cmdWithVars).Msg("Step failed")
			if step.IgnoreError {
				fmt.Println(utils.OutWarning("Command failed, but ignoring error and continuing..."))
			} else {
				return fmt.Errorf("step %d failed: %w", i+1, err)
			}
		}
	}
	fmt.Println(utils.OutSuccess("\nTemplate execution completed successfully"))
	return nil
}

func processTemplateVariables(cmdStr string, variables map[string]string) (string, error) {
	tmpl, err := template.New("command").Delims("{{", "}}").Parse(cmdStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse command template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", fmt.Errorf("failed to execute command template: %w", err)
	}
	return buf.String(), nil
}
