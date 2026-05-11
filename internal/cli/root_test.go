package cli

import (
	"bytes"
	"strings"
	"testing"
)

// TestVersionCmd is a table-driven test — the idiomatic Go testing pattern.
// Each entry in the table is an independent scenario. This scales cleanly
// as you add cases without duplicating boilerplate.
func TestVersionCmd(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		wantContain string
	}{
		{
			name:        "dev build",
			version:     "dev",
			wantContain: "dev",
		},
		{
			name:        "tagged release",
			version:     "1.2.3",
			wantContain: "1.2.3",
		},
	}

	for _, tt := range tests {
		// t.Run creates a sub-test. You can run one by name:
		//   go test ./internal/cli/ -run TestVersionCmd/tagged_release
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout so we can assert on it without printing to terminal.
			var buf bytes.Buffer

			cmd := versionCmd(tt.version)
			cmd.SetOut(&buf)
			// SetArgs is essential — cobra reads os.Args by default,
			// which breaks in test environments.
			cmd.SetArgs([]string{})

			if err := cmd.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// We use Contains rather than exact match so the output format
			// can evolve (e.g. add a git SHA) without breaking the test.
			got := buf.String()
			if !strings.Contains(got, tt.wantContain) {
				t.Errorf("output %q does not contain %q", got, tt.wantContain)
			}
		})
	}
}
