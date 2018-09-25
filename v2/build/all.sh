#!/bin/bash
set -Euxo pipefail

# Base V2 directory which parents this script.
V2DIR="$(cd "$(dirname "$0")/.." && pwd)"

# Build and test Gazette. This image includes all Gazette source, vendored
# dependencies, compiled packages and binaries, and only completes after
# all tests pass. The gazette-build image is a local-only artifact and never
# published (eg, to Docker hub).
docker build ${V2DIR} --file ${V2DIR}/build/Dockerfile.gazette-build --tag gazette-build

# Create the `gazette` command container, which plucks the `gazette` and
# `gazctl` binaries onto the `gazette-base` image.
docker build ${V2DIR} --file ${V2DIR}/build/cmd/Dockerfile.gazette --tag liveramp/gazette

# Gazette examples also have Dockerized build targets. Build the `stream-sum`
# example, which builds the stream-sum plugin and adds it to `gazette-base`
# with the `run-consumer` binary.
docker build ${V2DIR} --file ${V2DIR}/build/examples/Dockerfile.stream-sum --tag liveramp/gazette-example-stream-sum

# Build the `word-count` example.
docker build ${V2DIR} --file ${V2DIR}/build/examples/Dockerfile.word-count --tag liveramp/gazette-example-word-count
