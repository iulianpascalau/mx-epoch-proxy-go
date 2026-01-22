#!/bin/bash

source ./config.cfg

mxpy contract call "${CONTRACT_ADDRESS}" --recall-nonce "${MXPY_SIGN[@]}" \
    --gas-limit=3000000 --function="changeNumRequestsPerEGLD" \
    --arguments "${NUM_REQUESTS_PER_EGLD}" \
    --send --proxy="${PROXY}" --chain="${CHAIN_ID}"