# Devcontainer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create `.devcontainer/devcontainer.json` for the kubernetes-autoscaler monorepo with Go 1.25, Docker-in-Docker, Azure CLI, kubectl, Helm, minikube, kind, and CI tool pre-installation.

**Architecture:** Single `devcontainer.json` file using the `mcr.microsoft.com/devcontainers/go:1.25` base image with devcontainer features for Docker-in-Docker, Azure CLI, kubectl/Helm/minikube, and GitHub CLI. Lifecycle commands install system deps and pre-warm Go module caches.

**Tech Stack:** devcontainer spec, Go 1.25, Docker-in-Docker, Azure CLI, kubectl, Helm, minikube, kind

**Spec:** `docs/superpowers/specs/2026-03-23-devcontainer-design.md`

---

## File Structure

- **Create:** `.devcontainer/devcontainer.json` — the complete devcontainer configuration

No other files are created or modified.

---

### Task 1: Create devcontainer.json

**Files:**
- Create: `.devcontainer/devcontainer.json`

- [ ] **Step 1: Create the `.devcontainer/` directory and `devcontainer.json`**

Create `.devcontainer/devcontainer.json` with this exact content:

```jsonc
{
  "name": "Kubernetes Autoscaler Development",
  "image": "mcr.microsoft.com/devcontainers/go:1.25",

  "features": {
    "ghcr.io/devcontainers/features/docker-in-docker:2": {
      "dockerDashComposeVersion": "v2"
    },
    "ghcr.io/devcontainers/features/azure-cli:1": {},
    "ghcr.io/devcontainers/features/kubectl-helm-minikube:1": {
      "minikube": "latest"
    },
    "ghcr.io/devcontainers/features/github-cli:1": {}
  },

  "onCreateCommand": "sudo apt-get update && sudo apt-get install -y --no-install-recommends libseccomp-dev",

  "postCreateCommand": "cd cluster-autoscaler && go mod download && cd .. && cd vertical-pod-autoscaler && go mod download && cd .. && hack/install-verify-tools.sh && go install sigs.k8s.io/kind@latest",

  "customizations": {
    "vscode": {
      "extensions": [
        "golang.go"
      ],
      "settings": {
        "go.testFlags": ["-race", "-count=1"],
        "go.lintTool": "golangci-lint",
        "go.lintOnSave": "workspace"
      }
    }
  }
}
```

- [ ] **Step 2: Validate the JSON is well-formed**

Run: `python3 -c "import json; json.load(open('.devcontainer/devcontainer.json')); print('Valid JSON')"`
Expected: `Valid JSON`

- [ ] **Step 3: Commit**

```bash
git add .devcontainer/devcontainer.json
git commit -m "feat: add devcontainer for development environment

Adds a devcontainer configuration with:
- Go 1.25 base image (mcr.microsoft.com/devcontainers/go:1.25)
- Docker-in-Docker for container image builds
- Azure CLI, kubectl, Helm, minikube for AKS workflows
- GitHub CLI for PR workflows
- libseccomp-dev system dependency
- Pre-warmed Go module caches for cluster-autoscaler and VPA
- CI verification tools (golint, godep, misspell) and kind pre-installed
- Go extension with CI-matching test and lint settings"
```
