package stats

import (
	"cmp"
	"slices"

	"github.com/fanoxiz/Git-frame/internal/domain"
)

func BuildResults(facts []domain.FileFact, orderBy string) []domain.BriefAuthorStats {
	statsMap := make(map[string]*domain.FullAuthorStats)
	for _, fact := range facts {
		updateStats(statsMap, fact.Name, fact.CommitHash, fact.Filename, fact.Lines)
	}

	results := rearrangeData(statsMap)
	sortResults(results, orderBy)
	return results
}

func updateStats(statsMap map[string]*domain.FullAuthorStats, name, commitHash, filename string, numLines int) {
	stats, exists := statsMap[name]
	if !exists {
		stats = &domain.FullAuthorStats{
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

func rearrangeData(statsMap map[string]*domain.FullAuthorStats) []domain.BriefAuthorStats {
	result := make([]domain.BriefAuthorStats, 0, len(statsMap))
	for _, author := range statsMap {
		result = append(result, domain.BriefAuthorStats{
			Name:    author.Name,
			Lines:   author.Lines,
			Commits: len(author.Commits),
			Files:   len(author.Files),
		})
	}
	return result
}

func sortResults(results []domain.BriefAuthorStats, orderBy string) {
	slices.SortFunc(results,
		func(a, b domain.BriefAuthorStats) int {
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
