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
        document.documentElement.classList.add('dark-theme', 'dark');
        document.documentElement.classList.remove('light-theme', 'light');
        
        // For Tailwind dark mode
        document.documentElement.classList.add('dark-mode');
        document.documentElement.classList.remove('light-mode');
      } else {
        document.documentElement.classList.add('light-theme', 'light');
        document.documentElement.classList.remove('dark-theme', 'dark');
        
        // For Tailwind dark mode
        document.documentElement.classList.add('light-mode');
        document.documentElement.classList.remove('dark-mode');
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

<slot />

<style>
  :global(html) {
    transition: background-color 0.3s ease, color 0.3s ease;
  }
  
  :global(html.dark-theme) {
    background-color: #000000;
    color: #ffffff;
  }
  
  :global(html.light-theme) {
    background-color: #ffffff;
    color: #000000;
  }
</style>