# API Gateway Unit Tests

This directory contains unit tests for various components of the API Gateway using the Testify framework.

## Test Files

1. **string_similarity_test.go** - Tests for string similarity comparison functions
2. **token_manager_test.go** - Tests for token generation and validation
3. **auth_login_test.go** - Tests for authentication/login functionality

## Running the Tests

### Running All Tests

To run all tests in this directory:

```bash
cd backend/api-gateway
go test -v ./test/...
```

### Running Specific Tests

To run a specific test file:

```bash
go test -v ./test/string_similarity_test.go
go test -v ./test/token_manager_test.go
go test -v ./test/auth_login_test.go
```

To run a specific test function:

```bash
go test -v ./test -run TestDamerauLevenshteinDistance
go test -v ./test -run TestTokenManager
go test -v ./test -run TestLogin
```

## Test Dependencies

Make sure you have the Testify package installed:

```bash
go get github.com/stretchr/testify/assert
```

## Notes

- The login test uses the specified credentials (kolina@gmail.com / Miawmiaw123@) but will likely fail without a connection to the actual services.
- These tests are simplified examples and may need to be extended for production use. 