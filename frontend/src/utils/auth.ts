import { createLoggerWithPrefix } from './logger';
import appConfig from '../config/appConfig';

const logger = createLoggerWithPrefix('Auth');
const TOKEN_VALIDATION_INTERVAL = 1000 * 60 * 5;
let tokenValidationTimer: number | null = null;

const roleCache: Record<string, {role: string, timestamp: number}> = {};
const ROLE_CACHE_TTL = 5 * 60 * 1000; 

export function getAuthToken(): string {
  try {
    const authData = localStorage.getItem('auth');
    if (authData) {
      const auth = JSON.parse(authData);
      if (auth.access_token) {
        return auth.access_token;
      }
    }
  } catch (err) {
    logger.error("Error parsing auth data:", err);
  }
  return '';
}

export function isAuthenticated(): boolean {
  try {
    const authData = localStorage.getItem('auth');
    if (authData) {
      const auth = JSON.parse(authData);
      if (auth.is_authenticated && auth.access_token) {
        if (auth.expires_at) {
          return Date.now() < auth.expires_at;
        }
        return true;
      }
    }
  } catch (err) {
    logger.error("Error checking authentication status:", err);
  }
  return false;
}

export function getUserId(): string | null {
  try {
    const authData = localStorage.getItem('auth');
    if (authData) {
      const auth = JSON.parse(authData);
      if (auth.user_id) {
        return auth.user_id;
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
    if (!authData || !authData.expires_at) return false;

    const refreshBuffer = appConfig.auth.tokenRefreshBuffer || 5 * 60 * 1000; // Default to 5 minutes
    const timeUntilExpiry = authData.expires_at - Date.now();
    const isNearExpiry = timeUntilExpiry < refreshBuffer;

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
      return;
    }

    const expiresAt = userData.expiresAt || (Date.now() + 3600 * 1000);
    
    const authData = {
      is_authenticated: true,
      user_id: userData.userId,
      access_token: userData.accessToken,
      refresh_token: userData.refreshToken || null,
      expires_at: expiresAt,
      is_admin: userData.is_admin || false
    };

    localStorage.setItem('auth', JSON.stringify(authData));
    
    setupTokenValidation();
  } catch (err) {
    logger.error("Error setting auth data:", err);
  }
}

export function clearAuthData(): void {
  try {
    localStorage.removeItem('auth');
    
    if (tokenValidationTimer !== null) {
      window.clearInterval(tokenValidationTimer);
      tokenValidationTimer = null;
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
      authData.expires_at = newExpiresAt;
      authData.expiresAt = newExpiresAt; // For backward compatibility
      localStorage.setItem('auth', JSON.stringify(authData));
      logger.debug(`Token expiry updated to ${new Date(newExpiresAt).toLocaleString()}`);
    }
  } catch (err) {
    logger.error("Error updating token expiry:", err);
  }
}

export async function getUserRole(): Promise<string> {
  try {

    const authData = getAuthData();
    const userId = getUserId();

    if (!userId) {
      logger.warn('Cannot get user role - not logged in');
      return 'user';
    }

    const now = Date.now();
    const cachedEntry = roleCache[userId];
    if (cachedEntry && (now - cachedEntry.timestamp) < ROLE_CACHE_TTL) {
      logger.debug(`Using memory-cached user role: ${cachedEntry.role}`);
      return cachedEntry.role;
    }

    if (authData && authData.userRole) {

      roleCache[userId] = {
        role: authData.userRole,
        timestamp: now
      };
      logger.debug(`Using stored user role: ${authData.userRole}`);
      return authData.userRole;
    }

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

        roleCache[userId] = { role: defaultRole, timestamp: now };

        return defaultRole;
      }

      const data = await response.json();
      const role = data.role || 'user';

      if (authData) {
        authData.userRole = role;
        localStorage.setItem('auth', JSON.stringify(authData));
      }

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