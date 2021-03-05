# mage-loot

![ci](https://github.com/aserto-dev/mage-loot/workflows/ci/badge.svg?branch=main)
[![Coverage Status](https://coveralls.io/repos/github/aserto-dev/mage-loot/badge.svg?branch=main&t=4v6ABX&service=github)](https://coveralls.io/github/aserto-dev/mage-loot?branch=main)

Collection of reusable, useful code for magefiles



This is a sample `Depfile` that you might want to use - it contains all
dependencies for all helpers in `mage-loot`. 

```yaml
---
go:
  wire:
    importPath: "github.com/google/wire/cmd/wire"
    version: "v0.5.0"
  calc-version:
    importPath: "github.com/aserto-dev/calc-version"
    version: "v1.1.2"
  gotestsum:
    importPath: "gotest.tools/gotestsum"
    version: "v1.6.2"
  golangci-lint:
    importPath: "github.com/golangci/golangci-lint/cmd/golangci-lint"
    version: "v1.38.0"
  protoc-gen-go:
    importPath: "google.golang.org/protobuf/cmd/protoc-gen-go"
    version: "v1.25.0"
  protoc-gen-go-grpc:
    importPath: "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
    version: "v1.1.0"
  protoc-gen-grpc-gateway:
    importPath: "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway"
    version: "v2.3.0"
  protoc-gen-openapiv2:
    importPath: "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2"
    version: "v2.3.0"
  protoc-gen-doc:
    importPath: "github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc"
    version: "v1.4.1"
bin:
  protoc:
    url: "https://github.com/protocolbuffers/protobuf/releases/download/v{{.Version}}/protoc-{{.Version}}-linux-x86_64.zip"
    version: "3.15.4"
    sha: "14cca6414353c965ecf3c6bfc5aefb5b54cbd2f572b61aa67bf1ca435b086db9"
    zipPaths:
    - "bin/protoc"
lib:
  protoc-gen-openapiv2:
    url: "https://github.com/grpc-ecosystem/grpc-gateway/archive/v{{.Version}}.zip"
    version: "2.3.0"
    sha: "73d5f1f7373c4148a9934bff39afeb04c167895cc6b94e47075842b7e62c1f2e"
    libPrefix: "grpc-gateway-{{.Version}}"
    zipPaths:
    - "*/protoc-gen-openapiv2/options/*.proto"
  googleapis:
    url: "https://github.com/googleapis/googleapis/archive/{{.Version}}.zip"
    version: "8f117308d5bb55816953a0d6ad1a7d27a69a7d3f"
    sha: "103c32a32a994fa89565b8895697ae4bc987c50f58b2c7954d322a29429802d7"
    libPrefix: "googleapis-{{.Version}}"
    zipPaths:
    - "*/google/api/annotations.proto"
    - "*/google/api/field_behavior.proto"
    - "*/google/api/http.proto"
    - "*/google/api/httpbody.proto"
    - "*/google/rpc/code.proto"
    - "*/google/rpc/error_details.proto"
    - "*/google/rpc/status.proto"
  protobuf:
    url: "https://github.com/protocolbuffers/protobuf/archive/v{{.Version}}.zip"
    version: "3.15.5"
    sha: "f94faa42d49c0450226d1e9700ab5f5c3d8e5b757df41bc741bd304fd353eb63"
    libPrefix: "protobuf-{{.Version}}/src"
    zipPaths:
    - "*/src/google/protobuf/compiler/plugin.proto"
    - "*/src/google/protobuf/any.proto"
    - "*/src/google/protobuf/api.proto"
    - "*/src/google/protobuf/descriptor.proto"
    - "*/src/google/protobuf/duration.proto"
    - "*/src/google/protobuf/empty.proto"
    - "*/src/google/protobuf/field_mask.proto"
    - "*/src/google/protobuf/source_context.proto"
    - "*/src/google/protobuf/struct.proto"
    - "*/src/google/protobuf/timestamp.proto"
    - "*/src/google/protobuf/type.proto"
    - "*/src/google/protobuf/wrappers.proto"

```