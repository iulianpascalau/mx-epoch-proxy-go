# Crypto-Payment Feature - Manual Testing Flows

**Document Version:** 1.0
**Date:** January 2026
**Branch:** rc/v1.3
**Companion Document:** crypto-payment-testing-strategy.md
**Commit hash** 0cd2ffe496506819f8cf2799b6b5e67ef36502f3
---

## 1. Purpose

This document provides step-by-step manual testing flows for the Crypto-Payment feature. Each flow includes prerequisites, steps, expected results, and notes on potential integration test conversion.

**Legend:**
- **[IT-CANDIDATE]** - This scenario is a good candidate for conversion to an integration test
- **[E2E-ONLY]** - This scenario requires real blockchain interaction and should remain manual/E2E
- **[UI-ONLY]** - This scenario tests UI behavior that cannot be easily automated at backend level

---

## 2. Test Environment Setup

### 2.1 Prerequisites

| Item | Details |
|------|---------|
| Proxy Service | Running on `http://localhost:8080` |
| Crypto-Payment Service | Running on `http://localhost:8081` |
| Network | MultiversX Devnet |
| Browser | Chrome/Firefox with DevTools |
| MultiversX Web Wallet | Access to devnet wallet |
| Test eGLD | Minimum 1 eGLD in test wallet |

### 2.2 Test Accounts

Prepare the following accounts before testing:

| Account | Type | Purpose |
|---------|------|---------|
| `free-user@test.com` | Free | Standard free user for upgrade flows |
| `premium-user@test.com` | Premium (credits) | User with purchased credits |
| `unlimited-user@test.com` | Premium (unlimited) | Admin-granted unlimited account |
| `depleted-user@test.com` | Premium (depleted) | User who exhausted their credits |
| `admin@test.com` | Admin | For admin panel testing |

### 2.3 Funding Test Wallet

1. Go to MultiversX Devnet Faucet: https://devnet-wallet.multiversx.com/faucet
2. Request test eGLD (10 eGLD recommended)
3. Verify balance in wallet

---

## 3. Core Payment Flows

### 3.1 Flow: First-Time Address Creation

**ID:** FLOW-001
**Priority:** Critical
**Tag:** [IT-CANDIDATE]

**Preconditions:**
- User `free-user@test.com` has no `crypto_payment_id` in database
- Crypto-payment service is running
- Contract is NOT paused

**Steps:**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login as `free-user@test.com` | Dashboard loads successfully |
| 2 | Navigate to Dashboard | "Upgrade to Premium" section visible |
| 3 | Click "Request Payment Address" | Loading indicator appears |
| 4 | Wait for response | Payment details section appears |
| 5 | Verify Payment ID displayed | Non-zero numeric ID shown (e.g., "Payment ID: 42") |
| 6 | Verify Address displayed | Valid erd1... address (62 characters) |
| 7 | Verify "Current Requests Balance" | Shows "0" |
| 8 | Verify "Rate" displayed | Shows configured rate (e.g., "1,000,000 requests per 1 eGLD") |
| 9 | Verify buttons visible | "Copy Address", "Refresh Balance", "Open Wallet" present |
| 10 | Check database | `users.crypto_payment_id = 42` for this user |

**Verification Queries:**
```sql
-- Check user has payment ID
SELECT username, crypto_payment_id, max_requests FROM users WHERE username = 'free-user@test.com';
```

**Notes:**
- This flow tests the happy path of address creation
- Can be converted to integration test by mocking the crypto-payment service response

---

### 3.2 Flow: Duplicate Address Request Prevention

**ID:** FLOW-002
**Priority:** High
**Tag:** [IT-CANDIDATE]

**Preconditions:**
- User `free-user@test.com` already has a `crypto_payment_id`

**Steps:**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login as `free-user@test.com` | Dashboard loads |
| 2 | Navigate to Dashboard | Payment details already visible (not "Request" button) |
| 3 | Using browser DevTools, call POST `/api/crypto-payment/create-address` directly | Returns 400 with "User already has a payment ID" |
| 4 | Verify database unchanged | Same `crypto_payment_id` as before |

