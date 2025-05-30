# Backend API Field Naming Convention

## Standards Enforced:
- All JSON response fields use `snake_case` format
- All Go struct fields use `CamelCase` format 
- JSON field names in struct tags use `snake_case` format

## Fixed Issues:

### Invalid JSON tag options:
Go struct tags don't support multiple field name options like: `json:"field_name,fieldName"`. We've fixed this in:

1. `api-gateway/models/response_models.go`:
   - Fixed `HasMore` field in `Pagination` struct

2. `api-gateway/models/models.go`:
   - Fixed `SecurityQuestion`, `SecurityAnswer`, and `SubscribeToNewsletter` in `RegisterRequest`
   - Fixed `ProfilePictureUrl` and `BannerUrl` in `RegisterRequest`

3. `api-gateway/handlers/user_handlers.go`:
   - Fixed `DisplayName`, `DateOfBirth`, `ProfilePicture`, `BackgroundBanner` fields
   - Fixed `ProfilePictureUrl` in `UpdateProfilePictureURLHandler`
   - Fixed `BannerUrl` in `UpdateBannerURLHandler`

4. `api-gateway/handlers/thread_handlers.go`:
   - Fixed `MediaUrls` in `UpdateThreadMediaURLsHandler`

### Backward Compatibility:
To maintain backward compatibility with clients that might be using camelCase fields, we've taken the following approach:

1. All API response fields use snake_case (via gin.H or struct JSON tags)
2. API handlers accept the primary snake_case version in request structs
3. The application code manually handles fallbacks:
   ```go
   name := input.Name
   if name == "" && input.DisplayName != "" {
       name = input.DisplayName
   }
   ```

## Future Improvements:
- Consider adding middleware that automatically transforms camelCase to snake_case in responses
- Consider adding middleware that can parse both camelCase and snake_case in requests
- Update GRPC client initialization to remove deprecated API warnings
- Update documentation to explicitly standardize on snake_case for API interfaces 