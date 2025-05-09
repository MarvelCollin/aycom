import { writable } from 'svelte/store';

// Create a store that mimics the $app/stores page store
export const page = writable({
  url: new URL(window.location.href),
  params: {},
  route: {
    id: window.location.pathname
  }
});

// Function to update the page store when navigation happens
export function updatePageStore() {
  page.update(() => ({
    url: new URL(window.location.href),
    params: getRouteParams(),
    route: {
      id: window.location.pathname
    }
  }));
}

// Helper function to extract route parameters
function getRouteParams() {
  const path = window.location.pathname;
  const params = {};
  
  // Check for user profile route pattern
  const userProfileMatch = path.match(/^\/user\/([^\/]+)$/);
  if (userProfileMatch) {
    params.userId = userProfileMatch[1];
  }
  
  // Add more route patterns as needed
  
  return params;
}

// Initialize the store
updatePageStore();

// Set up listeners to update the store when navigation occurs
if (typeof window !== 'undefined') {
  window.addEventListener('popstate', updatePageStore);
  
  // Intercept pushState and replaceState to update our store
  const originalPushState = history.pushState;
  history.pushState = function(state, title, url) {
    originalPushState.call(this, state, title, url);
    updatePageStore();
  };
  
  const originalReplaceState = history.replaceState;
  history.replaceState = function(state, title, url) {
    originalReplaceState.call(this, state, title, url);
    updatePageStore();
  };
} 