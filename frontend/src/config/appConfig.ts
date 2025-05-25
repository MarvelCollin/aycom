const appConfig = {
  environment: import.meta.env.MODE || process.env.NODE_ENV || 'development',
  
  auth: {
    enabled: true
  },
  
  api: {
    baseUrl: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8083/api/v1',
    aiServiceUrl: import.meta.env.VITE_AI_SERVICE_URL || 'http://localhost:5000'
  },

  ui: {
    showErrorToasts: true
  }
};

export default appConfig; 