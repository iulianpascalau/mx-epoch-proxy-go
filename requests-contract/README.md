# Requests Contract

A MultiversX smart contract for managing request credits. Users purchase requests by sending EGLD, and the contract owner can withdraw accumulated funds.

## Quick Start

### Build
```bash
sc-meta all build
```

### Deploy to Devnet
```bash
mxpy contract deploy \
    --bytecode output/requests-contract.wasm \
    --proxy https://devnet-api.multiversx.com \
    --chain D \
    --pem wallet.pem \
    --gas-limit 60000000 \
    --arguments 100 \
    --send
```

## Contract Functions

### Constructor: `init(numRequestsPerEgld)`
Initializes the contract with the exchange rate (requests per EGLD).

**Example**: `numRequestsPerEgld = 100` means 1 EGLD = 100 requests

### `addRequests(id)` - Payable
Add requests to a user ID by sending EGLD.

**Example**: Send 2.5 EGLD with `numRequestsPerEgld = 100` → 250 requests added

### `getRequests(id)` - View
Get the current request balance for a user ID.

**Returns**: BigUint (0 if never credited)

### `changeNumRequestsPerEGLD(newNumRequestsPerEGLD)` - Owner Only
Change the exchange rate (requests per EGLD).

**Example**: Change from 100 to 200 requests per EGLD

### `withdrawAll()` - Owner Only
Withdraw all accumulated EGLD to the owner's address.

## Storage

- **numRequestsPerEgld**: Exchange rate set during initialization
- **requests[id]**: Request count per user ID

## Events

- **addRequests**: Emitted when requests are added
- **changeNumRequestsPerEGLD**: Emitted when exchange rate is changed
- **withdraw**: Emitted when owner withdraws EGLD

## Documentation

See [DEPLOYMENT_GUIDE.md](./DEPLOYMENT_GUIDE.md) for detailed deployment instructions, testing examples, and integration guides.

## Project Structure

```
requests-contract/
├── src/
│   └── lib.rs              # Main contract code
├── meta/
│   ├── Cargo.toml
│   └── src/main.rs         # Build configuration
├── wasm/
│   ├── Cargo.toml
│   └── src/lib.rs          # WASM entry point
├── Cargo.toml              # Main dependencies
├── README.md               # This file
└── DEPLOYMENT_GUIDE.md     # Detailed deployment guide
```

## Building

### Prerequisites
- Rust 1.83.0+
- sc-meta: `cargo install multiversx-sc-meta --locked`

### Build Command
```bash
sc-meta all build
```

### Output
- `output/requests-contract.wasm` - Contract bytecode
- `output/requests-contract.abi.json` - Contract ABI

## Testing

Deploy to devnet and test the functions:

1. Add requests for user 42 with 1 EGLD
2. Query requests for user 42 (should be 100)
3. Add more requests with 0.5 EGLD
4. Query again (should be 150)
5. Withdraw all EGLD as owner

See [DEPLOYMENT_GUIDE.md](./DEPLOYMENT_GUIDE.md) for detailed testing examples.

## Security

- ✅ Owner-only withdrawal
- ✅ Non-zero exchange rate validation
- ✅ Payment amount validation
- ✅ BigUint overflow protection

## Networks

- **Devnet**: https://devnet-api.multiversx.com (Chain D)
- **Testnet**: https://testnet-api.multiversx.com (Chain T)
- **Mainnet**: https://api.multiversx.com (Chain 1)

## License

MIT
