# team-harness installer bootstrap (Windows PowerShell)
# Pipeable: irm https://valianx.github.io/team-harness/install.ps1 | iex
# Or run from a clone: .\bin\install.ps1
$ErrorActionPreference = "Stop"

$Repo    = "valianx/team-harness"
$BaseUrl = "https://github.com/$Repo/releases/latest/download"

# Detect arch (Windows-only script; OS is implicitly windows).
$Arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default {
        Write-Host "Error: unsupported arch '$($env:PROCESSOR_ARCHITECTURE)'." -ForegroundColor Red
        Write-Host "  team-harness supports amd64 and arm64 on Windows."
        Write-Host "  See: https://github.com/$Repo/releases"
        exit 1
    }
}

$Asset = "install-windows-$Arch.exe"
$Url   = "$BaseUrl/$Asset"

$TmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.IO.Path]::GetRandomFileName())
New-Item -ItemType Directory -Path $TmpDir | Out-Null

try {
    $InstallerPath = Join-Path $TmpDir "install.exe"
    Write-Host "Downloading $Asset from latest release..."
    try {
        Invoke-WebRequest -Uri $Url -OutFile $InstallerPath -UseBasicParsing -TimeoutSec 120
    } catch {
        Write-Host "Error: download failed from $Url" -ForegroundColor Red
        Write-Host "  This usually means: (a) no release has been tagged yet, (b) GitHub is"
        Write-Host "  unreachable from this network, or (c) your firewall blocks github.com."
        Write-Host "  Releases: https://github.com/$Repo/releases"
        exit 1
    }

    Write-Host "Launching installer..."
    # Use .NET ProcessStartInfo directly with UseShellExecute=$false to avoid
    # Start-Process's ShellExecuteEx path, which triggers UAC mediation
    # ("requires elevation") when the working directory is protected (e.g.,
    # C:\Windows\System32) or the executable carries Mark-of-the-Web from the
    # download. With UseShellExecute=$false and no stream redirection, the
    # child inherits the parent PowerShell console — same window, all output
    # visible. Cross-compatible with PowerShell 5.1 and 7.x.
    $psi = New-Object System.Diagnostics.ProcessStartInfo
    $psi.FileName = $InstallerPath
    $psi.UseShellExecute = $false
    $psi.CreateNoWindow = $false
    if ($args.Count -gt 0) {
        $psi.Arguments = ($args -join ' ')
    }

    $proc = [System.Diagnostics.Process]::Start($psi)
    $proc.WaitForExit()
    exit $proc.ExitCode
} finally {
    Remove-Item -Recurse -Force $TmpDir -ErrorAction SilentlyContinue
}
