# Requests Contract - Technical Specification

## Contract Overview

**Name**: RequestsContract  
**Language**: Rust (MultiversX SC Framework v0.54.0)  
**Network**: MultiversX (Devnet, Testnet, Mainnet)  
**Type**: Stateful smart contract with owner privileges

## Architecture

### State Variables

#### 1. `numRequestsPerEgld: BigUint`
- **Storage Key**: `"numRequestsPerEgld"`
- **Type**: SingleValueMapper<BigUint>
- **Purpose**: Exchange rate - number of requests earned per 1 EGLD
- **Set By**: Constructor (`init`)
- **Constraints**: Must be > 0
- **Immutable**: Can be changed via upgrade

#### 2. `requests: Map<u64, BigUint>`
- **Storage Key**: `"requests" + id`
- **Type**: SingleValueMapper<BigUint> per ID
- **Purpose**: Track request balance for each user ID
- **Default Value**: 0 (if ID never credited)
- **Constraints**: None (can grow indefinitely)
- **Access**: Public read via `getRequests`, internal write via `addRequests`

---

## Function Specifications

### 1. Constructor: `init(numRequestsPerEgld: BigUint)`

**Visibility**: Public (called only during deployment)  
**Payable**: No  
**Access Control**: None (called by deployer)  
**Gas Estimate**: ~5,000 gas

#### Parameters
| Name | Type | Description |
|------|------|-------------|
| numRequestsPerEgld | BigUint | Exchange rate (requests per 1 EGLD) |

#### Validation
```
require!(numRequestsPerEgld > 0, "Number of requests per EGLD must be non-zero")
```

#### Logic
```
1. Validate numRequestsPerEgld > 0
2. Store numRequestsPerEgld in storage
3. Contract ready for use
```

#### Storage Changes
- `numRequestsPerEgld` ← parameter value

#### Events
- None

#### Error Cases
| Error | Cause | Resolution |
|-------|-------|-----------|
| "Number of requests per EGLD must be non-zero" | numRequestsPerEgld = 0 | Provide positive value |

---

### 2. Endpoint: `addRequests(id: u64)`

**Visibility**: Public  
**Payable**: Yes (EGLD only)  
**Access Control**: None (anyone can call)  
**Gas Estimate**: ~50,000 gas  
**Annotation**: `#[payable("EGLD")]` `#[endpoint(addRequests)]`

#### Parameters
| Name | Type | Description |
|------|------|-------------|
| id | u64 | User identifier (0 to 2^64-1) |

#### Payment
- **Token**: EGLD only
- **Amount**: Must be > 0
- **Denominated In**: Wei (1 EGLD = 10^18 wei)

#### Validation
```
require!(payment_amount > 0, "Payment amount must be greater than 0")
```

#### Logic
```
1. Get EGLD payment amount from call_value()
2. Validate amount > 0
3. Calculate: requests_to_add = payment_amount * numRequestsPerEgld
4. Update: requests[id] += requests_to_add
5. Emit AddRequests event
6. EGLD remains in contract (not transferred)
```

#### Storage Changes
- `requests[id]` ← `requests[id]` + (payment_amount * numRequestsPerEgld)

#### Events
```
Event: addRequests
Indexed:
  - id (u64): User identifier
  - egld_amount (BigUint): EGLD sent
Data:
  - requests_added (BigUint): Total requests added
```

#### Error Cases
| Error | Cause | Resolution |
|-------|-------|-----------|
| "Payment amount must be greater than 0" | No EGLD sent or 0 EGLD | Send EGLD with transaction |

#### Example Execution
```
Input:
  - id = 42
  - EGLD sent = 2.5 EGLD (2500000000000000000 wei)
  - numRequestsPerEgld = 100

Calculation:
  - requests_to_add = 2500000000000000000 * 100
  - requests_to_add = 250000000000000000000

Result:
  - requests[42] += 250000000000000000000
  - Event emitted: addRequests(42, 2500000000000000000, 250000000000000000000)
```

---

### 3. View: `getRequests(id: u64) -> BigUint`

**Visibility**: Public  
**Payable**: No  
**Access Control**: None (anyone can call)  
**Gas Estimate**: ~2,500 gas (view function, no state change)  
**Annotation**: `#[view(getRequests)]`

#### Parameters
| Name | Type | Description |
|------|------|-------------|
| id | u64 | User identifier |

#### Return Value
| Type | Description |
|------|-------------|
| BigUint | Current request balance for user (0 if never credited) |

#### Logic
```
1. Retrieve requests[id] from storage
2. Return value (0 if not found)
```

#### Storage Changes
- None (read-only)

#### Events
- None

#### Error Cases
- None (always succeeds)

#### Example Execution
```
Input:
  - id = 42

Scenarios:
  1. If requests[42] = 150
     → Return: 150

  2. If requests[42] never set
     → Return: 0
```

---

### 4. Endpoint: `changeNumRequestsPerEGLD(newNumRequestsPerEGLD: BigUint)`

