# Requests Contract - Deployment & Usage Guide

## Overview

The **Requests Contract** is a MultiversX smart contract that manages request credits for users. Users can purchase requests by sending EGLD, and the contract owner can withdraw accumulated funds.

### Key Features

- **Constructor**: Initializes the contract with a configurable exchange rate (requests per EGLD)
- **addRequests**: Payable endpoint that credits requests to a user ID based on EGLD sent
- **getRequests**: View function to check request balance for a user ID
- **changeNumRequestsPerEGLD**: Owner-only function to change the exchange rate
- **withdrawAll**: Owner-only function to withdraw all contract EGLD

---

## Contract Functions

### 1. Constructor: `init(numRequestsPerEgld: BigUint)`

**Purpose**: Initialize the contract with the exchange rate

**Parameters**:
- `numRequestsPerEgld` (BigUint): Number of requests earned per 1 EGLD sent
  - Must be greater than 0
  - Example: If set to 100, sending 1 EGLD = 100 requests

**Example**:
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

---

### 2. Endpoint: `addRequests(id: u64)`

**Type**: Payable (EGLD only)

**Purpose**: Add requests to a user ID by sending EGLD

**Parameters**:
- `id` (u64): User identifier (any number from 0 to 2^64-1)

**Logic**:
```
requests_added = (EGLD_amount_sent) * numRequestsPerEgld
user_requests[id] += requests_added
```

**Example**: If `numRequestsPerEgld = 100` and you send 2.5 EGLD:
- Requests added = 2.5 * 100 = 250 requests
- User's total requests increase by 250

**Call with mxpy**:
```bash
mxpy contract call <contract-address> \
    --proxy https://devnet-api.multiversx.com \
    --chain D \
    --pem wallet.pem \
    --gas-limit 5000000 \
    --function "addRequests" \
    --arguments 42 \
    --value 1000000000000000000 \
    --send
```

**Call with xExchange/Web3 SDK**:
```javascript
const { TransactionBuilder } = require("@multiversx/sdk-core");

const transaction = new TransactionBuilder({
    config: networkConfig,
    sender: senderAddress,
    receiver: contractAddress,
    gasLimit: 5000000,
    data: "addRequests@2a", // 2a is hex for 42
    value: "1000000000000000000", // 1 EGLD in wei
});

await transaction.send();
```

---

### 3. View: `getRequests(id: u64) -> BigUint`

**Type**: View (read-only, no gas cost)

**Purpose**: Check the current request balance for a user ID

**Parameters**:
- `id` (u64): User identifier

**Returns**:
- `BigUint`: Number of requests for that ID (0 if never credited)

**Example Call**:
```bash
mxpy contract query <contract-address> \
    --proxy https://devnet-api.multiversx.com \
    --function "getRequests" \
    --arguments 42
```

**Response**: Returns the BigUint value representing total requests for user 42

---

### 4. Endpoint: `changeNumRequestsPerEGLD(newNumRequestsPerEGLD: BigUint)`

**Type**: Owner-only, no payment required

**Purpose**: Change the exchange rate (requests per EGLD)

**Requirements**:
- Caller must be the contract owner
- New value must be greater than 0

**Parameters**:
- `newNumRequestsPerEGLD` (BigUint): New exchange rate

**Logic**:
```
Verify caller is owner
Validate newNumRequestsPerEGLD > 0
Get old value
Store new value
Emit changeNumRequestsPerEGLD event
```

**Call with mxpy**:
```bash
mxpy contract call <contract-address> \
    --proxy https://devnet-api.multiversx.com \
    --chain D \
    --pem wallet.pem \
    --gas-limit 5000000 \
    --function "changeNumRequestsPerEGLD" \
    --arguments 200 \
    --send
```

**Example**: Change exchange rate from 100 to 200 requests per EGLD
- Old rate: 1 EGLD = 100 requests
- New rate: 1 EGLD = 200 requests
- Future addRequests calls will use the new rate

**Error Cases**:
- **"Only the owner can change the exchange rate"**: Caller is not owner
- **"Number of requests per EGLD must be non-zero"**: New value is 0

---

### 5. Endpoint: `withdrawAll()`

**Type**: Owner-only, no payment required

