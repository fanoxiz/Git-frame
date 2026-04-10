package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"text/tabwriter"

	"github.com/fanoxiz/Git-frame/internal/domain"
)

func WriteResults(results []domain.BriefAuthorStats, format string, writer io.Writer) error {
	switch format {
	case "tabular":
		return printTabular(results, writer)
	case "csv":
		return printCSV(results, writer)
	case "json":
		return printJSON(results, writer)
	case "json-lines":
		return printJSONLines(results, writer)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func printTabular(results []domain.BriefAuthorStats, writer io.Writer) error {
	w := tabwriter.NewWriter(writer, 0, 0, 1, ' ', tabwriter.StripEscape)
	if _, err := fmt.Fprintln(w, "Name\tLines\tCommits\tFiles"); err != nil {
		return err
	}
	for _, res := range results {
		if _, err := fmt.Fprintf(w, "\xff%s\xff\t%d\t%d\t%d\n", res.Name, res.Lines, res.Commits, res.Files); err != nil {
			return err
		}
	}
	return w.Flush()
}

func printCSV(results []domain.BriefAuthorStats, writer io.Writer) error {
	w := csv.NewWriter(writer)
	if err := w.Write([]string{"Name", "Lines", "Commits", "Files"}); err != nil {
		return err
	}
	for _, res := range results {
		if err := w.Write([]string{
			res.Name,
			strconv.Itoa(res.Lines),
			strconv.Itoa(res.Commits),
			strconv.Itoa(res.Files),
		}); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printJSON(results []domain.BriefAuthorStats, writer io.Writer) error {
	return json.NewEncoder(writer).Encode(results)
}

func printJSONLines(results []domain.BriefAuthorStats, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	for _, res := range results {
		if err := encoder.Encode(res); err != nil {
			return err
		}
	}
	return nil
}
