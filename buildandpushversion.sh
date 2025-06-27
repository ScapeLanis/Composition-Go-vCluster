#!/usr/bin/env bash

# Prüfen, ob eine Version übergeben wurde
if [ -z "$1" ]; then
  echo "Usage: $0 <version>"
  exit 1
fi

VERSION=$1

docker build . --quiet --platform=linux/amd64 --tag runtime-amd64
docker build . --quiet --platform=linux/arm64 --tag runtime-arm64

crossplane xpkg build \
    --package-root=package \
    --embed-runtime-image=runtime-amd64 \
    --package-file=function-amd64.xpkg

crossplane xpkg build \
    --package-root=package \
    --embed-runtime-image=runtime-arm64 \
    --package-file=function-arm64.xpkg

crossplane xpkg push \
    --package-files=function-amd64.xpkg,function-arm64.xpkg \
    ghcr.io/scapelanis/praxisgo:$VERSION


sleep 1

kubectl patch function vcluster \
  --type merge \
  --patch "{\"spec\": {\"package\": \"ghcr.io/scapelanis/praxisgo:$VERSION\"}}"
