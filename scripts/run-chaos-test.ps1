param(
  [int]$TargetRps = 50,
  [int]$DurationSeconds = 120,
  [string]$FlagKey = 'menu_itau_mobile'
)
$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
$password = 'mvp-chaos-local-only'
$env:POSTGRES_PASSWORD = $password

Write-Host "[1/5] Starting stack"
docker compose -f (Join-Path $root 'docker-compose.yml') up -d --build
Start-Sleep -Seconds 5

Write-Host "[2/5] Enabling target flag"
$headers = @{ 'Content-Type' = 'application/json' }
Invoke-WebRequest -UseBasicParsing -Method Patch -Uri "http://localhost:8080/v1/flags/$FlagKey" -Headers $headers -Body '{"enabled":true}' | Out-Null

Write-Host "[3/5] Starting distributed-load-compatible k6 driver at $TargetRps req/s"
$script = Join-Path $root 'scripts'
$k6Args = @('run','-e',"BASE_URL=http://host.docker.internal:8080",'-e',"FLAG_KEY=$FlagKey",'-e',"TARGET_RPS=$TargetRps",'-e',"DURATION=${DurationSeconds}s",'-e','PREALLOCATED_VUS=100','-e','MAX_VUS=500','/scripts/evaluate-load.js')
$k6 = Start-Process -FilePath 'docker' -ArgumentList (@('run','--rm','--add-host=host.docker.internal:host-gateway','-v',"${script}:/scripts:ro",'grafana/k6') + $k6Args) -PassThru -NoNewWindow

Start-Sleep -Seconds ([Math]::Max(10, [Math]::Floor($DurationSeconds * 0.25)))
Write-Host "[4/5] Chaos: restarting NATS"
docker compose -f (Join-Path $root 'docker-compose.yml') restart nats
Start-Sleep -Seconds ([Math]::Max(10, [Math]::Floor($DurationSeconds * 0.25)))
Write-Host "[4/5] Chaos: restarting PostgreSQL"
docker compose -f (Join-Path $root 'docker-compose.yml') restart postgres
Start-Sleep -Seconds ([Math]::Max(10, [Math]::Floor($DurationSeconds * 0.25)))
Write-Host "[4/5] Chaos: recreating application"
docker compose -f (Join-Path $root 'docker-compose.yml') up -d --no-deps --build app

Write-Host "[5/5] Waiting for k6 report"
$k6.WaitForExit()
if ($k6.ExitCode -ne 0) { throw "k6 chaos test failed with exit code $($k6.ExitCode)" }
Write-Host "Chaos test completed. Preserve the k6 p95/p99 report and container logs as evidence."
