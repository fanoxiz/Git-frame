//go:build !solution

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

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

func processFile(repoPath, revision, file string, useCommitter bool, statsMap map[string]*FullAuthorStats) error {
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

func parseBlamePorcelain(output []byte, filename string, useCommitter bool, statsMap map[string]*FullAuthorStats) error {
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
