<script lang="ts">
  import LeftSide from './LeftSide.svelte';
  import RightSide from './RightSide.svelte';
  import Toast from '../common/Toast.svelte';
  import { useTheme } from '../../hooks/useTheme';
  import type { ITrend, ISuggestedFollow } from '../../interfaces/ISocialMedia';
  import { writable } from 'svelte/store';
  import { onMount, tick } from 'svelte';
  import { logStore, LogLevel, clearLogs, setGlobalLogLevel, type LogEntry } from '../../utils/logger';
  import appConfig from '../../config/appConfig';

  export let username = "";
  export let displayName = "";
  export let avatar = "👤";
  export let trends: ITrend[] = [];
  export let suggestedFollows: ISuggestedFollow[] = [];
  
  export let showLeftSidebar = true;
  export let showRightSidebar = true;

  const { theme } = useTheme();
  $: isDarkMode = $theme === 'dark';

  function handleToggleComposeModal() {
    dispatch('toggleComposeModal');
  }

  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();
  
  // Debug Modal logic
  const isDebugVisible = writable(false);
  let activeTab = 'logs';
  let konamiSequence = '';
  
  // Keyboard shortcut: "kowlin" to toggle the modal
  const handleKeyDown = (event: KeyboardEvent) => {    
    if (event.ctrlKey && event.shiftKey && event.key === 'D') {
      event.preventDefault();
      isDebugVisible.update(v => !v);
    }
    
    // Konami code: "kowlin"
    const pressedKey = event.key.toLowerCase();
    konamiSequence = (konamiSequence + pressedKey).slice(-6);
    if (konamiSequence === 'kowlin') {
      isDebugVisible.set(true);
    }
  };
  
  // Logs state
  let logs: LogEntry[] = [];
  let filteredLogs: LogEntry[] = [];
  let sourceFilter = 'all';
  let levelFilter = 'all';
  let searchTerm = '';
  let sources: string[] = [];
  let autoScrollEnabled = true;
  let logsContainer: HTMLElement;
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
  
  // Track if scrolled manually away from bottom
  let userScrolled = false;
  const handleLogScroll = () => {
    if (!logsContainer) return;
    
    const { scrollTop, scrollHeight, clientHeight } = logsContainer;
    const isAtBottom = Math.abs(scrollHeight - clientHeight - scrollTop) < 50;
    
    if (!isAtBottom) {
      userScrolled = true;
    } else {
      userScrolled = false;
    }
  };
  
  // Subscribe to the log store for updates
  const unsubscribeLogs = logStore.subscribe(newLogs => {
    logs = newLogs;
    applyLogFilters();
    
    // Scroll logs to bottom on update if auto-scroll is enabled
    tick().then(() => {
      if (autoScrollEnabled && !userScrolled && logsContainer) {
        logsContainer.scrollTop = logsContainer.scrollHeight;
      }
    });
    
    // Update unique sources
    const uniqueSources = new Set<string>();
    newLogs.forEach(log => uniqueSources.add(log.source));
    sources = Array.from(uniqueSources).sort();
  });
  
  // Filter logic for logs
  function applyLogFilters() {
    filteredLogs = logs.filter(log => {
      // Source filter
      if (sourceFilter !== 'all' && log.source !== sourceFilter) {
        return false;
      }
      
      // Level filter
      if (levelFilter !== 'all') {
        const logLevelValue = log.level;
        const filterLevelValue = 
          levelFilter === 'trace' ? LogLevel.TRACE :
          levelFilter === 'debug' ? LogLevel.DEBUG :
          levelFilter === 'info' ? LogLevel.INFO :
          levelFilter === 'warn' ? LogLevel.WARN :
          levelFilter === 'error' ? LogLevel.ERROR : -1;
        
        if (logLevelValue !== filterLevelValue) {
          return false;
        }
      }
      
      // Search term
      if (searchTerm && !log.message.toLowerCase().includes(searchTerm.toLowerCase()) &&
          !log.source.toLowerCase().includes(searchTerm.toLowerCase())) {
        return false;
      }
      
      return true;
    });
  }
  
  // Handle filter changes
  function handleFilterChange() {
    applyLogFilters();
  }
  
  // Clear logs
  function handleClearLogs() {
    clearLogs();
  }
  
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
  
  // Get color for log level
  function getLogLevelColor(level: LogLevel): string {
    switch (level) {
      case LogLevel.TRACE: return 'text-gray-400';
      case LogLevel.DEBUG: return 'text-blue-400';
      case LogLevel.INFO: return 'text-green-400';
      case LogLevel.WARN: return 'text-yellow-400';
      case LogLevel.ERROR: return 'text-red-400';
      default: return 'text-white';
    }
  }
  
  // Format timestamp
  function formatTimestamp(timestamp: string): string {
    const date = new Date(timestamp);
    return date.toLocaleTimeString(undefined, { 
      hour: '2-digit', 
      minute: '2-digit', 
      second: '2-digit',
      fractionalSecondDigits: 3 
    });
  }
  
  onMount(() => {
    const updateRoute = () => {
      currentRoute = window.location.pathname;
      isAuthenticated = localStorage.getItem('aycom_authenticated') === 'true';
    };
    
    window.addEventListener('popstate', updateRoute);
    window.addEventListener('keydown', handleKeyDown);
    
    // Expose toggleDebug to window
    if (typeof window !== 'undefined') {
      (window as any).toggleDebug = () => isDebugVisible.update(v => !v);
    }
    
    return () => {
      window.removeEventListener('popstate', updateRoute);
      window.removeEventListener('keydown', handleKeyDown);
      unsubscribeLogs();
    };
  });
