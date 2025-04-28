<script lang="ts">
  import { useTheme } from '../../hooks/useTheme';
  
  // Get the theme store and toggleTheme function from our hook
  const { theme, toggleTheme } = useTheme();
  
  // Reactive declaration to update isDarkMode when theme changes
  $: isDarkMode = $theme === 'dark';
  
  // Optional position props with defaults
  export let position = 'fixed';
  export let top = '4';
  export let right = '4';
</script>

<button 
  class="pos-{position} top-{top} right-{right} p-2 rounded-full z-10 {isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}"
  on:click={toggleTheme}
  aria-label="Toggle theme"
>
  {#if isDarkMode}
    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-yellow-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
    </svg>
  {:else}
    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-gray-800" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
    </svg>
  {/if}
</button>

<style>
  .theme-toggle {
    border: none;
    outline: none;
    cursor: pointer;
    transform: scale(1);
  }
  
  .theme-toggle:hover {
    transform: scale(1.05);
  }
  
  .theme-toggle:active {
    transform: scale(0.95);
  }
  
  /* Animation for icon transition */
  .theme-toggle svg {
    transition: transform 0.5s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  }
  
  .theme-toggle:hover svg {
    transform: rotate(15deg);
  }
</style>