#!/bin/bash

# Set the directory to the proto files
PROTO_DIR="$(pwd)"

# Generate user service protobuf
echo "Generating user service protobuf..."
protoc -I=$PROTO_DIR \
  --go_out=$PROTO_DIR --go_opt=paths=source_relative \
  --go-grpc_out=$PROTO_DIR --go-grpc_opt=paths=source_relative \
  user/user.proto

echo "Proto generation complete!" 