import { writable, get } from 'svelte/store';
import type { IUser, IUserProfile, IUserUpdateRequest } from '../interfaces/IUser';

export function useProfile() {
  // Profile data store
  const profile = writable<IUserProfile | null>(null);
  const loading = writable(false);
  const error = writable<string | null>(null);

  // Function to fetch user profile data
  async function fetchProfile(userId: string): Promise<boolean> {
    loading.set(true);
    error.set(null);

    try {
      // Mock API call - replace with actual API call
      console.log('Fetching profile for user:', userId);
      
      // Simulate API response
      const mockUserData: IUserProfile = {
        id: userId,
        name: 'John Doe',
        username: 'johndoe',
        email: 'johndoe@example.com',
        bio: 'This is my bio',
        location: 'New York',
        website: 'https://example.com',
        profile_picture: 'https://via.placeholder.com/150',
        banner: 'https://via.placeholder.com/1500x500',
        verified: false,
        followers_count: 250,
        following_count: 120,
        tweets_count: 65,
        joined_date: '2023-01-15',
        birthday: '1990-05-20'
      };
      
      profile.set(mockUserData);
      return true;
    } catch (err) {
      console.error('Failed to fetch profile:', err);
      error.set('Failed to load profile data');
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
      // Mock API call - replace with actual API call
      console.log('Updating profile with data:', data);
      
      // Simulate API response
      const currentProfile = get(profile);
      
      if (!currentProfile) {
        throw new Error('No profile loaded');
      }
      
      // Update the profile with the new data
      profile.set({
        ...currentProfile,
        ...data,
      });
      
      return { success: true, message: 'Profile updated successfully' };
    } catch (err) {
      console.error('Failed to update profile:', err);
      error.set('Failed to update profile');
      return { success: false, message: 'Failed to update profile' };
    } finally {
      loading.set(false);
    }
  }

  // Function to check if a username is available
  async function checkUsernameAvailability(username: string): Promise<boolean> {
    try {
      // Mock API call - replace with actual API call
      console.log('Checking username availability:', username);
      
      // Simulate a delay for the API call
      await new Promise(resolve => setTimeout(resolve, 500));
      
      // For demonstration, consider usernames containing 'taken' as unavailable
      return !username.toLowerCase().includes('taken');
    } catch (err) {
      console.error('Failed to check username availability:', err);
      return false;
    }
  }

  // Function to follow a user
  async function followUser(userId: string): Promise<boolean> {
    try {
      // Mock API call - replace with actual API call
      console.log('Following user:', userId);
      
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
      // Mock API call - replace with actual API call
      console.log('Unfollowing user:', userId);
      
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
      // Mock API call - replace with actual API call
      console.log(`Getting followers for user: ${userId}, page: ${page}, limit: ${limit}`);
      
      // Simulate API response with mock followers
      return Array(10).fill(null).map((_, i) => ({
        id: `follower-${i}`,
        name: `Follower ${i}`,
        username: `follower${i}`,
        profile_picture: `https://via.placeholder.com/150?text=F${i}`,
        verified: i % 3 === 0,
        isFollowing: i % 2 === 0
      }));
    } catch (err) {
      console.error('Failed to get followers:', err);
      return [];
    }
  }

  // Function to get following list
  async function getFollowing(userId: string, page = 1, limit = 20): Promise<IUser[]> {
    try {
      // Mock API call - replace with actual API call
      console.log(`Getting following for user: ${userId}, page: ${page}, limit: ${limit}`);
      
      // Simulate API response with mock following users
      return Array(10).fill(null).map((_, i) => ({
        id: `following-${i}`,
        name: `Following ${i}`,
        username: `following${i}`,
        profile_picture: `https://via.placeholder.com/150?text=F${i}`,
        verified: i % 4 === 0,
        isFollowing: true
      }));
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