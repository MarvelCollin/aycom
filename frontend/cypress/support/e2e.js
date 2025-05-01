// ***********************************************************
// This support file is loaded automatically by Cypress.
// It allows you to add commands to Cypress which will be accessible in all your tests.
// ***********************************************************

// Import commands.js using ES2015 syntax:
import './commands'

// Configure Cypress for file uploads and attachments
import 'cypress-file-upload';

// Use the existing Cypress Blob utility instead of overriding it
// The Cypress.Blob utilities are already available and include base64StringToBlob

// Alternatively you can use CommonJS syntax:
// require('./commands')

// Hide fetch/XHR requests in the Command Log
const app = window.top;
if (!app.document.head.querySelector('[data-hide-command-log-request]')) {
  const style = app.document.createElement('style');
  style.innerHTML =
    '.command-name-request, .command-name-xhr { display: none }';
  style.setAttribute('data-hide-command-log-request', '');
  app.document.head.appendChild(style);
}