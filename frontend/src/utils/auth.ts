// Auth utility for consistent token management

/**
 * Helper to get the authentication token from localStorage
 */
export function getAuthToken(): string {
  try {
    const authData = localStorage.getItem('auth');
    if (authData) {
      const auth = JSON.parse(authData);
      return auth.accessToken || '';
    }
  } catch (err) {
    console.error("Error parsing auth data:", err);
  }
  
  return '';
}

/**
 * Helper to check if user is authenticated
 */
export function isAuthenticated(): boolean {
  try {
    const authData = localStorage.getItem('auth');
    if (authData) {
      const auth = JSON.parse(authData);
      return !!auth.accessToken && auth.isAuthenticated === true;
    }
  } catch (err) {
    console.error("Error checking authentication status:", err);
  }
  
  return false;
}

/**
 * Get current user ID
 */
export function getUserId(): string | null {
  try {
    const authData = localStorage.getItem('auth');
    if (authData) {
      const auth = JSON.parse(authData);
      return auth.userId || null;
    }
  } catch (err) {
    console.error("Error getting user ID:", err);
  }
  
  return null;
}

/**
 * Set auth data to localStorage and mark user as authenticated
 */
export function setAuthData(userData: {
  accessToken: string;
  refreshToken?: string;
  userId: string;
  expiresAt?: number;
}): void {
  try {
    const authData = {
      isAuthenticated: true,
      userId: userData.userId,
      accessToken: userData.accessToken,
      refreshToken: userData.refreshToken || null,
      expiresAt: userData.expiresAt || null
    };
    
    localStorage.setItem('auth', JSON.stringify(authData));
    localStorage.setItem('aycom_authenticated', 'true');
  } catch (err) {
    console.error("Error setting auth data:", err);
  }
}

/**
 * Clear auth data and mark user as logged out
 */
export function clearAuthData(): void {
  try {
    localStorage.removeItem('auth');
    localStorage.setItem('aycom_authenticated', 'false');
  } catch (err) {
    console.error("Error clearing auth data:", err);
  }
} 