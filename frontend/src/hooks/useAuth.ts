import { writable, get } from 'svelte/store';
import type { IUserRegistration, IGoogleCredentialResponse, ITokenResponse, IAuthStore } from '../interfaces/IAuth';
import { setAuthData, clearAuthData, getAuthToken } from '../utils/auth';
import * as authApi from '../api/auth';
import appConfig from '../config/appConfig';

// Use the API base URL from appConfig
const API_URL = appConfig.api.baseUrl;
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
      const data = await authApi.refreshToken(refreshToken);
      
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
      const data = await authApi.register(userData);
      return {
        success: data.success,
        message: data.message || 'Registration successful! Check your email for verification code.'
      };
    } catch (error) {
      console.error('Registration error:', error);
      return handleApiError(error);
    }
  };
  
  const verifyEmail = async (email: string, code: string) => {
    try {
      const data = await authApi.verifyEmail(email, code);
      
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
      const data = await authApi.resendVerification(email);
      return {
        success: data.success,
        message: data.message || 'Verification code sent!'
      };
    } catch (error) {
      console.error('Resend verification error:', error);
      return handleApiError(error);
    }
  };
  
  const getCurrentUser = async () => {
    // Use the appropriate API call from user.ts when available
    return null;
  };
  
  const login = async (email: string, password: string) => {
    try {
      const data = await authApi.login(email, password);
      
      if (data.success && data.access_token) {
        // Calculate token expiry if not provided
        const expiresAt = data.expires_in 
          ? Date.now() + (data.expires_in * 1000) 
          : Date.now() + (3600 * 1000); // Default 1 hour
        
        // Create auth state
        const authState: AuthState = {
          isAuthenticated: true,
          userId: data.user_id,
          accessToken: data.access_token,
          refreshToken: data.refresh_token,
          expiresAt: expiresAt
        };
        
        // Update store
        authStore.set(authState);
        
        return {
          success: true,
          message: data.message || 'Login successful!'
        };
      } else {
        return {
          success: false,
          message: data.message || 'Login failed. Please check your credentials.'
        };
      }
    } catch (error) {
      console.error('Login error:', error);
      return handleApiError(error);
    }
  };
  
  const handleGoogleAuth = async (response: IGoogleCredentialResponse) => {
    try {
      if (!response.credential) {
        throw new Error('No Google credential provided');
      }
      
      const data = await authApi.googleLogin(response.credential);
      
      if (data.success && data.access_token) {
        // Calculate token expiry
        const expiresAt = data.expires_in 
          ? Date.now() + (data.expires_in * 1000) 
          : Date.now() + (3600 * 1000); // Default 1 hour
        
        // Create auth state
        const authState: AuthState = {
          isAuthenticated: true,
          userId: data.user_id,
          accessToken: data.access_token,
          refreshToken: data.refresh_token,
          expiresAt: expiresAt
        };
        
        // Update store
        authStore.set(authState);
        
        return {
          success: true,
          message: data.message || 'Google login successful!'
        };
      } else {
        return {
          success: false,
          message: data.message || 'Google login failed.'
        };
      }
    } catch (error) {
      console.error('Google auth error:', error);
      return handleApiError(error);
    }
  };
  
  const logout = async () => {
    // No need for server logout call yet, just clear local storage
    authStore.logout();
    return {
      success: true,
      message: 'Logged out successfully'
    };
  };
  
  // Access the current auth state
  const getAuthState = () => {
    // return get(authStore);
    const store = get(authStore);
    
    // Check if the token is expired or about to expire
    if (store.expiresAt && store.expiresAt - TOKEN_EXPIRY_BUFFER < Date.now()) {
      // If we have a refresh token, we could auto-refresh here
      if (store.refreshToken && store.isAuthenticated) {
        console.log('Token is expired or about to expire. Refreshing...');
        // This would need to be handled async-friendly in a real app
        // For now we'll just warn and continue with the stale token
      }
    }
    
    return store;
  };
  
  // Get auth token for use in API calls
  const getAuthToken = () => {
    const state = get(authStore);
    return state.accessToken;
  };
  
  return {
    register,
    login,
    verifyEmail,
    resendVerificationCode,
    handleGoogleAuth,
    logout,
    refreshToken: authStore.refreshToken,
    getAuthState,
    getAuthToken,
    subscribe: authStore.subscribe
  };
}