**API Test:**
```bash
curl -X POST http://localhost:8080/api/crypto-payment/create-address \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json"
# Expected: 400 {"error": "User already has a payment ID"}
```

---

### 3.3 Flow: Concurrent Address Request (Race Condition)

**ID:** FLOW-003
**Priority:** High
**Tag:** [IT-CANDIDATE]

**Preconditions:**
- Fresh user with no `crypto_payment_id`
- Ability to send concurrent requests

**Steps:**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Prepare two terminal windows with curl commands | Ready to execute |
| 2 | Execute both POST `/api/crypto-payment/create-address` simultaneously | One succeeds (200), one fails (400) OR both get same ID |
| 3 | Check database | User has exactly ONE `crypto_payment_id` |
| 4 | Check crypto-payment database | No orphaned addresses |

**Concurrent Test Script:**
```bash
#!/bin/bash
TOKEN="<jwt_token>"
# Run in parallel
curl -X POST http://localhost:8080/api/crypto-payment/create-address -H "Authorization: Bearer $TOKEN" &
curl -X POST http://localhost:8080/api/crypto-payment/create-address -H "Authorization: Bearer $TOKEN" &
wait
```

**Notes:**
- Tests `UserMutexManager` effectiveness
- Critical for preventing orphaned addresses

---

### 3.4 Flow: Complete Payment Cycle (End-to-End)

**ID:** FLOW-004
**Priority:** Critical
**Tag:** [E2E-ONLY]

**Preconditions:**
- Fresh user with newly created payment address
- Test wallet with sufficient eGLD (0.1 eGLD minimum)

**Steps:**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Note the deposit address from UI | Address recorded |
| 2 | Open MultiversX Web Wallet | Wallet loads |
| 3 | Send 0.05 eGLD to deposit address | Transaction submitted |
| 4 | Wait for transaction confirmation | ~6 seconds on devnet |
| 5 | Note the transaction hash | Hash recorded for verification |
| 6 | Wait for balance processor cycle | Up to 60 seconds (configurable) |
| 7 | Click "Refresh Balance" in UI | Balance updates to show purchased requests |
| 8 | Verify balance calculation | `0.05 * requestsPerEGLD` requests shown |
| 9 | Verify account upgraded | User now treated as Premium |
| 10 | Check `users.max_requests` in database | Updated to match contract balance |

**Verification:**
```bash
# Check transaction on explorer
https://devnet-explorer.multiversx.com/transactions/<tx_hash>

# Query smart contract directly
curl "https://devnet-gateway.multiversx.com/vm-values/query" \
  -H "Content-Type: application/json" \
  -d '{"scAddress":"erd1qqqqqqqqqqqqqpgqf4tjq8wpjm6nh0v8y0cmqp7qwqk9eymf945syjea0j","funcName":"getRequests","args":["<payment_id_hex>"]}'
```

**Timing Notes:**
- Balance processor runs every 60 seconds by default
- Relayed transaction may take additional ~6 seconds to confirm
- Total wait: up to ~70 seconds after initial deposit confirms

---

### 3.5 Flow: Minimum Balance Threshold

**ID:** FLOW-005
**Priority:** High
**Tag:** [E2E-ONLY]

**Preconditions:**
- User with payment address
- Exact 0.009 eGLD available (below 0.01 threshold)

**Steps:**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Send exactly 0.009 eGLD to deposit address | Transaction confirms |
| 2 | Wait for balance processor cycle (60s+) | No relay transaction triggered |
| 3 | Verify address still has 0.009 balance | Balance not swept |
| 4 | Send additional 0.001 eGLD (total now 0.01) | Transaction confirms |
| 5 | Wait for balance processor cycle | Relay transaction triggered |
| 6 | Verify balance swept | Address balance now ~0 |
| 7 | Check smart contract | User credited with requests |

**Notes:**
- Tests `MinimumBalanceToProcess = 0.01` configuration
- Edge case at exact threshold value (0.01) should process

---

### 3.6 Flow: Top-Up Existing Balance

**ID:** FLOW-006
**Priority:** High
**Tag:** [E2E-ONLY]

**Preconditions:**
- Premium user with existing balance (e.g., 50,000 requests)

