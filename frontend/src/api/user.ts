import appConfig from '../config/appConfig';
import { getAuthToken } from '../utils/auth';

const API_BASE_URL = appConfig.api.baseUrl;

export async function getProfile() {
  const token = getAuthToken();
  
  const response = await fetch(`${API_BASE_URL}/users/profile`, {
    method: "GET",
    headers: { 
      "Content-Type": "application/json",
      "Authorization": token ? `Bearer ${token}` : ''
    },
    credentials: "include",
  });
  if (!response.ok) {
    try {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to fetch user profile");
    } catch (parseError) {
      throw new Error("Failed to fetch user profile");
    }
  }
  return response.json();
}

export async function updateProfile(data: Record<string, any>) {
  const token = getAuthToken();
  
  const response = await fetch(`${API_BASE_URL}/users/profile`, {
    method: "PUT",
    headers: { 
      "Content-Type": "application/json",
      "Authorization": token ? `Bearer ${token}` : ''
    },
    body: JSON.stringify(data),
    credentials: "include",
  });
  if (!response.ok) {
    try {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to update user profile");
    } catch (parseError) {
      throw new Error("Failed to update user profile");
    }
  }
  return response.json();
} 