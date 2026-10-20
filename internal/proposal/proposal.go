package proposal

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"gaia-proposal/templates"
)

const defaultBasePath = "base_proposal.json"

type proposalDocument struct {
	Messages  []softwareUpgradeMessage `json:"messages"`
	Metadata  string                   `json:"metadata"`
	Deposit   string                   `json:"deposit"`
	Title     string                   `json:"title"`
	Summary   string                   `json:"summary"`
	Expedited bool                     `json:"expedited"`
}

type softwareUpgradeMessage struct {
	Type      string      `json:"@type"`
	Authority string      `json:"authority"`
	Plan      upgradePlan `json:"plan"`
}

type upgradePlan struct {
	Name                string          `json:"name"`
	Time                string          `json:"time"`
	Height              string          `json:"height"`
	Info                string          `json:"info"`
	UpgradedClientState json.RawMessage `json:"upgraded_client_state"`
}

type proposalOptions struct {
	basePath       string
	outputPath     string
	planFile       string
	planName       string
	height         string
	planInfo       string
	planInfoFile   string
	authority      string
	metadata       string
	deposit        string
	title          string
	summaryFile    string
	expedited      bool
	force          bool
	nonInteractive bool
	set            map[string]bool
}

func runProposalCommand(args []string, in io.Reader, out, errOut io.Writer) error {
	fs := flag.NewFlagSet("proposal", flag.ContinueOnError)
	fs.SetOutput(errOut)
	var opts proposalOptions
	fs.StringVar(&opts.basePath, "base", defaultBasePath, "base draft proposal JSON")
	fs.StringVar(&opts.outputPath, "output", "proposal.json", "output proposal JSON")
	fs.StringVar(&opts.planFile, "plan-file", "", "JSON file containing plan name, height, and info")
	fs.StringVar(&opts.planName, "plan-name", "", "software upgrade plan name")
	fs.StringVar(&opts.height, "height", "", "software upgrade height")
	fs.StringVar(&opts.planInfo, "plan-info", "", "literal plan info string")
	fs.StringVar(&opts.planInfoFile, "plan-info-file", "", "JSON file stored as an escaped plan info string")
	fs.StringVar(&opts.authority, "authority", "", "upgrade message authority")
	fs.StringVar(&opts.metadata, "metadata", "", "proposal metadata")
	fs.StringVar(&opts.deposit, "deposit", "", "initial deposit, for example 1uatom")
	fs.StringVar(&opts.title, "title", "", "proposal title")
	fs.StringVar(&opts.summaryFile, "summary-file", "", "proposal summary markdown file")
	fs.BoolVar(&opts.expedited, "expedited", false, "create an expedited proposal")
	fs.BoolVar(&opts.force, "force", false, "overwrite the output file")
	fs.BoolVar(&opts.nonInteractive, "non-interactive", false, "do not prompt; use flags and base values")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: gaia-proposal proposal [options]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	opts.set = make(map[string]bool)
	fs.Visit(func(f *flag.Flag) { opts.set[f.Name] = true })

	proposal, err := loadBaseProposal(opts.basePath)
	if err != nil {
		return err
	}
	if err := proposal.validateShape(); err != nil {
		return fmt.Errorf("invalid base proposal: %w", err)
	}
	if err := applyProposalOptions(&proposal, opts); err != nil {
		return err
	}

	if !opts.nonInteractive && isTerminalReader(in) {
		if err := promptProposal(in, out, &proposal, &opts); err != nil {
			return err
		}
	}

	warnings := proposal.warnings()
	for _, warning := range warnings {
		fmt.Fprintf(errOut, "warning: %s\n", warning)
	}
	if err := writeProposal(opts.outputPath, proposal, opts.force); err != nil {
		return err
	}
	fmt.Fprintf(out, "wrote %s\n", opts.outputPath)
	return nil
}

