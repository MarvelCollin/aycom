import { writable, get } from 'svelte/store';
import type { IUserRegistration, IGoogleCredentialResponse, ITokenResponse } from '../interfaces/IAuth';

const API_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';

const createAuthStore = () => {
  const auth = writable({
    isAuthenticated: false,
    userId: null as string | null,
    accessToken: null as string | null,
    refreshToken: null as string | null
  });
  
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

const authStore = createAuthStore();
authStore.init();

export function useAuth() {
  const register = async (userData: IUserRegistration) => {
    try {
      const response = await fetch(`${API_URL}/auth/register`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(userData)
      });
      
      const data = await response.json();
      
      return {
        success: data.success,
        message: data.message || 'Registration successful! Check your email for verification code.'
      };
    } catch (error) {
      console.error('Registration error:', error);
      return {
        success: false,
        message: error instanceof Error ? error.message : 'Registration failed. Please try again.'
      };
    }
  };
  
  const verifyEmail = async (email: string, code: string) => {
    try {
      const response = await fetch(`${API_URL}/auth/verify-email`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ email, verification_code: code })
      });
      
      const data = await response.json();
      
      if (data.success && data.access_token) {
        authStore.set({
          isAuthenticated: true,
          userId: data.user_id,
          accessToken: data.access_token,
          refreshToken: data.refresh_token
        });
      }
      
      return {
        success: data.success,
        message: data.message || 'Email verification successful!'
      };
    } catch (error) {
      console.error('Email verification error:', error);
      return {
        success: false,
        message: error instanceof Error ? error.message : 'Email verification failed. Please try again.'
      };
    }
  };
  
  const resendVerificationCode = async (email: string) => {
    try {
      const response = await fetch(`${API_URL}/auth/resend-code`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ email })
      });
      
      const data = await response.json();
      
      return {
        success: data.success,
        message: data.message || 'Verification code has been resent.'
      };
    } catch (error) {
      console.error('Resend verification code error:', error);
      return {
        success: false,
        message: error instanceof Error ? error.message : 'Failed to resend verification code. Please try again.'
      };
    }
  };
  
  const login = async (email: string, password: string) => {
    try {
      const response = await fetch(`${API_URL}/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ email, password })
      });
      
      const data = await response.json();
      
      if (data.success && data.access_token) {
        authStore.set({
          isAuthenticated: true,
          userId: data.user_id,
          accessToken: data.access_token,
          refreshToken: data.refresh_token
        });
      }
      
      return {
        success: data.success,
        message: data.message || 'Login successful!'
      };
    } catch (error) {
      console.error('Login error:', error);
      return {
        success: false,
        message: error instanceof Error ? error.message : 'Login failed. Please check your credentials and try again.'
      };
    }
  };
  
  const handleGoogleAuth = async (response: IGoogleCredentialResponse) => {
    try {
      const apiResponse = await fetch(`${API_URL}/auth/google`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ token_id: response.credential })
      });
      
      const data = await apiResponse.json();
      
      if (data.success && data.access_token) {
        authStore.set({
          isAuthenticated: true,
          userId: data.user_id,
          accessToken: data.access_token,
          refreshToken: data.refresh_token
        });
      }
      
      return {
        success: data.success,
        message: data.message || 'Google authentication successful!'
      };
    } catch (error) {
      console.error('Google auth error:', error);
      return {
        success: false,
        message: error instanceof Error ? error.message : 'Google authentication failed. Please try again.'
      };
    }
  };
  
  const logout = () => {
    const token = get(authStore).accessToken;
    if (token) {
      fetch(`${API_URL}/auth/logout`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ 
          access_token: get(authStore).accessToken,
          refresh_token: get(authStore).refreshToken
        })
      }).catch(error => {
        console.error('Logout error:', error);
      });
    }
    
    authStore.logout();
  };
  
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