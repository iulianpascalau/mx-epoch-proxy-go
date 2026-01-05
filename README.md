# mx-epoch-proxy
Epoch based proxy forwarder

## Installation notes

On the target VM the following steps should be completed:

### 1. Prerequisites on the VM
```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install prerequisites packages so the go build will succeed
sudo apt install -y git curl zip rsync jq gcc wget
sudo apt install -y build-essential

# Install Go (Adjust version if needed, your go.mod says 1.24 so you need a very recent version)
GO_LATEST_TESTED="1.24.11"
ARCH=$(dpkg --print-architecture)
wget https://dl.google.com/go/go${GO_LATEST_TESTED}.linux-${ARCH}.tar.gz
sudo tar -C /usr/local -xzf go${GO_LATEST_TESTED}.linux-${ARCH}.tar.gz
rm go${GO_LATEST_TESTED}.linux-${ARCH}.tar.gz

echo "export GOPATH=$HOME/go" >> ~/.profile
echo "export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin" >> ~/.profile
echo "export GOPATH=$HOME/go" >> ~/.profile
source ~/.profile
go version

# Install Node.js (for building frontend)
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs
```

### 2. Clone & Build the Application
```bash
# Clone the repository (replace with your actual repo URL)
cd ~
git clone https://github.com/iulianpascalau/mx-epoch-proxy-go.git epoch-proxy
cd epoch-proxy

# Ensure you are on the main branch (after you merge your PR)
git checkout main
git pull origin main

# --- Build Backend ---
# Create a binary named 'epoch-proxy-server'
cd ./services/proxy
go build -v -ldflags="-X main.appVersion=$(git describe --tags --long --dirty)" -o epoch-proxy-server main.go

# --- Build Frontend ---
cd frontend
npm install
npm run build
# This creates a 'dist' folder with your static site
```

## Configuration steps:

### 1. Cloudflare setup for deep-history subdomain

#### 1.1. Installing cloudflared

```bash
cd
sudo mkdir -p --mode=0755 /usr/share/keyrings
curl -fsSL https://pkg.cloudflare.com/cloudflare-main.gpg | sudo tee /usr/share/keyrings/cloudflare-main.gpg >/dev/null
echo 'deb [signed-by=/usr/share/keyrings/cloudflare-main.gpg] https://pkg.cloudflare.com/cloudflared jammy main' | sudo tee /etc/apt/sources.list.d/cloudflared.list
sudo apt-get update && sudo apt-get install cloudflared
```

#### 1.2. Authenticate
```bash
cloudflared tunnel login
```
Follow the URL to authorize your domain.

#### 1.3. Create Tunnel
```bash
cloudflared tunnel create mvx-deep-history
# Note the UUID and credentials path outputted.
```

#### 1.4. Configure Tunnel

```bash
# Create the system directory
sudo mkdir -p /etc/cloudflared

# Copy your credentials JSON file
sudo cp ~/.cloudflared/*.json /etc/cloudflared/

# Create/edit the config file 
sudo nano /etc/cloudflared/config.yml
```
Suppose the domain is set to xxx.yyy.zzz

```yaml
tunnel: <YOUR_TUNNEL_UUID>
credentials-file: /etc/cloudflared/<YOUR_TUNNEL_UUID>.json

ingress:
  # Route all traffic to the proxy on port 8080
  - hostname: xxx.yyy.zzz
    service: http://localhost:8080
  
  # Catch-all
  - service: http_status:404
```

#### 1.5. DNS Routing
```bash
cloudflared tunnel route dns mvx-deep-history xxx.yyy.zzz
```

#### 5.6. Start Tunnel Service
```bash
sudo cloudflared service install
sudo systemctl start cloudflared
```

