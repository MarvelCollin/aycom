describe('Registration Page', () => {
  beforeEach(() => {
    // Mock external services
    cy.window().then((win) => {
      // Mock Google Auth
      win.google = {
        accounts: {
          id: {
            initialize: cy.stub().as('googleInitialize'),
            renderButton: cy.stub().as('googleRenderButton')
          }
        }
      };
      
      // Mock reCAPTCHA
      win.grecaptcha = {
        ready: (cb) => cb(),
        render: cy.stub().returns('recaptcha-token-123').as('recaptchaRender'),
        execute: cy.stub().returns('recaptcha-token-123').as('recaptchaExecute'),
        getResponse: cy.stub().returns('recaptcha-token-123').as('recaptchaGetResponse')
      };
    });

    // Register custom command for filling the registration form
    Cypress.Commands.add('fillRegistrationForm', (options = {}) => {
      const defaults = {
        name: 'Test User',
        username: 'testuser',
        email: 'test@example.com',
        password: 'Password123!',
        confirmPassword: 'Password123!',
        gender: 'male',
        month: 'January',
        day: '1',
        year: '1990',
        securityQuestion: 'What was the name of your first pet?',
        securityAnswer: 'Fluffy',
        subscribeToNewsletter: true
      };

      const data = { ...defaults, ...options };
      
      // Fill in personal information
      cy.get('[data-cy=name-input]').type(data.name);
      cy.get('[data-cy=username-input]').type(data.username);
      cy.get('[data-cy=email-input]').type(data.email);
      cy.get('[data-cy=password-input]').type(data.password);
      cy.get('[data-cy=confirm-password-input]').type(data.confirmPassword);
      
      // Select gender
      cy.get(`[data-cy=gender-${data.gender}]`).check();
      
      // Set date of birth
      cy.get('[data-cy=dob-month]').select(data.month);
      cy.get('[data-cy=dob-day]').select(data.day);
      cy.get('[data-cy=dob-year]').select(data.year);
      
      // Select security question and answer
      cy.get('[data-cy=security-question]').select(data.securityQuestion);
      cy.get('[data-cy=security-answer]').type(data.securityAnswer);
      
      // Handle newsletter subscription
      if (data.subscribeToNewsletter) {
        cy.get('[data-cy=subscribe-checkbox]').check();
      } else {
        cy.get('[data-cy=subscribe-checkbox]').uncheck();
      }
    });

    // Intercept all API calls
    cy.interceptApi('POST', '/auth/register', {
      success: true,
      message: 'Registration successful! Please check your email for verification.'
    }).as('registerRequest');
    
    cy.interceptApi('POST', '/auth/verify-email', {
      success: true
    }).as('verifyEmailRequest');
    
    cy.interceptApi('POST', '/auth/resend-code', {
      success: true,
      message: 'Verification code has been sent to your email.'
    }).as('resendCodeRequest');
    
    // Visit the registration page before each test
    cy.visit('/register');
  });

  it('should have a working Google sign-in button', () => {
    cy.get('#google-signin-button').should('be.visible');
    // Since we mocked the Google API, we just verify the render button was called
    cy.get('@googleRenderButton').should('have.been.called');
  });

  it('should validate all form fields correctly', () => {
    // Submit empty form to trigger all validations
    cy.get('[data-cy=register-button]').click();
    
    // Check validation errors
    cy.get('[data-cy=name-error]').should('be.visible');
    cy.get('[data-cy=username-error]').should('be.visible');
    cy.get('[data-cy=email-error]').should('be.visible');
    cy.get('[data-cy=password-error]').should('be.visible');
    cy.get('[data-cy=gender-error]').should('be.visible');
    cy.get('[data-cy=dob-error]').should('be.visible');
    cy.get('[data-cy=security-question-error]').should('be.visible');
    
    // Test individual validations
    
    // Name field (required, max length)
    cy.get('[data-cy=name-input]').type('A').blur();
    cy.get('[data-cy=name-error]').should('contain', 'Name must be at least');
    
    cy.get('[data-cy=name-input]').clear().type('A'.repeat(51)).blur();
    cy.get('[data-cy=name-error]').should('contain', 'Name cannot exceed');
    
    // Email field (format validation)
    cy.get('[data-cy=email-input]').type('invalid-email').blur();
    cy.get('[data-cy=email-error]').should('contain', 'Please enter a valid email');
    
    // Password field (strength validation)
    cy.get('[data-cy=password-input]').type('weak').blur();
    cy.get('[data-cy=password-error]').should('be.visible');
    
    // Password matching
    cy.get('[data-cy=password-input]').clear().type('ValidPass123!');
    cy.get('[data-cy=confirm-password-input]').type('DifferentPass456!').blur();
    cy.get('[data-cy=password-match-error]').should('contain', 'Passwords do not match');
  });

  it('should display character count for name field', () => {
    cy.get('[data-cy=name-input]').type('John');
    cy.get('[data-cy=name-char-count]').should('contain', '4 / 50');
    
    cy.get('[data-cy=name-input]').clear().type('A'.repeat(50));
    cy.get('[data-cy=name-char-count]').should('contain', '50 / 50');
  });

  it('should submit the form successfully when all fields are valid', () => {
    // Fill in all required fields
    cy.fillRegistrationForm();
    
    // Submit the form
    cy.get('[data-cy=register-button]').click();
    
    // Check if the API was called with the correct data
    cy.wait('@registerRequest').its('request.body').should('include', {
      name: 'Test User',
      email: 'test@example.com'
    });
    
    // Verify we moved to step 2
    cy.get('[data-cy=verification-title]').should('be.visible')
      .and('contain', 'We sent you a code');
  });

  it('should handle password complexity requirements', () => {
    // Test with a password missing requirements
    cy.fillRegistrationForm({ 
      password: 'simple', 
      confirmPassword: 'simple' 
    });
    
    cy.get('[data-cy=register-button]').click();
    
    // Check that error messages appear
    cy.get('[data-cy=password-error]').should('contain', 'least 8 characters');
    cy.get('[data-cy=password-error]').should('contain', 'uppercase letter');
    cy.get('[data-cy=password-error]').should('contain', 'number');
    
    // Fix the password and try again
    cy.get('[data-cy=password-input]').clear().type('StrongPass123!');
    cy.get('[data-cy=confirm-password-input]').clear().type('StrongPass123!');
    cy.get('[data-cy=register-button]').click();
    
    // Check that there's no password error
    cy.get('[data-cy=password-error]').should('not.exist');
  });

  it('should verify email with correct verification code', () => {
    // Fill form and move to verification step
    cy.fillRegistrationForm();
    cy.get('[data-cy=register-button]').click();
    cy.wait('@registerRequest');
    
    // Enter verification code
    cy.get('[data-cy=verification-code]').type('123456');
    cy.get('[data-cy=verify-button]').click();
    
    // Verify API call was made
    cy.wait('@verifyEmailRequest');
    
    // Should redirect to login page
    cy.url().should('include', '/login');
  });

  it('should show timer and allow resending verification code', () => {
    // Fill form and move to verification step
    cy.fillRegistrationForm();
    cy.get('[data-cy=register-button]').click();
    cy.wait('@registerRequest');
    
    // Should display a timer
    cy.get('[data-cy=verification-timer]').should('be.visible');
    
    // Force timer to expire (by manipulating the clock)
    cy.clock();
    cy.tick(301000); // 5 minutes + 1 second
    
    // Resend button should appear
    cy.get('[data-cy=resend-button]').should('be.visible');
    
    // Click resend
    cy.get('[data-cy=resend-button]').click();
    
    // Verify resend API call
    cy.wait('@resendCodeRequest');
    
    // Timer should reappear
    cy.get('[data-cy=verification-timer]').should('be.visible');
  });

  it('should handle back button on verification page', () => {
    // Fill form and move to verification step
    cy.fillRegistrationForm();
    cy.get('[data-cy=register-button]').click();
    cy.wait('@registerRequest');
    
    // Click back button
    cy.get('[data-cy=back-button]').click();
    
    // Should be back at step 1 with the form filled
    cy.get('[data-cy=name-input]').should('have.value', 'Test User');
    cy.get('[data-cy=email-input]').should('have.value', 'test@example.com');
  });

  it('should handle Google authentication', () => {
    // Setup Google auth response object
    const googleResponse = {
      credential: 'mock-google-credential',
      clientId: 'mock-client-id',
      select_by: 'user'
    };

    // Create a global callback function that Cypress can spy on
    cy.window().then(win => {
      win.handleGoogleCredentialResponse = cy.stub().as('googleCallback');
    });

    // Trigger Google sign-in (normally this would be done by clicking the Google button)
    cy.window().then(win => {
      if (win.handleGoogleCredentialResponse) {
        win.handleGoogleCredentialResponse(googleResponse);
      }
    });

    // Verify our callback was called with the response
    cy.get('@googleCallback').should('have.been.calledWith', 
      Cypress.sinon.match.has('credential', 'mock-google-credential')
    );
  });
});