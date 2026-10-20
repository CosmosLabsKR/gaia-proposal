package proposal

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	govv1 "cosmossdk.io/api/cosmos/gov/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

var errInvalidLineEnding = errors.New("line-ending must be one of: crlf, lf, keep")

type options struct {
	lineEnding       string
	keepFinalNewline bool
	valueOnly        bool
}

func runSummaryCommand(args []string, out, errOut io.Writer) error {
	var opts options
	fs := flag.NewFlagSet("summary", flag.ContinueOnError)
	fs.SetOutput(errOut)
	fs.StringVar(&opts.lineEnding, "line-ending", "crlf", "output line ending: crlf, lf, or keep")
	fs.BoolVar(&opts.keepFinalNewline, "keep-final-newline", false, "keep the editor-added final newline")
	fs.BoolVar(&opts.valueOnly, "value-only", false, "print only the JSON string literal")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: gaia-proposal summary [options] <markdown-file>")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return errors.New("exactly one markdown file is required")
	}
	return runSummary(out, fs.Arg(0), opts)
}

func runSummary(out io.Writer, path string, opts options) error {
	input, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %q: %w", path, err)
	}

	summary, err := normalizeMarkdown(string(input), opts.lineEnding, opts.keepFinalNewline)
	if err != nil {
		return err
	}

	var encoded []byte
	if opts.valueOnly {
		encoded, err = json.Marshal(summary)
	} else {
		// Proposal is the generated Cosmos SDK cosmos.gov.v1 protobuf type.
		// With only Summary set, protojson emits exactly {"summary":"..."}.
		proposal := &govv1.Proposal{Summary: summary}
		encoded, err = (protojson.MarshalOptions{UseProtoNames: true}).Marshal(proposal)
	}
	if err != nil {
		return fmt.Errorf("encode JSON: %w", err)
	}

	if _, err := fmt.Fprintln(out, string(encoded)); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}

func normalizeMarkdown(input, lineEnding string, keepFinalNewline bool) (string, error) {
	if lineEnding == "keep" {
		if !keepFinalNewline {
			input = trimOneFinalNewline(input)
		}
		return input, nil
	}
	if lineEnding != "crlf" && lineEnding != "lf" {
		return "", errInvalidLineEnding
	}

	// Normalize first so LF, CRLF, and old-style CR input behave identically.
	normalized := strings.ReplaceAll(input, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	if !keepFinalNewline {
		normalized = strings.TrimSuffix(normalized, "\n")
	}
	if lineEnding == "crlf" {
		normalized = strings.ReplaceAll(normalized, "\n", "\r\n")
	}
	return normalized, nil
}

func trimOneFinalNewline(input string) string {
	if strings.HasSuffix(input, "\r\n") {
		return strings.TrimSuffix(input, "\r\n")
	}
	if strings.HasSuffix(input, "\n") || strings.HasSuffix(input, "\r") {
		return input[:len(input)-1]
	}
	return input
}
