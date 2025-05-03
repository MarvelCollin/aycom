describe('Register Page', () => {
  beforeEach(() => {
    cy.visit('/register');
  });

  it('shows Google authentication option', () => {
    cy.get('[data-cy="google-login-button"]').should('exist');
  });

  it('shows all required fields and options', () => {
    cy.get('[data-cy="name-input"]').should('exist');
    cy.get('[data-cy="username-input"]').should('exist');
    cy.get('[data-cy="email-input"]').should('exist');
    cy.get('[data-cy="password-input"]').should('exist');
    cy.get('[data-cy="confirm-password-input"]').should('exist');
    cy.get('[data-cy="gender-male"]').should('exist');
    cy.get('[data-cy="gender-female"]').should('exist');
    cy.get('[data-cy="dob-month"]').should('exist');
    cy.get('[data-cy="dob-day"]').should('exist');
    cy.get('[data-cy="dob-year"]').should('exist');
    cy.get('[data-cy="profile-picture-input"]').should('exist');
    cy.get('[data-cy="banner-input"]').should('exist');
    cy.get('[data-cy="security-question"]').should('exist');
    cy.get('[data-cy="security-answer"]').should('exist');
    cy.get('[data-cy="newsletter-checkbox"]').should('exist');
    cy.get('[data-cy="register-button"]').should('exist');
  });

  it('validates name field', () => {
    cy.get('[data-cy="name-input"]').clear().type('A1!').blur();
    cy.get('[data-cy="name-error"]').should('contain', 'more than 4 characters');
    cy.get('[data-cy="name-input"]').clear().type('12345').blur();
    cy.get('[data-cy="name-error"]').should('contain', 'must not contain symbols or numbers');
    cy.get('[data-cy="name-input"]').clear().type('Valid Name').blur();
    cy.get('[data-cy="name-error"]').should('not.exist');
  });

  it('validates username field', () => {
    cy.get('[data-cy="username-input"]').clear().type('a').blur();
    cy.get('[data-cy="username-error"]').should('contain', 'at least 3 characters');
    cy.get('[data-cy="username-input"]').clear().type('invalid!user').blur();
    cy.get('[data-cy="username-error"]').should('contain', 'can only contain letters, numbers, and underscores');
  });

  it('validates email field', () => {
    cy.get('[data-cy="email-input"]').clear().type('invalidemail').blur();
    cy.get('[data-cy="email-error"]').should('contain', 'valid email');
    cy.get('[data-cy="email-input"]').clear().type('test@example.com').blur();
    cy.get('[data-cy="email-error"]').should('not.exist');
  });

  it('validates password field', () => {
    cy.get('[data-cy="password-input"]').clear().type('short').blur();
    cy.get('[data-cy="password-error"]').should('contain', 'at least 8 characters');
    cy.get('[data-cy="password-input"]').clear().type('alllowercase1!').blur();
    cy.get('[data-cy="password-error"]').should('contain', 'one uppercase letter');
    cy.get('[data-cy="password-input"]').clear().type('ALLUPPERCASE1!').blur();
    cy.get('[data-cy="password-error"]').should('contain', 'one lowercase letter');
    cy.get('[data-cy="password-input"]').clear().type('NoNumber!').blur();
    cy.get('[data-cy="password-error"]').should('contain', 'one number');
    cy.get('[data-cy="password-input"]').clear().type('Valid1Password!').blur();
    cy.get('[data-cy="password-error"]').should('not.exist');
  });

  it('validates confirm password field', () => {
    cy.get('[data-cy="password-input"]').clear().type('Valid1Password!');
    cy.get('[data-cy="confirm-password-input"]').clear().type('DifferentPassword').blur();
    cy.get('[data-cy="password-match-error"]').should('contain', 'Passwords do not match');
    cy.get('[data-cy="confirm-password-input"]').clear().type('Valid1Password!').blur();
    cy.get('[data-cy="password-match-error"]').should('not.exist');
  });

  it('validates gender selection', () => {
    cy.get('[data-cy="gender-male"]').check();
    cy.get('[data-cy="gender-female"]').check();
  });

  it('validates age >= 13 years', () => {
    const thisYear = new Date().getFullYear();
    cy.get('[data-cy="dob-month"]').select('January');
    cy.get('[data-cy="dob-day"]').select('1');
    cy.get('[data-cy="dob-year"]').select((thisYear - 10).toString()).blur();
    cy.get('[data-cy="dob-error"]').should('contain', 'at least 13 years old');
    cy.get('[data-cy="dob-year"]').select((thisYear - 20).toString()).blur();
    cy.get('[data-cy="dob-error"]').should('not.exist');
  });

  it('shows all security questions', () => {
    cy.get('[data-cy="security-question"]').click();
    cy.get('[data-cy="security-question"]').find('option').should('have.length.at.least', 5);
  });

  it('can check newsletter subscription', () => {
    cy.get('[data-cy="newsletter-checkbox"]').check().should('be.checked');
  });

  it('shows reCAPTCHA error if not completed', () => {
    cy.get('[data-cy="register-button"]').click();
    cy.get('[data-cy="recaptcha-error"]').should('exist');
  });

  it('shows link to login page', () => {
    cy.get('[data-cy="login-link"]').should('exist');
  });

  it('completes registration and shows verification step', () => {
    // Fill all required fields with valid data
    cy.get('[data-cy="name-input"]').clear().type('Valid Name');
    cy.get('[data-cy="username-input"]').clear().type('validusername');
    cy.get('[data-cy="email-input"]').clear().type('validuser@example.com');
    cy.get('[data-cy="password-input"]').clear().type('Valid1Password!');
    cy.get('[data-cy="confirm-password-input"]').clear().type('Valid1Password!');
    cy.get('[data-cy="gender-male"]').check();
    cy.get('[data-cy="dob-month"]').select('January');
    cy.get('[data-cy="dob-day"]').select('1');
    cy.get('[data-cy="dob-year"]').select('2000');
    cy.get('[data-cy="profile-picture-input"]').selectFile('cypress/fixtures/profile.jpg', { force: true });
    cy.get('[data-cy="banner-input"]').selectFile('cypress/fixtures/banner.jpg', { force: true });
    cy.get('[data-cy="security-question"]').select(1);
    cy.get('[data-cy="security-answer"]').clear().type('MyAnswer');
    cy.get('[data-cy="register-button"]').click();
    // Simulate reCAPTCHA and backend response as needed
    // Should show verification step
    cy.contains('Enter it below to verify').should('exist');
  });

  it('shows resend code option after timer expires', () => {
    // Simulate timer expiry
    cy.window().then(win => {
      win.document.dispatchEvent(new Event('timer-expired'));
    });
    cy.contains('Resend code').should('exist');
  });
});
