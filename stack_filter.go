package cerror

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var defaultExclude = []string{"WithStack", "stack_filter.go"}

func NewStackTraceFilter(basePatterns, excludePatterns []string) *StackTraceFilter {
	if len(basePatterns) == 0 {
		basePatterns = []string{".*"}
	}
	if len(excludePatterns) == 0 {
		excludePatterns = defaultExclude
	} else {
		excludePatterns = append(excludePatterns, defaultExclude...)
	}
	compilePatterns := func(patterns []string) []*regexp.Regexp {
		compiled := make([]*regexp.Regexp, len(patterns))
		for i, pattern := range patterns {
			compiled[i] = regexp.MustCompile(pattern)
		}
		return compiled
	}

	return &StackTraceFilter{
		BasePatterns:    compilePatterns(basePatterns),
		ExcludePatterns: compilePatterns(excludePatterns),
	}
}

type StackTraceFilter struct {
	BasePatterns    []*regexp.Regexp
	ExcludePatterns []*regexp.Regexp
}

func (f *StackTraceFilter) AddBasePath(pattern string) {
	f.BasePatterns = append(f.BasePatterns, regexp.MustCompile(pattern))
}

func (f *StackTraceFilter) AddExcludePath(pattern string) {
	f.ExcludePatterns = append(f.ExcludePatterns, regexp.MustCompile(pattern))
}

func (f *StackTraceFilter) Filter(rawStack, format string) string {
	lines := strings.Split(rawStack, "\n")
	filtered := make([]string, 0)

	for i := 0; i < len(lines)-1; i += 2 {
		function := strings.TrimSpace(lines[i])
		location := strings.TrimSpace(lines[i+1])

		if f.isInBasePaths(location) && !f.isInExcludePaths(location) {
			filtered = append(filtered, fmt.Sprintf("%s (%s)", function, location))
		}
	}

	formattedStack := ""
	switch format {
	case "json":
		jsonFormat, err := formatToJSON(strings.Join(filtered, "\n"))
		if err != nil {
			return ""
		}
		formattedStack = string(jsonFormat)
	default:
		formattedStack = formatPlain(strings.Join(filtered, "\n"))

	}

	return formattedStack
}

func (f *StackTraceFilter) FilterOnlyBasePath(basePath, rawStack, format string) string {
	lines := strings.Split(rawStack, "\n")
	filtered := make([]string, 0)

	for i := 0; i < len(lines)-1; i += 2 {
		function := strings.TrimSpace(lines[i])
		location := strings.TrimSpace(lines[i+1])

		if f.isInBasePaths(location) && !f.isInExcludePaths(location) {
			filtered = append(filtered, fmt.Sprintf("%s (%s)", function, location))
		}
	}
	newFiltered := make([]string, len(filtered))
	for _, filter := range filtered {
		newFiltered = append(newFiltered, filter[strings.Index(filter, basePath):])
	}

	formattedStack := ""
	switch format {
	case "json":
		jsonFormat, err := formatToJSON(strings.Join(newFiltered, "\n"))
		if err != nil {
			return ""
		}
		formattedStack = string(jsonFormat)
	default:
		formattedStack = formatPlain(strings.Join(newFiltered, "\n"))

	}

	return formattedStack
}

func (f *StackTraceFilter) isInBasePaths(location string) bool {
	for _, pattern := range f.BasePatterns {
		if pattern.MatchString(location) {
			return true
		}
	}
	return false
}

func (f *StackTraceFilter) isInExcludePaths(location string) bool {
	for _, pattern := range f.ExcludePatterns {
		if pattern.MatchString(location) {
			return true
		}
	}
	return false
}

func formatToJSON(stackTrace string) ([]byte, error) {
	lines := strings.Split(stackTrace, "\n")
	formattedStack := []map[string]string{}

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, " ")
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid stack trace format: %s", line)
		}

		fn := strings.Split(parts[0], ".")
		lastFn := fn[len(fn)-1]

		//fileParts := strings.Split(parts[1], ":")
		//lineNumber := ""
		//if len(fileParts) > 1 {
		//	lineNumber = fileParts[1]
		//}
		formattedStack = append(formattedStack, map[string]string{
			"function": lastFn,
			"file":     parts[1],
			//"line":     lineNumber,
		})
	}

	return json.MarshalIndent(formattedStack, "", "  ")
}

// formatPlain
func formatPlain(stackTrace string) string {
	lines := strings.Split(stackTrace, "\n")
	formattedStack := []string{}

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, " ")
		if len(parts) < 2 {
			return ""
		}

		fn := strings.Split(parts[0], ".")
		lastFn := fn[len(fn)-1]

		formattedStack = append(formattedStack, fmt.Sprintf("fn:%s (%s)", lastFn, parts[1]))
	}

	return strings.Join(formattedStack, "\n")
}
