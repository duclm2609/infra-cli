package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"gopkg.in/yaml.v3"
)

// Format represents supported output formats
type Format string

const (
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
	FormatTable Format = "table"
)

// Formatter handles output formatting in multiple formats
type Formatter struct {
	format Format
	writer io.Writer
}

// NewFormatter creates a new output formatter
func NewFormatter(format string) *Formatter {
	f := &Formatter{
		writer: os.Stdout,
	}
	f.SetFormat(format)
	return f
}

// SetFormat sets the output format
func (f *Formatter) SetFormat(format string) error {
	switch strings.ToLower(format) {
	case "json":
		f.format = FormatJSON
	case "yaml", "yml":
		f.format = FormatYAML
	case "table", "":
		f.format = FormatTable
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
	return nil
}

// SetWriter sets the output writer
func (f *Formatter) SetWriter(w io.Writer) {
	f.writer = w
}

// GetFormat returns the current format
func (f *Formatter) GetFormat() Format {
	return f.format
}

// Format converts data to the specified format and returns as string
func (f *Formatter) Format(data interface{}) (string, error) {
	switch f.format {
	case FormatJSON:
		return FormatJSON_Data(data)
	case FormatYAML:
		return FormatYAML_Data(data)
	case FormatTable:
		return FormatTable_Data(data)
	default:
		return FormatTable_Data(data)
	}
}

// Print formats and prints data to the configured writer
func (f *Formatter) Print(data interface{}) error {
	output, err := f.Format(data)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(f.writer, output)
	return err
}

// FormatJSON_Data formats data as indented JSON
func FormatJSON_Data(data interface{}) (string, error) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format as JSON: %w", err)
	}
	return string(bytes), nil
}

// FormatYAML_Data formats data as YAML
func FormatYAML_Data(data interface{}) (string, error) {
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to format as YAML: %w", err)
	}
	return strings.TrimSpace(string(bytes)), nil
}

// FormatTable_Data formats data as an aligned table
func FormatTable_Data(data interface{}) (string, error) {
	var builder strings.Builder
	table := tablewriter.NewTable(&builder,
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Formatting: tw.CellFormatting{
					AutoWrap: tw.WrapNone,
				},
			},
		}),
	)

	switch v := data.(type) {
	case []map[string]interface{}:
		if len(v) == 0 {
			return "No data", nil
		}
		// Extract headers from first row
		headers := extractHeaders(v[0])
		table.Header(toInterfaceSlice(headers)...)

		// Add rows
		for _, row := range v {
			rowData := make([]interface{}, len(headers))
			for i, h := range headers {
				rowData[i] = fmt.Sprintf("%v", row[h])
			}
			table.Append(rowData...)
		}

	case map[string]interface{}:
		table.Header("Key", "Value")
		for k, val := range v {
			table.Append(k, fmt.Sprintf("%v", val))
		}

	case [][]string:
		if len(v) == 0 {
			return "No data", nil
		}
		// First row is headers
		if len(v) > 0 {
			table.Header(toInterfaceSlice(v[0])...)
		}
		// Rest are data rows
		for i := 1; i < len(v); i++ {
			table.Append(toInterfaceSlice(v[i])...)
		}

	default:
		// For simple types, just return string representation
		return fmt.Sprintf("%v", data), nil
	}

	table.Render()

	return strings.TrimSpace(builder.String()), nil
}

// toInterfaceSlice converts a string slice to interface slice
func toInterfaceSlice(s []string) []interface{} {
	result := make([]interface{}, len(s))
	for i, v := range s {
		result[i] = v
	}
	return result
}

// extractHeaders extracts keys from a map in a consistent order
func extractHeaders(m map[string]interface{}) []string {
	headers := make([]string, 0, len(m))
	for k := range m {
		headers = append(headers, k)
	}
	return headers
}

// TableData represents structured data for table output
type TableData struct {
	Headers []string
	Rows    [][]string
}

// FormatTableData formats TableData as an aligned table
func FormatTableData(data TableData) (string, error) {
	var builder strings.Builder
	table := tablewriter.NewTable(&builder,
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Formatting: tw.CellFormatting{
					AutoWrap: tw.WrapNone,
				},
			},
		}),
	)

	table.Header(toInterfaceSlice(data.Headers)...)
	for _, row := range data.Rows {
		table.Append(toInterfaceSlice(row)...)
	}

	table.Render()

	return strings.TrimSpace(builder.String()), nil
}
