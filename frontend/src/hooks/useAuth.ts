import { writable, get } from 'svelte/store';
import type { IUserRegistration, IGoogleCredentialResponse, ITokenResponse, IAuthStore } from '../interfaces/IAuth';
import { setAuthData, clearAuthData, getAuthData, ensureTokenFreshness } from '../utils/auth';
import { createLoggerWithPrefix } from '../utils/logger';

// Use a consistent API URL with port 8081 (matches API Gateway in docker-compose.yml)
const API_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081/api/v1';
const TOKEN_EXPIRY_BUFFER = 300000; // 5 minutes in milliseconds
const logger = createLoggerWithPrefix('AuthHook');

interface AuthState extends IAuthStore {
  expiresAt: number | null;
}

const createAuthStore = () => {
  const initialState: AuthState = {
    isAuthenticated: false,
    userId: null,
    accessToken: null,
    refreshToken: null,
    expiresAt: null
  };

  const auth = writable<AuthState>(initialState);
  
  const initAuth = () => {
    try {
      const storedAuth = localStorage.getItem('auth');
      if (storedAuth) {
        const parsedAuth = JSON.parse(storedAuth) as AuthState;
        
        logger.debug('Found stored auth data', { userId: parsedAuth.userId });
        
        if (parsedAuth.expiresAt && parsedAuth.expiresAt > Date.now()) {
          logger.info('Initializing with valid token');
          auth.set(parsedAuth);
          
          // Even if token is valid, if it's close to expiring, refresh it
          if (parsedAuth.expiresAt - Date.now() < TOKEN_EXPIRY_BUFFER) {
            logger.info('Token valid but close to expiry, refreshing');
            if (parsedAuth.refreshToken) {
              refreshExpiredToken(parsedAuth.refreshToken)
                .catch(err => logger.error('Failed to refresh near-expiry token', err));
            }
          }
        } else if (parsedAuth.refreshToken) {
          // Token expired but we have refresh token
          logger.info('Token expired, attempting refresh');
          refreshExpiredToken(parsedAuth.refreshToken)
            .catch(() => {
              logger.error('Failed to refresh expired token, logging out');
              clearAuth();
            });
        } else {
          // No valid tokens, clear auth state
          logger.warn('No valid tokens available, clearing auth state');
          clearAuth();
        }
      } else {
        logger.info('No stored auth data found');
      }
    } catch (error) {
      logger.error('Failed to initialize auth from localStorage:', error);
      clearAuth();
    }
  };
  
  const persistAuth = (authState: AuthState) => {
    try {
      // Use the auth utility to set data
      if (authState.accessToken && authState.userId) {
        logger.debug('Persisting valid auth state');
        setAuthData({
          accessToken: authState.accessToken,
          refreshToken: authState.refreshToken || undefined,
          userId: authState.userId,
          expiresAt: authState.expiresAt || undefined
        });
      } else {
        logger.warn('Persisting potentially invalid auth state');
        localStorage.setItem('auth', JSON.stringify(authState));
      }
    } catch (error) {
      logger.error('Failed to persist auth to localStorage:', error);
    }
  };

  const clearAuth = () => {
    // Use the auth utility to clear data
    logger.info('Clearing authentication state');
    clearAuthData();
    auth.set(initialState);
  };

  const refreshExpiredToken = async (refreshToken: string): Promise<boolean> => {
    try {
      logger.info('Attempting to refresh token');
      const response = await fetch(`${API_URL}/auth/refresh-token`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ refresh_token: refreshToken })
      });
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        logger.error(`Token refresh failed: ${response.status}`, errorData);
        throw new Error(`HTTP error! Status: ${response.status}`);
      }
      
      const data = await response.json();
      
      if (data.success && data.access_token) {
        const expiresAt = Date.now() + (data.expires_in * 1000);
        // Get current state to preserve username and displayName
        const currentState = get(authStore);
        
        const newState: AuthState = {
          isAuthenticated: true,
          userId: data.user_id,
          accessToken: data.access_token,
          refreshToken: data.refresh_token,
          username: currentState.username || null,
          displayName: currentState.displayName || null,
          expiresAt
        };
        logger.info('Token refreshed successfully', { 
          userId: data.user_id, 
          expiresAt: new Date(expiresAt).toLocaleString() 
        });
        auth.set(newState);
        persistAuth(newState);
        return true;
      } else {
        logger.warn('Refresh response did not contain a valid token', { data });
        clearAuth();
        return false;
      }
    } catch (error) {
      logger.error('Failed to refresh token:', error);
      clearAuth();
      return false;
    }
  };
  
  return {
    subscribe: auth.subscribe,
    set: (value: AuthState) => {
      auth.set(value);
      persistAuth(value);
    },
    update: (updater: (value: AuthState) => AuthState) => {
      auth.update((value) => {
        const updated = updater(value);
        persistAuth(updated);
        return updated;
      });
    },
    init: initAuth,
    logout: () => clearAuth(),
    refreshToken: refreshExpiredToken
  };
};

