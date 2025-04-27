<script lang="ts">
  import { theme, toggleTheme } from '../hooks/useTheme';
  
  export let size: string = "6"; // Default size in Tailwind units (1.5rem)
  export let className: string = ""; // Additional classes
</script>

<button 
  class="p-2 rounded-full transition-colors {className}"
  class:bg-gray-800={$theme === 'light'} 
  class:bg-gray-200={$theme === 'dark'}
  on:click={toggleTheme}
  aria-label="Toggle theme"
>
  {#if $theme === 'dark'}
    <!-- Sun icon for light mode -->
    <svg xmlns="http://www.w3.org/2000/svg" class="h-{size} w-{size} text-yellow-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
    </svg>
  {:else}
    <!-- Moon icon for dark mode -->
    <svg xmlns="http://www.w3.org/2000/svg" class="h-{size} w-{size} text-gray-800" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
    </svg>
  {/if}
</button>