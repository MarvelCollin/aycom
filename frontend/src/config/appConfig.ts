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
      ? `${window.location.protocol}//${window.location.hostname}:8083/api/v1`  // Browser accessing API on mapped port 8083
      : (import.meta.env.VITE_API_BASE_URL || 'http://api_gateway:8081/api/v1'), // Inside Docker network use internal port 8081
    // Use HTTP for WebSocket if on HTTP, WSS for HTTPS
    wsUrl: (typeof window !== 'undefined')
      ? `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.hostname}:8083/api/v1`
      : (import.meta.env.VITE_WS_URL || 'ws://api_gateway:8081/api/v1'),
    aiServiceUrl: (typeof window !== 'undefined')
      ? `${window.location.protocol}//${window.location.hostname}:5000`
      : (import.meta.env.VITE_AI_SERVICE_URL || 'http://ai_service:5000')
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
console.log('[Config] API URL:', appConfig.api.baseUrl);
console.log('[Config] WebSocket URL:', appConfig.api.wsUrl);
console.log('[Config] AI Service URL:', appConfig.api.aiServiceUrl);

// Add more comprehensive API health check
export const checkApiHealth = async () => {
  try {
    console.log(`[Config] Testing API connection to: ${appConfig.api.baseUrl}`);
    
    // Add origin information to help with debugging CORS issues
    const currentOrigin = typeof window !== 'undefined' ? window.location.origin : 'unknown';
    console.log(`[Config] Current origin: ${currentOrigin}`);
    
    // Check if API gateway is accessible
    const response = await fetch(`${appConfig.api.baseUrl}/health`, {
      method: 'GET',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json',
        'Origin': currentOrigin
      },
      mode: 'cors', // Ensure CORS is enabled
      credentials: 'include' // Send cookies if available
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
      
      // Check CORS headers in response
      const allowOrigin = response.headers.get('Access-Control-Allow-Origin');
      const allowMethods = response.headers.get('Access-Control-Allow-Methods');
      const allowHeaders = response.headers.get('Access-Control-Allow-Headers');
      
      console.log('[Config] CORS headers in response:', {
        'Access-Control-Allow-Origin': allowOrigin || 'not present',
        'Access-Control-Allow-Methods': allowMethods || 'not present',
        'Access-Control-Allow-Headers': allowHeaders || 'not present'
      });
      
      // Try to parse response
      try {
        const responseData = await response.json();
        console.log('[Config] API health response:', responseData);
      } catch(e) {
        console.warn('[Config] Could not parse API response as JSON');
      }
    }
    
    // Also check WebSocket connectivity
    console.log('[Config] Testing WebSocket connectivity...');
    try {
      const wsUrl = appConfig.api.wsUrl.replace('/api/v1', '') + '/api/v1/health/ws';
      console.log(`[Config] Attempting WebSocket connection to: ${wsUrl}`);
      
      const ws = new WebSocket(wsUrl);
      const wsTimeout = setTimeout(() => {
        console.warn('[Config] WebSocket connection test timed out');
        ws.close();
      }, 5000);
      
      ws.onopen = () => {
        clearTimeout(wsTimeout);
        console.log('[Config] WebSocket connection test successful');
        ws.close();
      };
      
      ws.onerror = (error) => {
        clearTimeout(wsTimeout);
        console.error('[Config] WebSocket connection test failed:', error);
        ws.close();
      };
    } catch (wsError) {
      console.error('[Config] Error testing WebSocket connection:', wsError);
    }
    
    // Also check chats endpoint specifically
    console.log('[Config] Testing chats endpoint...');
    try {
      const chatsResponse = await fetch(`${appConfig.api.baseUrl}/chats`, {
        method: 'GET',
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/json',
          'Origin': currentOrigin
        },
        mode: 'cors',
        credentials: 'include'
      });
      
      console.log(`[Config] Chats endpoint status: ${chatsResponse.status}`);
      
      if (!chatsResponse.ok) {
        console.warn('[Config] Chats endpoint not responding correctly');
        const contentType = chatsResponse.headers.get('content-type');
        console.log('[Config] Chats endpoint content-type:', contentType || 'none');
      } else {
        console.log('[Config] Chats endpoint responding correctly');
      }
    } catch(e) {
      console.error('[Config] Error testing chats endpoint:', e);
    }
  } catch (error) {
    console.error('[Config] Error connecting to API:', error);
    
    // CORS errors don't provide much info in the error object
    if (error instanceof TypeError && error.message.includes('Failed to fetch')) {
      console.error('[Config] This might be a CORS issue or the API server is not running');
      console.error('[Config] Try checking the Network tab in DevTools for more details');
    }
  }
};

export default appConfig; 