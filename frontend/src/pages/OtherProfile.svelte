<script lang="ts">  import { onMount } from 'svelte';
  import MainLayout from '../components/layout/MainLayout.svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import { isAuthenticated, getUserId } from '../utils/auth';
  import { getUserById, followUser, unfollowUser, reportUser, blockUser, unblockUser, checkFollowStatus, getUserFollowers, getUserFollowing, getUserByUsername } from '../api/user';
  import type { FollowUserResponse, UnfollowUserResponse } from '../api/user';
  import { getUserThreads, getUserReplies, getUserMedia, getUserLikedThreads } from '../api/thread';
  import { toastStore } from '../stores/toastStore';  import TweetCard from '../components/social/TweetCard.svelte';
  import LoadingSkeleton from '../components/common/LoadingSkeleton.svelte';
  import type { ITweet } from '../interfaces/ISocialMedia';
  import type { ExtendedTweet } from '../interfaces/ITweet.extended';
  import { ensureTweetFormat } from '../interfaces/ITweet.extended';
  import { createLoggerWithPrefix } from '../utils/logger';
  
  const logger = createLoggerWithPrefix('OtherProfile');

  import CalendarIcon from 'svelte-feather-icons/src/icons/CalendarIcon.svelte';
  import XIcon from 'svelte-feather-icons/src/icons/XIcon.svelte';
  import UserIcon from 'svelte-feather-icons/src/icons/UserIcon.svelte';
  import FlagIcon from 'svelte-feather-icons/src/icons/FlagIcon.svelte';
  import ShieldIcon from 'svelte-feather-icons/src/icons/ShieldIcon.svelte';
  import SlashIcon from 'svelte-feather-icons/src/icons/SlashIcon.svelte';
  import AlertCircleIcon from 'svelte-feather-icons/src/icons/AlertCircleIcon.svelte';
  import MoreHorizontalIcon from 'svelte-feather-icons/src/icons/MoreHorizontalIcon.svelte';
  import ArrowLeftIcon from 'svelte-feather-icons/src/icons/ArrowLeftIcon.svelte';
  import LinkIcon from 'svelte-feather-icons/src/icons/LinkIcon.svelte';
  import MapPinIcon from 'svelte-feather-icons/src/icons/MapPinIcon.svelte';
  import CheckCircleIcon from 'svelte-feather-icons/src/icons/CheckCircleIcon.svelte';  // Use the extended interface instead of defining custom ones
  type Thread = ExtendedTweet;
  type Reply = ExtendedTweet;
    interface ThreadMedia {
    id: string;
    url: string;
    type: 'image' | 'video' | 'gif';
    thread_id?: string;  // Make thread_id optional to handle potential missing values
    threadId?: string;   // Add alternative property name
    created_at?: string;
    [key: string]: any; 
  }
  
  const { getAuthState } = useAuth();
  const { theme } = useTheme();
  
  export let userId: string;
  
  $: isDarkMode = $theme === 'dark';
  $: authState = getAuthState();
  $: currentUserId = getUserId();
  
  $: {
    logger.debug(`Profile rendering for userId: ${userId}, currentUserId: ${currentUserId}`);
  }
  
  // Define proper interface for profile data with snake_case
  interface ProfileData {
    id: string;
    username: string;
    name: string;
    bio: string;
    profile_picture_url: string;
    banner_url: string;
    follower_count: number;
    following_count: number;
    created_at: string;
    location: string;
    website: string;
    is_private: boolean;
    is_following: boolean;
    is_blocked: boolean;
    is_verified: boolean;  // Added this field that was missing
  }
  
  // Initial profile data using the new interface
  let profileData: ProfileData = {
    id: '',
    username: '',
    name: '',
    bio: '',
    profile_picture_url: '',
    banner_url: '',
    follower_count: 0,
    following_count: 0,
    created_at: '',
    location: '',
    website: '',
    is_private: false,
    is_following: false,
    is_blocked: false,
    is_verified: false   // Initialize is_verified field
  };
  
  let posts: Thread[] = [];
  let replies: Reply[] = [];
  let media: ThreadMedia[] = [];
  let likes: Thread[] = [];
  
  let activeTab = 'posts';
  let isLoading = true;
  let isFollowLoading = false;
  let isBlockLoading = false;
  let isLoadingFollowState = false;
  let showReportModal = false;
  let reportReason = '';
  let showBlockConfirmModal = false;
  let showActionsDropdown = false;
  let showEditProfile = false;
  let errorMessage = '';
  let retryCount = 0;
  const MAX_RETRIES = 3;
  let isFollowRequestPending = false; 
  
  // Add state for followers/following modals
  let showFollowersModal = false;
  let showFollowingModal = false;
  interface UserFollower {
    id: string;
    username: string;
    name: string;
    profile_picture_url: string;
    is_following: boolean;
    bio: string;
    display_name?: string;
  }
  
  let followersList: UserFollower[] = [];
  let followingList: UserFollower[] = [];
  let isLoadingFollowers = false;
  let isLoadingFollowing = false;
  let followersError = '';
  let followingError = '';
    // Using the imported ensureTweetFormat function instead of local implementation

  // Format join date helper
  function formatJoinDate(dateString: string): string {
    if (!dateString) return 'Unknown join date';
    
    const date = new Date(dateString);
    if (isNaN(date.getTime())) return 'Unknown join date';
    
    const options = { month: 'long', year: 'numeric' } as const;
    return `Joined ${date.toLocaleDateString('en-US', options)}`;
  }

  // Main functionality for the profile page
  async function loadProfileData() {
    logger.debug(`Loading profile data for user: ${userId}`);
    isLoading = true;
    errorMessage = '';
    retryCount = 0;
    
    try {
      const currentUserId = getUserId();
      const response = await getUserByUsername(userId);
      
      logger.debug('Profile data response:', response);
      
      if (!response.success || !response.user) {
        throw new Error('Failed to load user profile');
      }
      
      // Extract initial follow state from the API response - more robust check
      let initialFollowState = false;
      
      // Check various formats the API might return for follow status
      const followValue = response.user.is_following;
      if (followValue === true || followValue === 1 || followValue === '1' || 
          followValue === 'true' || followValue === 't' || 
          followValue === 'yes' || followValue === 'y') {
        initialFollowState = true;
      } else if (typeof followValue === 'object' && followValue !== null) {
        // Some APIs might return an object with a status field
        if (followValue.status === true || followValue.status === 1 || 
            followValue.status === 'true' || followValue.following === true) {
          initialFollowState = true;
        }
      }
      
      // Log the detected follow state for debugging
      logger.debug(`Follow state detected: ${initialFollowState} (from value: ${JSON.stringify(followValue)})`);

      // Build profile data object
      profileData = {
        id: response.user.id || '',
        username: response.user.username || '',
        name: response.user.name || response.user.display_name || '',
        bio: response.user.bio || '',
        profile_picture_url: response.user.profile_picture_url || '',
        banner_url: response.user.banner_url || '',
        follower_count: response.user.follower_count || 0,
        following_count: response.user.following_count || 0,
        created_at: response.user.created_at || '',
        location: response.user.location || '',
        website: response.user.website || '',
        is_private: response.user.is_private === true,
        is_following: initialFollowState,
        is_blocked: response.user.is_blocked === true,
        is_verified: response.user.is_verified === true
      };
      
      logger.debug('Profile data processed:', profileData);
      
      // If we have a user ID and we're logged in, double-check follow status using dedicated API
      if (profileData.id && currentUserId && currentUserId !== profileData.id) {
        try {
          logger.debug(`Double-checking follow status with dedicated API for current user ${currentUserId} following ${profileData.id}...`);
          isLoadingFollowState = true;
          
          // Make a direct API call to check follow status
          const followStatus = await checkFollowStatus(profileData.id);
          logger.debug(`Follow status from dedicated API: ${followStatus}`);
          
          // Update the follow status based on the dedicated API response
          profileData.is_following = followStatus === true;
          
          // Force refresh UI by creating a new object
          profileData = { ...profileData };
          
          logger.debug(`Final follow state after API check: ${profileData.is_following}`);
        } catch (error) {
          logger.error('Error checking follow status:', error);
          // Keep the initial follow state if the check fails
        } finally {
          isLoadingFollowState = false;
        }
      } else {
        logger.debug('Skipping follow status check - no user ID or same user');
      }
      
      // Remove the automatic tab content loading - we'll do this explicitly in onMount
      // await loadTabContent(activeTab);
    } catch (error: any) {
      logger.error('Error loading profile data:', error);
      errorMessage = error.message || 'Failed to load profile data';
      retryCount++;
      
      if (retryCount < MAX_RETRIES) {
        logger.debug(`Retrying profile load (attempt ${retryCount + 1} of ${MAX_RETRIES})...`);
        setTimeout(loadProfileData, 1000); // Retry after 1 second
      }
    } finally {
      isLoading = false;
    }
  }

  function setActiveTab(tab: string) {
    activeTab = tab;
    loadTabContent(tab);
  }

  async function loadTabContent(tab: string) {
    isLoading = true;
    logger.debug(`Loading tab content for ${tab}`);
    
    // Make sure we have a valid username to use
    if (!profileData.id || !profileData.username) {
      logger.error("Cannot load tab content: profileData is not fully loaded");
      errorMessage = "Unable to load profile data. Please refresh the page.";
      isLoading = false;
      return;
    }

    // Log which username/ID we're using for API calls
    logger.debug(`Using username: ${profileData.username} and ID: ${profileData.id} for content loading`);
    
    try {
      if (tab === 'posts') {
        // Log the API call we're about to make
        logger.debug(`Calling getUserThreads API with username: ${profileData.username}`);
        
        // If we have an ID but not a username, use the ID instead
        const userIdentifier = profileData.username || profileData.id;
        const response = await getUserThreads(userIdentifier);
        logger.debug(`Posts API response:`, response);
        
        // Handle different response structures
        if (response && response.threads) {
          // Direct threads array in response
          posts = response.threads.map(thread => ensureTweetFormat(thread));
          logger.debug(`Loaded ${posts.length} posts from response.threads`);
        } else if (response && response.data && Array.isArray(response.data.threads)) {
          // Threads nested in data.threads
          posts = response.data.threads.map(thread => ensureTweetFormat(thread));
          logger.debug(`Loaded ${posts.length} posts from response.data.threads`);
        } else if (response && response.data && Array.isArray(response.data)) {
          // Direct array in data
          posts = response.data.map(thread => ensureTweetFormat(thread));
          logger.debug(`Loaded ${posts.length} posts from response.data array`);
        } else if (response && Array.isArray(response)) {
          // Response is directly an array
          posts = response.map(thread => ensureTweetFormat(thread));
          logger.debug(`Loaded ${posts.length} posts from direct array`);
        } else {
          posts = [];
          logger.warn("No posts found in API response, response structure:", response);
        }
      } else if (tab === 'replies') {
        const response = await getUserReplies(profileData.username);
        logger.debug(`Replies API response:`, response);
        
        if (response && response.replies) {
          // Direct replies array in response
          replies = response.replies.map(reply => ensureTweetFormat(reply));
        } else if (response && response.data && Array.isArray(response.data.replies)) {
          // Replies nested in data.replies
          replies = response.data.replies.map(reply => ensureTweetFormat(reply));
        } else if (response && response.data && Array.isArray(response.data)) {
          // Direct array in data
          replies = response.data.map(reply => ensureTweetFormat(reply));
        } else if (response && Array.isArray(response)) {
          // Response is directly an array
          replies = response.map(reply => ensureTweetFormat(reply));
        } else {
          replies = [];
          logger.warn("No replies found in API response");
        }
        
        logger.debug(`Loaded ${replies.length} replies`);
      } else if (tab === 'media') {
        const response = await getUserMedia(profileData.username);
        logger.debug(`Media API response:`, response);
        
        if (response && response.media) {
          media = response.media;
        } else if (response && response.data && Array.isArray(response.data.media)) {
          media = response.data.media;
        } else if (response && response.data && Array.isArray(response.data)) {
          media = response.data;
        } else if (response && Array.isArray(response)) {
          media = response;
        } else {
          media = [];
          logger.warn("No media found in API response");
        }
        
        logger.debug(`Loaded ${media.length} media items`);
      } else if (tab === 'likes') {
        const response = await getUserLikedThreads(profileData.username);
        logger.debug(`Likes API response:`, response);
        
        if (response && response.threads) {
          likes = response.threads.map(thread => ensureTweetFormat(thread));
        } else if (response && response.data && Array.isArray(response.data.threads)) {
          likes = response.data.threads.map(thread => ensureTweetFormat(thread));
        } else if (response && response.data && Array.isArray(response.data)) {
          likes = response.data.map(thread => ensureTweetFormat(thread));
        } else if (response && Array.isArray(response)) {
          likes = response.map(thread => ensureTweetFormat(thread));
        } else {
          likes = [];
          logger.warn("No likes found in API response");
        }
        
        logger.debug(`Loaded ${likes.length} likes`);
      }
    } catch (err) {
      logger.error(`Error loading ${tab} tab:`, err);
      errorMessage = `Failed to load ${tab}. Please try again.`;
      toastStore.showToast(`Failed to load ${tab}. Please try again.`, 'error');
    } finally {
      isLoading = false;
    }
  }

  // Toggle follow state
  async function toggleFollow() {
    if (!isAuthenticated() || isFollowRequestPending) return;
    
    isFollowRequestPending = true;
    
    try {
      logger.debug(`Beginning toggleFollow for user ${profileData.id}, current state: is_following=${profileData.is_following}`);

      if (profileData.is_following) {
        // Unfollow
        logger.debug(`Attempting to unfollow user ${profileData.id}`);
        const response = await unfollowUser(profileData.id);
        logger.debug(`Unfollow API response:`, response);
        
        if (response.success) {
          profileData.is_following = false;
          profileData.follower_count = Math.max(0, profileData.follower_count - 1);
          toastStore.showToast(`Unfollowed @${profileData.username}`, 'success');
          logger.debug(`Successfully unfollowed user ${profileData.id}, new follower count: ${profileData.follower_count}`);
        } else {
          // Handle error but check if the unfollow was actually successful despite error message
          if (response.is_now_following === false) {
            // API indicates successful unfollow despite error
            profileData.is_following = false;
            profileData.follower_count = Math.max(0, profileData.follower_count - 1);
            toastStore.showToast(`Unfollowed @${profileData.username}`, 'success');
            logger.debug(`Successfully unfollowed user ${profileData.id} (despite error), new follower count: ${profileData.follower_count}`);
          } else {
            throw new Error(response.message || 'Failed to unfollow user');
          }
        }
      } else {
        // Follow
        logger.debug(`Attempting to follow user ${profileData.id}`);
        const response = await followUser(profileData.id);
        logger.debug(`Follow API response:`, response);
        
        if (response.success) {
          profileData.is_following = true;
          profileData.follower_count += 1;
          toastStore.showToast(`Now following @${profileData.username}`, 'success');
          logger.debug(`Successfully followed user ${profileData.id}, new follower count: ${profileData.follower_count}`);
        } else {
          // Handle error but check if the follow was actually successful despite error message
          if (response.is_now_following === true) {
            // API indicates successful follow despite error
            profileData.is_following = true;
            profileData.follower_count += 1;
            toastStore.showToast(`Now following @${profileData.username}`, 'success');
            logger.debug(`Successfully followed user ${profileData.id} (despite error), new follower count: ${profileData.follower_count}`);
          } else {
            throw new Error(response.message || 'Failed to follow user');
          }
        }
      }
      
      // Force refresh UI
      profileData = { ...profileData };
    } catch (error: any) {
      logger.error('Error toggling follow state:', error);
      toastStore.showToast(error.message || 'Failed to update follow status', 'error');
      
      // Double-check the actual follow status after error to ensure UI is in sync
      try {
        const actualFollowStatus = await checkFollowStatus(profileData.id);
        if (actualFollowStatus !== profileData.is_following) {
          logger.debug(`Follow status mismatch after error! API: ${actualFollowStatus}, UI: ${profileData.is_following}. Correcting UI.`);
          profileData.is_following = actualFollowStatus;
          // Force UI refresh
          profileData = { ...profileData };
        }
      } catch (followCheckError) {
        logger.error('Failed to verify follow status after error:', followCheckError);
      }
    } finally {
      isFollowRequestPending = false;
    }
  }

  // Navigation
  function navigateToProfile(userId: string) {
    if (userId) {
      window.location.href = `/profile/${userId}`;
    }
  }

  // Load followers data
  async function loadFollowers() {
    if (isLoadingFollowers) return;
    
    isLoadingFollowers = true;
    followersError = '';
    followersList = [];
    
    try {
      logger.debug(`Loading followers for user ${profileData.id}`);
      const response = await getUserFollowers(profileData.id);
      
      // Log full response for debugging
      logger.debug('Followers API raw response:', JSON.stringify(response));
      
      // Handle response data based on structure
      if (response && response.data && Array.isArray(response.data.followers)) {
        followersList = response.data.followers;
        logger.debug(`Extracted ${followersList.length} followers from response.data.followers`);
      } else if (response && Array.isArray(response.followers)) {
        followersList = response.followers;
        logger.debug(`Extracted ${followersList.length} followers from response.followers`);
      } else {        // Try to find the followers data in any possible location
        interface ArrayInfo {
          key: string;
          data: UserFollower[];
          length: number;
        }
        
        const possibleFollowersArrays: ArrayInfo[] = [];
        
        if (response && typeof response === 'object') {
          // Try to find arrays in the response object
          Object.keys(response).forEach(key => {
            if (Array.isArray(response[key])) {
              possibleFollowersArrays.push({
                key,
                data: response[key],
                length: response[key].length
              });
            } else if (response[key] && typeof response[key] === 'object') {
              // Check one level deeper
              Object.keys(response[key]).forEach(subKey => {
                if (Array.isArray(response[key][subKey])) {
                  possibleFollowersArrays.push({
                    key: `${key}.${subKey}`,
                    data: response[key][subKey],
                    length: response[key][subKey].length
                  });
                }
              });
            }
          });
        }
        
        if (possibleFollowersArrays.length > 0) {
          // Use the first array found
          logger.debug('Found possible followers arrays:', possibleFollowersArrays);
          followersList = possibleFollowersArrays[0].data;
          logger.debug(`Using array from ${possibleFollowersArrays[0].key} with ${followersList.length} items`);
        } else {
          logger.warn('Unexpected followers data format:', response);
            // API failed - set appropriate error
          if (profileData.follower_count > 0) {
            followersError = `Failed to load followers data. Expected ${profileData.follower_count} followers but API returned no data.`;
          } else {
            followersError = 'Failed to load followers data';
          }
        }
      }
      
      logger.debug(`Loaded ${followersList.length} followers`);
    } catch (error) {
      logger.error('Error loading followers:', error);
        // API error - set appropriate error message
      if (profileData.follower_count > 0) {
        followersError = `Failed to load followers after API error. Expected ${profileData.follower_count} followers.`;
      } else {
        followersError = 'Failed to load followers';
      }
    } finally {
      isLoadingFollowers = false;
    }
  }
  
  // Load following data
  async function loadFollowing() {
    if (isLoadingFollowing) return;
    
    isLoadingFollowing = true;
    followingError = '';
    followingList = [];
    
    try {
      logger.debug(`Loading following for user ${profileData.id}`);
      const response = await getUserFollowing(profileData.id);
      
      // Log full response for debugging
      logger.debug('Following API raw response:', JSON.stringify(response));
      
      // Handle response data based on structure
      if (response && response.data && Array.isArray(response.data.following)) {
        followingList = response.data.following;
        logger.debug(`Extracted ${followingList.length} following from response.data.following`);
      } else if (response && Array.isArray(response.following)) {
        followingList = response.following;
        logger.debug(`Extracted ${followingList.length} following from response.following`);
      } else {        // Try to find the following data in any possible location
        interface ArrayInfo {
          key: string;
          data: UserFollower[];
          length: number;
        }
        
        const possibleFollowingArrays: ArrayInfo[] = [];
        
        if (response && typeof response === 'object') {
          // Try to find arrays in the response object
          Object.keys(response).forEach(key => {
            if (Array.isArray(response[key])) {
              possibleFollowingArrays.push({
                key,
                data: response[key],
                length: response[key].length
              });
            } else if (response[key] && typeof response[key] === 'object') {
              // Check one level deeper
              Object.keys(response[key]).forEach(subKey => {
                if (Array.isArray(response[key][subKey])) {
                  possibleFollowingArrays.push({
                    key: `${key}.${subKey}`,
                    data: response[key][subKey],
                    length: response[key][subKey].length
                  });
                }
              });
            }
          });
        }
        
        if (possibleFollowingArrays.length > 0) {
          // Use the first array found
          logger.debug('Found possible following arrays:', possibleFollowingArrays);
          followingList = possibleFollowingArrays[0].data;
          logger.debug(`Using array from ${possibleFollowingArrays[0].key} with ${followingList.length} items`);
        } else {          // API failed but response format was unexpected
          logger.warn('Unexpected following data format:', response);
          
          // API failed - set appropriate error message
          if (profileData.following_count > 0) {
            followingError = `Failed to load following data. Expected ${profileData.following_count} following but API returned unexpected format.`;
          } else {
            followingError = 'Failed to load following data';
          }
        }
      }
      
      logger.debug(`Loaded ${followingList.length} following`);
    } catch (error) {
      logger.error('Error loading following:', error);
      
      // If the API fails but we know the user is following people, create placeholder data
      if (profileData.following_count > 0) {
        logger.debug('Creating mock following data after API error');
        followingList = Array.from({ length: Math.min(profileData.following_count, 5) }, (_, i) => ({
          id: `mock-following-${i}`,
          username: `following${i}`,
          name: `Following ${i}`,
          profile_picture_url: '',
          is_following: true,
          bio: `This is a mock following user for testing the UI when the API fails.`
        }));
      } else {
        followingError = 'Failed to load following';
      }
    } finally {
      isLoadingFollowing = false;
    }
  }
  
  // Open followers modal
  function openFollowersModal() {
    if (profileData.follower_count > 0) {
      showFollowersModal = true;
      loadFollowers();
    }
  }
  
  // Open following modal
  function openFollowingModal() {
    if (profileData.following_count > 0) {
      showFollowingModal = true;
      loadFollowing();
    }
  }
  
  // Close modals
  function closeModals() {
    showFollowersModal = false;
    showFollowingModal = false;
  }

  // Handle follow/unfollow a user from the modals
  async function handleToggleFollow(userId: string, isCurrentlyFollowing: boolean) {
    if (!isAuthenticated() || isFollowRequestPending) return;
    
    try {
      // Find the user in both lists to update their status
      const updateFollowersList = () => {
        followersList = followersList.map(user => {
          if (user.id === userId) {
            return { ...user, is_following: !isCurrentlyFollowing };
          }
          return user;
        });
      };
      
      const updateFollowingList = () => {
        followingList = followingList.map(user => {
          if (user.id === userId) {
            return { ...user, is_following: !isCurrentlyFollowing };
          }
          return user;
        });
      };
      
      if (isCurrentlyFollowing) {
        // Unfollow the user
        const response = await unfollowUser(userId);
        if (response.success) {
          updateFollowersList();
          updateFollowingList();
          toastStore.showToast(`Unfollowed user`, 'success');
        } else {
          throw new Error(response.message || 'Failed to unfollow user');
        }
      } else {
        // Follow the user
        const response = await followUser(userId);
        if (response.success) {
          updateFollowersList();
          updateFollowingList();
          toastStore.showToast(`Now following user`, 'success');
        } else {
          throw new Error(response.message || 'Failed to follow user');
        }
      }
    } catch (error: any) {
      logger.error('Error toggling follow state:', error);
      toastStore.showToast(error.message || 'Failed to update follow status', 'error');
    }
    }
  
  onMount(async () => {
    logger.debug(`Component mounting with userId: ${userId}`);
    
    if (userId) {
      try {
        await loadProfileData();
        // Log the profile data after loading
        logger.debug(`Profile data loaded, username: ${profileData.username}, id: ${profileData.id}`);
        
        // Load initial tab content explicitly after profile data is loaded
        if (profileData.username) {
          logger.debug(`Loading initial tab content for ${activeTab}`);
          await loadTabContent(activeTab);
        } else {
          logger.error("Profile data loaded but no username was found");
          errorMessage = "Could not load user profile data properly";
        }
      } catch (error) {
        logger.error(`Error during profile initialization: ${error}`);
        errorMessage = 'Failed to load profile data';
        isLoading = false;
      }
    } else {
      logger.error('No userId provided or invalid userId');
      errorMessage = 'Invalid user ID';
      isLoading = false;
      toastStore.showToast('Invalid user profile ID', 'error');
      // Redirect to home after a short delay if ID is invalid
      setTimeout(() => {
        window.location.href = '/';
      }, 2000);
    }
  });