**Steps:**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Note current balance (UI and database) | e.g., 50,000 requests |
| 2 | Send 0.1 eGLD to same deposit address | Transaction confirms |
| 3 | Wait for relay cycle | ~70 seconds |
| 4 | Click "Refresh" in UI | Balance updates |
| 5 | Verify new balance | `50,000 + (0.1 * requestsPerEGLD)` |
| 6 | Verify `max_requests` increased | Database updated (only increases) |

---

## 4. UI Component Flows

### 4.1 Flow: Free User Dashboard View

**ID:** FLOW-UI-001
**Priority:** Medium
**Tag:** [UI-ONLY]

**Preconditions:**
- User with `account_type = 'free'`
- No `crypto_payment_id`

**Steps:**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login as free user | Dashboard loads |
| 2 | Verify Account Status badge | Shows "Free" badge |
| 3 | Verify "Upgrade to Premium" section | Section visible below account status |
| 4 | Verify upgrade description text | "Get unlimited requests by making a crypto payment." |
| 5 | Verify "Request Payment Address" button | Button enabled and clickable |
| 6 | Verify Premium section NOT visible | No progress bar, no "Add More Requests" |

---

### 4.2 Flow: Premium User with Credits Dashboard View

**ID:** FLOW-UI-002
**Priority:** Medium
**Tag:** [UI-ONLY]

**Preconditions:**
- User with purchased credits
- `request_count < max_requests`

**Steps:**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login as `premium-user@test.com` | Dashboard loads |
| 2 | Verify "Premium Account" header | Shows with checkmark |
| 3 | Verify Payment ID displayed | Shows assigned ID |
| 4 | Verify Address displayed | Shows erd1... address |
| 5 | Verify progress bar | Shows "X / Y" requests (percentage used) |
| 6 | Verify progress bar color | Green/normal when under limit |
| 7 | Verify "Add More Requests" button | Visible and enabled |
| 8 | Verify rate displayed | Shows requestsPerEGLD rate |
| 9 | Verify "Upgrade" section hidden | Free-user upgrade section NOT visible |

---

### 4.3 Flow: Unlimited Premium User Dashboard View

**ID:** FLOW-UI-003
**Priority:** Medium
**Tag:** [UI-ONLY]

**Preconditions:**
- User with `max_requests = 0` (unlimited)
- `account_type = 'premium'`

**Steps:**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login as `unlimited-user@test.com` | Dashboard loads |
| 2 | Verify "Premium Account" header | Shows with infinity symbol |
| 3 | Verify "Requests Used" counter | Shows total requests made |
| 4 | Verify "Limit: Unlimited" text | Displayed instead of progress bar |
| 5 | Verify NO progress bar | Progress bar hidden |
| 6 | Verify NO "Add More Requests" button | Button hidden |

---

### 4.4 Flow: Depleted Credits Warning

**ID:** FLOW-UI-004
**Priority:** High
**Tag:** [UI-ONLY]

**Preconditions:**
- User with `request_count >= max_requests`

**Steps:**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login as `depleted-user@test.com` | Dashboard loads |
| 2 | Verify progress bar at 100% | Bar completely filled |
| 3 | Verify warning message | "Limit reached. Add more requests to continue using the service." |
| 4 | Verify warning styling | Yellow/orange warning banner |
| 5 | Verify "Add More Requests" button | Prominently displayed |
| 6 | Attempt API call | Should be throttled (free tier behavior) |

---

### 4.5 Flow: Copy Address Functionality

**ID:** FLOW-UI-005
**Priority:** Medium
**Tag:** [UI-ONLY]

**Steps:**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | View payment details section | Address visible |
| 2 | Click "Copy Address" button | Visual feedback (button text changes or tooltip) |
| 3 | Paste into text editor | Correct erd1... address pasted |
| 4 | Verify no extra whitespace | Clean address copied |

---

### 4.6 Flow: Open Wallet External Link

**ID:** FLOW-UI-006
**Priority:** Low
**Tag:** [UI-ONLY]

**Steps:**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Click "Open Wallet" button | New browser tab opens |
| 2 | Verify URL | Opens configured walletURL (devnet-wallet.multiversx.com) |
| 3 | Verify link security | Opens with `target="_blank" rel="noopener"` |

