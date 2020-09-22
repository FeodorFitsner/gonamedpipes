// +build windows

package main

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestCommandLoop(t *testing.T) {
	pc, _ := newPipeImpl("111")

	go func() {
		for {
			// command echo loop
			cmd := <-pc.commands
			t.Log(cmd)

			pc.writeResult(fmt.Sprintf("%s - OK", strings.TrimSpace(strings.Split(cmd, "\n")[0])))
		}
	}()

	out, err := runPowerShell("test-commands.ps1", 10*time.Second)

	if err != nil {
		t.Fatalf("cmd.Run() failed with %s\n", err)
	}
	t.Logf("combined out:\n%s\n", out)
}

func runPowerShell(script string, timeout time.Duration) (out string, err error) {
	// current tests directory
	_, filename, _, _ := runtime.Caller(0)
	psPath := filepath.Join(filepath.Dir(filename), "tests", script)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel() // The cancel should be deferred so resources are cleaned up

	cmd := exec.CommandContext(ctx, "powershell.exe", psPath)
	outBytes, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		err = ctx.Err()
		return
	}

	if err != nil {
		return
	}
	out = string(outBytes)
	return
}
