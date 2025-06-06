import { writable, get } from 'svelte/store';
import type { IUserRegistration, IGoogleCredentialResponse, ITokenResponse, IAuthStore } from '../interfaces/IAuth';
import { setAuthData, clearAuthData, getAuthToken, ensureTokenFreshness } from '../utils/auth';
import * as authApi from '../api/auth';
import appConfig from '../config/appConfig';
import { uploadFile } from '../utils/supabase';
import { getProfile, checkAdminStatus } from '../api/user';
import { createLoggerWithPrefix } from '../utils/logger';

const API_URL = appConfig.api.baseUrl;
const TOKEN_EXPIRY_BUFFER = 300000; // 5 minutes in milliseconds
const logger = createLoggerWithPrefix('Auth');

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
      const storedAuth = localStorage.getItem('auth');
      if (storedAuth) {
        const parsedAuth = JSON.parse(storedAuth) as AuthState;
        
        const now = Date.now();
        const isAboutToExpire = parsedAuth.expires_at && (parsedAuth.expires_at - now) < TOKEN_EXPIRY_BUFFER;
        
        if (parsedAuth.expires_at && !isAboutToExpire) {
          auth.set(parsedAuth);
          
          if (parsedAuth.expires_at) {
            startTokenRefreshTimer(parsedAuth.expires_at, parsedAuth.refresh_token);
          }
        } else if (parsedAuth.refresh_token) {
          refreshExpiredToken(parsedAuth.refresh_token);
        } else {
          clearAuth();
        }
      }
    } catch (error) {
      console.error('Failed to initialize auth from localStorage:', error);
      clearAuth();
    }
  };

  // Setup a timer to check and refresh token before it expires
  let tokenRefreshTimer: number | null = null;
  
  const startTokenRefreshTimer = (expiresAt: number, refreshToken: string | null) => {
    if (tokenRefreshTimer !== null) {
      window.clearTimeout(tokenRefreshTimer);
      tokenRefreshTimer = null;
    }
    
    if (!refreshToken) return;
    
    const timeUntilRefresh = expiresAt - Date.now() - TOKEN_EXPIRY_BUFFER;
    
    if (timeUntilRefresh > 0) {
      console.log(`Token refresh scheduled in ${Math.round(timeUntilRefresh/1000)} seconds`);
      tokenRefreshTimer = window.setTimeout(() => {
        console.log('Token refresh timer triggered');
        if (refreshToken) {
          refreshExpiredToken(refreshToken);
        }
      }, timeUntilRefresh);
    } else {
      // Token is already expired or about to expire, refresh it immediately
      if (refreshToken) {
        refreshExpiredToken(refreshToken);
      }
    }
  };
  
  const persistAuth = (authState: AuthState) => {
    try {
      if (authState.access_token && authState.userId) {
        setAuthData({
          accessToken: authState.access_token,
          refreshToken: authState.refresh_token || undefined,
          userId: authState.userId,
          expiresAt: authState.expires_at || undefined,
          is_admin: authState.is_admin
        });
        
        // Setup refresh timer whenever we persist auth state
        if (authState.expires_at && authState.refresh_token) {
          startTokenRefreshTimer(authState.expires_at, authState.refresh_token);
        }
      } else {
        localStorage.setItem('auth', JSON.stringify(authState));
      }
    } catch (error) {
      console.error('Failed to persist auth to localStorage:', error);
    }
  };

  const clearAuth = () => {
    if (tokenRefreshTimer !== null) {
      window.clearTimeout(tokenRefreshTimer);
      tokenRefreshTimer = null;
    }
    clearAuthData();
    auth.set(initialState);
  };

  const refreshExpiredToken = async (refreshToken: string) => {
    try {
      const data = await authApi.refreshToken(refreshToken);
      
      if (data.success && data.access_token) {
        const expiresAt = Date.now() + (data.expires_in * 1000);
        const newState: AuthState = {
          is_authenticated: true,
          user_id: data.user_id,
          access_token: data.access_token,
          refresh_token: data.refresh_token,
          expires_at: expiresAt,
          is_admin: data.user_data?.is_admin || false
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
  
  const registerWithMedia = async (userData: IUserRegistration, profilePicture: File | null, banner: File | null) => {
    try {
      let profilePictureUrl: string | null = null;
      let profileUploadError = false;
      if (profilePicture) {
        try {
          profilePictureUrl = await uploadFile(profilePicture, 'profile-pictures', 'users');
          if (!profilePictureUrl) {
            profileUploadError = true;
            console.warn('Profile picture upload failed, continuing without profile picture');
          }
        } catch (uploadError) {
          profileUploadError = true;
          console.error('Profile picture upload error:', uploadError);
        }
      }
      
      let bannerUrl: string | null = null;
      let bannerUploadError = false;
      if (banner) {
        try {
          bannerUrl = await uploadFile(banner, 'banners', 'users');
          if (!bannerUrl) {
            bannerUploadError = true;
            console.warn('Banner upload failed, continuing without banner');
          }
        } catch (uploadError) {
          bannerUploadError = true;
          console.error('Banner upload error:', uploadError);
        }
      }
      
      const enrichedUserData: IUserRegistration = {
        ...userData,
        profile_picture_url: profilePictureUrl || '',
        banner_url: bannerUrl || ''
      };
      
      console.log('Registering user with data:', JSON.stringify(enrichedUserData, null, 2));
      
      const data = await authApi.register(enrichedUserData);
      
      let message = data.message || 'Registration successful! Check your email for verification code.';
      if (profileUploadError || bannerUploadError) {
        message += ' Note: Some media uploads failed. You can update your profile later.';
      }
      
      return {
        success: data.success,
        message: message
      };
    } catch (error) {
      console.error('Registration with media error:', error);
      return handleApiError(error);
    }
  };
  
  const verifyEmail = async (email: string, code: string) => {
    try {
      const data = await authApi.verifyEmail(email, code);
      
      if (data.success && data.access_token) {
        const expiresAt = Date.now() + (data.expires_in * 1000);
        authStore.set({
          is_authenticated: true,
          user_id: data.user_id,
          access_token: data.access_token,
          refresh_token: data.refresh_token,
          expires_at: expiresAt,
          is_admin: false
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
    return null;
  };
  
  const login = async (email: string, password: string) => {
    try {
      // Clear any existing auth data before login to prevent token issues
      clearAuthData();
      
      const data = await authApi.login(email, password);
      
      logger.info(`Login response received with token: ${data.access_token ? 'yes' : 'no'}`);
      
      if (data.access_token) {
        // Simple JWT inspection to debug token issues
        try {
          const tokenParts = data.access_token.split('.');
          if (tokenParts.length === 3) {
            // Base64 decode the payload (middle part)
            const payload = JSON.parse(atob(tokenParts[1]));
            logger.info(`JWT payload inspection: sub=${payload.sub}, user_id=${payload.user_id}, exp=${payload.exp}`);
          } else {
            logger.warn(`JWT token has unexpected format (${tokenParts.length} parts instead of 3)`);
          }
        } catch (tokenError) {
          logger.error('Error inspecting JWT token:', tokenError);
        }
        
        const expiresAt = Date.now() + ((data.expires_in || 3600) * 1000);
        let isAdmin = false;
        
        if (data.is_admin === true || data.user?.is_admin === true || data.user_data?.is_admin === true) {
          isAdmin = true;
        }
        
        const authState: AuthState = {
          is_authenticated: true,
          user_id: data.user_id,
          access_token: data.access_token,
          refresh_token: data.refresh_token || null,
          expires_at: expiresAt,
          is_admin: isAdmin,
          username: data.user?.username,
          display_name: data.user?.name || data.user?.display_name
        };
        
        logger.info(`Setting auth state with user_id=${data.user_id}, expires_at=${new Date(expiresAt).toISOString()}`);
        authStore.set(authState);
        
        try {
          const adminCheck = await checkAdminStatus();
          if (adminCheck && !authState.is_admin) {
            authStore.update(state => ({ ...state, is_admin: true }));
          }
          
          const userProfile = await getProfile();
          if (userProfile?.user) {
            const userData = userProfile.user;
            authStore.update(state => ({
              ...state,
              username: userData?.username,
              display_name: userData?.name || userData?.display_name,
              is_admin: userData?.is_admin === true || state.is_admin
            }));
          }
        } catch (profileError) {
          console.error('Failed to get user profile after login:', profileError);
        }
        
        return {
          success: true,
          message: 'Login successful!'
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
        const expiresAt = data.expires_in 
          ? Date.now() + (data.expires_in * 1000) 
          : Date.now() + (3600 * 1000);
        
        const authState: AuthState = {
          is_authenticated: true,
          user_id: data.user_id,
          access_token: data.access_token,
          refresh_token: data.refresh_token,
          expires_at: expiresAt,
          is_admin: data.user_data?.is_admin || false
        };
        
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
    authStore.logout();
    return {
      success: true,
      message: 'Logged out successfully'
    };
  };
    const getAuthState = () => {
    const store = get(authStore);
    
    if (store.expires_at && store.expires_at - TOKEN_EXPIRY_BUFFER < Date.now()) {
      if (store.refresh_token && store.is_authenticated) {
        console.log('Token is expired or about to expire. Refreshing...');
      }
    }
    
    return store;
  };
  
  const getAuthToken = () => {
    const state = get(authStore);
    return state.access_token;
  };
  
  // Check if token needs refresh and do it proactively
  const checkAndRefreshTokenIfNeeded = async () => {
    const state = get(authStore);
    if (!state.is_authenticated) {
      return Promise.resolve();
    }
    
    if (state && state.refresh_token && state.expires_at) {
      const now = Date.now();
      const timeUntilExpiry = state.expires_at - now;
      const needsRefresh = timeUntilExpiry < TOKEN_EXPIRY_BUFFER;
      
      if (needsRefresh) {
        logger.info(`Token expires in ${Math.floor(timeUntilExpiry/1000)}s, refreshing now...`);
        try {
          return await authStore.refreshToken(state.refresh_token);
        } catch (error) {
          logger.error('Failed to refresh token during proactive check:', error);
          
          // If token is completely expired, clear auth state
          if (timeUntilExpiry <= 0) {
            logger.warn('Token is completely expired, clearing auth state');
            authStore.logout();
          }
          
          throw error;
        }
      } else {
        logger.debug(`Token still valid for ${Math.floor(timeUntilExpiry/1000)}s, no refresh needed`);
      }
    } else if (state.is_authenticated) {
      // We're authenticated but don't have proper refresh info
      logger.warn('Missing refresh token or expiry but user is authenticated - this may cause issues');
    }
    
    return Promise.resolve();
  };
  
  return {
    register,
    registerWithMedia,
    login,
    verifyEmail,
    resendVerificationCode,
    handleGoogleAuth,
    logout,
    refreshToken: authStore.refreshToken,
    getAuthState,
    getAuthToken,
    subscribe: authStore.subscribe,
    checkAndRefreshTokenIfNeeded
  };
}