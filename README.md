
# GoLang EKS Monitor

> Production-grade Kubernetes cluster monitoring app with a full DevSecOps pipeline

[![Feature Pipeline](https://github.com/mayanksekhar/GoLang-EKS-Monitor-app/actions/workflows/feature.yml/badge.svg)](https://github.com/mayanksekhar/GoLang-EKS-Monitor-app/actions/workflows/feature.yml)
[![Develop Pipeline](https://github.com/mayanksekhar/GoLang-EKS-Monitor-app/actions/workflows/develop.yml/badge.svg)](https://github.com/mayanksekhar/GoLang-EKS-Monitor-app/actions/workflows/develop.yml)
[![Main Pipeline](https://github.com/mayanksekhar/GoLang-EKS-Monitor-app/actions/workflows/main.yml/badge.svg)](https://github.com/mayanksekhar/GoLang-EKS-Monitor-app/actions/workflows/main.yml)

---

## What This Does

A real Go application that runs **inside an EKS cluster** and queries the Kubernetes metrics-server API to display live resource usage:

- Live CPU and memory utilisation at cluster and node level
- Pod counts by status (Running / Pending / Failed) across all namespaces
- Namespace inventory
- Auto-refresh every 30 seconds
- Thinkwerke terminal-style dark dashboard

This is not a placeholder — it uses `client-go` with a least-privilege `ClusterRole` to read real metrics from the cluster it runs in.

---

## Pipeline Architecture
Push to feature/* → SAST + SCA + Kyverno Policy Validation

Push to develop   → SAST + SCA + Kyverno + Build + Push to GHCR (dev tag)

Push to main      → Full — SAST + SCA + Kyverno + Build + Cosign sign + Deploy to EKS + KBOM + Falco
Push to feature/* → SAST + SCA + Kyverno Policy Validation

Push to develop   → SAST + SCA + Kyverno + Build + Push to GHCR (dev tag)

Push to main      → Full — SAST + SCA + Kyverno + Build + Cosign sign + Deploy to EKS + KBOM + Falco

main          → Protected — requires PR from develop, all checks must pass

└── develop → Integration — SAST + SCA + Kyverno + Build + GHCR push

└── feature/eks-monitor-app    Go app and live dashboard

└── feature/pipeline-security  SAST + SCA workflows

└── feature/kyverno-policy     Kyverno admission policies

└── feature/cosign-signing     Cosign + GHCR integration (planned)

└── feature/helm-deploy        Helm chart + RBAC + EKS deploy (planned)

└── feature/kbom-falco         KBOM scan + Falco runtime (planned)

main          → Protected — requires PR from develop, all checks must pass

└── develop → Integration — SAST + SCA + Kyverno + Build + GHCR push

└── feature/eks-monitor-app    Go app and live dashboard

└── feature/pipeline-security  SAST + SCA workflows

└── feature/kyverno-policy     Kyverno admission policies

└── feature/cosign-signing     Cosign + GHCR integration (planned)

└── feature/helm-deploy        Helm chart + RBAC + EKS deploy (planned)

└── feature/kbom-falco         KBOM scan + Falco runtime (planned)

This passes the policy gate
image: ghcr.io/mayanksekhar/golang-eks-monitor-app@sha256:abc11cec...
This fails the policy gate — pipeline stops
image: ghcr.io/mayanksekhar/golang-eks-monitor-app:develop
This also fails — wrong registry
image: nginx:1.21

This passes the policy gate
image: ghcr.io/mayanksekhar/golang-eks-monitor-app@sha256:abc11cec...
This fails the policy gate — pipeline stops
image: ghcr.io/mayanksekhar/golang-eks-monitor-app:develop
This also fails — wrong registry
image: nginx:1.21

Registry:   ghcr.io/mayanksekhar/golang-eks-monitor-app

Tags:       :develop, :main, :dev-<sha>, :<sha>

Signing:    Cosign keyless (GitHub OIDC)

Verify the image signature
cosign verify ghcr.io/mayanksekhar/golang-eks-monitor-app:main 

--certificate-identity-regexp="https://github.com/mayanksekhar/GoLang-EKS-Monitor-app" 

--certificate-oidc-issuer="https://token.actions.githubusercontent.com"

Verify the image signature
cosign verify ghcr.io/mayanksekhar/golang-eks-monitor-app:main 

--certificate-identity-regexp="https://github.com/mayanksekhar/GoLang-EKS-Monitor-app" 

--certificate-oidc-issuer="https://token.actions.githubusercontent.com"

Cloud:       AWS EKS (us-east-1)

Node:        t3.medium (1 managed node — sized for Falco UI)

K8s version: 1.31

Namespaces:  eks-monitor, ingress-nginx, falco, kube-system

---

## RBAC — Least Privilege ClusterRole

The EKS Monitor queries the metrics-server API which requires specific Kubernetes permissions.
The Helm chart creates a minimal ClusterRole — read-only access to nodes, pods, and namespaces only:

```yaml
rules:
  - apiGroups: [""]
    resources: ["nodes", "pods", "namespaces"]
    verbs: ["get", "list"]
  - apiGroups: ["metrics.k8s.io"]
    resources: ["nodes", "pods"]
    verbs: ["get", "list"]
```

No secrets. No configmaps. No write access. No cluster-admin.

---

## Cosign Verification

Every image pushed to `main` is signed using Cosign keyless signing with GitHub OIDC.
No private keys are stored — the identity is proved by the GitHub Actions workflow itself.

```bash
# Verify any image from this repo
cosign verify \
  ghcr.io/mayanksekhar/golang-eks-monitor-app@sha256:<digest> \
  --certificate-identity-regexp="https://github.com/mayanksekhar/GoLang-EKS-Monitor-app/.github/workflows/main.yml" \
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com"
```

---

## Thinkwerke

This project is part of the **Thinkwerke DevSecOps portfolio** — a set of real, end-to-end security engineering projects built to demonstrate staff/principal-level judgment:

| Project | Focus | Stack |
|---|---|---|
| **GoLang EKS Monitor** (this) | Detect + Prevent | Go, EKS, Falco, Kyverno, Cosign, KBOM |
| **SBOM-gated pipeline** | Prevent | Go, GitLab CI, Syft, Grype, Cosign |
| **OWASP LLM Scanner** | Audit | Python, Ollama, GitLab CI |

Portfolio: `docs.thinkwerke.com`
GitLab: `gitlab.com/mayanksekhar`
GitHub: `github.com/mayanksekhar`

---

## Local Development

```bash
# Requires kubeconfig with access to a running cluster
export CLUSTER_NAME=eks-monitor-demo
go run main.go
# Open http://localhost:8080
```

---

## Teardown

```bash
eksctl delete cluster -f ~/eks-monitor-demo.yaml
```