---

### 4.7 Flow: Refresh Balance Button

**ID:** FLOW-UI-007
**Priority:** High
**Tag:** [IT-CANDIDATE]

**Preconditions:**
- User with payment address

**Steps:**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | View payment details | Current balance shown |
| 2 | Click "Refresh Balance" | Loading indicator appears |
| 3 | Wait for response | Balance updates (or stays same if unchanged) |
| 4 | Verify network request | GET `/api/crypto-payment/account` called |
| 5 | Rapid-click refresh multiple times | No errors, requests handled gracefully |

---

## 5. Service Status and Error Flows

### 5.1 Flow: Crypto Service Status Indicator

**ID:** FLOW-ERR-001
**Priority:** Medium
**Tag:** [IT-CANDIDATE]

**Scenario A: Service Online, Contract Active**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Both services running, contract not paused | Green indicator: "Online" |
| 2 | Contract status shows "Active" | Displayed correctly |

**Scenario B: Service Online, Contract Paused**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Pause the smart contract (admin action) | Contract paused |
| 2 | Refresh dashboard | Yellow indicator |
| 3 | Verify warning banner | "Service is paused - payments not being processed" |
| 4 | Verify "Request Payment Address" button | Disabled with tooltip explaining why |

**Scenario C: Service Unreachable**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Stop crypto-payment service | Service down |
| 2 | Refresh dashboard | Red indicator: "Offline" |
| 3 | Verify graceful degradation | Rest of dashboard functional |
| 4 | Verify crypto controls | All crypto-related buttons disabled |

---

### 5.2 Flow: API Error Handling - Service Unavailable

**ID:** FLOW-ERR-002
**Priority:** High
**Tag:** [IT-CANDIDATE]

**Preconditions:**
- Crypto-payment service stopped

**Steps:**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Stop crypto-payment service | Service down |
| 2 | Call `GET /api/crypto-payment/config` | Returns 503 with `isAvailable: false` |
| 3 | Call `POST /api/crypto-payment/create-address` | Returns 503 "Crypto payment service unavailable" |
| 4 | Call `GET /api/crypto-payment/account` | Returns 503 |
| 5 | Start crypto-payment service | Service up |
| 6 | Retry calls | All succeed |

---

### 5.3 Flow: Contract Paused State

**ID:** FLOW-ERR-003
**Priority:** High
**Tag:** [IT-CANDIDATE]

**Preconditions:**
- Smart contract in paused state

**Steps:**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Call `GET /api/crypto-payment/config` | Returns `isPaused: true` |
| 2 | UI shows paused warning | Banner visible |
| 3 | Balance processor logs | Shows "contract is paused, skipping" |
| 4 | Send eGLD to deposit address | Transaction confirms on chain |
| 5 | Wait for balance processor | NO relay transaction (balance stays) |
| 6 | Unpause contract | Contract active again |
| 7 | Wait for next cycle | Relay transaction processes backlog |

---

### 5.4 Flow: Invalid Payment ID Query

**ID:** FLOW-ERR-004
**Priority:** Medium
**Tag:** [IT-CANDIDATE]

**Steps:**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Call `GET /api/crypto-payment/account` with user who has no payment ID | Returns 404 "No payment ID associated with this account" |
| 2 | Call crypto-payment directly: `GET /account?id=999999` (non-existent) | Returns 500 (entry not found) |
| 3 | Call `GET /account?id=abc` (non-numeric) | Returns 400 (invalid parameter) |
| 4 | Call `GET /account` (missing id) | Returns 400 (missing parameter) |

---

### 5.5 Flow: Authentication Failures

**ID:** FLOW-ERR-005
**Priority:** Critical
**Tag:** [IT-CANDIDATE]

**Steps:**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Call crypto-payment `/create-address` without API key | Returns 401 |
| 2 | Call with wrong API key | Returns 401 |
| 3 | Call with empty API key header | Returns 401 |
| 4 | Call `/config` without API key | Returns 200 (public endpoint) |
| 5 | Call proxy endpoints without JWT | Returns 401 |
| 6 | Call proxy endpoints with expired JWT | Returns 401 |

