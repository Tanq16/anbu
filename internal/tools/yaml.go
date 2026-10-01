package tools

import (
	"encoding/json/jsontext"
	"fmt"

	"github.com/goccy/go-yaml"
)

func YAMLToJSON(input string) (string, error) {
	out, err := yaml.YAMLToJSON([]byte(input))
	if err != nil {
		return "", err
	}
	v := jsontext.Value(out)
	if err := v.Indent(jsontext.WithIndent("  ")); err != nil {
		return "", err
	}
	return string(v) + "\n", nil
}

func JSONToYAML(input string) (string, error) {
	if !jsontext.Value(input).IsValid() {
		return "", fmt.Errorf("input is not valid JSON")
	}
	out, err := yaml.JSONToYAML([]byte(input))
	if err != nil {
		return "", err
	}
	return string(out), nil
}
