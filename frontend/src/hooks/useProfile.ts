import { writable, get } from 'svelte/store';
import type { IUser, IUserProfile, IUserUpdateRequest } from '../interfaces/IUser';
import { getAuthToken } from '../utils/auth';

export function useProfile() {
  const profile = writable<IUserProfile | null>(null);
  const loading = writable(false);
  const error = writable<string | null>(null);
  
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081/api/v1';

  async function fetchProfile(userId: string): Promise<boolean> {
    loading.set(true);
    error.set(null);

    try {
      const token = getAuthToken();
      
      const endpoint = userId === 'me' 
        ? `${API_BASE_URL}/users/profile` 
        : `${API_BASE_URL}/users/${userId}`;
      
      const response = await fetch(endpoint, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': token ? `Bearer ${token}` : ''
        }
      });
      
      if (!response.ok) {
        throw new Error(`Failed to fetch profile: ${response.status}`);
      }
      
      const data = await response.json();
      
      if (data && data.user) {
        const userData = data.user;
        
        const userProfile: IUserProfile = {
          id: userData.id,
          name: userData.name || userData.display_name,
          username: userData.username,
          email: userData.email,
          bio: userData.bio || '',
          location: userData.location || '',
          website: userData.website || '',
          profile_picture: userData.profile_picture_url || '',
          banner: userData.banner_url || '',
          verified: userData.verified || false,
          followers_count: userData.followers_count || 0,
          following_count: userData.following_count || 0,
          tweets_count: userData.tweets_count || 0,
          joined_date: userData.created_at || new Date().toISOString(),
          birthday: userData.birthday || ''
        };
        
        profile.set(userProfile);
        return true;
      } else {
        throw new Error('Invalid profile data received from API');
      }
    } catch (err) {
      console.error('Failed to fetch profile:', err);
      error.set(err instanceof Error ? err.message : 'Failed to load profile data');
      return false;
    } finally {
      loading.set(false);
    }
  }

  // Function to update user profile
  async function updateProfile(data: IUserUpdateRequest): Promise<{ success: boolean; message?: string }> {
    loading.set(true);
    error.set(null);

    try {
      const token = getAuthToken();
      
      const response = await fetch(`${API_BASE_URL}/users/profile`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': token ? `Bearer ${token}` : ''
        },
        body: JSON.stringify(data)
      });
      
      if (!response.ok) {
        throw new Error(`Failed to update profile: ${response.status}`);
      }
      
      const responseData = await response.json();
      
      // Update the local profile with the new data
      const currentProfile = get(profile);
      if (currentProfile) {
        profile.set({
          ...currentProfile,
          ...data,
        });
      }
      
      return { 
        success: true, 
        message: responseData.message || 'Profile updated successfully' 
      };
    } catch (err) {
      console.error('Failed to update profile:', err);
      error.set(err instanceof Error ? err.message : 'Failed to update profile');
      return { 
        success: false, 
        message: err instanceof Error ? err.message : 'Failed to update profile' 
      };
    } finally {
      loading.set(false);
    }
  }

  // Function to check if a username is available
  async function checkUsernameAvailability(username: string): Promise<boolean> {
    try {
      const response = await fetch(`${API_BASE_URL}/users/check-username?username=${encodeURIComponent(username)}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json'
        }
      });
      
      if (!response.ok) {
        throw new Error(`Failed to check username: ${response.status}`);
      }
      
      const data = await response.json();
      return data.available;
    } catch (err) {
      console.error('Failed to check username availability:', err);
      return false;
    }
  }

  // Function to follow a user
  async function followUser(userId: string): Promise<boolean> {
    try {
      const token = getAuthToken();
      
      const response = await fetch(`${API_BASE_URL}/users/${userId}/follow`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': token ? `Bearer ${token}` : ''
        }
      });
      
      if (!response.ok) {
        throw new Error(`Failed to follow user: ${response.status}`);
      }
      
      // Update followers count in the profile
      const currentProfile = get(profile);
      if (currentProfile) {
        profile.set({
          ...currentProfile,
          followers_count: currentProfile.followers_count + 1
        });
      }
      
      return true;
    } catch (err) {
      console.error('Failed to follow user:', err);
      return false;
    }
  }

  // Function to unfollow a user
  async function unfollowUser(userId: string): Promise<boolean> {
    try {
      const token = getAuthToken();
      
      const response = await fetch(`${API_BASE_URL}/users/${userId}/unfollow`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': token ? `Bearer ${token}` : ''
        }
      });
      
      if (!response.ok) {
        throw new Error(`Failed to unfollow user: ${response.status}`);
      }
      
      // Update followers count in the profile
      const currentProfile = get(profile);
      if (currentProfile && currentProfile.followers_count > 0) {
        profile.set({
          ...currentProfile,
          followers_count: currentProfile.followers_count - 1
        });
      }
      
      return true;
    } catch (err) {
      console.error('Failed to unfollow user:', err);
      return false;
    }
  }

  // Function to get followers list
  async function getFollowers(userId: string, page = 1, limit = 20): Promise<IUser[]> {
    try {
      const token = getAuthToken();
      
      const response = await fetch(`${API_BASE_URL}/users/${userId}/followers?page=${page}&limit=${limit}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': token ? `Bearer ${token}` : ''
        }
      });
      
      if (!response.ok) {
        throw new Error(`Failed to get followers: ${response.status}`);
      }
      
      const data = await response.json();
      
      if (data && data.followers) {
        return data.followers.map((follower: any) => ({
          id: follower.id,
          name: follower.name || follower.display_name,
          username: follower.username,
          profile_picture: follower.profile_picture_url || '👤',
          verified: follower.verified || false,
          isFollowing: follower.is_following || false
        }));
      }
      
      return [];
    } catch (err) {
      console.error('Failed to get followers:', err);
      return [];
    }
  }

  // Function to get following list
  async function getFollowing(userId: string, page = 1, limit = 20): Promise<IUser[]> {
    try {
      const token = getAuthToken();
      
      const response = await fetch(`${API_BASE_URL}/users/${userId}/following?page=${page}&limit=${limit}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': token ? `Bearer ${token}` : ''
        }
      });
      
      if (!response.ok) {
        throw new Error(`Failed to get following: ${response.status}`);
      }
      
      const data = await response.json();
      
      if (data && data.following) {
        return data.following.map((following: any) => ({
          id: following.id,
          name: following.name || following.display_name,
          username: following.username,
          profile_picture: following.profile_picture_url || '👤',
          verified: following.verified || false,
          isFollowing: true
        }));
      }
      
      return [];
    } catch (err) {
      console.error('Failed to get following list:', err);
      return [];
    }
  }

  return {
    profile,
    loading,
    error,
    fetchProfile,
    updateProfile,
    checkUsernameAvailability,
    followUser,
    unfollowUser,
    getFollowers,
    getFollowing
  };
} 