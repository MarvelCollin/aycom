import { writable } from 'svelte/store';

type ThemeType = 'light' | 'dark';

// Create a writable store for the theme
const createThemeStore = () => {
  // Get stored theme or user's preferred color scheme
  const getInitialTheme = (): ThemeType => {
    try {
      const storedTheme = localStorage.getItem('theme') as ThemeType;
      if (storedTheme) {
        return storedTheme;
      }
      
      // Check user's preferred color scheme
      if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
        return 'dark';
      }
    } catch (error) {
      console.error('Error getting initial theme:', error);
    }
    
    // Default to light
    return 'light';
  };
  
  // Create the writable store with the initial theme
  const theme = writable<ThemeType>('light');
  
  // Initialize when in browser
  if (typeof window !== 'undefined') {
    theme.set(getInitialTheme());
  }
  
  // Listen for system preference changes
  if (typeof window !== 'undefined' && window.matchMedia) {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    
    const handleChange = (event: MediaQueryListEvent) => {
      const currentTheme = localStorage.getItem('theme');
      
      // Only update based on system preference if user hasn't explicitly set a theme
      if (!currentTheme) {
        theme.set(event.matches ? 'dark' : 'light');
      }
    };
    
    // Add the listener
    if (mediaQuery.addEventListener) {
      mediaQuery.addEventListener('change', handleChange);
    } else {
      // For older browsers
      mediaQuery.addListener(handleChange);
    }
  }
  
  return {
    subscribe: theme.subscribe,
    set: (value: ThemeType) => {
      theme.set(value);
      try {
        localStorage.setItem('theme', value);
        
        // Update document with the theme class
        if (typeof document !== 'undefined') {
          document.documentElement.classList.remove('light', 'dark');
          document.documentElement.classList.add(value);
        }
      } catch (error) {
        console.error('Error setting theme:', error);
      }
    }
  };
};

// Create the theme store
const themeStore = createThemeStore();

export function useTheme() {
  /**
   * Toggles between light and dark themes
   */
  const toggleTheme = () => {
    themeStore.update(currentTheme => {
      const newTheme = currentTheme === 'light' ? 'dark' : 'light';
      
      try {
        localStorage.setItem('theme', newTheme);
        
        // Update document with the theme class
        if (typeof document !== 'undefined') {
          document.documentElement.classList.remove('light', 'dark');
          document.documentElement.classList.add(newTheme);
        }
      } catch (error) {
        console.error('Error toggling theme:', error);
      }
      
      return newTheme;
    });
  };
  
  return {
    theme: themeStore,
    toggleTheme
  };
}