# Protocol Buffers (Proto) Structure

This directory contains the Protocol Buffer definitions for all microservices in the AYCOM backend.

## Directory Structure

```
proto/
├── auth/              # Auth service definitions
│   └── auth.proto
├── community/         # Community service definitions 
│   └── community.proto
├── thread/            # Thread service definitions
│   └── thread.proto
├── user/              # User service definitions
│   └── user.proto
├── generate.bat       # Script to generate Go code from proto files
└── README.md          # This file
```

## Naming Conventions

1. **File Names**: Service proto files are named after the service (e.g., `auth.proto`)
2. **Go Package**: Proto files use a consistent pattern: `option go_package = "aycom/backend/proto/{service}"`
3. **Message Names**: Follow CamelCase naming (e.g., `UserResponse`)
4. **RPCs**: Follow CamelCase naming with no spaces between parameters (e.g., `GetUserById`)
5. **Field Names**: Use snake_case for field names (e.g., `user_id`)

## Code Generation

To generate Go code from these proto files, use the provided `generate.bat` script:

```bash
cd proto
generate.bat
```

## Cross-Service References

If a service needs to reference entities defined in another service's proto file, import the other service's proto file.

Example:
```protobuf
import "proto/user/user.proto";

message ThreadResponse {
  Thread thread = 1;
  user.User user = 2; // Reference to User from user service
}
```

## Best Practices

1. **Don't Repeat Yourself**: Avoid duplicating message definitions across services
2. **Backward Compatibility**: Follow proto3 conventions for field deprecation and additions
3. **Documentation**: Add comments for all services, RPCs, and messages
4. **Consistent Field Numbering**: Once deployed, never change field numbers 