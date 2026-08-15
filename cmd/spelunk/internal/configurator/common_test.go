package configurator_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	cliBinaryPath string
	buildOnce     sync.Once
	buildErr      error
)

func buildCLI(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		tmpDir, err := os.MkdirTemp("", "spelunk-e2e-*")
		if err != nil {
			buildErr = err
			return
		}
		cliBinaryPath = filepath.Join(tmpDir, "spelunk")
		cmd := exec.Command("go", "build", "-o", cliBinaryPath, "../../main.go")
		cmd.Env = os.Environ()
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			buildErr = err
		}
	})
	require.NoError(t, buildErr)
	return cliBinaryPath
}

func cleanEnv(t *testing.T, extraEnv ...string) []string {
	t.Helper()
	baseKeys := map[string]bool{
		"PATH":       true,
		"SYSTEMROOT": true,
		"TEMP":       true,
		"TMP":        true,
		"TMPDIR":     true,
	}
	var env []string
	for _, e := range os.Environ() {
		for k := range baseKeys {
			if len(e) > len(k)+1 && e[:len(k)+1] == k+"=" {
				env = append(env, e)
				break
			}
		}
	}
	env = append(env, "HOME="+t.TempDir())
	env = append(env, extraEnv...)
	return env
}

type execResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

func runCLI(ctx context.Context, bin string, env []string, args ...string) execResult {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Env = env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return execResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		Err:      err,
	}
}
