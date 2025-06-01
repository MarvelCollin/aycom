import { writable, get } from 'svelte/store';
import type { IUserRegistration, IGoogleCredentialResponse, ITokenResponse, IAuthStore } from '../interfaces/IAuth';
import { setAuthData, clearAuthData, getAuthToken } from '../utils/auth';
import * as authApi from '../api/auth';
import appConfig from '../config/appConfig';
import { uploadFile } from '../utils/supabase';
import { getProfile, checkAdminStatus } from '../api/user';

const API_URL = appConfig.api.baseUrl;
const TOKEN_EXPIRY_BUFFER = 300000;

interface AuthState extends IAuthStore {
  expires_at: number | null;
  username?: string;
  display_name?: string;
  is_admin: boolean;
  userId?: string; // Add userId property for compatibility
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
        
        // Set userId for compatibility (maps from user_id)
        if (parsedAuth.user_id && !parsedAuth.userId) {
          parsedAuth.userId = parsedAuth.user_id;
        }
        
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
      console.log(`Login attempt for email: ${email}`);
      const data = await authApi.login(email, password);
      console.log('Login API response:', JSON.stringify(data, null, 2));
      
      // Check if response contains token information - various formats possible
      if ((data.success && data.access_token) || data.access_token) {
        // Use the token data from whatever format we receive
        const accessToken = data.access_token;
        const userId = data.user_id;
        const expiresIn = data.expires_in || 3600;
        const refreshToken = data.refresh_token || null;
        
        // Properly extract admin status from response
        let isAdmin = false;
        
        // Check all possible locations for admin flag in the response
        if (data.is_admin === true) {
          isAdmin = true;
          console.log('Admin status found directly in login response');
        } else if (data.user_data && data.user_data.is_admin === true) {
          isAdmin = true;
          console.log('Admin status found in user_data of login response');
        } else if (data.user && data.user.is_admin === true) {
          isAdmin = true;
          console.log('Admin status found in user object of login response');
        } else {
          console.log('No admin status found in login response');
        }
        
        const expiresAt = Date.now() + (expiresIn * 1000);
        
        console.log(`Login successful. Token exists: ${!!accessToken}, User ID: ${userId}, Admin: ${isAdmin}`);
        
        // Store auth data immediately to avoid issues with subsequent requests
        const initialAuthState: AuthState = {
          isAuthenticated: true,
          userId,
          accessToken,
          refreshToken,
          expiresAt,
          is_admin: isAdmin
        };
        
        // Update the store immediately
        authStore.set(initialAuthState);
        
        // Try multiple methods to determine admin status
        try {
          // Direct admin check via API
          const adminCheck = await checkAdminStatus();
          if (adminCheck) {
            isAdmin = true;
            console.log('Admin status confirmed via admin check API');
            authStore.update(state => ({ ...state, is_admin: true }));
          }
          
          // Also get the user's complete profile to update any missing information
          const userProfile = await getProfile();
          const userData = userProfile?.user;
          
          if (userData) {
            // Check admin status from profile response and preserve any existing admin status
            isAdmin = userData.is_admin === true || isAdmin;
            
            // Update auth state with user info including admin status
            const authState: AuthState = {
              ...initialAuthState,
              username: userData?.username,
              displayName: userData?.name || userData?.display_name,
              is_admin: isAdmin
            };
            
            authStore.set(authState);
            console.log('Auth state updated with user profile, admin status:', isAdmin);
          } else {
            console.log('User profile data not found in response, keeping initial auth state');
          }
        } catch (profileError) {
          console.error('Failed to get user profile after login:', profileError);
          // Continue with login even if profile fetch fails
        }
        
        return {
          success: true,
          message: 'Login successful!'
        };
      } else {
        console.error('Login failed: Invalid or missing token in response');
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
    
    // Set userId from user_id for compatibility
    if (store.user_id && !store.userId) {
      store.userId = store.user_id;
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