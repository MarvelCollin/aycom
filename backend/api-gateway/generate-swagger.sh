#!/bin/bash

# Generate Swagger documentation
echo "Generating Swagger documentation..."
cd "$(dirname "$0")"
swag init -g main.go -o ./docs

# Convert to JSON format (for Docker usage)
echo "Converting to JSON format..."
cp ./docs/swagger.json ./docs/swagger.json

echo "Swagger documentation generated successfully!" 