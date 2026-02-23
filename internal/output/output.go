package output

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
)

// Print writes data in the requested format to w.
// For "json" format, data is marshaled as JSON.
// For "table" format, data must be a Table.
func Print(w io.Writer, format string, data any) error {
	switch format {
	case "json":
		return printJSON(w, data)
	case "table":
		t, ok := data.(*Table)
		if !ok {
			return fmt.Errorf("table format requires *output.Table, got %T", data)
		}
		return printTable(w, t)
	default:
		return fmt.Errorf("unknown output format: %s", format)
	}
}

func printJSON(w io.Writer, data any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// Table holds headers and rows for tabwriter output.
type Table struct {
	Headers []string
	Rows    [][]string
}

func printTable(w io.Writer, t *Table) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	for i, h := range t.Headers {
		if i > 0 {
			fmt.Fprint(tw, "\t")
		}
		fmt.Fprint(tw, h)
	}
	fmt.Fprintln(tw)

	for _, row := range t.Rows {
		for i, col := range row {
			if i > 0 {
				fmt.Fprint(tw, "\t")
			}
			fmt.Fprint(tw, col)
		}
		fmt.Fprintln(tw)
	}

	return tw.Flush()
}
