#!/bin/sh

echo "Manually generating Swagger documentation..."

# Convert YAML to JSON for Swagger UI
if command -v python3 >/dev/null 2>&1; then
    # Using Python if available
    echo "Using Python to convert YAML to JSON..."
    python3 -c "import sys, yaml, json; json.dump(yaml.safe_load(open('./docs/swagger.yaml', 'r')), open('./docs/swagger.json', 'w'), indent=2)"
elif command -v yq >/dev/null 2>&1; then
    # Using yq if available
    echo "Using yq to convert YAML to JSON..."
    yq -o=json eval ./docs/swagger.yaml > ./docs/swagger.json
else
    echo "Error: Neither Python nor yq is available. Cannot convert YAML to JSON."
    echo "Please install either Python (with PyYAML) or yq to run this script."
    exit 1
fi

echo "Swagger documentation generated successfully at ./docs/swagger.json"
echo "Access the Swagger UI at: http://localhost:8083/swagger/index.html"
