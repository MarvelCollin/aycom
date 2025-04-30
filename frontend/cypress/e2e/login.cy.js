describe('Login Page', () => {
  beforeEach(() => {
    // Clear cookies and localStorage between tests
    cy.clearCookies();
    cy.clearLocalStorage();
    // Visit the login page before each test
    cy.visit('/login');
  });

  it('should have the correct UI elements', () => {
    // Check for page title
    cy.get('[data-cy=page-title]').should('contain', 'Login');
    
    // Check for form elements
    cy.get('[data-cy=email-input]').should('be.visible');
    cy.get('[data-cy=password-input]').should('be.visible');
    cy.get('[data-cy=login-button]').should('be.visible');
    
    // Check for "Remember me" checkbox
    cy.get('[data-cy=remember-me]').should('be.visible');
    
    // Check for "Forgot password" link
    cy.get('[data-cy=forgot-password]').should('be.visible');
    
    // Check for "Register" link
    cy.get('[data-cy=register-link]').should('be.visible');
    
    // Check for social login options
    cy.get('[data-cy=google-login-button]').should('be.visible');
  });

  it('should successfully login with valid credentials', () => {
    // Intercept the auth API call
    cy.interceptApi('POST', '/auth/login', {
      success: true,
      access_token: 'fake-access-token',
      refresh_token: 'fake-refresh-token'
    }).as('loginRequest');
    
    // Enter credentials
    cy.get('[data-cy=email-input]').type('test@example.com');
    cy.get('[data-cy=password-input]').type('password123');
    cy.get('[data-cy=login-button]').click();
    
    // Wait for API call to complete
    cy.wait('@loginRequest');
    
    // Verify tokens are stored in localStorage
    cy.window().its('localStorage')
      .invoke('getItem', 'accessToken')
      .should('equal', 'fake-access-token');
    
    cy.window().its('localStorage')
      .invoke('getItem', 'refreshToken')
      .should('equal', 'fake-refresh-token');
    
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
    
    // Enter wrong credentials
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
  
  it('should validate form fields before submission', () => {
    // Try to submit empty form
    cy.get('[data-cy=login-button]').click();
    
    // Check validation errors
    cy.get('[data-cy=email-error]').should('be.visible');
    cy.get('[data-cy=password-error]').should('be.visible');
    
    // Fix email but leave password empty
    cy.get('[data-cy=email-input]').type('valid@example.com');
    cy.get('[data-cy=login-button]').click();
    
    // Should still show password error
    cy.get('[data-cy=email-error]').should('not.exist');
    cy.get('[data-cy=password-error]').should('be.visible');
    
    // Clear email and type password
    cy.get('[data-cy=email-input]').clear();
    cy.get('[data-cy=password-input]').type('password123');
    cy.get('[data-cy=login-button]').click();
    
    // Should now show email error
    cy.get('[data-cy=email-error]').should('be.visible');
    cy.get('[data-cy=password-error]').should('not.exist');
  });
  
  it('should validate email format', () => {
    // Test invalid email formats
    cy.get('[data-cy=email-input]').type('invalid-email').blur();
    cy.get('[data-cy=email-error]').should('be.visible');
    
    cy.get('[data-cy=email-input]').clear().type('user@').blur();
    cy.get('[data-cy=email-error]').should('be.visible');
    
    cy.get('[data-cy=email-input]').clear().type('user@domain').blur();
    cy.get('[data-cy=email-error]').should('be.visible');
    
    // Test valid email format
    cy.get('[data-cy=email-input]').clear().type('user@domain.com').blur();
    cy.get('[data-cy=email-error]').should('not.exist');
  });

  it('should remember user when "Remember me" is checked', () => {
    // Intercept the auth API call
    cy.interceptApi('POST', '/auth/login', {
      success: true,
      access_token: 'fake-access-token',
      refresh_token: 'fake-refresh-token'
    }).as('loginRequest');
    
    // Enter credentials and check remember me
    cy.get('[data-cy=email-input]').type('test@example.com');
    cy.get('[data-cy=password-input]').type('password123');
    cy.get('[data-cy=remember-me]').check();
    cy.get('[data-cy=login-button]').click();
    
    // Wait for API call to complete
    cy.wait('@loginRequest');
    
    // Verify redirection
    cy.url().should('include', '/feed');
    
    // Now visit login page again
    cy.clearCookies(); // Clear cookies but not localStorage
    cy.visit('/login');
    
    // Email should be pre-filled
    cy.get('[data-cy=email-input]').should('have.value', 'test@example.com');
  });

  it('should navigate to register page', () => {
    cy.get('[data-cy=register-link]').click();
    cy.url().should('include', '/register');
  });

  it('should navigate to forgot password page', () => {
    cy.get('[data-cy=forgot-password]').click();
    cy.url().should('include', '/forgot-password');
  });

  it('should handle login with Google', () => {
    // Mock window.open for Google OAuth
    cy.window().then(win => {
      cy.stub(win, 'open').as('windowOpen');
    });
    
    cy.get('[data-cy=google-login-button]').click();
    
    // Verify window.open was called with Google OAuth URL
    cy.get('@windowOpen').should('be.called');
    cy.get('@windowOpen').should('be.calledWith', Cypress.sinon.match(/accounts\.google\.com/));
    
    // Simulate successful OAuth callback
    cy.interceptApi('GET', '/auth/oauth/google/callback', {
      success: true,
      access_token: 'google-access-token',
      refresh_token: 'google-refresh-token'
    }).as('oauthCallback');
    
    // Simulate the callback by visiting directly
    cy.visit('/auth/oauth/google/callback?code=test-code');
    
    // Verify tokens are stored
    cy.window().its('localStorage')
      .invoke('getItem', 'accessToken')
      .should('equal', 'google-access-token');
    
    // Verify redirection to feed page
    cy.url().should('include', '/feed');
  });
}); 