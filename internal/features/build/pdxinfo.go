package build

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func loadPDXInfo(path string) ([]byte, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read application pdxinfo: %w", err)
	}
	if err := validatePDXInfo(contents); err != nil {
		return nil, fmt.Errorf("validate application pdxinfo: %w", err)
	}
	return contents, nil
}

func validatePDXInfo(contents []byte) error {
	values := make(map[string]string)
	for lineNumber, line := range strings.Split(string(contents), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if !found || key == "" || value == "" {
			return fmt.Errorf("line %d must be key=value", lineNumber+1)
		}
		if _, duplicate := values[key]; duplicate {
			return fmt.Errorf("field %s is duplicated", key)
		}
		values[key] = value
	}
	for _, field := range []string{"name", "author", "bundleID", "version", "buildNumber"} {
		if values[field] == "" {
			return fmt.Errorf("field %s is required", field)
		}
	}
	if bundleID := values["bundleID"]; !strings.Contains(bundleID, ".") || strings.ContainsAny(bundleID, " \t") {
		return fmt.Errorf("field bundleID must use reverse DNS notation")
	}
	buildNumber, err := strconv.ParseUint(values["buildNumber"], 10, 64)
	if err != nil || buildNumber == 0 {
		return fmt.Errorf("field buildNumber must be a positive integer")
	}
	return nil
}
