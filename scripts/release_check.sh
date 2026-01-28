#!/usr/bin/env bash
set -euo pipefail

TAG="$(git describe --tags --exact-match 2>/dev/null || true)"
if [[ -z "${TAG}" ]]; then
  echo "ERROR: release build must be on an exact git tag."
  exit 1
fi
echo "OK: on tag ${TAG}"