const authStore = createAuthStore();
authStore.init();

export function useAuth() {
  // Common fetch with timeout and error handling
  const fetchWithTimeout = async (url: string, options: RequestInit, timeout = 10000): Promise<Response> => {
    const controller = new AbortController();
    const id = setTimeout(() => controller.abort(), timeout);
    
    try {
      // Log the request URL (helpful for debugging)
      logger.debug(`Fetching: ${url}`);
      
      const response = await fetch(url, {
        ...options,
        signal: controller.signal
      });
      
      clearTimeout(id);
      return response;
    } catch (error) {
      // Make sure to clear the timeout if fetch fails
      clearTimeout(id);
      throw error;
    }
  };

  const handleApiError = (error: unknown): { success: false, message: string } => {
    if (error instanceof Error) {
      if (error.name === 'AbortError') {
        return { success: false, message: 'Request timed out. Please try again.' };
      }
      return { success: false, message: error.message };
    }
    return { success: false, message: 'An unexpected error occurred.' };
  };

  // Function to check for and handle unauthorized responses
  const handleUnauthorizedResponse = (response: Response): boolean => {
    if (response.status === 401) {
      logger.warn('Received 401 Unauthorized response');
      // Try to refresh the token first
      const authData = getAuthData();
      if (authData?.refreshToken) {
        logger.info('Attempting to refresh token after 401');
        return true; // Signal caller to try refreshing
      } else {
        // No refresh token available, must log out
        clearAuthData();
        window.location.href = '/login?session_expired=true';
        return false;
      }
    }
    return false;
  };

  // Function to manually try refreshing the token
  const tryTokenRefresh = async (): Promise<boolean> => {
    const authData = getAuthData();
    if (authData?.refreshToken) {
      logger.info('Manually attempting token refresh');
      return await authStore.refreshToken(authData.refreshToken);
    }
    return false;
  };

  const register = async (userData: IUserRegistration) => {
    try {
      const response = await fetchWithTimeout(
        `${API_URL}/users/register`, 
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(userData)
        }
      );
      const data = await response.json();
      return {
        success: data.success,
        message: data.message || 'Registration successful! Check your email for verification code.'
      };
    } catch (error) {
      logger.error('Registration error:', error);
      return handleApiError(error);
    }
  };
  
  const registerWithMedia = async (formData: FormData) => {
    try {
      const response = await fetchWithTimeout(
        `${API_URL}/auth/register-with-media`,
        {
          method: 'POST',
          body: formData
        }
      );
      
      const data = await response.json();
      
      return {
        success: data.success,
        message: data.message || 'Registration successful! Check your email for verification code.'
      };
    } catch (error) {
      logger.error('Registration with media error:', error);
      return handleApiError(error);
    }
  };
  
  const verifyEmail = async (email: string, code: string) => {
    try {
      const response = await fetchWithTimeout(
        `${API_URL}/auth/verify-email`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email, verification_code: code })
        }
      );
      
      const data = await response.json();
      
      if (data.success && data.access_token) {
        const expiresAt = Date.now() + (data.expires_in * 1000);
        authStore.set({
          isAuthenticated: true,
          userId: data.user_id,
          accessToken: data.access_token,
          refreshToken: data.refresh_token,
          expiresAt
        });
      }
      
      return {
        success: data.success,
        message: data.message || 'Email verification successful!'
      };
    } catch (error) {
      logger.error('Email verification error:', error);
      return handleApiError(error);
    }
  };
  
  const resendVerificationCode = async (email: string) => {
    try {
      const response = await fetchWithTimeout(
        `${API_URL}/auth/resend-code`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email })
        }
      );
      
      const data = await response.json();
      
      return {
        success: data.success,
        message: data.message || 'Verification code has been resent.'
      };
    } catch (error) {
      logger.error('Resend verification code error:', error);
      return handleApiError(error);
    }
  };
  
  const getCurrentUser = async () => {
    try {
      const state = get(authStore);
      if (!state.accessToken) {
        logger.warn('getCurrentUser called without access token');
        return { success: false, message: 'Not authenticated' };
      }
      
      logger.debug('Fetching current user profile');
      logger.debug(`Using token: ${state.accessToken.substring(0, 15)}...`);
      
      const response = await fetchWithTimeout(
        `${API_URL}/users/me`,
        {
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${state.accessToken}`,
            'Content-Type': 'application/json'
          }
        },
        15000 // Extend timeout for profile fetch
      );
      
      logger.debug(`User profile response status: ${response.status}`);
      
      if (handleUnauthorizedResponse(response)) {
        // Try refreshing token
        const refreshed = await tryTokenRefresh();
        logger.debug(`Token refresh result: ${refreshed ? 'Success' : 'Failed'}`);
        
        if (refreshed) {
          // Retry with new token
          const newState = get(authStore);
          logger.debug(`Retrying with new token: ${newState.accessToken ? newState.accessToken.substring(0, 15) : 'undefined'}...`);
          
          const retryResponse = await fetchWithTimeout(
            `${API_URL}/users/me`,
            {
              method: 'GET',
              headers: {
                'Authorization': `Bearer ${newState.accessToken}`,
                'Content-Type': 'application/json'
              }
            }
          );
          
          logger.debug(`Retry response status: ${retryResponse.status}`);
          
          if (!retryResponse.ok) {
            const errorText = await retryResponse.text().catch(() => '');
            logger.error(`Failed to get user profile after token refresh: ${retryResponse.status}`, { errorText });
            throw new Error(`Failed to get user profile after token refresh: ${retryResponse.status}`);
          }
          
          const data = await retryResponse.json();
          logger.debug('Successfully retrieved user profile after token refresh');
          return data;
        } else {
          logger.error('Failed to refresh token during getCurrentUser');
          throw new Error('Session expired. Please log in again.');
        }
      }
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        logger.error(`Failed to get user profile: ${response.status}`, errorData);
        throw new Error(errorData.message || `Failed to get user profile: ${response.status}`);
      }
      
      const data = await response.json();
      logger.debug('Successfully retrieved user profile');
      return data;
    } catch (error) {
      logger.error('Get current user error:', error);
      return handleApiError(error);
    }
  };

  const login = async (email: string, password: string) => {
    try {
      logger.info(`Attempting login for ${email}`);
      
      // Log request details before sending
      logger.debug(`Login request to: ${API_URL}/users/login with email: ${email}`);
      
      const response = await fetchWithTimeout(
        `${API_URL}/users/login`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email, password })
        }
      );
      
      // Log response status
      logger.debug(`Login response status: ${response.status} ${response.statusText}`);
      
      const data = await response.json();
      logger.debug(`Login response data: ${JSON.stringify(data)}`);
      
      if (data.success) {
        // Check for both token naming patterns since backends can vary
        const accessToken = data.access_token || data.token;
        const refreshToken = data.refresh_token || data.refreshToken;
        const expiresIn = data.expires_in || 3600; // Default to 1 hour if not provided
        const userId = data.user_id || (data.user ? data.user.id : null);
        const username = data.user ? (data.user.username || '') : '';
        const displayName = data.user ? (data.user.name || data.user.displayName || '') : '';
        
        logger.debug('Token data extracted', { 
          hasAccessToken: !!accessToken,
          hasRefreshToken: !!refreshToken,
          hasUserId: !!userId,
          username,
          displayName,
          tokenPrefix: accessToken ? accessToken.substring(0, 10) : 'NONE'
        });
        
        if (accessToken) {
          // Set auth data with proper tokens
          const authData = {
            isAuthenticated: true,
            userId,
            username,
            displayName,
            accessToken,
            refreshToken, 
            expiresAt: Date.now() + (expiresIn * 1000)
          };
          logger.info('Setting auth store with new login data');
          authStore.set(authData);
          
          // Verify what was actually stored
          const currentState = get(authStore);
          logger.debug('Auth state set successfully', { 
            isAuthenticated: currentState.isAuthenticated,
            hasUserId: !!currentState.userId,
            username: currentState.username,
            displayName: currentState.displayName,
            tokenPrefix: currentState.accessToken ? currentState.accessToken.substring(0, 10) : 'NONE'
          });
          
          // Verify with backend immediately after login
          setTimeout(() => {
            logger.debug('Starting post-login validation');
            validateAuth()
              .then(valid => {
                logger.debug(`Post-login validation result: ${valid ? 'Valid' : 'Invalid'}`);
              })
              .catch(err => {
                logger.error('Post-login validation failed:', err);
              });
          }, 500);
        } else {
          logger.error('Login response missing access token');
        }
      } else {
        logger.warn(`Login failed: ${data.message || 'Unknown error'}`);
      }
      
      return {
        success: data.success,
        message: data.message || (data.success ? 'Login successful!' : 'Login failed'),
        token: data.access_token || data.token,
        user: data.user
      };
    } catch (error) {
      logger.error('Login error:', error);
      return handleApiError(error);
    }
  };
  
  const handleGoogleAuth = async (response: IGoogleCredentialResponse) => {
    try {
      const apiResponse = await fetchWithTimeout(
        `${API_URL}/auth/google`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ token_id: response.credential })
        }
      );
      
      const data = await apiResponse.json();
      
      if (data.success && data.access_token) {
        const expiresAt = Date.now() + (data.expires_in * 1000);
        authStore.set({
          isAuthenticated: true,
          userId: data.user_id,
          accessToken: data.access_token,
          refreshToken: data.refresh_token,
          expiresAt
        });
        
        // Verify with backend immediately after login
        setTimeout(() => {
          validateAuth().catch(err => {
            logger.error('Post-Google-login validation failed:', err);
          });
        }, 500);
      }
      
      return {
        success: data.success,
        message: data.message || 'Google authentication successful!'
      };
    } catch (error) {
      logger.error('Google auth error:', error);
      return handleApiError(error);
    }
  };
  
  const logout = async () => {
    const state = get(authStore);
    if (state.accessToken) {
      try {
        logger.info('Sending logout request to server');
        await fetchWithTimeout(
          `${API_URL}/auth/logout`,
          {
            method: 'POST',
            headers: {
              'Authorization': `Bearer ${state.accessToken}`,
              'Content-Type': 'application/json'
            },
            body: JSON.stringify({ 
              access_token: state.accessToken,
              refresh_token: state.refreshToken
            })
          }
        );
        logger.info('Logout request completed');
      } catch (error) {
        logger.error('Logout request error:', error);
        // Continue with local logout even if server logout fails
      }
    }
    
    logger.info('Clearing local auth state');
    authStore.logout();
  };
  
  const getAuthState = () => {
    const state = get(authStore);
    
    // Check if token is about to expire and needs refresh
    if (state.isAuthenticated && state.expiresAt && state.refreshToken) {
      const now = Date.now();
      const timeUntilExpiry = state.expiresAt - now;
      
      if (timeUntilExpiry < 0) {
        // Token has already expired, trigger refresh immediately
        logger.warn('Token has expired, refreshing');
        authStore.refreshToken(state.refreshToken)
          .catch(() => {
            logger.error('Failed to refresh expired token on getAuthState');
            authStore.logout();
          });
      } else if (timeUntilExpiry < TOKEN_EXPIRY_BUFFER) {
        // Token will expire soon, refresh it
        logger.info(`Token expires in ${Math.round(timeUntilExpiry/1000)}s, refreshing`);
        authStore.refreshToken(state.refreshToken)
          .catch(err => logger.error('Failed to refresh token on getAuthState', err));
      }
    }
    
    return state;
  };
  
  // Validate current auth state with backend
  const validateAuth = async (): Promise<boolean> => {
    try {
      const state = get(authStore);
      if (!state.isAuthenticated || !state.accessToken) {
        logger.info('No auth state to validate');
        return false;
      }
      
      logger.debug('Validating current authentication with backend', {
        userId: state.userId, 
        tokenPrefix: state.accessToken.substring(0, 10),
        tokenExpiry: state.expiresAt ? new Date(state.expiresAt).toLocaleString() : 'undefined'
      });
      
      // Make a direct call to check the token rather than using the getCurrentUser wrapper
      const response = await fetch(`${API_URL}/users/me`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${state.accessToken}`,
          'Content-Type': 'application/json'
        }
      });
      
      logger.debug(`Auth validation response status: ${response.status}`);
      
      if (response.status === 401) {
        logger.warn('Auth validation failed with 401 Unauthorized');
        
        // Try to get the error response for debugging
        try {
          const errorData = await response.clone().json();
          logger.warn('Auth validation error details:', errorData);
        } catch (e) {
          // Ignore parsing errors
        }
        
        // Try refreshing token
        if (state.refreshToken) {
          logger.debug('Attempting token refresh during validation');
          const refreshSuccess = await authStore.refreshToken(state.refreshToken);
          logger.debug(`Token refresh result: ${refreshSuccess ? 'Success' : 'Failed'}`);
          
          if (refreshSuccess) {
            // Verify the refreshed token works
            const newState = get(authStore);
            if (newState.accessToken) {
              logger.debug('Verifying refreshed token works');
              const verifyResponse = await fetch(`${API_URL}/users/me`, {
                method: 'GET',
                headers: {
                  'Authorization': `Bearer ${newState.accessToken}`,
                  'Content-Type': 'application/json'
                }
              });
              
              logger.debug(`Verification response status: ${verifyResponse.status}`);
              
              if (verifyResponse.ok) {
                logger.info('Token refresh and verification successful');
                return true;
              } else {
                logger.error(`Token refresh succeeded but verification failed: ${verifyResponse.status}`);
              }
            }
          }
          return refreshSuccess;
        }
        return false;
      }
      
      if (!response.ok) {
        const errorText = await response.text().catch(() => '');
        logger.error(`Auth validation failed with status ${response.status}`, { errorText });
        return false;
      }
      
      const data = await response.json();
      logger.debug('Auth validation successful', { user: data.user ? data.user.id : 'unknown' });
      return true;
    } catch (error) {
      logger.error('Auth validation error:', error);
      return false;
    }
  };
  
  // Immediately validate auth with server on initialization
  setTimeout(() => {
    const state = get(authStore);
    if (state.isAuthenticated) {
      logger.info('Validating authentication state with server on initialization');
      validateAuth().catch(err => {
        logger.error('Initial auth validation failed:', err);
      });
    }
  }, 1000);
  
  return {
    subscribe: authStore.subscribe,
    register,
    registerWithMedia,
    verifyEmail,
    resendVerificationCode,
    login,
    handleGoogleAuth,
    logout,
    getAuthState,
    getCurrentUser,
    tryTokenRefresh,
    validateAuth
  };
}