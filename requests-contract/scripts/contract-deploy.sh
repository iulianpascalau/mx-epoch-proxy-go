#!/bin/bash

source ./config.cfg

mxpy contract deploy --bytecode="${CONTRACT_WASM}" --metadata-payable --recall-nonce "${MXPY_SIGN[@]}" \
  --gas-limit=20000000 \
  --arguments "${NUM_REQUESTS_PER_EGLD}" \
  --send --proxy="${PROXY}" --chain="${CHAIN_ID}" || return


