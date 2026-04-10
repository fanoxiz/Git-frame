package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

var (
	repoRoot   string
	binaryPath string
	buildDir   string
)

func TestMain(m *testing.M) {
	if err := setupBinary(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	code := m.Run()
	if buildDir != "" {
		_ = os.RemoveAll(buildDir)
	}
	os.Exit(code)
}

func setupBinary() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	repoRoot, err = findRepoRoot(wd)
	if err != nil {
		return err
	}

	buildDir, err = os.MkdirTemp("", "gitframe-bin-")
	if err != nil {
		return err
	}

	binName := "gitframe-testbin"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binaryPath = filepath.Join(buildDir, binName)

	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", binaryPath, "./cmd/gitframe")
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to build gitframe binary: %w\n%s", err, string(output))
	}

	return nil
}

func findRepoRoot(startDir string) (string, error) {
	dir := startDir
	for {
		candidate := filepath.Join(dir, "go.mod")
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s up to filesystem root", startDir)
		}
		dir = parent
	}
}

type testDescription struct {
	Name   string
	Args   []string
	Bundle string
	Error  bool
	Format string
}

type testCase struct {
	testDescription
	Expected []byte
}

func TestGitFrameAllScenarios(t *testing.T) {
	testdataDir := filepath.Join(repoRoot, "tests")
	bundlesDir := filepath.Join(testdataDir, "bundles")
	testsDir := filepath.Join(testdataDir, "testdata")

	testDirs, err := listTestDirs(testsDir)
	if err != nil {
		t.Fatalf("list test dirs: %v", err)
	}

	for _, dirName := range testDirs {
		tcPath := filepath.Join(testsDir, dirName)
		tc, err := readTestCase(tcPath)
		if err != nil {
			t.Fatalf("read test case %s: %v", dirName, err)
		}

		t.Run(dirName+"/"+tc.Name, func(t *testing.T) {
			tmpRoot := t.TempDir()
			repoPath := filepath.Join(tmpRoot, "repo")
			bundlePath := filepath.Join(bundlesDir, tc.Bundle)

			if err := cloneBundle(bundlePath, repoPath); err != nil {
				t.Fatalf("clone bundle: %v", err)
			}

			headBefore, err := getHeadRef(repoPath)
			if err != nil {
				t.Fatalf("get head before: %v", err)
			}

			args := append([]string{"--repository", repoPath}, tc.Args...)
			cmd := exec.Command(binaryPath, args...)
			cmd.Dir = repoRoot

			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			stdout, err := cmd.Output()

			if tc.Error {
				if err == nil {
					t.Fatalf("expected command to fail, but it succeeded")
				}
			} else {
				if err != nil {
					t.Fatalf("command failed: %v\nstderr:\n%s", err, stderr.String())
				}
				if err := compareResults(tc.Expected, stdout, tc.Format); err != nil {
					t.Fatalf("output mismatch: %v", err)
				}
			}

			headAfter, err := getHeadRef(repoPath)
			if err != nil {
				t.Fatalf("get head after: %v", err)
			}
			if headBefore != headAfter {
				t.Fatalf("HEAD changed during test: before=%q after=%q", headBefore, headAfter)
			}
		})
	}
}

func listTestDirs(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}

	sort.Slice(names, func(i, j int) bool {
		left, _ := strconv.Atoi(names[i])
		right, _ := strconv.Atoi(names[j])
		return left < right
	})

	return names, nil
}

func readTestCase(path string) (*testCase, error) {
	desc, err := readDescription(filepath.Join(path, "description.yaml"))
	if err != nil {
		return nil, err
	}

	expected, err := os.ReadFile(filepath.Join(path, "expected.out"))
	if err != nil {
		return nil, err
	}

	return &testCase{testDescription: *desc, Expected: expected}, nil
}

