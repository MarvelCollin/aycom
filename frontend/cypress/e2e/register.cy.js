describe('Registration Page', () => {
  beforeEach(() => {
    // Clear cookies and localStorage between tests
    cy.clearCookies();
    cy.clearLocalStorage();
    // Visit the registration page before each test
    cy.visit('/register');
  });

  it('should have the correct UI elements', () => {
    // Check for page title
    cy.get('[data-cy=page-title]').should('contain', 'Register');
    
    // Check for form elements
    cy.get('[data-cy=name-input]').should('be.visible');
    cy.get('[data-cy=username-input]').should('be.visible');
    cy.get('[data-cy=email-input]').should('be.visible');
    cy.get('[data-cy=password-input]').should('be.visible');
    cy.get('[data-cy=confirm-password-input]').should('be.visible');
    cy.get('[data-cy=gender-select]').should('be.visible');
    cy.get('[data-cy=dob-day]').should('be.visible');
    cy.get('[data-cy=dob-month]').should('be.visible');
    cy.get('[data-cy=dob-year]').should('be.visible');
    cy.get('[data-cy=security-question]').should('be.visible');
    cy.get('[data-cy=security-answer]').should('be.visible');
    cy.get('[data-cy=profile-picture-upload]').should('be.visible');
    cy.get('[data-cy=banner-upload]').should('be.visible');
    cy.get('[data-cy=register-button]').should('be.visible');
    
    // Check for login link
    cy.get('[data-cy=login-link]').should('be.visible');
  });

  it('should validate all form fields correctly', () => {
    // Submit empty form to trigger all validations
    cy.get('[data-cy=register-button]').click();
    
    // Check validation errors
    cy.get('[data-cy=name-error]').should('be.visible');
    cy.get('[data-cy=username-error]').should('be.visible');
    cy.get('[data-cy=email-error]').should('be.visible');
    cy.get('[data-cy=password-error]').should('be.visible');
    cy.get('[data-cy=confirm-password-error]').should('be.visible');
    cy.get('[data-cy=gender-error]').should('be.visible');
    cy.get('[data-cy=dob-error]').should('be.visible');
    cy.get('[data-cy=security-question-error]').should('be.visible');
    cy.get('[data-cy=profile-picture-error]').should('be.visible');
    cy.get('[data-cy=banner-error]').should('be.visible');
  });

  it('should validate name field correctly', () => {
    // Test name with 4 characters (too short)
    cy.get('[data-cy=name-input]').type('John').blur();
    cy.get('[data-cy=name-error]').should('contain', 'Name must be more than 4 characters');
    
    // Test name with numbers (invalid)
    cy.get('[data-cy=name-input]').clear().type('John123').blur();
    cy.get('[data-cy=name-error]').should('contain', 'Name must not contain symbols or numbers');
    
    // Test name with symbols (invalid)
    cy.get('[data-cy=name-input]').clear().type('John@Doe').blur();
    cy.get('[data-cy=name-error]').should('contain', 'Name must not contain symbols or numbers');
    
    // Test valid name
    cy.get('[data-cy=name-input]').clear().type('John Doe').blur();
    cy.get('[data-cy=name-error]').should('not.exist');
  });

  it('should validate username field correctly', () => {
    // Test username that's too short
    cy.get('[data-cy=username-input]').type('jo').blur();
    cy.get('[data-cy=username-error]').should('contain', 'Username must be at least 3 characters');
    
    // Test username with spaces (invalid)
    cy.get('[data-cy=username-input]').clear().type('john doe').blur();
    cy.get('[data-cy=username-error]').should('contain', 'Username cannot contain spaces');
    
    // Test username with symbols (invalid)
    cy.get('[data-cy=username-input]').clear().type('john@doe').blur();
    cy.get('[data-cy=username-error]').should('contain', 'Username can only contain letters, numbers, and underscores');
    
    // Test valid username
    cy.get('[data-cy=username-input]').clear().type('john_doe123').blur();
    cy.get('[data-cy=username-error]').should('not.exist');
  });

  it('should validate email format correctly', () => {
    // Test invalid email formats
    cy.get('[data-cy=email-input]').type('invalid-email').blur();
    cy.get('[data-cy=email-error]').should('contain', 'valid email');
    
    cy.get('[data-cy=email-input]').clear().type('user@').blur();
    cy.get('[data-cy=email-error]').should('contain', 'valid email');
    
    cy.get('[data-cy=email-input]').clear().type('user@domain').blur();
    cy.get('[data-cy=email-error]').should('contain', 'valid email');
    
    // Test valid email format
    cy.get('[data-cy=email-input]').clear().type('user@domain.com').blur();
    cy.get('[data-cy=email-error]').should('not.exist');
  });

  it('should validate password complexity and matching', () => {
    // Test password that's too short
    cy.get('[data-cy=password-input]').type('short').blur();
    cy.get('[data-cy=password-error]').should('contain', 'at least 8 characters');
    
    // Test password without uppercase
    cy.get('[data-cy=password-input]').clear().type('password123!').blur();
    cy.get('[data-cy=password-error]').should('contain', 'uppercase letter');
    
    // Test password without number
    cy.get('[data-cy=password-input]').clear().type('Password!').blur();
    cy.get('[data-cy=password-error]').should('contain', 'number');
    
    // Test password without special character
    cy.get('[data-cy=password-input]').clear().type('Password123').blur();
    cy.get('[data-cy=password-error]').should('contain', 'special character');
    
    // Test valid password
    cy.get('[data-cy=password-input]').clear().type('Password123!').blur();
    cy.get('[data-cy=password-error]').should('not.exist');
    
    // Test non-matching confirmation
    cy.get('[data-cy=confirm-password-input]').type('DifferentPassword123!').blur();
    cy.get('[data-cy=password-match-error]').should('contain', 'Passwords do not match');
    
    // Test matching confirmation
    cy.get('[data-cy=confirm-password-input]').clear().type('Password123!').blur();
    cy.get('[data-cy=password-match-error]').should('not.exist');
  });

  it('should validate age requirement correctly', () => {
    // Calculate a date that would make the user 12 years old
    const today = new Date();
    const underageYear = today.getFullYear() - 12;
    
    // Set DOB for a 12-year-old (too young)
    cy.get('[data-cy=dob-month]').select('January');
    cy.get('[data-cy=dob-day]').select('1');
    cy.get('[data-cy=dob-year]').select(underageYear.toString());
    cy.get('[data-cy=name-input]').click(); // Trigger validation by clicking elsewhere
    
    cy.get('[data-cy=dob-error]').should('contain', 'at least 13 years old');
    
    // Now set DOB for 13-year-old (valid)
    const validYear = today.getFullYear() - 13;
    cy.get('[data-cy=dob-year]').select(validYear.toString());
    cy.get('[data-cy=name-input]').click(); // Trigger validation
    
    cy.get('[data-cy=dob-error]').should('not.exist');
  });

  it('should handle file uploads', () => {
    // Mock file uploads for profile picture and banner
    cy.get('[data-cy=profile-picture-upload]').mockFileUpload('profile.jpg');
    cy.get('[data-cy=profile-picture-preview]').should('be.visible');
    
    cy.get('[data-cy=banner-upload]').mockFileUpload('banner.jpg');
    cy.get('[data-cy=banner-preview]').should('be.visible');
  });

  it('should validate security question fields', () => {
    // Select security question but leave answer empty
    cy.get('[data-cy=security-question]').select('What was your first pet\'s name?');
    cy.get('[data-cy=register-button]').click();
    cy.get('[data-cy=security-answer-error]').should('be.visible');
    
    // Enter security answer
    cy.get('[data-cy=security-answer]').type('Fluffy');
    cy.get('[data-cy=security-answer-error]').should('not.exist');
  });

  it('should check for username availability', () => {
    // Mock the username check API call for taken username
    cy.interceptApi('GET', '/auth/check-username?username=taken_username', {
      statusCode: 409,
      body: {
        available: false,
        message: 'Username is already taken'
      }
    }).as('usernameTakenCheck');
    
    // Enter a taken username
    cy.get('[data-cy=username-input]').type('taken_username').blur();
    
    // Wait for API call to complete
    cy.wait('@usernameTakenCheck');
    
    // Should show username taken error
    cy.get('[data-cy=username-error]').should('contain', 'already taken');
    
    // Mock the username check API call for available username
    cy.interceptApi('GET', '/auth/check-username?username=available_username', {
      available: true,
      message: 'Username is available'
    }).as('usernameAvailableCheck');
    
    // Enter an available username
    cy.get('[data-cy=username-input]').clear().type('available_username').blur();
    
    // Wait for API call to complete
    cy.wait('@usernameAvailableCheck');
    
    // Should show username available message
    cy.get('[data-cy=username-available]').should('be.visible');
    cy.get('[data-cy=username-error]').should('not.exist');
  });

  it('should check for email availability', () => {
    // Mock the email check API call for taken email
    cy.interceptApi('GET', '/auth/check-email?email=taken@example.com', {
      statusCode: 409,
      body: {
        available: false,
        message: 'Email is already registered'
      }
    }).as('emailTakenCheck');
    
    // Enter a taken email
    cy.get('[data-cy=email-input]').type('taken@example.com').blur();
    
    // Wait for API call to complete
    cy.wait('@emailTakenCheck');
    
    // Should show email taken error
    cy.get('[data-cy=email-error]').should('contain', 'already registered');
    
    // Mock the email check API call for available email
    cy.interceptApi('GET', '/auth/check-email?email=available@example.com', {
      available: true,
      message: 'Email is available'
    }).as('emailAvailableCheck');
    
    // Enter an available email
    cy.get('[data-cy=email-input]').clear().type('available@example.com').blur();
    
    // Wait for API call to complete
    cy.wait('@emailAvailableCheck');
    
    // Should not show email error
    cy.get('[data-cy=email-error]').should('not.exist');
  });

  it('should successfully register a new account', () => {
    // Intercept all check API calls
    cy.interceptApi('GET', '/auth/check-username?username=newuser123', {
      available: true
    }).as('usernameCheck');
    
    cy.interceptApi('GET', '/auth/check-email?email=newuser@example.com', {
      available: true
    }).as('emailCheck');
    
    // Intercept the register API call
    cy.interceptApi('POST', '/auth/register', {
      success: true,
      message: 'Registration successful. Please check your email to verify your account.'
    }).as('registerRequest');
    
    // Fill out the complete form
    cy.get('[data-cy=name-input]').type('New Test User');
    cy.get('[data-cy=username-input]').type('newuser123');
    cy.get('[data-cy=email-input]').type('newuser@example.com');
    cy.get('[data-cy=password-input]').type('Password123!');
    cy.get('[data-cy=confirm-password-input]').type('Password123!');
    cy.get('[data-cy=gender-select]').select('Male');
    
    // Set date of birth for an 18-year-old
    const today = new Date();
    const birthYear = today.getFullYear() - 18;
    cy.get('[data-cy=dob-month]').select('January');
    cy.get('[data-cy=dob-day]').select('1');
    cy.get('[data-cy=dob-year]').select(birthYear.toString());
    
    // Set security question
    cy.get('[data-cy=security-question]').select('What was your first pet\'s name?');
    cy.get('[data-cy=security-answer]').type('Fluffy');
    
    // Mock file uploads
    cy.get('[data-cy=profile-picture-upload]').mockFileUpload('profile.jpg');
    cy.get('[data-cy=banner-upload]').mockFileUpload('banner.jpg');
    
    // Submit the form
    cy.get('[data-cy=register-button]').click();
    
    // Wait for API call to complete
    cy.wait('@registerRequest');
    
    // Verify success message
    cy.get('[data-cy=success-message]').should('be.visible');
    cy.contains('Registration successful').should('be.visible');
    
    // Verify redirection to login page or verification page
    cy.url().should('include', '/login');
  });

  it('should navigate to login page', () => {
    cy.get('[data-cy=login-link]').click();
    cy.url().should('include', '/login');
  });
});