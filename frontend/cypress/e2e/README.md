# Authentication E2E Tests

This directory contains end-to-end tests for the authentication flows in the AYCOM application.

## Test Files

- **auth.cy.js**: Tests for login and registration processes including validation, Google auth integration, and multi-step flows.

## Running Tests

You can run the authentication tests using the following commands:

- **Run in headless mode**: `npm run test:auth`
- **Run with UI**: `npm run test:ui` and select the auth.cy.js file

## Test Coverage

The authentication tests cover:

### Login Page
- Basic UI elements validation
- Form validation
- Error handling for invalid credentials
- Success flow with proper redirection
- Google Sign-In integration
- Redirection to registration and forgot password

### Registration Page - Step 1
- UI elements validation
- Form field validation for:
  - Name (4+ characters, no symbols/numbers)
  - Username (uniqueness check)
  - Email (format validation and uniqueness)
  - Password (multiple complexity validations)
  - Password confirmation
  - Gender selection
  - Age validation (13+ years)
  - Security question selection
- File uploads for profile picture and banner
- reCAPTCHA integration
- Proceeding to verification step

### Registration Page - Step 2 (Verification)
- Verification code validation
- Success flow with account activation
- Error handling for invalid codes
- Resending verification code
- Timer functionality for code expiration

## Notes for Developers

- The tests use mock API responses for external services and API calls
- File uploads are handled using the `cypress-file-upload` plugin
- Google authentication is simulated rather than actually integrating with Google OAuth

## Adding New Tests

When adding new authentication-related tests:
1. Add them to the existing auth.cy.js file in the appropriate describe block
2. Ensure all UI elements have data-cy attributes for reliable selection
3. Use the existing mocking patterns for API calls
4. Run the tests to verify they work as expected 