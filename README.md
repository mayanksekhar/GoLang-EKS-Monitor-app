# GoLang EKS Monitor

> Production-grade Kubernetes cluster monitoring app with a full DevSecOps pipeline

[![Feature Pipeline](https://github.com/mayanksekhar/GoLang-EKS-Monitor-app/actions/workflows/feature.yml/badge.svg)](https://github.com/mayanksekhar/GoLang-EKS-Monitor-app/actions/workflows/feature.yml)
[![Develop Pipeline](https://github.com/mayanksekhar/GoLang-EKS-Monitor-app/actions/workflows/develop.yml/badge.svg)](https://github.com/mayanksekhar/GoLang-EKS-Monitor-app/actions/workflows/develop.yml)
[![Main Pipeline](https://github.com/mayanksekhar/GoLang-EKS-Monitor-app/actions/workflows/main.yml/badge.svg)](https://github.com/mayanksekhar/GoLang-EKS-Monitor-app/actions/workflows/main.yml)

---

## What This Does

A real Go application that runs **inside an EKS cluster** and queries the Kubernetes metrics-server API to display live resource usage:

- CPU and memory utilisation at cluster and node level
- Pod counts by status (Running / Pending / Failed) across all namespaces
- Namespace inventory
- Auto-refresh every 30 seconds
- Thinkwerke terminal-style dark dashboard

This is not a placeholder — it uses `client-go` with a least-privilege `ClusterRole` to read real metrics from the cluster it runs in.

---

## Pipeline Architecture
Push to feature/* → SAST + SCA only

Push to develop   → SAST + SCA + Build + Push to GHCR (dev tag)

Push to main      → SAST + SCA + Build + Push + Cosign sign + Deploy to EKS + KBOM scan
### Stages

| Stage | Tool | What it checks |
|---|---|---|
| SAST | Semgrep | Static analysis of Go source and Dockerfile |
| SCA | Trivy + Nancy | Go module dependency vulnerabilities |
| Build | Docker | Multi-stage build → distroless image |
| Push | GHCR | Push to ghcr.io with SHA digest tag |
| Sign | Cosign | Keyless signing via GitHub OIDC — no keys stored |
| Deploy | Helm | Helm chart with RBAC ClusterRole for metrics-server access |
| KBOM | Trivy k8s | Cluster-wide CycloneDX inventory + vulnerability scan |
| Runtime | Falco | eBPF-based syscall monitoring with MITRE ATT&CK mapping |

---

## Multi-Branch Strategy
main          Protected — requires PR from develop, pipeline must pass

└── develop Integration branch — SAST + SCA + Build + GHCR push

└── feature/eks-monitor-app   Go app and dashboard

└── feature/pipeline-security SAST + SCA workflows

└── feature/cosign-signing    Cosign + GHCR integration

└── feature/helm-deploy       Helm chart + RBAC + EKS deploy

└── feature/kbom-falco        KBOM scan + Falco runtime
Each feature branch triggers fast-feedback security scanning only. Nothing deploys until a PR is merged to develop, and nothing goes to production until develop merges to main. This mirrors real enterprise DevSecOps workflow.

---

## Security Highlights

**Supply chain (build time)**
- Semgrep SAST catches misconfigurations in Go code and Dockerfile before merge
- Trivy + Nancy SCA gates on known CVEs in Go module dependencies
- Cosign keyless signing attaches a verifiable identity to every pushed image — no long-lived keys
- Image digest pinned in Helm values for every deployment

**Runtime (in-cluster)**
- Falco DaemonSet via eBPF — zero in-container agent, zero app changes required
- MITRE ATT&CK mapped alerts (T1555, T1552.001 confirmed in demo)
- ingress-nginx v1.6.4 intentionally deployed to demonstrate KBOM vulnerability discovery (CVE-2023-5043, CVE-2025-1974)

**Visibility**
- KBOM generated in CycloneDX JSON format — 16 components inventoried
- Falco UI enabled for visual alert browsing
- All pipeline artifacts (SBOM, KBOM, scan results) retained per run

---

## Container Image
Registry:   ghcr.io/mayanksekhar/golang-eks-monitor-app

Signing:    Cosign keyless (GitHub OIDC)

Verify:     cosign verify ghcr.io/mayanksekhar/golang-eks-monitor-app:main 

--certificate-identity-regexp="https://github.com/mayanksekhar/GoLang-EKS-Monitor-app" 

--certificate-oidc-issuer="https://token.actions.githubusercontent.com"
---

## Infrastructure
Cloud:       AWS EKS (us-east-1)

Node:        t3.medium (1 managed node)

K8s version: 1.31

Namespaces:  eks-monitor, ingress-nginx, falco, kube-system
---

## RBAC — Why the App Needs a ClusterRole

The EKS Monitor queries the metrics-server API, which requires specific Kubernetes permissions. Rather than granting broad access, the Helm chart creates a minimal ClusterRole:

```yaml
rules:
  - apiGroups: [""]
    resources: ["nodes", "pods", "namespaces"]
    verbs: ["get", "list"]
  - apiGroups: ["metrics.k8s.io"]
    resources: ["nodes", "pods"]
    verbs: ["get", "list"]
```

This is intentionally narrow — it can read node and pod metrics, nothing else. No secrets, no configmaps, no write access.

---

## Thinkwerke

This project is part of the Thinkwerke DevSecOps portfolio — a set of real, end-to-end security engineering projects built to demonstrate staff/principal-level judgment:

- **Detect** → Falco runtime security + KBOM scanning (this project)
- **Prevent** → SBOM-gated supply chain pipeline with Cosign attestation
- **Audit** → OWASP LLM Top 10 scanner suite

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

