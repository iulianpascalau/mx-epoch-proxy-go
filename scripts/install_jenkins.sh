#!/bin/bash
set -e

# Jenkins Installation Script for Ubuntu/Debian
# Usage: sudo ./install_jenkins.sh

if [ "$EUID" -ne 0 ]; then 
  echo "Please run as root (sudo)"
  exit 1
fi

echo "Step 1: Installing Java (required for Jenkins)..."
apt-get update
apt-get install -y fontconfig openjdk-17-jre

echo "Step 2: Adding Jenkins Repository..."
# Download key
wget -O /usr/share/keyrings/jenkins-keyring.asc \
  https://pkg.jenkins.io/debian-stable/jenkins.io-2023.key

# Add entry to sources.list.d
echo "deb [signed-by=/usr/share/keyrings/jenkins-keyring.asc]" \
  "https://pkg.jenkins.io/debian-stable binary/" | \
  tee /etc/apt/sources.list.d/jenkins.list > /dev/null

echo "Step 3: Installing Jenkins..."
apt-get update
apt-get install -y jenkins

echo "Step 4: Starting Jenkins Service..."
systemctl enable jenkins
systemctl start jenkins

echo "Step 5: Configuring Firewall (Port 8080)..."
if command -v ufw &> /dev/null; then
    ufw allow 8080
    echo "Port 8080 allowed through UFW."
else
    echo "UFW not found, ensure port 8080 is open in your cloud security groups."
fi

# Wait a moment for Jenkins to create the initial password file
sleep 5

echo "========================================================"
echo "Jenkins Installation Complete!"
echo "========================================================"
echo "Access Jenkins at: http://YOUR_VM_IP:8080"
echo ""
echo "Your Initial Admin Password is:"
if [ -f /var/lib/jenkins/secrets/initialAdminPassword ]; then
    cat /var/lib/jenkins/secrets/initialAdminPassword
else
    echo "Password file not found yet. It may take a few seconds."
    echo "Run: sudo cat /var/lib/jenkins/secrets/initialAdminPassword"
fi
echo "========================================================"
