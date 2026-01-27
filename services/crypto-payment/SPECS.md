# Crypto Payment Service - Feature Specification

## 1. Overview
This feature allows users of the **Mx Epoch Proxy** to upgrade their accounts from "Free" to "Premium" by making a cryptocurrency payment in eGLD on the MultiversX blockchain.

The system utilizes a secure, segregated microservice (`crypto-payment`) to manage deposit addresses and relay funds to a licensing Smart Contract. This ensures that the main Proxy service never handles private keys and that the user experience is seamless (no gas required for the user's deposit transaction).

## 2. Architecture

The solution consists of three main components interacting asynchronously:

1.  **Epoch Proxy Service** (Existing): Handles the user UI and final entitlement (premium status).
2.  **Crypto Payment Service** (New Microservice):
    -   **Responsibility**: Secure key management, address generation, blockchain monitoring, and transaction relaying.
    -   **Security**: Runs in a restricted environment (separate VM); holds the master mnemonic.
3.  **Licensing Smart Contract**:
    -   **Responsibility**: Receives funds, calculates entitlement (e.g., 1 eGLD = 1M requests), and stores the "Allowed Requests" quota for each Payment ID.

## 3. Detailed Data Flow

### Phase 1: Initiation
1.  **User Action**: User clicks "Upgrade to Premium" on the Proxy Dashboard.
2.  **Proxy Request**: The Proxy Backend requests a new `Payment ID` and `Address` from the internal `crypto-payment` API.
3.  **Address Generation**:
    -   The `crypto-payment` service increments its internal index counter.
    -   It derives a new public key/address from its master mnemonic using the HD path (e.g., `m/44'/508'/0'/0/index`).
    -   It stores the `{index, address}` tuple in its local database.
    -   It returns the `index` (Payment ID) to the Proxy.
4.  **Display**:
    -   The Proxy links this `Payment ID` to the User in its database.
    -   The Proxy calls `GET /account?id=<id>` to retrieve the address to display to the User.
    -   The Proxy displays the `address` and `numberOfRequests` to the User.

### Phase 2: Payment & Monitoring
5.  **User Payment**: The user sends eGLD to the displayed unique address.
6.  **Monitoring**: The `crypto-payment` service monitors the blockchain (via Observer Nodes) for any incoming transactions to its known generated addresses.

### Phase 3: Sweeping (Relayed Transaction)
7.  **Detection**: When a balance > 0 is detected on a generated address:
8.  **Construction**: The `crypto-payment` service constructs a **Relayed Transaction v3**:
    -   **Inner Transaction**:
        -   **Sender**: The generated deposit address.
        -   **Receiver**: The Licensing Smart Contract address.
        -   **Value**: The entire balance of the deposit address.
        -   **Data**: `buy_credits` endpoint call, passing the `Payment ID` as an argument (e.g., `buy_credits@<payment_id_hex>`).
        -   **Gas Limit**: Sufficient for SC execution.
        -   **Signature**: Signed by the derived private key of the deposit address.
    -   **Relayer**:
        -   **Sender**: The `crypto-payment` service's "Hot Wallet" (Gas Payer).
        -   **Signature**: Signed by the Hot Wallet's private key.
9.  **Submission**: The transaction is broadcast to the network. The Hot Wallet pays the gas fees; the User's deposit is forwarded intact.

### Phase 4: Entitlement & Sync
10. **Smart Contract**: Executes `buy_credits`. It calculates the credit amount based on the `Value` transferred and updates its storage: `storage[payment_id] = new_total_credits`.
11. **Synchronization**:
    -   *Option A (Polling)*: The Proxy Service periodically queries the Smart Contract for the credit balance of its users' `Payment IDs`.
    -   *Option B (Trigger)*: The `crypto-payment` service notifies the Proxy Service via web-hook upon successful relay.
12. **Update**: When the Proxy Service detects an increase in credits on the SC, it updates the local `users` table:
    -   `max_requests` updated to the new limit.
    -   `account_type` set to `premium`.

## 4. API Endpoints

### 4.1. `POST /create-address`
Generates a new deposit address.

**Request:** Empty body.

**Response:**
```json
{
  "id": 123
}
```

### 4.2. `GET /account`
Retrieves account details for a payment ID.

**Query Parameters:**
- `id` (required): The Payment ID (uint64).

**Response:**
```json
{
  "address": "erd1...",
  "numberOfRequests": 1500000
}
```

### 4.3. `GET /config`
Retrieves public configuration required by the client/proxy.

**Response:**
```json
{
  "walletUrl": "https://devnet-wallet.multiversx.com",
  "explorerUrl": "https://devnet-explorer.multiversx.com",
  "contractAddress": "erd1...",
  "requestsPerEGLD": 1000000,
  "isContractPaused": false
}
```

## 5. Data Models

### Proxy Database (`sqlite.db`)
Changes to `users` table:
-   `payment_id` (Integer, Nullable): Links the user to the crypto-payment system.

### Crypto Payment Database (`crypto-payment.db`)
New SQLite database for the microservice.
**Table: `payment_address`** (Created by `sqliteWrapper`)
-   `id` (Integer, Primary Key): The derivation index (Payment ID).
-   `address` (Text, Unique): The derived Bech32 address.

## 6. Implementation Details
-   **Balance Processor**: A background process (`BalanceProcessor`) periodically scans all generated addresses for balances.
-   **Relayed Transactions**: If a balance is found, a relayed transaction is constructed and sent to the network to sweep funds to the contract.
-   **Contract Query**: The service caches (`timeCacher`) contract query results (`isPaused`, `requestsPerEgld`) to ensure performance and reduce network load.
-   **Caching**: A simple in-memory cache with a global flush strategy is used to store high-frequency data (like contract config values).

## 7. Security Considerations
-   **Isolation**: The `crypto-payment` service contains the "Crown Jewels" (the Mnemonic). It must be isolated.
-   **Relayer Wallet**: The Hot Wallet key is also sensitive as it holds gas funds.
-   **Input Validation**: The Proxy must validate that the `crypto-payment` service is the only one calling its internal webhooks (if used).

## 8. Configuration

### Crypto Payment Service
-   `MNEMONIC` (Env variable): 24-word secret phrase for user address derivation.
-   `config.toml`:
    -   `Port`: API Port (default 8080).
    -   `WalletURL`: URL of the Web Wallet (e.g., `devnet-wallet.multiversx.com`).
    -   `ExplorerURL`: URL of the Explorer (e.g., `devnet-explorer.multiversx.com`).
    -   `ProxyURL`: URL of the MultiversX Proxy/Gateway.
    -   `ContractAddress`: Address of the Licensing Smart Contract.
    -   `CallSCGasLimit`: Gas limit for the relayed transaction (default 40,000,000).
    -   `SCSettingsCacheInSeconds`: Cache TTL for contract queries (e.g., 60s).
    -   `MinimumBalanceToProcess`: Minimum eGLD balance to trigger a sweep (e.g., 0.01).
    -   `TimeToProcessAddressesInSeconds`: Interval for the Balance Processor cron job.
-   `*.pem` files in working directory: Loaded as Relayer Sk(s) for paying gas.
