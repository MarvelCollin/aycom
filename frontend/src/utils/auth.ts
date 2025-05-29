import { createLoggerWithPrefix } from './logger';
import appConfig from '../config/appConfig';

const logger = createLoggerWithPrefix('Auth');
const TOKEN_VALIDATION_INTERVAL = 1000 * 60 * 5;
let tokenValidationTimer: number | null = null;

// Add a memory cache for roles to avoid repeated API calls
const roleCache: Record<string, {role: string, timestamp: number}> = {};
const ROLE_CACHE_TTL = 5 * 60 * 1000; // 5 minutes in milliseconds

export function getAuthToken(): string {
  try {
    const authData = localStorage.getItem('auth');
    if (authData) {
      const auth = JSON.parse(authData);
      
      if (auth.accessToken) {
        if (auth.expiresAt && Date.now() > auth.expiresAt) {
          logger.warn('Token has expired, will need refresh');
          
          if (!auth.refreshToken) {
            logger.warn('No refresh token available for expired token');
            return '';
          }
        }
        
        logger.debug(`Found token: ${auth.accessToken.substring(0, 10)}...`);
        return auth.accessToken;
      } else {
        logger.warn('Token exists in auth data but is empty');
      }
    }
  } catch (err) {
    logger.error("Error parsing auth data:", err);
  }
  
  logger.warn('No auth token found');
  return '';
}

export function isAuthenticated(): boolean {
  try {
    const authData = localStorage.getItem('auth');
    if (authData) {
      const auth = JSON.parse(authData);
      
      if (auth.accessToken && auth.isAuthenticated === true) {
        if (auth.expiresAt) {
          const isValid = Date.now() < auth.expiresAt;
          logger.debug(`Token expiry check: now=${Date.now()}, expires=${auth.expiresAt}, isValid=${isValid}`);
          return isValid;
        }
        return true;
      } else {
        logger.warn(`Auth validation failed: token=${!!auth.accessToken}, isAuthenticated=${auth.isAuthenticated}`);
      }
    } else {
      logger.warn('No auth data in localStorage');
    }
  } catch (err) {
    logger.error("Error checking authentication status:", err);
  }
  
  if (localStorage.getItem('aycom_authenticated') === 'true') {
    logger.warn('Simple auth flag is true but auth data validation failed - clearing flag');
    localStorage.setItem('aycom_authenticated', 'false');
  }
  
  return false;
}

export function getUserId(): string | null {
  try {
    const authData = localStorage.getItem('auth');
    if (authData) {
      const auth = JSON.parse(authData);
      // Check both camelCase and snake_case versions for backward compatibility
      if (auth.userId || auth.user_id) {
        const id = auth.user_id || auth.userId;
        logger.debug(`Found user ID: ${id}`);
        return id;
      } else {
        logger.warn('Auth data exists but no user ID found');
      }
    }
  } catch (err) {
    logger.error("Error getting user ID:", err);
  }
  
  return null;
}

export function ensureTokenFreshness(): boolean {
  try {
    const authData = getAuthData();
    if (!authData || !authData.expiresAt) return false;
    
    const fiveMinutesMs = 5 * 60 * 1000;
    const timeUntilExpiry = authData.expiresAt - Date.now();
    const isNearExpiry = timeUntilExpiry < fiveMinutesMs;
    
    if (isNearExpiry) {
      logger.info(`Token is near expiry (expires in ${Math.round(timeUntilExpiry/1000)}s), should be refreshed`);
    }
    
    return isNearExpiry;
  } catch (error) {
    logger.error('Error checking token freshness:', error);
    return false;
  }
}

export function setAuthData(userData: {
  accessToken: string;
  refreshToken?: string;
  userId: string;
  expiresAt?: number;
  is_admin?: boolean;
} | null): void {
  try {
    if (userData === null) {
      clearAuthData();
      logger.info('Auth data cleared');
      return;
    }
    
    const expiresAt = userData.expiresAt || (Date.now() + 3600 * 1000);
    
    // Get existing auth data to preserve fields like is_admin if not provided
    let existingData = {};
    try {
      const existingAuth = localStorage.getItem('auth');
      if (existingAuth) {
        existingData = JSON.parse(existingAuth);
      }
    } catch (e) {
      logger.error('Error reading existing auth data:', e);
    }
    
    const authData = {
      ...existingData, // Preserve existing fields
      isAuthenticated: true,
      userId: userData.userId, // Keep camelCase for backward compatibility
      user_id: userData.userId, // Add snake_case version for new code
      accessToken: userData.accessToken,
      refreshToken: userData.refreshToken || null,
      expiresAt: expiresAt,
      // Use provided is_admin or preserve existing value
      is_admin: userData.is_admin !== undefined ? userData.is_admin : (existingData as any).is_admin || false
    };
    
    localStorage.setItem('auth', JSON.stringify(authData));
    localStorage.setItem('aycom_authenticated', 'true');
    
    logger.info('Auth data updated successfully');
    logger.debug(`User ID: ${userData.userId}, Token: ${userData.accessToken.substring(0, 10)}..., Expires: ${new Date(expiresAt).toLocaleString()}, Admin: ${authData.is_admin}`);
    
    setupTokenValidation();
  } catch (err) {
    logger.error("Error setting auth data:", err);
  }
}

