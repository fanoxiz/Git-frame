//go:build !solution

package main

import (
	"fmt"
	"os"
)

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

	paths := getPaths(cfg.Repository, cfg.Revision)
	filteredPaths := filterPaths(cfg, paths, langMap)
	
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
