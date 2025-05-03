describe('Login Page', () => {
  beforeEach(() => {
    cy.visit('/login');
  });

  it('redirects logged-in users to /feed', () => {
    window.localStorage.setItem('auth', JSON.stringify({ accessToken: 'dummy' }));
    cy.visit('/login');
    cy.url().should('include', '/feed');
    window.localStorage.removeItem('auth');
  });

  it('shows Google authentication option', () => {
    cy.get('[data-cy="google-login-button"]').should('exist');
  });

  it('shows required fields and links', () => {
    cy.get('[data-cy="email-input"]').should('exist');
    cy.get('[data-cy="password-input"]').should('exist');
    cy.get('[data-cy="login-button"]').should('exist');
    cy.get('[data-cy="landing-link"]').should('exist');
    cy.get('[data-cy="forgot-password"]').should('exist');
  });

  it('validates required fields', () => {
    cy.get('[data-cy="login-button"]').click();
    cy.get('[data-cy="email-error"]').should('exist');
    cy.get('[data-cy="password-error"]').should('exist');
  });

  it('shows reCAPTCHA error if not completed', () => {
    cy.get('[data-cy="email-input"]').type('user@example.com');
    cy.get('[data-cy="password-input"]').type('Password123!');
    cy.get('[data-cy="login-button"]').click();
    cy.get('[data-cy="recaptcha-error"]').should('exist');
  });
});

describe('Forgotten Account Page', () => {
  beforeEach(() => {
    cy.visit('/forgot-password');
  });

  it('redirects logged-in users to /feed', () => {
    window.localStorage.setItem('auth', JSON.stringify({ accessToken: 'dummy' }));
    cy.visit('/forgot-password');
    cy.url().should('include', '/feed');
    window.localStorage.removeItem('auth');
  });

  it('shows required fields and landing link on step 1', () => {
    cy.get('input[type="email"]').should('exist');
    cy.contains('Back to Landing Page').should('exist');
    cy.get('button[type="submit"]').should('exist');
  });

  it('shows reCAPTCHA error if not completed', () => {
    cy.get('input[type="email"]').type('user@example.com');
    cy.get('button[type="submit"]').click();
    cy.contains('reCAPTCHA').should('exist');
  });

  it('shows security question on valid email (mocked)', () => {
    cy.intercept('POST', '**/auth/forgot-password-question', {
      statusCode: 200,
      body: { security_question: 'What is your favorite video game?', old_password_hash: 'oldhash' }
    }).as('getQuestion');
    cy.get('input[type="email"]').type('user@example.com');
    cy.window().then(win => {
      win.document.querySelector('form').recaptchaToken = 'dummy';
    });
    cy.get('button[type="submit"]').click();
    cy.wait('@getQuestion');
    cy.contains('What is your favorite video game?').should('exist');
  });

  it('shows error for incorrect answer (mocked)', () => {
    cy.intercept('POST', '**/auth/forgot-password-question', {
      statusCode: 200,
      body: { security_question: 'What is your favorite video game?', old_password_hash: 'oldhash' }
    });
    cy.get('input[type="email"]').type('user@example.com');
    cy.window().then(win => {
      win.document.querySelector('form').recaptchaToken = 'dummy';
    });
    cy.get('button[type="submit"]').click();
    cy.intercept('POST', '**/auth/forgot-password-verify', {
      statusCode: 400,
      body: { message: 'Incorrect answer.' }
    }).as('verifyAnswer');
    cy.get('input[type="text"]').type('wronganswer');
    cy.get('button[type="submit"]').click();
    cy.wait('@verifyAnswer');
    cy.contains('Incorrect answer.').should('exist');
  });

  it('allows password reset on correct answer (mocked)', () => {
    cy.intercept('POST', '**/auth/forgot-password-question', {
      statusCode: 200,
      body: { security_question: 'What is your favorite video game?', old_password_hash: 'oldhash' }
    });
    cy.get('input[type="email"]').type('user@example.com');
    cy.window().then(win => {
      win.document.querySelector('form').recaptchaToken = 'dummy';
    });
    cy.get('button[type="submit"]').click();
    cy.intercept('POST', '**/auth/forgot-password-verify', {
      statusCode: 200,
      body: {}
    });
    cy.get('input[type="text"]').type('correctanswer');
    cy.get('button[type="submit"]').click();
    cy.get('input[type="password"]').should('exist');
  });

  it('validates new password is not same as old (mocked)', () => {
    cy.intercept('POST', '**/auth/forgot-password-question', {
      statusCode: 200,
      body: { security_question: 'What is your favorite video game?', old_password_hash: 'oldhash' }
    });
    cy.get('input[type="email"]').type('user@example.com');
    cy.window().then(win => {
      win.document.querySelector('form').recaptchaToken = 'dummy';
    });
    cy.get('button[type="submit"]').click();
    cy.intercept('POST', '**/auth/forgot-password-verify', {
      statusCode: 200,
      body: {}
    });
    cy.get('input[type="text"]').type('correctanswer');
    cy.get('button[type="submit"]').click();
    cy.get('input[type="password"]').type('oldhash');
    cy.get('button[type="submit"]').click();
    cy.contains('cannot be the same as the old password').should('exist');
  });
}); 