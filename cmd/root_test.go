package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantErr   bool
		contains  []string
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
			name:     "Invalid command",
			args:     []string{"invalid-command"},
			wantErr:  true,
			contains: []string{"unknown command"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset RootCmd
			RootCmd.ResetFlags()
			
			// Create buffers to capture output
			outBuf := new(bytes.Buffer)
			errBuf := new(bytes.Buffer)
			
			// Set output and error streams
			RootCmd.SetOut(outBuf)
			RootCmd.SetErr(errBuf)
			
			// Set args
			RootCmd.SetArgs(tt.args)
			
			// Execute command
			err := RootCmd.Execute()
			
			// Check if error was expected
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			// Get combined output
			output := outBuf.String() + errBuf.String()
			
			// Check if output contains expected strings
			for _, s := range tt.contains {
				if !strings.Contains(output, s) {
					t.Errorf("Output doesn't contain expected string %q", s)
				}
			}
		})
	}
}

func TestExecute(t *testing.T) {
	// This is more of a smoke test since we can't easily test 
	// the exit code without modifying the Execute function
	
	// Save original stdout/stderr
	oldStdout := RootCmd.OutOrStdout()
	oldStderr := RootCmd.ErrOrStderr()
	
	// Restore after test
	defer func() {
		RootCmd.SetOut(oldStdout)
		RootCmd.SetErr(oldStderr)
	}()
	
	// Create buffer to capture output
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	
	// Set a valid command that will succeed
	RootCmd.SetArgs([]string{"--help"})
	
	// Execute function should not panic
	Execute()
} 