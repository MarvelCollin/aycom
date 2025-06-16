/// <reference types="cypress" />

declare namespace Cypress {
  interface Chainable<Subject> {
    /**
     * Custom command to login a user
     * @example cy.login('user@example.com', 'password123')
     */
    login(email: string, password: string): Chainable<any>

    /**
     * Custom command to login with Google
     * @example cy.loginWithGoogle()
     */
    loginWithGoogle(): Chainable<any>

    /**
     * Custom command to register a new user
     * @example cy.register('user@example.com', 'password123', 'Test User')
     */
    register(email: string, password: string, name: string): Chainable<any>

    /**
     * Custom command to intercept API calls
     * @example cy.interceptApi('GET', '/api/users', { users: [] })
     */
    interceptApi(method: string, path: string, response?: object): Chainable<any>

    /**
     * Custom command to mock file upload without actual files
     * @example cy.get('input[type="file"]').mockFileUpload('test.jpg', 'image/jpeg')
     */
    mockFileUpload(fileName?: string, fileType?: string): Chainable<any>

    /**
     * Custom command to attach a file to an input
     * @example cy.get('input[type="file"]').attachFile({ fileContent: 'base64Content', fileName: 'test.jpg', mimeType: 'image/jpeg' })
     */
    attachFile(options: { fileContent: string, fileName: string, mimeType: string }): Chainable<any>
  }
}