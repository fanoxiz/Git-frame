package config

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

type Language struct {
	Name       string   `json:"name"`
	Extensions []string `json:"extensions"`
}

func ParseCommand() (Config, error) {
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
		return cfg, fmt.Errorf("invalid order-by value: %s", cfg.OrderBy)
	}
	if !slices.Contains([]string{"tabular", "csv", "json", "json-lines"}, cfg.Format) {
		return cfg, fmt.Errorf("invalid format value: %s", cfg.Format)
	}
	return cfg, nil
}

func FindLanguageConfigPath() (string, error) {
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

func LoadLanguageMap(jsonPath string) (map[string][]string, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read language config: %w", err)
	}

	var languages []Language
	if err := json.Unmarshal(data, &languages); err != nil {
		return nil, fmt.Errorf("failed to parse language config: %w", err)
	}

	langMap := make(map[string][]string)
	for _, lang := range languages {
		langMap[strings.ToLower(lang.Name)] = lang.Extensions
	}
	return langMap, nil
}
