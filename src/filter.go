//go:build !solution

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
		if useExtFilter {
			ext := filepath.Ext(file)
			if _, ok := allowedExt[ext]; !ok {
				continue
			}
		}
		if len(cfg.Exclude) > 0 && matchesAnyPattern(file, cfg.Exclude) {
			continue
		}
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
