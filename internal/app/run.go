package app

import (
	"fmt"
	"os"

	"github.com/fanoxiz/Git-frame/internal/config"
	"github.com/fanoxiz/Git-frame/internal/domain"
	"github.com/fanoxiz/Git-frame/internal/filter"
	"github.com/fanoxiz/Git-frame/internal/git"
	"github.com/fanoxiz/Git-frame/internal/output"
	"github.com/fanoxiz/Git-frame/internal/stats"
)

func Run() error {
	cfg, err := config.ParseCommand()
	if err != nil {
		return err
	}

	langMap := make(map[string][]string)
	if len(cfg.Languages) > 0 {
		cfgPath, err := config.FindLanguageConfigPath(cfg.Repository)
		if err != nil {
			return fmt.Errorf("failed to locate language config: %w", err)
		}

		langMap, err = config.LoadLanguageMap(cfgPath)
		if err != nil {
			return err
		}
	}

	paths, err := git.GetPaths(cfg.Repository, cfg.Revision)
	if err != nil {
		return err
	}

	filterOptions := domain.FilterOptions{
		Extensions: cfg.Extensions,
		Languages:  cfg.Languages,
		Exclude:    cfg.Exclude,
		RestrictTo: cfg.RestrictTo,
	}

	filteredPaths := filter.FilterPaths(filterOptions, paths, langMap)
	facts := make([]domain.FileFact, 0, len(filteredPaths))

	for _, file := range filteredPaths {
		fileFacts, err := git.CollectFileFacts(cfg.Repository, cfg.Revision, file, cfg.UseCommitter)
		if err != nil {
			return fmt.Errorf("error processing file %s: %w", file, err)
		}
		facts = append(facts, fileFacts...)
	}

	results := stats.BuildResults(facts, cfg.OrderBy)
	return output.WriteResults(results, cfg.Format, os.Stdout)
}
