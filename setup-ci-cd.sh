#!/bin/bash

# CI/CD Setup Helper Script
# This script helps you set up the SSH key and verify prerequisites

set -e

echo "================================"
echo "Kafka CI/CD Setup Helper"
echo "================================"
echo ""

# Check prerequisites
echo "🔍 Checking prerequisites..."
echo ""

# Check Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker not found. Please install Docker first."
    exit 1
fi
echo "✅ Docker: $(docker --version)"

# Check Docker Compose
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose not found. Please install Docker Compose."
    exit 1
fi
echo "✅ Docker Compose installed"

# Check Git
if ! command -v git &> /dev/null; then
    echo "❌ Git not found. Please install Git first."
    exit 1
fi
echo "✅ Git: $(git --version)"

echo ""
echo "================================"
echo "SSH Key Setup"
echo "================================"
echo ""

# Check for existing key
KEY_PATH="${HOME}/.ssh/kafka-deploy-key"
if [ -f "$KEY_PATH" ]; then
    echo "⚠️  SSH key already exists at $KEY_PATH"
    read -p "Generate new key? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Using existing key"
    else
        rm "$KEY_PATH" "$KEY_PATH.pub"
        echo "✅ Removed old key"
    fi
fi

# Generate key if needed
if [ ! -f "$KEY_PATH" ]; then
    echo "🔑 Generating SSH key..."
    mkdir -p ~/.ssh
    ssh-keygen -t ed25519 -f "$KEY_PATH" -N "" -C "kafka-deploy-$(date +%s)"
    chmod 600 "$KEY_PATH"
    chmod 644 "$KEY_PATH.pub"
    echo "✅ SSH key generated at $KEY_PATH"
fi

echo ""
echo "================================"
echo "Next Steps"
echo "================================"
echo ""
echo "1️⃣  Copy your private key to GitHub Secrets:"
echo "   cat ~/.ssh/kafka-deploy-key"
echo "   Then go to:"
echo "   Repository Settings → Secrets and variables → Actions"
echo "   Create secret: VM_SSH_KEY (paste the entire key)"
echo ""
echo "2️⃣  Add public key to VM:"
echo "   ssh-copy-id -i ~/.ssh/kafka-deploy-key.pub user@vm-ip"
echo ""
echo "   OR manually:"
echo "   cat ~/.ssh/kafka-deploy-key.pub | ssh user@vm-ip 'mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys'"
echo ""
echo "3️⃣  Test SSH connection:"
echo "   ssh -i ~/.ssh/kafka-deploy-key user@vm-ip docker ps"
echo ""
echo "4️⃣  Set GitHub Secrets:"
echo "   - VM_HOST: Your VM IP or hostname"
echo "   - VM_USER: Your SSH username"
echo "   - VM_SSH_KEY: Your private key (from step 1)"
echo "   - PROJECT_PATH: /path/to/Kafka/on/vm"
echo "   - VM_PORT: 22 (or your custom SSH port)"
echo ""
echo "5️⃣  Test the pipeline:"
echo "   git add . && git commit -m 'setup: CI/CD ready' && git push origin main"
echo ""
echo "✅ Setup helper completed!"
