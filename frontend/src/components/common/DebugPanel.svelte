<script lang="ts">
  import { LogLevel, setGlobalLogLevel, logger } from '../../utils/logger';
  import { writable } from 'svelte/store';
  
  // Store for debug panel visibility
  const isVisible = writable(false);
  
  // Toggle the debug panel visibility
  function togglePanel() {
    isVisible.update(value => !value);
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
  
  // Environment info
  const envInfo = {
    isDev: import.meta.env.DEV,
    isProd: import.meta.env.PROD,
    mode: import.meta.env.MODE
  };
  
  // System info
  const systemInfo = {
    userAgent: typeof navigator !== 'undefined' ? navigator.userAgent : 'Unknown',
    language: typeof navigator !== 'undefined' ? navigator.language : 'Unknown',
    screenWidth: typeof window !== 'undefined' ? window.screen.width : 'Unknown',
    screenHeight: typeof window !== 'undefined' ? window.screen.height : 'Unknown'
  };
  
  // API debug helper function
  function testApiEndpoint() {
    logger.info('Testing API endpoint', null, { showToast: true });
    fetch('/api/v1/health')
      .then(response => response.json())
      .then(data => {
        logger.info('API test successful', { data }, { showToast: true });
      })
      .catch(error => {
        logger.error('API test failed', { error }, { showToast: true });
      });
  }
  
  // Test logging at each level
  function testLogLevels() {
    logger.trace('This is a TRACE message');
    logger.debug('This is a DEBUG message');
    logger.info('This is an INFO message', null, { showToast: true });
    logger.warn('This is a WARN message', null, { showToast: true });
    logger.error('This is an ERROR message', null, { showToast: true });
  }
  
  // Expose debug commands to window
  if (typeof window !== 'undefined') {
    (window as any).debugCommands = {
      testApi: testApiEndpoint,
      testLogs: testLogLevels,
      toggleDebugPanel: togglePanel
    };
  }
</script>

<!-- Debug icon in bottom left corner -->
<button 
  class="fixed bottom-4 left-4 z-50 bg-gray-700 text-white p-2 rounded-full shadow-lg opacity-30 hover:opacity-100 transition-opacity"
  on:click={togglePanel}
  title="Toggle Debug Panel"
>
  <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
    <path fill-rule="evenodd" d="M11.3 1.046A1 1 0 0112 2v5h4a1 1 0 01.82 1.573l-7 10A1 1 0 018 18v-5H4a1 1 0 01-.82-1.573l7-10a1 1 0 011.12-.38z" clip-rule="evenodd" />
  </svg>
</button>

<!-- Debug panel -->
{#if $isVisible}
  <div class="fixed bottom-16 left-4 z-50 bg-gray-800 text-white p-4 rounded-lg shadow-lg max-w-md overflow-auto max-h-[80vh]">
    <div class="flex justify-between items-center mb-4">
      <h2 class="text-lg font-semibold">Debug Panel</h2>
      <button class="text-gray-400 hover:text-white" on:click={togglePanel}>
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
        </svg>
      </button>
    </div>
    
    <div class="mb-4">
      <label class="block text-sm font-medium mb-1">Log Level:</label>
      <select 
        value={currentLogLevel} 
        on:change={updateLogLevel}
        class="w-full bg-gray-700 border border-gray-600 rounded py-1 px-2 text-white"
      >
        {#each logLevelOptions as option}
          <option value={option.value}>{option.label}</option>
        {/each}
      </select>
    </div>
    
    <div class="mb-4">
      <div class="flex space-x-2">
        <button 
          class="bg-blue-600 text-white px-3 py-1 rounded hover:bg-blue-700 text-sm"
          on:click={testApiEndpoint}
        >
          Test API
        </button>
        <button 
          class="bg-green-600 text-white px-3 py-1 rounded hover:bg-green-700 text-sm"
          on:click={testLogLevels}
        >
          Test Logs
        </button>
      </div>
    </div>
    
    <div class="mb-4">
      <h3 class="text-sm font-semibold mb-1">Environment:</h3>
      <div class="bg-gray-700 rounded p-2 text-xs">
        <pre>{JSON.stringify(envInfo, null, 2)}</pre>
      </div>
    </div>
    
    <div class="mb-4">
      <h3 class="text-sm font-semibold mb-1">System:</h3>
      <div class="bg-gray-700 rounded p-2 text-xs overflow-x-auto">
        <pre>{JSON.stringify(systemInfo, null, 2)}</pre>
      </div>
    </div>
    
    <div class="text-xs text-gray-400 mt-2">
      <p>Debug commands available in console:</p>
      <ul class="list-disc list-inside mt-1">
        <li>window.debugCommands.testApi()</li>
        <li>window.debugCommands.testLogs()</li>
        <li>window.debugCommands.toggleDebugPanel()</li>
        <li>window.logger.setLevel(window.LogLevel.DEBUG)</li>
      </ul>
    </div>
  </div>
{/if} 