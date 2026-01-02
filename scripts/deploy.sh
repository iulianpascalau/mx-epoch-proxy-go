#!/bin/bash
set -e

# Configuration
PROJECT_DIR="/home/ubuntu/app"
BACKEND_SERVICE="epoch-proxy-server"
FRONTEND_SERVICE="epoch-proxy-frontend"

# Check argument
BRANCH=$1
if [ -z "$BRANCH" ]; then
    echo "Usage: $0 <branch_or_tag>"
    echo "Example: $0 main"
    exit 1
fi

echo "=========================================="
echo "Starting deployment for branch: $BRANCH"
echo "=========================================="

# Navigate to project directory
if [ ! -d "$PROJECT_DIR" ]; then
    echo "Error: Project directory $PROJECT_DIR does not exist."
    exit 1
fi
cd "$PROJECT_DIR"

# 1. Stop Services
echo "Step 1: Stopping services..."
sudo systemctl stop $FRONTEND_SERVICE $BACKEND_SERVICE

# 2. Checkout Code
echo "Step 2: Checking out code..."
git fetch --all
git checkout "$BRANCH"
git pull origin "$BRANCH"

# 3. Recompile Backend
echo "Step 3: Recompiling Backend..."
# Adjust go path if necessary, assuming it is in path or /usr/local/go/bin/go
if command -v go &> /dev/null; then
    GO_CMD="go"
elif [ -f "/usr/local/go/bin/go" ]; then
    GO_CMD="/usr/local/go/bin/go"
else
    echo "Error: Go binary not found."
    exit 1
fi

$GO_CMD build -v -ldflags="-X main.appVersion=$(git describe --tags --long --dirty)" -o server main.go
if [ $? -ne 0 ]; then
    echo "Backend build failed!"
    exit 1
fi
echo "Backend build successful."

# 4. Update Frontend
echo "Step 4: Updating Frontend..."
cd frontend
# Install dependencies
npm install
# Note: The service currently runs 'npm run dev', so we don't 'build' for production serving yet.
# If you switch to 'npm run build', add it here.
cd ..

# 5. Restart Services
echo "Step 5: Restarting services..."
sudo systemctl start $BACKEND_SERVICE $FRONTEND_SERVICE

# 6. Monitor
echo "Step 6: Monitoring status..."
sleep 5

if systemctl is-active --quiet $BACKEND_SERVICE; then
    echo "✅ $BACKEND_SERVICE is active."
else
    echo "❌ $BACKEND_SERVICE failed to start."
    sudo journalctl -u $BACKEND_SERVICE -n 20 --no-pager
    exit 1
fi

if systemctl is-active --quiet $FRONTEND_SERVICE; then
    echo "✅ $FRONTEND_SERVICE is active."
else
    echo "❌ $FRONTEND_SERVICE failed to start."
    sudo journalctl -u $FRONTEND_SERVICE -n 20 --no-pager
    exit 1
fi

echo "=========================================="
echo "Deployment Finished Successfully!"
echo "=========================================="
