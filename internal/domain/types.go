package domain

type FilterOptions struct {
	Extensions []string
	Languages  []string
	Exclude    []string
	RestrictTo []string
}

type FileFact struct {
	Name       string
	CommitHash string
	Filename   string
	Lines      int
}

type BriefAuthorStats struct {
	Name    string `json:"name"`
	Lines   int    `json:"lines"`
	Commits int    `json:"commits"`
	Files   int    `json:"files"`
}

type FullAuthorStats struct {
	Name    string
	Lines   int
	Commits map[string]struct{}
	Files   map[string]struct{}
}
