package proposal

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeMarkdown(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		lineEnding string
		keepFinal  bool
		want       string
	}{
		{"LF to CRLF", "# Title\n\nA \"quote\".\n", "crlf", false, "# Title\r\n\r\nA \"quote\"."},
		{"mixed to LF", "a\r\nb\rc\n", "lf", false, "a\nb\nc"},
		{"keep final newline", "a\n", "crlf", true, "a\r\n"},
		{"preserve input", "a\r\nb\n", "keep", false, "a\r\nb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeMarkdown(tt.input, tt.lineEnding, tt.keepFinal)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRunOutputsCosmosGovSummaryJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "1052-prop.txt")
	if err := os.WriteFile(path, []byte("# Background\n\nA \"quote\".\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := runSummary(&out, path, options{lineEnding: "crlf"})
	if err != nil {
		t.Fatal(err)
	}
	want := "{\"summary\":\"# Background\\r\\n\\r\\nA \\\"quote\\\".\"}\n"
	if out.String() != want {
		t.Fatalf("got %q, want %q", out.String(), want)
	}
}

func TestRunValueOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "proposal.txt")
	if err := os.WriteFile(path, []byte("line 1\nline 2"), 0o600); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := runSummary(&out, path, options{lineEnding: "lf", valueOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := out.String(), "\"line 1\\nline 2\"\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
