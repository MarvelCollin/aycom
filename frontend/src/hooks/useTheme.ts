import { writable } from 'svelte/store';

// Check if we're in a browser environment
const isBrowser = typeof window !== 'undefined';

// Get the user's preferred theme from localStorage or system preference
const getUserPreference = (): 'light' | 'dark' => {
  // Check for saved theme preference
  const savedTheme = isBrowser ? localStorage.getItem('theme') : null;
  
  if (savedTheme === 'light' || savedTheme === 'dark') {
    return savedTheme;
  }
  
  // Check for system preference
  if (isBrowser && window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
    return 'dark';
  }
  
  // Default to light
  return 'light';
};

// Create a theme store
const theme = writable<'light' | 'dark'>(getUserPreference());

// Subscribe to theme changes to update localStorage
if (isBrowser) {
  theme.subscribe((value) => {
    localStorage.setItem('theme', value);
    document.documentElement.setAttribute('data-theme', value);
    
    // Also add/remove the dark class on the html element for convenience
    if (value === 'dark') {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  });

  // Listen for system theme changes
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (event) => {
    // Only update if user hasn't explicitly set a preference
    if (!localStorage.getItem('theme')) {
      theme.set(event.matches ? 'dark' : 'light');
    }
  });
}

// Export the theme hook
export function useTheme() {
  const toggleTheme = () => {
    theme.update((current) => (current === 'light' ? 'dark' : 'light'));
  };

  const setTheme = (newTheme: 'light' | 'dark') => {
    theme.set(newTheme);
  };

  return {
    theme,
    toggleTheme,
    setTheme
  };
}