---

## 6. Admin Panel Flows

### 6.1 Flow: Admin View User Payment Details

**ID:** FLOW-ADMIN-001
**Priority:** Medium
**Tag:** [IT-CANDIDATE]

**Preconditions:**
- Admin user logged in
- Target user has payment ID

**Steps:**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login as `admin@test.com` | Admin dashboard loads |
| 2 | Navigate to User Management | User list displayed |
| 3 | Locate user with payment ID | Payment ID column shows value |
| 4 | Click "View Payment Details" | Modal opens |
| 5 | Verify modal content | Shows username, payment ID, address, balance |
| 6 | Call API directly: `GET /api/admin-crypto-payment/account?username=premium-user@test.com` | Returns full payment details |

---

### 6.2 Flow: Admin View User Without Payment

**ID:** FLOW-ADMIN-002
**Priority:** Low
**Tag:** [IT-CANDIDATE]

**Steps:**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | View user management as admin | List displayed |
| 2 | Find user without payment ID | Payment ID column shows "-" or empty |
| 3 | "View Payment Details" action | Disabled or shows "No payment info" |

---

## 7. Background Sync Flows

### 7.1 Flow: Automatic Balance Sync

**ID:** FLOW-SYNC-001
**Priority:** High
**Tag:** [IT-CANDIDATE]

**Preconditions:**
- User has payment ID with credits on smart contract
- `users.max_requests` is lower than contract balance

**Steps:**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Set `users.max_requests = 1000` directly in DB | Value set |
| 2 | Verify contract has higher balance (e.g., 5000) | Contract shows 5000 |
| 3 | Wait for sync interval (default 5 minutes) | Sync job runs |
| 4 | Check database | `max_requests` updated to 5000 |
| 5 | Verify only increased | If DB was 10000 and contract 5000, DB stays 10000 |

---

### 7.2 Flow: Sync Respects Unlimited Accounts

**ID:** FLOW-SYNC-002
**Priority:** High
**Tag:** [IT-CANDIDATE]

**Preconditions:**
- User with `max_requests = 0` (unlimited, admin-granted)
- User has payment ID

**Steps:**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Verify `users.max_requests = 0` | Unlimited account |
| 2 | Add credits via blockchain payment | Contract balance increases |
| 3 | Wait for sync | Sync job runs |
| 4 | Verify `max_requests` still 0 | Unlimited status preserved |

---

## 8. Edge Cases and Stress Scenarios

### 8.1 Flow: Very Large Payment

**ID:** FLOW-EDGE-001
**Priority:** Medium
**Tag:** [E2E-ONLY]

**Steps:**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Send 10 eGLD to deposit address | Large transaction |
| 2 | Wait for relay | Transaction processes |
| 3 | Verify contract balance | `10 * requestsPerEGLD` credited |
| 4 | Verify no overflow | Large numbers handled correctly |
| 5 | UI displays correctly | Numbers formatted (e.g., "10,000,000") |

---

### 8.2 Flow: Multiple Small Payments Before Processing

**ID:** FLOW-EDGE-002
**Priority:** Medium
**Tag:** [E2E-ONLY]

**Preconditions:**
- Temporarily increase processing interval or pause processor

**Steps:**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Send 0.01 eGLD | Transaction 1 confirms |
| 2 | Send 0.02 eGLD (before processor runs) | Transaction 2 confirms |
| 3 | Send 0.03 eGLD (before processor runs) | Transaction 3 confirms |
| 4 | Allow processor to run | Single relay with total 0.06 eGLD |
| 5 | Verify balance | `0.06 * requestsPerEGLD` credited |

---

### 8.3 Flow: Service Restart During Processing

**ID:** FLOW-EDGE-003
**Priority:** Medium
**Tag:** [E2E-ONLY]

**Steps:**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Send eGLD to deposit address | Transaction confirms |
| 2 | Restart crypto-payment service immediately | Service restarts |
| 3 | Verify no duplicate processing | Balance credited once |
| 4 | Verify address persistence | Payment ID → address mapping intact |

---

### 8.4 Flow: Relayer Wallet Low Balance

