# Overview

This proposal schedules a coordinated upgrade of the Cosmos Hub to Gaia `v29.0.0`.

**If the proposal passes, validators will be required to update the Cosmos Hub binary at the specified upgrade height.**

**This upgrade is state-breaking.**

## Highlights

- Remove the `x/params` module.
- Update the Cosmos SDK from `v0.53.4` to `v0.53.8`.
- Update Go from `1.25` to `1.26`.

### Breaking Changes

- The legacy `ParameterChangeProposal` governance route and the `cosmos.params.v1beta1.Query` service are no longer available following the removal of `x/params`.

### Known Issues

- In Cosmos SDK `v0.53.8`, gas simulation with `--gas auto` is broken for multisig senders. Multisig transactions must specify a fixed `--gas` value until the issue is fixed upstream.

## Release Details

- [Gaia v29.0.0 release](https://github.com/cosmos/gaia/releases#release-v29.0.0)
- [Changelog](https://github.com/cosmos/gaia/blob/main/CHANGELOG.md)

# Upgrade Instructions and Verification

Please check the announcements in the established communication channels for release details and upgrade instructions.

Public verification keys and instructions for verifying Gaia release signatures are available in the [official Gaia release signing directory](https://github.com/cosmos/security/tree/main/release/gaia/signing).

# Voting

This expedited proposal has an initial voting period of **THREE DAYS**.

The following items summarize the voting options and what they mean for this proposal:

YES - You agree that the Cosmos Hub should be updated with this release.

NO - You disagree that the Cosmos Hub should be updated with this release.

NO WITH VETO - A `NoWithVeto` vote indicates that the proposal either (1) is deemed to be spam, i.e., irrelevant to Cosmos Hub, (2) disproportionately infringes on minority interests, or (3) violates or encourages violation of the rules of engagement as currently set out by Cosmos Hub governance. Veto thresholds and any deposit burning are governed by the chain's active governance parameters.

ABSTAIN - You wish to contribute to the quorum but formally decline to vote either for or against the proposal.