export function clearAuthData(): void {
  try {
    localStorage.removeItem('auth');
    localStorage.setItem('aycom_authenticated', 'false');
    logger.info('Auth data cleared');
    
    if (tokenValidationTimer !== null) {
      window.clearInterval(tokenValidationTimer);
      tokenValidationTimer = null;
      logger.debug('Token validation interval cleared');
    }
  } catch (err) {
    logger.error("Error clearing auth data:", err);
  }
}

export function getAuthData() {
  try {
    const authData = localStorage.getItem('auth');
    if (authData) {
      return JSON.parse(authData);
    }
  } catch (err) {
    logger.error("Error getting auth data:", err);
  }
  
  return null;
}

function setupTokenValidation() {
  if (tokenValidationTimer !== null) {
    window.clearInterval(tokenValidationTimer);
  }
  
  tokenValidationTimer = window.setInterval(() => {
    logger.debug('Running periodic token validation');
    const isValid = isAuthenticated();
    if (!isValid) {
      logger.warn('Token validation failed during periodic check');
    }
  }, TOKEN_VALIDATION_INTERVAL);
  
  logger.debug('Token validation interval setup');
}

export function updateTokenExpiry(newExpiresAt: number): void {
  try {
    const authData = getAuthData();
    if (authData) {
      authData.expiresAt = newExpiresAt;
      localStorage.setItem('auth', JSON.stringify(authData));
      logger.debug(`Token expiry updated to ${new Date(newExpiresAt).toLocaleString()}`);
    }
  } catch (err) {
    logger.error("Error updating token expiry:", err);
  }
}

/**
 * Gets the user's role from auth data or fetches it from the API
 * @returns User role: 'admin', 'moderator', or 'user' (default)
 */
export async function getUserRole(): Promise<string> {
  try {
    // First check if we have the role cached in auth data
    const authData = getAuthData();
    const userId = getUserId();
    
    // Return quickly if no user ID - default to regular user
    if (!userId) {
      logger.warn('Cannot get user role - not logged in');
      return 'user';
    }
    
    // Check memory cache first (faster than localStorage)
    const now = Date.now();
    const cachedEntry = roleCache[userId];
    if (cachedEntry && (now - cachedEntry.timestamp) < ROLE_CACHE_TTL) {
      logger.debug(`Using memory-cached user role: ${cachedEntry.role}`);
      return cachedEntry.role;
    }
    
    // Then check localStorage cache
    if (authData && authData.userRole) {
      // Update memory cache
      roleCache[userId] = {
        role: authData.userRole,
        timestamp: now
      };
      logger.debug(`Using stored user role: ${authData.userRole}`);
      return authData.userRole;
    }
    
    // If not cached anywhere, fetch from the API
    const token = getAuthToken();
    if (!token) {
      logger.warn('Cannot get user role - no auth token');
      return 'user';
    }
    
    const API_BASE_URL = appConfig.api.baseUrl;
    logger.debug('Fetching user role from API');
    
    try {
      const response = await fetch(`${API_BASE_URL}/users/${userId}/role`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        credentials: 'include'
      });
      
      if (response.status === 404) {
        const defaultRole = 'user';
        logger.warn('User role endpoint not found (404) - using default role');
        
        // Cache the default role to avoid repeated API calls
        if (authData) {
          authData.userRole = defaultRole;
          localStorage.setItem('auth', JSON.stringify(authData));
        }
        roleCache[userId] = { role: defaultRole, timestamp: now };
        
        return defaultRole;
      }
      
      if (!response.ok) {
        const defaultRole = 'user';
        logger.warn(`Failed to get user role: ${response.status} ${response.statusText}`);
        
        // Cache the default role briefly to avoid repeated API calls
        roleCache[userId] = { role: defaultRole, timestamp: now };
        
        return defaultRole;
      }
      
      const data = await response.json();
      const role = data.role || 'user';
      
      // Cache the role in auth data
      if (authData) {
        authData.userRole = role;
        localStorage.setItem('auth', JSON.stringify(authData));
      }
      
      // Update memory cache
      roleCache[userId] = { role, timestamp: now };
      logger.debug(`Cached user role: ${role}`);
      
      return role;
    } catch (fetchError) {
      logger.warn('Failed to fetch user role - network error:', fetchError);
      return 'user';
    }
  } catch (error) {
    logger.error('Error getting user role:', error);
    return 'user';
  }
}