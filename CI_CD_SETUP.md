# CI/CD Setup Guide

## Important

For this project setup, the active deployment path is a Windows self-hosted runner on your own computer. The old VM/SSH instructions below are legacy notes from the earlier version of the workflow.

## Overview

This document describes the automated CI/CD pipeline and deployment process for the Kafka Order Service application.

## Pipeline Architecture

### 1. **Build Stage** (Automatic on every push)
- Triggered on push to `main` branch
- Builds Docker image
- Runs container health checks
- Validates successful startup

### 2. **Deploy Stage** (After successful build)
- Only runs on successful build
- Connects to VM via SSH
- Pulls latest code from git
- Rebuilds and restarts Docker containers
- Verifies deployment status

## GitHub Actions Configuration

### Workflow File
- Location: `.github/workflows/ci-cd.yml`
- Triggers: Push to main branch, Pull requests to main

### Jobs

#### Build Job
```yaml
- Builds Docker image: kafka-app:{SHA} and kafka-app:latest
- Starts docker-compose services
- Performs health checks
- Cleans up test environment
```

#### Deploy Job
```yaml
- Runs only on main branch pushes (not on PRs)
- Depends on successful build
- SSH connection to VM
- Git pull and docker-compose rebuild
```

## Required GitHub Secrets

You need to configure these secrets in your GitHub repository settings:

### **VM_HOST**
- VM's public IP address or hostname
- Example: `203.0.113.42` or `deploy.example.com`

### **VM_USER**
- SSH username on the VM
- Example: `ubuntu` or `deploy`

### **VM_SSH_KEY**
- **IMPORTANT**: Private SSH key for authentication
- Must be in OpenSSH format (starts with `-----BEGIN OPENSSH PRIVATE KEY-----`)
- Do NOT include passphrase in the key
- Example format:
  ```
  -----BEGIN OPENSSH PRIVATE KEY-----
  b3BlbnNzaC1rZXktdjEAAAAABG5vbmUtbm9uZWQtYWVzMjU2AAAA...
  -----END OPENSSH PRIVATE KEY-----
  ```

### **PROJECT_PATH**
- Full path to project directory on VM
- Example: `/home/ubuntu/Kafka` or `/opt/projects/kafka`

### **VM_PORT** (Optional)
- SSH port on VM (default: 22)
- Example: `2222`

## Setup Instructions

### Step 1: Generate SSH Key (if not already done)

```bash
# On your local machine
ssh-keygen -t ed25519 -f kafka-deploy-key -N ""

# Or RSA (for older systems):
ssh-keygen -t rsa -b 4096 -f kafka-deploy-key -N ""
```

### Step 2: Add Public Key to VM

```bash
# Copy public key to VM
cat kafka-deploy-key.pub | ssh user@vm-host "mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys"

# Or manually on the VM:
cat >> ~/.ssh/authorized_keys < kafka-deploy-key.pub
```

### Step 3: Configure GitHub Secrets

1. Go to GitHub repository → Settings → Secrets and variables → Actions
2. Click "New repository secret" for each secret:
   - `VM_HOST`: Your VM's IP or hostname
   - `VM_USER`: SSH username
   - `VM_SSH_KEY`: Contents of `kafka-deploy-key` (private key)
   - `PROJECT_PATH`: Path to project on VM (e.g., `/home/ubuntu/Kafka`)
   - `VM_PORT`: SSH port (optional, defaults to 22)

### Step 4: Verify VM Setup

Ensure on your VM:
- [ ] Git is installed: `git --version`
- [ ] Docker is installed: `docker --version`
- [ ] Docker Compose is installed: `docker-compose --version`
- [ ] Project is cloned: `git clone <repo-url>` at PROJECT_PATH
- [ ] SSH key is in authorized_keys
- [ ] User can run Docker without sudo: `sudo usermod -aG docker $USER`

## Deployment Flow

