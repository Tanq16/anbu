package anbuGenerics

import (
	"fmt"
	"os"
	"strings"

	u "github.com/tanq16/anbu/utils"
	"gopkg.in/yaml.v3"
)

// Convert Docker run command to Docker Compose format
func convertDockerToCompose(input string) {
	// Trim and clean the input
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "docker run") {
		u.PrintError("Invalid docker run command. Must start with 'docker run'")
		return
	}

	// Parse the docker run command
	composeConfig := map[string]interface{}{
		"version": "3",
		"services": map[string]interface{}{
			"app": map[string]interface{}{},
		},
	}
	service := composeConfig["services"].(map[string]interface{})["app"].(map[string]interface{})

	// Split the command by spaces but respect quotes
	parts := splitCommand(input[len("docker run"):])

	// Default command mode
	detached := false
	var ports []string
	var volumes []string
	var environment []string
	var command []string

	// Process docker run options
	imageName := ""
	skipNext := false

	for i := 0; i < len(parts); i++ {
		if skipNext {
			skipNext = false
			continue
		}

		part := parts[i]
		if part == "" {
			continue
		}

		// Handle options with values
		if strings.HasPrefix(part, "-") {
			switch {
			case part == "-d" || part == "--detach":
				detached = true

			case part == "-p" || part == "--publish":
				if i+1 < len(parts) {
					ports = append(ports, parts[i+1])
					skipNext = true
				}

			case strings.HasPrefix(part, "-p=") || strings.HasPrefix(part, "--publish="):
				value := strings.Split(part, "=")[1]
				ports = append(ports, value)

			case part == "-v" || part == "--volume":
				if i+1 < len(parts) {
					volumes = append(volumes, parts[i+1])
					skipNext = true
				}

			case strings.HasPrefix(part, "-v=") || strings.HasPrefix(part, "--volume="):
				value := strings.Split(part, "=")[1]
				volumes = append(volumes, value)

			case part == "-e" || part == "--env":
				if i+1 < len(parts) {
					environment = append(environment, parts[i+1])
					skipNext = true
				}

			case strings.HasPrefix(part, "-e=") || strings.HasPrefix(part, "--env="):
				value := strings.Split(part, "=")[1]
				environment = append(environment, value)

			case part == "--name":
				if i+1 < len(parts) {
					// Use container name as service name
					serviceName := parts[i+1]
					composeConfig["services"].(map[string]interface{})["app"] = nil
					composeConfig["services"].(map[string]interface{})[serviceName] = service
					skipNext = true
				}

			default:
				// Skip unknown options for simplicity
				if i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "-") {
					skipNext = true
				}
			}
		} else if imageName == "" {
			// First non-flag argument is the image name
			imageName = part
		} else {
			// Remaining arguments form the command
			command = append(command, part)
		}
	}

	// Set the values in the compose structure
	if imageName != "" {
		service["image"] = imageName
	}

	if detached {
		// No need to set this in compose as it's the default
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

	// Convert to YAML
	yamlData, err := yaml.Marshal(composeConfig)
	if err != nil {
		u.PrintError(fmt.Sprintf("Failed to generate YAML: %s", err))
		return
	}

	// Output the docker-compose.yml file
	outputFile := "docker-compose.yml"
	if err := os.WriteFile(outputFile, yamlData, 0644); err != nil {
		u.PrintError(fmt.Sprintf("Failed to write output file: %s", err))
		return
	}

	u.PrintSuccess(fmt.Sprintf("Docker run command converted to Docker Compose: %s", outputFile))
}

// Convert Docker Compose to Docker run commands
func convertComposeToDocker(inputFile string) {
	data, err := os.ReadFile(inputFile)
	if err != nil {
		u.PrintError(fmt.Sprintf("Failed to read input file: %s", err))
		return
	}

	var composeConfig map[string]interface{}
	if err := yaml.Unmarshal(data, &composeConfig); err != nil {
		u.PrintError(fmt.Sprintf("Failed to parse YAML: %s", err))
		return
	}

	services, ok := composeConfig["services"].(map[string]interface{})
	if !ok {
		u.PrintError("No services found in the Docker Compose file")
		return
	}

	var dockerCommands []string

	for serviceName, serviceConfig := range services {
		service, ok := serviceConfig.(map[string]interface{})
		if !ok {
			continue
		}

		var command strings.Builder
		command.WriteString("docker run -d")

		// Add container name
		command.WriteString(fmt.Sprintf(" --name %s", serviceName))

		// Add ports
		if ports, ok := service["ports"].([]interface{}); ok {
			for _, port := range ports {
				command.WriteString(fmt.Sprintf(" -p %v", port))
			}
		}

		// Add volumes
		if volumes, ok := service["volumes"].([]interface{}); ok {
			for _, volume := range volumes {
				command.WriteString(fmt.Sprintf(" -v %v", volume))
			}
		}

		// Add environment variables
		if env, ok := service["environment"].([]interface{}); ok {
			for _, e := range env {
				command.WriteString(fmt.Sprintf(" -e %v", e))
			}
		} else if envMap, ok := service["environment"].(map[string]interface{}); ok {
			for key, value := range envMap {
				command.WriteString(fmt.Sprintf(" -e %s=%v", key, value))
			}
		}

		// Add image
		if image, ok := service["image"].(string); ok {
			command.WriteString(fmt.Sprintf(" %s", image))
		}

		// Add command if exists
		if cmd, ok := service["command"].(string); ok {
			command.WriteString(fmt.Sprintf(" %s", cmd))
		}

		dockerCommands = append(dockerCommands, command.String())
	}

	// Output results
	fmt.Println("\nDocker run commands for services in Docker Compose file:")
	for _, cmd := range dockerCommands {
		fmt.Println("\n" + u.FSuccess(cmd))
	}
}

// Helper function to split a command into parts respecting quotes
func splitCommand(command string) []string {
	var parts []string
	var inQuote bool
	var quoteChar byte
	var current strings.Builder

	for i := 0; i < len(command); i++ {
		c := command[i]

		if (c == '"' || c == '\'') && (i == 0 || command[i-1] != '\\') {
			if inQuote && quoteChar == c {
				// End quote
				inQuote = false
				parts = append(parts, current.String())
				current.Reset()
			} else if !inQuote {
				// Start quote
				inQuote = true
				quoteChar = c
				// If we have content, add it
				if current.Len() > 0 {
					parts = append(parts, current.String())
					current.Reset()
				}
			} else {
				// Quote char inside another quote, treat as normal char
				current.WriteByte(c)
			}
		} else if c == ' ' && !inQuote {
			// Space outside of quotes separates parts
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			// Normal character
			current.WriteByte(c)
		}
	}

	// Add the last part if it exists
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}
