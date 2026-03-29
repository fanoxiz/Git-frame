//go:build !solution

package main

import (
	"cmp"
	"slices"
)

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