**Visibility**: Public  
**Payable**: No  
**Access Control**: Owner only  
**Gas Estimate**: ~50,000 gas  
**Annotation**: `#[endpoint(changeNumRequestsPerEGLD)]`

#### Parameters
| Name | Type | Description |
|------|------|-------------|
| newNumRequestsPerEGLD | BigUint | New exchange rate (requests per 1 EGLD) |

#### Access Control
```
require!(caller == owner, "Only the owner can change the exchange rate")
```

#### Validation
```
require!(newNumRequestsPerEGLD > 0, "Number of requests per EGLD must be non-zero")
```

#### Logic
```
1. Verify caller is contract owner
2. Validate newNumRequestsPerEGLD > 0
3. Get old value from storage
4. Store new value in storage
5. Emit ChangeNumRequestsPerEGLD event with old and new values
```

#### Storage Changes
- `numRequestsPerEgld` ← newNumRequestsPerEGLD

#### Events
```
Event: changeNumRequestsPerEGLD
Data:
  - old_value (BigUint): Previous exchange rate
  - new_value (BigUint): New exchange rate
```

#### Error Cases
| Error | Cause | Resolution |
|-------|-------|-----------|
| "Only the owner can change the exchange rate" | Caller is not owner | Use owner's wallet |
| "Number of requests per EGLD must be non-zero" | newNumRequestsPerEGLD = 0 | Provide positive value |

#### Example Execution
```
Input:
  - Caller: owner_address
  - newNumRequestsPerEGLD: 200

Current State:
  - numRequestsPerEgld: 100

Result:
  - numRequestsPerEgld: 200
  - Event emitted: changeNumRequestsPerEGLD(100, 200)
  - Future addRequests calls use new rate (200)
```

---

### 5. Endpoint: `withdrawAll()`

**Visibility**: Public  
**Payable**: No  
**Access Control**: Owner only  
**Gas Estimate**: ~100,000 gas  
**Annotation**: `#[endpoint(withdrawAll)]`

#### Parameters
- None

#### Access Control
```
require!(caller == owner, "Only the owner can withdraw")
```

#### Validation
```
require!(contract_balance > 0, "No EGLD to withdraw")
```

#### Logic
```
1. Verify caller is contract owner
2. Get contract EGLD balance
3. Validate balance > 0
4. Transfer all EGLD to owner address
5. Emit Withdraw event
```

#### Storage Changes
- None (contract balance decreases, but not in storage)

#### Events
```
Event: withdraw
Indexed:
  - recipient (ManagedAddress): Owner address
Data:
  - amount (BigUint): EGLD withdrawn
```

#### Error Cases
| Error | Cause | Resolution |
|-------|-------|-----------|
| "Only the owner can withdraw" | Caller is not owner | Use owner's wallet |
| "No EGLD to withdraw" | Contract balance = 0 | Wait for users to send EGLD |

#### Example Execution
```
Input:
  - Caller: owner_address
  - Contract balance: 5.5 EGLD (5500000000000000000 wei)

Result:
  - Transfer 5500000000000000000 wei to owner_address
  - Event emitted: withdraw(owner_address, 5500000000000000000)
  - Contract balance: 0
```

---

## Data Flow Diagrams

### addRequests Flow
```
User sends EGLD
    ↓
addRequests(id) called
    ↓
Validate EGLD > 0
    ↓
Calculate: requests = EGLD * numRequestsPerEgld
    ↓
requests[id] += requests
    ↓
Emit event
    ↓
EGLD stored in contract
```

### getRequests Flow
```
Query getRequests(id)
    ↓
Retrieve requests[id]
    ↓
Return value (or 0)
    ↓
No state change
```

### changeNumRequestsPerEGLD Flow
```
Owner calls changeNumRequestsPerEGLD(newValue)
    ↓
Verify caller = owner
    ↓
Validate newValue > 0
    ↓
Get old value
    ↓
numRequestsPerEgld = newValue
    ↓
Emit event with old and new values
```

### withdrawAll Flow
```
Owner calls withdrawAll()
    ↓
Verify caller = owner
    ↓
Get contract balance
    ↓
Validate balance > 0
    ↓
Transfer to owner
    ↓
Emit event
    ↓
Contract balance = 0
```

---

## Storage Layout

### Storage Mappers

```
Storage Key: "numRequestsPerEgld"
Type: SingleValueMapper<BigUint>
Size: ~32 bytes
Access: Read in addRequests, write in init

Storage Key: "requests" + id (concatenated)
Type: SingleValueMapper<BigUint> per ID
Size: ~32 bytes per ID
Access: Read in getRequests, write in addRequests
```

### Example Storage State
```
After deployment with numRequestsPerEgld = 100:
  numRequestsPerEgld → 100

After addRequests(42, 1 EGLD):
  numRequestsPerEgld → 100
  requests[42] → 100

After addRequests(42, 0.5 EGLD):
  numRequestsPerEgld → 100
  requests[42] → 150

After addRequests(99, 2 EGLD):
  numRequestsPerEgld → 100
  requests[42] → 150
  requests[99] → 200
```

