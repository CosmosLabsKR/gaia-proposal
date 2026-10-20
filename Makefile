SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

ifneq ($(filter generate-mainnet,$(MAKECMDGOALS)),)
ifneq ($(filter generate-testnet,$(MAKECMDGOALS)),)
$(error Run only one network generation target at a time)
endif
override NETWORK = mainnet
endif
NETWORK ?= mainnet
ifneq ($(filter generate-testnet,$(MAKECMDGOALS)),)
override NETWORK = testnet
endif
ifeq ($(filter mainnet testnet,$(NETWORK)),)
$(error NETWORK must be mainnet or testnet)
endif
include config/$(NETWORK).mk
-include config/$(NETWORK).local.mk

GO ?= go
BINARY ?= bin/gaia-proposal
PROPOSAL_DIR = $(word 2,$(MAKECMDGOALS))
DIRECTORY_COMMANDS = generate generate-mainnet generate-testnet

ifneq ($(filter $(firstword $(MAKECMDGOALS)),$(DIRECTORY_COMMANDS)),)
ifneq ($(PROPOSAL_DIR),)
.PHONY: $(PROPOSAL_DIR)
$(PROPOSAL_DIR):
	@:
endif
endif

.PHONY: build install check generate generate-all generate-mainnet generate-testnet

build:
	mkdir -p $(dir $(BINARY))
	$(GO) build -trimpath -o $(BINARY) ./cmd/gaia-proposal

install: 
	$(GO) install ./cmd/gaia-proposal

generate: build
	$(if $(PROPOSAL_DIR),,$(error proposal directory is required. Example: make generate v1.0.0-test))
	$(BINARY) generate "proposals/$(PROPOSAL_DIR)"

generate-all: build
	$(BINARY) generate-all

check: build
	$(BINARY) generate-all --check

# These targets only generate a transaction and print a submission command.
generate-mainnet generate-testnet: generate
	@test -n "$(TX_FROM)" || { echo "Set TX_FROM in config/$(NETWORK).local.mk" >&2; exit 1; }
	@test -n "$(RPC_NODE)" || { echo "Set RPC_NODE in config/$(NETWORK).local.mk" >&2; exit 1; }
	@gaiad tx gov submit-proposal "./proposals/$(PROPOSAL_DIR)/proposal.json" --from "$(TX_FROM)" --generate-only --chain-id "$(CHAIN_ID)" | jq -e '.'
	@echo 'gaiad tx gov submit-proposal "./proposals/$(PROPOSAL_DIR)/proposal.json" --from "$(TX_FROM)" --node "$(RPC_NODE)" --chain-id "$(CHAIN_ID)" --gas-prices "$(GAS_PRICES)" --gas "$(GAS)"'

.PHONY: deposit
deposit:
	@test -n "$(TX_FROM)" -a -n "$(RPC_NODE)" || { echo "Set TX_FROM and RPC_NODE in config/$(NETWORK).local.mk" >&2; exit 1; }
	@test -n "$(PROPOSAL_ID)" -a -n "$(DEPOSIT_AMOUNT)" || { echo "Set PROPOSAL_ID and DEPOSIT_AMOUNT" >&2; exit 1; }
	@echo 'gaiad tx gov deposit "$(PROPOSAL_ID)" "$(DEPOSIT_AMOUNT)" --from "$(TX_FROM)" --node "$(RPC_NODE)" --chain-id "$(CHAIN_ID)" --gas-prices "$(GAS_PRICES)" --gas "$(GAS)"'
