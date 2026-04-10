package filter

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/fanoxiz/Git-frame/internal/domain"
)

func FilterPaths(cfg domain.FilterOptions, paths []string, langMap map[string][]string) []string {
	var filteredPaths []string
	allowedExt := make(map[string]struct{})
	useExtFilter := (len(cfg.Extensions) > 0)

	for _, ext := range cfg.Extensions {
		normalizedExt := normalizeExtension(ext)
		if normalizedExt == "" {
			continue
		}
		allowedExt[normalizedExt] = struct{}{}
	}
	for _, lang := range cfg.Languages {
		extensions, exists := langMap[strings.ToLower(lang)]
		if !exists {
			fmt.Fprintf(os.Stderr, "warning: unknown language '%s'\n", lang)
			continue
		}
		useExtFilter = true
		for _, ext := range extensions {
			normalizedExt := normalizeExtension(ext)
			if normalizedExt == "" {
				continue
			}
			allowedExt[normalizedExt] = struct{}{}
		}
	}

	for _, file := range paths {
		if useExtFilter {
			ext := strings.ToLower(filepath.Ext(file))
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

func normalizeExtension(ext string) string {
	ext = strings.TrimSpace(strings.ToLower(ext))
	if ext == "" {
		return ""
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return ext
}

func matchesAnyPattern(filePath string, patterns []string) bool {
	normalizedPath := filepath.ToSlash(filePath)
	for _, pattern := range patterns {
		normalizedPattern := filepath.ToSlash(pattern)
		matched, err := path.Match(normalizedPattern, normalizedPath)
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