---

## Arithmetic & Precision

### BigUint Operations
- All amounts use `BigUint` (arbitrary precision unsigned integers)
- No overflow possible (BigUint grows as needed)
- Multiplication: `payment_amount * numRequestsPerEgld`

### EGLD Denomination
- 1 EGLD = 10^18 wei
- All contract operations use wei internally
- Example: 2.5 EGLD = 2500000000000000000 wei

### Example Calculation
```
User sends: 2.5 EGLD = 2500000000000000000 wei
numRequestsPerEgld: 100

requests_to_add = 2500000000000000000 * 100
                = 250000000000000000000

This is 250 * 10^18, representing 250 requests
```

---

## Security Analysis

### Access Control
✅ **withdrawAll**: Owner-only via `blockchain().get_owner_address()`  
✅ **changeNumRequestsPerEGLD**: Owner-only via `blockchain().get_owner_address()`  
✅ **addRequests**: Public (anyone can add requests)  
✅ **getRequests**: Public read-only  
✅ **init**: Deployment-only (implicit)

### Input Validation
✅ **numRequestsPerEgld > 0**: Checked in init  
✅ **payment_amount > 0**: Checked in addRequests  
✅ **contract_balance > 0**: Checked in withdrawAll

### Arithmetic Safety
✅ **BigUint**: No overflow possible  
✅ **Multiplication**: Safe with arbitrary precision  
✅ **Addition**: Safe with arbitrary precision

### Token Handling
✅ **EGLD only**: Enforced by `#[payable("EGLD")]`  
✅ **No token transfers in addRequests**: EGLD stays in contract  
✅ **Atomic withdrawal**: All EGLD transferred in one operation

### Potential Issues
⚠️ **No rate change mechanism**: numRequestsPerEgld can only be changed via upgrade  
⚠️ **No pause mechanism**: Contract always active  
⚠️ **No request consumption**: Requests only accumulate, never decrease  

---

## Gas Estimates

| Function | Operation | Gas Cost |
|----------|-----------|----------|
| init | Storage write | ~5,000 |
| addRequests | Storage read + write + event | ~50,000 |
| getRequests | Storage read | ~2,500 |
| withdrawAll | Storage read + transfer + event | ~100,000 |

---

## Events

### AddRequests Event
```
Name: addRequests
Indexed Topics:
  - id (u64)
  - egld_amount (BigUint)
Data:
  - requests_added (BigUint)
```

### ChangeNumRequestsPerEGLD Event
```
Name: changeNumRequestsPerEGLD
Data:
  - old_value (BigUint)
  - new_value (BigUint)
```

### Withdraw Event
```
Name: withdraw
Indexed Topics:
  - recipient (ManagedAddress)
Data:
  - amount (BigUint)
```

---

## ABI Interface

### Constructor
```json
{
  "name": "init",
  "onlyOwner": false,
  "inputs": [
    {
      "name": "num_requests_per_egld",
      "type": "BigUint"
    }
  ],
  "outputs": []
}
```

### addRequests
```json
{
  "name": "addRequests",
  "onlyOwner": false,
  "payable": true,
  "payableInTokens": ["EGLD"],
  "inputs": [
    {
      "name": "id",
      "type": "u64"
    }
  ],
  "outputs": []
}
```

### getRequests
```json
{
  "name": "getRequests",
  "onlyOwner": false,
  "readonly": true,
  "inputs": [
    {
      "name": "id",
      "type": "u64"
    }
  ],
  "outputs": [
    {
      "type": "BigUint"
    }
  ]
}
```

### changeNumRequestsPerEGLD
```json
{
  "name": "changeNumRequestsPerEGLD",
  "onlyOwner": true,
  "payable": false,
  "inputs": [
    {
      "name": "new_num_requests_per_egld",
      "type": "BigUint"
    }
  ],
  "outputs": []
}
```

### withdrawAll
```json
{
  "name": "withdrawAll",
  "onlyOwner": true,
  "payable": false,
  "inputs": [],
  "outputs": []
}
```

---

## Deployment Checklist

- [ ] Build contract: `sc-meta all build`
- [ ] Verify WASM output: `output/requests-contract.wasm`
- [ ] Verify ABI output: `output/requests-contract.abi.json`
- [ ] Choose network (devnet/testnet/mainnet)
- [ ] Prepare wallet with sufficient EGLD for gas
- [ ] Choose numRequestsPerEgld value
- [ ] Deploy contract with mxpy
- [ ] Verify contract address
- [ ] Test addRequests with small amount
- [ ] Test getRequests
- [ ] Test withdrawAll as owner
- [ ] Monitor contract events

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2025-01-13 | Initial release |

---

## References

- [MultiversX Docs](https://docs.multiversx.com)
- [SC Framework](https://docs.multiversx.com/developers/smart-contracts)
- [Storage Mappers](https://docs.multiversx.com/developers/developer-reference/storage-mappers)
- [Payments](https://docs.multiversx.com/developers/developer-reference/sc-payments)
