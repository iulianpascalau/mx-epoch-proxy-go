#!/bin/bash
set -e

# Configuration
PROJECT_DIR="/home/ubuntu/epoch-proxy"
SERVICE_NAME="crypto-payment-service"

# Check argument
BRANCH=$1
if [ -z "$BRANCH" ]; then
    echo "Usage: $0 <branch_or_tag>"
    echo "Example: $0 main"
    exit 1
fi

echo "=========================================="
echo "Starting Crypto Payment deployment for branch: $BRANCH"
echo "=========================================="

# Navigate to project directory
if [ ! -d "$PROJECT_DIR" ]; then
    echo "Error: Project directory $PROJECT_DIR does not exist."
    exit 1
fi
cd "$PROJECT_DIR"

# 1. Stop Service
echo "Step 1: Stopping service..."
sudo systemctl stop $SERVICE_NAME || echo "Service $SERVICE_NAME not found or not running, skipping stop."

# 2. Checkout Code
echo "Step 2: Checking out code..."
git fetch --all
git checkout "$BRANCH"
git pull origin "$BRANCH"

# 3. Recompile Backend
echo "Step 3: Recompiling Crypto Payment Service..."
# Load common functions
source ./scripts/common.sh

# Ensure Go is installed
ensure_go_installed
GO_CMD="go"

cd ./services/crypto-payment
$GO_CMD build -v -ldflags="-X main.appVersion=$(git describe --tags --long --dirty)" -o crypto-payment-server main.go
if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi
echo "Build successful."

# 4. Restart Service
echo "Step 4: Restarting service..."
sudo systemctl start $SERVICE_NAME

# 5. Monitor
echo "Step 5: Monitoring status..."
sleep 5

if systemctl is-active --quiet $SERVICE_NAME; then
    echo "✅ $SERVICE_NAME is active."
else
    echo "❌ $SERVICE_NAME failed to start."
    sudo journalctl -u $SERVICE_NAME -n 20 --no-pager
    exit 1
fi

echo "=========================================="
echo "Deployment Finished Successfully!"
echo "=========================================="
