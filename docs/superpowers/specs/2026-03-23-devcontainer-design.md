# Devcontainer for kubernetes-autoscaler

**Date:** 2026-03-23
**Status:** Approved

## Goal

Create a devcontainer configuration at the repo root suitable for ongoing development of the kubernetes-autoscaler monorepo, with a focus on the Azure cloud provider. The configuration should be upstreamable as a PR to kubernetes/autoscaler, work in both GitHub Codespaces and local VS Code, and be a superset of the default Codespaces experience.

## Decisions

| Question | Decision | Rationale |
|---|---|---|
| Audience | Broader contributor community | Upstreamable, not personal |
| Docker tooling | Docker-in-Docker | Superset of Codespaces; needed for `make make-image` and multi-arch builds |
| Azure/K8s CLIs | Azure CLI + kubectl + Helm | Full stack for deploying/testing CA on AKS |
| Configuration style | Feature-based only | Declarative, maintainable, no custom Dockerfile |
| VS Code extensions | Go extension only | Minimal; contributors add their own |
| Post-create setup | Download deps + install verify tools | Pre-warm module caches and install CI lint tools |
| System deps | libseccomp-dev only | Required for builds; protoc installable on demand |
| Local cluster tools | minikube + kind | Easy local testing |

## Architecture

Single file: `.devcontainer/devcontainer.json`. No Dockerfile, no helper scripts.

### Base Image

`mcr.microsoft.com/devcontainers/go:1.25` — the official Microsoft devcontainer Go image (Debian-based). Go 1.25 matches CI, the builder Dockerfile, and the majority of go.mod files. Go's backward compatibility means it handles the older modules (balancer at 1.19) without issue.

### Devcontainer Features

| Feature | Purpose | Config |
|---|---|---|
| `ghcr.io/devcontainers/features/docker-in-docker:2` | Full Docker + buildx for container image builds | `dockerDashComposeVersion: "v2"` |
| `ghcr.io/devcontainers/features/azure-cli:1` | `az` CLI for AKS interaction and Azure auth | defaults |
| `ghcr.io/devcontainers/features/kubectl-helm-minikube:1` | kubectl, Helm, minikube for cluster operations | `minikube: "latest"` |
| `ghcr.io/devcontainers/features/github-cli:1` | `gh` CLI for PR workflows | defaults |

Kind is installed separately via `go install sigs.k8s.io/kind@latest` in the post-create step (no dedicated feature exists).

### Lifecycle Commands

**`onCreateCommand`** — runs once when the container is first created, before workspace mount:
```
sudo apt-get update && sudo apt-get install -y --no-install-recommends libseccomp-dev
```

**`postCreateCommand`** — runs after the workspace is mounted:
```bash
cd cluster-autoscaler && go mod download && cd .. && \
cd vertical-pod-autoscaler && go mod download && cd .. && \
hack/install-verify-tools.sh && \
go install sigs.k8s.io/kind@latest
```

This pre-warms the Go module cache for the two main sub-projects and installs the CI verification tools (`golint`, `godep`, `misspell`) plus kind.

### VS Code Customizations

**Extensions:**
- `golang.go`

**Settings:**
- `go.testFlags: ["-race", "-count=1"]` — matches CI test invocations
- `go.lintTool: "golangci-lint"` — matches VPA's CI lint configuration
- `go.lintOnSave: "workspace"`

### What's Intentionally Excluded

- **protoc** — protobuf codegen is rare; installable on demand
- **Node.js / Python** — not needed for this Go-only repo
- **Custom Dockerfile** — features cover everything needed
- **Per-sub-project configurations** — the sub-projects share tooling and are converging on Go 1.25
- **Additional VS Code extensions** — kept minimal for contributor neutrality

## Complete devcontainer.json

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
