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

	u "github.com/tanq16/anbu/utils"
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

func RunTemplate(filePath string, overrideVars map[string]string) {
	logger := u.NewManager(0)
	logger.StartDisplay()
	templateFuncID := logger.Register("Run Template")
	logger.SetMessage(templateFuncID, fmt.Sprintf("Initializing template: %s", filePath))
	defer logger.StopDisplay()

	// Load template
	data, err := os.ReadFile(filePath)
	if err != nil {
		logger.ReportError(templateFuncID, fmt.Errorf("failed to read template file: %s", err))
		logger.SetMessage(templateFuncID, "Error reading template file")
		return
	}
	var config TemplateConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		logger.ReportError(templateFuncID, fmt.Errorf("failed to parse template YAML: %s", err))
		logger.SetMessage(templateFuncID, "Error parsing template YAML")
		return
	}
	if config.Variables == nil {
		config.Variables = make(map[string]string)
	}
	maps.Copy(config.Variables, overrideVars)
	logger.SetMessage(templateFuncID, fmt.Sprintf("Running template: %s", config.Name))
	// if config.Description != "" {
	// 	fmt.Println(u.OutDebug(config.Description))
	// }

	// Run template steps
	for i, step := range config.Steps {
		stepID := logger.Register(fmt.Sprintf("Step %d: %s", i+1, step.Name))
		logger.SetMessage(stepID, fmt.Sprintf("Executing step %d: %s", i+1, step.Name))
		cmdWithVars, err := processTemplateVariables(step.Command, config.Variables)
		if err != nil {
			logger.ReportError(stepID, fmt.Errorf("failed to process step %d variables: %s", i+1, err))
			return
		}

		// Execute command with streaming output
		cmd := exec.Command("sh", "-c", cmdWithVars)
		stdoutPipe, _ := cmd.StdoutPipe()
		stderrPipe, _ := cmd.StderrPipe()
		if err := cmd.Start(); err != nil {
			logger.ReportError(stepID, fmt.Errorf("failed to start command: %s", err))
			logger.SetMessage(templateFuncID, "Template failed midway")
			return
		}
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				logger.AddStreamLine(stepID, scanner.Text())
			}
		}()
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stderrPipe)
			for scanner.Scan() {
				logger.AddStreamLine(stepID, scanner.Text())
			}
		}()
		wg.Wait()
		err = cmd.Wait()

		if err != nil {
			logger.ReportError(stepID, fmt.Errorf("command failed: %s", err))
			if step.IgnoreError {
				logger.SetMessage(stepID, fmt.Sprintf("Step %d errored but ignored", i+1))
			} else {
				logger.SetMessage(stepID, fmt.Sprintf("Step %d failed", i+1))
				logger.SetMessage(templateFuncID, "Template failed midway")
				return
			}
		} else {
			logger.Complete(stepID, fmt.Sprintf("Step %d completed successfully", i+1))
		}
		logger.AddProgressBarToStream(templateFuncID, int64(i+1), int64(len(config.Steps)), fmt.Sprintf("%d / %d", i+1, len(config.Steps)))
	}
	logger.Complete(templateFuncID, fmt.Sprintf("Template %s completed successfully", config.Name))
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
