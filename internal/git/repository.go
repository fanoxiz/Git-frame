package git

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/fanoxiz/Git-frame/internal/domain"
)

func GetPaths(repoPath string, revision string) ([]string, error) {
	cmd := exec.Command("git", "ls-tree", "-r", "--name-only", revision)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		if !isGitRepository(repoPath) {
			return nil, fmt.Errorf("repository is not a git repository: %s", repoPath)
		}

		hasCommits, checkErr := hasAnyCommit(repoPath)
		if checkErr != nil {
			return nil, checkErr
		}
		if !hasCommits {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to execute git ls-tree: %s", strings.TrimSpace(string(output)))
	}

	text := strings.TrimSpace(string(output))
	if text == "" {
		return []string{}, nil
	}
	return strings.Split(text, "\n"), nil
}

func isGitRepository(repoPath string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(output)) == "true"
}

func hasAnyCommit(repoPath string) (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--verify", "HEAD")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return false, nil
	}

	return true, nil
}

func CollectFileFacts(repoPath, revision, file string, useCommitter bool) ([]domain.FileFact, error) {
	cmd := exec.Command("git", "blame", "--porcelain", revision, "--", file)
	cmd.Dir = repoPath
	output, err := cmd.Output()

	if err == nil && len(output) > 0 {
		return parseBlamePorcelain(output, file, useCommitter)
	}

	logCmd := exec.Command("git", "log", "-1", "--format=%H%x00%an%x00%cn", revision, "--", file)
	logCmd.Dir = repoPath
	logOutput, logErr := logCmd.Output()
	if logErr != nil {
		return nil, nil
	}

	logText := strings.TrimRight(string(logOutput), "\r\n")
	if logText == "" {
		return nil, nil
	}

	parts := strings.Split(logText, "\x00")
	if len(parts) == 3 {
		hash, name := parts[0], parts[1]
		if useCommitter {
			name = parts[2]
		}

		return []domain.FileFact{{
			Name:       name,
			CommitHash: hash,
			Filename:   file,
			Lines:      0,
		}}, nil
	}

	return nil, nil
}
