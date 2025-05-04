<script lang="ts">
  import { LogLevel, setGlobalLogLevel, logger } from '../../utils/logger';
  import { writable } from 'svelte/store';
  import { useAuth } from '../../hooks/useAuth';
  import { getAuthData } from '../../utils/auth';
  
  // Get auth state
  const { getAuthState, subscribe } = useAuth();
  
  // Initialize with real data from localStorage 
  let authState;
  try {
    const storedAuth = localStorage.getItem('auth');
    if (storedAuth) {
      try {
        const parsedAuth = JSON.parse(storedAuth);
        
        // Verify if the stored auth data has real values or sample values
        if (parsedAuth.userId === "sample-user-id") {
          console.warn("Detected sample user ID in localStorage. Attempting to fetch real auth state.");
          authState = getAuthState(); // Try to get the real state
        } else {
          authState = parsedAuth; // Use the stored auth data
        }
      } catch (parseError) {
        console.error('Failed to parse auth data from localStorage:', parseError);
        authState = getAuthState();
      }
    } else {
      authState = getAuthState(); // Fallback to hook state if no localStorage data
    }
  } catch (err) {
    console.error('Error loading auth data from storage:', err);
    authState = getAuthState(); // Fallback on error
  }
  
  // Before rendering UI, do one final check for sample data
  if (authState && authState.userId === "sample-user-id") {
    console.warn("Still showing sample user ID. Checking for alternative auth sources.");
    
    // Try to get auth from another source
    try {
      const authStatusFromLocalStorage = localStorage.getItem('aycom_authenticated');
      if (authStatusFromLocalStorage === 'true') {
        // We know user is authenticated, try to get real data
        const jwtToken = localStorage.getItem('accessToken') || sessionStorage.getItem('accessToken');
        if (jwtToken) {
          // Decode JWT to get user information
          try {
            const base64Url = jwtToken.split('.')[1];
            const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
            const jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
              return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
            }).join(''));
            
            const tokenData = JSON.parse(jsonPayload);
            if (tokenData.sub) {
              // Use real user ID from token
              authState.userId = tokenData.sub;
            }
          } catch (tokenError) {
            console.error('Failed to decode JWT token:', tokenError);
          }
        }
      }
    } catch (alternativeAuthError) {
      console.error('Failed to check alternative auth sources:', alternativeAuthError);
    }
  }
  
  // Subscribe to auth changes
  subscribe(newState => {
    authState = newState;
  });
  
  // Store for debug panel visibility
  const isVisible = writable(false);
  let panelVisible = false;
  
  // Store for log messages
  const logMessages = writable<{level: LogLevel; message: string; timestamp: Date; data?: any; stack?: string}[]>([]);
  
  // Toggle the debug panel visibility
  function togglePanel() {
    panelVisible = !panelVisible;
    isVisible.set(panelVisible);
    
    // Toggle body class for scroll locking
    if (typeof document !== 'undefined') {
      if (panelVisible) {
        document.body.classList.add('debug-panel-open');
      } else {
        document.body.classList.remove('debug-panel-open');
      }
    }
  }
  
  // Get log level options for the dropdown
  const logLevelOptions = [
    { value: LogLevel.TRACE, label: 'TRACE' },
    { value: LogLevel.DEBUG, label: 'DEBUG' },
    { value: LogLevel.INFO, label: 'INFO' },
    { value: LogLevel.WARN, label: 'WARN' },
    { value: LogLevel.ERROR, label: 'ERROR' },
    { value: LogLevel.NONE, label: 'NONE' }
  ];
  
  // Current log level (reactive)
  let currentLogLevel = logger.getLevel();
  
  // Update the log level when changed
  function updateLogLevel(event: Event) {
    const select = event.target as HTMLSelectElement;
    const newLevel = parseInt(select.value) as LogLevel;
    setGlobalLogLevel(newLevel);
    currentLogLevel = newLevel;
  }
  
  // Function to navigate to a route with bypass
  function navigateWithBypass(route: string) {
    // Add a bypass parameter to skip any guards or checks
    const separator = route.includes('?') ? '&' : '?';
    const bypassParam = `${separator}debug_bypass=true`;
    window.location.href = `${route}${bypassParam}`;
  }
  
  // Navigation routes
  const commonRoutes = [
    '/',
    '/login',
    '/register',
    '/home',
    '/feed',
    '/profile'
  ];
  
  // Function to get log level label
  function getLogLevelLabel(level: LogLevel): string {
    const option = logLevelOptions.find(opt => opt.value === level);
    return option ? option.label : 'UNKNOWN';
  }
  
  // Function to get log level color
  function getLogLevelColor(level: LogLevel): string {
    switch(level) {
      case LogLevel.ERROR: return 'text-red-500';
      case LogLevel.WARN: return 'text-yellow-500';
      case LogLevel.INFO: return 'text-blue-500';
      case LogLevel.DEBUG: return 'text-gray-400';
      case LogLevel.TRACE: return 'text-gray-500';
      default: return 'text-white';
    }
  }
  
  // Clear all logs
  function clearLogs() {
    logMessages.set([]);
  }
  
  // Utility function to safely stringify objects
  function safeStringify(obj: any, indent = 2): string {
    try {
      if (obj === null || obj === undefined) return String(obj);
      if (typeof obj === 'object') {
        // Handle circular references and complex objects
        const cache = new Set();
        const result = JSON.stringify(obj, (key, value) => {
          if (typeof value === 'object' && value !== null) {
            if (cache.has(value)) {
              return '[Circular Reference]';
            }
            cache.add(value);
          }
          // Handle Error objects specially
          if (value instanceof Error) {
            const errorObj: Record<string, any> = {};
            
            // Get all properties including non-enumerable ones
            Object.getOwnPropertyNames(value).forEach(propName => {
              errorObj[propName] = value[propName as keyof Error];
            });
            
            // Make sure we have the important properties
            if (!('message' in errorObj)) errorObj.message = value.message;
            if (!('stack' in errorObj)) errorObj.stack = value.stack;
            
            return errorObj;
          }
          return value;
        }, indent);
        return result;
      }
      return String(obj);
    } catch (e) {
      return `[Error stringifying object: ${e instanceof Error ? e.message : String(e)}]`;
    }
  }
  
  // Function to extract stack trace from error
  function getStackTrace(error: any): string | undefined {
    if (!error) return undefined;
    
    if (error instanceof Error) {
      return error.stack;
    } else if (typeof error === 'object' && 'stack' in error) {
      return error.stack as string;
    } else if (typeof error === 'object' && 'trace' in error) {
      return error.trace as string;
    }
    
    return undefined;
  }
  
  // Test logging at each level
  function testLogLevels() {
    logger.trace('This is a TRACE message');
    logger.debug('This is a DEBUG message');
    logger.info('This is an INFO message', null, { showToast: true });
    logger.warn('This is a WARN message', null, { showToast: true });
    
    // Create a real error for testing
    try {
      throw new Error('This is a test error with stack trace');
    } catch (error) {
      logger.error('This is an ERROR message', { error }, { showToast: true });
    }
  }
  
  // Fetch user profile
  let userProfileInfo = null;
  
  async function fetchUserProfile() {
    if (!authState.isAuthenticated || !authState.userId) {
      logger.warn('Cannot fetch user profile - not authenticated', null, { showToast: true });
      logMessages.update(logs => [
        { level: LogLevel.WARN, message: 'Cannot fetch user profile - not authenticated', timestamp: new Date() },
        ...logs
      ]);
      return;
    }
    
    try {
      logger.info('Fetching user profile...', null, { showToast: true });
      logMessages.update(logs => [
        { level: LogLevel.INFO, message: 'Fetching user profile...', timestamp: new Date() },
        ...logs
      ]);
      
      const apiUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8083/api/v1';
      
      // Log API request details for debugging
      logger.debug('API Request Details', { 
        url: `${apiUrl}/users/profile`,
        headers: { 'Authorization': `Bearer ${authState.accessToken}` },
        tokenInfo: {
          accessToken: authState.accessToken ? authState.accessToken.substring(0, 10) + '...' : null,
          tokenLength: authState.accessToken ? authState.accessToken.length : 0
        }
      });
      
      const response = await fetch(`${apiUrl}/users/profile`, {
        headers: {
          'Authorization': `Bearer ${authState.accessToken}`
        }
      });
      
      if (response.ok) {
        const data = await response.json();
        userProfileInfo = data;
        logger.info('User profile fetched successfully', { profileData: data }, { showToast: true });
        logMessages.update(logs => [
          { level: LogLevel.INFO, message: `User profile fetched successfully: ${JSON.stringify(data)}`, timestamp: new Date() },
          ...logs
        ]);
      } else {
        const errorText = await response.text();
        logger.error('Failed to fetch user profile', { 
          status: response.status, 
          statusText: response.statusText,
          responseBody: errorText
        }, { showToast: true });
        logMessages.update(logs => [
          { level: LogLevel.ERROR, message: `Failed to fetch user profile. Status: ${response.status}. Details: ${errorText}`, timestamp: new Date() },
          ...logs
        ]);
      }
    } catch (err) {
      logger.error('Error fetching user profile', { error: err }, { showToast: true });
      logMessages.update(logs => [
        { level: LogLevel.ERROR, message: `Error fetching user profile: ${err}`, timestamp: new Date() },
        ...logs
      ]);
    }
  }
  
  // Debug keyboard shortcuts handler
  function setupKeyboardShortcuts() {
    let konamiSequence = '';
    
    const handleKeyDown = (event: KeyboardEvent) => {    
      if (event.ctrlKey && event.shiftKey && event.key === 'D') {
        event.preventDefault();
        togglePanel();
      }
      
      // Konami code: "kowlin"
      const pressedKey = event.key.toLowerCase();
      konamiSequence = (konamiSequence + pressedKey).slice(-6);
      if (konamiSequence === 'kowlin') {
        panelVisible = true;
        isVisible.set(true);
        if (typeof document !== 'undefined') {
          document.body.classList.add('debug-panel-open');
        }
      }
    };
    
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }
  
  // Override logger functions to capture logs
  const originalTrace = logger.trace;
  const originalDebug = logger.debug;
  const originalInfo = logger.info;
  const originalWarn = logger.warn;
  const originalError = logger.error;
  
  // Override logger to capture logs
  logger.trace = function(message: string, data?: any, options?: any) {
    logMessages.update(logs => [
      { level: LogLevel.TRACE, message: String(message), timestamp: new Date(), data },
      ...logs
    ]);
    return originalTrace.call(logger, message, data, options);
  };
  
  logger.debug = function(message: string, data?: any, options?: any) {
    logMessages.update(logs => [
      { level: LogLevel.DEBUG, message: String(message), timestamp: new Date(), data },
      ...logs
    ]);
    return originalDebug.call(logger, message, data, options);
  };
  
  logger.info = function(message: string, data?: any, options?: any) {
    logMessages.update(logs => [
      { level: LogLevel.INFO, message: String(message), timestamp: new Date(), data },
      ...logs
    ]);
    return originalInfo.call(logger, message, data, options);
  };
  
  logger.warn = function(message: string, data?: any, options?: any) {
    logMessages.update(logs => [
      { level: LogLevel.WARN, message: String(message), timestamp: new Date(), data },
      ...logs
    ]);
    return originalWarn.call(logger, message, data, options);
  };
  
  logger.error = function(message: string, data?: any, options?: any) {
    // Extract stack trace if data is an error
    const stack = getStackTrace(data?.error || data);
    
    logMessages.update(logs => [
      { 
        level: LogLevel.ERROR, 
        message: String(message), 
        timestamp: new Date(), 
        data, 
        stack 
      },
      ...logs
    ]);
    return originalError.call(logger, message, data, options);
  };
  
  // Expose debug commands to window
  if (typeof window !== 'undefined') {
    (window as any).debugCommands = {
      toggleDebugPanel: togglePanel,
      navigateWithBypass: navigateWithBypass,
      fetchUserProfile: fetchUserProfile,
      testLogs: testLogLevels,
      clearLogs: clearLogs
    };
  }
  
  // Set up keyboard shortcuts when mounted
  import { onMount, onDestroy } from 'svelte';
  
  let cleanupFunction;
  let logs: {level: LogLevel; message: string; timestamp: Date; data?: any; stack?: string}[] = [];
  
  // Subscribe to logs
  const unsubscribeLogs = logMessages.subscribe(value => {
    logs = value;
  });
  
  // Service status section
  const services = [
    { name: 'API Gateway', url: 'http://localhost:8083/health', http: true },
    // Keep API Gateway proxied endpoints for reference
    { name: 'User Service (via Gateway)', url: 'http://localhost:8083/api/v1/system/status/user', http: true },
    { name: 'Thread Service (via Gateway)', url: 'http://localhost:8083/api/v1/system/status/thread', http: true },
    { name: 'Community Service (via Gateway)', url: 'http://localhost:8083/api/v1/system/status/community', http: true },
    { name: 'Event Bus (via Gateway)', url: 'http://localhost:8083/api/v1/system/status/event-bus', http: true },
    // Add direct health check endpoints
    { name: 'User Service (direct)', url: 'http://localhost:8081/health', http: true },
    { name: 'Thread Service (direct)', url: 'http://localhost:9092/health', http: true },
    { name: 'Community Service (direct)', url: 'http://localhost:9093/health', http: true },
    { name: 'Event Bus (direct)', url: 'http://localhost:8000/health', http: true },
    // { name: 'AI Service', url: 'http://localhost:5000/health', http: true }, // Removed AI service check
    { name: 'Redis', http: false },
    { name: 'User DB', http: false },
    { name: 'Thread DB', http: false },
    { name: 'RabbitMQ', http: false },
  ];
  
  let serviceStatuses = {};
  let serviceErrors = {};
  
  async function checkServiceStatus() {
    for (const service of services) {
      if (!service.http || !service.url) continue;
      try {
        console.log(`Checking service health: ${service.name} at ${service.url}`);
        const res = await fetch(service.url, { 
          method: 'GET',
          headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json'
          },
          // Don't use credentials for health checks
          mode: 'cors' 
        });
        if (res.ok) {
          serviceStatuses[service.name] = true;
          serviceErrors[service.name] = '';
          console.log(`Service ${service.name} is healthy`);
        } else {
          serviceStatuses[service.name] = false;
          let reason = `HTTP ${res.status} - ${res.statusText}`;
          try {
            const text = await res.text();
            if (text) reason += `: ${text}`;
          } catch {}
          serviceErrors[service.name] = reason;
          console.log(`Service ${service.name} is unhealthy: ${reason}`);
        }
      } catch (e) {
        serviceStatuses[service.name] = false;
        serviceErrors[service.name] = (e && (e as any).message) ? (e as any).message : String(e);
        console.log(`Error checking service ${service.name}:`, e);
      }
    }
  }
  
  onMount(() => {
    cleanupFunction = setupKeyboardShortcuts();
    
    // Try to load user profile if authenticated
    if (authState.isAuthenticated) {
      fetchUserProfile();
    }
    
    // Add initial log
    logger.info('Debug panel initialized');
    
    checkServiceStatus();
    const interval = setInterval(checkServiceStatus, 10000);
    return () => clearInterval(interval);
  });
  
  onDestroy(() => {
    if (cleanupFunction) cleanupFunction();
    if (unsubscribeLogs) unsubscribeLogs();
    
    // Clean up body class
    if (typeof document !== 'undefined') {
      document.body.classList.remove('debug-panel-open');
    }
    
    // Restore original logger functions
    logger.trace = originalTrace;
    logger.debug = originalDebug;
    logger.info = originalInfo;
    logger.warn = originalWarn;
    logger.error = originalError;
  });
