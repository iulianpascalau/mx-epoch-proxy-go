# Requests Contract - Testing Guide

## Overview

This document describes the comprehensive unit tests for the Requests Contract. Tests are written using the MultiversX scenario framework and cover all contract functions and edge cases.

## Test Structure

### Test Framework
- **Framework**: MultiversX Scenario Testing Framework
- **Language**: Rust
- **Location**: `/tests/integration_test.rs`
- **Scenarios**: `/scenarios/`

### Running Tests

#### Prerequisites
```bash
# Build the contract first
sc-meta all build

# Ensure Rust is installed
rustc --version
```

#### Run All Tests
```bash
cd /app/project/requests-contract
cargo test
```

#### Run Specific Test
```bash
cargo test test_init_with_valid_value
```

#### Run Tests with Output
```bash
cargo test -- --nocapture
```

#### Run Tests in Parallel
```bash
cargo test -- --test-threads=4
```

---

## Test Cases

### 1. Constructor Tests

#### `test_init_with_valid_value`
**File**: `scenarios/init_valid.scen.json`

**Purpose**: Verify contract deploys successfully with valid exchange rate

**Steps**:
1. Deploy contract with `numRequestsPerEgld = 100`
2. Verify deployment succeeds (status 0)

**Expected Result**: ✅ Contract deployed successfully

---

#### `test_init_with_zero_value`
**File**: `scenarios/init_zero.scen.json`

**Purpose**: Verify constructor rejects zero exchange rate

**Steps**:
1. Attempt to deploy with `numRequestsPerEgld = 0`
2. Verify deployment fails

**Expected Result**: ✅ Deployment fails with error "Number of requests per EGLD must be non-zero"

---

### 2. Add Requests Tests

#### `test_add_requests_single_user`
**File**: `scenarios/add_requests_single.scen.json`

**Purpose**: Verify single user can add requests correctly

**Steps**:
1. Deploy contract with rate 100
2. User sends 1 EGLD for ID 42
3. Query requests for ID 42

**Expected Result**: ✅ User has 100 requests (1 EGLD * 100 rate)

---

#### `test_add_requests_multiple_users`
**File**: `scenarios/add_requests_multiple.scen.json`

**Purpose**: Verify multiple users have independent request balances

**Steps**:
1. Deploy contract with rate 100
2. User1 sends 1 EGLD for ID 1
3. User2 sends 2 EGLD for ID 2
4. Query both users

**Expected Result**: ✅ User1 has 100 requests, User2 has 200 requests

---

#### `test_add_requests_accumulation`
**File**: `scenarios/add_requests_accumulation.scen.json`

**Purpose**: Verify requests accumulate for same user

**Steps**:
1. Deploy contract with rate 100
2. User sends 1 EGLD for ID 42 → 100 requests
3. Query: should have 100
4. User sends 0.5 EGLD for ID 42 → 50 more requests
5. Query: should have 150

**Expected Result**: ✅ Requests accumulate correctly (100 + 50 = 150)

---

### 3. Get Requests Tests

#### `test_get_requests_existing_user`
**File**: `scenarios/get_requests_existing.scen.json`

**Purpose**: Verify querying existing user returns correct balance

**Steps**:
1. Deploy contract with rate 100
2. User sends 2 EGLD for ID 99
3. Query requests for ID 99

**Expected Result**: ✅ Returns 200 requests (2 EGLD * 100 rate)

---

#### `test_get_requests_nonexistent_user`
**File**: `scenarios/get_requests_nonexistent.scen.json`

**Purpose**: Verify querying nonexistent user returns 0

**Steps**:
1. Deploy contract with rate 100
2. Query requests for ID 999 (never credited)

**Expected Result**: ✅ Returns 0

---

### 4. Change Exchange Rate Tests

#### `test_change_exchange_rate_valid`
**File**: `scenarios/change_rate_valid.scen.json`

**Purpose**: Verify owner can change exchange rate

**Steps**:
1. Deploy contract with rate 100
2. Owner calls changeNumRequestsPerEGLD(200)
3. Verify transaction succeeds

**Expected Result**: ✅ Exchange rate changed successfully

