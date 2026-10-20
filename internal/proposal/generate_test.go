package proposal

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateDirectoryAndCheck(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, manifestFilename), `{
  "title": "Gaia v28.0.0 Upgrade",
  "deposit": "500000000uatom"
}`)
	writeTestFile(t, filepath.Join(dir, planFilename), `{
  "name": "v28.0.0",
  "height": "32785900",
  "info": {"binaries":{"linux/amd64":"https://example.com/gaiad"}}
}`)
	writeTestFile(t, filepath.Join(dir, summaryFilename), "# Upgrade\n\nDetails.\n")

	var out bytes.Buffer
	if err := generateDirectory(dir, false, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "generated") {
		t.Fatalf("unexpected output %q", out.String())
	}
	if err := generateDirectory(dir, true, ioDiscard{}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, outputFilename))
	if err != nil {
		t.Fatal(err)
	}
	var proposal proposalDocument
	if err := json.Unmarshal(data, &proposal); err != nil {
		t.Fatal(err)
	}
	if got, want := proposal.Messages[0].Plan.Name, "v28.0.0"; got != want {
		t.Fatalf("plan name got %q, want %q", got, want)
	}
	if got, want := proposal.Deposit, "500000000uatom"; got != want {
		t.Fatalf("deposit got %q, want %q", got, want)
	}
}

func TestGenerateCheckDetectsDrift(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, manifestFilename), `{"title":"Upgrade"}`)
	writeTestFile(t, filepath.Join(dir, planFilename), `{"name":"v1","height":"10","info":{}}`)
	writeTestFile(t, filepath.Join(dir, summaryFilename), "Summary\n")
	writeTestFile(t, filepath.Join(dir, outputFilename), "{}\n")

	err := generateDirectory(dir, true, ioDiscard{})
	if err == nil || !strings.Contains(err.Error(), "out of date") {
		t.Fatalf("expected out-of-date error, got %v", err)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
