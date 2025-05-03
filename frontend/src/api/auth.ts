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
  const response = await fetch(`${API_BASE_URL}/auth/google`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ token_id: tokenId }),
    credentials: "include",
  });
  if (!response.ok) {
    try {
      const errorData = await response.json();
      throw new Error(errorData.message || "Google login failed");
    } catch (parseError) {
      throw new Error("Google login failed");
    }
  }
  return response.json();
} 