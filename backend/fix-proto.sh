#!/bin/bash

# This script ensures proper proto directories are created
# and proto files are copied to the correct locations

set -e

cd $(dirname $0)

# Create proto directories for each service
mkdir -p services/auth/proto
mkdir -p services/user/proto
mkdir -p services/product/proto

# Copy the proto files to service directories
cp proto/auth.proto services/auth/proto/
cp proto/user.proto services/user/proto/
cp proto/product.proto services/product/proto/

echo "Proto files copied to service directories"

# If protoc is available, generate Go stubs
if command -v protoc &> /dev/null; then
    echo "Generating Go stubs from proto files..."
    
    # Auth
    protoc --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        services/auth/proto/auth.proto
    
    # User
    protoc --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        services/user/proto/user.proto
    
    # Product
    protoc --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        services/product/proto/product.proto
    
    echo "Go stubs generated successfully"
else
    echo "Warning: protoc not found. Go stubs not generated."
    echo "Install protoc and the Go plugins to generate Go stubs from proto files."
fi

echo "Proto directory setup complete" 