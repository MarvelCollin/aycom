import { writable } from 'svelte/store';
import type { TokenResponse, GoogleCredentialResponse } from '../interfaces/auth';

// Authentication store
export const authStore = writable({
  isAuthenticated: false,
  userId: null,
  accessToken: null,
  refreshToken: null
});

// Hook for authentication functions
export function useAuth() {
  // Login with email and password
  const login = async (email: string, password: string) => {
    try {
      const response = await fetch('/api/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ email, password })
      });
      
      const result = await response.json();
      
      if (result.access_token) {
        storeTokens(result);
        return { success: true };
      } else {
        return { success: false, message: result.message || 'Login failed' };
      }
    } catch (error) {
      console.error('Login error:', error);
      return { success: false, message: 'An error occurred during login' };
    }
  };

  // Register a new user
  const register = async (userData: any) => {
    try {
      const response = await fetch('/api/auth/register', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(userData)
      });
      
      const result = await response.json();
      return result;
    } catch (error) {
      console.error('Registration error:', error);
      return { success: false, message: 'An error occurred during registration' };
    }
  };

  // Verify email with code
  const verifyEmail = async (email: string, verificationCode: string) => {
    try {
      const response = await fetch('/api/auth/verify-email', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          email,
          verification_code: verificationCode
        })
      });
      
      const result = await response.json();
      
      if (result.access_token) {
        storeTokens(result);
        return { success: true };
      } else {
        return { success: false, message: 'Verification failed' };
      }
    } catch (error) {
      console.error('Verification error:', error);
      return { success: false, message: 'An error occurred during verification' };
    }
  };

  // Resend verification code
  const resendVerificationCode = async (email: string) => {
    try {
      const response = await fetch('/api/auth/resend-code', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ email })
      });
      
      const result = await response.json();
      return result;
    } catch (error) {
      console.error('Resend code error:', error);
      return { success: false, message: 'An error occurred' };
    }
  };

  // Handle Google authentication
  const handleGoogleAuth = async (response: GoogleCredentialResponse) => {
    try {
      const result = await fetch('/api/auth/google', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          token_id: response.credential
        })
      });
      
      const data = await result.json();
      
      if (data.access_token) {
        storeTokens(data);
        return { success: true };
      } else {
        return { success: false, message: data.message || 'Google authentication failed' };
      }
    } catch (error) {
      console.error('Error during Google authentication:', error);
      return { success: false, message: 'An error occurred during Google authentication' };
    }
  };

  // Logout user
  const logout = () => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user_id');
    
    authStore.set({
      isAuthenticated: false,
      userId: null,
      accessToken: null,
      refreshToken: null
    });
    
    // Redirect to login page
    window.location.href = '/login';
  };

  // Store authentication tokens
  const storeTokens = (tokenData: TokenResponse) => {
    localStorage.setItem('access_token', tokenData.access_token);
    localStorage.setItem('refresh_token', tokenData.refresh_token);
    localStorage.setItem('user_id', tokenData.user_id);
    
    authStore.set({
      isAuthenticated: true,
      userId: tokenData.user_id,
      accessToken: tokenData.access_token,
      refreshToken: tokenData.refresh_token
    });
  };

  // Initialize auth state from localStorage
  const initAuth = () => {
    const accessToken = localStorage.getItem('access_token');
    const refreshToken = localStorage.getItem('refresh_token');
    const userId = localStorage.getItem('user_id');
    
    if (accessToken && userId) {
      authStore.set({
        isAuthenticated: true,
        userId,
        accessToken,
        refreshToken
      });
    }
  };

  return {
    login,
    register,
    verifyEmail,
    resendVerificationCode,
    handleGoogleAuth,
    logout,
    initAuth,
    authStore
  };
} 