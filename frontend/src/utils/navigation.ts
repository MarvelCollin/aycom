import { toastStore } from '../stores/toastStore';

/**
 * Navigate to a specific route
 * @param route The route to navigate to
 * @param options Options for navigation
 */
export function navigate(
  route: string, 
  options: { 
    replace?: boolean,
    showToast?: boolean,
    toastMessage?: string,
    toastType?: 'info' | 'success' | 'warning' | 'error'
  } = {}
) {
  const { 
    replace = false, 
    showToast = false, 
    toastMessage,
    toastType = 'info' 
  } = options;
  
  if (showToast && toastMessage) {
    toastStore.showToast(toastMessage, toastType);
  }
  
  if (replace) {
    window.history.replaceState({}, '', route);
  } else {
    window.history.pushState({}, '', route);
  }
  
  // Dispatch a custom navigation event
  window.dispatchEvent(new CustomEvent('navigate', { detail: { route } }));
}

/**
 * Navigate to the login page
 * @param options Options for navigation
 */
export function navigateToLogin(
  options: {
    replace?: boolean,
    showToast?: boolean,
    toastMessage?: string
  } = {}
) {
  const {
    replace = true,
    showToast = true,
    toastMessage = 'You need to log in to access this feature'
  } = options;
  
  navigate('/login', {
    replace,
    showToast,
    toastMessage,
    toastType: 'warning'
  });
}

/**
 * Navigate to the home/feed page
 * @param options Options for navigation
 */
export function navigateToHome(
  options: {
    replace?: boolean,
    showToast?: boolean,
    toastMessage?: string,
    toastType?: 'info' | 'success' | 'warning' | 'error'
  } = {}
) {
  navigate('/feed', options);
}

/**
 * Navigate to user profile
 * @param userId The user ID to navigate to
 * @param options Options for navigation
 */
export function navigateToUserProfile(
  userId: string,
  options: {
    replace?: boolean,
    showToast?: boolean,
    toastMessage?: string,
    toastType?: 'info' | 'success' | 'warning' | 'error'
  } = {}
) {
  navigate(`/user/${userId}`, options);
}

/**
 * Navigate back in history
 * @param fallbackRoute The route to navigate to if there's no history
 */
export function goBack(fallbackRoute = '/') {
  if (window.history.length > 1) {
    window.history.back();
  } else {
    navigate(fallbackRoute, { replace: true });
  }
} 