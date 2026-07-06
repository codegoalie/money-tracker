// Package uat holds the Green/Green Auto-UAT harness. See doc.go for the
// package-level overview.
//
// This file builds the moneytracker binary once for the whole test run
// (TestMain) and provides runSUT, the helper every black-box test uses to
// exec that binary against a twin.
package uat

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// sutPath is the path to the moneytracker binary built once by TestMain.
// It is populated before any test runs and is safe to read (never written)
// from concurrent tests thereafter.
var sutPath string

// sutTempDir is the temp directory holding the built binary, removed after
// the test run completes.
var sutTempDir string

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

// runTests builds the SUT, runs the test suite, cleans up, and returns the
// process exit code. It exists as a separate function from TestMain because
// os.Exit does not run deferred cleanup, so cleanup must happen explicitly
// before we return the exit code to TestMain.
func runTests(m *testing.M) int {
	repoRoot, err := repoRootDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "uat: failed to locate repo root: %v\n", err)
		return 1
	}

	tempDir, err := os.MkdirTemp("", "moneytracker-uat-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "uat: failed to create temp dir: %v\n", err)
		return 1
	}
	sutTempDir = tempDir
	defer func() { _ = os.RemoveAll(sutTempDir) }()

	binPath := filepath.Join(sutTempDir, "moneytracker")

	buildCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(buildCtx, "go", "build", "-o", binPath, "./cmd/moneytracker")
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "uat: failed to build SUT (moneytracker):\n%s\n", out)
		fmt.Fprintf(os.Stderr, "uat: build error: %v\n", err)
		return 1
	}

	sutPath = binPath

	return m.Run()
}

// repoRootDir resolves the repo root (one directory up from the uat/ module)
// robustly, via runtime.Caller, since go test sets the test binary's working
// directory to the package directory regardless of where `go test` was
// invoked from.
func repoRootDir() (string, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("runtime.Caller failed to resolve this source file's path")
	}
	uatDir := filepath.Dir(thisFile)
	repoRoot := filepath.Dir(uatDir)
	return repoRoot, nil
}

// runSUT execs the built moneytracker binary with the given args, pointing
// it at a twin via LUNCHMONEY_API_URL/LUNCHMONEY_TOKEN environment
// variables. It returns the process exit code and captured stdout/stderr.
func runSUT(t *testing.T, twinURL, token string, args ...string) (exitCode int, stdout, stderr string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, sutPath, args...)
	cmd.Env = append(os.Environ(),
		"LUNCHMONEY_API_URL="+twinURL,
		"LUNCHMONEY_TOKEN="+token,
	)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if err == nil {
		return 0, outBuf.String(), errBuf.String()
	}

	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
		return exitErr.ExitCode(), outBuf.String(), errBuf.String()
	}

	t.Fatalf("runSUT: command failed in a way that isn't a plain exit error: %v", err)
	return -1, outBuf.String(), errBuf.String()
}
