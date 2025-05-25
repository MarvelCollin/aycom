import { writable, get } from 'svelte/store';
import type { IUserRegistration, IGoogleCredentialResponse, ITokenResponse, IAuthStore } from '../interfaces/IAuth';
import { setAuthData, clearAuthData, getAuthToken } from '../utils/auth';
import * as authApi from '../api/auth';
import appConfig from '../config/appConfig';
import { uploadFile } from '../utils/supabase';
import { getProfile } from '../api/user';

const API_URL = appConfig.api.baseUrl;
const TOKEN_EXPIRY_BUFFER = 300000;

interface AuthState extends IAuthStore {
  expiresAt: number | null;
  username?: string;
  displayName?: string;
  is_admin: boolean;
}

const createAuthStore = () => {
  const initialState: AuthState = {
    isAuthenticated: false,
    userId: null,
    accessToken: null,
    refreshToken: null,
    expiresAt: null,
    is_admin: false
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
          refreshExpiredToken(parsedAuth.refreshToken);
        } else {
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
          expiresAt,
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
          isAuthenticated: true,
          userId: data.user_id,
          accessToken: data.access_token,
          refreshToken: data.refresh_token,
          expiresAt,
          is_admin: data.user_data?.is_admin || false
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
      const data = await authApi.login(email, password);
      
      if (data.success && data.access_token) {
        const expiresAt = Date.now() + (data.expires_in * 1000);
        
        // Get the user's profile to check if they are an admin
        const userProfile = await getProfile();
        const userData = userProfile?.user;
        
        // Update auth state with user info including admin status
        const authState: AuthState = {
          isAuthenticated: true,
          userId: data.user_id,
          accessToken: data.access_token,
          refreshToken: data.refresh_token,
          expiresAt,
          username: userData?.username,
          displayName: userData?.name || userData?.display_name,
          is_admin: userData?.is_admin || false
        };
        
        authStore.set(authState);
        
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
          isAuthenticated: true,
          userId: data.user_id,
          accessToken: data.access_token,
          refreshToken: data.refresh_token,
          expiresAt,
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
    
    if (store.expiresAt && store.expiresAt - TOKEN_EXPIRY_BUFFER < Date.now()) {
      if (store.refreshToken && store.isAuthenticated) {
        console.log('Token is expired or about to expire. Refreshing...');
      }
    }
    
    return store;
  };
  
  const getAuthToken = () => {
    const state = get(authStore);
    return state.accessToken;
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
    subscribe: authStore.subscribe
  };
}