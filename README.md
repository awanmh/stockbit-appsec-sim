# Fintech AppSec Simulation (Stockbit Case Study)

![Build Passing](https://img.shields.io/badge/build-passing-brightgreen?style=flat-square)
![Go](https://img.shields.io/badge/Go-1.24-blue?style=flat-square&logo=go)
![Docker](https://img.shields.io/badge/Docker-Enabled-blue?style=flat-square&logo=docker)
![Security](https://img.shields.io/badge/Security-Hardened-red?style=flat-square&logo=security)
![Architecture](https://img.shields.io/badge/Architecture-Clean-orange?style=flat-square)

## 📋 Executive Summary

This project is a **Secure SDLC (Software Development Life Cycle) Simulation** demonstrating how to engineer security into a Fintech application from the ground up. It moves beyond simple coding to showcase the entire lifecycle: **Design -> Attack -> Fix -> Automate**.

Built with **Golang (Gin)** and **Clean Architecture**, this repository serves as a practical case study for:

1.  **Secure Design**: Threat Modeling with STRIDE.
2.  **Vulnerability Management**: Identifying and patching **IDOR** (Insecure Direct Object Reference) flaws.
3.  **DevSecOps**: Automating security checks (SAST, Secret Scanning) in CI/CD.

---

## ⚔️ The Vulnerability Showcase (IDOR)

**Scenario**: "Attacker accessing Victim's Order History"
We implemented two endpoints to demonstrate a critical **Insecure Direct Object Reference (IDOR)** vulnerability, which allows an attacker to view sensitive financial orders belonging to other user accounts.

### ❌ Vulnerable Implementation

`GET /api/v1/vuln/orders/:id`
The system fetches the order by ID but **fails to verify ownership**.

```go
// ⛔ UNSAFE: No verification if order.UserID == currentUserID
order, _ := h.OrderUseCase.GetOrder(ctx, id)
c.JSON(200, order)
```

### ✅ Secure Implementation

`GET /api/v1/secure/orders/:id`
The system verifies that the authenticated user owns the resource before returning data.

```go
// 🛡️ SECURE: Explicit ownership check
order, _ := h.OrderUseCase.GetOrder(ctx, id)
if order.UserID != ctxUser.ID {
    c.AbortWithStatus(403) // Forbidden
}
c.JSON(200, order)
```

> 🕵️‍♂️ **Try it yourself**: A reproduction script [`reproduce_issue.ps1`](reproduce_issue.ps1) is included to automate the attack and verification.

---

## 📂 Security Artifacts

We follow industry-standard documentation practices to bridge the gap between Engineering and Security.

| Document                                      | Description                                                                                         |
| --------------------------------------------- | --------------------------------------------------------------------------------------------------- |
| 🛡️ **[Threat Model](docs/THREAT_MODEL.md)**   | **STRIDE Analysis** & Mermaid.js Sequence Diagrams comparing vulnerable vs secure flows.            |
| 🚨 **[Triage Report](docs/TRIAGE_REPORT.md)** | Professional **Bug Bounty Report** detailing the IDOR severity (CVSS 6.5), impact, and remediation. |

---

## 🚀 DevSecOps Pipeline

Security is automated via **GitHub Actions** (`.github/workflows/security.yml`) to ensure no vulnerabilities reach production.

- 🔑 **Gitleaks**: Scans for hardcoded secrets (API keys, credentials).
- 🛡️ **Gosec**: Performs Static Application Security Testing (SAST) for Go.
- 📦 **Govulncheck**: Scans dependencies for known CVEs.

---

## 🛠️ How to Run

### Prerequisites

- Docker & Docker Compose

### Quick Start

1. **Clone and Run**:
   ```bash
   docker-compose up --build
   ```
2. **Auto-Seeding**:
   The app will automatically seed two users (`attacker@example.com`, `victim@example.com`) and a sample order for the victim on startup.

3. **Verify**:
   ```bash
   # Test IDOR (Requires PowerShell)
   ./reproduce_issue.ps1
   ```
