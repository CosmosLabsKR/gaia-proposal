package main

import (
	"fmt"
	"os"

	"gaia-proposal/internal/proposal"
)

func main() {
	if err := proposal.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "gaia-proposal: %v\n", err)
		os.Exit(1)
	}
}
