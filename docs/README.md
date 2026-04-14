# MultiversX Epoch Proxy - Infrastructure & Dataflow

This document describes the hardware and software architecture of the Epoch Proxy infrastructure, designed to serve both live and historical (deep history) blockchain data efficiently.

## Overall Architecture

The system consists of a proxy layer that intelligently routes requests either to a live blockchain node (for recent data) or a deep-history node (for older epochs). It integrates a crypto payment service to handle access rights and limits.

### Architecture Diagram

![Architecture Diagram](diagram.drawio.png)

## Hardware Infrastructure

The physical infrastructure is hosted in a dedicated rack equipped with redundant power supplies and high-speed networking.

1. **Hypervisor Host (Compute)**: 
   - **Dell PowerEdge R640**: High-density 1U server acting as the main compute node.
   - Runs **Proxmox Virtual Environment**.
2. **Network Attached Storage (Data)**:
   - **UniFi UNAS PRO 4**: Provides high-capacity network storage.
   - Configured with a **24TB RAID1** drive array for redundancy.
   - Exposes Samba shares for accessing the massive historical blockchain data.
3. **Networking**: The rack is equipped with UniFi routing and switching equipment for high-throughput local connectivity.

## Virtual Machine Infrastructure

Inside the Proxmox environment (`pve-R640`), several targeted VMs are provisioned to handle distinct parts of the service workload.

### 1. `mvx-epoch-proxy`
- **Role**: Core ingress point and intelligent routing service.
- **Function**: Receives external requests from `mvx-deep-history.jls-software.net` and `admin-mvx-deep-history.jls-software.net`. Based on the requested epoch or data block, it forwards the request to either the live mainnet node or the deep history node.
- **Integration**: Communicates bi-directionally with the `mvx-crypto-payment-dh` VM to validate and record payment info.
- **VM specs**: 4 CPUs, 8GB RAM, 20GB disk.

### 2. `mvx-crypto-payment-dh`
- **Role**: Payment processor and manager.
- **Function**: Handles the crypto payment infrastructure, ensuring users have active access/balances to query the deep history infrastructure.
- **Dataflow**: Syncs payment info data directly with the Epoch Proxy VM.
- **VM specs**: 4 CPUs, 4GB RAM, 20GB disk.

### 3. `mvx-mainnet`
- **Role**: Live network node.
- **Function**: Runs a MultiversX Mainnet environment comprising **4 observers + proxy**.
- **State**: Kept fully in sync with the tip of the live blockchain.
- **Dataflow**: Responds to queries routed from the Epoch Proxy for recent/current epochs.
- **VM specs**: 20 CPUs, 64GB RAM, 12TB disk.

### 4. `mvx-dh1`
- **Role**: Archive node for historical epochs between 1 and 2050.
- **Function**: Runs 4 observers + proxy specifically configured for historical data archiving and retrieval. 
- **Storage**: Due to the massive storage footprint of historical epochs, this VM retrieves its state from the local UniFi UNAS appliance via a **Samba share**, preventing the primary VM SSDs from becoming saturated.
- **VM specs**: 16 CPUs, 64GB RAM, 400GB disk + 24TB NAS shared storage.

## Software solutions

### 1. `mvx-epoch-proxy`
- **URL**: https://github.com/iulianpascalau/mx-epoch-proxy-go
- **Language**: Go, Typescript, HTML, CSS, JavaScript, Shell
- **Purpose**: Main entry point, accessible from the internet. It manages the data access through tokens, integrates the crypto payment service and performs the data serving, intelligently routing the requests on the available and configured data nodes.
- **Local DB**: Yes, SQLite3.

### 2. `mvx-crypto-payment-dh`
- **URLs**: https://github.com/iulianpascalau/mx-crypto-payments-go, https://github.com/iulianpascalau/mx-credits-contract-rs
- **Language**: Go, Shell, Rust
- **Purpose**: Service that manages crypto payments providing access to dedicated accounts for each new user through private/public keys derived from a seed phrase. Handles contract interactions using the MultiversX relayed transactions v3.
- **Local DB**: Yes, SQLite3.

