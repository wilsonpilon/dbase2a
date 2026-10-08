# ==============================================================================
# Script de Build e Empacotamento do dBase II++ para Windows (PowerShell)
# Autor: Equipe dBase II++
# Licença: GNU General Public License v3.0 (GPL-3.0)
# ==============================================================================

[CmdletBinding()]
param (
    [string]$Mode = "all",
    [switch]$SkipTests = $false,
    [switch]$NoIncrement = $false,
    [string]$CustomCodename = "",
    [string]$Tags = "nogui"
)

$ErrorActionPreference = "Stop"

Write-Host "================================================================================" -ForegroundColor Cyan
Write-Host "                dBase II++ (PC & MSX Edition) - Build System (Windows)          " -ForegroundColor Yellow
Write-Host "================================================================================" -ForegroundColor Cyan

# 1. Carregamento e incremento da versão (vX.Y.Z) onde Z é o contador de builds
$VersionFile = "version.json"
if (-not (Test-Path $VersionFile)) {
    $VersionData = @{
        major = 0
        minor = 1
        patch = 0
        codename = "Nosferatu"
        build_date = (Get-Date -Format "yyyy-MM-dd")
        commit = "dev"
    }
} else {
    $JsonContent = Get-Content $VersionFile -Raw
    $VersionData = $JsonContent | ConvertFrom-Json
}

# Incrementa Z (patch / build number) se não solicitado NoIncrement
if (-not $NoIncrement) {
    $VersionData.patch = [int]$VersionData.patch + 1
}
$VersionData.build_date = (Get-Date -Format "yyyy-MM-dd")

if ($CustomCodename -ne "") {
    $VersionData.codename = $CustomCodename
}

# Tenta obter o commit hash do Git se o repositório estiver inicializado
try {
    $GitCommit = (git rev-parse --short HEAD 2>$null)
    if ($GitCommit) {
        $VersionData.commit = $GitCommit.Trim()
    }
} catch {
    $VersionData.commit = "local"
}

# Salva version.json atualizado
$UpdatedJson = $VersionData | ConvertTo-Json -Depth 5
[System.IO.File]::WriteAllText((Resolve-Path $VersionFile), $UpdatedJson, [System.Text.Encoding]::UTF8)

$VersionStr = "v$($VersionData.major).$($VersionData.minor).$($VersionData.patch)"
Write-Host ">>> Nova Versão de Build: $VersionStr `"$($VersionData.codename)`" (Build: $($VersionData.build_date))" -ForegroundColor Green

# 2. Resolução de dependências
Write-Host "`n>>> Sincronizando dependências (go mod tidy)..." -ForegroundColor Cyan
go mod tidy

# 3. Execução de testes automatizados
if (-not $SkipTests) {
    Write-Host "`n>>> Executando suíte de testes unitários..." -ForegroundColor Cyan
    go test ./pkg/dbf ./pkg/parser ./pkg/eval ./pkg/engine ./pkg/help ./pkg/msx ./pkg/version ./pkg/storage/sqlite ./pkg/ui/cli
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Falha na execução dos testes unitários. Abortando compilação."
    }
    Write-Host ">>> Todos os testes passaram com sucesso!" -ForegroundColor Green
}

# 4. Criação de pastas de saída
$BinDir = "bin"
$ReleaseDir = "release"
$DistDir = "dist"

if (-not (Test-Path $BinDir)) { New-Item -ItemType Directory -Path $BinDir | Out-Null }
if (-not (Test-Path $ReleaseDir)) { New-Item -ItemType Directory -Path $ReleaseDir | Out-Null }
if (-not (Test-Path $DistDir)) { New-Item -ItemType Directory -Path $DistDir | Out-Null }

# 5. Compilação do executável Windows
Write-Host "`n>>> Compilando executável dBase II++ para Windows (amd64)..." -ForegroundColor Cyan
$ExeName = "dbase2plus.exe"
$ExePath = Join-Path $BinDir $ExeName

