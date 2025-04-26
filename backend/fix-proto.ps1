# PowerShell script to ensure proper proto directories are created
# and proto files are copied to the correct locations

# Move to the script's directory
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptDir

# Create proto directories for each service
New-Item -Path "services/auth/proto" -ItemType Directory -Force | Out-Null
New-Item -Path "services/user/proto" -ItemType Directory -Force | Out-Null
New-Item -Path "services/product/proto" -ItemType Directory -Force | Out-Null

# Copy the proto files to service directories
Copy-Item -Path "proto/auth.proto" -Destination "services/auth/proto/" -Force
Copy-Item -Path "proto/user.proto" -Destination "services/user/proto/" -Force
Copy-Item -Path "proto/product.proto" -Destination "services/product/proto/" -Force

Write-Host "Proto files copied to service directories"

# Check if protoc is available
if (Get-Command "protoc" -ErrorAction SilentlyContinue) {
    Write-Host "Generating Go stubs from proto files..."
    
    # Auth
    & protoc --go_out=. --go_opt=paths=source_relative `
        --go-grpc_out=. --go-grpc_opt=paths=source_relative `
        services/auth/proto/auth.proto
    
    # User
    & protoc --go_out=. --go_opt=paths=source_relative `
        --go-grpc_out=. --go-grpc_opt=paths=source_relative `
        services/user/proto/user.proto
    
    # Product
    & protoc --go_out=. --go_opt=paths=source_relative `
        --go-grpc_out=. --go-grpc_opt=paths=source_relative `
        services/product/proto/product.proto
    
    Write-Host "Go stubs generated successfully"
}
else {
    Write-Host "Warning: protoc not found. Go stubs not generated."
    Write-Host "Install protoc and the Go plugins to generate Go stubs from proto files."
}

Write-Host "Proto directory setup complete" 