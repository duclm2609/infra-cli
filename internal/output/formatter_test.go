package output

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"gopkg.in/yaml.v3"
)

// Feature: devops-cli-tool, Property 3: JSON Output Round-Trip
// For any valid data structure, formatting it as JSON and then parsing
// the JSON output SHALL produce a data structure equivalent to the original.
func TestJSONRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("JSON format is reversible for string maps", prop.ForAll(
		func(data map[string]string) bool {
			formatted, err := FormatJSON_Data(data)
			if err != nil {
				return false
			}

			var parsed map[string]string
			err = json.Unmarshal([]byte(formatted), &parsed)
			if err != nil {
				return false
			}

			// Compare maps
			if len(data) != len(parsed) {
				return false
			}
			for k, v := range data {
				if parsed[k] != v {
					return false
				}
			}
			return true
		},
		gen.MapOf(gen.Identifier(), gen.Identifier()),
	))

	properties.Property("JSON output is valid JSON", prop.ForAll(
		func(key, value string) bool {
			data := map[string]string{key: value}
			formatted, err := FormatJSON_Data(data)
			if err != nil {
				return false
			}

			var parsed interface{}
			return json.Unmarshal([]byte(formatted), &parsed) == nil
		},
		gen.Identifier(),
		gen.Identifier(),
	))

	properties.TestingRun(t)
}

// Feature: devops-cli-tool, Property 4: YAML Output Round-Trip
// For any valid data structure, formatting it as YAML and then parsing
// the YAML output SHALL produce a data structure equivalent to the original.
func TestYAMLRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("YAML format is reversible for string maps", prop.ForAll(
		func(data map[string]string) bool {
			formatted, err := FormatYAML_Data(data)
			if err != nil {
				return false
			}

			var parsed map[string]string
			err = yaml.Unmarshal([]byte(formatted), &parsed)
			if err != nil {
				return false
			}

			// Compare maps
			if len(data) != len(parsed) {
				return false
			}
			for k, v := range data {
				if parsed[k] != v {
					return false
				}
			}
			return true
		},
		gen.MapOf(gen.Identifier(), gen.Identifier()),
	))

	properties.Property("YAML output is valid YAML", prop.ForAll(
		func(key, value string) bool {
			data := map[string]string{key: value}
			formatted, err := FormatYAML_Data(data)
			if err != nil {
				return false
			}

			var parsed interface{}
			return yaml.Unmarshal([]byte(formatted), &parsed) == nil
		},
		gen.Identifier(),
		gen.Identifier(),
	))

	properties.TestingRun(t)
}

// Feature: devops-cli-tool, Property 5: Table Column Alignment
// For any tabular data with multiple rows, all cells in the same column
// SHALL start at the same horizontal position in the formatted output.
func TestTableColumnAlignment(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("table columns are aligned", prop.ForAll(
		func(rows [][]string) bool {
			if len(rows) < 2 {
				return true // Need at least header + 1 row
			}

			// Ensure all rows have same number of columns
			numCols := len(rows[0])
			for _, row := range rows {
				if len(row) != numCols {
					return true // Skip invalid input
				}
			}

			data := TableData{
				Headers: rows[0],
				Rows:    rows[1:],
			}

			formatted, err := FormatTableData(data)
			if err != nil {
				return false
			}

			// Check that all lines have consistent column positions
			lines := strings.Split(formatted, "\n")
			if len(lines) < 2 {
				return true
			}

			// Find pipe positions in first data line
			var pipePositions []int
			for i, line := range lines {
				if strings.Contains(line, "|") {
					positions := findPipePositions(line)
					if i == 0 {
						pipePositions = positions
					} else {
						// All lines should have pipes at same positions
						if len(positions) != len(pipePositions) {
							return false
						}
						for j, pos := range positions {
							if pos != pipePositions[j] {
								return false
							}
						}
					}
				}
			}

			return true
		},
		gen.SliceOfN(3, gen.SliceOfN(2, gen.Identifier())),
	))

	properties.TestingRun(t)
}

func findPipePositions(line string) []int {
	var positions []int
	for i, ch := range line {
		if ch == '|' {
			positions = append(positions, i)
		}
	}
	return positions
}

// Unit tests
func TestNewFormatter(t *testing.T) {
	f := NewFormatter("json")
	if f.GetFormat() != FormatJSON {
		t.Errorf("Expected JSON format, got %v", f.GetFormat())
	}

	f = NewFormatter("yaml")
	if f.GetFormat() != FormatYAML {
		t.Errorf("Expected YAML format, got %v", f.GetFormat())
	}

	f = NewFormatter("table")
	if f.GetFormat() != FormatTable {
		t.Errorf("Expected Table format, got %v", f.GetFormat())
	}

	f = NewFormatter("")
	if f.GetFormat() != FormatTable {
		t.Errorf("Expected Table format for empty string, got %v", f.GetFormat())
	}
}

func TestSetFormatInvalid(t *testing.T) {
	f := NewFormatter("table")
	err := f.SetFormat("invalid")
	if err == nil {
		t.Error("Expected error for invalid format")
	}
}

func TestFormatJSONSimple(t *testing.T) {
	data := map[string]string{"name": "test", "value": "123"}
	result, err := FormatJSON_Data(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(result, "name") || !strings.Contains(result, "test") {
		t.Error("JSON output missing expected content")
	}
}

func TestFormatYAMLSimple(t *testing.T) {
	data := map[string]string{"name": "test", "value": "123"}
	result, err := FormatYAML_Data(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(result, "name:") || !strings.Contains(result, "test") {
		t.Error("YAML output missing expected content")
	}
}

func TestFormatTableSimple(t *testing.T) {
	data := TableData{
		Headers: []string{"Name", "Value"},
		Rows: [][]string{
			{"test1", "123"},
			{"test2", "456"},
		},
	}
	result, err := FormatTableData(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// tablewriter v1.1.3 uses uppercase headers by default
	if !strings.Contains(strings.ToUpper(result), "NAME") || !strings.Contains(result, "test1") {
		t.Errorf("Table output missing expected content. Got:\n%s", result)
	}
}
