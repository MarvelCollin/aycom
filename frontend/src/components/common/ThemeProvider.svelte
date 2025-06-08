<script lang="ts">
  import { onMount } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  
  // Get the theme store from our hook
  const { theme } = useTheme();
  
  // Initialize theme and font preferences on component mount
  onMount(() => {
    // Initialize font preferences from localStorage
    try {
      // Apply font size if stored in localStorage
      const storedFontSize = localStorage.getItem('fontSize');
      if (storedFontSize) {
        document.documentElement.classList.remove('font-small', 'font-medium', 'font-large');
        document.documentElement.classList.add(`font-${storedFontSize}`);
      } else {
        // Default to medium if not set
        document.documentElement.classList.add('font-medium');
      }
      
      // Apply font color if stored in localStorage
      const storedFontColor = localStorage.getItem('fontColor');
      if (storedFontColor) {
        document.documentElement.classList.remove('text-default', 'text-blue', 'text-green', 'text-purple');
        document.documentElement.classList.add(`text-${storedFontColor}`);
        
        // Add class to the html element as well for better specificity
        const htmlElement = document.querySelector('html');
        if (htmlElement) {
          htmlElement.classList.remove('text-default', 'text-blue', 'text-green', 'text-purple');
          htmlElement.classList.add(`text-${storedFontColor}`);
        }
      } else {
        // Default to default if not set
        document.documentElement.classList.add('text-default');
        const htmlElement = document.querySelector('html');
        if (htmlElement) {
          htmlElement.classList.add('text-default');
        }
      }
    } catch (error) {
      console.error('Error initializing font preferences:', error);
    }
    
    // Subscribe to theme changes and apply them
    const unsubscribe = theme.subscribe(currentTheme => {
      // Apply theme to document
      document.documentElement.setAttribute('data-theme', currentTheme);
      
      const htmlElement = document.querySelector('html');
      if (htmlElement) {
        htmlElement.setAttribute('data-theme', currentTheme);
      }
      
      // Add or remove the appropriate theme classes
      if (currentTheme === 'dark') {
        document.documentElement.classList.add('dark-theme');
        document.documentElement.classList.remove('light-theme');
        
        if (htmlElement) {
          htmlElement.classList.add('dark-theme');
          htmlElement.classList.remove('light-theme');
        }
      } else {
        document.documentElement.classList.add('light-theme');
        document.documentElement.classList.remove('dark-theme');
        
        if (htmlElement) {
          htmlElement.classList.add('light-theme');
          htmlElement.classList.remove('dark-theme');
        }
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