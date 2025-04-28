import { writable, get } from 'svelte/store';
import type { IUserRegistration, IGoogleCredentialResponse, ITokenResponse } from '../interfaces/IAuth';

// In a real app, these would be fetched from an API server
const API_URL = import.meta.env.VITE_API_URL || 'https://api.example.com';

// Create auth store
const createAuthStore = () => {
  const auth = writable({
    isAuthenticated: false,
    userId: null as string | null,
    accessToken: null as string | null,
    refreshToken: null as string | null
  });
  
  // Initialize store from localStorage
  const initAuth = () => {
    try {
      const storedAuth = localStorage.getItem('auth');
      if (storedAuth) {
        const parsedAuth = JSON.parse(storedAuth);
        auth.set(parsedAuth);
      }
    } catch (error) {
      console.error('Failed to initialize auth from localStorage:', error);
    }
  };
  
  // Save auth state to localStorage
  const persistAuth = (authState: any) => {
    try {
      localStorage.setItem('auth', JSON.stringify(authState));
    } catch (error) {
      console.error('Failed to persist auth to localStorage:', error);
    }
  };
  
  return {
    subscribe: auth.subscribe,
    set: (value: any) => {
      auth.set(value);
      persistAuth(value);
    },
    update: (updater: (value: any) => any) => {
      auth.update((value) => {
        const updated = updater(value);
        persistAuth(updated);
        return updated;
      });
    },
    init: initAuth,
    logout: () => {
      auth.set({
        isAuthenticated: false,
        userId: null,
        accessToken: null,
        refreshToken: null
      });
      localStorage.removeItem('auth');
    }
  };
};

// Create and initialize the auth store
const authStore = createAuthStore();
authStore.init();

export function useAuth() {
  /**
   * Handles user registration
   * @param userData User registration data
   * @returns Result of the registration attempt
   */
  const register = async (userData: IUserRegistration) => {
    try {
      // In a real app, this would be an API call
      console.log('Registering user:', userData);
      
      // Mock successful registration
      return {
        success: true,
        message: 'Registration successful! Check your email for verification code.'
      };
    } catch (error) {
      console.error('Registration error:', error);
      return {
        success: false,
        message: error instanceof Error ? error.message : 'Registration failed. Please try again.'
      };
    }
  };
  
  /**
   * Verifies a user's email with the provided verification code
   * @param email User's email
   * @param code Verification code sent to the user's email
   * @returns Result of the verification attempt
   */
  const verifyEmail = async (email: string, code: string) => {
    try {
      // In a real app, this would be an API call
      console.log('Verifying email:', email, 'with code:', code);
      
      // Mock successful verification
      return {
        success: true,
        message: 'Email verification successful!'
      };
    } catch (error) {
      console.error('Email verification error:', error);
      return {
        success: false,
        message: error instanceof Error ? error.message : 'Email verification failed. Please try again.'
      };
    }
  };
  
  /**
   * Resends the verification code to the user's email
   * @param email User's email
   * @returns Result of the resend attempt
   */
  const resendVerificationCode = async (email: string) => {
    try {
      // In a real app, this would be an API call
      console.log('Resending verification code to:', email);
      
      // Mock successful resend
      return {
        success: true,
        message: 'Verification code has been resent.'
      };
    } catch (error) {
      console.error('Resend verification code error:', error);
      return {
        success: false,
        message: error instanceof Error ? error.message : 'Failed to resend verification code. Please try again.'
      };
    }
  };
  
  /**
   * Logs in a user with email and password
   * @param email User's email
   * @param password User's password
   * @returns Result of the login attempt
   */
  const login = async (email: string, password: string) => {
    try {
      // In a real app, this would be an API call
      console.log('Logging in user:', email);
      
      // Mock successful login
      const tokenResponse: ITokenResponse = {
        access_token: 'mock-access-token',
        refresh_token: 'mock-refresh-token',
        user_id: '123456',
        token_type: 'Bearer',
        expires_in: 3600
      };
      
      // Update auth store
      authStore.set({
        isAuthenticated: true,
        userId: tokenResponse.user_id,
        accessToken: tokenResponse.access_token,
        refreshToken: tokenResponse.refresh_token
      });
      
      return {
        success: true,
        message: 'Login successful!'
      };
    } catch (error) {
      console.error('Login error:', error);
      return {
        success: false,
        message: error instanceof Error ? error.message : 'Login failed. Please check your credentials and try again.'
      };
    }
  };
  
  /**
   * Handles Google authentication
   * @param response Google credential response
   * @returns Result of the Google auth attempt
   */
  const handleGoogleAuth = async (response: IGoogleCredentialResponse) => {
    try {
      // In a real app, this would be an API call to verify the token and get user info
      console.log('Handling Google auth with credential:', response.credential);
      
      // Mock successful Google auth
      const tokenResponse: ITokenResponse = {
        access_token: 'mock-google-access-token',
        refresh_token: 'mock-google-refresh-token',
        user_id: '789012',
        token_type: 'Bearer',
        expires_in: 3600
      };
      
      // Update auth store
      authStore.set({
        isAuthenticated: true,
        userId: tokenResponse.user_id,
        accessToken: tokenResponse.access_token,
        refreshToken: tokenResponse.refresh_token
      });
      
      return {
        success: true,
        message: 'Google authentication successful!'
      };
    } catch (error) {
      console.error('Google auth error:', error);
      return {
        success: false,
        message: error instanceof Error ? error.message : 'Google authentication failed. Please try again.'
      };
    }
  };
  
  /**
   * Logs out the current user
   */
  const logout = () => {
    authStore.logout();
  };
  
  /**
   * Gets the current authentication state
   * @returns Current auth state
   */
  const getAuthState = () => get(authStore);
  
  return {
    subscribe: authStore.subscribe,
    register,
    verifyEmail,
    resendVerificationCode,
    login,
    handleGoogleAuth,
    logout,
    getAuthState
  };
} 