$LdFlags = "-s -w -X 'dbase2a/pkg/version.Current.Major=$($VersionData.major)' -X 'dbase2a/pkg/version.Current.Minor=$($VersionData.minor)' -X 'dbase2a/pkg/version.Current.Patch=$($VersionData.patch)' -X 'dbase2a/pkg/version.Current.Codename=$($VersionData.codename)' -X 'dbase2a/pkg/version.Current.BuildDate=$($VersionData.build_date)' -X 'dbase2a/pkg/version.Current.Commit=$($VersionData.commit)'"

if ($Tags -ne "") {
    go build -tags $Tags -ldflags $LdFlags -o $ExePath ./cmd/dbase2plus
} else {
    go build -ldflags $LdFlags -o $ExePath ./cmd/dbase2plus
}

if (-not (Test-Path $ExePath)) {
    Write-Error "Erro: Executável $ExePath não foi gerado."
}
Write-Host ">>> Executável gerado: $ExePath ($((Get-Item $ExePath).Length / 1MB | ForEach-Object { '{0:N2} MB' -f $_ }))" -ForegroundColor Green

# 6. Preparação da pasta release/
Write-Host "`n>>> Copiando artefatos para a pasta '$ReleaseDir'..." -ForegroundColor Cyan
Get-ChildItem -Path $ReleaseDir -Filter "*.zip" | Remove-Item -Force -ErrorAction SilentlyContinue
Get-ChildItem -Path $ReleaseDir -Filter "*.tar.gz" | Remove-Item -Force -ErrorAction SilentlyContinue
Copy-Item $ExePath -Destination $ReleaseDir -Force
Copy-Item "version.json" -Destination $ReleaseDir -Force
Copy-Item "README.md" -Destination $ReleaseDir -Force
Copy-Item "SPEC.md" -Destination $ReleaseDir -Force
Copy-Item "CHANGELOG.md" -Destination $ReleaseDir -Force
Copy-Item "RELEASE.md" -Destination $ReleaseDir -Force
Copy-Item "HANDOVER.md" -Destination $ReleaseDir -Force
Copy-Item "LICENSE" -Destination $ReleaseDir -Force

if (Test-Path "resource") {
    $RelResDir = Join-Path $ReleaseDir "resource"
    if (Test-Path $RelResDir) {
        Remove-Item $RelResDir -Recurse -Force
    }
    New-Item -ItemType Directory -Path $RelResDir | Out-Null
    if (Test-Path "resource\files") { Copy-Item "resource\files" -Destination $RelResDir -Recurse -Force }
    if (Test-Path "resource\disks") { Copy-Item "resource\disks" -Destination $RelResDir -Recurse -Force }
}

# 7. Geração do pacote compactado ZIP para distribuição
$ZipName = "dbase2plus-$VersionStr-windows-amd64.zip"
$ZipDistPath = Join-Path $DistDir $ZipName
$ZipReleasePath = Join-Path $ReleaseDir $ZipName

Write-Host "`n>>> Gerando pacote ZIP de distribuição ($ZipName)..." -ForegroundColor Cyan
if (Test-Path $ZipDistPath) { Remove-Item $ZipDistPath -Force }
Compress-Archive -Path "$ReleaseDir\*" -DestinationPath $ZipDistPath -CompressionLevel Fastest
Copy-Item $ZipDistPath -Destination $ZipReleasePath -Force

Write-Host "`n================================================================================" -ForegroundColor Green
Write-Host "  BUILD CONCLUÍDO COM SUCESSO!                                                  " -ForegroundColor Green
Write-Host "  Versão: $VersionStr  |  Codinome: $($VersionData.codename)                     " -ForegroundColor Yellow
Write-Host "  Pasta de Release: $ReleaseDir                                                 " -ForegroundColor White
Write-Host "  Pacote de Distribuição: $ZipDistPath                                          " -ForegroundColor White
Write-Host "================================================================================" -ForegroundColor Green
