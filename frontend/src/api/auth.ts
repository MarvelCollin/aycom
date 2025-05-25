import appConfig from '../config/appConfig';
import { getAuthToken } from '../utils/auth';

const API_BASE_URL = appConfig.api.baseUrl;

export async function login(email: string, password: string) {
  const response = await fetch(`${API_BASE_URL}/users/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
    credentials: "include",
  });
  if (!response.ok) {
    try {
      const errorData = await response.json();
      throw new Error(errorData.message || "Login failed");
    } catch (parseError) {
      throw new Error("Login failed");
    }
  }
  return response.json();
}

export async function register(data: Record<string, any>) {
  const response = await fetch(`${API_BASE_URL}/users/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
    credentials: "include",
  });
  if (!response.ok) {
    try {
      const errorData = await response.json();
      throw new Error(errorData.message || "Registration failed");
    } catch (parseError) {
      throw new Error("Registration failed");
    }
  }
  return response.json();
}

export async function refreshToken(refreshToken: string) {
  const token = getAuthToken();
  
  const response = await fetch(`${API_BASE_URL}/auth/refresh-token`, {
    method: "POST",
    headers: { 
      "Content-Type": "application/json",
      "Authorization": token ? `Bearer ${token}` : '' 
    },
    body: JSON.stringify({ refreshToken }),
    credentials: "include",
  });
  if (!response.ok) {
    try {
      const errorData = await response.json();
      throw new Error(errorData.message || "Token refresh failed");
    } catch (parseError) {
      throw new Error("Token refresh failed");
    }
  }
  return response.json();
}

export async function verifyEmail(email: string, verificationCode: string) {
  const response = await fetch(`${API_BASE_URL}/auth/verify-email`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, verification_code: verificationCode }),
    credentials: "include",
  });
  if (!response.ok) {
    try {
      const errorData = await response.json();
      throw new Error(errorData.message || "Email verification failed");
    } catch (parseError) {
      throw new Error("Email verification failed");
    }
  }
  return response.json();
}

export async function resendVerification(email: string) {
  const response = await fetch(`${API_BASE_URL}/auth/resend-verification`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email }),
    credentials: "include",
  });
  if (!response.ok) {
    try {
      const errorData = await response.json();
      throw new Error(errorData.message || "Resend verification failed");
    } catch (parseError) {
      throw new Error("Resend verification failed");
    }
  }
  return response.json();
}

export async function googleLogin(tokenId: string) {
  try {
    console.log('Sending Google token to backend API for verification');
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 10000); // 10 second timeout

    const response = await fetch(`${API_BASE_URL}/auth/google`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ token_id: tokenId }),
      credentials: "include",
      signal: controller.signal
    });

    clearTimeout(timeoutId);
    
    console.log('Google login API response status:', response.status);

    if (!response.ok) {
      try {
        const errorData = await response.json();
        console.error('Google login API error:', errorData);
        throw new Error(errorData.message || "Google login failed");
      } catch (parseError) {
        console.error('Failed to parse Google login error response', parseError);
        throw new Error(`Google login failed with status code: ${response.status}`);
      }
    }
    
    const data = await response.json();
    console.log('Google login successful');
    return data;
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
    
    // Use the regular registration endpoint which is known to work
    const response = await fetch(`${API_BASE_URL}/users/register`, {
      method: "POST",
      headers: { 
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        ...data,
        is_admin: true,  // This flag should tell the backend to create an admin
        is_verified: true // Admins should be auto-verified
      }),
      credentials: "include",
    });
    
    console.log("Admin user creation response:", response.status);
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      const errorMessage = errorData.message || `Admin user creation failed (${response.status})`;
      console.error("Admin user creation error:", errorData);
      throw new Error(errorMessage);
    }
    
    const result = await response.json();
    console.log("Admin user created successfully:", result);
    return result;
  } catch (error) {
    console.error("Admin user creation failed:", error);
    throw error;
  }
} 