**Purpose**: Withdraw all accumulated EGLD from the contract

**Requirements**:
- Caller must be the contract owner
- Contract must have EGLD balance > 0

**Logic**:
```
balance = contract_EGLD_balance
transfer(balance, owner_address)
emit WithdrawEvent(owner_address, balance)
```

**Call with mxpy**:
```bash
mxpy contract call <contract-address> \
    --proxy https://devnet-api.multiversx.com \
    --chain D \
    --pem wallet.pem \
    --gas-limit 5000000 \
    --function "withdrawAll" \
    --send
```

---

## Building the Contract

### Prerequisites

Ensure you have Rust installed:
```bash
rustup update
rustc --version
```

Install sc-meta (MultiversX smart contract meta tool):
```bash
cargo install multiversx-sc-meta --locked
```

### Build Steps

1. Navigate to the contract directory:
```bash
cd requests-contract
```

2. Build the contract:
```bash
sc-meta all build
```

3. Verify output files:
```bash
ls -la output/
# Should contain:
# - requests-contract.wasm (contract bytecode)
# - requests-contract.abi.json (contract ABI)
```

### Build Options

**Debug build with WAT output** (for debugging):
```bash
cd meta && cargo run build-dbg
```

**Clean build**:
```bash
sc-meta all clean
sc-meta all build
```

---

## Deployment

### Prerequisites

1. Have a MultiversX wallet with EGLD for gas fees
2. Export your wallet private key:
```bash
# Create wallet.pem file with your private key
echo "your_private_key_here" > wallet.pem
```

3. Ensure you have mxpy installed:
```bash
pip install multiversx-sdk-py
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

**Parameters**:
- `--bytecode`: Path to compiled WASM file
- `--proxy`: Network endpoint (devnet/testnet/mainnet)
- `--chain`: Chain ID (D=devnet, T=testnet, 1=mainnet)
- `--pem`: Wallet private key file
- `--gas-limit`: Gas limit for deployment (60M recommended)
- `--arguments`: Constructor argument (numRequestsPerEgld = 100)

**Output**: Returns the contract address (e.g., `erd1abc...xyz`)

### Deploy to Testnet

```bash
mxpy contract deploy \
    --bytecode output/requests-contract.wasm \
    --proxy https://testnet-api.multiversx.com \
    --chain T \
    --pem wallet.pem \
    --gas-limit 60000000 \
    --arguments 100 \
    --send
```

### Deploy to Mainnet

```bash
mxpy contract deploy \
    --bytecode output/requests-contract.wasm \
    --proxy https://api.multiversx.com \
    --chain 1 \
    --pem wallet.pem \
    --gas-limit 60000000 \
    --arguments 100 \
    --send
```

---

## Upgrade Contract

To upgrade an existing contract with new code:

```bash
mxpy contract upgrade <contract-address> \
    --bytecode output/requests-contract.wasm \
    --proxy https://devnet-api.multiversx.com \
    --chain D \
    --pem wallet.pem \
    --gas-limit 60000000 \
    --arguments 100 \
    --send
```

---

## Contract Storage

### Storage Mappers

1. **numRequestsPerEgld** (SingleValueMapper<BigUint>)
   - Stores the exchange rate set during initialization
   - Key: `"numRequestsPerEgld"`

2. **requests** (SingleValueMapper<BigUint> per ID)
   - Stores request count for each user ID
   - Key: `"requests" + id`
   - Returns 0 if ID never credited

---

## Events

### AddRequests Event

Emitted when requests are successfully added:
```
Event: addRequests
Indexed Topics:
  - id (u64): User ID
  - egld_amount (BigUint): EGLD sent
Data:
  - requests_added (BigUint): Total requests added
```

### Withdraw Event

Emitted when owner withdraws EGLD:
```
Event: withdraw
Indexed Topics:
  - recipient (ManagedAddress): Owner address
Data:
  - amount (BigUint): EGLD withdrawn
