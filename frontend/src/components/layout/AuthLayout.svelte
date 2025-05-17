<script lang="ts">
  import { useTheme } from '../../hooks/useTheme';
  import Logo from '../common/Logo.svelte';
  import Toast from '../common/Toast.svelte';
  import { onMount } from 'svelte';
  import lightLogo from '../../assets/logo/light-logo.jpeg';
  import darkLogo from '../../assets/logo/dark-logo.jpeg';
  
  export let title = '';
  export let showLogo = true;
  export let showCloseButton = false;
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

<div class="auth-container {isDarkMode ? 'auth-container-dark' : ''}">
  <div class="auth-left">
    <div class="auth-left-logo">
      {#if isDarkMode}
        <img src={lightLogo} alt="AYCOM Logo" class="auth-logo-image" />
      {:else}
        <img src={darkLogo} alt="AYCOM Logo" class="auth-logo-image" />
      {/if}
    </div>
    <div class="auth-left-bg"></div>
  </div>
  
  <div class="auth-right">
    <div class="auth-form">
      <div class="auth-header">
        {#if showBackButton}
          <button 
            class="text-blue-500 hover:text-blue-600 transition-colors absolute top-4 left-4"
            on:click={onBack}
            data-cy="back-button"
            aria-label="Go back"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
          </button>
        {/if}
        
        {#if showLogo}
          <div class="auth-logo">
            {#if isDarkMode}
              <img src={lightLogo} alt="AYCOM Logo" class="auth-header-logo-image" />
            {:else}
              <img src={darkLogo} alt="AYCOM Logo" class="auth-header-logo-image" />
            {/if}
          </div>
        {/if}
        
        {#if title}
          <h1 class="auth-title" data-cy="page-title">{title}</h1>
        {/if}
      </div>
      
      <slot />
    </div>
  </div>
</div>

<style>
  .theme-container {
    background-color: var(--bg-secondary);
    color: var(--text-primary);
    transition: background-color 0.3s ease, color 0.3s ease;
  }
  
  .auth-left-logo {
    z-index: 10;
    display: flex;
    justify-content: center;
    align-items: center;
  }
  
  .auth-logo-image {
    width: 60%;
    height: auto;
    max-width: 400px;
  }
  
  .auth-header-logo-image {
    width: 40px;
    height: 40px;
    object-fit: contain;
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