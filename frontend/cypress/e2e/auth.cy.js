describe('Authentication', () => {
  beforeEach(() => {
    // Clear cookies and localStorage between tests
    cy.clearCookies();
    cy.clearLocalStorage();
  });

  describe('Login', () => {
    it('should successfully login with valid credentials', () => {
      // Intercept the auth API call
      cy.interceptApi('POST', '/auth/login', {
        success: true,
        access_token: 'fake-access-token',
        refresh_token: 'fake-refresh-token'
      }).as('loginRequest');
      
      cy.visit('/login');
      cy.get('[data-cy=email-input]').type('test@example.com');
      cy.get('[data-cy=password-input]').type('password123');
      cy.get('[data-cy=login-button]').click();
      
      // Wait for API call to complete
      cy.wait('@loginRequest');
      
      // Verify redirection to feed page
      cy.url().should('include', '/feed');
    });
    
    it('should display error with invalid credentials', () => {
      // Intercept the auth API call with error
      cy.interceptApi('POST', '/auth/login', {
        statusCode: 401,
        body: {
          success: false,
          message: 'Invalid email or password'
        }
      }).as('loginRequest');
      
      cy.visit('/login');
      cy.get('[data-cy=email-input]').type('wrong@example.com');
      cy.get('[data-cy=password-input]').type('wrongpassword');
      cy.get('[data-cy=login-button]').click();
      
      // Wait for API call to complete
      cy.wait('@loginRequest');
      
      // Check that error message is displayed
      cy.get('[data-cy=error-message]').should('be.visible');
      cy.contains('Invalid email or password').should('be.visible');
      
      // Verify we're still on the login page
      cy.url().should('include', '/login');
    });
    
    it('should validate form fields', () => {
      cy.visit('/login');
      
      // Try to submit empty form
      cy.get('[data-cy=login-button]').click();
      
      // Check validation errors
      cy.get('[data-cy=email-error]').should('be.visible');
      cy.get('[data-cy=password-error]').should('be.visible');
    });
  });
  
  describe('Registration', () => {
    it('should successfully register a new account', () => {
      // Intercept the register API call
      cy.interceptApi('POST', '/auth/register', {
        success: true,
        message: 'Registration successful'
      }).as('registerRequest');
      
      cy.visit('/register');
      cy.get('[data-cy=name-input]').type('Test User');
      cy.get('[data-cy=email-input]').type('newuser@example.com');
      cy.get('[data-cy=password-input]').type('securePassword123');
      cy.get('[data-cy=confirm-password-input]').type('securePassword123');
      cy.get('[data-cy=register-button]').click();
      
      // Wait for API call to complete
      cy.wait('@registerRequest');
      
      // Verify redirection to login page or verification page
      cy.url().should('include', '/login');
    });
    
    it('should validate matching passwords', () => {
      cy.visit('/register');
      cy.get('[data-cy=name-input]').type('Test User');
      cy.get('[data-cy=email-input]').type('test@example.com');
      cy.get('[data-cy=password-input]').type('password123');
      cy.get('[data-cy=confirm-password-input]').type('different-password');
      cy.get('[data-cy=register-button]').click();
      
      // Check validation error for password match
      cy.get('[data-cy=password-match-error]').should('be.visible');
      cy.contains('Passwords do not match').should('be.visible');
    });
    
    it('should validate required fields', () => {
      cy.visit('/register');
      cy.get('[data-cy=register-button]').click();
      
      // Check validation errors for all required fields
      cy.get('[data-cy=name-error]').should('be.visible');
      cy.get('[data-cy=email-error]').should('be.visible');
      cy.get('[data-cy=password-error]').should('be.visible');
    });
  });
  
  describe('Google Authentication', () => {
    it('should handle Google login button click', () => {
      cy.visit('/login');
      // We can't fully test OAuth without mocking, but we can check button works
      cy.get('[data-cy=google-login-button]').should('be.visible').click();
      // Additional assertions would depend on how your app handles Google auth
    });
  });
});