```
Push to main
    ↓
GitHub Actions triggered
    ↓
Build Job:
  - Build Docker image
  - Run health checks
    ↓
Build Successful?
    ├─ YES → Deploy Job:
    │         - SSH to VM
    │         - git pull
    │         - docker compose up -d --build
    │         - Verify deployment
    │
    └─ NO → Pipeline fails
```

## Testing the Pipeline

### Test 1: Verify Build Pipeline
1. Make a small change to code (e.g., in a comment)
2. Commit and push to main: `git push origin main`
3. Go to GitHub repository → Actions
4. Watch the workflow run
5. Verify "Build & Test Docker Image" job succeeds

### Test 2: Verify Deploy Pipeline
1. Same as Test 1
2. Verify "Deploy to VM" job completes
3. SSH to VM and check: `docker compose ps`
4. Verify all services are running

### Test 3: End-to-End Test with Visible Changes
1. Modify application code (e.g., response message)
2. Commit with message: `git commit -m "test: visible change for CI/CD validation"`
3. Push: `git push origin main`
4. Monitor Actions tab until deployment completes
5. SSH to VM and verify the change is live:
   ```bash
   curl http://localhost:8000/health
   # Check if new code is running
   docker logs order-api | tail -20
   ```

## Troubleshooting

### Build Fails
- Check Docker build logs in Actions tab
- Ensure Dockerfile is valid: `docker build -t test .`
- Check docker-compose.yml syntax: `docker-compose config`

### Deploy Fails
- **SSH Connection Error**: 
  - Verify VM_HOST, VM_USER, VM_PORT are correct
  - Check SSH key is in authorized_keys: `cat ~/.ssh/authorized_keys`
  - Test manually: `ssh -i key.pem user@host`

- **Git Pull Error**:
  - Verify PROJECT_PATH exists and is a git repository
  - Check user has permission to read/write directory
  - Verify origin remote: `git remote -v`

- **Docker Command Error**:
  - Check user is in docker group: `groups $USER`
  - May need: `sudo usermod -aG docker $USER`
  - Restart Docker daemon if needed

### Services Not Starting
- Check logs: `docker compose logs`
- Verify docker-compose.yml is valid
- Check port availability: `lsof -i :8000` (example)

## Environment Variables

If your application uses environment variables, add them to `docker-compose.yml`:

```yaml
services:
  order-api:
    environment:
      - LOG_LEVEL=info
      - KAFKA_BROKER=kafka:29092
```

Or use `.env` file in project directory:

```bash
echo "LOG_LEVEL=info" > /path/to/project/.env
```

## Monitoring Deployments

### GitHub Actions UI
- Repository → Actions tab
- Click on workflow run to see detailed logs
- Each step shows execution time and status

### VM Monitoring
```bash
# Check running containers
docker compose ps

# View logs
docker compose logs -f order-api

# Check resource usage
docker stats
```

## Security Best Practices

1. **SSH Key**:
   - Use ed25519 keys (stronger than RSA)
   - Never commit private keys to git
   - Rotate keys periodically

2. **Secrets**:
   - Don't store in code or docker-compose.yml
   - Use GitHub Secrets for sensitive data
   - Rotate credentials regularly

3. **Access Control**:
   - Restrict deploy user permissions on VM
   - Use specific SSH key for CI/CD only
   - Monitor Actions logs for suspicious activity

## Performance Optimization

### Caching Dependencies
Add to workflow to speed up builds:
```yaml
- name: Cache Docker layers
  uses: actions/cache@v3
  with:
    path: /tmp/.buildx-cache
    key: ${{ runner.os }}-buildx-${{ github.sha }}
```

### Parallel Jobs
Current setup runs build first, then deploy serially. Can be optimized based on your CI capacity.

## Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [appleboy/ssh-action](https://github.com/appleboy/ssh-action)
- [Docker Build Action](https://docs.github.com/en/actions/publishing-packages/publishing-docker-images)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
