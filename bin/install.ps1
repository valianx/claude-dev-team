# claude-dev-team bootstrap (Windows PowerShell)
# Ensures uv is installed, then runs the Python installer via `uv run`.
$ErrorActionPreference = "Stop"

$RepoRoot = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)

Write-Host "claude-dev-team bootstrap"

if (-not (Get-Command uv -ErrorAction SilentlyContinue)) {
    Write-Host "uv not found. Installing via astral.sh..."
    Invoke-Expression (Invoke-RestMethod https://astral.sh/uv/install.ps1)

    # Refresh PATH for the current session so `uv` resolves
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine") + ";" +
                [System.Environment]::GetEnvironmentVariable("Path", "User")
}

if (-not (Get-Command uv -ErrorAction SilentlyContinue)) {
    Write-Error "uv install did not succeed. Install it manually (https://docs.astral.sh/uv/) and re-run."
    exit 1
}

Write-Host ("uv: " + (& uv --version))
Write-Host "Running installer..."
Write-Host ""

& uv run (Join-Path $RepoRoot "bin\install.py") @args
exit $LASTEXITCODE