---

#### `test_change_exchange_rate_zero`
**File**: `scenarios/change_rate_zero.scen.json`

**Purpose**: Verify changing rate to zero is rejected

**Steps**:
1. Deploy contract with rate 100
2. Owner tries to change rate to 0
3. Verify transaction fails

**Expected Result**: ✅ Transaction fails with error "Number of requests per EGLD must be non-zero"

---

#### `test_change_exchange_rate_non_owner`
**File**: `scenarios/change_rate_non_owner.scen.json`

**Purpose**: Verify non-owner cannot change exchange rate

**Steps**:
1. Deploy contract with rate 100
2. Non-owner tries to change rate to 200
3. Verify transaction fails

**Expected Result**: ✅ Transaction fails with error "Only the owner can change the exchange rate"

---

### 5. Withdraw Tests

#### `test_withdraw_all_success`
**File**: `scenarios/withdraw_success.scen.json`

**Purpose**: Verify owner can withdraw accumulated EGLD

**Steps**:
1. Deploy contract with rate 100
2. User sends 2 EGLD for ID 42
3. Owner calls withdrawAll()
4. Verify transaction succeeds

**Expected Result**: ✅ Owner receives 2 EGLD

---

#### `test_withdraw_all_empty_contract`
**File**: `scenarios/withdraw_empty.scen.json`

**Purpose**: Verify withdrawing from empty contract fails

**Steps**:
1. Deploy contract with rate 100
2. Owner tries to withdrawAll() without any deposits
3. Verify transaction fails

**Expected Result**: ✅ Transaction fails with error "No EGLD to withdraw"

---

#### `test_withdraw_all_non_owner`
**File**: `scenarios/withdraw_non_owner.scen.json`

**Purpose**: Verify non-owner cannot withdraw

**Steps**:
1. Deploy contract with rate 100
2. User sends 2 EGLD for ID 42
3. Non-owner tries to withdrawAll()
4. Verify transaction fails

**Expected Result**: ✅ Transaction fails with error "Only the owner can withdraw"

---

### 6. Integration Tests

#### `test_full_workflow`
**File**: `scenarios/full_workflow.scen.json`

**Purpose**: Verify complete workflow works correctly

**Steps**:
1. Deploy contract with rate 100
2. User1 sends 1 EGLD for ID 1 → 100 requests
3. User2 sends 2 EGLD for ID 2 → 200 requests
4. Query both users
5. Owner changes rate to 150
6. User1 sends 1 EGLD for ID 1 → 150 more requests
7. Query User1 (should have 250 total)
8. Owner withdraws all EGLD

**Expected Result**: ✅ All operations succeed with correct values

---

#### `test_rate_change_affects_future_requests`
**File**: `scenarios/rate_change_affects_future.scen.json`

**Purpose**: Verify rate change only affects future requests

**Steps**:
1. Deploy contract with rate 100
2. User sends 1 EGLD for ID 42 → 100 requests
3. Query: should have 100
4. Owner changes rate to 200
5. User sends 1 EGLD for ID 42 → 200 more requests
6. Query: should have 300 (100 + 200)

**Expected Result**: ✅ Rate change affects only new requests (100 + 200 = 300)

---

## Test Coverage

### Functions Tested
- ✅ `init(numRequestsPerEgld)` - Constructor
- ✅ `addRequests(id)` - Payable endpoint
- ✅ `getRequests(id)` - View function
- ✅ `changeNumRequestsPerEGLD(newNumRequestsPerEGLD)` - Owner-only
- ✅ `withdrawAll()` - Owner-only

### Scenarios Covered
- ✅ Valid operations
- ✅ Invalid inputs (zero values)
- ✅ Access control (owner-only functions)
- ✅ Edge cases (empty contract, nonexistent users)
- ✅ State changes (accumulation, rate changes)
- ✅ Complex workflows (multiple users, rate changes)

### Error Cases Tested
- ✅ Zero exchange rate in constructor
- ✅ Zero exchange rate in changeNumRequestsPerEGLD
- ✅ Non-owner calling changeNumRequestsPerEGLD
- ✅ Non-owner calling withdrawAll
- ✅ Withdrawing from empty contract

