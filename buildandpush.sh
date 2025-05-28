#!/usr/bin/env bash


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


crossplane xpkg push   --package-files=function-amd64.xpkg,function-arm64.xpkg   ghcr.io/scapelanis/praxisgo:v0.2.14

