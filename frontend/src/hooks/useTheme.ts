import { writable } from 'svelte/store';

type ThemeType = 'light' | 'dark';

// Create a writable store for the theme
const createThemeStore = () => {
  // Get stored theme or user's preferred color scheme
  const getInitialTheme = (): ThemeType => {
    try {
      // First check localStorage
      const storedTheme = localStorage.getItem('theme') as ThemeType;
      if (storedTheme && (storedTheme === 'light' || storedTheme === 'dark')) {
        return storedTheme;
      }
      
      // Then check system preference
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
  
  // Function to apply theme to DOM - moved here before it's used
  const applyThemeToDOM = (themeValue: ThemeType) => {
    if (typeof document === 'undefined') return;
    
    // Apply to document element
    document.documentElement.classList.remove('light', 'dark', 'light-theme', 'dark-theme', 'light-mode', 'dark-mode');
    document.documentElement.classList.add(themeValue, `${themeValue}-theme`, `${themeValue}-mode`);
    document.documentElement.setAttribute('data-theme', themeValue);
    
    // Apply to body element as well
    document.body.setAttribute('data-theme', themeValue);
  };
  
  // Initialize when in browser
  if (typeof window !== 'undefined') {
    const initialTheme = getInitialTheme();
    theme.set(initialTheme);

    // Apply theme classes immediately
    if (typeof document !== 'undefined') {
      applyThemeToDOM(initialTheme);
    }
  }
  
  // Listen for system preference changes
  if (typeof window !== 'undefined' && window.matchMedia) {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    
    const handleChange = (event: MediaQueryListEvent) => {
      const currentTheme = localStorage.getItem('theme');
      
      // Only update based on system preference if user hasn't explicitly set a theme
      if (!currentTheme) {
        const newTheme = event.matches ? 'dark' : 'light';
        theme.set(newTheme);
        applyThemeToDOM(newTheme);
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
          applyThemeToDOM(value);
        }
      } catch (error) {
        console.error('Error setting theme:', error);
      }
    },
    // Add update method to the store
    update: (callback: (value: ThemeType) => ThemeType) => {
      // Initialize with a default value to satisfy TypeScript
      let currentValue: ThemeType = 'light';
      
      // Get current value
      const unsubscribe = theme.subscribe((value) => {
        currentValue = value;
      });
      unsubscribe();
      
      // Calculate new value
      const newValue = callback(currentValue);
      
      // Set the new value
      theme.set(newValue);
      
      // Update localStorage and classes
      try {
        localStorage.setItem('theme', newValue);
        
        // Update document with the theme class
        if (typeof document !== 'undefined') {
          applyThemeToDOM(newValue);
        }
      } catch (error) {
        console.error('Error updating theme:', error);
      }
      
      return newValue;
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
      return newTheme;
    });
  };
  
  return {
    theme: themeStore,
    toggleTheme
  };
}