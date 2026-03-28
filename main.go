//go:build !solution

package main

import (
	"bufio"
	"bytes"
	"cmp"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"

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

type FullAuthorStats struct {
	Name    string
	Lines   int
	Commits map[string]struct{}
	Files   map[string]struct{}
}

type BriefAuthorStats struct {
	Name    string `json:"name"`
	Lines   int    `json:"lines"`
	Commits int    `json:"commits"`
	Files   int    `json:"files"`
}

func main() {
	cfg := parseCommand()
	langMap := make(map[string][]string)

	if len(cfg.Languages) > 0 {
		cfgPath, err := findLanguageConfigPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: failed to locate language config: %v\n", err)
			os.Exit(1)
		}
		langMap = loadLanguageMap(cfgPath)
	}

	filteredPaths := filterPaths(cfg, getPaths(cfg.Repository, cfg.Revision), langMap)
	statsMap := make(map[string]*FullAuthorStats)

	for _, file := range filteredPaths {
		err := processFile(cfg.Repository, cfg.Revision, file, cfg.UseCommitter, statsMap)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: error processing file %s: %v\n", file, err)
			os.Exit(1)
		}
	}

	results := rearrangeData(statsMap)
	sortResults(results, cfg.OrderBy)
	writeResults(results, cfg.Format)
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

func getPaths(repoPath string, revision string) []string {
	cmd := exec.Command("git", "ls-tree", "-r", "--name-only", revision)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		checkCmd := exec.Command("git", "log", "-1")
		checkCmd.Dir = repoPath
		if checkErr := checkCmd.Run(); checkErr != nil {
			return []string{}
		}
		fmt.Fprintf(os.Stderr, "error: failed to execute git ls-tree: %s\n", strings.TrimSpace(string(output)))
		os.Exit(1)
	}

	text := strings.TrimSpace(string(output))
	if text == "" {
		return []string{}
	}
	return strings.Split(text, "\n")
}

func filterPaths(cfg Config, paths []string, langMap map[string][]string) []string {
	var filteredPaths []string
	allowedExt := make(map[string]struct{})
	useExtFilter := (len(cfg.Extensions) > 0)

	for _, ext := range cfg.Extensions {
		allowedExt[ext] = struct{}{}
	}
	for _, lang := range cfg.Languages {
		extensions, exists := langMap[strings.ToLower(lang)]
		if !exists {
			fmt.Fprintf(os.Stderr, "warning: unknown language '%s'\n", lang)
			continue
		}
		useExtFilter = true
		for _, ext := range extensions {
			allowedExt[ext] = struct{}{}
		}
	}

	for _, file := range paths {
		// languages, extensions
		if useExtFilter {
			ext := filepath.Ext(file)
			if _, ok := allowedExt[ext]; !ok {
				continue
			}
		}
		// exclude
		if len(cfg.Exclude) > 0 && matchesAnyPattern(file, cfg.Exclude) {
			continue
		}
		// restrict-to
		if len(cfg.RestrictTo) > 0 && !matchesAnyPattern(file, cfg.RestrictTo) {
			continue
		}
		filteredPaths = append(filteredPaths, file)
	}

	return filteredPaths
}

func matchesAnyPattern(filePath string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: unknown glob pattern '%s': %v\n", pattern, err)
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

func processFile(repoPath, revision, file string, useCommitter bool,
	statsMap map[string]*FullAuthorStats) error {
	cmd := exec.Command("git", "blame", "--porcelain", revision, "--", file)
	cmd.Dir = repoPath
	output, err := cmd.Output()

	if err == nil && len(output) > 0 {
		if err := parseBlamePorcelain(output, file, useCommitter, statsMap); err != nil {
			return err
		}
	} else {
		logCmd := exec.Command("git", "log", "-1", "--format=%H%x00%an%x00%cn", revision, "--", file)
		logCmd.Dir = repoPath
		logOutput, logErr := logCmd.Output()
		if logErr != nil {
			return nil
		}

		logText := strings.TrimRight(string(logOutput), "\r\n")
		if logText == "" {
			return nil
		}

		parts := strings.Split(logText, "\x00")
		if len(parts) == 3 {
			hash, name := parts[0], parts[1]
			if useCommitter {
				name = parts[2]
			}
			updateStats(statsMap, name, hash, file, 0)
		}
	}
	return nil
}

func parseBlamePorcelain(output []byte, filename string, useCommitter bool,
	statsMap map[string]*FullAuthorStats) error {
	scanner := bufio.NewScanner(bytes.NewReader(output))
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)

	targetPrefix := "author "
	if useCommitter {
		targetPrefix = "committer "
	}

	var curCommitHash, curName string
	var numLines int
	commitToName := make(map[string]string)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "\t") {
			continue
		}
		fields := strings.Fields(line)

		if isBlameHeader(fields) {
			curCommitHash = fields[0]
			numLines = 0

			if len(fields) >= 4 {
				if n, err := strconv.Atoi(fields[3]); err == nil && n > 0 {
					numLines = n
				} else {
					numLines = 1
				}
			}

			if cachedName, ok := commitToName[curCommitHash]; ok && numLines > 0 {
				updateStats(statsMap, cachedName, curCommitHash, filename, numLines)
				numLines = 0
			}
			continue
		}

		if strings.HasPrefix(line, targetPrefix) {
			curName = strings.TrimPrefix(line, targetPrefix)
			commitToName[curCommitHash] = curName
			updateStats(statsMap, curName, curCommitHash, filename, numLines)
			numLines = 0
			continue
		}
	}

	return scanner.Err()
}

func updateStats(statsMap map[string]*FullAuthorStats, name, commitHash, filename string, numLines int) {
	stats, exists := statsMap[name]
	if !exists {
		stats = &FullAuthorStats{
			Name:    name,
			Lines:   0,
			Commits: make(map[string]struct{}),
			Files:   make(map[string]struct{}),
		}
		statsMap[name] = stats
	}

	stats.Lines += numLines
	stats.Commits[commitHash] = struct{}{}
	stats.Files[filename] = struct{}{}
}

func isBlameHeader(fields []string) bool {
	if len(fields) < 3 {
		return false
	}
	if !isUint(fields[1]) || !isUint(fields[2]) {
		return false
	}
	if len(fields) >= 4 && !isUint(fields[3]) {
		return false
	}
	return true
}

func isUint(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func rearrangeData(statsMap map[string]*FullAuthorStats) []BriefAuthorStats {
	result := make([]BriefAuthorStats, 0, len(statsMap))
	for _, author := range statsMap {
		result = append(result, BriefAuthorStats{
			Name:    author.Name,
			Lines:   author.Lines,
			Commits: len(author.Commits),
			Files:   len(author.Files),
		})
	}
	return result
}

func sortResults(results []BriefAuthorStats, orderBy string) {
	slices.SortFunc(results,
		func(a, b BriefAuthorStats) int {
			var res int
			switch orderBy {
			case "lines":
				res = cmp.Compare(b.Lines, a.Lines)
			case "commits":
				res = cmp.Compare(b.Commits, a.Commits)
			case "files":
				res = cmp.Compare(b.Files, a.Files)
			}
			if res != 0 {
				return res
			}

			if orderBy != "lines" {
				if r := cmp.Compare(b.Lines, a.Lines); r != 0 {
					return r
				}
			}
			if orderBy != "commits" {
				if r := cmp.Compare(b.Commits, a.Commits); r != 0 {
					return r
				}
			}
			if orderBy != "files" {
				if r := cmp.Compare(b.Files, a.Files); r != 0 {
					return r
				}
			}
			return cmp.Compare(a.Name, b.Name)
		},
	)
}

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