### 3. `mvx-mainnet`
- **URLs**: https://github.com/multiversx/mx-chain-go, https://github.com/multiversx/mx-chain-mainnet-config, https://github.com/multiversx/mx-chain-scripts
- **Language**: Go, Shell
- **Purpose**: Official MultiversX Mainnet node code, configuration files and scripts.
- **Local DB**: Yes, LevelDB.

### 4. `mvx-dh1`
- **URLs**: https://github.com/iulianpascalau/mx-chain-deep-history-go, https://github.com/multiversx/mx-chain-mainnet-config, https://github.com/iulianpascalau/mx-chain-deep-history-scripts
- **Language**: Go, Shell
- **Purpose**: Altered node code and scripts repositories, specifically designed for historical data retrieval. The node code is tuned so only the API engine is running, the p2p, blocks syncing, heartbeat and consensus go routines are not started to conserve resources.
- **Local DB**: Yes, LevelDB.

## Networking & Domains

The Epoch Proxy exposes two external endpoints to interact with the underlying virtual machines:

- **Client Access**: `mvx-deep-history.jls-software.net`
- **Administrative Access**: `admin-mvx-deep-history.jls-software.net`

Requests hitting these URLs are appropriately terminated by the `mvx-epoch-proxy` VM, which serves as the protective and routing proxy over the internal ecosystem.

---

## Solution development backlog

### 1. Initial solution as of November 2024:
The configuration consisted of: 
- 1 physical server running the official mainnet node (epochs starting from 1543 and onwards)
- 1 physical server running the deep history node version (epochs starting from 953 to 1543) 
- 1 physical server running the deep hisotry node version (epochs starting from 1 to 953)
- 1 VM running the Epoch Proxy initial application, version v1.0.1
The solution ran on 3 servers with SSD storage. The access was quite fast, but it was not a sustainable model due to the cost of adding one server for ~1 year worth of data. The system was mostly used internally by the MultiversX team as it lacked access management.

### 2. v1.0.x && v1.1.x versions (2025):
The hardware configuration remained basically the same, only the epoch proxy software was changed to accommodate manually added whitelisted tokens. The access was given to users outside the MultiversX team.
During the 2025 setup monitoring, the disk space usage became the main concern in the current setup. During 2 meetings with the MultiversX team members, it was decided to move away from server-hosted-storage towards a more sustainable solution using network attached storage. The decision was to use an Unifi UNAS Pro 4 device which was unavailable at that time. The hardware solution was postponed until the UNAS became available.

### 3. v1.2.x versions (January 2026)
While waiting for the hardware availability, the epoch proxy solution was re-evaluated and the administrative panel was implemented using some AI tools. Frontend, DB interaction and infrastructure settings were carried out.

### 4. v1.3.x versions (January 2026)
The work started to implement the crypto payment service as a mean to migrate from the free tier to the premium one, automatically. 

### 5. v1.4.x versions (February 2026)
To increase the reusability of the solution, the code was heavily refactored, essentially splitting the epoch-proxy, crypto-payment and the credits-contract solution in dedicated repositories.

### 6. Hardware refactoring (February – April 2026)
With the new hardware available, the time came to assess the biggest problem of the entire solution: disk usage. The data was slowly, sometimes painfully slowly copied to the UNAS hardware. After the required data was copied and filled up almost the entire UNAS 24TB drive the following changes took place:
- the physical servers hosting historical data were turned off and repurposed
- the physical server running the official mainnet (R640) nodes was reinstalled and configured to host the Proxmox hypervisor
- on this server 2 VMs were added: one for the official mainnet node and one for the deep history node
- the VMs for the crypto payment service and epoch proxy were moved to this server.
