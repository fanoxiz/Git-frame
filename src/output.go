//go:build !solution

package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
)

func writeResults(results []BriefAuthorStats, format string) {
	switch format {
	case "tabular":
		printTabular(results)
	case "csv":
		printCSV(results)
	case "json":
		printJSON(results)
	case "json-lines":
		printJSONLines(results)
	}
}

func printTabular(results []BriefAuthorStats) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.StripEscape)
	fmt.Fprintln(w, "Name\tLines\tCommits\tFiles")
	for _, res := range results {
		fmt.Fprintf(w, "\xff%s\xff\t%d\t%d\t%d\n", res.Name, res.Lines, res.Commits, res.Files)
	}
	w.Flush()
}

func printCSV(results []BriefAuthorStats) {
	w := csv.NewWriter(os.Stdout)
	_ = w.Write([]string{"Name", "Lines", "Commits", "Files"})
	for _, res := range results {
		_ = w.Write([]string{
			res.Name,
			strconv.Itoa(res.Lines),
			strconv.Itoa(res.Commits),
			strconv.Itoa(res.Files),
		})
	}
	w.Flush()
}

func printJSON(results []BriefAuthorStats) {
	_ = json.NewEncoder(os.Stdout).Encode(results)
}

func printJSONLines(results []BriefAuthorStats) {
	encoder := json.NewEncoder(os.Stdout)
	for _, res := range results {
		_ = encoder.Encode(res)
	}
}
