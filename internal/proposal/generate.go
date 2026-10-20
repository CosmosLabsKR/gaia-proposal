package proposal

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"gaia-proposal/templates"
)

const (
	manifestFilename = "manifest.json"
	planFilename     = "plan.json"
	summaryFilename  = "summary.md"
	outputFilename   = "proposal.json"
)

type manifestDocument struct {
	Title     string  `json:"title"`
	Deposit   *string `json:"deposit,omitempty"`
	Metadata  *string `json:"metadata,omitempty"`
	Expedited *bool   `json:"expedited,omitempty"`
	Authority *string `json:"authority,omitempty"`
	Base      string  `json:"base,omitempty"`
}

func runGenerateCommand(args []string, out, errOut io.Writer) error {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	fs.SetOutput(errOut)
	check := fs.Bool("check", false, "fail if proposal.json is missing or out of date")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: gaia-proposal generate [--check] <proposal-directory>")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return errors.New("exactly one proposal directory is required")
	}
	return generateDirectory(fs.Arg(0), *check, out)
}

func runGenerateAllCommand(args []string, out, errOut io.Writer) error {
	fs := flag.NewFlagSet("generate-all", flag.ContinueOnError)
	fs.SetOutput(errOut)
	check := fs.Bool("check", false, "fail if any proposal.json is missing or out of date")
	root := fs.String("root", "proposals", "root directory containing proposal directories")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: gaia-proposal generate-all [options]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}

	directories, err := proposalDirectories(*root)
	if err != nil {
		return err
	}
	for _, directory := range directories {
		if err := generateDirectory(directory, *check, out); err != nil {
			return err
		}
	}
	return nil
}

func proposalDirectories(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read proposals root %q: %w", root, err)
	}
	var directories []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		directory := filepath.Join(root, entry.Name())
		if _, err := os.Stat(filepath.Join(directory, manifestFilename)); err == nil {
			directories = append(directories, directory)
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("inspect %q: %w", directory, err)
		}
	}
	if len(directories) == 0 {
		return nil, fmt.Errorf("no %s files found under %q", manifestFilename, root)
	}
	sort.Strings(directories)
	return directories, nil
}

func generateDirectory(directory string, check bool, out io.Writer) error {
	manifest, err := readManifest(filepath.Join(directory, manifestFilename))
	if err != nil {
		return err
	}
	proposal, err := proposalFromManifest(directory, manifest)
	if err != nil {
		return err
	}
	data, err := marshalProposal(proposal)
	if err != nil {
		return err
	}
	outputPath := filepath.Join(directory, outputFilename)

	if check {
		existing, err := os.ReadFile(outputPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("%s is missing; run make generate", outputPath)
			}
			return fmt.Errorf("read generated proposal %q: %w", outputPath, err)
		}
		if !bytes.Equal(existing, data) {
			return fmt.Errorf("%s is out of date; run make generate", outputPath)
		}
		fmt.Fprintf(out, "checked %s\n", outputPath)
		return nil
	}

	changed, err := writeGeneratedFile(outputPath, data)
	if err != nil {
		return err
	}
	if changed {
		fmt.Fprintf(out, "generated %s\n", outputPath)
	} else {
		fmt.Fprintf(out, "unchanged %s\n", outputPath)
	}
	return nil
}

func readManifest(path string) (manifestDocument, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return manifestDocument{}, fmt.Errorf("read manifest %q: %w", path, err)
	}
	var manifest manifestDocument
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return manifestDocument{}, fmt.Errorf("decode manifest %q: %w", path, err)
	}
	if manifest.Title == "" {
		return manifestDocument{}, fmt.Errorf("manifest %q: title cannot be empty", path)
	}
	return manifest, nil
}

func proposalFromManifest(directory string, manifest manifestDocument) (proposalDocument, error) {
	var proposal proposalDocument
	var err error
	if manifest.Base == "" {
		if err := json.Unmarshal(templates.BaseProposalJSON(), &proposal); err != nil {
			return proposalDocument{}, fmt.Errorf("decode embedded base proposal: %w", err)
		}
	} else {
		basePath := manifest.Base
		if !filepath.IsAbs(basePath) {
			basePath = filepath.Join(directory, basePath)
		}
		proposal, err = loadBaseProposal(basePath)
		if err != nil {
			return proposalDocument{}, err
		}
	}
	if err := proposal.validateShape(); err != nil {
		return proposalDocument{}, fmt.Errorf("invalid base proposal: %w", err)
	}
	if err := applyPlanFile(&proposal.Messages[0].Plan, filepath.Join(directory, planFilename)); err != nil {
		return proposalDocument{}, err
	}
	proposal.Summary, err = readSummary(filepath.Join(directory, summaryFilename))
	if err != nil {
		return proposalDocument{}, err
	}
	proposal.Title = manifest.Title
	if manifest.Deposit != nil {
		proposal.Deposit = *manifest.Deposit
	}
	if manifest.Metadata != nil {
		proposal.Metadata = *manifest.Metadata
	}
	if manifest.Expedited != nil {
		proposal.Expedited = *manifest.Expedited
	}
	if manifest.Authority != nil {
		proposal.Messages[0].Authority = *manifest.Authority
	}
	if proposal.Deposit == "" {
		return proposalDocument{}, errors.New("deposit cannot be empty")
	}
	if proposal.Messages[0].Authority == "" {
		return proposalDocument{}, errors.New("authority cannot be empty")
	}
	if warnings := proposal.warnings(); len(warnings) > 0 {
		return proposalDocument{}, fmt.Errorf("proposal is incomplete: %v", warnings)
	}
	return proposal, nil
}

func marshalProposal(proposal proposalDocument) ([]byte, error) {
	data, err := json.MarshalIndent(proposal, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode proposal: %w", err)
	}
	return append(data, '\n'), nil
}

func writeGeneratedFile(path string, data []byte) (bool, error) {
	existing, err := os.ReadFile(path)
	if err == nil && bytes.Equal(existing, data) {
		return false, nil
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("read generated proposal %q: %w", path, err)
	}

	temp, err := os.CreateTemp(filepath.Dir(path), ".proposal-*.tmp")
	if err != nil {
		return false, fmt.Errorf("create temporary proposal: %w", err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err := temp.Chmod(0o644); err != nil {
		temp.Close()
		return false, fmt.Errorf("set proposal permissions: %w", err)
	}
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return false, fmt.Errorf("write temporary proposal: %w", err)
	}
	if err := temp.Close(); err != nil {
		return false, fmt.Errorf("close temporary proposal: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return false, fmt.Errorf("replace generated proposal %q: %w", path, err)
	}
	return true, nil
}
