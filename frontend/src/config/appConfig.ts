/**
 * Main application configuration
 */

const appConfig = {
  // Authentication settings
  auth: {
    // Set to false to bypass authentication for design testing
    enabled: false,
    
    // Default mock user data to use when auth is disabled
    mockUser: {
      userId: 'mock-user-id',
      username: 'johndoe',
      displayName: 'John Doe',
      avatar: '👤',
      profilePicture: '👤'
    }
  },
  
  // API settings
  api: {
    baseUrl: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'
  }
};

export default appConfig; 