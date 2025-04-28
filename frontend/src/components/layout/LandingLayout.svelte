<script lang="ts">
  import { useTheme } from '../../hooks/useTheme';
  import { onMount } from 'svelte';
  
  // Get the theme store and toggleTheme function from our hook
  const { theme, toggleTheme } = useTheme();
  
  // Reactive declaration to update isDarkMode when theme changes
  $: isDarkMode = $theme === 'dark';
  
  onMount(() => {
    // Apply theme class to document when component mounts
    document.documentElement.classList.add(isDarkMode ? 'dark' : 'light');
  });
</script>

<div class="theme-container {isDarkMode ? 'dark-mode' : 'light-mode'} min-h-screen w-full overflow-x-hidden">
  <div class="{isDarkMode ? 'bg-gray-900 text-white' : 'bg-white text-black'} min-h-screen w-full">
    <button 
      class="absolute top-4 right-4 p-2 rounded-full z-10 {isDarkMode ? 'bg-gray-800 hover:bg-gray-700' : 'bg-gray-100 hover:bg-gray-200'} transition-colors"
      on:click={toggleTheme}
      aria-label="Toggle theme"
    >
      {#if isDarkMode}
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-yellow-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
        </svg>
      {:else}
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-800" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
        </svg>
      {/if}
    </button>
    
    <slot />
  </div>
</div>

<style>
  /* Additional theme-related styles */
  .theme-container {
    transition: background-color 0.3s ease, color 0.3s ease;
  }
  
  .dark-mode {
    --bg-color: #1a202c;
    --text-color: #f7fafc;
  }
  
  .light-mode {
    --bg-color: #ffffff;
    --text-color: #1a202c;
  }
</style> 