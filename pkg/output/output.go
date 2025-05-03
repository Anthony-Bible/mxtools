// Package output provides output formatting functionality for the MXToolbox clone.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

// Format represents the output format.
type Format string

const (
	// FormatText is the text output format.
	FormatText Format = "text"
	// FormatJSON is the JSON output format.
	FormatJSON Format = "json"
)

// Formatter is an interface for formatting output.
type Formatter interface {
	// Format formats the data and writes it to the writer.
	Format(w io.Writer, data interface{}) error
}

// TextFormatter is a formatter for text output.
type TextFormatter struct {
	// Indent is the indentation string.
	Indent string
}

// JSONFormatter is a formatter for JSON output.
type JSONFormatter struct {
	// Indent is the indentation string.
	Indent string
	// Prefix is the prefix string.
	Prefix string
}

// NewFormatter creates a new formatter for the specified format.
func NewFormatter(format Format) Formatter {
	switch format {
	case FormatJSON:
		return &JSONFormatter{
			Indent: "  ",
			Prefix: "",
		}
	default:
		return &TextFormatter{
			Indent: "  ",
		}
	}
}

// Format formats the data as text and writes it to the writer.
func (f *TextFormatter) Format(w io.Writer, data interface{}) error {
	// Check if the data implements the Stringer interface
	if stringer, ok := data.(fmt.Stringer); ok {
		_, err := fmt.Fprintln(w, stringer.String())
		return err
	}

	// Check if the data is a map
	if m, ok := data.(map[string]interface{}); ok {
		return f.formatMap(w, m, 0)
	}

	// Check if the data is a slice
	if s, ok := data.([]interface{}); ok {
		return f.formatSlice(w, s, 0)
	}

	// Default to fmt.Println
	_, err := fmt.Fprintln(w, data)
	return err
}

// formatMap formats a map as text and writes it to the writer.
func (f *TextFormatter) formatMap(w io.Writer, m map[string]interface{}, level int) error {
	indent := strings.Repeat(f.Indent, level)
	for k, v := range m {
		// Format the key
		_, err := fmt.Fprintf(w, "%s%s: ", indent, k)
		if err != nil {
			return err
		}

		// Format the value
		switch val := v.(type) {
		case map[string]interface{}:
			_, err := fmt.Fprintln(w)
			if err != nil {
				return err
			}
			err = f.formatMap(w, val, level+1)
			if err != nil {
				return err
			}
		case []interface{}:
			_, err := fmt.Fprintln(w)
			if err != nil {
				return err
			}
			err = f.formatSlice(w, val, level+1)
			if err != nil {
				return err
			}
		default:
			_, err := fmt.Fprintln(w, val)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// formatSlice formats a slice as text and writes it to the writer.
func (f *TextFormatter) formatSlice(w io.Writer, s []interface{}, level int) error {
	indent := strings.Repeat(f.Indent, level)
	for i, v := range s {
		// Format the index
		_, err := fmt.Fprintf(w, "%s%d: ", indent, i)
		if err != nil {
			return err
		}

		// Format the value
		switch val := v.(type) {
		case map[string]interface{}:
			_, err := fmt.Fprintln(w)
			if err != nil {
				return err
			}
			err = f.formatMap(w, val, level+1)
			if err != nil {
				return err
			}
		case []interface{}:
			_, err := fmt.Fprintln(w)
			if err != nil {
				return err
			}
			err = f.formatSlice(w, val, level+1)
			if err != nil {
				return err
			}
		default:
			_, err := fmt.Fprintln(w, val)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Format formats the data as JSON and writes it to the writer.
func (f *JSONFormatter) Format(w io.Writer, data interface{}) error {
	// Marshal the data to JSON
	jsonData, err := json.MarshalIndent(data, f.Prefix, f.Indent)
	if err != nil {
		return err
	}

	// Write the JSON data to the writer
	_, err = w.Write(jsonData)
	if err != nil {
		return err
	}

	// Write a newline
	_, err = fmt.Fprintln(w)
	return err
}

// FormatTable formats tabular data and writes it to the writer.
func FormatTable(w io.Writer, headers []string, rows [][]string) error {
	// Create a tabwriter
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	// Write the headers
	for i, header := range headers {
		if i > 0 {
			fmt.Fprint(tw, "\t")
		}
		fmt.Fprint(tw, header)
	}
	fmt.Fprintln(tw)

	// Write a separator
	for i, header := range headers {
		if i > 0 {
			fmt.Fprint(tw, "\t")
		}
		fmt.Fprint(tw, strings.Repeat("-", len(header)))
	}
	fmt.Fprintln(tw)

	// Write the rows
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				fmt.Fprint(tw, "\t")
			}
			fmt.Fprint(tw, cell)
		}
		fmt.Fprintln(tw)
	}

	// Flush the tabwriter
	return tw.Flush()
}

// WriteResult writes a result to the specified writer in the specified format.
func WriteResult(w io.Writer, data interface{}, format Format) error {
	formatter := NewFormatter(format)
	return formatter.Format(w, data)
}

// PrintResult prints a result to stdout in the specified format.
func PrintResult(data interface{}, format Format) error {
	return WriteResult(os.Stdout, data, format)
}