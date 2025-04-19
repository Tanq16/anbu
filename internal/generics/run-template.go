package anbuGenerics

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sync"
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
	fmt.Println(utils.OutDetail(fmt.Sprintf("\nRunning template: %s", config.Name)))
	if config.Description != "" {
		fmt.Println(utils.OutDebug(config.Description))
	}

	// Run template steps
	for i, step := range config.Steps {
		cmdWithVars, err := processTemplateVariables(step.Command, config.Variables)
		if err != nil {
			return fmt.Errorf("failed to process variables in step %d: %w", i+1, err)
		}
		fmt.Printf("\n%s %s\n", utils.OutCyan(fmt.Sprintf("[Step %d]", i+1)), utils.OutCyan(step.Name))
		fmt.Printf("%s %s\n", utils.OutSuccess("Command:"), utils.OutSuccess(cmdWithVars))

		// Execute command with streaming output
		cmd := exec.Command("sh", "-c", cmdWithVars)
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("failed to create stdout pipe: %w", err)
		}
		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			return fmt.Errorf("failed to create stderr pipe: %w", err)
		}
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start command: %w", err)
		}
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				fmt.Println(utils.OutDebug(scanner.Text()))
			}
		}()
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stderrPipe)
			for scanner.Scan() {
				fmt.Println(utils.OutDebug(scanner.Text()))
			}
		}()
		wg.Wait()
		err = cmd.Wait()

		if err != nil {
			logger.Debug().Err(err).Str("command", cmdWithVars).Msg("Step failed")
			if step.IgnoreError {
				fmt.Println(utils.OutWarning("Command failed, but ignoring error and continuing..."))
			} else {
				return fmt.Errorf("step %d failed: %w", i+1, err)
			}
		}
	}
	fmt.Println(utils.OutInfo("\nTemplate execution completed successfully"))
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
