import { writable, get } from 'svelte/store';
import type { IUserRegistration, IGoogleCredentialResponse, ITokenResponse, IAuthStore } from '../interfaces/IAuth';
import { setAuthData, clearAuthData } from '../utils/auth';

// Use a consistent API URL with port 8081 (matches API Gateway in docker-compose.yml)
const API_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081/api/v1';
const TOKEN_EXPIRY_BUFFER = 300000; // 5 minutes in milliseconds

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
        
        if (parsedAuth.expiresAt && parsedAuth.expiresAt > Date.now()) {
          auth.set(parsedAuth);
        } else if (parsedAuth.refreshToken) {
          // Token expired but we have refresh token
          refreshExpiredToken(parsedAuth.refreshToken);
        } else {
          // No valid tokens, clear auth state
          clearAuth();
        }
      }
    } catch (error) {
      console.error('Failed to initialize auth from localStorage:', error);
      clearAuth();
    }
  };
  
  const persistAuth = (authState: AuthState) => {
    try {
      // Use the auth utility to set data
      if (authState.accessToken && authState.userId) {
        setAuthData({
          accessToken: authState.accessToken,
          refreshToken: authState.refreshToken || undefined,
          userId: authState.userId,
          expiresAt: authState.expiresAt || undefined
        });
      } else {
        localStorage.setItem('auth', JSON.stringify(authState));
      }
    } catch (error) {
      console.error('Failed to persist auth to localStorage:', error);
    }
  };

  const clearAuth = () => {
    // Use the auth utility to clear data
    clearAuthData();
    auth.set(initialState);
  };

  const refreshExpiredToken = async (refreshToken: string) => {
    try {
      const response = await fetch(`${API_URL}/auth/refresh-token`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ refresh_token: refreshToken })
      });
      
      if (!response.ok) {
        throw new Error(`HTTP error! Status: ${response.status}`);
      }
      
      const data = await response.json();
      
      if (data.success && data.access_token) {
        const expiresAt = Date.now() + (data.expires_in * 1000);
        const newState: AuthState = {
          isAuthenticated: true,
          userId: data.user_id,
          accessToken: data.access_token,
          refreshToken: data.refresh_token,
          expiresAt
        };
        auth.set(newState);
        persistAuth(newState);
      } else {
        clearAuth();
      }
    } catch (error) {
      console.error('Failed to refresh token:', error);
      clearAuth();
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
      console.log(`Fetching: ${url}`);
      
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
      console.error('Registration error:', error);
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
      console.error('Registration with media error:', error);
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
      console.error('Email verification error:', error);
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
      console.error('Resend verification code error:', error);
      return handleApiError(error);
    }
  };
  
  const getCurrentUser = async () => {
    
  }

  const login = async (email: string, password: string) => {
    try {
      console.log(`Attempting login for ${email} at ${API_URL}/users/login`);
      const response = await fetchWithTimeout(
        `${API_URL}/users/login`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email, password })
        }
      );
      const data = await response.json();
      console.log('Login response (full):', JSON.stringify(data, null, 2));
      
      if (data.success) {
        // Check for both token naming patterns since backends can vary
        const accessToken = data.access_token || data.token;
        const refreshToken = data.refresh_token || data.refreshToken;
        const expiresIn = data.expires_in || 3600; // Default to 1 hour if not provided
        const userId = data.user_id || (data.user ? data.user.id : null);
        
        console.log('Token data extracted:', { 
          accessToken: accessToken || 'MISSING',
          refreshToken: refreshToken || 'MISSING',
          userId: userId || 'MISSING',
          expiresIn
        });
        
        if (accessToken) {
          // Set auth data with proper tokens
          const authData = {
            isAuthenticated: true,
            userId: userId,
            accessToken: accessToken,
            refreshToken: refreshToken, 
            expiresAt: Date.now() + (expiresIn * 1000)
          };
          console.log('Setting auth store with:', JSON.stringify(authData, null, 2));
          authStore.set(authData);
          
          // Verify what was actually stored
          const currentState = get(authStore);
          console.log('Current auth state after setting:', JSON.stringify(currentState, null, 2));
        } else {
          console.error('Login response missing access token:', data);
        }
      }
      
      return {
        success: data.success,
        message: data.message || (data.success ? 'Login successful!' : 'Login failed'),
        token: data.access_token || data.token,
        user: data.user
      };
    } catch (error) {
      console.error('Login error:', error);
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
      }
      
      return {
        success: data.success,
        message: data.message || 'Google authentication successful!'
      };
    } catch (error) {
      console.error('Google auth error:', error);
      return handleApiError(error);
    }
  };
  
  const logout = async () => {
    const state = get(authStore);
    if (state.accessToken) {
      try {
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
      } catch (error) {
        console.error('Logout error:', error);
      }
    }
    
    authStore.logout();
  };
  
  const getAuthState = () => {
    const state = get(authStore);
    
    // Check if token is about to expire and needs refresh
    if (state.expiresAt && state.refreshToken) {
      const now = Date.now();
      if (state.expiresAt - now < TOKEN_EXPIRY_BUFFER) {
        // Token will expire soon, refresh it
        authStore.refreshToken(state.refreshToken);
      }
    }
    
    return state;
  };
  
  return {
    subscribe: authStore.subscribe,
    register,
    registerWithMedia,
    verifyEmail,
    resendVerificationCode,
    login,
    handleGoogleAuth,
    logout,
    getAuthState
  };
}