func readDescription(path string) (*testDescription, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	desc := &testDescription{}
	for _, raw := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		sep := strings.Index(line, ":")
		if sep <= 0 {
			continue
		}

		key := strings.TrimSpace(line[:sep])
		value := strings.TrimSpace(line[sep+1:])

		switch key {
		case "name":
			desc.Name = trimYAMLScalar(value)
		case "args":
			desc.Args, err = parseArgsList(value)
			if err != nil {
				return nil, fmt.Errorf("parse args in %s: %w", path, err)
			}
		case "bundle":
			desc.Bundle = trimYAMLScalar(value)
		case "error":
			desc.Error, err = strconv.ParseBool(strings.ToLower(value))
			if err != nil {
				return nil, fmt.Errorf("parse error flag in %s: %w", path, err)
			}
		case "format":
			desc.Format = trimYAMLScalar(value)
		}
	}

	if desc.Name == "" {
		return nil, fmt.Errorf("description has empty name: %s", path)
	}
	if desc.Bundle == "" {
		return nil, fmt.Errorf("description has empty bundle: %s", path)
	}

	return desc, nil
}

func trimYAMLScalar(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		if (value[0] == '\'' && value[len(value)-1] == '\'') || (value[0] == '"' && value[len(value)-1] == '"') {
			return value[1 : len(value)-1]
		}
	}
	return value
}

func parseArgsList(value string) ([]string, error) {
	value = strings.TrimSpace(value)
	if value == "[]" {
		return []string{}, nil
	}
	if !strings.HasPrefix(value, "[") || !strings.HasSuffix(value, "]") {
		return nil, fmt.Errorf("args must be an inline list, got %q", value)
	}

	inner := strings.TrimSpace(value[1 : len(value)-1])
	if inner == "" {
		return []string{}, nil
	}

	parts := make([]string, 0, 8)
	var token strings.Builder
	inSingle := false
	inDouble := false

	for _, ch := range inner {
		switch ch {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
				continue
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
				continue
			}
		case ',':
			if !inSingle && !inDouble {
				part := strings.TrimSpace(token.String())
				if part != "" {
					parts = append(parts, part)
				}
				token.Reset()
				continue
			}
		}
		token.WriteRune(ch)
	}

	if inSingle || inDouble {
		return nil, fmt.Errorf("unterminated quote in args list %q", value)
	}

	part := strings.TrimSpace(token.String())
	if part != "" {
		parts = append(parts, part)
	}

	return parts, nil
}

func cloneBundle(src, dst string) error {
	cmd := exec.Command("git", "clone", src, dst)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w\n%s", err, string(output))
	}
	return nil
}

func getHeadRef(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func compareResults(expected, actual []byte, format string) error {
	switch format {
	case "json":
		return compareJSON(expected, actual)
	case "json-lines":
		return compareJSONLines(expected, actual)
	default:
		if string(expected) != string(actual) {
			return fmt.Errorf("expected %q, got %q", string(expected), string(actual))
		}
		return nil
	}
}

func compareJSON(expected, actual []byte) error {
	trimmedExpected := bytes.TrimSpace(expected)
	trimmedActual := bytes.TrimSpace(actual)
	if len(trimmedExpected) == 0 {
		if len(trimmedActual) != 0 {
			return fmt.Errorf("expected empty json output, got %q", string(actual))
		}
		return nil
	}

	var exp any
	if err := json.Unmarshal(trimmedExpected, &exp); err != nil {
		return fmt.Errorf("bad expected json: %w", err)
	}

	var act any
	if err := json.Unmarshal(trimmedActual, &act); err != nil {
		return fmt.Errorf("bad actual json: %w", err)
	}

	if !reflect.DeepEqual(exp, act) {
		return fmt.Errorf("json mismatch: expected %s, got %s", string(trimmedExpected), string(trimmedActual))
	}
	return nil
}

func compareJSONLines(expected, actual []byte) error {
	expLines := parseJSONLines(expected)
	actLines := parseJSONLines(actual)
	if len(expLines) != len(actLines) {
		return fmt.Errorf("json-lines line count mismatch: expected %d, got %d", len(expLines), len(actLines))
	}

	for i := range expLines {
		if err := compareJSON(expLines[i], actLines[i]); err != nil {
			return fmt.Errorf("line %d: %w", i+1, err)
		}
	}

	return nil
}

func parseJSONLines(data []byte) [][]byte {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil
	}
	return bytes.Split(trimmed, []byte("\n"))
}
