<script lang="ts">
  import { useTheme } from '../../hooks/useTheme';
  
  export let size: 'small' | 'medium' | 'large' = 'medium';
  
  // Get the theme store
  const { theme } = useTheme();
  
  // Reactive declaration to update isDarkMode when theme changes
  $: isDarkMode = $theme === 'dark';
  
  // Determine size class based on prop
  $: sizeClass = {
    small: 'h-8 w-auto',
    medium: 'h-12 w-auto',
    large: 'h-16 w-auto'
  }[size];
</script>

<img 
  src={isDarkMode ? "/src/assets/logo/light-logo.jpeg" : "/src/assets/logo/dark-logo.jpeg"} 
  alt="AYCOM Logo" 
  class="{sizeClass} rounded-xl shadow-sm"
/> 