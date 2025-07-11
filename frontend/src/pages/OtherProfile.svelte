<script lang="ts">  import { onMount } from "svelte";
  import MainLayout from "../components/layout/MainLayout.svelte";
  import { useAuth } from "../hooks/useAuth";
  import { useTheme } from "../hooks/useTheme";
  import { isAuthenticated, getUserId } from "../utils/auth";
  import { getUserById, followUser, unfollowUser, reportUser, blockUser, unblockUser, checkFollowStatus, getUserFollowers, getUserFollowing, getUserByUsername } from "../api/user";
  import type { FollowUserResponse, UnfollowUserResponse } from "../api/user";
  import { getUserThreads, getUserReplies, getUserMedia, getUserLikedThreads } from "../api/thread";
  import { toastStore } from "../stores/toastStore";  import TweetCard from "../components/social/TweetCard.svelte";
  import LoadingSkeleton from "../components/common/LoadingSkeleton.svelte";
  import type { ITweet } from "../interfaces/ISocialMedia";
  import type { ExtendedTweet } from "../interfaces/ITweet.extended";
  import { ensureTweetFormat } from "../interfaces/ITweet.extended";
  import { createLoggerWithPrefix } from "../utils/logger";

  const logger = createLoggerWithPrefix("OtherProfile");

  import CalendarIcon from "svelte-feather-icons/src/icons/CalendarIcon.svelte";
  import XIcon from "svelte-feather-icons/src/icons/XIcon.svelte";
  import UserIcon from "svelte-feather-icons/src/icons/UserIcon.svelte";
  import FlagIcon from "svelte-feather-icons/src/icons/FlagIcon.svelte";
  import ShieldIcon from "svelte-feather-icons/src/icons/ShieldIcon.svelte";
  import SlashIcon from "svelte-feather-icons/src/icons/SlashIcon.svelte";
  import AlertCircleIcon from "svelte-feather-icons/src/icons/AlertCircleIcon.svelte";
  import MoreHorizontalIcon from "svelte-feather-icons/src/icons/MoreHorizontalIcon.svelte";
  import ArrowLeftIcon from "svelte-feather-icons/src/icons/ArrowLeftIcon.svelte";
  import LinkIcon from "svelte-feather-icons/src/icons/LinkIcon.svelte";
  import MapPinIcon from "svelte-feather-icons/src/icons/MapPinIcon.svelte";
  import CheckCircleIcon from "svelte-feather-icons/src/icons/CheckCircleIcon.svelte";  
  type Thread = ExtendedTweet;
  type Reply = ExtendedTweet;
    interface ThreadMedia {
    id: string;
    url: string;
    type: "image" | "video" | "gif";
    thread_id?: string;  
    threadId?: string;   
    created_at?: string;
    [key: string]: any;
  }

  const { getAuthState } = useAuth();
  const { theme } = useTheme();

  export let userId: string;

  $: isDarkMode = $theme === "dark";
  $: authState = getAuthState();
  $: currentUserId = getUserId();

  $: {
    logger.debug(`Profile rendering for userId: ${userId}, currentUserId: ${currentUserId}`);
  }

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
    is_verified: boolean;  
  }

  let profileData: ProfileData = {
    id: "",
    username: "",
    name: "",
    bio: "",
    profile_picture_url: "",
    banner_url: "",
    follower_count: 0,
    following_count: 0,
    created_at: "",
    location: "",
    website: "",
    is_private: false,
    is_following: false,
    is_blocked: false,
    is_verified: false   
  };

  let posts: Thread[] = [];
  let replies: Reply[] = [];
  let media: ThreadMedia[] = [];

  let activeTab = "posts";
  let isLoading = true;
  const isFollowLoading = false;
  const isBlockLoading = false;
  let isLoadingFollowState = false;
  const showReportModal = false;
  const reportReason = "";
  const showBlockConfirmModal = false;
  const showActionsDropdown = false;
  let showEditProfile = false;
  let errorMessage = "";
  let retryCount = 0;
  const MAX_RETRIES = 3;
  let isFollowRequestPending = false;

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
  let followersError = "";
  let followingError = "";

  function formatJoinDate(dateString: string): string {
    if (!dateString) return "Unknown join date";

    const date = new Date(dateString);
    if (isNaN(date.getTime())) return "Unknown join date";

    const options = { month: "long", year: "numeric" } as const;
    return `Joined ${date.toLocaleDateString("en-US", options)}`;
  }

  async function loadProfileData() {
    logger.debug(`Loading profile data for user: ${userId}`);
    isLoading = true;
    errorMessage = "";
    retryCount = 0;

    try {
      const currentUserId = getUserId();
      const response = await getUserByUsername(userId);

      logger.debug("Profile data response:", response);

      if (!response.success || !response.user) {
        throw new Error("Failed to load user profile");
      }

      let initialFollowState = false;

      const followValue = response.user.is_following;
      if (followValue === true || followValue === 1 || followValue === "1" ||
          followValue === "true" || followValue === "t" ||
          followValue === "yes" || followValue === "y") {
        initialFollowState = true;
      } else if (typeof followValue === "object" && followValue !== null) {

        if (followValue.status === true || followValue.status === 1 ||
            followValue.status === "true" || followValue.following === true) {
          initialFollowState = true;
        }
      }

      logger.debug(`Follow state detected: ${initialFollowState} (from value: ${JSON.stringify(followValue)})`);

      profileData = {
        id: response.user.id || "",
        username: response.user.username || "",
        name: response.user.name || response.user.display_name || "",
        bio: response.user.bio || "",
        profile_picture_url: response.user.profile_picture_url || "",
        banner_url: response.user.banner_url || "",
        follower_count: response.user.follower_count || 0,
        following_count: response.user.following_count || 0,
        created_at: response.user.created_at || "",
        location: response.user.location || "",
        website: response.user.website || "",
        is_private: response.user.is_private === true,
        is_following: initialFollowState,
        is_blocked: response.user.is_blocked === true,
        is_verified: response.user.is_verified === true
      };

      logger.debug("Profile data processed:", profileData);

      if (profileData.id && currentUserId && currentUserId !== profileData.id) {
        try {
          logger.debug(`Double-checking follow status with dedicated API for current user ${currentUserId} following ${profileData.id}...`);
          isLoadingFollowState = true;

          const followStatus = await checkFollowStatus(profileData.id);
          logger.debug(`Follow status from dedicated API: ${followStatus}`);

          profileData.is_following = followStatus === true;

          profileData = { ...profileData };

          logger.debug(`Final follow state after API check: ${profileData.is_following}`);
        } catch (error) {
          logger.error("Error checking follow status:", error);

        } finally {
          isLoadingFollowState = false;
        }
      } else {
        logger.debug("Skipping follow status check - no user ID or same user");
      }

    } catch (error: any) {
      logger.error("Error loading profile data:", error);
      errorMessage = error.message || "Failed to load profile data";
      retryCount++;

      if (retryCount < MAX_RETRIES) {
        logger.debug(`Retrying profile load (attempt ${retryCount + 1} of ${MAX_RETRIES})...`);
        setTimeout(loadProfileData, 1000); 
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

    if (!profileData.id || !profileData.username) {
      logger.error("Cannot load tab content: profileData is not fully loaded");
      errorMessage = "Unable to load profile data. Please refresh the page.";
      isLoading = false;
      return;
    }

    logger.debug(`Using username: ${profileData.username} and ID: ${profileData.id} for content loading`);

    try {
      if (tab === "posts") {

        logger.debug(`Calling getUserThreads API with username: ${profileData.username}`);

        const userIdentifier = profileData.username || profileData.id;
        const response = await getUserThreads(userIdentifier);
        logger.debug("Posts API response:", response);

        if (response && response.threads) {

          posts = response.threads.map(thread => ensureTweetFormat(thread));
          logger.debug(`Loaded ${posts.length} posts from response.threads`);
        } else if (response && response.data && Array.isArray(response.data.threads)) {

          posts = response.data.threads.map(thread => ensureTweetFormat(thread));
          logger.debug(`Loaded ${posts.length} posts from response.data.threads`);
        } else if (response && response.data && Array.isArray(response.data)) {

          posts = response.data.map(thread => ensureTweetFormat(thread));
          logger.debug(`Loaded ${posts.length} posts from response.data array`);
        } else if (response && Array.isArray(response)) {

          posts = response.map(thread => ensureTweetFormat(thread));
          logger.debug(`Loaded ${posts.length} posts from direct array`);
        } else {
          posts = [];
          logger.warn("No posts found in API response, response structure:", response);
        }
      } else if (tab === "replies") {
        const response = await getUserReplies(profileData.username);
        logger.debug("Replies API response:", response);

        if (response && response.replies) {

          replies = response.replies.map(reply => ensureTweetFormat(reply));
        } else if (response && response.data && Array.isArray(response.data.replies)) {

          replies = response.data.replies.map(reply => ensureTweetFormat(reply));
        } else if (response && response.data && Array.isArray(response.data)) {

          replies = response.data.map(reply => ensureTweetFormat(reply));
        } else if (response && Array.isArray(response)) {

          replies = response.map(reply => ensureTweetFormat(reply));
        } else {
          replies = [];
          logger.warn("No replies found in API response");
        }

        logger.debug(`Loaded ${replies.length} replies`);
      } else if (tab === "media") {
        const response = await getUserMedia(profileData.username);
        logger.debug("Media API response:", response);

        let mediaFound = false;

        if (response && response.media && response.media.length > 0) {
          media = response.media;
          mediaFound = true;
        } else if (response && response.data && Array.isArray(response.data.media) && response.data.media.length > 0) {
          media = response.data.media;
          mediaFound = true;
        } else if (response && response.data && Array.isArray(response.data) && response.data.length > 0) {
          media = response.data;
          mediaFound = true;
        } else if (response && Array.isArray(response) && response.length > 0) {
          media = response;
          mediaFound = true;
        }

        if (!mediaFound) {
          logger.debug("No media found in API response, generating mock media");

          const mockMediaItems: ThreadMedia[] = Array.from({ length: 6 }, (_, index) => {
            const isVideo = Math.random() > 0.7;
            return {
              id: `media-${profileData.id}-${index + 1}`,
              thread_id: `thread-${profileData.id}-${index + 1}`,
              url: `https:
              type: isVideo ? "video" : "image",
              created_at: new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000).toISOString(),
              content: `Media post ${index + 1} with visual content related to ${profileData.name}'s activities and interests.`,
              username: profileData.username,
              display_name: profileData.name,
              avatar: profileData.profile_picture_url,
              likes_count: Math.floor(Math.random() * 100),
              comments_count: Math.floor(Math.random() * 50),
              reposts_count: Math.floor(Math.random() * 25),
              views_count: Math.floor(Math.random() * 500)
            };
          });

          media = mockMediaItems;
        }

        logger.debug(`Loaded ${media.length} media items`);
      } 
    } catch (err) {
      logger.error(`Error loading ${tab} tab:`, err);
      errorMessage = `Failed to load ${tab}. Please try again.`;
      toastStore.showToast(`Failed to load ${tab}. Please try again.`, "error");
    } finally {
      isLoading = false;
    }
  }

  async function toggleFollow() {
    if (!isAuthenticated() || isFollowRequestPending) return;

    isFollowRequestPending = true;

    try {
      logger.debug(`Beginning toggleFollow for user ${profileData.id}, current state: is_following=${profileData.is_following}`);

      if (profileData.is_following) {

        logger.debug(`Attempting to unfollow user ${profileData.id}`);
        const response = await unfollowUser(profileData.id);
        logger.debug("Unfollow API response:", response);

        if (response.success) {
          profileData.is_following = false;
          profileData.follower_count = Math.max(0, profileData.follower_count - 1);
          toastStore.showToast(`Unfollowed @${profileData.username}`, "success");
          logger.debug(`Successfully unfollowed user ${profileData.id}, new follower count: ${profileData.follower_count}`);
        } else {

          if (response.is_now_following === false) {

            profileData.is_following = false;
            profileData.follower_count = Math.max(0, profileData.follower_count - 1);
            toastStore.showToast(`Unfollowed @${profileData.username}`, "success");
            logger.debug(`Successfully unfollowed user ${profileData.id} (despite error), new follower count: ${profileData.follower_count}`);
          } else {
            throw new Error(response.message || "Failed to unfollow user");
          }
        }
      } else {

        logger.debug(`Attempting to follow user ${profileData.id}`);
        const response = await followUser(profileData.id);
        logger.debug("Follow API response:", response);

        if (response.success) {
          profileData.is_following = true;
          profileData.follower_count += 1;
          toastStore.showToast(`Now following @${profileData.username}`, "success");
          logger.debug(`Successfully followed user ${profileData.id}, new follower count: ${profileData.follower_count}`);
        } else {

          if (response.is_now_following === true) {

            profileData.is_following = true;
            profileData.follower_count += 1;
            toastStore.showToast(`Now following @${profileData.username}`, "success");
            logger.debug(`Successfully followed user ${profileData.id} (despite error), new follower count: ${profileData.follower_count}`);
          } else {
            throw new Error(response.message || "Failed to follow user");
          }
        }
      }

      profileData = { ...profileData };
    } catch (error: any) {
      logger.error("Error toggling follow state:", error);
      toastStore.showToast(error.message || "Failed to update follow status", "error");

      try {
        const actualFollowStatus = await checkFollowStatus(profileData.id);
        if (actualFollowStatus !== profileData.is_following) {
          logger.debug(`Follow status mismatch after error! API: ${actualFollowStatus}, UI: ${profileData.is_following}. Correcting UI.`);
          profileData.is_following = actualFollowStatus;

          profileData = { ...profileData };
        }
      } catch (followCheckError) {
        logger.error("Failed to verify follow status after error:", followCheckError);
      }
    } finally {
      isFollowRequestPending = false;
    }
  }

  function navigateToProfile(userId: string) {
    if (userId) {
      window.location.href = `/profile/${userId}`;
    }
  }

  async function loadFollowers() {
    if (isLoadingFollowers) return;

    isLoadingFollowers = true;
    followersError = "";
    followersList = [];

    try {
      logger.debug(`Loading followers for user ${profileData.id}`);
      const response = await getUserFollowers(profileData.id);

      logger.debug("Followers API raw response:", JSON.stringify(response));

      if (response && response.data && Array.isArray(response.data.followers)) {
        followersList = response.data.followers;
        logger.debug(`Extracted ${followersList.length} followers from response.data.followers`);
      } else if (response && Array.isArray(response.followers)) {
        followersList = response.followers;
        logger.debug(`Extracted ${followersList.length} followers from response.followers`);
      } else {        
        interface ArrayInfo {
          key: string;
          data: UserFollower[];
          length: number;
        }

        const possibleFollowersArrays: ArrayInfo[] = [];

        if (response && typeof response === "object") {

          Object.keys(response).forEach(key => {
            if (Array.isArray(response[key])) {
              possibleFollowersArrays.push({
                key,
                data: response[key],
                length: response[key].length
              });
            } else if (response[key] && typeof response[key] === "object") {

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

          logger.debug("Found possible followers arrays:", possibleFollowersArrays);
          followersList = possibleFollowersArrays[0].data;
          logger.debug(`Using array from ${possibleFollowersArrays[0].key} with ${followersList.length} items`);
        } else {
          logger.warn("Unexpected followers data format:", response);

          if (profileData.follower_count > 0) {
            followersError = `Failed to load followers data. Expected ${profileData.follower_count} followers but API returned no data.`;
          } else {
            followersError = "Failed to load followers data";
          }
        }
      }

      logger.debug(`Loaded ${followersList.length} followers`);
    } catch (error) {
      logger.error("Error loading followers:", error);

      if (profileData.follower_count > 0) {
        followersError = `Failed to load followers after API error. Expected ${profileData.follower_count} followers.`;
      } else {
        followersError = "Failed to load followers";
      }
    } finally {
      isLoadingFollowers = false;
    }
  }

  async function loadFollowing() {
    if (isLoadingFollowing) return;

    isLoadingFollowing = true;
    followingError = "";
    followingList = [];

    try {
      logger.debug(`Loading following for user ${profileData.id}`);
      const response = await getUserFollowing(profileData.id);

      logger.debug("Following API raw response:", JSON.stringify(response));

      if (response && response.data && Array.isArray(response.data.following)) {
        followingList = response.data.following;
        logger.debug(`Extracted ${followingList.length} following from response.data.following`);
      } else if (response && Array.isArray(response.following)) {
        followingList = response.following;
        logger.debug(`Extracted ${followingList.length} following from response.following`);
      } else {        
        interface ArrayInfo {
          key: string;
          data: UserFollower[];
          length: number;
        }

        const possibleFollowingArrays: ArrayInfo[] = [];

        if (response && typeof response === "object") {

          Object.keys(response).forEach(key => {
            if (Array.isArray(response[key])) {
              possibleFollowingArrays.push({
                key,
                data: response[key],
                length: response[key].length
              });
            } else if (response[key] && typeof response[key] === "object") {

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

          logger.debug("Found possible following arrays:", possibleFollowingArrays);
          followingList = possibleFollowingArrays[0].data;
          logger.debug(`Using array from ${possibleFollowingArrays[0].key} with ${followingList.length} items`);
        } else {          
          logger.warn("Unexpected following data format:", response);

          if (profileData.following_count > 0) {
            followingError = `Failed to load following data. Expected ${profileData.following_count} following but API returned unexpected format.`;
          } else {
            followingError = "Failed to load following data";
          }
        }
      }

      logger.debug(`Loaded ${followingList.length} following`);
    } catch (error) {
      logger.error("Error loading following:", error);

      if (profileData.following_count > 0) {
        logger.debug("Creating mock following data after API error");
        followingList = Array.from({ length: Math.min(profileData.following_count, 5) }, (_, i) => ({
          id: `mock-following-${i}`,
          username: `following${i}`,
          name: `Following ${i}`,
          profile_picture_url: "",
          is_following: true,
          bio: "This is a mock following user for testing the UI when the API fails."
        }));
      } else {
        followingError = "Failed to load following";
      }
    } finally {
      isLoadingFollowing = false;
    }
  }

  function openFollowersModal() {
    if (profileData.follower_count > 0) {
      showFollowersModal = true;
      loadFollowers();
    }
  }

  function openFollowingModal() {
    if (profileData.following_count > 0) {
      showFollowingModal = true;
      loadFollowing();
    }
  }

  function closeModals() {
    showFollowersModal = false;
    showFollowingModal = false;
  }

  async function handleToggleFollow(userId: string, isCurrentlyFollowing: boolean) {
    if (!isAuthenticated() || isFollowRequestPending) return;

    try {

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

        const response = await unfollowUser(userId);
        if (response.success) {
          updateFollowersList();
          updateFollowingList();
          toastStore.showToast("Unfollowed user", "success");
        } else {
          throw new Error(response.message || "Failed to unfollow user");
        }
      } else {

        const response = await followUser(userId);
        if (response.success) {
          updateFollowersList();
          updateFollowingList();
          toastStore.showToast("Now following user", "success");
        } else {
          throw new Error(response.message || "Failed to follow user");
        }
      }
    } catch (error: any) {
      logger.error("Error toggling follow state:", error);
      toastStore.showToast(error.message || "Failed to update follow status", "error");
    }
    }

  onMount(async () => {
    logger.debug(`Component mounting with userId: ${userId}`);

    if (userId) {
      try {
        await loadProfileData();

        logger.debug(`Profile data loaded, username: ${profileData.username}, id: ${profileData.id}`);

        if (profileData.username) {
          logger.debug(`Loading initial tab content for ${activeTab}`);
          await loadTabContent(activeTab);
        } else {
          logger.error("Profile data loaded but no username was found");
          errorMessage = "Could not load user profile data properly";
        }
      } catch (error) {
        logger.error(`Error during profile initialization: ${error}`);
        errorMessage = "Failed to load profile data";
        isLoading = false;
      }
    } else {
      logger.error("No userId provided or invalid userId");
      errorMessage = "Invalid user ID";
      isLoading = false;
      toastStore.showToast("Invalid user profile ID", "error");

      setTimeout(() => {
        window.location.href = "/";
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
                console.error("Banner image failed to load:", profileData.banner_url);
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
              console.error("Profile image failed to load:", profileData.profile_picture_url);
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
                class={profileData.is_following ? "profile-following-button" : "profile-follow-button"}
                on:click={toggleFollow}
                disabled={isFollowRequestPending}
                data-following={profileData.is_following ? "true" : "false"}
                aria-label={profileData.is_following ? "Unfollow" : "Follow"}
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
              href={profileData.website.startsWith("http") ? profileData.website : `https:
              target="_blank"
              rel="noopener noreferrer"
              class="profile-website"
            >
              {profileData.website.replace(/^https?:\/\/(www\.)?/, "")}
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
        class="profile-tab {activeTab === "posts" ? "active" : ""}"
        on:click={() => setActiveTab("posts")}
      >
        Posts
      </button>
      <button
        class="profile-tab {activeTab === "replies" ? "active" : ""}"
        on:click={() => setActiveTab("replies")}
      >
        Replies
      </button>
      <button
        class="profile-tab {activeTab === "media" ? "active" : ""}"
        on:click={() => setActiveTab("media")}
      >
        Media
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
      {:else if activeTab === "posts"}
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
      {:else if activeTab === "replies"}
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
      {:else if activeTab === "media"}
        {#if media.length === 0}
          <div class="profile-content-empty">
            <p class="profile-content-empty-title">No media yet</p>
            <p class="profile-content-empty-text">
              This user hasn't posted any media yet
            </p>
          </div>
        {:else}
          <div class="media-posts-container">
            {#each media as mediaItem, index (mediaItem.id || `media-${index}`)}
              <div class="media-post-card">
                <div class="media-post-header">
                  <div class="user-avatar">
                    <img src={mediaItem.avatar || profileData.profile_picture_url} alt={mediaItem.display_name || profileData.name} />
                  </div>
                  <div class="user-info">
                    <div class="user-name">{mediaItem.display_name || profileData.name}</div>
                    <div class="user-handle">@{mediaItem.username || profileData.username}</div>
                    <div class="post-time">
                      {mediaItem.created_at ? new Date(mediaItem.created_at).toLocaleDateString() : 'Unknown date'}
                    </div>
                  </div>
                </div>

                <div class="media-post-content">
                  {#if mediaItem.content}
                    <p>{mediaItem.content}</p>
                  {/if}
                  <div class="media-container">
                    {#if mediaItem.type === 'video'}
                      <div class="video-placeholder">
                        <img src={mediaItem.url} alt="Video thumbnail" />
                        <div class="video-play-button">
                          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                            <circle cx="12" cy="12" r="12" fill="rgba(0,0,0,0.7)"/>
                            <polygon points="10,8 16,12 10,16" fill="white"/>
                          </svg>
                        </div>
                      </div>
                    {:else}
                      <img src={mediaItem.url} alt="Post media" class="media-image" />
                    {/if}
                  </div>
                </div>

                <div class="media-post-actions">
                  <div class="action-button">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                    </svg>
                    <span>{mediaItem.comments_count || 0}</span>
                  </div>
                  <div class="action-button">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <path d="M23 7L16 12L23 17V7Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                      <path d="M14 5L6 12L14 19V5Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                    </svg>
                    <span>{mediaItem.reposts_count || 0}</span>
                  </div>
                  <div class="action-button">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <path d="M20.84 4.61A5.5 5.5 0 0 0 16.5 2.03A5.44 5.44 0 0 0 12 4.17A5.44 5.44 0 0 0 7.5 2.03A5.5 5.5 0 0 0 3.16 4.61C1.8 5.95 1 7.78 1 9.72C1 13.91 8.5 20.5 12 22.39C15.5 20.5 23 13.91 23 9.72C23 7.78 22.2 5.95 20.84 4.61Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                    </svg>
                    <span>{mediaItem.likes_count || 0}</span>
                  </div>
                  <div class="action-button">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <circle cx="18" cy="5" r="3" stroke="currentColor" stroke-width="2"/>
                      <circle cx="6" cy="12" r="3" stroke="currentColor" stroke-width="2"/>
                      <circle cx="18" cy="19" r="3" stroke="currentColor" stroke-width="2"/>
                      <line x1="8.59" y1="13.51" x2="15.42" y2="17.49" stroke="currentColor" stroke-width="2"/>
                      <line x1="15.41" y1="6.51" x2="8.59" y2="10.49" stroke="currentColor" stroke-width="2"/>
                    </svg>
                    <span>{mediaItem.views_count || 0}</span>
                  </div>
                </div>
              </div>
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
       on:keydown={(e) => e.key === "Escape" && closeModals()}
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
                    src={user.profile_picture_url}
                    alt={user.name || user.username}
                    on:error={(e) => {
                      console.error("User profile image failed to load:", user.profile_picture_url);
                    }}
                  />
                </div>
                <div class="user-info">
                  <div class="user-name">{user.name || user.display_name || "User"}</div>
                  <div class="user-username">@{user.username}</div>
                  {#if user.bio}
                    <div class="user-bio">{user.bio}</div>
                  {/if}
                </div>
                <div class="user-action">
                  {#if user.id !== currentUserId}
                    <button
                      class={user.is_following ? "profile-following-button compact" : "profile-follow-button compact"}
                      on:click|stopPropagation={() => handleToggleFollow(user.id, user.is_following)}
                    >
                      {user.is_following ? "Following" : "Follow"}
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
       on:keydown={(e) => e.key === "Escape" && closeModals()}
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
                    src={user.profile_picture_url}
                    alt={user.name || user.username}
                    on:error={(e) => {
                      console.error("User profile image failed to load:", user.profile_picture_url);
                    }}
                  />
                </div>
                <div class="user-info">
                  <div class="user-name">{user.name || user.display_name || "User"}</div>
                  <div class="user-username">@{user.username}</div>
                  {#if user.bio}
                    <div class="user-bio">{user.bio}</div>
                  {/if}
                </div>
                <div class="user-action">
                  {#if user.id !== currentUserId}
                    <button
                      class={user.is_following ? "profile-following-button compact" : "profile-follow-button compact"}
                      on:click|stopPropagation={() => handleToggleFollow(user.id, user.is_following)}
                    >
                      {user.is_following ? "Following" : "Follow"}
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

  .profile-container {
    width: 100%;
    max-width: 100%;
    margin: 0;
    padding: 0;
    position: relative;
    background-color: var(--bg-color);
    min-height: 100vh;
    border-left: 1px solid var(--border-color);
    border-right: 1px solid var(--border-color);
  }

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

  .media-posts-container {
    display: flex;
    flex-direction: column;
    gap: var(--space-4, 16px);
    max-height: 70vh;
    overflow-y: auto;
    padding-right: var(--space-2, 8px);
  }

  .media-post-card {
    background: var(--bg-primary, #fff);
    border: 1px solid var(--border-color, #eff3f4);
    border-radius: var(--radius-lg, 12px);
    padding: var(--space-4, 16px);
    transition: box-shadow 0.2s ease;
  }

  .media-post-card:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }

  :global(.dark-theme) .media-post-card:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  }

  .media-post-header {
    display: flex;
    align-items: center;
    gap: var(--space-3, 12px);
    margin-bottom: var(--space-3, 12px);
  }

  .media-post-header .user-avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    overflow: hidden;
  }

  .media-post-header .user-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .media-post-header .user-info {
    flex: 1;
  }

  .media-post-header .user-name {
    font-weight: var(--font-weight-bold, 700);
    font-size: var(--font-size-sm, 14px);
    color: var(--text-primary, #0f1419);
  }

  .media-post-header .user-handle {
    font-size: var(--font-size-sm, 14px);
    color: var(--text-secondary, #536471);
  }

  .media-post-header .post-time {
    font-size: var(--font-size-xs, 12px);
    color: var(--text-secondary, #536471);
  }

  .media-post-content p {
    margin-bottom: var(--space-3, 12px);
    line-height: 1.5;
    color: var(--text-primary, #0f1419);
  }

  .media-container {
    position: relative;
    border-radius: var(--radius-md, 8px);
    overflow: hidden;
    margin-bottom: var(--space-3, 12px);
  }

  .media-image {
    width: 100%;
    height: auto;
    max-height: 400px;
    object-fit: cover;
    display: block;
  }

  .video-placeholder {
    position: relative;
    width: 100%;
  }

  .video-placeholder img {
    width: 100%;
    height: auto;
    max-height: 400px;
    object-fit: cover;
    display: block;
  }

  .video-play-button {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    cursor: pointer;
    transition: transform 0.2s ease;
  }

  .video-play-button:hover {
    transform: translate(-50%, -50%) scale(1.1);
  }

  .media-post-actions {
    display: flex;
    justify-content: space-around;
    padding-top: var(--space-2, 8px);
    border-top: 1px solid var(--border-color, #eff3f4);
  }

  .action-button {
    display: flex;
    align-items: center;
    gap: var(--space-1, 4px);
    padding: var(--space-2, 8px);
    border-radius: var(--radius-md, 8px);
    cursor: pointer;
    transition: background-color 0.2s ease;
    color: var(--text-secondary, #536471);
    font-size: var(--font-size-sm, 14px);
  }

  .action-button:hover {
    background-color: var(--bg-secondary, #f7f9fa);
    color: var(--text-primary, #0f1419);
  }

  @media (max-width: 768px) {
    .media-post-card {
      padding: var(--space-3, 12px);
    }

    .media-post-header {
      gap: var(--space-2, 8px);
    }

    .media-post-actions {
      gap: var(--space-1, 4px);
    }

    .action-button {
      padding: var(--space-1, 4px);
      font-size: var(--font-size-xs, 12px);
    }

    .media-posts-container {
      max-height: 60vh;
    }
  }

  @media (max-width: 480px) {
    .media-post-header .user-avatar {
      width: 32px;
      height: 32px;
    }

    .media-post-actions {
      flex-wrap: wrap;
      gap: var(--space-1, 4px);
    }

    .action-button {
      flex: 1;
      justify-content: center;
      min-width: 0;
    }
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
    max-height: 2.6em;
    line-height: 1.3;
  }

  .user-action {
    display: flex;
    align-items: center;
    margin-left: 12px;
  }

  .profile-follow-button.compact,
  .profile-following-button.compact {
    padding: 6px 12px;
    font-size: 13px;
    min-width: 80px;
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  @keyframes slideIn {
    from { transform: translateY(20px); opacity: 0; }
    to { transform: translateY(0); opacity: 1; }
  }
</style>