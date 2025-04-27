import { writable } from 'svelte/store';

// Define theme types
export type Theme = 'dark' | 'light';

// Check for system preference or saved preference
function getInitialTheme(): Theme {
  if (typeof window !== 'undefined') {
    // Check local storage first
    const savedTheme = localStorage.getItem('theme') as Theme;
    if (savedTheme === 'dark' || savedTheme === 'light') {
      return savedTheme;
    }
    
    // Fall back to system preference
    if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
      return 'dark';
    }
  }
  
  // Default to dark theme
  return 'dark';
}

// Create theme store
export const theme = writable<Theme>(getInitialTheme());

// Theme toggler function
export function toggleTheme(): void {
  theme.update(currentTheme => {
    const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
    
    if (typeof window !== 'undefined') {
      localStorage.setItem('theme', newTheme);
      applyThemeToDocument(newTheme);
    }
    
    return newTheme;
  });
}

// Apply theme to document
export function applyThemeToDocument(currentTheme: Theme): void {
  if (typeof document !== 'undefined') {
    const html = document.documentElement;
    
    if (currentTheme === 'dark') {
      html.classList.add('dark-theme');
      html.classList.remove('light-theme');
    } else {
      html.classList.add('light-theme');
      html.classList.remove('dark-theme');
    }
  }
}

// Initialize theme on import
if (typeof window !== 'undefined') {
  const currentTheme = getInitialTheme();
  localStorage.setItem('theme', currentTheme);
  applyThemeToDocument(currentTheme);
}