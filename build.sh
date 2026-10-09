#!/usr/bin/env bash
# ==============================================================================
# Script de Build e Empacotamento do dBase II++ para Linux / WSL (Bash)
# Autor: Equipe dBase II++
# Licença: GNU General Public License v3.0 (GPL-3.0)
# ==============================================================================

set -euo pipefail

echo "================================================================================"
echo "                dBase II++ (PC & MSX Edition) - Build System (Linux/WSL)        "
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

NO_INCREMENT=0
for arg in "$@"; do
    case "$arg" in
        --no-increment)
            NO_INCREMENT=1
            ;;
    esac
done

# Extrai e incrementa Z
MAJOR=$(grep -oP '"major":\s*\K[0-9]+' "$VERSION_FILE" || echo 0)
MINOR=$(grep -oP '"minor":\s*\K[0-9]+' "$VERSION_FILE" || echo 1)
PATCH=$(grep -oP '"patch":\s*\K[0-9]+' "$VERSION_FILE" || echo 0)
CODENAME=$(grep -oP '"codename":\s*"\K[^"]+' "$VERSION_FILE" || echo "Nosferatu")
BUILD_DATE=$(date +%Y-%m-%d)
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "local")

if [ "$NO_INCREMENT" -eq 0 ]; then
    PATCH=$((PATCH + 1))
fi
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
go test ./pkg/dbf ./pkg/parser ./pkg/eval ./pkg/engine ./pkg/help ./pkg/msx ./pkg/version ./pkg/storage/sqlite ./pkg/ui/cli

BIN_DIR="bin"
RELEASE_DIR="release"
DIST_DIR="dist"

mkdir -p "$BIN_DIR" "$RELEASE_DIR" "$DIST_DIR"

echo ""
echo ">>> Compilando executável dBase II++ para Linux (amd64)..."
EXE_PATH="${BIN_DIR}/dbase2plus"
TAGS="${TAGS:-nogui}"
go build -tags "${TAGS}" -ldflags "-s -w -X 'dbase2a/pkg/version.BuildMajor=${MAJOR}' -X 'dbase2a/pkg/version.BuildMinor=${MINOR}' -X 'dbase2a/pkg/version.BuildPatch=${PATCH}' -X 'dbase2a/pkg/version.BuildCodename=${CODENAME}' -X 'dbase2a/pkg/version.BuildDateVal=${BUILD_DATE}' -X 'dbase2a/pkg/version.BuildCommit=${COMMIT}'" -o "$EXE_PATH" ./cmd/dbase2plus

echo ">>> Executável gerado: ${EXE_PATH}"

echo ""
echo ">>> Gerando disquete bootável MSX-DOS 2 (720 KB) para MSX 2+..."
DSK_RELEASE_PATH="${RELEASE_DIR}/dbase2plus.dsk"
DSK_DIST_NAME="dbase2plus-${VERSION_STR}-msx2p-dos2.dsk"
DSK_DIST_PATH="${DIST_DIR}/${DSK_DIST_NAME}"
"${EXE_PATH}" -dsk "${DSK_RELEASE_PATH}"
cp "${DSK_RELEASE_PATH}" "${DSK_DIST_PATH}"
echo ">>> Disquete MSX gerado: ${DSK_RELEASE_PATH} (720 KB FAT12 / MSX-DOS 2)"

echo ""
echo ">>> Copiando artefatos para a pasta '${RELEASE_DIR}'..."
rm -f "${RELEASE_DIR}"/*.tar.gz "${RELEASE_DIR}"/*.zip
cp "$EXE_PATH" "${RELEASE_DIR}/"
cp "version.json" "README.md" "SPEC.md" "CHANGELOG.md" "RELEASE.md" "HANDOVER.md" "LICENSE" "${RELEASE_DIR}/"

if [ -d "resource" ]; then
    rm -rf "${RELEASE_DIR}/resource"
    mkdir -p "${RELEASE_DIR}/resource"
    if [ -d "resource/files" ]; then cp -r "resource/files" "${RELEASE_DIR}/resource/"; fi
    if [ -d "resource/disks" ]; then cp -r "resource/disks" "${RELEASE_DIR}/resource/"; fi
    if [ -d "resource/msxdos2" ]; then cp -r "resource/msxdos2" "${RELEASE_DIR}/resource/"; fi
    if [ -f "vram.alf" ]; then cp "vram.alf" "${RELEASE_DIR}/"; cp "vram.alf" "${RELEASE_DIR}/resource/"; fi
    if [ -f "resource/GRAPHOS.ALF" ]; then cp "resource/GRAPHOS.ALF" "${RELEASE_DIR}/resource/"; fi
fi

TAR_NAME="dbase2plus-${VERSION_STR}-linux-amd64.tar.gz"
TAR_DIST_PATH="${DIST_DIR}/${TAR_NAME}"
TAR_RELEASE_PATH="${RELEASE_DIR}/${TAR_NAME}"

echo ""
echo ">>> Gerando pacote TAR.GZ de distribuição (${TAR_NAME})..."
tar -I "gzip -1" -cf "$TAR_DIST_PATH" -C "${RELEASE_DIR}" .
cp "$TAR_DIST_PATH" "$TAR_RELEASE_PATH"

echo ""
echo "================================================================================"
echo "  BUILD CONCLUÍDO COM SUCESSO!                                                  "
echo "  Versão: ${VERSION_STR}  |  Codinome: ${CODENAME}                              "
echo "  Pasta de Release: ${RELEASE_DIR}                                              "
echo "  Pacote de Distribuição: ${TAR_DIST_PATH}                                      "
echo "  Disquete MSX 2+ Bootável: ${DSK_DIST_PATH}                                    "
echo "================================================================================"
