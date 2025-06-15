import { writable, get } from 'svelte/store';
import type { IUserRegistration, IGoogleCredentialResponse, ITokenResponse, IAuthStore } from '../interfaces/IAuth';
import { setAuthData, clearAuthData, getAuthToken, ensureTokenFreshness } from '../utils/auth';
import * as authApi from '../api/auth';
import appConfig from '../config/appConfig';
import { uploadFile } from '../utils/supabase';
import { getProfile, checkAdminStatus } from '../api/user';
import { createLoggerWithPrefix } from '../utils/logger';
import * as userApi from '../api/user';

const API_BASE_URL = appConfig.api.baseUrl;
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
      if (authState.access_token && authState.user_id) {
        setAuthData({
          accessToken: authState.access_token,
          refreshToken: authState.refresh_token || undefined,
          userId: authState.user_id,
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
  
  // Add a function to update admin status
  const updateAdminStatus = (isAdmin: boolean) => {
    auth.update((state) => {
      return { ...state, is_admin: isAdmin };
    });
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
    refreshToken: refreshExpiredToken,
    updateAdminStatus
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
      
      const enrichedUserData = {
        ...userData,
        ...(profilePictureUrl ? { profile_picture_url: profilePictureUrl } : {}),
        ...(bannerUrl ? { banner_url: bannerUrl } : {})
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
    try {
      return await userApi.getProfile();
    } catch (error) {
      console.error('Failed to get current user:', error);
      return null;
    }
  };
  
  // Alias for getCurrentUser for clarity and consistency with API
  const getProfile = getCurrentUser;
  
  const login = async (email: string, password: string, recaptchaToken: string | null = null) => {
    try {
      // Clear any existing auth data before login to prevent token issues
      clearAuthData();
      
      const data = await authApi.login(email, password, recaptchaToken);
      
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
        
        // Safely check for is_admin property in various locations
        const userData = data.user || {};
        const userDataObj = (data as any).user_data || {};
        
        if ((data as any).is_admin === true || userData.is_admin === true || userDataObj.is_admin === true) {
          isAdmin = true;
        }
        
        const authState: AuthState = {
          is_authenticated: true,
          user_id: data.user_id,
          access_token: data.access_token,
          refresh_token: data.refresh_token || null,
          expires_at: expiresAt,
          is_admin: isAdmin,
          username: userData.username,
          display_name: userData.name || userData.display_name
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
          message: (data as any).message || 'Login successful!'
        };
      } else {
        return {
          success: false,
          message: (data as any).message || 'Login failed. Please check your credentials.'
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
      
      // Log attempt to authenticate with Google
      console.log('Attempting to authenticate with Google token');
      const data = await authApi.googleLogin(response.credential);
      console.log('Google login API response:', { 
        success: data.success, 
        has_token: !!data.access_token,
        user_id: data.user_id,
        is_new_user: (data as any).is_new_user
      });
      
      if (data.success && data.access_token) {
        // Calculate token expiration time
        const expiresAt = data.expires_in 
          ? Date.now() + (data.expires_in * 1000) 
          : Date.now() + (3600 * 1000);
        
        // Get user data from the most appropriate source
        const userData = data.user || (data as any).user_data || {};
        
        // Construct and set authentication state
        const authState: AuthState = {
          is_authenticated: true,
          user_id: data.user_id,
          access_token: data.access_token,
          refresh_token: data.refresh_token || null,
          expires_at: expiresAt,
          is_admin: userData.is_admin || false,
          username: userData.username,
          display_name: userData.name || userData.display_name
        };
        
        console.log('Setting auth state with token and user info');
        authStore.set(authState);
        
        // Check if profile information is complete
        let requiredFields = ['gender', 'date_of_birth', 'security_question', 'security_answer'];
        let missingFields: string[] = [];
        
        // Always check for missing profile fields, regardless of what the server says
        // This ensures we catch any incomplete fields even if the server doesn't flag it
          try {
          const logger = createLoggerWithPrefix('ProfileCheck');
          logger.info('Fetching profile to check for missing information');
            const userProfile = await getProfile();
          
            if (userProfile?.user) {
            const profileData = userProfile.user;
              
              // Check for missing required fields
            if (!profileData.gender || profileData.gender === 'unknown') {
              missingFields.push('gender');
              logger.info('Missing gender information');
            }
            
            if (!profileData.date_of_birth || profileData.date_of_birth === '') {
              missingFields.push('date_of_birth');
              logger.info('Missing date of birth information');
            }
            
            if (!profileData.security_question) {
              missingFields.push('security_question');
              logger.info('Missing security question');
            }
            
            if (!profileData.security_answer) {
              missingFields.push('security_answer');
              logger.info('Missing security answer');
            }
            
            logger.info(`Profile check complete. Missing fields: ${missingFields.length > 0 ? missingFields.join(', ') : 'None'}`);
              
              // Update auth store with any available information
              authStore.update(state => ({
                ...state,
              username: profileData.username || state.username,
              display_name: profileData.name || profileData.display_name || state.display_name,
              is_admin: profileData.is_admin === true || state.is_admin
              }));
          } else {
            logger.warn('User profile data not available');
            // If we can't get profile data, assume we need to complete profile
            missingFields = requiredFields;
            }
          } catch (profileError) {
            console.error('Failed to get user profile after Google login:', profileError);
          // If we can't fetch the profile, assume we need to complete it
          missingFields = requiredFields;
        }
        
        return {
          success: true,
          message: (data as any).message || 'Google login successful!',
          is_new_user: (data as any).is_new_user || false,
          missing_fields: missingFields,
          requires_profile_completion: missingFields.length > 0
        };
      } else {
        console.error('Google login response lacks access token:', data);
        return {
          success: false,
          message: (data as any).message || 'Google login failed.'
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
  
  const updateProfile = async (data: Record<string, any>) => {
    try {
      const response = await fetch(`${API_BASE_URL}/users/profile`, {
        method: "PUT",
        headers: { 
          "Content-Type": "application/json",
          "Authorization": `Bearer ${getAuthToken()}`
        },
        body: JSON.stringify(data),
        credentials: "include",
      });
      
      const responseData = await handleApiResponse(response, 'Update profile failed');
      
      return {
        success: true,
        message: responseData.message || 'Profile updated successfully'
      };
    } catch (error) {
      console.error('Update profile error:', error);
      return handleApiError(error);
    }
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
    getProfile,
    subscribe: authStore.subscribe,
    checkAndRefreshTokenIfNeeded,
    updateProfile
  };
}

// Add this helper function to handle API responses
async function handleApiResponse(response: Response, errorMessage: string) {
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.message || errorMessage || 'API request failed');
  }
  return await response.json();
}