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
  let iconSize = '';
  
  $: {
    switch (size) {
      case 'sm':
        iconSize = '16';
        break;
      case 'lg':
        iconSize = '24';
        break;
      case 'md':
      default:
        iconSize = '20';
        break;
    }
  }
</script>

<button 
  type="button" 
  class="btn-icon {size === 'sm' ? 'btn-icon-sm' : size === 'lg' ? 'btn-icon-lg' : ''}"
  on:click={handleToggle}
  aria-label={$theme === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'}
>
  {#if $theme === 'dark'}
    <div class="flex items-center text-warning">
      <SunIcon size={iconSize} />
      {#if showLabel}
      <span class="ml-2 text-sm text-white">Light Mode</span>
      {/if}
    </div>
  {:else}
    <div class="flex items-center text-primary">
      <MoonIcon size={iconSize} />
      {#if showLabel}
      <span class="ml-2 text-sm">Dark Mode</span>
      {/if}
    </div>
  {/if}
</button>