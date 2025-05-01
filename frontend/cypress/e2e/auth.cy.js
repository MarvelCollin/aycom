// Auth E2E tests covering Login and Registration

describe('Authentication', () => {
  // Login Tests
  describe('Login Page', () => {
    beforeEach(() => {
      cy.visit('/login');
    });

    it('should display the login page correctly', () => {
      cy.contains('h1', 'Sign in to AYCOM');
      cy.get('[data-cy="google-login-button"]').should('exist');
      cy.get('[data-cy="email-input"]').should('exist');
      cy.get('[data-cy="password-input"]').should('exist');
      cy.get('[data-cy="remember-me"]').should('exist');
      cy.get('[data-cy="login-button"]').should('exist');
      cy.get('[data-cy="forgot-password"]').should('exist');
      cy.get('[data-cy="register-link"]').should('exist');
    });

    it('should validate required fields', () => {
      cy.get('[data-cy="login-button"]').click();
      cy.get('[data-cy="error-message"]').should('be.visible');
    });

    it('should show error for invalid credentials', () => {
      cy.get('[data-cy="email-input"]').type('invalid@example.com');
      cy.get('[data-cy="password-input"]').type('wrongpassword');
      cy.get('[data-cy="login-button"]').click();
      cy.get('[data-cy="error-message"]').should('be.visible');
    });

    it('should redirect to forgot password page', () => {
      cy.get('[data-cy="forgot-password"]').click();
      cy.url().should('include', '/forgot-password');
    });

    it('should redirect to register page', () => {
      cy.get('[data-cy="register-link"]').click();
      cy.url().should('include', '/register');
    });

    it('should login successfully and redirect to feed', () => {
      // Intercept the login API call to mock a successful response
      cy.intercept('POST', '**/api/v1/auth/login', {
        statusCode: 200,
        body: {
          success: true,
          token: 'fake-jwt-token',
          user: {
            id: '1',
            name: 'Test User',
            email: 'test@example.com'
          }
        }
      }).as('loginRequest');

      cy.get('[data-cy="email-input"]').type('test@example.com');
      cy.get('[data-cy="password-input"]').type('Password123!');
      cy.get('[data-cy="login-button"]').click();
      
      cy.wait('@loginRequest');
      cy.url().should('include', '/feed');
    });

    it('should not allow access for banned or deactivated accounts', () => {
      // Intercept the login API call to mock a banned account response
      cy.intercept('POST', '**/api/v1/auth/login', {
        statusCode: 403,
        body: {
          success: false,
          message: 'This account has been banned or deactivated'
        }
      }).as('bannedAccountRequest');

      cy.get('[data-cy="email-input"]').type('banned@example.com');
      cy.get('[data-cy="password-input"]').type('Password123!');
      cy.get('[data-cy="login-button"]').click();
      
      cy.wait('@bannedAccountRequest');
      cy.get('[data-cy="error-message"]').should('contain', 'banned or deactivated');
    });

    it('should handle Google Sign-In button click', () => {
      // Since we can't fully test Google auth in E2E, we can verify the button interaction
      cy.window().then(win => {
        cy.stub(win, 'open').as('windowOpen');
      });
      
      cy.get('[data-cy="google-login-button"]').find('button').click();
      // Verify that the OAuth window would have been opened
      cy.get('@windowOpen').should('have.been.called');
    });
  });

  // Registration Tests
  describe('Registration Page - Step 1', () => {
    beforeEach(() => {
      cy.visit('/register');
    });

    it('should display the registration page correctly', () => {
      cy.contains('h1', 'Create your account');
      cy.get('[data-cy="google-login-button"]').should('exist');
      // Verify form fields exist
      cy.get('[data-cy="name-input"]').should('exist');
      cy.get('[data-cy="username-input"]').should('exist');
      cy.get('[data-cy="email-input"]').should('exist');
      cy.get('[data-cy="password-input"]').should('exist');
      cy.get('[data-cy="confirm-password-input"]').should('exist');
      cy.get('[data-cy="gender-input"]').should('exist');
      cy.get('[data-cy="dob-month"]').should('exist');
      cy.get('[data-cy="dob-day"]').should('exist');
      cy.get('[data-cy="dob-year"]').should('exist');
      cy.get('[data-cy="security-question"]').should('exist');
      cy.get('[data-cy="security-answer"]').should('exist');
      cy.get('[data-cy="subscribe-newsletter"]').should('exist');
      cy.get('[data-cy="profile-picture-input"]').should('exist');
      cy.get('[data-cy="banner-input"]').should('exist');
      cy.get('[data-cy="register-button"]').should('exist');
      cy.get('[data-cy="login-link"]').should('exist');
    });

    it('should validate name field', () => {
      // Test too short
      cy.get('[data-cy="name-input"]').type('Ab').blur();
      cy.get('[data-cy="name-error"]').should('be.visible');
      
      // Test with numbers
      cy.get('[data-cy="name-input"]').clear().type('John123').blur();
      cy.get('[data-cy="name-error"]').should('be.visible');
      
      // Test with symbols
      cy.get('[data-cy="name-input"]').clear().type('John@Smith').blur();
      cy.get('[data-cy="name-error"]').should('be.visible');
      
      // Test valid name
      cy.get('[data-cy="name-input"]').clear().type('John Smith').blur();
      cy.get('[data-cy="name-error"]').should('not.exist');
    });

    it('should validate username field uniqueness', () => {
      // Mock an API call that checks username uniqueness
      cy.intercept('GET', '**/api/v1/auth/check-username?username=taken_username', {
        statusCode: 200,
        body: { available: false }
      }).as('checkTakenUsername');
      
      cy.intercept('GET', '**/api/v1/auth/check-username?username=available_username', {
        statusCode: 200,
        body: { available: true }
      }).as('checkAvailableUsername');
      
      // Test taken username
      cy.get('[data-cy="username-input"]').type('taken_username').blur();
      cy.wait('@checkTakenUsername');
      cy.get('[data-cy="username-error"]').should('be.visible');
      
      // Test available username
      cy.get('[data-cy="username-input"]').clear().type('available_username').blur();
      cy.wait('@checkAvailableUsername');
      cy.get('[data-cy="username-error"]').should('not.exist');
    });

    it('should validate email format and uniqueness', () => {
      // Invalid format
      cy.get('[data-cy="email-input"]').type('notanemail').blur();
      cy.get('[data-cy="email-error"]').should('be.visible');
      
      // Mock an API call that checks email uniqueness
      cy.intercept('GET', '**/api/v1/auth/check-email?email=taken@gmail.com', {
        statusCode: 200,
        body: { available: false }
      }).as('checkTakenEmail');
      
      cy.intercept('GET', '**/api/v1/auth/check-email?email=available@gmail.com', {
        statusCode: 200,
        body: { available: true }
      }).as('checkAvailableEmail');
      
      // Test taken email
      cy.get('[data-cy="email-input"]').clear().type('taken@gmail.com').blur();
      cy.wait('@checkTakenEmail');
      cy.get('[data-cy="email-error"]').should('be.visible');
      
      // Test available email
      cy.get('[data-cy="email-input"]').clear().type('available@gmail.com').blur();
      cy.wait('@checkAvailableEmail');
      cy.get('[data-cy="email-error"]').should('not.exist');
    });

    it('should validate password complexity', () => {
      // Too short
      cy.get('[data-cy="password-input"]').type('short').blur();
      cy.get('[data-cy="password-error"]').should('be.visible');
      
      // No uppercase
      cy.get('[data-cy="password-input"]').clear().type('password123!').blur();
      cy.get('[data-cy="password-error"]').should('be.visible');
      
      // No lowercase
      cy.get('[data-cy="password-input"]').clear().type('PASSWORD123!').blur();
      cy.get('[data-cy="password-error"]').should('be.visible');
      
      // No number
      cy.get('[data-cy="password-input"]').clear().type('Password!').blur();
      cy.get('[data-cy="password-error"]').should('be.visible');
      
      // No special character
      cy.get('[data-cy="password-input"]').clear().type('Password123').blur();
      cy.get('[data-cy="password-error"]').should('be.visible');
      
      // Valid password
      cy.get('[data-cy="password-input"]').clear().type('Password123!').blur();
      cy.get('[data-cy="password-error"]').should('not.exist');
    });

    it('should validate password confirmation', () => {
      cy.get('[data-cy="password-input"]').type('Password123!');
      cy.get('[data-cy="confirm-password-input"]').type('DifferentPassword123!').blur();
      cy.get('[data-cy="confirm-password-error"]').should('be.visible');
      
      cy.get('[data-cy="confirm-password-input"]').clear().type('Password123!').blur();
      cy.get('[data-cy="confirm-password-error"]').should('not.exist');
    });

    it('should validate gender selection', () => {
      cy.get('[data-cy="gender-input"]').select('').blur();
      cy.get('[data-cy="gender-error"]').should('be.visible');
      
      cy.get('[data-cy="gender-input"]').select('Male').blur();
      cy.get('[data-cy="gender-error"]').should('not.exist');
    });

    it('should validate age requirement (13+ years)', () => {
      const currentYear = new Date().getFullYear();
      const tooYoungYear = currentYear - 12; // 12 years old (too young)
      const oldEnoughYear = currentYear - 13; // 13 years old (old enough)
      
      // Too young
      cy.get('[data-cy="dob-month"]').select('January');
      cy.get('[data-cy="dob-day"]').select('1');
      cy.get('[data-cy="dob-year"]').select(tooYoungYear.toString());
      cy.get('[data-cy="dob-year"]').blur();
      cy.get('[data-cy="dob-error"]').should('be.visible');
      
      // Old enough
      cy.get('[data-cy="dob-month"]').select('January');
      cy.get('[data-cy="dob-day"]').select('1');
      cy.get('[data-cy="dob-year"]').select(oldEnoughYear.toString());
      cy.get('[data-cy="dob-year"]').blur();
      cy.get('[data-cy="dob-error"]').should('not.exist');
    });

    it('should validate security question selection and answer', () => {
      cy.get('[data-cy="security-question"]').select('').blur();
      cy.get('[data-cy="security-question-error"]').should('be.visible');
      
      cy.get('[data-cy="security-question"]').select('What was the name of your first pet?').blur();
      cy.get('[data-cy="security-question-error"]').should('not.exist');
      
      cy.get('[data-cy="security-answer"]').type('Fluffy').blur();
      cy.get('[data-cy="security-answer-error"]').should('not.exist');
    });

    it('should handle file uploads for profile picture and banner', () => {
      // Stub file upload
      cy.fixture('test-profile.jpg', 'base64').then(fileContent => {
        cy.get('[data-cy="profile-picture-input"]').attachFile({
          fileContent,
          fileName: 'test-profile.jpg',
          mimeType: 'image/jpeg'
        });
      });
      
      cy.fixture('test-banner.jpg', 'base64').then(fileContent => {
        cy.get('[data-cy="banner-input"]').attachFile({
          fileContent,
          fileName: 'test-banner.jpg',
          mimeType: 'image/jpeg'
        });
      });
    });

    it('should verify reCAPTCHA implementation', () => {
      // Since we can't fully test reCAPTCHA in E2E, we can verify it exists
      cy.get('[data-cy="g-recaptcha"]').should('exist');
      
      // Mock successful reCAPTCHA verification
      cy.window().then(win => {
        win.formData = win.formData || {};
        win.formData.recaptchaToken = 'mock-recaptcha-token';
      });
    });

    it('should handle form submission and proceed to verification step', () => {
      // Fill out the entire form with valid data
      cy.get('[data-cy="name-input"]').type('John Smith');
      cy.get('[data-cy="username-input"]').type('johnsmith');
      cy.get('[data-cy="email-input"]').type('johnsmith@gmail.com');
      cy.get('[data-cy="password-input"]').type('Password123!');
      cy.get('[data-cy="confirm-password-input"]').type('Password123!');
      cy.get('[data-cy="gender-input"]').select('Male');
      
      // Set DOB to 18 years ago
      const currentYear = new Date().getFullYear();
      cy.get('[data-cy="dob-month"]').select('January');
      cy.get('[data-cy="dob-day"]').select('1');
      cy.get('[data-cy="dob-year"]').select((currentYear - 18).toString());
      
      cy.get('[data-cy="security-question"]').select('What was the name of your first pet?');
      cy.get('[data-cy="security-answer"]').type('Fluffy');
      cy.get('[data-cy="subscribe-newsletter"]').check();
      
      // Mock file uploads
      cy.fixture('test-profile.jpg', 'base64').then(fileContent => {
        cy.get('[data-cy="profile-picture-input"]').attachFile({
          fileContent,
          fileName: 'test-profile.jpg',
          mimeType: 'image/jpeg'
        });
      });
      
      cy.fixture('test-banner.jpg', 'base64').then(fileContent => {
        cy.get('[data-cy="banner-input"]').attachFile({
          fileContent,
          fileName: 'test-banner.jpg',
          mimeType: 'image/jpeg'
        });
      });
      
      // Mock reCAPTCHA token
      cy.window().then(win => {
        win.formData = win.formData || {};
        win.formData.recaptchaToken = 'mock-recaptcha-token';
      });
      
      // Intercept registration API call
      cy.intercept('POST', '**/api/v1/auth/register', {
        statusCode: 200,
        body: {
          success: true,
          message: 'Verification code sent to email'
        }
      }).as('registerRequest');
      
      // Submit the form
      cy.get('[data-cy="register-button"]').click();
      cy.wait('@registerRequest');
      
      // Check redirection to verification step
      cy.contains('h1', 'We sent you a code').should('be.visible');
      cy.get('[data-cy="verification-code-input"]').should('exist');
    });
  });

  describe('Registration Page - Step 2 (Verification)', () => {
    beforeEach(() => {
      // Setup: Register first, then go to verification step
      cy.visit('/register');
      
      // Mock registration API
      cy.intercept('POST', '**/api/v1/auth/register', {
        statusCode: 200,
        body: {
          success: true,
          message: 'Verification code sent to email'
        }
      }).as('registerRequest');
      
      // Fill minimal data and submit to reach step 2
      cy.window().then(win => {
        win.formData = win.formData || {};
        win.formData.recaptchaToken = 'mock-recaptcha-token';
      });
      
      cy.get('[data-cy="name-input"]').type('John Smith');
      cy.get('[data-cy="username-input"]').type('johnsmith');
      cy.get('[data-cy="email-input"]').type('johnsmith@gmail.com');
      cy.get('[data-cy="password-input"]').type('Password123!');
      cy.get('[data-cy="confirm-password-input"]').type('Password123!');
      cy.get('[data-cy="gender-input"]').select('Male');
      
      // Set DOB to 18 years ago
      const currentYear = new Date().getFullYear();
      cy.get('[data-cy="dob-month"]').select('January');
      cy.get('[data-cy="dob-day"]').select('1');
      cy.get('[data-cy="dob-year"]').select((currentYear - 18).toString());
      
      cy.get('[data-cy="security-question"]').select('What was the name of your first pet?');
      cy.get('[data-cy="security-answer"]').type('Fluffy');
      
      cy.get('[data-cy="register-button"]').click();
      cy.wait('@registerRequest');
      
      // Now we should be on the verification step
      cy.contains('h1', 'We sent you a code').should('be.visible');
    });

    it('should validate the verification code input', () => {
      cy.get('[data-cy="verify-button"]').click();
      cy.get('[data-cy="error-message"]').should('be.visible');
    });

    it('should handle successful verification', () => {
      // Intercept verification API call
      cy.intercept('POST', '**/api/v1/auth/verify', {
        statusCode: 200,
        body: {
          success: true,
          message: 'Email verified successfully'
        }
      }).as('verifyRequest');
      
      cy.get('[data-cy="verification-code-input"]').type('123456');
      cy.get('[data-cy="verify-button"]').click();
      cy.wait('@verifyRequest');
      
      // Should redirect to login page
      cy.url().should('include', '/login');
    });

    it('should handle invalid verification code', () => {
      // Intercept verification API call with error
      cy.intercept('POST', '**/api/v1/auth/verify', {
        statusCode: 400,
        body: {
          success: false,
          message: 'Invalid verification code'
        }
      }).as('verifyRequest');
      
      cy.get('[data-cy="verification-code-input"]').type('000000');
      cy.get('[data-cy="verify-button"]').click();
      cy.wait('@verifyRequest');
      
      cy.get('[data-cy="error-message"]').should('be.visible');
    });

    it('should handle expired verification code', () => {
      // Intercept verification API call with expired error
      cy.intercept('POST', '**/api/v1/auth/verify', {
        statusCode: 400,
        body: {
          success: false,
          message: 'Verification code has expired'
        }
      }).as('verifyRequest');
      
      cy.get('[data-cy="verification-code-input"]').type('123456');
      cy.get('[data-cy="verify-button"]').click();
      cy.wait('@verifyRequest');
      
      cy.get('[data-cy="error-message"]').should('contain', 'expired');
    });

    it('should allow resending verification code', () => {
      // Test that timer shows initially
      cy.get('[data-cy="resend-timer"]').should('be.visible');
      
      // Wait for timer to expire and resend button to appear
      // For testing, we can force the timer to expire
      cy.window().then(win => {
        win.formState = win.formState || {};
        win.formState.showResendOption = true;
      });
      
      // Intercept resend verification API call
      cy.intercept('POST', '**/api/v1/auth/resend-verification', {
        statusCode: 200,
        body: {
          success: true,
          message: 'Verification code resent'
        }
      }).as('resendRequest');
      
      cy.get('[data-cy="resend-button"]').should('be.visible').click();
      cy.wait('@resendRequest');
      
      // Timer should be visible again
      cy.get('[data-cy="resend-timer"]').should('be.visible');
    });

    it('should allow going back to step 1', () => {
      cy.get('[data-cy="back-button"]').click();
      cy.contains('h1', 'Create your account').should('be.visible');
    });
  });
}); 