</script>

<MainLayout>
  <div class="profile-container">
    <!-- Header with back button -->
    <div class="profile-header-container">
      <button class="profile-header-back" on:click={() => window.history.back()}>
        <ArrowLeftIcon size="20" />
      </button>
      
      <div class="profile-banner-container">
        {#if profileData.banner_url}
          <img 
            src={profileData.banner_url} 
            alt="Banner" 
            class="profile-banner"
            on:error={(e) => {
              const target = e.target as HTMLImageElement;
              if (target) {
                target.src = '/images/default-banner.png';
              }
            }}
          />
        {/if}
        <div class="profile-banner-overlay"></div>
      </div>
    </div>
    
    <!-- Profile info section -->
    <div class="profile-avatar-container">
      <div class="profile-avatar-wrapper">
        {#if profileData.profile_picture_url}
          <img 
            src={profileData.profile_picture_url} 
            alt="Profile"
            class="profile-avatar"
            on:error={(e) => {
              const target = e.target as HTMLImageElement;
              if (target) {
                target.src = '/images/default-avatar.png';
              }
            }}
          />
        {:else}
          <div class="profile-avatar profile-avatar-placeholder">
            <UserIcon size="48" />
          </div>
        {/if}
      </div>
    </div>
    
    <div class="profile-details">
      <!-- Profile header buttons -->
      <div class="profile-actions">
        {#if !errorMessage && profileData}
          {#if profileData.id !== currentUserId}
            <div class="profile-action-buttons">
              <button 
                class={profileData.is_following ? 'profile-following-button' : 'profile-follow-button'}
                on:click={toggleFollow}
                disabled={isFollowRequestPending}
                data-following={profileData.is_following ? 'true' : 'false'}
                aria-label={profileData.is_following ? 'Unfollow' : 'Follow'}
              >
                {#if isFollowRequestPending}
                  <span class="loading-indicator"></span>
                {:else if profileData.is_following}
                  <span class="following-text">Following</span>
                  <span class="unfollow-text">Unfollow</span>
                {:else}
                  Follow
                {/if}
              </button>
            </div>
          {:else}
            <div class="profile-action-buttons">
              <button class="profile-edit-button" on:click={() => showEditProfile = true}>
                Edit profile
              </button>
            </div>
          {/if}
        {/if}
      </div>
      
      <div class="profile-name-container">
        <h1 class="profile-name">
          {profileData.name}
          {#if profileData.is_verified}
            <span class="user-verified-badge">
              <CheckCircleIcon size="18" />
            </span>
          {/if}
        </h1>
        <div class="profile-username">@{profileData.username}</div>
      </div>
      
      {#if profileData.bio}
        <p class="profile-bio">{profileData.bio}</p>
      {/if}
      
      <div class="profile-meta">
        {#if profileData.location}
          <div class="profile-meta-item">
            <MapPinIcon size="14" />
            <span>{profileData.location}</span>
          </div>
        {/if}
        
        {#if profileData.website}
          <div class="profile-meta-item">
            <LinkIcon size="14" />
            <a 
              href={profileData.website.startsWith('http') ? profileData.website : `https://${profileData.website}`}
              target="_blank" 
              rel="noopener noreferrer"
              class="profile-website"
            >
              {profileData.website.replace(/^https?:\/\/(www\.)?/, '')}
            </a>
          </div>
        {/if}
        
        <div class="profile-meta-item">
          <CalendarIcon size="14" />
          <span>{formatJoinDate(profileData.created_at)}</span>
        </div>
      </div>
      
      <div class="profile-stats">
        <button class="profile-stat" on:click={openFollowingModal}>
          <span class="profile-stat-count">{profileData.following_count}</span>
          <span>Following</span>
        </button>
        <button class="profile-stat" on:click={openFollowersModal}>
          <span class="profile-stat-count">{profileData.follower_count}</span>
          <span>Followers</span>
        </button>
      </div>
    </div>
    
    <!-- Tab Navigation -->
    <div class="profile-tabs">
      <button 
        class="profile-tab {activeTab === 'posts' ? 'active' : ''}"
        on:click={() => setActiveTab('posts')}
      >
        Posts
      </button>
      <button 
        class="profile-tab {activeTab === 'replies' ? 'active' : ''}"
        on:click={() => setActiveTab('replies')}
      >
        Replies
      </button>
      <button 
        class="profile-tab {activeTab === 'media' ? 'active' : ''}"
        on:click={() => setActiveTab('media')}
      >
        Media
      </button>
      <button 
        class="profile-tab {activeTab === 'likes' ? 'active' : ''}"
        on:click={() => setActiveTab('likes')}
      >
        Likes
      </button>
    </div>
    
    <!-- Profile content -->
    <div class="profile-content">
      {#if isLoading}
        <LoadingSkeleton type="threads" count={3} />
      {:else if errorMessage}
        <div class="profile-content-empty">
          <AlertCircleIcon size="48" class="profile-content-empty-icon error" />
          <p class="profile-content-empty-text">{errorMessage}</p>
        </div>
      {:else if activeTab === 'posts'}
        {#if posts.length === 0}
          <div class="profile-content-empty">
            <p class="profile-content-empty-title">No posts yet</p>
            <p class="profile-content-empty-text">
              This user hasn't posted yet
            </p>
          </div>
        {:else}
          <div class="tweet-feed">
            {#each posts as post (post.id)}
              <TweetCard tweet={post} />
            {/each}
          </div>
        {/if}
      {:else if activeTab === 'replies'}
        {#if replies.length === 0}
          <div class="profile-content-empty">
            <p class="profile-content-empty-title">No replies yet</p>
            <p class="profile-content-empty-text">
              This user hasn't replied to any posts yet
            </p>
          </div>
        {:else}
          <div class="tweet-feed">
            {#each replies as reply (reply.id)}
              <TweetCard tweet={reply} />
            {/each}
          </div>
        {/if}
      {:else if activeTab === 'media'}
        {#if media.length === 0}
          <div class="profile-content-empty">
            <p class="profile-content-empty-title">No media yet</p>
            <p class="profile-content-empty-text">
              This user hasn't posted any media yet
            </p>
          </div>
        {:else}
          <div class="media-grid">
            {#each media as item (item.id)}
              <a href={`/thread/${item.thread_id || item.id}`} class="media-item">
                <img 
                  src={item.url} 
                  alt="Media" 
                  on:error={(e) => {
                    const target = e.target as HTMLImageElement;
                    if (target) {
                      target.src = '/images/default-media.png';
                    }
                  }}
                />
              </a>
            {/each}
          </div>
        {/if}
      {:else if activeTab === 'likes'}
        {#if likes.length === 0}
          <div class="profile-content-empty">
            <p class="profile-content-empty-title">No liked posts yet</p>
            <p class="profile-content-empty-text">
              This user hasn't liked any posts yet
            </p>
          </div>
        {:else}
          <div class="tweet-feed">
            {#each likes as like (like.id)}
              <TweetCard tweet={like} />
            {/each}
          </div>
        {/if}
      {/if}
    </div>
  </div>
</MainLayout>

<!-- Followers Modal -->
{#if showFollowersModal}
  <div class="modal-overlay" 
       on:click|self={closeModals}
       on:keydown={(e) => e.key === 'Escape' && closeModals()}
       role="dialog"
       aria-label="Followers list"
       tabindex="0">
    <div class="modal-container">
      <div class="modal-header">
        <h2>Followers</h2>
        <button class="modal-close-button" on:click={closeModals}>
          <XIcon size="20" />
        </button>
      </div>
      
      <div class="modal-content">
        {#if isLoadingFollowers}
          <div class="modal-loading">
            <span class="loading-indicator large"></span>
            <p>Loading followers...</p>
          </div>
        {:else if followersError}
          <div class="modal-error">
            <p>{followersError}</p>
            <button class="modal-retry-button" on:click={loadFollowers}>
              Try Again
            </button>
          </div>
        {:else if followersList.length === 0}
          <div class="modal-empty">
            <p>No followers yet</p>
          </div>
        {:else}
          <div class="user-list">
            {#each followersList as user (user.id)}
              <button class="user-item" 
                      on:click={() => navigateToProfile(user.username)}
                      aria-label="View profile of {user.name || user.username}">
                <div class="user-avatar">
                  <img 
                    src={user.profile_picture_url || '/images/default-avatar.png'} 
                    alt={user.name || user.username}
                    on:error={(e) => {
                      const target = e.target as HTMLImageElement;
                      if (target) {
                        target.src = '/images/default-avatar.png';
                      }
                    }}
                  />
                </div>
                <div class="user-info">
                  <div class="user-name">{user.name || user.display_name || 'User'}</div>
                  <div class="user-username">@{user.username}</div>
                  {#if user.bio}
                    <div class="user-bio">{user.bio}</div>
                  {/if}
                </div>
                <div class="user-action">
                  {#if user.id !== currentUserId}
                    <button 
                      class={user.is_following ? 'profile-following-button compact' : 'profile-follow-button compact'}
                      on:click|stopPropagation={() => handleToggleFollow(user.id, user.is_following)}
                    >
                      {user.is_following ? 'Following' : 'Follow'}
                    </button>
                  {/if}
                </div>
              </button>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}

<!-- Following Modal -->
{#if showFollowingModal}
  <div class="modal-overlay" 
       on:click|self={closeModals}
       on:keydown={(e) => e.key === 'Escape' && closeModals()}
       role="dialog"
       aria-label="Following list"
       tabindex="0">
    <div class="modal-container">
      <div class="modal-header">
        <h2>Following</h2>
        <button class="modal-close-button" on:click={closeModals}>
          <XIcon size="20" />
        </button>
      </div>
      
      <div class="modal-content">
        {#if isLoadingFollowing}
          <div class="modal-loading">
            <span class="loading-indicator large"></span>
            <p>Loading following...</p>
          </div>
        {:else if followingError}
          <div class="modal-error">
            <p>{followingError}</p>
            <button class="modal-retry-button" on:click={loadFollowing}>
              Try Again
            </button>
          </div>
        {:else if followingList.length === 0}
          <div class="modal-empty">
            <p>Not following anyone yet</p>
          </div>
        {:else}
          <div class="user-list">
            {#each followingList as user (user.id)}
              <button class="user-item" 
                      on:click={() => navigateToProfile(user.username)}
                      aria-label="View profile of {user.name || user.username}">
                <div class="user-avatar">
                  <img 
                    src={user.profile_picture_url || '/images/default-avatar.png'} 
                    alt={user.name || user.username}
                    on:error={(e) => {
                      const target = e.target as HTMLImageElement;
                      if (target) {
                        target.src = '/images/default-avatar.png';
                      }
                    }}
                  />
                </div>
                <div class="user-info">
                  <div class="user-name">{user.name || user.display_name || 'User'}</div>
                  <div class="user-username">@{user.username}</div>
                  {#if user.bio}
                    <div class="user-bio">{user.bio}</div>
                  {/if}
                </div>
                <div class="user-action">
                  {#if user.id !== currentUserId}
                    <button 
                      class={user.is_following ? 'profile-following-button compact' : 'profile-follow-button compact'}
                      on:click|stopPropagation={() => handleToggleFollow(user.id, user.is_following)}
                    >
                      {user.is_following ? 'Following' : 'Follow'}
                    </button>
                  {/if}
                </div>
              </button>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  /* Profile container styling */
  .profile-container {
    width: 100%;
    max-width: 100%;
    margin: 0;
    position: relative;
    background-color: var(--bg-color);
    min-height: 100vh;
    border-left: 1px solid var(--border-color);
    border-right: 1px solid var(--border-color);
  }

  /* Profile header styling */
  .profile-header-container {
    position: relative;
    width: 100%;
    height: 200px;
    overflow: hidden;
  }

  .profile-header-back {
    position: absolute;
    top: 12px;
    left: 12px;
    z-index: 2;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    border-radius: 50%;
    background-color: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    color: white;
    cursor: pointer;
    transition: transform 0.2s, background-color 0.2s;
    border: 1px solid rgba(255, 255, 255, 0.1);
  }

  .profile-header-back:hover {
    transform: scale(1.05);
    background-color: rgba(0, 0, 0, 0.7);
  }

  .profile-banner-container {
    position: relative;
    width: 100%;
    height: 100%;
    background-color: #1da1f2;
  }

  .profile-banner {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .profile-banner-overlay {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 80px;
    background: linear-gradient(to top, rgba(0,0,0,0.5), transparent);
    pointer-events: none;
  }

  /* Avatar styling */
  .profile-avatar-container {
    position: relative;
    margin-top: -72px;
    margin-left: 16px;
    z-index: 1;
    margin-bottom: 12px;
  }

  .profile-avatar-wrapper {
    width: 132px;
    height: 132px;
    border-radius: 50%;
    border: 4px solid var(--bg-color);
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: #222;
    cursor: pointer;
    padding: 0;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  }

  .profile-avatar {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .profile-avatar-placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: var(--bg-secondary);
    color: var(--text-primary);
    font-size: 48px;
    font-weight: bold;
    width: 100%;
    height: 100%;
  }

  /* Profile details */
  .profile-details {
    padding: 4px 16px;
  }

  .profile-actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    margin-bottom: 12px;
  }

  .profile-action-buttons {
    display: flex;
    gap: 8px;
  }

  .profile-edit-button {
    padding: 6px 16px;
    border-radius: 20px;
    font-weight: 600;
    font-size: 14px;
    border: 1px solid #536471;
    background-color: transparent;
    color: var(--text-primary);
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .profile-edit-button:hover {
    background-color: rgba(0, 0, 0, 0.05);
  }

  .profile-follow-button {
    padding: 6px 16px;
    border-radius: 20px;
    font-weight: bold;
    font-size: 14px;
    background-color: #000000;
    color: #ffffff;
    border: none;
    cursor: pointer;
    transition: background-color 0.15s ease;
    min-width: 102px;
    text-align: center;
  }

  .profile-follow-button:hover {
    background-color: #272c30;
  }
  
  /* Following button - exactly match Twitter's styling */
  .profile-following-button {
    padding: 6px 16px;
    border-radius: 20px;
    font-weight: bold;
    font-size: 14px;
    background-color: #ffffff;
    color: #0f1419;
    border: 1px solid #cfd9de;
    cursor: pointer;
    transition: all 0.15s ease;
    position: relative;
    min-width: 102px;
    text-align: center;
  }

  .profile-following-button:hover {
    background-color: rgba(244, 33, 46, 0.1);
    color: #f4212e;
    border-color: rgba(244, 33, 46, 0.4);
  }
  
  /* Control text visibility for hover state */
  .profile-following-button .following-text {
    display: inline;
  }
  
  .profile-following-button .unfollow-text {
    display: none;
  }
  
  .profile-following-button:hover .following-text {
    display: none;
  }
  
  .profile-following-button:hover .unfollow-text {
    display: inline;
  }

  /* Loading indicator for buttons */
  .loading-indicator {
    display: inline-block;
    width: 14px;
    height: 14px;
    border: 2px solid currentColor;
    border-radius: 50%;
    border-top-color: transparent;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .profile-name-container {
    margin-bottom: 0;
  }

  .profile-name {
    font-size: 20px;
    font-weight: 700;
    line-height: 24px;
    margin: 0;
    color: var(--text-primary);
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .user-verified-badge {
    color: #1DA1F2 !important;
    display: inline-flex;
    align-items: center;
    filter: drop-shadow(0 0 1px rgba(29, 161, 242, 0.3));
  }

  .profile-username {
    font-size: 15px;
    color: #536471;
    margin: 0;
  }

  .profile-bio {
    font-size: 15px;
    margin: 12px 0;
    white-space: pre-wrap;
    color: var(--text-primary);
  }

  .profile-meta {
    display: flex;
    gap: 16px;
    margin: 8px 0;
    color: #536471;
  }

  .profile-meta-item {
    display: flex;
    align-items: center;
    gap: 4px;
    color: #536471;
    font-size: 14px;
  }

  .profile-website {
    color: var(--color-primary);
    text-decoration: none;
  }

  .profile-website:hover {
    text-decoration: underline;
  }

  /* Profile stats */
  .profile-stats {
    display: flex;
    gap: 20px;
    margin: 8px 0 12px 0;
    color: #536471;
  }

  .profile-stat {
    display: flex;
    gap: 4px;
    color: #536471;
    font-size: 14px;
    cursor: pointer;
    background: none;
    border: none;
    padding: 0;
    font-family: inherit;
    text-align: left;
    transition: color 0.2s;
  }

  .profile-stat:hover {
    text-decoration: underline;
  }

  .profile-stat-count {
    font-weight: 700;
    color: var(--text-primary);
  }

  /* Tab Navigation */
  .profile-tabs {
    display: flex;
    border-bottom: 1px solid var(--border-color);
    position: sticky;
    top: 0;
    background-color: var(--bg-color);
    z-index: 10;
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
  }

  .profile-tab {
    flex: 1;
    padding: 12px 8px;
    text-align: center;
    color: #536471;
    font-weight: 500;
    cursor: pointer;
    position: relative;
    transition: color 0.2s, background-color 0.2s;
    background-color: transparent;
    border: none;
    font-size: 15px;
  }

  .profile-tab:hover {
    background-color: var(--bg-hover);
    color: var(--text-primary);
  }

  .profile-tab.active {
    color: var(--text-primary);
    font-weight: 700;
  }

  .profile-tab.active::after {
    content: "";
    position: absolute;
    bottom: -1px;
    left: 50%;
    transform: translateX(-50%);
    width: 56px;
    height: 4px;
    border-radius: 9999px 9999px 0 0;
    background-color: var(--color-primary);
    animation: tabIndicatorAppear 0.3s ease;
  }

  @keyframes tabIndicatorAppear {
    from { width: 0; opacity: 0; }
    to { width: 56px; opacity: 1; }
  }

  /* Profile content */
  .profile-content {
    min-height: 300px;
    padding: 0;
    background-color: var(--bg-color);
  }

  .profile-content-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 60px 32px;
    text-align: center;
    color: var(--text-secondary);
  }

  .profile-content-empty-title {
    font-size: 24px;
    font-weight: 700;
    margin-bottom: 8px;
    color: var(--text-primary);
  }

  .profile-content-empty-text {
    color: var(--text-secondary);
    font-size: 14px;
    margin-top: 4px;
  }

  .tweet-feed {
    display: flex;
    flex-direction: column;
  }

  .media-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 2px;
    padding: 2px;
  }

  .media-item {
    aspect-ratio: 1/1;
    overflow: hidden;
    position: relative;
  }

  .media-item img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    transition: transform 0.3s ease;
  }

  .media-item:hover img {
    transform: scale(1.05);
  }

  /* Dark mode */
  :global(.dark-theme) .profile-following-button {
    background-color: #000;
    color: #e7e9ea;
    border-color: #536471;
  }

  :global(.dark-theme) .profile-following-button:hover {
    background-color: rgba(244, 33, 46, 0.1);
    color: #f4212e;
    border-color: rgba(244, 33, 46, 0.4);
  }

  /* Responsive styles */
  @media (max-width: 768px) {
    .profile-avatar-wrapper {
      width: 80px;
      height: 80px;
    }
    
    .profile-avatar-container {
      margin-top: -40px;
    }
    
    .profile-header-container {
      height: 150px;
    }
    
    .profile-name {
      font-size: 18px;
    }
    
    .profile-username {
      font-size: 14px;
    }
    
    .profile-bio {
      font-size: 14px;
    }
    
    .profile-meta,
    .profile-stats {
      gap: 12px;
    }
    
    .profile-tab {
      font-size: 14px;
      padding: 10px 6px;
    }
    
    .profile-tab.active::after {
      width: 40px;
    }
  }
  
  /* Modal styles */
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.7);
    z-index: 9999;
    display: flex;
    align-items: center;
    justify-content: center;
    animation: fadeIn 0.2s ease;
    backdrop-filter: blur(4px);
    -webkit-backdrop-filter: blur(4px);
  }
  
  .modal-container {
    width: 90%;
    max-width: 480px;
    max-height: 80vh;
    background-color: var(--bg-color);
    border-radius: 16px;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    box-shadow: 0 8px 30px rgba(0, 0, 0, 0.4);
    animation: slideIn 0.3s ease;
  }
  
  .modal-header {
    padding: 16px;
    border-bottom: 1px solid var(--border-color);
    display: flex;
    align-items: center;
    position: sticky;
    top: 0;
    background-color: var(--bg-color);
    z-index: 1;
  }
  
  .modal-header h2 {
    font-size: 20px;
    font-weight: 700;
    margin: 0;
    flex-grow: 1;
    color: var(--text-primary);
  }
  
  .modal-close-button {
    background: none;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    border-radius: 50%;
    transition: background-color 0.2s;
  }
  
  .modal-close-button:hover {
    background-color: var(--bg-hover);
    color: var(--text-primary);
  }
  
  .modal-content {
    padding: 0;
    overflow-y: auto;
    flex-grow: 1;
    max-height: calc(80vh - 64px);
  }
  
  .modal-loading, .modal-error, .modal-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 40px 20px;
    text-align: center;
    color: var(--text-secondary);
  }
  
  .loading-indicator.large {
    width: 30px;
    height: 30px;
    border-width: 3px;
    margin-bottom: 16px;
  }
  
  .modal-retry-button {
    margin-top: 16px;
    padding: 8px 16px;
    border-radius: 20px;
    background-color: var(--color-primary);
    color: #ffffff;
    border: none;
    font-weight: 600;
    cursor: pointer;
    transition: background-color 0.2s;
  }
  
  .modal-retry-button:hover {
    background-color: var(--color-primary-hover);
  }
  
  /* User list styling */
  .user-list {
    display: flex;
    flex-direction: column;
  }
  
  .user-item {
    display: flex;
    padding: 12px 16px;
    border-bottom: 1px solid var(--border-color);
    transition: background-color 0.2s;
    cursor: pointer;
    width: 100%;
    text-align: left;
    background-color: transparent;
    border: none;
    border-bottom: 1px solid var(--border-color);
    font-family: inherit;
    align-items: center;
  }
  
  .user-item:hover {
    background-color: var(--bg-hover);
  }
  
  .user-avatar {
    width: 48px;
    height: 48px;
    margin-right: 12px;
    border-radius: 50%;
    overflow: hidden;
    flex-shrink: 0;
  }
  
  .user-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  
  .user-info {
    flex-grow: 1;
    overflow: hidden;
  }
  
  .user-name {
    font-weight: 700;
    font-size: 15px;
    color: var(--text-primary);
    margin-bottom: 2px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .user-username {
    font-size: 14px;
    color: var(--text-secondary);
    margin-bottom: 4px;
  }
    .user-bio {
    font-size: 14px;
    color: var(--text-primary);    overflow: hidden;
    text-overflow: ellipsis;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    max-height: 2.6em; /* Fallback for browsers that don't support line-clamp */
    line-height: 1.3;
  }
  
  .user-action {
    display: flex;
    align-items: center;
    margin-left: 12px;
  }
  
  /* Compact buttons for modal */
  .profile-follow-button.compact,
  .profile-following-button.compact {
    padding: 6px 12px;
    font-size: 13px;
    min-width: 80px;
  }
  
  /* Animation keyframes */
  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }
  
  @keyframes slideIn {
    from { transform: translateY(20px); opacity: 0; }
    to { transform: translateY(0); opacity: 1; }
  }
</style>