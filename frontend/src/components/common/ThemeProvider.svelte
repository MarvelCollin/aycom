<script lang="ts">
  import { onMount } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  
  // Get the theme store from our hook
  const { theme } = useTheme();
  
  // Initialize theme on component mount
  onMount(() => {
    // Subscribe to theme changes and apply them
    const unsubscribe = theme.subscribe(currentTheme => {
      // Apply theme to document
      document.documentElement.setAttribute('data-theme', currentTheme);
      
      // Add or remove the appropriate theme classes
      if (currentTheme === 'dark') {
        document.documentElement.classList.add('dark-theme');
        document.documentElement.classList.remove('light-theme');
      } else {
        document.documentElement.classList.add('light-theme');
        document.documentElement.classList.remove('dark-theme');
      }
      
      // Update the body class as well for component-specific styling
      document.body.setAttribute('data-theme', currentTheme);
    });
    
    // Clean up subscription
    return () => {
      unsubscribe();
    };
  });
</script>

<div class="theme-provider">
  <slot />
</div>

<style>
  .theme-provider {
    display: contents;
  }
</style>