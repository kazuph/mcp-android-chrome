package format

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/kazuph/mcp-android-chrome/internal/loader"
)

// Format represents the output format type
type Format string

const (
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
)

// TabFormatter handles formatting of tab data in different formats
type TabFormatter struct {
	format Format
}

// NewTabFormatter creates a new tab formatter
func NewTabFormatter(format Format) *TabFormatter {
	return &TabFormatter{
		format: format,
	}
}

// FormatTabs formats a slice of tabs in the specified format
func (f *TabFormatter) FormatTabs(tabs []loader.Tab) (string, error) {
	switch f.format {
	case FormatJSON:
		return f.formatJSON(tabs)
	case FormatYAML:
		return f.formatYAML(tabs)
	default:
		return "", fmt.Errorf("unsupported format: %s", f.format)
	}
}

// FormatSingleTab formats a single tab in the specified format
func (f *TabFormatter) FormatSingleTab(tab loader.Tab) (string, error) {
	return f.FormatTabs([]loader.Tab{tab})
}

// formatJSON formats tabs as pretty-printed JSON
func (f *TabFormatter) formatJSON(tabs []loader.Tab) (string, error) {
	data, err := json.MarshalIndent(tabs, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(data), nil
}

// formatYAML formats tabs as YAML
func (f *TabFormatter) formatYAML(tabs []loader.Tab) (string, error) {
	data, err := yaml.Marshal(tabs)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}
	return string(data), nil
}

// ParseFormat parses a format string and returns the Format enum
func ParseFormat(formatStr string) (Format, error) {
	switch formatStr {
	case "json", "JSON":
		return FormatJSON, nil
	case "yaml", "YAML", "yml", "YML":
		return FormatYAML, nil
	default:
		return FormatJSON, fmt.Errorf("unsupported format: %s (supported: json, yaml)", formatStr)
	}
}

// GetMimeType returns the MIME type for the format
func (f *TabFormatter) GetMimeType() string {
	switch f.format {
	case FormatJSON:
		return "application/json"
	case FormatYAML:
		return "application/x-yaml"
	default:
		return "text/plain"
	}
}

// DefaultFormatter returns a JSON formatter (backward compatibility)
func DefaultFormatter() *TabFormatter {
	return NewTabFormatter(FormatJSON)
}

// YAMLFormatter returns a YAML formatter
func YAMLFormatter() *TabFormatter {
	return NewTabFormatter(FormatYAML)
}

// JSONFormatter returns a JSON formatter
func JSONFormatter() *TabFormatter {
	return NewTabFormatter(FormatJSON)
}

// SearchResult represents a search result with relevance scoring
type SearchResult struct {
	Tab   loader.Tab `json:"tab" yaml:"tab"`
	Score float64    `json:"score" yaml:"score"`
}

// FormatSearchResults formats search results in the specified format
func (f *TabFormatter) FormatSearchResults(results interface{}) (string, error) {
	switch f.format {
	case FormatJSON:
		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal search results as JSON: %w", err)
		}
		return string(data), nil
	case FormatYAML:
		data, err := yaml.Marshal(results)
		if err != nil {
			return "", fmt.Errorf("failed to marshal search results as YAML: %w", err)
		}
		return string(data), nil
	default:
		return "", fmt.Errorf("unsupported format: %s", f.format)
	}
}