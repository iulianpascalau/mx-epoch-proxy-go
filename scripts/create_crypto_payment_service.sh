#!/bin/bash

# Configuration
USER_NAME="ubuntu"
APP_NAME="crypto-payment-service"
APP_DIR="/home/${USER_NAME}/epoch-proxy/services/crypto-payment"
EXEC_PATH="${APP_DIR}/crypto-payment-server"

# Create the service file content
SERVICE_CONTENT="[Unit]
Description=Mvx Crypto Payment Service Go Backend
After=network-online.target

[Service]
User=${USER_NAME}
WorkingDirectory=${APP_DIR}
ExecStart=${EXEC_PATH} -log-save -log-level *:DEBUG
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
"

# Path to the systemd service file
SERVICE_FILE="/etc/systemd/system/${APP_NAME}.service"

# Write the service file
echo "Creating systemd service file at ${SERVICE_FILE}..."
sudo bash -c "echo '${SERVICE_CONTENT}' > ${SERVICE_FILE}"

# Reload systemd daemon
echo "Reloading systemd daemon..."
sudo systemctl daemon-reload

# Enable the service
echo "Enabling ${APP_NAME} service..."
sudo systemctl enable ${APP_NAME}

# Start the service
echo "Starting ${APP_NAME} service..."
sudo systemctl start ${APP_NAME}

# Show status
echo "Service status:"
sudo systemctl status ${APP_NAME} --no-pager