</script>

<div class="flex relative w-full h-screen {isDarkMode ? 'bg-black text-white' : 'bg-white text-black'}">
  {#if showLeftSidebar}
    <div class="fixed left-0 top-0 z-40 h-screen border-r {isDarkMode ? 'border-gray-800' : 'border-gray-200'} overflow-y-auto {isDarkMode ? 'bg-black' : 'bg-white'}" style="width: 275px;">
      <LeftSide 
        {username}
        {displayName}
        {avatar}
        {isDarkMode}
        on:toggleComposeModal={handleToggleComposeModal}
      />
    </div>
    <div class="flex-shrink-0" style="width: 275px;"></div>
  {/if}
  
  <main class="flex-grow h-screen overflow-y-auto relative {isDarkMode ? 'bg-black' : 'bg-white'}">
    <slot></slot>
  </main>
  
  {#if showRightSidebar}
    <div class="hidden md:block fixed right-0 top-0 z-40 h-screen {isDarkMode ? 'bg-black' : 'bg-white'} border-l {isDarkMode ? 'border-gray-800' : 'border-gray-200'} overflow-y-auto" style="width: 350px;">
      <div class="p-4">
        <RightSide 
          {isDarkMode}
          {trends}
          suggestedFollows={suggestedFollows}
        />
      </div>
    </div>
    <div class="hidden md:block flex-shrink-0" style="width: 350px;"></div>
  {/if}
  
  <Toast />
  
  <!-- Debug Modal -->
  {#if $isDebugVisible}
  <div 
    class="debug-overlay" 
    on:click={() => isDebugVisible.set(false)}
    on:keydown={(e) => e.key === 'Escape' && isDebugVisible.set(false)}
    role="dialog"
    aria-modal="true"
    tabindex="0"
  >
    <div class="debug-modal" on:click|stopPropagation>
      <div class="debug-header">
        <h1>AYCOM Debug Panel</h1>
        <div class="debug-tabs">
          <button 
            class={activeTab === 'logs' ? 'active' : ''} 
            on:click={() => activeTab = 'logs'}
          >
            Logs
          </button>
          <button 
            class={activeTab === 'navigation' ? 'active' : ''} 
            on:click={() => activeTab = 'navigation'}
          >
            Navigation
          </button>
          <button 
            class={activeTab === 'help' ? 'active' : ''} 
            on:click={() => activeTab = 'help'}
          >
            Help
          </button>
        </div>
        <div class="auth-toggle">
          <span>Auth: </span>
          <span class={isAuthenticated ? "auth-on" : "auth-off"}>
            {isAuthenticated ? "✓" : "✗"}
          </span>
          <button class="toggle-btn" on:click={toggleAuth}>
            {isAuthenticated ? "Log Out" : "Log In"}
          </button>
          <button class="close-btn" on:click={() => isDebugVisible.set(false)}>
            ✕
          </button>
        </div>
      </div>
      
      <!-- Logs Tab -->
      {#if activeTab === 'logs'}
        <div class="debug-content">
          <div class="filter-bar">
            <div class="filter-group">
              <select bind:value={sourceFilter} on:change={handleFilterChange}>
                <option value="all">All Sources</option>
                {#each sources as source}
                  <option value={source}>{source}</option>
                {/each}
              </select>
              
              <select bind:value={levelFilter} on:change={handleFilterChange}>
                <option value="all">All Levels</option>
                <option value="trace">TRACE</option>
                <option value="debug">DEBUG</option>
                <option value="info">INFO</option>
                <option value="warn">WARN</option>
                <option value="error">ERROR</option>
              </select>
              
              <input 
                type="text" 
                placeholder="Search logs..." 
                bind:value={searchTerm} 
                on:input={handleFilterChange}
              />
            </div>
            
            <div class="filter-actions">
              <label class="auto-scroll">
                <input type="checkbox" bind:checked={autoScrollEnabled} />
                <span>Auto-scroll</span>
              </label>
              <button class="clear-btn" on:click={handleClearLogs}>Clear</button>
            </div>
          </div>
          
          <div class="logs-container" bind:this={logsContainer} on:scroll={handleLogScroll}>
            {#if filteredLogs.length === 0}
              <div class="empty-logs">
                <p>No logs to display</p>
                {#if logs.length > 0 && filteredLogs.length === 0}
                  <p class="hint">Try changing your filters</p>
                {/if}
              </div>
            {:else}
              <table class="logs-table">
                <thead>
                  <tr>
                    <th class="time-col">Time</th>
                    <th class="level-col">Level</th>
                    <th class="source-col">Source</th>
                    <th class="message-col">Message</th>
                  </tr>
                </thead>
                <tbody>
                  {#each filteredLogs as log}
                    <tr class="log-row">
                      <td class="time-col">{formatTimestamp(log.timestamp)}</td>
                      <td class="level-col">
                        <span class={getLogLevelColor(log.level)}>{log.levelName}</span>
                      </td>
                      <td class="source-col">{log.source}</td>
                      <td class="message-col">
                        <div class="message">{log.message}</div>
                        {#if log.data}
                          <details class="data-details">
                            <summary>Data</summary>
                            <pre>{JSON.stringify(log.data, null, 2)}</pre>
                          </details>
                        {/if}
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            {/if}
          </div>
          
          <div class="logs-status">
            <span>Showing {filteredLogs.length} of {logs.length} logs</span>
          </div>
        </div>
      {/if}
      
      <!-- Navigation Tab -->
      {#if activeTab === 'navigation'}
        <div class="debug-content">
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
        </div>
      {/if}
      
      <!-- Help Tab -->
      {#if activeTab === 'help'}
        <div class="debug-content">
          <h2>Debug Panel Usage</h2>
          <div class="help-section">
            <h3>Keyboard Shortcuts</h3>
            <ul>
              <li><code>Ctrl+Shift+D</code>: Toggle debug panel</li>
              <li>Type <code>kowlin</code> anywhere to open debug panel</li>
            </ul>
            
            <h3>Logs</h3>
            <ul>
              <li>Filter logs by source, level, or search term</li>
              <li>Toggle auto-scroll to follow new logs</li>
              <li>Click "Clear" to remove all logs</li>
              <li>Click on "Data" to expand additional log information</li>
            </ul>
            
            <h3>Navigation</h3>
            <ul>
              <li>View current application state</li>
              <li>Navigate to any route directly</li>
              <li>Toggle authentication state for testing</li>
            </ul>
            
            <h3>Console Commands</h3>
            <ul>
              <li><code>window.toggleDebug()</code>: Toggle debug panel</li>
              <li><code>window.logger.info('message')</code>: Log a message</li>
              <li><code>window.setLogLevel(window.LogLevel.DEBUG)</code>: Set log level</li>
              <li><code>window.clearLogs()</code>: Clear all logs</li>
            </ul>
          </div>
        </div>
      {/if}
    </div>
  </div>
  {/if}
</div>

<style>
  /* Original MainLayout styles (preserving them) */
  /* ... existing code ... */
  
  /* Debug modal styles */
  .debug-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.7);
    z-index: 9999;
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 20px;
  }
  
  .debug-modal {
    background-color: #161616;
    color: #e6e6e6;
    width: 90%;
    max-width: 1200px;
    height: 80vh;
    border-radius: 8px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
    display: flex;
    flex-direction: column;
    font-family: system-ui, -apple-system, BlinkMacSystemFont, sans-serif;
  }
  
  .debug-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 10px 20px;
    border-bottom: 1px solid #2f2f2f;
    flex-shrink: 0;
  }
  
  .debug-tabs {
    display: flex;
    gap: 10px;
  }
  
  .debug-tabs button {
    background: transparent;
    border: none;
    color: #aaa;
    padding: 6px 12px;
    cursor: pointer;
    border-radius: 4px;
  }
  
  .debug-tabs button.active {
    background-color: #1da1f2;
    color: white;
  }
  
  .debug-content {
    flex: 1;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }
  
  .filter-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 10px 15px;
    background-color: #1e1e1e;
    border-bottom: 1px solid #2f2f2f;
  }
  
  .filter-group {
    display: flex;
    gap: 10px;
    flex: 1;
  }
  
  .filter-group select,
  .filter-group input {
    background-color: #2a2a2a;
    border: 1px solid #3a3a3a;
    color: #e6e6e6;
    padding: 5px 10px;
    border-radius: 4px;
  }
  
  .filter-group input {
    flex: 1;
  }
  
  .filter-actions {
    display: flex;
    align-items: center;
    gap: 15px;
  }
  
  .auto-scroll {
    display: flex;
    align-items: center;
    gap: 5px;
    color: #aaa;
  }
  
  .auto-scroll input {
    margin: 0;
  }
  
  .logs-container {
    flex: 1;
    overflow-y: auto;
    background-color: #1a1a1a;
  }
  
  .logs-table {
    width: 100%;
    border-collapse: collapse;
    font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
    font-size: 12px;
  }
  
  .logs-table thead {
    position: sticky;
    top: 0;
    background-color: #212121;
    z-index: 1;
  }
  
  .logs-table th {
    text-align: left;
    padding: 8px 10px;
    font-weight: normal;
    color: #999;
    border-bottom: 1px solid #2f2f2f;
  }
  
  .logs-table td {
    padding: 4px 10px;
    vertical-align: top;
    border-bottom: 1px solid #252525;
  }
  
  .logs-table tr:hover {
    background-color: #252525;
  }
  
  .time-col {
    width: 100px;
    white-space: nowrap;
  }
  
  .level-col {
    width: 70px;
    white-space: nowrap;
  }
  
  .source-col {
    width: 120px;
    white-space: nowrap;
  }
  
  .message-col {
    min-width: 300px;
  }
  
  .message {
    word-break: break-word;
  }
  
  .data-details {
    margin-top: 4px;
    cursor: pointer;
  }
  
  .data-details summary {
    color: #999;
    font-size: 11px;
  }
  
  .data-details pre {
    margin: 5px 0;
    padding: 5px 8px;
    background-color: #232323;
    border-radius: 4px;
    overflow-x: auto;
    max-height: 300px;
  }
  
  .empty-logs {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: #777;
  }
  
  .empty-logs .hint {
    font-size: 12px;
    color: #555;
    margin-top: 8px;
  }
  
  .logs-status {
    padding: 5px 15px;
    font-size: 12px;
    color: #777;
    background-color: #1e1e1e;
    border-top: 1px solid #2f2f2f;
  }
  
  h1 {
    font-size: 18px;
    color: #1da1f2;
    margin: 0;
    white-space: nowrap;
  }
  
  h2 {
    font-size: 18px;
    margin: 15px 0 10px;
    color: #1da1f2;
    padding: 0 15px;
  }
  
  h3 {
    font-size: 16px;
    margin: 12px 0 8px;
    color: #e6e6e6;
  }
  
  .auth-toggle {
    display: flex;
    align-items: center;
    gap: 8px;
    white-space: nowrap;
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
    padding: 5px 10px;
    cursor: pointer;
    font-size: 13px;
  }
  
  .toggle-btn {
    background-color: #1da1f2;
    color: white;
    border: none;
    border-radius: 4px;
  }
  
  .close-btn {
    background-color: transparent;
    color: #aaa;
    border: none;
    font-size: 16px;
    padding: 0 6px;
  }
  
  .clear-btn {
    background-color: #333;
    color: #ddd;
    border: none;
    border-radius: 4px;
  }
  
  .current-state {
    background-color: #1e1e1e;
    margin: 15px;
    padding: 0 0 15px;
    border-radius: 4px;
  }
  
  .state-item {
    margin: 10px 15px;
  }
  
  .navigation {
    background-color: #1e1e1e;
    margin: 15px;
    padding: 0 0 15px;
    border-radius: 4px;
  }
  
  .navigation p {
    padding: 0 15px;
  }
  
  .route-buttons {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 10px;
    padding: 0 15px;
  }
  
  .route-buttons button {
    background-color: #2f2f2f;
    border: none;
    border-radius: 4px;
    color: #ddd;
  }
  
  .route-buttons button.active {
    background-color: #1da1f2;
  }
  
  .help-section {
    padding: 0 15px 15px;
    background-color: #1e1e1e;
    margin: 0 15px 15px;
    border-radius: 4px;
  }
  
  .help-section ul {
    margin: 8px 0;
    padding-left: 20px;
  }
  
  .help-section li {
    margin-bottom: 6px;
  }
  
  .help-section code {
    background-color: #2a2a2a;
    padding: 2px 4px;
    border-radius: 3px;
    font-family: monospace;
  }
</style>