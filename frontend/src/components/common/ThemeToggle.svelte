<script>
  import { useTheme } from '../../hooks/useTheme';
  import { createLoggerWithPrefix } from '../../utils/logger';
  import { SunIcon, MoonIcon } from 'svelte-feather-icons';

  // Props
  export let size = 'md'; // 'sm', 'md', 'lg'
  export let showLabel = false;
  
  // Get the theme store
  const { theme, toggleTheme } = useTheme();
  
  // Logging
  const logger = createLoggerWithPrefix('ThemeToggle');
  
  // Toggle the theme
  function handleToggle() {
    toggleTheme();
    logger.debug('Theme toggled', { newTheme: $theme });
  }

  // Define icon size based on prop
  let iconSizeClass = '';
  let iconSize = '';
  
  $: {
    switch (size) {
      case 'sm':
        iconSizeClass = 'h-4 w-4';
        iconSize = '16';
        break;
      case 'lg':
        iconSizeClass = 'h-6 w-6';
        iconSize = '24';
        break;
      case 'md':
      default:
        iconSizeClass = 'h-5 w-5';
        iconSize = '20';
        break;
    }
  }
</script>

<button 
  type="button" 
  class="rounded-full p-2 hover:bg-gray-200 dark:hover:bg-gray-800 focus:outline-none transition-colors"
  on:click={handleToggle}
  aria-label={$theme === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'}
>
  {#if $theme === 'dark'}
    <div class="flex items-center text-yellow-400">
      <SunIcon size={iconSize} />
      {#if showLabel}
      <span class="ml-2 hidden md:inline text-sm text-white">Light Mode</span>
      {/if}
    </div>
  {:else}
    <div class="flex items-center text-blue-900">
      <MoonIcon size={iconSize} />
      {#if showLabel}
      <span class="ml-2 hidden md:inline text-sm">Dark Mode</span>
      {/if}
    </div>
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

  /* Define light mode colors */
  :global(.light-theme) .theme-toggle {
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }
  
  /* Define dark mode colors */
  :global(.dark-theme) .theme-toggle {
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
  }
</style>