const appConfig = {
  auth: {
    enabled: true
  },
  
  api: {
    baseUrl: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8083/api/v1'
  },

  ui: {
    showErrorToasts: true
  }
};

export default appConfig; 