package proposal

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProposalCommandBuildsSubmitProposalJSON(t *testing.T) {
	dir := t.TempDir()
	summaryPath := filepath.Join(dir, "summary.md")
	planPath := filepath.Join(dir, "plan.json")
	outputPath := filepath.Join(dir, "proposal.json")
	if err := os.WriteFile(summaryPath, []byte("# Background\n\nA \"quote\".\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(planPath, []byte("{\n  \"name\": \"v28.0.0\",\n  \"height\": \"32785900\",\n  \"info\": {\n    \"binaries\": {\n      \"linux/amd64\": \"https://example.com/gaiad\"\n    }\n  }\n}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	args := []string{
		"--non-interactive",
		"--output", outputPath,
		"--plan-file", planPath,
		"--title", "Gaia v28.0.0 Upgrade",
		"--summary-file", summaryPath,
	}
	if err := runProposalCommand(args, nil, ioDiscard{}, ioDiscard{}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	var proposal proposalDocument
	if err := json.Unmarshal(data, &proposal); err != nil {
		t.Fatal(err)
	}
	if got, want := proposal.Messages[0].Plan.Info, `{"binaries":{"linux/amd64":"https://example.com/gaiad"}}`; got != want {
		t.Fatalf("info got %q, want %q", got, want)
	}
	if got, want := proposal.Summary, "# Background\r\n\r\nA \"quote\"."; got != want {
		t.Fatalf("summary got %q, want %q", got, want)
	}
}

func TestPlanFileCanUseAlreadyEncodedInfoString(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plan.json")
	data := []byte(`{"name":"v28.0.0","height":"32785900","info":"{\"binaries\":{}}"}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	plan := upgradePlan{Time: "0001-01-01T00:00:00Z", UpgradedClientState: json.RawMessage("null")}
	if err := applyPlanFile(&plan, path); err != nil {
		t.Fatal(err)
	}
	if got, want := plan.Info, `{"binaries":{}}`; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestLoadBaseProposalFallsBackToEmbeddedDefault(t *testing.T) {
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldDir) })

	proposal, err := loadBaseProposal(defaultBasePath)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := proposal.Summary, "summary"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestPromptShowsFileOptionsAsDefaults(t *testing.T) {
	proposal, err := loadBaseProposal(defaultBasePath)
	if err != nil {
		t.Fatal(err)
	}
	opts := proposalOptions{
		outputPath:  "v28-proposal.json",
		planFile:    "plan.json",
		summaryFile: "1052-prop.txt",
	}
	// Accept every displayed default.
	in := strings.NewReader("\n\n\n\n\n\n\n\n\n")
	var out bytes.Buffer
	if err := promptProposal(in, &out, &proposal, &opts); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Plan JSON file [plan.json]:") {
		t.Fatalf("plan file default missing from prompt: %q", out.String())
	}
	if !strings.Contains(out.String(), "Summary markdown file [1052-prop.txt]:") {
		t.Fatalf("summary file default missing from prompt: %q", out.String())
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }
