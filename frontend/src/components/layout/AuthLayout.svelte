<script lang="ts">
  import { useTheme } from '../../hooks/useTheme';
  import Logo from '../common/Logo.svelte';
  
  export let title = '';
  export let showLogo = true;
  export let showCloseButton = true;
  export let showBackButton = false;
  export let onBack = () => {};
  
  // Get theme store
  const { theme } = useTheme();
  
  // Reactive declaration to update isDarkMode when theme changes
  $: isDarkMode = $theme === 'dark';
</script>

<div class="{isDarkMode ? 'bg-black text-white' : 'bg-white text-black'} min-h-screen w-full flex justify-center items-center p-4">
  <div class="w-full max-w-md bg-dark-900 rounded-lg shadow-lg p-6">
    <div class="flex items-center justify-between mb-6">
      {#if showBackButton}
        <button 
          class="text-blue-500 hover:text-blue-600 transition-colors"
          on:click={onBack}
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
        </button>
      {:else if showCloseButton}
        <a href="/" class="text-blue-500 hover:text-blue-600 transition-colors">
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
      <h1 class="text-2xl font-bold mb-6 text-center">{title}</h1>
    {/if}
    
    <slot />
  </div>
</div> 