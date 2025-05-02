<script lang="ts">
  import { useTheme } from '../../hooks/useTheme';
  import Logo from '../common/Logo.svelte';
  import Toast from '../common/Toast.svelte';
  import { onMount } from 'svelte';
  
  export let title = '';
  export let showLogo = true;
  export let showCloseButton = true;
  export let showBackButton = false;
  export let onBack = () => {};
  
  // Get theme store
  const { theme } = useTheme();
  
  // Reactive declaration to update isDarkMode when theme changes
  $: isDarkMode = $theme === 'dark';
  
  onMount(() => {
    // Apply theme class to document when component mounts
    document.documentElement.classList.add(isDarkMode ? 'dark' : 'light');
  });
</script>

<!-- Render the Toast component here -->
<Toast />

<div class="theme-container {isDarkMode ? 'dark-mode' : 'light-mode'} min-h-screen w-full flex justify-center items-center p-4">
  <div class="{isDarkMode ? 'bg-gray-900 text-white' : 'bg-white text-gray-900'} w-full max-w-md rounded-lg shadow-lg p-6 transition-colors">
    <div class="flex items-center justify-between mb-6">
      {#if showBackButton}
        <button 
          class="text-blue-500 hover:text-blue-600 transition-colors"
          on:click={onBack}
          data-cy="back-button"
          aria-label="Go back"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
        </button>
      {:else if showCloseButton}
        <a 
          href="/" 
          class="text-blue-500 hover:text-blue-600 transition-colors" 
          data-cy="close-button"
          aria-label="Close"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </a>
      {:else}
        <div class="w-6"></div> <!-- spacer -->
      {/if}
      
      {#if showLogo}
        <div class="mx-auto">
          <Logo size="small" />
        </div>
      {:else}
        <div></div> <!-- empty div for flex layout -->
      {/if}
      
      <div class="w-6"></div> <!-- spacer for balance -->
    </div>
    
    {#if title}
      <h1 class="text-2xl font-bold mb-6 text-center" data-cy="page-title">{title}</h1>
    {/if}
    
    <slot />
  </div>
</div>

<style>
  .theme-container {
    background-color: var(--bg-secondary);
    color: var(--text-primary);
    transition: background-color 0.3s ease, color 0.3s ease;
  }
  
  /* Apply these styles to auth buttons so they look more like Twitter */
  :global(.auth-btn) {
    @apply w-full py-3 rounded-full font-semibold transition-colors;
  }
  
  :global(.auth-btn-primary) {
    @apply bg-blue-500 text-white hover:bg-blue-600;
  }
  
  :global(.auth-btn-secondary) {
    @apply border dark:border-gray-700 border-gray-300 text-blue-500 hover:bg-gray-100 dark:hover:bg-gray-800;
  }
  
  :global(.auth-input) {
    @apply w-full p-2 rounded focus:outline-none focus:ring-2 focus:ring-blue-500;
    background-color: var(--bg-primary);
    border-color: var(--border-color);
  }
</style> 