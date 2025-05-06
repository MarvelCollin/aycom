describe('User Registration', () => {
  beforeEach(() => {
    cy.visit('/register');
  });

  it('should register a new user successfully with specific credentials', () => {
    const random = Math.floor(Math.random() * 100000);
    const testUser = {
      name: `miawmiawa`,
      username: 'miawmiaw',
      email: 'miawmiawa@gmail.com',
      password: 'Miawmiaw123@',
      confirmPassword: 'Miawmiaw123@',
      gender: 'female',
      dateOfBirth: {
        month: 'February',
        day: '14',
        year: '1999',
      },
      securityQuestion: 'What was the name of your first pet?',
      securityAnswer: 'Fluffy',
      subscribeToNewsletter: false,
      profilePictureFixture: 'profile.jpg',
      bannerFixture: 'banner.jpg',
    };

    cy.get('[data-cy=name-input]').type(testUser.name);
    cy.get('[data-cy=username-input]').type(testUser.username);
    cy.get('[data-cy=email-input]').type(testUser.email);
    cy.get('[data-cy=password-input]').type(testUser.password);
    cy.get('[data-cy=confirm-password-input]').type(testUser.confirmPassword);
    cy.get('[data-cy=gender-female]').check({ force: true });
    cy.get('[data-cy=dob-month]').select(testUser.dateOfBirth.month);
    cy.get('[data-cy=dob-day]').select(testUser.dateOfBirth.day);
    cy.get('[data-cy=dob-year]').select(testUser.dateOfBirth.year);
    cy.get('[data-cy=security-question]').select(testUser.securityQuestion);
    cy.get('[data-cy=security-answer]').type(testUser.securityAnswer);
    cy.get('[data-cy=profile-picture-input]').selectFile(`cypress/fixtures/${testUser.profilePictureFixture}`, { force: true });
    cy.get('[data-cy=banner-input]').selectFile(`cypress/fixtures/${testUser.bannerFixture}`, { force: true });

    cy.get('[data-cy=register-button]').click();

    cy.get('[data-cy=success-message]', { timeout: 10000 })
      .should('exist')
      .and('contain', 'Registration successful');
  });
});
