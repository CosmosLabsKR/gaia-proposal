// Package templates provides proposal templates embedded in the binary.
package templates

import _ "embed"

//go:embed base_proposal.json
var baseProposalJSON string

// BaseProposalJSON returns a copy of the default proposal template.
func BaseProposalJSON() []byte {
	return []byte(baseProposalJSON)
}
