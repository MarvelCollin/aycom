<script>
  import { useTheme } from '../../hooks/useTheme';
  
  // Access the theme store and toggle function
  const { theme, toggleTheme } = useTheme();
  
  // Props
  export let size = 'md'; // sm, md, lg
  
  // Compute the button size classes based on the size prop
  $: buttonSizeClass = 
    size === 'sm' ? 'w-8 h-8' : 
    size === 'lg' ? 'w-12 h-12' : 
    'w-10 h-10';
    
  // Compute the icon size based on the size prop
  $: iconSizeClass = 
    size === 'sm' ? 'w-4 h-4' : 
    size === 'lg' ? 'w-6 h-6' : 
    'w-5 h-5';
</script>

<button 
  on:click={toggleTheme} 
  class="theme-toggle {buttonSizeClass} rounded-full flex items-center justify-center transition-all duration-300 ease-in-out {$theme === 'dark' ? 'bg-gray-800 hover:bg-gray-700' : 'bg-white hover:bg-gray-100'} shadow-md"
  aria-label="Toggle theme"
>
  {#if $theme === 'dark'}
    <!-- Sun icon for dark mode -->
    <svg xmlns="http://www.w3.org/2000/svg" class="{iconSizeClass} text-yellow-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
    </svg>
  {:else}
    <!-- Moon icon for light mode -->
    <svg xmlns="http://www.w3.org/2000/svg" class="{iconSizeClass} text-blue-900" fill="none" viewBox="0 0 24 24" stroke="currentColor">
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