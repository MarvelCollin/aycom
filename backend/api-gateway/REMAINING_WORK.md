# API Standardization - Remaining Work

This document outlines the remaining tasks needed to complete the API standardization effort across the AYCOM backend.

## Remaining Handler Files

The following handler files still need to be fully reviewed and updated:

1. `user_handlers.go` - Some direct `c.JSON()` calls remain
2. `thread_handlers.go` - May contain inconsistent field naming
3. `community_handlers.go` - Needs review for consistent error codes
4. `social_handlers.go` - Large file that needs thorough review
5. `bookmark_handlers.go` - Check for consistent pagination
6. `admin_handlers.go` - Ensure admin-specific responses follow conventions
7. `auth_handlers.go` - Authentication responses need standardization

## WebSocket Message Standardization

WebSocket message payloads need to be standardized:

1. `notification_websocket_handlers.go`:
   - Ensure notification event payloads use snake_case
   - Standardize event type naming

2. `chat_websocket_handlers.go`:
   - Standardize message payload format
   - Ensure consistent field naming in chat events

## Service Client Files

Review service client files for consistent field mapping between gRPC/protobuf and JSON:

1. `user_service_clients.go`
2. `thread_service_clients.go`
3. `community_service_clients.go`

These files often transform protobuf messages to JSON responses and need to ensure they follow the naming conventions.

## Proto File Field Inconsistencies

Some proto files may define fields with inconsistent naming. While changing proto files requires more careful coordination due to backward compatibility concerns, we should:

1. Document inconsistencies between proto field names and API response field names
2. Plan for future proto file updates to align with API naming conventions
3. Ensure service clients properly transform proto fields to follow API conventions

## Request Payload Standardization

While we've focused on response standardization, request payload field naming should also be standardized:

1. Review all request binding structs for consistent field naming
2. Document request payload conventions in NAMING_CONVENTIONS.md
3. Update validation error messages to reference correct field names

## Testing

Comprehensive testing is needed to ensure standardization doesn't break existing functionality:

1. Create API tests that verify response format compliance
2. Test pagination across different endpoints
3. Verify error responses follow the standard format
4. Test WebSocket message format compliance

## Documentation Updates

1. Update API documentation to reflect standardized field names
2. Create examples of standard request/response formats for different entity types
3. Document error codes and their meanings

## Next Steps

1. Prioritize WebSocket message standardization as it affects real-time features
2. Address user_handlers.go and auth_handlers.go next as they are critical paths
3. Create a test plan to verify standardization doesn't break existing functionality 