</script>

<!-- Debug icon button -->
<button 
  class="fixed bottom-4 left-4 z-50 bg-gray-700 text-white p-2 rounded-full shadow-lg opacity-30 hover:opacity-100 transition-opacity"
  on:click={togglePanel}
  title="Toggle Debug Panel"
>
  <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
    <path fill-rule="evenodd" d="M11.3 1.046A1 1 0 0112 2v5h4a1 1 0 01.82 1.573l-7 10A1 1 0 018 18v-5H4a1 1 0 01-.82-1.573l7-10a1 1 0 011.12-.38z" clip-rule="evenodd" />
  </svg>
</button>

<!-- Debug panel modal -->
{#if panelVisible}
  <div class="debug-panel-modal-overlay">
    <div class="debug-panel simplified">
      <!-- Header -->
      <header class="debug-panel-header">
        <h2>Debug Panel</h2>
        <button 
          class="close-btn"
          on:click={togglePanel}
        >
          ✕
        </button>
      </header>
      
      <!-- Simple content -->
      <div class="debug-panel-content">
        <div class="section">
          <h3>Current User</h3>
          <div class="card">
            {#if authState.isAuthenticated}
              <h4>Auth Status: <span class="auth-status-authenticated">Authenticated</span></h4>
              <div class="code-block">
                <pre>{JSON.stringify({
                  userId: authState.userId,
                  isAuthenticated: authState.isAuthenticated,
                  tokenExpires: authState.expiresAt ? new Date(authState.expiresAt).toLocaleString() : null
                }, null, 2)}</pre>
              </div>
              <button class="debug-btn green" on:click={fetchUserProfile}>
                Refresh User Data
              </button>
              
              {#if userProfileInfo}
                <h4 class="mt-4">User Profile</h4>
                <div class="code-block">
                  <pre>{JSON.stringify(userProfileInfo, null, 2)}</pre>
                </div>
              {/if}
            {:else}
              <div class="placeholder warning">
                <p>Not authenticated. No user informationa available.</p>
              </div>
            {/if}
          </div>
        </div>
        
        <div class="section">
          <h3>Logging</h3>
          <div class="card">
            <div class="flex-row">
              <div>
                <label for="logLevel">Log Level:</label>
                <select 
                  id="logLevel"
                  value={currentLogLevel} 
                  on:change={updateLogLevel}
                  class="select-input"
                >
                  {#each logLevelOptions as option}
                    <option value={option.value}>{option.label}</option>
                  {/each}
                </select>
              </div>
              <div>
                <button class="debug-btn orange mr-2" on:click={testLogLevels}>
                  Test Logs
                </button>
                <button class="debug-btn red" on:click={clearLogs}>
                  Clear Logs
                </button>
              </div>
            </div>
            
            <!-- Log display area -->
            <div class="log-display mt-4">
              <h4>Log Output</h4>
              <div class="log-container">
                {#if logs.length === 0}
                  <div class="empty-logs">No logs to display</div>
                {:else}
                  {#each logs as log}
                    <div class="log-entry {getLogLevelColor(log.level)}">
                      <div class="log-header">
                        <span class="log-timestamp">[{log.timestamp.toLocaleTimeString()}]</span>
                        <span class="log-level">[{getLogLevelLabel(log.level)}]</span>
                        <span class="log-message">{log.message}</span>
                      </div>
                      
                      {#if log.data && (log.level === LogLevel.ERROR || log.level === LogLevel.WARN)}
                        <div class="log-details">
                          <div class="log-data">
                            <pre>{safeStringify(log.data)}</pre>
                          </div>
                        </div>
                      {/if}
                      
                      {#if log.stack && log.level === LogLevel.ERROR}
                        <div class="log-stack-trace">
                          <div class="stack-header">Stack Trace:</div>
                          <pre>{log.stack}</pre>
                        </div>
                      {/if}
                    </div>
                  {/each}
                {/if}
              </div>
            </div>
          </div>
        </div>
        
        <div class="section">
          <h3>Redirect with Bypass</h3>
          <div class="card">
            <div class="route-buttons">
              {#each commonRoutes as route}
                <button
                  class="route-btn"
                  on:click={() => navigateWithBypass(route)}
                >
                  {route}
                </button>
              {/each}
            </div>
            <div class="mt-3">
              <input 
                type="text" 
                id="customRoute" 
                placeholder="Enter custom route..." 
                class="text-input"
              />
              <button 
                class="debug-btn blue" 
                on:click={() => {
                  const input = document.getElementById('customRoute') as HTMLInputElement;
                  if (input && input.value) {
                    navigateWithBypass(input.value);
                  }
                }}
              >
                Go
              </button>
            </div>
          </div>
        </div>
        
        <!-- Service Status Section -->
        <div class="section">
          <h3>Service Status</h3>
          <div class="card">
            <ul>
              {#each services as service}
                <li>
                  <span>{service.name}:</span>
                  {#if service.http}
                    <span style="font-weight:bold; color:{serviceStatuses[service.name] === undefined ? 'gray' : serviceStatuses[service.name] ? 'limegreen' : 'red'}">
                      {serviceStatuses[service.name] === undefined ? 'Checking...' : serviceStatuses[service.name] ? 'Active' : 'Inactive'}
                    </span>
                    {#if serviceStatuses[service.name] === false && serviceErrors[service.name]}
                      <span style="color:orange; font-size:12px; margin-left:8px;">{serviceErrors[service.name]}</span>
                    {/if}
                  {:else}
                    <span style="color:gray; font-size:12px; margin-left:8px;">Not HTTP health-checkable</span>
                  {/if}
                </li>
              {/each}
            </ul>
          </div>
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
  /* Debug Panel Modal Overlay */
  .debug-panel-modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.7);
    z-index: 9999;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  /* Debug Panel Container */
  .debug-panel {
    width: 100%;
    max-width: 700px;
    max-height: 90vh;
    background-color: #0f172a;
    color: white;
    border-radius: 8px;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
  
  /* Debug Panel Header */
  .debug-panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    background-color: #1e293b;
    border-bottom: 1px solid #334155;
  }
  
  .debug-panel-header h2 {
    font-size: 20px;
    font-weight: 600;
    color: #3b82f6;
    margin: 0;
  }
  
  .close-btn {
    background: none;
    border: none;
    color: #94a3b8;
    font-size: 18px;
    cursor: pointer;
    width: 28px;
    height: 28px;
    border-radius: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: color 0.2s;
  }
  
  .close-btn:hover {
    color: white;
    background-color: rgba(255, 255, 255, 0.1);
  }
  
  /* Debug Panel Content */
  .debug-panel-content {
    display: flex;
    flex-direction: column;
    flex: 1;
    overflow-y: auto;
    padding: 20px;
  }
  
  /* Sections */
  .section {
    margin-bottom: 24px;
  }
  
  .section h3 {
    font-size: 18px;
    font-weight: 600;
    color: #3b82f6;
    margin-top: 0;
    margin-bottom: 12px;
  }
  
  .section h4 {
    font-size: 14px;
    font-weight: 600;
    color: #94a3b8;
    margin-top: 0;
    margin-bottom: 8px;
  }
  
  /* Cards */
  .card {
    background-color: #1e293b;
    border-radius: 6px;
    padding: 16px;
    margin-bottom: 16px;
  }
  
  /* Code Blocks */
  .code-block {
    background-color: #0f172a;
    border-radius: 6px;
    padding: 12px;
    overflow: auto;
    max-height: 200px;
    font-family: monospace;
    font-size: 12px;
    color: #e2e8f0;
    margin-bottom: 12px;
  }
  
  pre {
    margin: 0;
    white-space: pre-wrap;
  }
  
  code {
    font-family: monospace;
    background-color: #0f172a;
    padding: 2px 4px;
    border-radius: 3px;
    font-size: 12px;
  }
  
  /* Placeholders */
  .placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: #0f172a;
    border-radius: 6px;
    padding: 20px;
    color: #94a3b8;
    font-size: 14px;
    min-height: 100px;
  }
  
  .placeholder.warning {
    color: #fbbf24;
  }
  
  /* Buttons */
  .debug-btn {
    padding: 6px 12px;
    border-radius: 4px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    border: none;
    color: white;
    transition: background-color 0.2s;
  }
  
  .debug-btn.blue {
    background-color: #2563eb;
  }
  
  .debug-btn.blue:hover {
    background-color: #1d4ed8;
  }
  
  .debug-btn.green {
    background-color: #10b981;
  }
  
  .debug-btn.green:hover {
    background-color: #059669;
  }
  
  .debug-btn.orange {
    background-color: #f59e0b;
  }
  
  .debug-btn.orange:hover {
    background-color: #d97706;
  }
  
  .debug-btn.red {
    background-color: #ef4444;
  }
  
  .debug-btn.red:hover {
    background-color: #dc2626;
  }
  
  .mr-2 {
    margin-right: 8px;
  }
  
  /* Route Buttons */
  .route-buttons {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 12px;
  }
  
  .route-btn {
    padding: 6px 12px;
    border-radius: 4px;
    font-size: 14px;
    cursor: pointer;
    background-color: #1e293b;
    color: #e2e8f0;
    border: 1px solid #334155;
    transition: all 0.2s;
  }
  
  .route-btn:hover {
    background-color: #334155;
  }
  
  /* Select Input */
  .select-input {
    background-color: #1e293b;
    color: white;
    border: 1px solid #334155;
    border-radius: 4px;
    padding: 8px 12px;
    font-size: 14px;
    outline: none;
    margin-right: 8px;
  }
  
  .select-input:focus {
    border-color: #3b82f6;
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.3);
  }
  
  /* Text Input */
  .text-input {
    background-color: #1e293b;
    color: white;
    border: 1px solid #334155;
    border-radius: 4px;
    padding: 8px 12px;
    font-size: 14px;
    outline: none;
    margin-right: 8px;
    width: calc(100% - 60px);
  }
  
  .text-input:focus {
    border-color: #3b82f6;
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.3);
  }
  
  /* Utility classes */
  .mt-3 {
    margin-top: 12px;
  }
  
  .mt-4 {
    margin-top: 16px;
  }
  
  .flex-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  
  /* Auth status */
  .auth-status-authenticated {
    color: #4ade80;
  }
  
  /* Log display */
  .log-display {
    margin-top: 16px;
  }
  
  .log-container {
    background-color: #0f172a;
    border-radius: 6px;
    padding: 10px;
    max-height: 300px;
    overflow-y: auto;
    font-family: monospace;
    font-size: 12px;
  }
  
  .log-entry {
    padding: 4px 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    white-space: pre-wrap;
    word-break: break-word;
    margin-bottom: 8px;
  }
  
  .log-header {
    display: flex;
    align-items: flex-start;
  }
  
  .log-timestamp {
    color: #94a3b8;
    margin-right: 6px;
    flex-shrink: 0;
  }
  
  .log-level {
    font-weight: bold;
    margin-right: 6px;
    flex-shrink: 0;
  }
  
  .log-message {
    color: white;
  }
  
  .log-details {
    padding-left: 20px;
    margin-top: 4px;
    font-size: 11px;
    color: #cbd5e1;
  }
  
  .log-data {
    background-color: rgba(0, 0, 0, 0.2);
    border-radius: 4px;
    padding: 6px;
    margin-top: 2px;
  }
  
  .log-stack-trace {
    padding-left: 20px;
    margin-top: 4px;
    font-size: 11px;
    color: #cbd5e1;
    border-left: 2px solid #ef4444;
  }
  
  .stack-header {
    font-weight: bold;
    margin-bottom: 4px;
    color: #f87171;
  }
  
  .log-data pre, .log-stack-trace pre {
    margin: 0;
    white-space: pre-wrap;
    overflow-x: auto;
  }
  
  .empty-logs {
    color: #64748b;
    text-align: center;
    padding: 20px;
    font-style: italic;
  }
  
  /* Log colors */
  .text-red-500 {
    color: #ef4444;
  }
  
  .text-yellow-500 {
    color: #f59e0b;
  }
  
  .text-blue-500 {
    color: #3b82f6;
  }
  
  .text-gray-400 {
    color: #94a3b8;
  }
  
  .text-gray-500 {
    color: #64748b;
  }
</style>