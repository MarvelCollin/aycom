/**
 * Main application configuration
 */

const appConfig = {
  // Authentication settings
  auth: {
    // Set to true to enable authentication
    enabled: true
  },
  
  // API settings
  api: {
    baseUrl: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081/api/v1'
  },

  ui: {
    showErrorToasts: true
  }
};

export default appConfig; 