---

## Test Results

### Expected Output
```
running 15 tests
test test_init_with_valid_value ... ok
test test_init_with_zero_value ... ok
test test_add_requests_single_user ... ok
test test_add_requests_multiple_users ... ok
test test_add_requests_accumulation ... ok
test test_get_requests_existing_user ... ok
test test_get_requests_nonexistent_user ... ok
test test_change_exchange_rate_valid ... ok
test test_change_exchange_rate_zero ... ok
test test_change_exchange_rate_non_owner ... ok
test test_withdraw_all_success ... ok
test test_withdraw_all_empty_contract ... ok
test test_withdraw_all_non_owner ... ok
test test_full_workflow ... ok
test test_rate_change_affects_future_requests ... ok

test result: ok. 15 passed; 0 failed; 0 ignored
```

---

## Scenario File Format

Each scenario file follows this structure:

```json
{
  "name": "Test name",
  "comment": "Test description",
  "steps": [
    {
      "step": "setState",
      "accounts": {
        "account_name": {
          "nonce": 0,
          "balance": "amount_in_wei"
        }
      }
    },
    {
      "step": "scDeploy",
      "txId": "deploy",
      "tx": {
        "from": "owner",
        "contractCode": "file:output/requests-contract.wasm",
        "arguments": ["100"],
        "gasLimit": 60000000,
        "gasPrice": 1000000000
      }
    },
    {
      "step": "scCall",
      "txId": "call_id",
      "tx": {
        "from": "caller",
        "to": "sc:deploy",
        "function": "functionName",
        "arguments": ["arg1", "arg2"],
        "value": "amount_in_wei",
        "gasLimit": 5000000,
        "gasPrice": 1000000000
      },
      "expect": {
        "status": 0,
        "logs": []
      }
    },
    {
      "step": "scQuery",
      "txId": "query_id",
      "tx": {
        "to": "sc:deploy",
        "function": "viewFunction",
        "arguments": ["arg1"]
      },
      "expect": {
        "out": ["expected_output"]
      }
    }
  ]
}
```

---

## Debugging Tests

### View Test Output
```bash
cargo test -- --nocapture --test-threads=1
```

### Run Single Test with Backtrace
```bash
RUST_BACKTRACE=1 cargo test test_init_with_valid_value -- --nocapture
```

### Check Scenario File Syntax
```bash
# Validate JSON syntax
cat scenarios/init_valid.scen.json | jq .
```

---

## Adding New Tests

### Steps to Add a Test

1. **Create scenario file** in `/scenarios/`:
   ```bash
   touch scenarios/new_test.scen.json
   ```

2. **Write scenario** following the format above

3. **Add test function** in `tests/integration_test.rs`:
   ```rust
   #[test]
   fn test_new_functionality() {
       let mut world = world();
       world.run("scenarios/new_test.scen.json");
   }
   ```

4. **Run test**:
   ```bash
   cargo test test_new_functionality
   ```

---

## Test Maintenance

### When to Update Tests
- When adding new functions
- When changing function behavior
- When fixing bugs
- When adding new validation

### Best Practices
- Keep tests independent
- Use descriptive names
- Document complex scenarios
- Test both success and failure cases
- Verify state changes
- Test edge cases

---

## Continuous Integration

### Running Tests in CI/CD
```bash
# Build contract
sc-meta all build

# Run all tests
cargo test

# Generate test report
cargo test -- --nocapture > test_results.txt
```

---

## Performance Considerations

### Gas Estimates
- `init`: ~5,000 gas
- `addRequests`: ~50,000 gas
- `getRequests`: ~2,500 gas (view, no gas cost)
- `changeNumRequestsPerEGLD`: ~50,000 gas
- `withdrawAll`: ~100,000 gas

### Test Execution Time
- All tests: ~5-10 seconds
- Individual test: ~0.5 seconds

---

## Support

For issues or questions about testing:
1. Check test output for error messages
2. Review scenario file syntax
3. Verify contract builds successfully
4. Check MultiversX documentation

---

**Status**: ✅ All tests ready to run

Run `cargo test` to execute all tests.
