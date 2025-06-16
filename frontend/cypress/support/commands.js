// ***********************************************
// This example commands.js shows you how to
// create various custom commands and overwrite
// existing commands.
//
// For more comprehensive examples of custom
// commands please read more here:
// https://on.cypress.io/custom-commands
// ***********************************************

// Custom login command to reuse across tests
Cypress.Commands.add("login", (email, password) => {
  cy.visit("/login");
  cy.get("[data-cy=email-input]").type(email);
  cy.get("[data-cy=password-input]").type(password);
  cy.get("[data-cy=login-button]").click();
  // Wait for login to complete and redirect
  cy.url().should("include", "/feed");
});

// Custom command to test Google OAuth flow
Cypress.Commands.add("loginWithGoogle", () => {
  cy.visit("/login");
  cy.get("[data-cy=google-login-button]").click();
  // Note: Testing actual OAuth requires additional mocking
  // This is a placeholder that would need to be enhanced
});

// Custom command to register a new user
Cypress.Commands.add("register", (email, password, name) => {
  cy.visit("/register");
  cy.get("[data-cy=name-input]").type(name);
  cy.get("[data-cy=email-input]").type(email);
  cy.get("[data-cy=password-input]").type(password);
  cy.get("[data-cy=confirm-password-input]").type(password);
  cy.get("[data-cy=register-button]").click();
  // Wait for registration to complete
  cy.url().should("include", "/login");
});

// Intercept API calls for testing
Cypress.Commands.add("interceptApi", (method, path, response = {}) => {
  return cy.intercept(
    method,
    `${Cypress.env("apiUrl")}${path}`,
    {
      statusCode: 200,
      body: response,
    }
  );
});

// Custom command to test file uploads without actual files
Cypress.Commands.add("mockFileUpload", { prevSubject: "element" }, (subject, fileName = "test.jpg", fileType = "image/jpeg") => {
  // Create a blob with simple content
  const blob = new Blob(["test file content"], { type: fileType });

  // Create a File from the blob
  const testFile = new File([blob], fileName, { type: fileType });

  // Create a DataTransfer instance
  const dataTransfer = new DataTransfer();
  dataTransfer.items.add(testFile);

  // Assign the created file list to the element
  subject[0].files = dataTransfer.files;

  // Trigger change event
  return cy.wrap(subject).trigger("change", { force: true });
});

// Custom command to attach file to an input
Cypress.Commands.add("attachFile", { prevSubject: "element" }, (subject, { fileContent, fileName, mimeType }) => {
  // Use the built-in Cypress.Blob utility
  const fileBlob = Cypress.Blob.base64StringToBlob(fileContent, mimeType);
  const testFile = new File([fileBlob], fileName, { type: mimeType });
  const dataTransfer = new DataTransfer();

  dataTransfer.items.add(testFile);
  subject[0].files = dataTransfer.files;

  return cy.wrap(subject).trigger("change", { force: true });
});