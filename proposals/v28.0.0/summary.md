# Background

The Gaia `v28.0.0` upgrade removes Interchain Security (ICS) from the Cosmos Hub.

The following will happen as part of this upgrade:

1. The `max_validators` parameter in the staking module will be updated to match the `max_provider_consensus_validators` parameter in the provider module.
2. All bonded validators that are ranked below the new `max_validator` threshold at the time of the upgrade will be set to unbonding. They will not be jailed or slashed.
3. All tokens in the consumer rewards pool account will be transferred to the community pool, and the consumer rewards pool will be blocked from receiving any additional tokens.
4. All pending VSCPackets will be deleted.
5. All `provider` port IBC channels will be closed.

The upgrade handler will transfer the full remaining balance of the legacy `ConsumerRewardsPool` module account to the community pool. These legacy consumer-chain rewards would otherwise be lost because these are denoms that have been "stuck" in the consumer rewards pool account (they were never registered in a distribution allowlist). Moving them to the community pool preserves the balances, but does not distribute them to validators or delegators. Validators or other independent community members can calculate historical allocations and submit a separate governance proposal if they choose.

Additionally, any queries for ICS messages will now return empty stubs instead of historical data.

If the proposal passes, validators will be required to update the Cosmos Hub binary at the specified upgrade height.

# Release Details

- The release cna be found [here](https://github.com/cosmos/gaia/releases/tag/v28.0.0).
- The changelog can be found [here](https://github.com/cosmos/gaia/blob/v28.0.0/CHANGELOG.md).

**This upgrade is state-breaking and mandatory once the on-chain vote passes and the upgrade height is reached.**

# Governance votes

- This proposal has a voting period of **SEVEN DAYS**.
- If the proposal passes, validators will be required to update the Cosmos Hub binary at the specified upgrade height.

The following items summarize the voting options and what it means for this proposal:

YES - You agree that the Cosmos Hub should be updated with this release.

NO - You disagree that the Cosmos Hub should be updated with this release.

NO WITH VETO - A ‘NoWithVeto’ vote indicates a proposal either (1) is deemed to be spam, i.e., irrelevant to Cosmos Hub, (2) disproportionately infringes on minority interests, or (3) violates or encourages violation of the rules of engagement as currently set out by Cosmos Hub governance. If the number of ‘NoWithVeto’ votes is greater than a third of total votes, the proposal is rejected and the deposits are burned.

ABSTAIN - You wish to contribute to the quorum but you formally decline to vote either for or against the proposal.
