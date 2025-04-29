import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export interface User {
  id: string;
  email: string;
  name: string;
  avatar?: string;
}

export interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  loading: boolean;
}

// Initialize state from localStorage if available
const getInitialState = (): AuthState => {
  if (browser) {
    const storedUser = localStorage.getItem('user');
    const storedToken = localStorage.getItem('token');
    
    return {
      user: storedUser ? JSON.parse(storedUser) : null,
      token: storedToken,
      isAuthenticated: !!storedToken,
      loading: false
    };
  }
  
  return {
    user: null,
    token: null,
    isAuthenticated: false,
    loading: false
  };
};

const createAuthStore = () => {
  const { subscribe, set, update } = writable<AuthState>(getInitialState());

  return {
    subscribe,
    
    login: async (email: string, password: string) => {
      update(state => ({ ...state, loading: true }));
      try {
        const response = await fetch('/api/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email, password }),
          credentials: 'include'
        });

        if (!response.ok) {
          const error = await response.json();
          throw new Error(error.message || 'Login failed');
        }

        const data = await response.json();
        const { token, user } = data;

        // Store in localStorage
        if (browser) {
          localStorage.setItem('token', token);
          localStorage.setItem('user', JSON.stringify(user));
        }

        set({
          user,
          token,
          isAuthenticated: true,
          loading: false
        });
        
        return { success: true };
      } catch (error) {
        update(state => ({ ...state, loading: false }));
        throw error;
      }
    },
    
    register: async (name: string, email: string, password: string) => {
      update(state => ({ ...state, loading: true }));
      try {
        const response = await fetch('/api/auth/register', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ name, email, password }),
          credentials: 'include'
        });

        if (!response.ok) {
          const error = await response.json();
          throw new Error(error.message || 'Registration failed');
        }

        const data = await response.json();
        const { token, user } = data;

        // Store in localStorage
        if (browser) {
          localStorage.setItem('token', token);
          localStorage.setItem('user', JSON.stringify(user));
        }

        set({
          user,
          token,
          isAuthenticated: true,
          loading: false
        });
        
        return { success: true };
      } catch (error) {
        update(state => ({ ...state, loading: false }));
        throw error;
      }
    },
    
    googleLogin: async (credential: string) => {
      update(state => ({ ...state, loading: true }));
      try {
        const response = await fetch('/api/auth/google', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ credential }),
          credentials: 'include'
        });

        if (!response.ok) {
          const error = await response.json();
          throw new Error(error.message || 'Google login failed');
        }

        const data = await response.json();
        const { token, user } = data;

        // Store in localStorage
        if (browser) {
          localStorage.setItem('token', token);
          localStorage.setItem('user', JSON.stringify(user));
        }

        set({
          user,
          token,
          isAuthenticated: true,
          loading: false
        });
        
        return { success: true };
      } catch (error) {
        update(state => ({ ...state, loading: false }));
        throw error;
      }
    },
    
    logout: () => {
      // Clear localStorage
      if (browser) {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
      }

      set({
        user: null,
        token: null,
        isAuthenticated: false,
        loading: false
      });
    },
    
    // Method to validate token and refresh user data
    checkAuth: async () => {
      const state = getInitialState();
      if (!state.token) return false;

      update(state => ({ ...state, loading: true }));
      try {
        const response = await fetch('/api/auth/me', {
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${state.token}`
          },
          credentials: 'include'
        });

        if (!response.ok) {
          throw new Error('Invalid token');
        }

        const user = await response.json();
        
        update(state => ({
          ...state,
          user,
          isAuthenticated: true,
          loading: false
        }));
        
        return true;
      } catch (error) {
        // Token invalid, clear auth
        if (browser) {
          localStorage.removeItem('token');
          localStorage.removeItem('user');
        }
        
        set({
          user: null,
          token: null,
          isAuthenticated: false,
          loading: false
        });
        
        return false;
      }
    }
  };
};

export const authStore = createAuthStore(); 