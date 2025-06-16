import { writable, get } from "svelte/store";
import { getAuthToken, clearAuthData } from "../utils/auth";
import type { IAuthStore } from "../interfaces/IAuth";
import { createLoggerWithPrefix } from "../utils/logger";

const logger = createLoggerWithPrefix("AuthStore");

interface AuthState extends IAuthStore {
  expires_at: number | null;
  username?: string;
  display_name?: string;
  is_admin: boolean;
}

const createAuthStore = () => {
  const initialState: AuthState = {
    is_authenticated: false,
    user_id: null,
    access_token: null,
    refresh_token: null,
    expires_at: null,
    is_admin: false
  };

  const auth = writable<AuthState>(initialState);

  const initAuth = () => {
    try {
      const storedAuth = localStorage.getItem("auth");
      if (storedAuth) {
        const parsedAuth = JSON.parse(storedAuth) as AuthState;
        auth.set(parsedAuth);
      }
    } catch (error) {
      logger.error("Failed to initialize auth from localStorage:", error);
    }
  };

  initAuth();

  return {
    subscribe: auth.subscribe,

    isAuthenticated: () => {
      const state = get(auth);
      return state.is_authenticated;
    },

    getUserId: () => {
      const state = get(auth);
      return state.user_id;
    },

    getToken: () => {
      const state = get(auth);
      return state.access_token;
    },

    logout: () => {
      clearAuthData();
      auth.set(initialState);
    },

    isAdmin: () => {
      const state = get(auth);
      return state.is_admin;
    }
  };
};

export const authStore = createAuthStore();