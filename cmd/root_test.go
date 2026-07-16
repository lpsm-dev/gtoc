package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// resetRootCmd restores RootCmd to the state it has right after this
// package's init() runs: all persistent flags registered and every
// subcommand attached. Every test in this package calls it before running a
// command so that ResetCommands()/ResetFlags() calls made by one test never
// leak into another, regardless of which file or order the tests execute in.
func resetRootCmd() (outBuf, errBuf *bytes.Buffer) {
	RootCmd.ResetFlags()
	RootCmd.ResetCommands()

	RootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "warn", "Set log level (debug, info, warn, error, fatal)")
	RootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "Log format (text, json)")
	RootCmd.PersistentFlags().BoolVar(&logNoColors, "log-no-colors", false, "Disable colors in logs")

	RootCmd.AddCommand(generateCmd)
	RootCmd.AddCommand(analyzeCmd)
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(upgradeCmd)

	outBuf = new(bytes.Buffer)
	errBuf = new(bytes.Buffer)
	RootCmd.SetOut(outBuf)
	RootCmd.SetErr(errBuf)

	return outBuf, errBuf
}

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		contains    []string
		errContains []string
	}{
		{
			name:     "Root command with no arguments",
			args:     []string{},
			wantErr:  false,
			contains: []string{"gtoc is a CLI tool", "Usage:", "Available Commands:"},
		},
		{
			name:     "Help flag",
			args:     []string{"--help"},
			wantErr:  false,
			contains: []string{"gtoc is a CLI tool", "Usage:", "Available Commands:", "Flags:"},
		},
		{
			name:        "Invalid command",
			args:        []string{"invalid-command"},
			wantErr:     true,
			errContains: []string{"unknown command"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outBuf, errBuf := resetRootCmd()
			RootCmd.SetArgs(tt.args)

			err := RootCmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			output := outBuf.String() + errBuf.String()
			for _, s := range tt.contains {
				if !strings.Contains(output, s) {
					t.Errorf("output doesn't contain expected string %q", s)
				}
			}

			// With SilenceErrors set, cobra no longer writes the error text
			// to the output buffer, so unknown-command style assertions
			// must be checked against the returned error instead.
			for _, s := range tt.errContains {
				if err == nil || !strings.Contains(err.Error(), s) {
					t.Errorf("error doesn't contain expected string %q, got %v", s, err)
				}
			}
		})
	}
}

func TestExecute(t *testing.T) {
	// This is more of a smoke test since we can't easily test the exit code
	// without modifying the Execute function.
	oldStdout := RootCmd.OutOrStdout()
	oldStderr := RootCmd.ErrOrStderr()
	defer func() {
		RootCmd.SetOut(oldStdout)
		RootCmd.SetErr(oldStderr)
	}()

	resetRootCmd()

	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)

	// Set a valid command that will succeed.
	RootCmd.SetArgs([]string{"--help"})

	// Execute function should not panic.
	Execute()
}
