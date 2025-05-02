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
      
      // Add or remove the dark class for convenience
      if (currentTheme === 'dark') {
        document.documentElement.classList.add('dark-theme');
        document.documentElement.classList.remove('light-theme');
      } else {
        document.documentElement.classList.add('light-theme');
        document.documentElement.classList.remove('dark-theme');
      }
    });
    
    // Clean up subscription
    return () => {
      unsubscribe();
    };
  });
</script>

<slot />