**ID:** FLOW-EDGE-004
**Priority:** High
**Tag:** [E2E-ONLY]

**Preconditions:**
- Drain relayer wallet to very low balance

**Steps:**

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Reduce relayer wallet to < 0.001 eGLD | Insufficient for gas |
| 2 | Send eGLD to user deposit address | Transaction confirms |
| 3 | Wait for balance processor | Relay transaction fails |
| 4 | Check logs | Error logged: insufficient gas |
| 5 | User deposit NOT lost | Funds still on deposit address |
| 6 | Fund relayer wallet | Relayer has funds |
| 7 | Next processor cycle | Relay succeeds |

---

## 9. Integration Test Conversion Recommendations

The following manual test flows are recommended for conversion to automated integration tests:

### High Priority Conversions

| Flow ID | Name | Reason |
|---------|------|--------|
| FLOW-001 | First-Time Address Creation | Critical path, easily mockable |
| FLOW-002 | Duplicate Address Prevention | Business rule validation |
| FLOW-003 | Concurrent Address Request | Race condition detection |
| FLOW-ERR-002 | Service Unavailable Handling | Error path coverage |
| FLOW-ERR-004 | Invalid Payment ID Query | Input validation |
| FLOW-ERR-005 | Authentication Failures | Security critical |
| FLOW-SYNC-001 | Automatic Balance Sync | Background job validation |
| FLOW-SYNC-002 | Sync Respects Unlimited | Business rule protection |

### Integration Test Implementation Notes

```go
// Example test structure for FLOW-002
func TestDuplicateAddressCreationRejected(t *testing.T) {
    // Setup: User already has crypto_payment_id
    db := setupTestDB(t)
    db.SetUserPaymentID("test@example.com", 42)

    // Mock crypto-payment service (should NOT be called)
    mockCryptoService := &MockCryptoPaymentService{}
    mockCryptoService.On("CreateAddress").Times(0) // Assert never called

    // Execute
    handler := NewCryptoPaymentHandler(db, mockCryptoService)
    resp := handler.CreateAddress(userContext("test@example.com"))

    // Assert
    assert.Equal(t, 400, resp.StatusCode)
    assert.Contains(t, resp.Body, "already has a payment ID")
    mockCryptoService.AssertExpectations(t)
}
```

### Flows to Keep Manual

| Flow ID | Name | Reason |
|---------|------|--------|
| FLOW-004 | Complete Payment Cycle | Requires real blockchain |
| FLOW-005 | Minimum Balance Threshold | Requires real blockchain timing |
| FLOW-006 | Top-Up Existing Balance | Requires real blockchain |
| FLOW-EDGE-001 | Very Large Payment | Real value transfer testing |
| FLOW-EDGE-003 | Service Restart During Processing | Infrastructure testing |

---

## 10. Test Execution Checklist

### Pre-Test Checklist

- [ ] All services running and healthy
- [ ] Database migrated to latest schema
- [ ] Test accounts created
- [ ] Test wallet funded (>1 eGLD)
- [ ] Relayer wallets funded
- [ ] Smart contract not paused
- [ ] Logging enabled for debugging

### Post-Test Checklist

- [ ] All critical flows (FLOW-001 to FLOW-006) passed
- [ ] All error flows tested
- [ ] UI flows validated on target browsers
- [ ] Admin flows validated
- [ ] Database state verified
- [ ] No orphaned addresses in crypto-payment database
- [ ] Logs reviewed for unexpected errors

---

## 11. Defect Reporting Template

When reporting issues found during manual testing:

```
**Flow ID:** FLOW-XXX
**Title:** [Brief description]
**Severity:** Critical/High/Medium/Low

**Environment:**
- Branch: rc/v1.3
- Commit: cfd4817
- Browser: [if UI]

**Steps to Reproduce:**
1. ...
2. ...
3. ...

**Expected Result:**
[What should happen]

**Actual Result:**
[What actually happened]

**Evidence:**
- Screenshot: [attached]
- Logs: [attached]
- Database state: [query results]

**Notes:**
[Additional context]
```

---

**Document Prepared By:** System Architect
**Review Required By:** QA Lead, Development Lead
