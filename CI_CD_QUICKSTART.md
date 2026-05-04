# Quick CI/CD Setup Checklist

## ✅ Prerequisite Tasks

### On Your Local Machine
- [ ] Clone the repository locally
- [ ] Verify project builds: `docker build -t test .`
- [ ] Verify docker-compose works: `docker-compose config`
- [ ] Push to your fork on GitHub

### On Your VM (Target Deployment Server)
- [ ] SSH access working
- [ ] Docker installed and running
- [ ] Docker Compose installed (v2.0+)
- [ ] Git installed
- [ ] Project cloned to VM at specific path
- [ ] User added to docker group: `sudo usermod -aG docker $USER`
- [ ] Restart shell to apply group changes: `newgrp docker`

## 🔑 SSH Key Setup

### Generate Key (one-time)
```bash
# On your local machine (if you don't have a key)
ssh-keygen -t ed25519 -f ~/.ssh/kafka-deploy-key -N ""
# or for older systems:
ssh-keygen -t rsa -b 4096 -f ~/.ssh/kafka-deploy-key -N ""
```

### Add to VM
```bash
# Copy key to VM
ssh-copy-id -i ~/.ssh/kafka-deploy-key.pub user@vm-ip
# or manually:
cat ~/.ssh/kafka-deploy-key.pub | ssh user@vm-ip "mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys"
```

### Test SSH Connection
```bash
ssh -i ~/.ssh/kafka-deploy-key user@vm-ip "docker ps"
# Should list Docker containers without password prompt
```

## 🔐 GitHub Secrets Configuration

Go to: **Repository Settings → Secrets and variables → Actions → New repository secret**

| Secret Name | Value | Example |
|------------|-------|---------|
| `VM_HOST` | VM IP or hostname | `203.0.113.42` |
| `VM_USER` | SSH username | `ubuntu` |
| `VM_SSH_KEY` | Private key contents | `-----BEGIN OPENSSH...` |
| `PROJECT_PATH` | Full path to project | `/home/ubuntu/Kafka` |
| `VM_PORT` | SSH port (optional) | `22` |

### Getting Private Key Contents
```bash
# Print key to copy (make sure no passphrase!)
cat ~/.ssh/kafka-deploy-key
# Then copy the entire output from "-----BEGIN" to "-----END"
```

## 🚀 Testing the Pipeline

### Test 1: Trigger Build Pipeline
```bash
# Make a change and push
echo "# Testing CI/CD" >> README.md
git add README.md
git commit -m "test: trigger CI/CD pipeline"
git push origin main
```

### Test 2: Monitor Execution
- Go to GitHub → Actions tab
- Click on the running workflow
- Watch the build and deploy jobs

### Test 3: Verify on VM
```bash
# SSH to VM
ssh user@vm-ip

# Check running services
cd /path/to/project
docker compose ps

# View logs
docker compose logs -f order-api
```

### Test 4: End-to-End Verification
1. Make visible code change (e.g., API response)
2. Push to main
3. Wait for pipeline completion
4. SSH to VM and verify change is live

## 📋 Expected Workflow Behavior

### On Successful Push to Main
```
✓ Build & Test Docker Image (5-10 min)
  ├─ Checkout code
  ├─ Build Docker image  
  ├─ Test with docker-compose
  ├─ Health checks
  └─ Cleanup

✓ Deploy to VM (2-5 min)
  ├─ SSH to VM
  ├─ Git pull
  ├─ Docker compose rebuild
  ├─ Verify services
  └─ Done!
```

## ❌ Troubleshooting

### Pipeline doesn't trigger
- [ ] Pushed to `main` branch (not `master`)
- [ ] Commit is in GitHub repository
- [ ] Workflow file exists at `.github/workflows/ci-cd.yml`
- [ ] Workflow syntax is valid

### Build fails
- [ ] `docker build` works locally
- [ ] `go mod` dependencies are correct
- [ ] Dockerfile has no syntax errors
- [ ] Check GitHub Actions logs

### Deploy fails
- [ ] All GitHub Secrets are set correctly
- [ ] SSH key doesn't have passphrase
- [ ] VM user can run Docker: `ssh user@vm "docker ps"`
- [ ] PROJECT_PATH exists on VM
- [ ] Git repository initialized at PROJECT_PATH

### Services don't start after deploy
- [ ] Check docker logs: `docker compose logs`
- [ ] Ports not in use: `lsof -i :8000`
- [ ] docker-compose.yml syntax: `docker-compose config`
- [ ] Environment variables set correctly

## 📞 Support Commands

```bash
# On VM - check everything is working
docker ps                      # Running containers
docker compose ps              # Services in project
docker-compose logs -f         # Live logs
git log --oneline -5           # Recent commits
docker stats                   # Resource usage

# On local machine - test SSH key
ssh -i ~/.ssh/kafka-deploy-key -v user@vm-ip "echo Success"

# Test GitHub Secrets presence (they won't show value)
# Just verify no errors in Actions tab
```

## 📝 Workflow Files

- Main workflow: `.github/workflows/ci-cd.yml`
- Full setup guide: `CI_CD_SETUP.md`
- This checklist: `CI_CD_QUICKSTART.md`

---

**Status**: Ready for deployment
**Last Updated**: 2026-05-04