```

---

## Error Handling

### Constructor Errors

- **"Number of requests per EGLD must be non-zero"**
  - Cause: numRequestsPerEgld = 0
  - Fix: Provide a positive number

### addRequests Errors

- **"Payment amount must be greater than 0"**
  - Cause: Called without sending EGLD
  - Fix: Send EGLD with the transaction

### withdrawAll Errors

- **"Only the owner can withdraw"**
  - Cause: Called by non-owner
  - Fix: Use the owner's wallet

- **"No EGLD to withdraw"**
  - Cause: Contract has 0 EGLD balance
  - Fix: Wait for users to send EGLD

---

## Testing

### Manual Testing on Devnet

1. **Deploy contract** with `numRequestsPerEgld = 100`

2. **Add requests** for user ID 42 with 1 EGLD:
```bash
mxpy contract call <contract-address> \
    --proxy https://devnet-api.multiversx.com \
    --chain D \
    --pem wallet.pem \
    --gas-limit 5000000 \
    --function "addRequests" \
    --arguments 42 \
    --value 1000000000000000000 \
    --send
```

3. **Query requests** for user 42:
```bash
mxpy contract query <contract-address> \
    --proxy https://devnet-api.multiversx.com \
    --function "getRequests" \
    --arguments 42
```
Expected: `100` (1 EGLD * 100 requests/EGLD)

4. **Add more requests** for user 42 with 0.5 EGLD:
```bash
mxpy contract call <contract-address> \
    --proxy https://devnet-api.multiversx.com \
    --chain D \
    --pem wallet.pem \
    --gas-limit 5000000 \
    --function "addRequests" \
    --arguments 42 \
    --value 500000000000000000 \
    --send
```

5. **Query requests** again:
```bash
mxpy contract query <contract-address> \
    --proxy https://devnet-api.multiversx.com \
    --function "getRequests" \
    --arguments 42
```
Expected: `150` (100 + 50)

6. **Withdraw all EGLD**:
```bash
mxpy contract call <contract-address> \
    --proxy https://devnet-api.multiversx.com \
    --chain D \
    --pem wallet.pem \
    --gas-limit 5000000 \
    --function "withdrawAll" \
    --send
```

---

## Network Endpoints

| Network | Proxy URL | Chain ID | Use Case |
|---------|-----------|----------|----------|
| Devnet | `https://devnet-api.multiversx.com` | D | Development & Testing |
| Testnet | `https://testnet-api.multiversx.com` | T | Pre-production Testing |
| Mainnet | `https://api.multiversx.com` | 1 | Production |

---

## Security Considerations

1. **Owner Verification**: Only the contract owner can withdraw funds
2. **Non-zero Exchange Rate**: Constructor validates that numRequestsPerEgld > 0
3. **Payment Validation**: addRequests requires EGLD payment > 0
4. **Arithmetic Safety**: Uses BigUint for overflow protection

---

## Integration Examples

### JavaScript/TypeScript Integration

```typescript
import { ApiNetworkProvider } from "@multiversx/sdk-network-providers";
import { SmartContract, AbiRegistry } from "@multiversx/sdk-core";
import * as fs from "fs";

const networkProvider = new ApiNetworkProvider("https://devnet-api.multiversx.com");
const abiJson = JSON.parse(fs.readFileSync("output/requests-contract.abi.json", "utf8"));
const abiRegistry = AbiRegistry.create(abiJson);

const contract = new SmartContract({
    address: "erd1abc...xyz",
    abi: abiRegistry,
});

// Query requests for user 42
const interaction = contract.methods.getRequests([42]);
const result = await networkProvider.queryContract(interaction);
console.log("Requests for user 42:", result.returnCode);
```

---

## Troubleshooting

### Build Fails with "Cannot find multiversx-sc"

```bash
# Update Cargo dependencies
cargo update

# Rebuild
sc-meta all build
```

### Deployment Fails with "Insufficient funds"

- Ensure wallet has enough EGLD for gas fees
- Devnet EGLD can be obtained from the faucet: https://devnet-faucet.multiversx.com

### Contract Call Returns Error

- Check contract address is correct
- Verify function name matches (case-sensitive)
- Ensure arguments are in correct format

---

## Support & Resources

- **MultiversX Docs**: https://docs.multiversx.com
- **sc-meta Documentation**: https://docs.multiversx.com/developers/meta/sc-meta
- **Example Contracts**: https://github.com/multiversx/mx-sdk-rs/tree/master/contracts/examples
- **Discord Community**: https://discord.gg/multiversx