func loadBaseProposal(path string) (proposalDocument, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && path == defaultBasePath {
			data = templates.BaseProposalJSON()
		} else {
			return proposalDocument{}, fmt.Errorf("read base %q: %w", path, err)
		}
	}
	var proposal proposalDocument
	if err := json.Unmarshal(data, &proposal); err != nil {
		return proposalDocument{}, fmt.Errorf("decode base %q: %w", path, err)
	}
	return proposal, nil
}

func (p proposalDocument) validateShape() error {
	if len(p.Messages) != 1 {
		return fmt.Errorf("expected exactly one message, got %d", len(p.Messages))
	}
	if p.Messages[0].Type != "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade" {
		return fmt.Errorf("unsupported message type %q", p.Messages[0].Type)
	}
	return nil
}

func applyProposalOptions(p *proposalDocument, opts proposalOptions) error {
	msg := &p.Messages[0]
	if opts.set["plan-file"] {
		if err := applyPlanFile(&msg.Plan, opts.planFile); err != nil {
			return err
		}
	}
	if opts.set["plan-name"] {
		msg.Plan.Name = opts.planName
	}
	if opts.set["height"] {
		if err := validateHeight(opts.height); err != nil {
			return err
		}
		msg.Plan.Height = opts.height
	}
	if opts.set["authority"] {
		msg.Authority = opts.authority
	}
	if opts.set["metadata"] {
		p.Metadata = opts.metadata
	}
	if opts.set["deposit"] {
		p.Deposit = opts.deposit
	}
	if opts.set["title"] {
		p.Title = opts.title
	}
	if opts.set["expedited"] {
		p.Expedited = opts.expedited
	}
	if opts.set["plan-info"] && opts.set["plan-info-file"] {
		return errors.New("use only one of --plan-info and --plan-info-file")
	}
	if opts.set["plan-info"] {
		msg.Plan.Info = opts.planInfo
	}
	if opts.set["plan-info-file"] {
		info, err := readCompactJSON(opts.planInfoFile)
		if err != nil {
			return err
		}
		msg.Plan.Info = info
	}
	if opts.set["summary-file"] {
		summary, err := readSummary(opts.summaryFile)
		if err != nil {
			return err
		}
		p.Summary = summary
	}
	return nil
}

func promptProposal(in io.Reader, out io.Writer, p *proposalDocument, opts *proposalOptions) error {
	reader := bufio.NewReader(in)
	var err error
	if opts.outputPath, err = prompt(reader, out, "Output file", opts.outputPath); err != nil {
		return err
	}
	msg := &p.Messages[0]
	if msg.Plan.Name, err = prompt(reader, out, "Plan name", msg.Plan.Name); err != nil {
		return err
	}
	if msg.Plan.Height, err = prompt(reader, out, "Upgrade height", msg.Plan.Height); err != nil {
		return err
	}
	if err := validateHeight(msg.Plan.Height); err != nil {
		return err
	}
	planPath, err := prompt(reader, out, "Plan JSON file", opts.planFile)
	if err != nil {
		return err
	}
	if planPath != "" && planPath != opts.planFile {
		if err := applyPlanFile(&msg.Plan, planPath); err != nil {
			return err
		}
		opts.planFile = planPath
	}
	if p.Metadata, err = prompt(reader, out, "Metadata", p.Metadata); err != nil {
		return err
	}
	if p.Deposit, err = prompt(reader, out, "Deposit", p.Deposit); err != nil {
		return err
	}
	if p.Title, err = prompt(reader, out, "Title", p.Title); err != nil {
		return err
	}
	summaryPath, err := prompt(reader, out, "Summary markdown file", opts.summaryFile)
	if err != nil {
		return err
	}
	if summaryPath != "" && summaryPath != opts.summaryFile {
		p.Summary, err = readSummary(summaryPath)
		if err != nil {
			return err
		}
		opts.summaryFile = summaryPath
	}
	expedited, err := prompt(reader, out, "Expedited (true/false)", strconv.FormatBool(p.Expedited))
	if err != nil {
		return err
	}
	p.Expedited, err = strconv.ParseBool(expedited)
	if err != nil {
		return fmt.Errorf("invalid expedited value %q", expedited)
	}
	return nil
}

