#!/usr/bin/env bash
set -euo pipefail

PKG="github.com/cainlara/gogit-branch/version"
OUT="${1:-gogit-branch}"

VERSION="$(git describe --tags --exact-match 2>/dev/null || true)"
if [[ -z "${VERSION}" ]]; then
  VERSION="$(git describe --tags --always --dirty)"
fi

COMMIT="$(git rev-parse --short HEAD)"
DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
DIRTY="false"
if ! git diff --quiet || ! git diff --cached --quiet; then
  DIRTY="true"
fi

LDFLAGS="-s -w \
  -X '${PKG}.Version=${VERSION}' \
  -X '${PKG}.Commit=${COMMIT}' \
  -X '${PKG}.Date=${DATE}' \
  -X '${PKG}.Dirty=${DIRTY}'"

go build -trimpath -ldflags "${LDFLAGS}" -o "${OUT}"

echo "Built ${OUT}: ${VERSION} (${COMMIT}) dirty=${DIRTY} date=${DATE}"
