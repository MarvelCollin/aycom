<script lang="ts">
  import { onMount } from 'svelte';
  import appConfig from '../config/appConfig';
  
  // Debug state
  let currentRoute = window.location.pathname;
  let allRoutes = [
    '/',
    '/login',
    '/register',
    '/home',
    '/feed',
    '/explore',
    '/notifications',
    '/messages',
    '/profile',
    '/bookmarks',
    '/communities',
    '/premium',
    '/verified-orgs',
    '/more',
    '/grok'
  ];
  
  // Auth state
  let isAuthenticated = localStorage.getItem('aycom_authenticated') === 'true';
  
  // Toggle authentication status
  function toggleAuth() {
    isAuthenticated = !isAuthenticated;
    localStorage.setItem('aycom_authenticated', isAuthenticated.toString());
  }
  
  // Navigate to a route with actual page redirect
  function navigateTo(route: string) {
    // Using window.location.href to trigger a full page redirect
    window.location.href = route;
  }
  
  // Update current route display when navigation happens
  onMount(() => {
    const updateRoute = () => {
      currentRoute = window.location.pathname;
      isAuthenticated = localStorage.getItem('aycom_authenticated') === 'true';
    };
    
    window.addEventListener('popstate', updateRoute);
    
    return () => {
      window.removeEventListener('popstate', updateRoute);
    };
  });
</script>

<div class="debug-container">
  <header class="debug-header">
    <h1>AYCOM Debug Panel</h1>
    <div class="auth-toggle">
      <span>Auth Status: </span>
      <span class={isAuthenticated ? "auth-on" : "auth-off"}>
        {isAuthenticated ? "Authenticated" : "Not Authenticated"}
      </span>
      <button on:click={toggleAuth}>
        {isAuthenticated ? "Log Out" : "Log In"}
      </button>
    </div>
  </header>
  
  <div class="current-state">
    <h2>Current Application State</h2>
    <div class="state-item">
      <strong>Current Route:</strong> {currentRoute}
    </div>
    <div class="state-item">
      <strong>App Config:</strong>
      <pre>{JSON.stringify(appConfig, null, 2)}</pre>
    </div>
  </div>
  
  <div class="navigation">
    <h2>Navigation</h2>
    <p>Click on any route to navigate directly to that page</p>
    <div class="route-buttons">
      {#each allRoutes as route}
        <button 
          class={route === currentRoute ? "active" : ""} 
          on:click={() => navigateTo(route)}
        >
          {route || "/"}
        </button>
      {/each}
    </div>
  </div>
  
  <div class="help-section">
    <h2>Debug Help</h2>
    <ul>
      <li>This debug panel bypasses all router authentication checks</li>
      <li>Toggle authentication to test application with/without auth</li>
      <li>Use the buttons above to navigate to any route directly</li>
      <li>Type "kowlin" anywhere to return to this debug panel</li>
      <li>Authentication status and AppConfig are displayed for reference</li>
    </ul>
  </div>
</div>

<style>
  .debug-container {
    padding: 20px;
    max-width: 1200px;
    margin: 0 auto;
    background-color: #161616;
    color: #e6e6e6;
    min-height: 100vh;
    font-family: system-ui, -apple-system, BlinkMacSystemFont, sans-serif;
  }
  
  .debug-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    padding-bottom: 10px;
    border-bottom: 1px solid #2f2f2f;
  }
  
  h1 {
    font-size: 24px;
    color: #1da1f2;
    margin: 0;
  }
  
  h2 {
    font-size: 20px;
    margin-top: 20px;
    margin-bottom: 10px;
    color: #1da1f2;
  }
  
  .auth-toggle {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  
  .auth-on {
    color: #4caf50;
    font-weight: bold;
  }
  
  .auth-off {
    color: #f44336;
    font-weight: bold;
  }
  
  button {
    padding: 6px 12px;
    background-color: #1da1f2;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 14px;
  }
  
  button:hover {
    background-color: #1a91da;
  }
  
  .current-state {
    background-color: #1e1e1e;
    padding: 15px;
    border-radius: 4px;
    margin-bottom: 20px;
  }
  
  .state-item {
    margin-bottom: 10px;
  }
  
  pre {
    background-color: #282828;
    padding: 10px;
    border-radius: 4px;
    overflow-x: auto;
    margin: 8px 0;
    font-family: monospace;
  }
  
  .navigation {
    background-color: #1e1e1e;
    padding: 15px;
    border-radius: 4px;
    margin-bottom: 20px;
  }
  
  .route-buttons {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 10px;
  }
  
  .route-buttons button {
    background-color: #2f2f2f;
  }
  
  .route-buttons button.active {
    background-color: #1da1f2;
  }
  
  .help-section {
    background-color: #1e1e1e;
    padding: 15px;
    border-radius: 4px;
  }
  
  .help-section ul {
    margin-top: 10px;
    padding-left: 20px;
  }
  
  .help-section li {
    margin-bottom: 8px;
  }
</style>
