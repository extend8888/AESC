#!/bin/bash

set -euo pipefail

# Ensure go is in PATH
export PATH="/usr/local/go/bin:$HOME/go/bin:$PATH"

echo "Generating protobuf code..."

rm -rf ./build/proto

# We have to build regen-network protoc-gen-gocosmos from source because
# the module uses replace directive, which makes it impossible to use
# go install like a healthy human being.
#
# As a workaround, we download the source code to a temporary location
# and build the binary. buf.gen.yaml then implicitly uses the path to the
# built binary. This is ugly but it works, and results in the least amount
# of changes across the repo to have _a_ working solution without accidentally
# breaking anything else or introduce too much change as part of automating
# the proto generation.
go get github.com/regen-network/cosmos-proto/protoc-gen-gocosmos@v0.3.1
mkdir -p ./build/proto/gocosmos
build_out="${PWD}/build/proto/gocosmos"
pushd "$(go env GOMODCACHE)/github.com/regen-network/cosmos-proto@v0.3.1" &&
  go build -o "${build_out}/protoc-gen-gocosmos" ./protoc-gen-gocosmos &&
  popd

go run github.com/bufbuild/buf/cmd/buf@v1.58.0 generate

# Copy generated files to the right places
cp -rf ./build/proto/gocosmos/github.com/sei-protocol/sei-chain/* ./

rm -rf ./build/proto

echo "Protobuf code generation complete."