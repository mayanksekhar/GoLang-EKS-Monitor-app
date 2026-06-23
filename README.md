# GoLang EKS Monitor

[![Build](https://github.com/mayanksekhar/GoLang-EKS-Monitor-app/actions/workflows/feature.yml/badge.svg)](https://github.com/mayanksekhar/GoLang-EKS-Monitor-app/actions)

A real-time EKS cluster monitoring dashboard built in Go, secured with a full DevSecOps pipeline.

## What it does
- Live CPU, memory metrics at cluster and node level
- Pod counts by status across all namespaces
- Auto-refresh every 30 seconds
- Thinkwerke-branded terminal aesthetic

## Pipeline
| Branch | Stages |
|---|---|
| `feature/*` | SAST + SCA |
| `develop` | SAST + SCA + Build + Push to GHCR |
| `main` | Full — SAST + SCA + Build + Push + Cosign sign + Deploy + KBOM |

## Runner
Executed on **son-of-anton** — self-hosted GitHub Actions runner on Fedora 44 ThinkPad X1 Carbon.
