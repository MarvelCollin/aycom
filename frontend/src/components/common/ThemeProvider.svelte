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

<style>
  /* Define theme variables for the entire app */
  :global(:root) {
    /* Common variables for both themes */
    --font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont,
      "Segoe UI", Roboto, "Helvetica Neue", Arial, "Noto Sans", sans-serif;
    --color-primary: #3b82f6;
    --color-success: #10b981;
    --color-danger: #ef4444;
    --color-warning: #f59e0b;
    --transition-speed: 0.3s;
  }
  
  /* Dark theme variables */
  :global(.dark-theme) {
    --bg-primary: #000000;
    --bg-secondary: #1f2937;
    --bg-tertiary: #374151;
    --text-primary: #ffffff;
    --text-secondary: #e5e7eb;
    --text-tertiary: #9ca3af;
    --border-color: #4b5563;
    --hover-bg: #1f2937;
  }
  
  /* Light theme variables */
  :global(.light-theme) {
    --bg-primary: #ffffff;
    --bg-secondary: #f3f4f6;
    --bg-tertiary: #e5e7eb;
    --text-primary: #111827;
    --text-secondary: #374151;
    --text-tertiary: #6b7280;
    --border-color: #d1d5db;
    --hover-bg: #f3f4f6;
  }
  
  /* Apply smooth transitions for theme changes */
  :global(body), :global(body *) {
    transition: background-color var(--transition-speed) ease, 
                color var(--transition-speed) ease,
                border-color var(--transition-speed) ease;
  }
</style>