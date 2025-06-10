const appConfig = {
  environment: import.meta.env.MODE || process.env.NODE_ENV || 'development',
  
  auth: {
    enabled: true,
    tokenRefreshBuffer: 5 * 60 * 1000, // 5 minutes in milliseconds
    tokenRefreshRetryDelay: 10 * 1000, // 10 seconds between failed refresh attempts
    loginUrl: '/login',
    registerUrl: '/register',
    logoutRedirectUrl: '/'
  },
  
  api: {
    // For development in Docker, use the service name
    // For browser access from outside Docker, use the port mapping
    baseUrl: (typeof window !== 'undefined') 
      ? `${window.location.protocol}//${window.location.hostname}:8083/api/v1`  // Browser accessing API on same hostname
      : (import.meta.env.VITE_API_BASE_URL || 'http://api_gateway:8081/api/v1'), // Inside Docker network
    wsUrl: (typeof window !== 'undefined')
      ? `ws://${window.location.hostname}:8083/api/v1`
      : (import.meta.env.VITE_WS_URL || 'ws://localhost:8083/api/v1'),
    aiServiceUrl: (typeof window !== 'undefined')
      ? `${window.location.protocol}//${window.location.hostname}:5000`
      : (import.meta.env.VITE_AI_SERVICE_URL || 'http://localhost:5000')
  },

  supabase: {
    url: import.meta.env.VITE_SUPABASE_URL || '',
    publicKey: import.meta.env.VITE_SUPABASE_PUBLIC_KEY || '',
    bucketName: import.meta.env.VITE_SUPABASE_BUCKET || 'aycom-media'
  },

  ui: {
    showErrorToasts: true
  }
};

// Log the configuration for debugging
console.log('[Config] Environment:', appConfig.environment);
console.log('[Config] API Base URL:', appConfig.api.baseUrl);
console.log('[Config] AI Service URL:', appConfig.api.aiServiceUrl);

// Helper function to log API health status
export const checkApiHealth = async () => {
  try {
    console.log(`[Config] Testing API connection to: ${appConfig.api.baseUrl}`);
    // Use /trends endpoint which we know is working from the logs
    const response = await fetch(`${appConfig.api.baseUrl}/trends`, {
      method: 'GET',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json'
      },
      mode: 'cors' // Ensure CORS is enabled
    });
    
    console.log(`[Config] API health check status: ${response.status}`);
    
    if (!response.ok) {
      console.warn(`[Config] API endpoint returned ${response.status} - API may not be working correctly`);
      
      // Try to read the error response
      try {
        const errorText = await response.text();
        console.warn(`[Config] API error response: ${errorText}`);
      } catch (textError) {
        console.warn(`[Config] Could not read API error response: ${textError}`);
      }
    } else {
      console.log('[Config] API connection successful');
    }
  } catch (error) {
    console.error('[Config] Error connecting to API:', error);
  }
};

export default appConfig; 