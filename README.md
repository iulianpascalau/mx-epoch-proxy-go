# mx-epoch-proxy
Epoch based proxy forwarder

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

