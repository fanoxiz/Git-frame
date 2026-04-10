package git

import (
	"bufio"
	"bytes"
	"strconv"
	"strings"

	"github.com/fanoxiz/Git-frame/internal/domain"
)

func parseBlamePorcelain(output []byte, filename string,
	useCommitter bool) ([]domain.FileFact, error) {
	scanner := bufio.NewScanner(bytes.NewReader(output))
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)

	targetPrefix := "author "
	if useCommitter {
		targetPrefix = "committer "
	}

	var curCommitHash string
	var numLines int
	commitToName := make(map[string]string)
	facts := make([]domain.FileFact, 0)

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
				facts = append(facts, domain.FileFact{
					Name:       cachedName,
					CommitHash: curCommitHash,
					Filename:   filename,
					Lines:      numLines,
				})
				numLines = 0
			}
			continue
		}

		if strings.HasPrefix(line, targetPrefix) {
			name := strings.TrimPrefix(line, targetPrefix)
			commitToName[curCommitHash] = name
			facts = append(facts, domain.FileFact{
				Name:       name,
				CommitHash: curCommitHash,
				Filename:   filename,
				Lines:      numLines,
			})
			numLines = 0
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return facts, nil
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
