#!/bin/bash

set -euo pipefail

# Change cwd to script path.
cd "$(dirname "$0")"

cd ..
mkdir -p script-playground && cd script-playground

rm -f go.mod go.sum || true
go mod init script-playground
