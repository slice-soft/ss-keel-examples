#!/usr/bin/env bash
# run-example.sh — Helper to run any example from the repo root.
#
# Usage: ./shared/scripts/run-example.sh <example-directory>
# Example: ./shared/scripts/run-example.sh examples/01-hello-world

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

if [[ $# -eq 0 ]]; then
  echo "Usage: $0 <example-directory>"
  echo ""
  echo "Available examples:"
  for dir in "${REPO_ROOT}/examples"/*/; do
    name=$(basename "$dir")
    echo "  examples/${name}"
  done
  exit 1
fi

EXAMPLE_DIR="${REPO_ROOT}/$1"

if [[ ! -d "${EXAMPLE_DIR}" ]]; then
  echo "Error: directory '${EXAMPLE_DIR}' does not exist."
  exit 1
fi

if [[ ! -f "${EXAMPLE_DIR}/go.mod" ]]; then
  echo "Error: '${EXAMPLE_DIR}' is not a Go module (missing go.mod)."
  exit 1
fi

# Copy .env.example if .env is missing
if [[ ! -f "${EXAMPLE_DIR}/.env" ]] && [[ -f "${EXAMPLE_DIR}/.env.example" ]]; then
  echo "→ Copying .env.example to .env"
  cp "${EXAMPLE_DIR}/.env.example" "${EXAMPLE_DIR}/.env"
fi

echo "→ Downloading dependencies..."
(cd "${EXAMPLE_DIR}" && go mod download)

echo "→ Running $(basename "${EXAMPLE_DIR}")..."
echo ""
(cd "${EXAMPLE_DIR}" && go run main.go)
