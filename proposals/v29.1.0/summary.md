# Overview

This proposal schedules a coordinated upgrade of the Cosmos Hub to Gaia `v29.1.0`.

**If the proposal passes, validators will be required to update the Cosmos Hub binary at the specified upgrade height.**

**This upgrade is state-breaking.**

## Highlights

- Restore decoding of historical `ParameterChangeProposal` content, fixing governance proposal list, status-filtered, single-proposal and tally queries, as well as state export.
- Update CosmWasm `wasmd` from `v0.60.9` to `v0.60.10`.
- Update CosmWasm `wasmvm` from `v2.3.5` to `v2.3.6`.

### State-Breaking Changes

- Legacy `ParameterChangeProposal` submissions now decode and reach ante processing before governance rejects their execution. As a result, fees and account sequences change for such submissions.

## Release Details

- [Gaia v29.1.0 release](https://github.com/cosmos/gaia/releases#release-v29.1.0)
- [Changelog](https://github.com/cosmos/gaia/blob/v29.1.0/CHANGELOG.md)
- [Upgrading guide](https://github.com/cosmos/gaia/blob/v29.1.0/UPGRADING.md)

# Upgrade Instructions and Verification

Please check the announcements in the established communication channels for release details and upgrade instructions.

Public verification keys and instructions for verifying Gaia release signatures are available in the [official Gaia release signing directory](https://github.com/cosmos/security/tree/main/release/gaia/signing).

# Voting

This proposal has a voting period of **SEVEN DAYS**.

The following items summarize the voting options and what they mean for this proposal:

YES - You agree that the Cosmos Hub should be updated with this release.

NO - You disagree that the Cosmos Hub should be updated with this release.

NO WITH VETO - A `NoWithVeto` vote indicates that the proposal either (1) is deemed to be spam, i.e., irrelevant to Cosmos Hub, (2) disproportionately infringes on minority interests, or (3) violates or encourages violation of the rules of engagement as currently set out by Cosmos Hub governance. Veto thresholds and any deposit burning are governed by the chain's active governance parameters.

ABSTAIN - You wish to contribute to the quorum but formally decline to vote either for or against the proposal.
