# Build .NET Analyzer for multiple platforms

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$analyzerDir = Join-Path $repoRoot "internal\dotnets\DeadCodeAnalyzer"

Write-Host ""
Write-Host "Building C# Analyzer (Multi-platform)..." -ForegroundColor Cyan
Write-Host ""

Push-Location $analyzerDir

try {
    # Platform configurations
    $platforms = @(
        @{ Name = "Windows"; RID = "win-x64"; File = "DeadCodeAnalyzer.exe"; Target = "deadcode_windows_amd64.exe" },
        @{ Name = "Linux";   RID = "linux-x64"; File = "DeadCodeAnalyzer"; Target = "deadcode_linux_amd64" },
        @{ Name = "macOS";   RID = "osx-x64"; File = "DeadCodeAnalyzer"; Target = "deadcode_darwin_amd64" }
    )

    $index = 1
    foreach ($platform in $platforms) {
        Write-Host "[$index/3] Building $($platform.Name) x64..." -ForegroundColor Yellow

        $output = dotnet publish -c Release -r $platform.RID --self-contained true --nologo 2>&1

        if ($LASTEXITCODE -ne 0) {
            Write-Host $output
            throw "$($platform.Name) build failed!"
        }

        $source = Join-Path $analyzerDir "bin\Release\net10.0\$($platform.RID)\publish\$($platform.File)"
        $target = Join-Path $repoRoot "internal\dotnets\$($platform.Target)"

        if (Test-Path $source) {
            Copy-Item -Path $source -Destination $target -Force
            $fileSize = [math]::Round((Get-Item $target).Length / 1MB, 2)
            Write-Host "[OK] $($platform.Name) completed ($fileSize MB)" -ForegroundColor Green
        } else {
            throw "Source file not found: $source"
        }

        $index++
    }

    Pop-Location

    Write-Host ""
    Write-Host "All platforms built successfully!" -ForegroundColor Green
    Write-Host "Files location: internal\dotnets\" -ForegroundColor White
    Write-Host "  - deadcode_windows_amd64.exe" -ForegroundColor Gray
    Write-Host "  - deadcode_linux_amd64" -ForegroundColor Gray
    Write-Host "  - deadcode_darwin_amd64" -ForegroundColor Gray
    Write-Host ""
    Write-Host "Now you can run build.bat to compile the Go project" -ForegroundColor Cyan
    Write-Host ""
} catch {
    Pop-Location
    Write-Host ""
    Write-Host "[ERROR] $_" -ForegroundColor Red
    Write-Host ""
    exit 1
}
