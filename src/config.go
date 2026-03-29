//go:build !solution

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	flag "github.com/spf13/pflag"
)

type Config struct {
	Repository   string
	Revision     string
	OrderBy      string
	UseCommitter bool
	Format       string

	Extensions []string
	Languages  []string
	Exclude    []string
	RestrictTo []string
}

func parseCommand() Config {
	var cfg Config

	flag.StringVar(&cfg.Repository, "repository", ".", "path to git repository")
	flag.StringVar(&cfg.Revision, "revision", "HEAD", "commit pointer")
	flag.StringVar(&cfg.OrderBy, "order-by", "lines", "sort key: lines, commits, files")
	flag.BoolVar(&cfg.UseCommitter, "use-committer", false, "use committer instead of author")
	flag.StringVar(&cfg.Format, "format", "tabular", "output format: tabular, csv, json, json-lines")

	flag.StringSliceVar(&cfg.Extensions, "extensions", []string{}, "list of extensions")
	flag.StringSliceVar(&cfg.Languages, "languages", []string{}, "list of languages")
	flag.StringSliceVar(&cfg.Exclude, "exclude", []string{}, "glob patterns to exclude")
	flag.StringSliceVar(&cfg.RestrictTo, "restrict-to", []string{}, "glob patterns to restrict to")

	flag.Parse()

	if !slices.Contains([]string{"lines", "commits", "files"}, cfg.OrderBy) {
		fmt.Fprintf(os.Stderr, "error: invalid order-by value: %s\n", cfg.OrderBy)
		os.Exit(1)
	}
	if !slices.Contains([]string{"tabular", "csv", "json", "json-lines"}, cfg.Format) {
		fmt.Fprintf(os.Stderr, "error: invalid format value: %s\n", cfg.Format)
		os.Exit(1)
	}
	return cfg
}

func findLanguageConfigPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		p := filepath.Join(dir, "configs", "language_extensions.json")
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("configs/language_extensions.json not found from %s up to root", wd)
}

func loadLanguageMap(jsonPath string) map[string][]string {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to read language config: %v\n", err)
		os.Exit(1)
	}
	type Language struct {
		Name       string   `json:"name"`
		Extensions []string `json:"extensions"`
	}

	var languages []Language
	if err := json.Unmarshal(data, &languages); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to parse language config: %v\n", err)
		os.Exit(1)
	}

	langMap := make(map[string][]string)
	for _, lang := range languages {
		langMap[strings.ToLower(lang.Name)] = lang.Extensions
	}
	return langMap
}
