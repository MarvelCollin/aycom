import appConfig from '../config/appConfig';
import { getAuthToken } from '../utils/auth';

const API_BASE_URL = appConfig.api.baseUrl;

async function handleApiResponse(response: Response, errorMessage: string = 'Operation failed') {
  console.log(`Response status: ${response.status}`);

  if (!response.ok) {
    try {
      const errorData = await response.json();
      console.error('Error data:', errorData);
      throw new Error(errorData.message || `${errorMessage} with status: ${response.status}`);
    } catch (parseError) {
      console.error('Failed to parse error response:', parseError);
      throw new Error(`${errorMessage} with status: ${response.status}`);
    }
  }

  const data = await response.json();
  console.log('Successful response with keys:', Object.keys(data));
  return data;
}

function standardizeUserResponse(data: any) {
  return {
    user: data.user ? {
      id: data.user.id,
      username: data.user.username,
      name: data.user.name || data.user.display_name,
      email: data.user.email,
      profile_picture_url: data.user.profile_picture_url || data.user.profilePictureUrl,
      is_verified: data.user.is_verified || data.user.verified || false,
      is_admin: data.user.is_admin || data.user.admin || false
    } : data.user,
    access_token: data.access_token || data.accessToken || data.token,
    refresh_token: data.refresh_token || data.refreshToken,
    expires_in: data.expires_in || data.expiresIn,
    token_type: data.token_type || data.tokenType || 'bearer',
    user_id: data.user_id || data.userId || (data.user ? data.user.id : null)
  };
}

export async function login(email: string, password: string) {
  console.log(`Attempting to login with email: ${email.substring(0, 3)}...${email.split('@')[1]}`);

  try {
    const response = await fetch(`${API_BASE_URL}/users/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
      credentials: "include",
    });

    const data = await handleApiResponse(response, 'Login failed');
    return standardizeUserResponse(data);
  } catch (error) {
    console.error('Login exception:', error);
    throw error;
  }
}

export async function register(data: Record<string, any>) {
  const response = await fetch(`${API_BASE_URL}/users/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
    credentials: "include",
  });

  return handleApiResponse(response, 'Registration failed');
}

export async function refreshToken(refreshToken: string) {
  const token = getAuthToken();

  const response = await fetch(`${API_BASE_URL}/auth/refresh-token`, {
    method: "POST",
    headers: { 
      "Content-Type": "application/json",
      "Authorization": token ? `Bearer ${token}` : '' 
    },
    body: JSON.stringify({ refresh_token: refreshToken }),
    credentials: "include",
  });

  return handleApiResponse(response, 'Token refresh failed');
}

export async function verifyEmail(email: string, verificationCode: string) {
  const response = await fetch(`${API_BASE_URL}/auth/verify-email`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, verification_code: verificationCode }),
    credentials: "include",
  });

  return handleApiResponse(response, 'Email verification failed');
}

export async function resendVerification(email: string) {
  const response = await fetch(`${API_BASE_URL}/auth/resend-verification`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email }),
    credentials: "include",
  });

  return handleApiResponse(response, 'Resend verification failed');
}

export async function googleLogin(tokenId: string) {
  try {
    console.log('Sending Google token to backend API for verification');
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 10000); 

    const response = await fetch(`${API_BASE_URL}/auth/google`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ token_id: tokenId }),
      credentials: "include",
      signal: controller.signal
    });

    clearTimeout(timeoutId);

    const data = await handleApiResponse(response, 'Google login failed');
    console.log('Google login successful');

    return standardizeUserResponse(data);
  } catch (error) {
    console.error('Google login request error:', error);
    if (error instanceof Error && error.name === 'AbortError') {
      throw new Error("Request timed out. The server might be down or not responding.");
    }
    throw error;
  }
}

export async function createAdminUser(data: Record<string, any>) {
  try {
    console.log("Creating admin user with data:", data);

    const response = await fetch(`${API_BASE_URL}/users/register`, {
      method: "POST",
      headers: { 
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        ...data,
        is_admin: true,  
        is_verified: true 
      }),
      credentials: "include",
    });

    return handleApiResponse(response, 'Admin user creation failed');
  } catch (error) {
    console.error("Admin user creation failed:", error);
    throw error;
  }
}