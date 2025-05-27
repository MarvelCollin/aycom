// Block/unblock/report API functions
export async function blockUser(userId: string): Promise<boolean> {
  try {
    const token = getAuthToken();
    
    const response = await fetch(`${API_BASE_URL}/users/${userId}/block`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      }
    });
    
    if (!response.ok) {
      throw new Error(`Failed to block user: ${response.status}`);
    }
    
    return true;
  } catch (err) {
    console.error('Failed to block user:', err);
    return false;
  }
}

export async function unblockUser(userId: string): Promise<boolean> {
  try {
    const token = getAuthToken();
    
    const response = await fetch(`${API_BASE_URL}/users/${userId}/unblock`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      }
    });
    
    if (!response.ok) {
      throw new Error(`Failed to unblock user: ${response.status}`);
    }
    
    return true;
  } catch (err) {
    console.error('Failed to unblock user:', err);
    return false;
  }
}

export async function getBlockedUsers(page = 1, limit = 20): Promise<any[]> {
  try {
    const token = getAuthToken();
    
    const response = await fetch(`${API_BASE_URL}/users/blocked?page=${page}&limit=${limit}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      }
    });
    
    if (!response.ok) {
      throw new Error(`Failed to get blocked users: ${response.status}`);
    }
    
    const data = await response.json();
    
    if (data && data.data && data.data.blocked_users) {
      return data.data.blocked_users.map((user: any) => ({
        id: user.id,
        name: user.name || user.display_name,
        username: user.username,
        profile_picture: user.profile_picture_url || '👤',
        verified: user.is_verified || false
      }));
    }
    
    return [];
  } catch (err) {
    console.error('Failed to get blocked users:', err);
    return [];
  }
}

export async function reportUser(userId: string, reason: string): Promise<boolean> {
  try {
    const token = getAuthToken();
    
    const response = await fetch(`${API_BASE_URL}/users/${userId}/report`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      body: JSON.stringify({ reason })
    });
    
    if (!response.ok) {
      throw new Error(`Failed to report user: ${response.status}`);
    }
    
    return true;
  } catch (err) {
    console.error('Failed to report user:', err);
    return false;
  }
}

export async function isUserBlocked(userId: string): Promise<boolean> {
  try {
    const token = getAuthToken();
    
    // This endpoint doesn't exist yet, so we'll skip the implementation for now
    // This function would check if the current user has blocked the specified user
    
    return false;
  } catch (err) {
    console.error('Failed to check if user is blocked:', err);
    return false;
  }
}
