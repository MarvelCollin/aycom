import { writable, get } from "svelte/store";
import type { IUser, IUserProfile, IUserUpdateRequest } from "../interfaces/IUser";
import * as userApi from "../api/user";

export function useProfile() {
  const profile = writable<IUserProfile | null>(null);
  const loading = writable(false);
  const error = writable<string | null>(null);

  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:8083/api/v1";

  async function fetchProfile(): Promise<boolean> {
    loading.set(true);
    error.set(null);

    try {
      // Use the getProfile API function
      const data = await userApi.getProfile();

      if (data && data.user) {
        const userData = data.user;

        const userProfile: IUserProfile = {
          id: userData.id,
          name: userData.name || userData.display_name,
          username: userData.username,
          email: userData.email,
          bio: userData.bio || "",
          location: userData.location || "",
          website: userData.website || "",
          profile_picture: userData.profile_picture_url || "",
          banner: userData.banner_url || "",
          verified: userData.verified || false,
          followers_count: userData.followers_count || 0,
          following_count: userData.following_count || 0,
          tweets_count: userData.tweets_count || 0,
          joined_date: userData.created_at || new Date().toISOString(),
          birthday: userData.birthday || ""
        };

        profile.set(userProfile);
        return true;
      } else {
        throw new Error("Invalid profile data received from API");
      }
    } catch (err) {
      console.error("Failed to fetch profile:", err);
      error.set(err instanceof Error ? err.message : "Failed to load profile data");
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
      // Use the updateProfile API function
      const responseData = await userApi.updateProfile(data);

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
        message: responseData.message || "Profile updated successfully"
      };
    } catch (err) {
      console.error("Failed to update profile:", err);
      error.set(err instanceof Error ? err.message : "Failed to update profile");
      return {
        success: false,
        message: err instanceof Error ? err.message : "Failed to update profile"
      };
    } finally {
      loading.set(false);
    }
  }

  // Function to check if a username is available
  async function checkUsernameAvailability(username: string): Promise<boolean> {
    try {
      return await userApi.checkUsernameAvailability(username);
    } catch (err) {
      console.error("Failed to check username availability:", err);
      return false;
    }
  }

  // Function to follow a user
  async function followUser(userId: string): Promise<boolean> {
    try {
      const success = await userApi.followUser(userId);

      // Update followers count in the profile
      if (success) {
        const currentProfile = get(profile);
        if (currentProfile) {
          profile.set({
            ...currentProfile,
            followers_count: currentProfile.followers_count + 1
          });
        }
      }

      return success;
    } catch (err) {
      console.error("Failed to follow user:", err);
      return false;
    }
  }

  // Function to unfollow a user
  async function unfollowUser(userId: string): Promise<boolean> {
    try {
      const success = await userApi.unfollowUser(userId);

      // Update followers count in the profile
      if (success) {
        const currentProfile = get(profile);
        if (currentProfile && currentProfile.followers_count > 0) {
          profile.set({
            ...currentProfile,
            followers_count: currentProfile.followers_count - 1
          });
        }
      }

      return success;
    } catch (err) {
      console.error("Failed to unfollow user:", err);
      return false;
    }
  }

  // Function to get followers list
  async function getFollowers(userId: string, page = 1, limit = 20): Promise<IUser[]> {
    try {
      return await userApi.getFollowers(userId, page, limit);
    } catch (err) {
      console.error("Failed to get followers:", err);
      return [];
    }
  }

  // Function to get following list
  async function getFollowing(userId: string, page = 1, limit = 20): Promise<IUser[]> {
    try {
      return await userApi.getFollowing(userId, page, limit);
    } catch (err) {
      console.error("Failed to get following list:", err);
      return [];
    }
  }

  return {
    profile: { subscribe: profile.subscribe },
    loading: { subscribe: loading.subscribe },
    error: { subscribe: error.subscribe },
    fetchProfile,
    updateProfile,
    checkUsernameAvailability,
    followUser,
    unfollowUser,
    getFollowers,
    getFollowing
  };
}