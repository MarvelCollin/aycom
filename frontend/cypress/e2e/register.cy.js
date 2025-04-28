describe('Registration Page', () => {
  beforeEach(() => {
    // Visit the registration page before each test
    cy.visit('/register');
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

  it('should validate age requirement correctly', () => {
    // Calculate a date that would make the user 12 years old
    const today = new Date();
    const underageYear = today.getFullYear() - 12;
    
    // Set DOB for a 12-year-old (too young)
    cy.get('[data-cy=dob-month]').select('January');
    cy.get('[data-cy=dob-day]').select('1');
    cy.get('[data-cy=dob-year]').select(underageYear.toString());
    // Trigger validation by changing something else
    cy.get('[data-cy=name-input]').click();
    
    cy.get('[data-cy=dob-error]').should('contain', 'at least 13 years old');
    
    // Now set DOB for 13-year-old (valid)
    const validYear = today.getFullYear() - 13;
    cy.get('[data-cy=dob-year]').select(validYear.toString());
    // Trigger validation
    cy.get('[data-cy=name-input]').click();
    
    cy.get('[data-cy=dob-error]').should('not.exist');
  });

  it('should display character count for name field', () => {
    cy.get('[data-cy=name-input]').type('John');
    cy.get('[data-cy=name-char-count]').should('contain', '4 / 50');
    
    cy.get('[data-cy=name-input]').clear().type('A'.repeat(50));
    cy.get('[data-cy=name-char-count]').should('contain', '50 / 50');
  });

  it('should handle password complexity requirements', () => {
    // Fill in a basic valid form 
    cy.get('[data-cy=name-input]').type('John Doe');
    cy.get('[data-cy=username-input]').type('johndoe');
    cy.get('[data-cy=email-input]').type('john@example.com');
    
    // Test with a password missing requirements
    cy.get('[data-cy=password-input]').type('simple');
    cy.get('[data-cy=confirm-password-input]').type('simple');
    
    cy.get('[data-cy=register-button]').click();
    
    // Check that error messages appear
    cy.get('[data-cy=password-error]').should('be.visible');
    
    // Fix the password and try again
    cy.get('[data-cy=password-input]').clear().type('StrongPass123!');
    cy.get('[data-cy=confirm-password-input]').clear().type('StrongPass123!');
    
    // Password errors should be gone
    cy.get('[data-cy=password-error]').should('not.exist');
  });
});