#!/usr/bin/env bash
# ==============================================================================
# Script de Build e Empacotamento do dBase II+ para Linux / WSL (Bash)
# Autor: Equipe dBase II+
# Licença: GNU General Public License v3.0 (GPL-3.0)
# ==============================================================================

set -euo pipefail

echo "================================================================================"
echo "                dBase II+ (PC & MSX Edition) - Build System (Linux/WSL)         "
echo "================================================================================"

VERSION_FILE="version.json"

if [ ! -f "$VERSION_FILE" ]; then
    cat <<EOF > "$VERSION_FILE"
{
  "major": 0,
  "minor": 1,
  "patch": 0,
  "codename": "Nosferatu",
  "build_date": "$(date +%Y-%m-%d)",
  "commit": "dev"
}
EOF
fi

# Extrai e incrementa Z
MAJOR=$(grep -oP '"major":\s*\K[0-9]+' "$VERSION_FILE" || echo 0)
MINOR=$(grep -oP '"minor":\s*\K[0-9]+' "$VERSION_FILE" || echo 1)
PATCH=$(grep -oP '"patch":\s*\K[0-9]+' "$VERSION_FILE" || echo 0)
CODENAME=$(grep -oP '"codename":\s*"\K[^"]+' "$VERSION_FILE" || echo "Nosferatu")
BUILD_DATE=$(date +%Y-%m-%d)
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "local")

PATCH=$((PATCH + 1))
VERSION_STR="v${MAJOR}.${MINOR}.${PATCH}"

# Atualiza version.json
cat <<EOF > "$VERSION_FILE"
{
  "major": ${MAJOR},
  "minor": ${MINOR},
  "patch": ${PATCH},
  "codename": "${CODENAME}",
  "build_date": "${BUILD_DATE}",
  "commit": "${COMMIT}"
}
EOF

echo ">>> Nova Versão de Build: ${VERSION_STR} \"${CODENAME}\" (Build: ${BUILD_DATE})"

echo ""
echo ">>> Sincronizando dependências (go mod tidy)..."
go mod tidy

echo ""
echo ">>> Executando suíte de testes unitários..."
go test ./pkg/dbf ./pkg/parser ./pkg/engine ./pkg/msx ./pkg/version ./pkg/storage/sqlite

BIN_DIR="bin"
RELEASE_DIR="release"
DIST_DIR="dist"

mkdir -p "$BIN_DIR" "$RELEASE_DIR" "$DIST_DIR"

echo ""
echo ">>> Compilando executável dBase II+ para Linux (amd64)..."
EXE_PATH="${BIN_DIR}/dbase2plus"
TAGS="${TAGS:-nogui}"
go build -tags "${TAGS}" -ldflags "-s -w -X 'dbase2a/pkg/version.Current.Major=${MAJOR}' -X 'dbase2a/pkg/version.Current.Minor=${MINOR}' -X 'dbase2a/pkg/version.Current.Patch=${PATCH}' -X 'dbase2a/pkg/version.Current.Codename=${CODENAME}' -X 'dbase2a/pkg/version.Current.BuildDate=${BUILD_DATE}' -X 'dbase2a/pkg/version.Current.Commit=${COMMIT}'" -o "$EXE_PATH" ./cmd/dbase2plus

echo ">>> Executável gerado: ${EXE_PATH}"

echo ""
echo ">>> Copiando artefatos para a pasta '${RELEASE_DIR}'..."
cp "$EXE_PATH" "${RELEASE_DIR}/"
cp "version.json" "README.md" "SPEC.md" "CHANGELOG.md" "RELEASE.md" "LICENSE" "${RELEASE_DIR}/"

if [ -d "resource" ]; then
    cp -r "resource" "${RELEASE_DIR}/"
fi

TAR_NAME="dbase2plus-${VERSION_STR}-linux-amd64.tar.gz"
TAR_DIST_PATH="${DIST_DIR}/${TAR_NAME}"
TAR_RELEASE_PATH="${RELEASE_DIR}/${TAR_NAME}"

echo ""
echo ">>> Gerando pacote TAR.GZ de distribuição (${TAR_NAME})..."
tar -czf "$TAR_DIST_PATH" -C "${RELEASE_DIR}" .
cp "$TAR_DIST_PATH" "$TAR_RELEASE_PATH"

echo ""
echo "================================================================================"
echo "  BUILD CONCLUÍDO COM SUCESSO!                                                  "
echo "  Versão: ${VERSION_STR}  |  Codinome: ${CODENAME}                              "
echo "  Pasta de Release: ${RELEASE_DIR}                                              "
echo "  Pacote de Distribuição: ${TAR_DIST_PATH}                                      "
echo "================================================================================"