func prompt(reader *bufio.Reader, out io.Writer, label, current string) (string, error) {
	fmt.Fprintf(out, "%s [%s]: ", label, current)
	value, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return current, nil
	}
	return value, nil
}

func isTerminalReader(in io.Reader) bool {
	file, ok := in.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func readCompactJSON(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read plan info %q: %w", path, err)
	}
	if !json.Valid(data) {
		return "", fmt.Errorf("plan info %q is not valid JSON", path)
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, data); err != nil {
		return "", fmt.Errorf("compact plan info %q: %w", path, err)
	}
	return compact.String(), nil
}

type planDocument struct {
	Name                string          `json:"name"`
	Time                string          `json:"time"`
	Height              string          `json:"height"`
	Info                json.RawMessage `json:"info"`
	UpgradedClientState json.RawMessage `json:"upgraded_client_state"`
}

func applyPlanFile(plan *upgradePlan, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read plan %q: %w", path, err)
	}
	var input planDocument
	if err := json.Unmarshal(data, &input); err != nil {
		return fmt.Errorf("decode plan %q: %w", path, err)
	}
	if input.Name == "" {
		return fmt.Errorf("plan %q: name cannot be empty", path)
	}
	if err := validateHeight(input.Height); err != nil {
		return fmt.Errorf("plan %q: %w", path, err)
	}
	if input.Height == "0" {
		return fmt.Errorf("plan %q: height must be greater than 0", path)
	}
	if len(input.Info) == 0 || bytes.Equal(input.Info, []byte("null")) {
		return fmt.Errorf("plan %q: info is required", path)
	}

	info, err := planInfoString(input.Info)
	if err != nil {
		return fmt.Errorf("plan %q: %w", path, err)
	}
	plan.Name = input.Name
	plan.Height = input.Height
	plan.Info = info
	if input.Time != "" {
		plan.Time = input.Time
	}
	if len(input.UpgradedClientState) != 0 {
		plan.UpgradedClientState = input.UpgradedClientState
	}
	return nil
}

func planInfoString(raw json.RawMessage) (string, error) {
	if len(raw) > 0 && raw[0] == '"' {
		var info string
		if err := json.Unmarshal(raw, &info); err != nil {
			return "", fmt.Errorf("decode info string: %w", err)
		}
		return info, nil
	}
	if !json.Valid(raw) {
		return "", errors.New("info is not valid JSON")
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, raw); err != nil {
		return "", fmt.Errorf("compact info: %w", err)
	}
	return compact.String(), nil
}

func readSummary(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read summary %q: %w", path, err)
	}
	return normalizeMarkdown(string(data), "crlf", false)
}

func validateHeight(height string) error {
	if height == "" {
		return errors.New("height cannot be empty")
	}
	if _, err := strconv.ParseUint(height, 10, 64); err != nil {
		return fmt.Errorf("height must be an unsigned integer: %w", err)
	}
	return nil
}

func (p proposalDocument) warnings() []string {
	var warnings []string
	if p.Messages[0].Plan.Name == "" {
		warnings = append(warnings, "plan name is empty")
	}
	if p.Messages[0].Plan.Height == "0" {
		warnings = append(warnings, "upgrade height is 0")
	}
	if p.Title == "" || p.Title == "title" {
		warnings = append(warnings, "title still has a draft value")
	}
	if p.Summary == "" || p.Summary == "summary" {
		warnings = append(warnings, "summary still has a draft value")
	}
	return warnings
}

func writeProposal(path string, proposal proposalDocument, force bool) error {
	data, err := json.MarshalIndent(proposal, "", "  ")
	if err != nil {
		return fmt.Errorf("encode proposal: %w", err)
	}
	data = append(data, '\n')
	flags := os.O_WRONLY | os.O_CREATE
	if force {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_EXCL
	}
	file, err := os.OpenFile(path, flags, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("output %q already exists; use --force to overwrite", path)
		}
		return fmt.Errorf("create output %q: %w", path, err)
	}
	defer file.Close()
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("write output %q: %w", path, err)
	}
	return nil
}
