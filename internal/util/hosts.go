package util

import (
	"bufio"
	"strings"
)

func ValidateIsHostFileFormat(content string) ([]string, bool) {
	hosts := make([]string, 0)
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Empty lines are valid
		if line == "" {
			continue
		}

		// Comment lines are valid
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Entry lines must have at least 2 whitespace-separated fields
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return nil, false
		}
		hosts = append(hosts, fields[1])
	}

	return hosts, true
}
