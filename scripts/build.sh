#!/bin/bash
# Builds every platform into bin/. Needs Go and nothing else: the binaries are
# static (CGO_ENABLED=0), so there is no libc to match on the far end.
#
#   scripts/build.sh            all five targets
#   scripts/build.sh host       just this machine's, for a quick loop
set -euo pipefail

ROOT=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd -P)
cd "$ROOT"
command -v go >/dev/null || { echo "go is required" >&2; exit 1; }

VERSION=$(python3 -c 'import json;print(json.load(open(".claude-plugin/plugin.json"))["version"])' 2>/dev/null || echo dev)
# -trimpath keeps build paths out of the binary; -s -w drop the symbol table and
# DWARF, which is about a third of the size and nothing anyone needs at runtime.
FLAGS=(-trimpath -ldflags "-s -w -X main.version=$VERSION")

TARGETS=(linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64)
if [ "${1:-}" = "host" ]; then
  TARGETS=("$(go env GOOS)/$(go env GOARCH)")
fi

mkdir -p bin
for target in "${TARGETS[@]}"; do
  os=${target%/*}; arch=${target#*/}
  out="bin/ccpet-$os-$arch"
  [ "$os" = windows ] && out="$out.exe"
  GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build "${FLAGS[@]}" -o "$out" ./cmd/ccpet
  printf '  %-28s %s\n' "$out" "$(du -h "$out" | cut -f1)"
done
echo "version $VERSION"
