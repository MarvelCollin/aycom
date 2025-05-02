const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081';

export async function getProfile() {
  const response = await fetch(`${API_BASE_URL}/users/profile`, {
    method: "GET",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
  });
  if (!response.ok) throw new Error("Failed to fetch user profile");
  return response.json();
}

export async function updateProfile(data: Record<string, any>) {
  const response = await fetch(`${API_BASE_URL}/users/profile`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
    credentials: "include",
  });
  if (!response.ok) throw new Error("Failed to update user profile");
  return response.json();
} 