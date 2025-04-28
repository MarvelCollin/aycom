<script lang="ts">
  import { useTheme } from '../../hooks/useTheme';
  
  export let size = 'medium'; // small, medium, large
  
  // Get theme store
  const { theme } = useTheme();
  
  // Reactive declaration to update isDarkMode when theme changes
  $: isDarkMode = $theme === 'dark';
  
  // Size classes based on the size prop
  $: sizeClass = size === 'small' ? 'h-8 w-auto' : 
                size === 'large' ? 'h-16 w-16' : 'h-12 w-12';
</script>

{#if size === 'large'}
  <div class="bg-blue-500 rounded-full {sizeClass} flex items-center justify-center text-2xl font-bold">
    AY
  </div>
{:else}
  <img 
    src={isDarkMode ? "/src/assets/logo/light-logo.jpeg" : "/src/assets/logo/dark-logo.jpeg"} 
    alt="AYCOM Logo" 
    class="{sizeClass}"
  />
{/if} 