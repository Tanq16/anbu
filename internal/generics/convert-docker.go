package anbuGenerics

import (
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	u "github.com/tanq16/anbu/internal/utils"
)

func convertDockerToCompose(input string) error {
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "docker run") {
		return fmt.Errorf("invalid docker run command: must start with 'docker run'")
	}
	composeConfig := map[string]any{
		"services": map[string]any{
			"app": map[string]any{},
		},
	}
	service := composeConfig["services"].(map[string]any)["app"].(map[string]any)
	parts := splitCommand(input[len("docker run"):])

	var ports []string
	var volumes []string
	var environment []string
	var command []string
	imageName := ""
	skipNext := false

	for i := range parts {
		if skipNext {
			skipNext = false
			continue
		}
		part := parts[i]
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, "-") {
			switch {
			case part == "-d" || part == "--detach":
			case part == "-p" || part == "--publish":
				if i+1 < len(parts) {
					ports = append(ports, parts[i+1])
					skipNext = true
				}
			case strings.HasPrefix(part, "-p=") || strings.HasPrefix(part, "--publish="):
				pParts := strings.SplitN(part, "=", 2)
				if len(pParts) == 2 {
					ports = append(ports, pParts[1])
				}
			case part == "-v" || part == "--volume":
				if i+1 < len(parts) {
					volumes = append(volumes, parts[i+1])
					skipNext = true
				}
			case strings.HasPrefix(part, "-v=") || strings.HasPrefix(part, "--volume="):
				vParts := strings.SplitN(part, "=", 2)
				if len(vParts) == 2 {
					volumes = append(volumes, vParts[1])
				}
			case part == "-e" || part == "--env":
				if i+1 < len(parts) {
					environment = append(environment, parts[i+1])
					skipNext = true
				}
			case strings.HasPrefix(part, "-e=") || strings.HasPrefix(part, "--env="):
				eParts := strings.SplitN(part, "=", 2)
				if len(eParts) == 2 {
					environment = append(environment, eParts[1])
				}
			case part == "--name":
				if i+1 < len(parts) {
					serviceName := parts[i+1]
					composeConfig["services"].(map[string]any)["app"] = nil
					composeConfig["services"].(map[string]any)[serviceName] = service
					skipNext = true
				}
			default:
				if i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "-") {
					skipNext = true
				}
			}
		} else if imageName == "" {
			imageName = part
		} else {
			command = append(command, part)
		}
	}
	if imageName != "" {
		service["image"] = imageName
	}
	if len(ports) > 0 {
		service["ports"] = ports
	}
	if len(volumes) > 0 {
		service["volumes"] = volumes
	}
	if len(environment) > 0 {
		service["environment"] = environment
	}
	if len(command) > 0 {
		service["command"] = strings.Join(command, " ")
	}

	yamlData, err := yaml.Marshal(composeConfig)
	if err != nil {
		return fmt.Errorf("failed to generate YAML: %w", err)
	}
	outputFile := "docker-compose.yml"
	if err := os.WriteFile(outputFile, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}
	u.PrintSuccess(fmt.Sprintf("Docker run command converted to Docker Compose: %s", outputFile))
	return nil
}

func convertComposeToDocker(inputFile string) error {
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}
	var composeConfig map[string]any
	if err := yaml.Unmarshal(data, &composeConfig); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}
	services, ok := composeConfig["services"].(map[string]any)
	if !ok {
		return fmt.Errorf("no services found in the Docker Compose file")
	}

	var dockerCommands []string
	for serviceName, serviceConfig := range services {
		service, ok := serviceConfig.(map[string]any)
		if !ok {
			continue
		}
		var command strings.Builder
		command.WriteString("docker run -d")
		command.WriteString(fmt.Sprintf(" --name %s", serviceName))
		if ports, ok := service["ports"].([]any); ok {
			for _, port := range ports {
				command.WriteString(fmt.Sprintf(" -p %v", port))
			}
		}
		if volumes, ok := service["volumes"].([]any); ok {
			for _, volume := range volumes {
				command.WriteString(fmt.Sprintf(" -v %v", volume))
			}
		}
		if env, ok := service["environment"].([]any); ok {
			for _, e := range env {
				command.WriteString(fmt.Sprintf(" -e %v", e))
			}
		} else if envMap, ok := service["environment"].(map[string]any); ok {
			for key, value := range envMap {
				command.WriteString(fmt.Sprintf(" -e %s=%v", key, value))
			}
		}
		if image, ok := service["image"].(string); ok {
			command.WriteString(fmt.Sprintf(" %s", image))
		}
		if cmd, ok := service["command"].(string); ok {
			command.WriteString(fmt.Sprintf(" %s", cmd))
		}
		dockerCommands = append(dockerCommands, command.String())
	}
	u.PrintGeneric("\nDocker run commands for services in Docker Compose file:")
	for _, cmd := range dockerCommands {
		u.PrintGeneric("\n" + u.FSuccess(cmd))
	}
	return nil
}

func splitCommand(command string) []string {
	var parts []string
	var inQuote bool
	var quoteChar byte
	var current strings.Builder
	for i := 0; i < len(command); i++ {
		c := command[i]
		if (c == '"' || c == '\'') && (i == 0 || command[i-1] != '\\') {
			if inQuote && quoteChar == c {
				inQuote = false
				parts = append(parts, current.String())
				current.Reset()
			} else if !inQuote {
				inQuote = true
				quoteChar = c
				if current.Len() > 0 {
					parts = append(parts, current.String())
					current.Reset()
				}
			} else {
				current.WriteByte(c)
			}
		} else if c == ' ' && !inQuote {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(c)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}
