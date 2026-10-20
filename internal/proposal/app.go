package proposal

import (
	"errors"
	"flag"
	"fmt"
	"io"
)

// Run dispatches a gaia-proposal CLI command.
func Run(args []string, in io.Reader, out, errOut io.Writer) error {
	if len(args) == 0 {
		printUsage(errOut)
		return errors.New("a command is required")
	}

	var err error
	switch args[0] {
	case "proposal":
		err = runProposalCommand(args[1:], in, out, errOut)
	case "generate":
		err = runGenerateCommand(args[1:], out, errOut)
	case "generate-all":
		err = runGenerateAllCommand(args[1:], out, errOut)
	case "summary":
		err = runSummaryCommand(args[1:], out, errOut)
	case "help", "-h", "--help":
		printUsage(out)
		return nil
	default:
		// Backward compatibility with the original summary-only CLI.
		err = runSummaryCommand(args, out, errOut)
	}
	if errors.Is(err, flag.ErrHelp) {
		return nil
	}
	return err
}

func printUsage(out io.Writer) {
	fmt.Fprintln(out, `Usage: gaia-proposal <command> [options]

Commands:
  proposal       Interactively build one proposal from explicit files
  generate       Generate one proposals/<version>/proposal.json
  generate-all   Generate or check every proposal directory
  summary        Convert a Markdown file to a JSON summary value`)
}
