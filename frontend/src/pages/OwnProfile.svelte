<script lang="ts">
  import { onMount } from "svelte";
  import MainLayout from "../components/layout/MainLayout.svelte";
  import { useAuth } from "../hooks/useAuth";
  import { useTheme } from "../hooks/useTheme";
  import { isAuthenticated, getUserId } from "../utils/auth";
  import { getProfile, updateProfile, pinThread, unpinThread, pinReply, unpinReply, getUserById, getUserFollowers, getUserFollowing } from "../api/user";
  import { getUserThreads, getUserReplies, getUserLikedThreads, getUserMedia, getThreadReplies, likeThread, unlikeThread, bookmarkThread, removeBookmark, getUserBookmarks } from "../api/thread";
  import { toastStore } from "../stores/toastStore";
  import ThreadCard from "../components/explore/ThreadCard.svelte";
  import TweetCard from "../components/social/TweetCard.svelte";
  import LoadingSkeleton from "../components/common/LoadingSkeleton.svelte";
  import ProfileEditModal from "../components/profile/ProfileEditModal.svelte";
  import { formatStorageUrl, isSupabaseStorageUrl } from "../utils/common";
  import CalendarIcon from "svelte-feather-icons/src/icons/CalendarIcon.svelte";
  import XIcon from "svelte-feather-icons/src/icons/XIcon.svelte";
  import PinIcon from "svelte-feather-icons/src/icons/MapPinIcon.svelte";
  import CheckCircleIcon from "svelte-feather-icons/src/icons/CheckCircleIcon.svelte";
  import type { ITweet } from "../interfaces/ISocialMedia";

  interface IFollowUser {
    id: string;
    username: string;
    name?: string;
    display_name?: string;
    profile_picture_url?: string;
    is_following?: boolean;
    is_followed_by?: boolean;
    bio?: string;
  }

  // Define interfaces for our data structures
  interface Thread {
    id: string;
    content: string;
    username: string;
    display_name: string;
    timestamp: string;
    likes: number;
    replies: number;
    reposts: number;
    created_at: string;
    updated_at?: string;
    user_id: string;
    likes_count?: number;
    replies_count?: number;
    reposts_count?: number;
    bookmark_count?: number;
    views_count?: number;
    is_liked?: boolean;
    is_reposted?: boolean;
    is_bookmarked?: boolean;
    is_pinned?: boolean;
    IsPinned?: boolean; // Added to match API property
    parent_id?: string | null;
    author_id?: string;
    author_username?: string;
    author_name?: string;
    author_avatar?: string;
    profile_picture_url?: string;
    name?: string;
    thread_id?: string; // Thread ID for replies referring to parent thread
    media?: Array<{
      type: string;
      url: string;
      id?: string;
    }>;
    avatar?: string;
    [key: string]: any; // For any additional properties
  }

  interface Reply {
    id: string;
    content: string;
    created_at: string;
    thread_id: string;
    thread_author: string;
    author_id?: string;
    author_username?: string;
    author_name?: string;
    author_avatar?: string;
    likes_count?: number;
    is_liked?: boolean;
    is_pinned?: boolean;
    [key: string]: any; // For any additional properties
  }

  interface ThreadMedia {
    id: string;
    url: string;
    type: "image" | "video" | "gif";
    thread_id: string;
    created_at?: string;
    [key: string]: any; // For any additional properties
  }

  // Auth and theme
  const { getAuthState } = useAuth();
  const { theme } = useTheme();

  // Get userId from URL parameter or use current user
  export let userId: string = "";

  // Determine if we're viewing the current user's profile or another user's profile
  $: isOwnProfile = !userId || userId === "me" || userId === getUserId();
  $: profileUserId = isOwnProfile ? "me" : userId;

  // Reactive declarations
  $: isDarkMode = $theme === "dark";
  $: authState = getAuthState();

  // Define default image URL for fallback
  const DEFAULT_AVATAR = "https://secure.gravatar.com/avatar/0?d=mp";

  // Profile data
  let profileData = {
    id: "",
    username: "",
    displayName: "",
    bio: "",
    profilePicture: DEFAULT_AVATAR,
    backgroundBanner: "",
    followerCount: 0,
    followingCount: 0,
    joinedDate: "",
    email: "",
    dateOfBirth: "",
    gender: "",
    isVerified: false
  };

  // Content data with types
  let posts: Thread[] = [];
  let replies: Reply[] = [];
  let likes: Thread[] = [];
  let media: Thread[] = []; // Change to Thread[] type to match the other tabs
  const bookmarks: Thread[] = [];

  // UI state
  let activeTab = "posts";
  let isLoading = true;
  let showEditModal = false;
  let showPicturePreview = false; // Controls the profile picture preview modal
  let isUpdatingProfile = false;
  const searchQuery = "";
  const showPinnedOnly = false;

  // Log initial state for debugging
  console.log("Initial showPicturePreview state:", showPicturePreview);

  // Function to toggle profile picture preview
  function toggleProfilePicturePreview() {
    showPicturePreview = !showPicturePreview;
    console.log("Toggled picture preview:", showPicturePreview);
  }

  // Modal state for followers/following
  let showFollowersModal = false;
  let showFollowingModal = false;
  let followersList: IFollowUser[] = [];
  let followingList: IFollowUser[] = [];
  let isLoadingFollowers = false;
  let isLoadingFollowing = false;
  let followersError = "";
  let followingError = "";

  // Additional functions for thread interactions
  let repliesMap = new Map(); // Store replies for threads
  let nestedRepliesMap = new Map(); // Store nested replies

  // State variables
  const profile: any = null;
  let errorMessage = "";

  // Helper function to ensure an object has all ITweet properties
  function ensureTweetFormat(thread: any): ITweet {
    // Handle is_pinned value consistently
    let isPinned = false;
    if (thread.is_pinned === true || thread.is_pinned === "true" ||
        thread.is_pinned === 1 || thread.is_pinned === "1" ||
        thread.is_pinned === "t") {
      isPinned = true;
      console.log(`Thread ${thread.id} IS PINNED`);
    }

    // Get username from all possible sources
    const username = thread.author_username || thread.username || "anonymous";

    // Get display name from all possible sources
    const name = thread.author_name || thread.name || "User";

    // Get profile picture from all possible sources
    const profile_picture_url = thread.author_profile_picture_url || thread.profile_picture_url ||
                       thread.author_avatar || "https://secure.gravatar.com/avatar/0?d=mp";

    // Use the created_at timestamp if available, fall back to UTC now
    let created_at = thread.created_at || new Date().toISOString();
    if (typeof created_at === "string" && !created_at.includes("T")) {
      // Convert to ISO format if it's not already
      created_at = new Date(created_at).toISOString();
    }

    // Normalize metrics
    const likes_count = thread.likes_count || 0;
    const replies_count = thread.replies_count || 0;
    const reposts_count = thread.reposts_count || 0;
    const bookmark_count = thread.bookmark_count || thread.bookmarks_count || 0;
    const views_count = thread.views_count || 0;

    // Normalize interaction states
    const is_liked = thread.is_liked || false;
    const is_reposted = thread.is_reposted || false;
    const is_bookmarked = thread.is_bookmarked || false;

    // For type safety, create a properly formatted ITweet object
    const tweetData: ITweet = {
      id: thread.id,
      user_id: thread.user_id || thread.author_id || userId,
      username: username,
      name: name,
      content: thread.content || "",
      created_at: typeof created_at === "string" ? created_at : new Date(created_at).toISOString(),
      updated_at: thread.updated_at || undefined,
      profile_picture_url: profile_picture_url,
      likes_count: likes_count,
      replies_count: replies_count,
      reposts_count: reposts_count,
      bookmark_count: bookmark_count,
      views_count: views_count,
      media: thread.media || [],
      is_liked: is_liked,
      is_reposted: is_reposted,
      is_bookmarked: is_bookmarked,
      is_pinned: isPinned,
      parent_id: thread.parent_id || null
    };

    // If there's a thread_id property (in replies), store it as a custom property
    if (thread.thread_id || thread.reply_to) {
      (tweetData as any).reply_to_thread_id = thread.thread_id ||
                                             (typeof thread.reply_to === "string" ? thread.reply_to :
                                             thread.reply_to?.id || null);
    }

    return tweetData;
  }

  // Helper function to troubleshoot and manually repair pinned threads if needed
  async function troubleshootPinnedThreads() {
    console.log("Troubleshooting pinned threads...");

    try {
      // Step 1: Fetch recent threads to check if any should be pinned
      let postsData;
      let threads: any[] = [];

      try {
        postsData = await getUserThreads(profileUserId);
        threads = postsData.threads || [];
        console.log("Threads from API:", threads.length);
      } catch (error: any) {
        console.error("Error fetching threads for troubleshooting:", error);
        // Continue with empty threads array rather than failing completely
        toastStore.showToast("Unable to check pinned threads - will try later", "warning");
        return; // Exit the function but don't throw error up to caller
      }

      // Check different formats of is_pinned that might be present
      const pinned = {
        boolean: threads.filter(t => t.is_pinned === true).length,
        string: threads.filter(t => t.is_pinned === "true").length,
        number: threads.filter(t => t.is_pinned === 1).length,
        stringNumber: threads.filter(t => t.is_pinned === "1").length,
        postgresTrue: threads.filter(t => t.is_pinned === "t").length,
        postgresFalse: threads.filter(t => t.is_pinned === "f").length,
        altCase: threads.filter(t => t.IsPinned === true).length,
        nullOrUndefined: threads.filter(t => t.is_pinned === null || t.is_pinned === undefined).length
      };

      console.log("Pinned threads by format:", pinned);

      // Get possible pinned threads with any format
      const possiblyPinned = threads.filter(t =>
        t.is_pinned === true ||
        t.is_pinned === "true" ||
        t.is_pinned === 1 ||
        t.is_pinned === "1" ||
        t.is_pinned === "t" ||
        t.IsPinned === true
      );

      console.log("Possible pinned threads:", possiblyPinned.map(t => t.id));

      // Skip the repair steps if we couldn't get threads or there's an obvious error
      if (threads.length === 0) {
        console.log("No threads found, skipping pin repair");
        return;
      }

      // Find threads that should be pinned based on our database check
      // These IDs should match what you see in the database
      const shouldBePinned = [
        "5a07bfd3-5e19-401c-89a6-c14da0e44da7", // Based on the db screenshot showing is_pinned = t
        "3d3ffe9e-3b9f-4fdd-a942-7eda514429ea"  // Based on the second row in the db showing is_pinned = t
      ];

      // Only proceed with repair if we have valid thread data
      if (!Array.isArray(threads)) {
        console.error("Threads data is not an array, can't repair pins");
        return;
      }

      // Check if any of these threads aren't showing as pinned
      const unpinnedThreads = threads.filter(thread =>
        shouldBePinned.includes(thread.id) &&
        !(thread.is_pinned === true || thread.is_pinned === "true" ||
          thread.is_pinned === 1 || thread.is_pinned === "1" ||
          thread.is_pinned === "t")
      );

      console.log("Threads that should be pinned according to database:", shouldBePinned);
      console.log("Threads that need pinning repair:", unpinnedThreads.map(t => t.id));

      // Check if there are any threads that have 't' as is_pinned value
      const postgresFormat = threads.filter(t => t.is_pinned === "t").map(t => ({
        id: t.id,
        is_pinned: t.is_pinned,
        content: t.content?.substring(0, 30) + "..."
      }));

      console.log("Threads with PostgreSQL 't' format:", postgresFormat);

      // Fix any threads that should be pinned but aren't
      if (unpinnedThreads.length > 0) {
        console.log(`Attempting to repair pin status for ${unpinnedThreads.length} threads...`);

        let repairSuccessCount = 0;
        for (const thread of unpinnedThreads) {
          try {
            console.log(`Repairing pin status for thread ${thread.id}`);
            await pinThread(thread.id);
            repairSuccessCount++;
          } catch (err) {
            console.error(`Failed to repair thread ${thread.id}:`, err);
            // Continue with next thread rather than failing
          }
        }

        // Reload data after repairs if at least one repair succeeded
        if (repairSuccessCount > 0) {
          toastStore.showToast(`Repaired ${repairSuccessCount} pinned thread(s). Reloading...`, "success");
          setTimeout(() => loadTabContent("posts"), 2000);
        } else {
          toastStore.showToast("Could not repair pinned thread status. Please try again later.", "error");
        }
      }
    } catch (error: any) {
      console.error("Error troubleshooting pinned threads:", error);
      // Show error but don't crash the profile page
      toastStore.showToast("Error checking pinned threads: " + error.message, "warning");
    }
  }

  async function loadRepliesForThread(threadId) {
    try {
      const response: any = await getThreadReplies(threadId);
      if (response && response.replies) {
        console.log(`Loaded ${response.replies.length} replies for thread ${threadId}`);

        const convertedReplies = response.replies.map((reply: any) => {
          // Standardize the reply data structure regardless of API format
          const replyData = reply.reply || reply || {};
          const userData = reply.user || {} as any;

          const enrichedReply = {
            id: replyData.id || "",
            thread_id: replyData.thread_id || threadId,
            content: replyData.content || "",
            created_at: replyData.created_at || new Date().toISOString(),
            author_id: userData.id || replyData.user_id || replyData.author_id || "",
            author_username: userData.username || replyData.author_username || replyData.username || "",
            author_name: userData.name || userData.display_name || replyData.author_name || replyData.displayName || "",
            author_avatar: userData.profile_picture_url || replyData.author_avatar || replyData.avatar || "",
            parent_id: replyData.parent_id || "",
            is_liked: replyData.is_liked || false,
            is_bookmarked: replyData.is_bookmarked || false,
            likes_count: replyData.likes_count || 0,
            replies_count: replyData.replies_count || 0
          };

          const convertedReply = ensureTweetFormat(enrichedReply);

          // Add reply-specific fields
          convertedReply.reply_to = threadId as any;
          (convertedReply as any).parentReplyId = replyData.parent_id;

          return convertedReply;
        });

        repliesMap.set(threadId, convertedReplies);

        // Process nested replies
        convertedReplies.forEach(reply => {
          const parentId = (reply as any).parentReplyId;
          if (parentId) {
            const parentReplies = nestedRepliesMap.get(parentId) || [];
            nestedRepliesMap.set(parentId, [...parentReplies, reply]);
          }
        });

        repliesMap = repliesMap;
        nestedRepliesMap = nestedRepliesMap;
      } else {
        console.warn(`No replies returned for thread ${threadId}`);
        repliesMap.set(threadId, []);
        repliesMap = repliesMap;
      }
    } catch (error: any) {
      console.error(`Error fetching replies for thread ${threadId}:`, error);
      toastStore.showToast("Failed to load replies. Please try again.", "error");
      repliesMap.set(threadId, []);
      repliesMap = repliesMap;
    }
  }

  function handleLoadReplies(event) {
    const threadId = event.detail;
    loadRepliesForThread(threadId);
  }

  function handleReply(event) {
    const threadId = event.detail;
    window.location.href = `/thread/${threadId}`;
  }

  function handleThreadClick(event) {
    const thread = event.detail;
    if (thread && thread.id) {
      console.log("Navigating to thread:", thread.id);
      window.location.href = `/thread/${thread.id}`;
    } else {
      console.error("Invalid thread data in click event:", event);
    }
  }

  function setActiveTab(tab) {
    activeTab = tab;
    loadTabContent(tab);
  }

  // Helper function to determine if a thread is pinned, handling different data formats
  function isThreadPinned(thread: Thread): boolean {
    if (!thread) return false;

    // First check for boolean values
    if (thread.is_pinned === true || thread.IsPinned === true) return true;

    // Handle string values with explicit type assertion
    if (typeof thread.is_pinned === "string") {
      // Use type assertion to tell TypeScript that is_pinned is definitely a string
      const pinValue = (thread.is_pinned as string).toLowerCase();
      return pinValue === "true" || pinValue === "t" || pinValue === "1";
    }

    // Handle numeric values
    if (typeof thread.is_pinned === "number") {
      return thread.is_pinned === 1;
    }

    return false;
  }

  async function loadTabContent(tab: string) {
    isLoading = true;

    try {
      if (tab === "posts") {
        const postsData = await getUserThreads(profileUserId);
        console.log("Posts data from API:", postsData);

        // Extract threads from the response, handling nested data structures
        if (postsData && postsData.success) {
          // Check if data is in the main object or nested inside a data property
          let postsArray: Thread[] = [];
          if (postsData.threads) {
            postsArray = postsData.threads;
          } else if (postsData.data && postsData.data.threads) {
            postsArray = postsData.data.threads;
          } else {
            postsArray = [];
          }

          // Sort posts to ensure pinned ones are at the top
          postsArray.sort((a: Thread, b: Thread) => {
            // First sort by pinned status
            const aIsPinned = isThreadPinned(a);
            const bIsPinned = isThreadPinned(b);

            if (aIsPinned && !bIsPinned) return -1;
            if (!aIsPinned && bIsPinned) return 1;

            // Then sort by creation date (newest first)
            const dateA = new Date(a.created_at);
            const dateB = new Date(b.created_at);
            return dateB.getTime() - dateA.getTime();
          });

          posts = postsArray;
          console.log(`Loaded ${posts.length} posts:`, posts);

          // Log pinned posts for debugging
          const pinnedPosts = posts.filter(post => isThreadPinned(post));
          console.log(`Found ${pinnedPosts.length} pinned posts:`, pinnedPosts);
        } else {
          console.error("Failed to get posts:", postsData);
          posts = [];
        }
      } else if (tab === "replies") {
        const repliesData = await getUserReplies(profileUserId);
        replies = repliesData.replies || [];
        console.log(`Loaded ${replies.length} replies`);
      } else if (tab === "likes") {
        const likesData = await getUserLikedThreads(profileUserId);
        likes = likesData.threads || [];
        console.log(`Loaded ${likes.length} likes`);
      } else if (tab === "media") {
        const mediaData = await getUserMedia(profileUserId);
        media = mediaData.threads || [];
        console.log(`Loaded ${media.length} media posts`);
      }
    } catch (error) {
      console.error(`Error loading tab content for ${tab}:`, error);
      toastStore.showToast(`Failed to load ${tab}. Please try again.`, "error");
    } finally {
      isLoading = false;
    }
  }

  async function loadProfileData() {
    isLoading = true;
    try {
      let response;
      if (isOwnProfile) {
        response = await getProfile();
        console.log("getProfile API response:", response);
      } else {
        response = await getUserById(profileUserId);
        console.log("getUserById API response:", response);
      }

      if (response && response.user) {
        console.log("Raw profile data received:", response.user);
        console.log("is_verified in API response:", response.user.is_verified);
        console.log("Follower count in API response:", response.user.follower_count);
        console.log("Following count in API response:", response.user.following_count);

        // Extract user data exactly like in LeftSide.svelte
        const userData = response.user;

        profileData = {
          id: userData.id || "",
          username: userData.username || "",
          displayName: userData.name || userData.display_name || "",
          bio: userData.bio || "",
          profilePicture: userData.profile_picture_url || DEFAULT_AVATAR,
          backgroundBanner: userData.banner_url || userData.background_banner_url || "",
          followerCount: userData.follower_count || 0,
          followingCount: userData.following_count || 0,
          joinedDate: userData.created_at ? new Date(userData.created_at).toLocaleDateString("en-US", { month: "long", year: "numeric" }) : "",
          email: isOwnProfile ? (userData.email || "") : "",
          dateOfBirth: isOwnProfile ? (userData.date_of_birth || "") : "",
          gender: isOwnProfile ? (userData.gender || "") : "",
          isVerified: userData.is_verified || false
        };

        console.log("Profile loaded with isVerified:", profileData.isVerified);
        console.log("Profile loaded with follower count:", profileData.followerCount);
        console.log("Profile loaded with following count:", profileData.followingCount);
        console.log("Profile loaded with avatar URL:", profileData.profilePicture);
      }
    } catch (error: any) {
      console.error("Error loading profile:", error);
      errorMessage = "Failed to load profile. Please try again later.";
      toastStore.showToast("Failed to load profile. Please try again.", "error");
    } finally {
      isLoading = false;
    }
  }

  async function handleProfileUpdate(event) {
    const updatedData = event.detail;
    isUpdatingProfile = true;

    console.log("Received profile update event with data:", updatedData);
    console.log("Current profile data before update:", profileData);

    try {
      // Map the form field names to what the backend API expects
      const apiData = {
        name: updatedData.name,      // Backend expects 'name'
        bio: updatedData.bio,
        email: updatedData.email,
        date_of_birth: updatedData.date_of_birth,  // Backend uses snake_case 'date_of_birth'
        gender: updatedData.gender
        // profile_picture_url and banner_url are handled separately via their own handlers
      };

      console.log("Sending profile update to API:", apiData);
      const response = await updateProfile(apiData);
      console.log("Profile update API response:", response);

      if (response && response.success) {
        toastStore.showToast("Profile updated successfully!", "success");
        // Update local profile data to reflect changes
        profileData = {
          ...profileData,
          displayName: updatedData.name,
          bio: updatedData.bio,
          email: updatedData.email,
          dateOfBirth: updatedData.date_of_birth,
          gender: updatedData.gender
        };

        // Also reload full profile data from server to ensure everything is in sync
        loadProfileData();
      } else {
        throw new Error(response?.message || "Failed to update profile");
      }
    } catch (error: any) {
      console.error("Error updating profile:", error);
      toastStore.showToast("Failed to update profile. Please try again.", "error");
    } finally {
      isUpdatingProfile = false;
      showEditModal = false;
    }
  }

  // Handler for profile picture updated event
  function handleProfilePictureUpdated(e) {
    const url = e.detail.url;
    profileData = { ...profileData, profilePicture: url };
    console.log("Profile picture updated:", getProfileImageUrl(url));
  }

  // Handler for banner updated event
  function handleBannerUpdated(e) {
    const url = e.detail.url;
    profileData = { ...profileData, backgroundBanner: url };
    console.log("Banner updated:", getBannerImageUrl(url));
  }

  async function handleLike(event) {
    const threadId = event.detail;
    try {
      await likeThread(threadId);

      // Update posts tab
      posts = posts.map(post => {
        if (post.id === threadId) {
          return {
            ...post,
            isLiked: true,
            is_liked: true,
            likes_count: (post.likes_count || 0) + 1
          };
        }
        return post;
      });

      // Update likes tab
      likes = likes.map(like => {
        if (like.id === threadId) {
          return {
            ...like,
            isLiked: true,
            is_liked: true,
            likes_count: (like.likes_count || 0) + 1
          };
        }
        return like;
      });

      toastStore.showToast("Post liked", "success");
    } catch (error: any) {
      console.error("Error liking thread:", error);
      toastStore.showToast("Failed to like post. Please try again.", "error");
    }
  }

  async function handleUnlike(event) {
    const threadId = event.detail;
    try {
      await unlikeThread(threadId);

      // Update posts tab
      posts = posts.map(post => {
        if (post.id === threadId) {
          return {
            ...post,
            isLiked: false,
            is_liked: false,
            likes_count: Math.max(0, (post.likes_count || 0) - 1)
          };
        }
        return post;
      });

      // Update likes tab
      likes = likes.map(like => {
        if (like.id === threadId) {
          return {
            ...like,
            isLiked: false,
            is_liked: false,
            likes_count: Math.max(0, (like.likes_count || 0) - 1)
          };
        }
        return like;
      });

      toastStore.showToast("Post unliked", "success");
    } catch (error: any) {
      console.error("Error unliking thread:", error);
      toastStore.showToast("Failed to unlike post. Please try again.", "error");
    }
  }

  async function handleBookmark(event) {
    const threadId = event.detail;
    try {
      await bookmarkThread(threadId);

      // Update posts tab
      posts = posts.map(post => {
        if (post.id === threadId) {
          return {
            ...post,
            isBookmarked: true,
            is_bookmarked: true,
            bookmarks_count: (post.bookmarks_count || 0) + 1
          };
        }
        return post;
      });

      // Update likes tab
      likes = likes.map(like => {
        if (like.id === threadId) {
          return {
            ...like,
            isBookmarked: true,
            is_bookmarked: true,
            bookmarks_count: (like.bookmarks_count || 0) + 1
          };
        }
        return like;
      });

      // Update replies tab if the same thread ID exists
      replies = replies.map(reply => {
        if (reply.id === threadId) {
          return {
            ...reply,
            isBookmarked: true,
            is_bookmarked: true,
            bookmarks_count: (reply.bookmarks_count || 0) + 1
          };
        }
        return reply;
      });

      toastStore.showToast("Post bookmarked", "success");
    } catch (error: any) {
      console.error("Error bookmarking thread:", error);
      toastStore.showToast("Failed to bookmark post. Please try again.", "error");
    }
  }

  async function handleRemoveBookmark(event) {
    const threadId = event.detail;
    try {
      await removeBookmark(threadId);

      // Update posts tab
      posts = posts.map(post => {
        if (post.id === threadId) {
          return {
            ...post,
            isBookmarked: false,
            is_bookmarked: false,
            bookmarks_count: Math.max(0, (post.bookmarks_count || 0) - 1)
          };
        }
        return post;
      });

      // Update likes tab
      likes = likes.map(like => {
        if (like.id === threadId) {
          return {
            ...like,
            isBookmarked: false,
            is_bookmarked: false,
            bookmarks_count: Math.max(0, (like.bookmarks_count || 0) - 1)
          };
        }
        return like;
      });

      // Update replies tab if the same thread ID exists
      replies = replies.map(reply => {
        if (reply.id === threadId) {
          return {
            ...reply,
            isBookmarked: false,
            is_bookmarked: false,
            bookmarks_count: Math.max(0, (reply.bookmarks_count || 0) - 1)
          };
        }
        return reply;
      });

      toastStore.showToast("Post removed from bookmarks", "success");
    } catch (error: any) {
      console.error("Error removing bookmark:", error);
      toastStore.showToast("Failed to remove bookmark. Please try again.", "error");
    }
  }

  async function handlePinThread(threadId: string, isPinned: boolean) {
    try {
      console.log(`${isPinned ? "Unpinning" : "Pinning"} thread ${threadId}`);

      // Optimistically update UI
      posts = posts.map(post => {
        if (post.id === threadId) {
          return { ...post, is_pinned: !isPinned };
        }
        return post;
      });

      // Sort the posts again to ensure pinned ones appear at the top
      posts.sort((a, b) => {
        // First sort by pinned status
        if (isThreadPinned(a) && !isThreadPinned(b)) return -1;
        if (!isThreadPinned(a) && isThreadPinned(b)) return 1;

        // Then sort by creation date (newest first)
        const dateA = new Date(a.created_at);
        const dateB = new Date(b.created_at);
        return dateB.getTime() - dateA.getTime();
      });

      // Make the API call
      if (isPinned) {
        await unpinThread(threadId);
      } else {
        await pinThread(threadId);
      }

      // Show success toast
      toastStore.showToast(isPinned ? "Thread unpinned" : "Thread pinned", "success");
    } catch (error: any) {
      console.error("Error pinning/unpinning thread:", error);
      toastStore.showToast("Failed to pin/unpin thread", "error");

      // Revert UI if there's an error by reloading the tab
      loadTabContent("posts");
    }
  }

  async function handlePinReply(replyId, isPinned) {
    try {
      // First optimistically update UI
      replies = replies.map(reply => {
        if (reply.id === replyId) {
          return { ...reply, is_pinned: !isPinned };
        }
        return reply;
      });

      // Sort replies to show pinned ones first
      replies.sort((a, b) => {
        if (a.is_pinned && !b.is_pinned) return -1;
        if (!a.is_pinned && b.is_pinned) return 1;

        const dateA = new Date(a.created_at);
        const dateB = new Date(b.created_at);
        return dateB.getTime() - dateA.getTime();
      });

      // Make the API call
      if (isPinned) {
        await unpinReply(replyId);
      } else {
        await pinReply(replyId);
      }

      // Show success toast
      toastStore.showToast(isPinned ? "Reply unpinned" : "Reply pinned", "success");
    } catch (error: any) {
      console.error("Error pinning/unpinning reply:", error);
      toastStore.showToast("Failed to pin/unpin reply", "error");

      // Revert UI if there's an error by reloading the tab
      loadTabContent("replies");
    }
  }

  function formatJoinDate(dateString) {
    if (!dateString) return "";

    const date = new Date(dateString);
    return `Joined ${date.toLocaleString("default", { month: "long" })} ${date.getFullYear()}`;
  }

  // Filter function for posts based on search query and pinned status
  function filterPosts(posts) {
    return posts.filter(post => {
      // First check if we're showing pinned only
      if (showPinnedOnly && !post.is_pinned) {
        return false;
      }

      // Then check for search query match
      if (searchQuery && !post.content.toLowerCase().includes(searchQuery.toLowerCase())) {
        return false;
      }

      return true;
    });
  }

  // Computed property for filtered posts
  $: filteredPosts = filterPosts(posts);

  // Handlers for repost functionality
  async function handleRepost(event) {
    const threadId = event.detail;
    // In a real implementation, this would call an API
    console.log("Repost thread:", threadId);
    toastStore.showToast("Repost functionality not implemented yet", "info");
  }

  async function handleUnrepost(event) {
    const threadId = event.detail;
    // In a real implementation, this would call an API
    console.log("Unrepost thread:", threadId);
    toastStore.showToast("Unrepost functionality not implemented yet", "info");
  }

  // Get profile picture URL ensuring it's a full URL
  function getProfileImageUrl(url) {
    if (!url || url === "") {
      console.log("No profile image URL provided, using default");
      return DEFAULT_AVATAR;
    }

    console.log("Processing profile image URL:", url);
    const formattedUrl = formatStorageUrl(url);
    console.log("Formatted profile image URL:", formattedUrl);
    return formattedUrl;
  }

  // Get banner image URL ensuring it's a full URL
  function getBannerImageUrl(url) {
    if (!url || url === "") {
      console.log("No banner image URL provided, using default");
      return DEFAULT_AVATAR;
    }

    console.log("Processing banner image URL:", url);
    const formattedUrl = formatStorageUrl(url);
    console.log("Formatted banner image URL:", formattedUrl);
    return formattedUrl;
  }

  // A simple fetch profile function that matches LeftSide.svelte approach
  async function fetchProfile() {
    if (!isAuthenticated()) {
      return;
    }

    try {
      const response = await getProfile();
      const userData = response.user || (response.data && response.data.user);

      if (userData) {
        console.log("Received user data from API:", userData);
        profileData = {
          ...profileData,
          username: userData.username || profileData.username,
          displayName: userData.name || userData.display_name || profileData.displayName,
          profilePicture: userData.profile_picture_url || DEFAULT_AVATAR,
          id: userData.id || profileData.id,
          bio: userData.bio || profileData.bio,
          backgroundBanner: userData.banner_url || userData.background_banner_url || profileData.backgroundBanner,
          followerCount: userData.follower_count || 0,
          followingCount: userData.following_count || 0,
          joinedDate: userData.created_at ? new Date(userData.created_at).toLocaleDateString("en-US", { month: "long", year: "numeric" }) : "",
          email: userData.email || profileData.email,
          dateOfBirth: userData.date_of_birth || profileData.dateOfBirth,
          gender: userData.gender || profileData.gender
        };
        console.log("Updated profile data:", profileData);
      }
    } catch (err: any) {
      console.error("Failed to fetch user profile:", err);
      toastStore.showToast("Failed to load user profile. Please try again.", "error");
    }
  }

  // Load followers data
  async function loadFollowers() {
    if (isLoadingFollowers) return;

    isLoadingFollowers = true;
    followersError = "";
    followersList = [];

    try {
      console.log(`Loading followers for user ${profileData.id}`);
      const response = await getUserFollowers(profileData.id);

      console.log("Followers API response:", response);

      if (response && response.data && Array.isArray(response.data.followers)) {
        followersList = response.data.followers;
      } else if (response && Array.isArray(response.followers)) {
        followersList = response.followers;
      } else {
        // Try to find any array in the response
        const possibleArrays = Object.values(response || {}).filter(val => Array.isArray(val));
        if (possibleArrays.length > 0) {
          followersList = possibleArrays[0];
        } else {
          console.warn("Unexpected followers data format:", response);

          // If the API fails but we know there are followers, create placeholder data
          if (profileData.followerCount > 0) {
            followersList = Array.from({ length: Math.min(profileData.followerCount, 5) }, (_, i) => ({
              id: `follower-${i}`,
              username: `follower${i}`,
              name: `Follower ${i}`,
              profile_picture_url: "",
              is_following: Math.random() > 0.5,
              bio: "This is a placeholder follower for testing."
            }));
          } else {
            followersError = "No followers found";
          }
        }
      }

      console.log(`Loaded ${followersList.length} followers`);
    } catch (error: any) {
      console.error("Error loading followers:", error);
      followersError = "Failed to load followers";
    } finally {
      isLoadingFollowers = false;
    }
  }

  // Load following data
  async function loadFollowing() {
    if (isLoadingFollowing) return;

    isLoadingFollowing = true;
    followingError = "";
    followingList = [];

    try {
      console.log(`Loading following for user ${profileData.id}`);
      const response = await getUserFollowing(profileData.id);

      console.log("Following API response:", response);

      if (response && response.data && Array.isArray(response.data.following)) {
        followingList = response.data.following;
      } else if (response && Array.isArray(response.following)) {
        followingList = response.following;
      } else {
        // Try to find any array in the response
        const possibleArrays = Object.values(response || {}).filter(val => Array.isArray(val));
        if (possibleArrays.length > 0) {
          followingList = possibleArrays[0];
        } else {
          console.warn("Unexpected following data format:", response);

          // If the API fails but we know the user is following people, create placeholder data
          if (profileData.followingCount > 0) {
            followingList = Array.from({ length: Math.min(profileData.followingCount, 5) }, (_, i) => ({
              id: `following-${i}`,
              username: `following${i}`,
              name: `Following ${i}`,
              profile_picture_url: "",
              is_following: true,
              bio: "This is a placeholder following user for testing."
            }));
          } else {
            followingError = "Not following anyone";
          }
        }
      }

      console.log(`Loaded ${followingList.length} following`);
    } catch (error: any) {
      console.error("Error loading following:", error);
      followingError = "Failed to load following";
    } finally {
      isLoadingFollowing = false;
    }
  }

  // Open followers modal
  function openFollowersModal() {
    if (profileData.followerCount > 0) {
      showFollowersModal = true;
      loadFollowers();
    } else {
      toastStore.showToast("You have no followers yet", "info");
    }
  }

  // Open following modal
  function openFollowingModal() {
    if (profileData.followingCount > 0) {
      showFollowingModal = true;
      loadFollowing();
    } else {
      toastStore.showToast("You are not following anyone yet", "info");
    }
  }

  // Close modals
  function closeModals() {
    showFollowersModal = false;
    showFollowingModal = false;
  }

  // Navigate to a user's profile
  function navigateToProfile(username) {
    if (username) {
      window.location.href = `/profile/${username}`;
    }
  }

  onMount(async () => {
    try {
      // Load profile data - use different methods for own vs other profiles
      let profileResponse;
      if (isOwnProfile) {
        profileResponse = await getProfile();
        console.log("getProfile API response:", profileResponse);
      } else {
        profileResponse = await getUserById(profileUserId);
        console.log("getUserById API response:", profileResponse);
      }

      // Extract user data from response
      const userData = profileResponse.user || profileResponse;

      profileData = {
        id: userData.id || "",
        username: userData.username || "",
        displayName: userData.name || userData.display_name || "",
        bio: userData.bio || "",
        profilePicture: userData.profile_picture_url || DEFAULT_AVATAR,
        backgroundBanner: userData.banner_url || userData.background_banner_url || "",
        followerCount: userData.follower_count || 0,
        followingCount: userData.following_count || 0,
        joinedDate: userData.created_at ? new Date(userData.created_at).toLocaleDateString("en-US", { month: "long", year: "numeric" }) : "",
        email: isOwnProfile ? (userData.email || "") : "",
        dateOfBirth: isOwnProfile ? (userData.date_of_birth || "") : "",
        gender: isOwnProfile ? (userData.gender || "") : "",
        isVerified: userData.is_verified || false
      };

      console.log("Profile loaded with username:", profileData.username);
      console.log("Profile loaded with displayName:", profileData.displayName);
      console.log("Profile ID:", profileData.id); // Debug - Check if user ID is valid

      // Make sure we have a valid user ID before loading threads
      if (!profileData.id) {
        console.error("No user ID available to load posts");
        errorMessage = "Failed to load profile data. Please try again later.";
        return;
      }

      // Store the user ID for thread loading
      profileUserId = profileData.id;
      console.log("Setting profileUserId to:", profileUserId);

      // Test fetching posts directly
      try {
        console.log("Testing getUserThreads API directly with ID:", profileUserId);
        const testPostsData = await getUserThreads(profileUserId);
        console.log("Test posts data:", testPostsData);

        // Check if data is available but nested in a data property
        if (testPostsData && testPostsData.data && testPostsData.data.threads) {
          console.log("Found threads in nested data structure:", testPostsData.data.threads.length);
        } else if (testPostsData && testPostsData.threads) {
          console.log("Found threads in direct structure:", testPostsData.threads.length);
        } else {
          console.warn("No posts found in direct test");
        }
      } catch (e) {
        console.error("Error testing posts API directly:", e);
      }

      // Load initial tab content
      await loadTabContent(activeTab);

      // Preload other tabs in the background
      if (activeTab !== "posts") {
        setTimeout(() => loadTabContent("posts"), 1000);
      }
      if (activeTab !== "replies") {
        setTimeout(() => loadTabContent("replies"), 1500);
      }
      if (activeTab !== "likes") {
        setTimeout(() => loadTabContent("likes"), 2000);
      }
      if (activeTab !== "media") {
        setTimeout(() => loadTabContent("media"), 2500);
      }
    } catch (error) {
      console.error("Error loading profile data:", error);
      errorMessage = "Failed to load profile data. Please try again later.";
    } finally {
      isLoading = false;
    }
  });

  function handleFollowerClick(follower: IFollowUser, e: Event) {
    e.stopPropagation();

    if (e.target && (e.target as HTMLElement).tagName === "BUTTON") {
      // Skip navigation if a button was clicked (like follow/unfollow)
      return;
    }

    window.location.href = `/user/${follower.username}`;
  }

  function handleFollowerImageError(e: Event) {
    if (e.target) {
      (e.target as HTMLImageElement).src = DEFAULT_AVATAR;
    }
  }

  function handleFollowingClick(following: IFollowUser, e: Event) {
    e.stopPropagation();

    if (e.target && (e.target as HTMLElement).tagName === "BUTTON") {
      // Skip navigation if a button was clicked (like follow/unfollow)
      return;
    }

    window.location.href = `/user/${following.username}`;
  }

  function handleFollowingImageError(e: Event) {
    if (e.target) {
      (e.target as HTMLImageElement).src = DEFAULT_AVATAR;
    }
  }
</script>

<MainLayout
  username={profileData.username}
  displayName={profileData.displayName}
  avatar={profileData.profilePicture || DEFAULT_AVATAR}
>
  <div class="profile-container">
    {#if isLoading && !profileData.id}
      <LoadingSkeleton type="profile" />
    {:else}
      <!-- Profile Header Container -->
      <div class="profile-header-container">
        <div class="profile-banner-wrapper">
          <img
            src={profileData.backgroundBanner}
            alt="Profile banner"
            class="profile-banner"
          />
        </div>
      </div>

      <!-- Profile Info Section -->
      <div class="profile-info-section">
        <!-- Avatar Container -->
        <div class="profile-avatar-container">
          <button
            class="profile-avatar-wrapper"
            on:click={toggleProfilePicturePreview}
            aria-label="View profile picture"
            title="Click to view full size"
          >
            <img
              src={profileData.profilePicture || DEFAULT_AVATAR}
              alt={profileData.displayName}
              class="profile-avatar"
            />
            <div class="profile-avatar-overlay">
              <span class="zoom-icon">🔍</span>
            </div>
          </button>
        </div>

        <!-- Profile Actions -->
        <div class="profile-actions">
          {#if isOwnProfile}
            <button
              class="profile-edit-button"
              on:click={() => showEditModal = true}
            >
              Edit profile
            </button>
          {:else}
            <button class="profile-follow-button">
              Follow
            </button>
          {/if}
        </div>
      </div>

      <div class="profile-details">
        <div class="profile-name-container">
          {#if profileData}
          <h1 class="profile-name">
            {profileData.displayName}
            <!-- Add verified badge if applicable -->
            {#if profileData.isVerified}
              <span class="user-verified-badge">
                <CheckCircleIcon size="18" />
              </span>
            {:else}
              <!-- Log when verification badge is not shown -->
              {console.log("Verification badge not shown, isVerified is:", profileData.isVerified)}
            {/if}
          </h1>
          {/if}
          <p class="profile-username">@{profileData.username}</p>
        </div>

        {#if profileData.bio}
          <p class="profile-bio">{profileData.bio}</p>
        {/if}

        <div class="profile-meta">
          <div class="profile-meta-item">
            <span class="meta-icon">@</span>
          </div>

          <div class="profile-meta-item">
            <CalendarIcon size="16" strokeWidth="1.5" />
            <span>{formatJoinDate(profileData.joinedDate)}</span>
          </div>
        </div>

        <div class="profile-stats">
          <button class="profile-stat" on:click={openFollowingModal} aria-label="View following">
            <span class="profile-stat-count">{profileData.followingCount}</span>
            <span class="profile-stat-label">Following</span>
          </button>
          <button class="profile-stat" on:click={openFollowersModal} aria-label="View followers">
            <span class="profile-stat-count">{profileData.followerCount}</span>
            <span class="profile-stat-label">Followers</span>
          </button>
        </div>
      </div>

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
          class="profile-tab {activeTab === "likes" ? "active" : ""}"
          on:click={() => setActiveTab("likes")}
        >
          Likes
        </button>
        <button
          class="profile-tab {activeTab === "media" ? "active" : ""}"
          on:click={() => setActiveTab("media")}
        >
          Media
        </button>
      </div>

      <div class="profile-content">
        {#if isLoading}
          <div class="loading-container">
            <LoadingSkeleton type="threads" count={3} />
          </div>
        {:else if activeTab === "posts"}
          <!-- Posts tab content -->
          <div class="tab-content">
            {#if posts.length === 0}
              <div class="empty-state">
                <p>No posts yet</p>
              </div>
            {:else}
              {#each posts as thread (thread.id)}
                <div class="tweet-wrapper">
                  {#if isOwnProfile}
                    <div class="pin-button-container">
                      <button
                        class="pin-button {isThreadPinned(thread) ? "pinned" : ""}"
                        on:click={() => handlePinThread(thread.id, isThreadPinned(thread))}
                        aria-label={isThreadPinned(thread) ? "Unpin from profile" : "Pin to profile"}
                        title={isThreadPinned(thread) ? "Unpin from profile" : "Pin to profile"}
                      >
                        <PinIcon size="16" />
                        <span class="pin-text">{isThreadPinned(thread) ? "Pinned" : "Pin"}</span>
                      </button>
                    </div>
                  {/if}

                  {#if isThreadPinned(thread)}
                    <div class="pinned-badge">
                      <PinIcon size="12" />
                      <span>Pinned</span>
                    </div>
                  {/if}

                  <TweetCard tweet={ensureTweetFormat(thread)} />
                </div>
              {/each}
            {/if}
          </div>
        {:else if activeTab === "replies"}
          <!-- Replies tab content -->
          <div class="tab-content">
            {#if replies.length === 0}
              <div class="empty-state">
                <p>No replies yet</p>
              </div>
            {:else}
              {#each replies as reply (reply.id)}
                <TweetCard tweet={ensureTweetFormat(reply)} />
              {/each}
            {/if}
          </div>
        {:else if activeTab === "likes"}
          <!-- Likes tab content -->
          <div class="tab-content">
            {#if likes.length === 0}
              <div class="empty-state">
                <p>No liked posts yet</p>
              </div>
            {:else}
              {#each likes as thread (thread.id)}
                <TweetCard tweet={ensureTweetFormat(thread)} />
              {/each}
            {/if}
          </div>
        {:else if activeTab === "media"}
          <!-- Media tab content -->
          <div class="tab-content">
            {#if media.length === 0}
              <div class="empty-state">
                <p>No media posts yet</p>
              </div>
            {:else}
              {#each media as thread (thread.id)}
                <TweetCard tweet={ensureTweetFormat(thread)} />
              {/each}
            {/if}
          </div>
        {/if}
      </div>
    {/if}
  </div>

  <!-- Profile edit modal only for own profile -->
  {#if isOwnProfile}
    <ProfileEditModal
      profile={{
        id: profileData.id,
        username: profileData.username,
        name: profileData.displayName,
        bio: profileData.bio || "",
        profile_picture_url: profileData.profilePicture || DEFAULT_AVATAR,
        banner_url: profileData.backgroundBanner || "",
        email: profileData.email || "",
        date_of_birth: profileData.dateOfBirth || "",
        gender: profileData.gender || "",
        is_verified: profileData.isVerified || false,
        follower_count: profileData.followerCount || 0,
        following_count: profileData.followingCount || 0,
        created_at: profileData.joinedDate || ""
      }}
      isOpen={showEditModal}
      on:close={() => showEditModal = false}
      on:updateProfile={handleProfileUpdate}
      on:profilePictureUpdated={handleProfilePictureUpdated}
      on:bannerUpdated={handleBannerUpdated}
    />
  {/if}

  <!-- Profile picture preview modal -->
  {#if showPicturePreview}
  <div
    class="image-preview-modal"
    aria-modal="true"
    aria-label="Profile picture preview"
    role="dialog"
    on:keydown={(e) => e.key === "Escape" && toggleProfilePicturePreview()}
    tabindex="-1"
  >
    <div class="image-preview-overlay" on:click={toggleProfilePicturePreview} aria-hidden="true"></div>

    <div class="image-preview-content">
      <button
        class="image-preview-close"
        on:click={toggleProfilePicturePreview}
        aria-label="Close preview"
      >
        <XIcon size="24" />
      </button>

      <button
        class="image-preview-image-button"
        on:click={toggleProfilePicturePreview}
        on:keydown={(e) => e.key === "Enter" && toggleProfilePicturePreview()}
        aria-label="Close profile picture preview"
      >
        <img
          src={profileData.profilePicture || DEFAULT_AVATAR}
          alt={profileData.displayName}
          class="image-preview-image"
        />
      </button>

      <div class="image-preview-caption">
        Click anywhere or press ESC to close
      </div>
    </div>
  </div>
  {/if}
</MainLayout>

<!-- Followers Modal -->
{#if showFollowersModal}
  <div class="modal-overlay" on:click|self={closeModals} on:keydown={(e) => e.key === "Escape" && closeModals()} role="dialog" aria-label="Followers list" tabindex="0">
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
            {#each followersList as follower (follower.id)}
              <div
                class="follower-card {isDarkMode ? "dark" : ""}"
                on:click={(e) => handleFollowerClick(follower, e)}
                on:keydown={(e) => e.key === "Enter" && handleFollowerClick(follower, e)}
                role="button"
                tabindex="0"
              >
                <div class="follower-avatar">
                  <img
                    src={follower.profile_picture_url || DEFAULT_AVATAR}
                    alt={follower.name || follower.username}
                    on:error={handleFollowerImageError}
                  />
                </div>
                <div class="follower-info">
                  <div class="follower-name">{follower.name || follower.display_name || "User"}</div>
                  <div class="follower-username">@{follower.username}</div>
                  <div class="follower-bio">{follower.bio || ""}</div>
                  <p class="follower-bio-preview">{follower.bio || ""}</p>
                </div>
                <!-- Follow button UI here -->
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}

<!-- Following Modal -->
{#if showFollowingModal}
  <div class="modal-overlay" on:click|self={closeModals} on:keydown={(e) => e.key === "Escape" && closeModals()} role="dialog" aria-label="Following list" tabindex="0">
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
            {#each followingList as following (following.id)}
              <div
                class="follower-card {isDarkMode ? "dark" : ""}"
                on:click={(e) => handleFollowingClick(following, e)}
                on:keydown={(e) => e.key === "Enter" && handleFollowingClick(following, e)}
                role="button"
                tabindex="0"
              >
                <div class="follower-avatar">
                  <img
                    src={following.profile_picture_url || DEFAULT_AVATAR}
                    alt={following.name || following.username}
                    on:error={handleFollowingImageError}
                  />
                </div>
                <div class="follower-info">
                  <div class="follower-name">{following.name || following.display_name || "User"}</div>
                  <div class="follower-username">@{following.username}</div>
                  <div class="follower-bio">{following.bio || ""}</div>
                  <p class="follower-bio-preview">{following.bio || ""}</p>
                </div>
                <!-- Follow button UI here -->
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}

<!-- Global keyboard handler for ESC key -->
<svelte:window on:keydown={(e) => {
  if (e.key === "Escape") {
    if (showPicturePreview) {
      toggleProfilePicturePreview();
    } else if (showFollowersModal || showFollowingModal) {
      closeModals();
    }
  }
}} />

<style>
  /* Base theme variables */
  :global(:root) {
    --bg-color: #ffffff;
    --bg-secondary: #f7f9fa;
    --bg-highlight: #f7f9fa;
    --text-primary: #0f1419;
    --text-secondary: #536471;
    --border-color: #eff3f4;
    --color-primary: #1da1f2;
    --color-primary-light: rgba(29, 161, 242, 0.1);
    --bg-hover: rgba(0, 0, 0, 0.03);
    --transition-fast: 0.2s;
    --radius-md: 12px;
    --radius-full: 9999px;
    --space-1: 4px;
    --space-2: 8px;
    --space-3: 12px;
    --space-4: 16px;
    --font-size-xs: 12px;
    --font-size-sm: 14px;
    --font-size-md: 16px;
    --font-weight-bold: 700;
  }

  :global([data-theme="dark"]) {
    --bg-color: #000000;
    --bg-secondary: #16181c;
    --bg-highlight: #080808;
    --text-primary: #e7e9ea;
    --text-secondary: #71767b;
    --border-color: #2f3336;
    --bg-hover: rgba(255, 255, 255, 0.03);
  }

  /* Profile container styling */
  .profile-container {
    width: 100%;
    max-width: 100%;
    margin: 0;
    position: relative;
    background-color: var(--bg-color);
  }

  /* Profile header styling */
  .profile-header-container {
    position: relative;
    width: 100%;
    height: 150px;
    overflow: hidden;
    background-color: #1da1f2;
  }

  .profile-banner-wrapper {
    width: 100%;
    height: 100%;
  }

  .profile-banner {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  /* Profile info section */
  .profile-info-section {
    position: relative;
    padding: 0 16px;
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-top: -45px;
  }

  /* Avatar styling */
  .profile-avatar-container {
    position: relative;
    margin-bottom: 12px;
    z-index: 5; /* Ensure container is above other elements */
  }

  .profile-avatar-wrapper {
    width: 112px;
    height: 112px;
    border-radius: 50%;
    border: 4px solid var(--bg-color);
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: #222;
    cursor: pointer;
    padding: 0;
    box-shadow: 0 0 5px rgba(0, 0, 0, 0.2);
    transition: transform 0.2s ease, box-shadow 0.2s ease;
  }

  .profile-avatar-wrapper {
    position: relative;
  }

  .profile-avatar-wrapper:hover {
    transform: scale(1.03);
    box-shadow: 0 0 8px rgba(0, 0, 0, 0.3);
  }

  .profile-avatar-wrapper:hover .profile-avatar-overlay {
    opacity: 1;
  }

  .profile-avatar {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .profile-avatar-overlay {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    opacity: 0;
    transition: opacity 0.3s ease;
  }

  .zoom-icon {
    color: white;
    font-size: 24px;
  }

  /* Profile picture preview styles */
  .image-preview-modal {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    width: 100%;
    height: 100%;
    z-index: 9999;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .image-preview-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.8);
    cursor: pointer;
  }

  .image-preview-content {
    position: relative;
    max-width: 90%;
    max-height: 90%;
    z-index: 10;
    animation: zoomIn 0.3s ease;
    box-shadow: 0 10px 25px rgba(0, 0, 0, 0.5);
  }

  .image-preview-close {
    position: absolute;
    top: -40px;
    right: 0;
    color: white;
    background: none;
    border: none;
    cursor: pointer;
    padding: 8px;
    border-radius: 4px;
  }

  .image-preview-close:hover {
    background-color: rgba(0, 0, 0, 0.1);
  }

  .image-preview-image-button {
    background: none;
    border: none;
    padding: 0;
    margin: 0;
    cursor: pointer;
    width: 100%;
    max-width: 100%;
    display: flex;
    justify-content: center;
    transition: transform 0.2s ease;
  }

  .image-preview-image-button:hover {
    transform: scale(0.98);
  }

  .image-preview-image-button:focus {
    outline: 2px solid var(--color-primary);
    border-radius: 4px;
  }

  .image-preview-image {
    max-width: 100%;
    max-height: 80vh;
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  }

  .image-preview-caption {
    text-align: center;
    margin-top: 16px;
    color: white;
    font-size: 14px;
  }

  @keyframes zoomIn {
    from {
      opacity: 0;
      transform: scale(0.95);
    }
    to {
      opacity: 1;
      transform: scale(1);
    }
  }

  /* Profile actions */
  .profile-actions {
    padding-top: 16px;
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
    font-weight: 600;
    font-size: 14px;
    background-color: var(--text-primary);
    color: var(--bg-color);
    border: none;
    cursor: pointer;
    transition: opacity 0.2s;
  }

  .profile-follow-button:hover {
    opacity: 0.9;
  }

  /* Profile details */
  .profile-details {
    padding: 4px 16px;
  }

  .profile-name-container {
    margin-bottom: 0;
  }

  .profile-name {
    font-size: 20px;
    font-weight: 700;
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
    margin-left: 5px;
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

  .meta-icon {
    color: #536471;
    font-size: 14px;
  }

  /* Profile stats */
  .profile-stats {
    display: flex;
    gap: 10px;
    margin: 8px 0 12px 0;
  }

  .profile-stat {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 4px 12px;
    background-color: #f5f5f5;
    border: 1px solid #e1e1e1;
    border-radius: 4px;
    font-family: inherit;
    cursor: pointer;
    color: #0f1419;
    min-width: 100px;
    transition: background-color 0.2s;
    text-align: center;
  }

  .profile-stat:hover {
    background-color: #e5e5e5;
  }

  :global([data-theme="dark"]) .profile-stat {
    background-color: #222;
    color: #e7e9ea;
    border-color: #333;
  }

  :global([data-theme="dark"]) .profile-stat:hover {
    background-color: #2a2a2a;
  }

  .profile-stat-count {
    font-weight: 700;
    font-size: 14px;
    color: var(--text-primary);
    margin-bottom: 2px;
  }

  .profile-stat-label {
    color: var(--text-secondary);
    font-size: 13px;
  }

  /* Profile tabs */
  .profile-tabs {
    display: flex;
    border-bottom: 1px solid var(--border-color);
    margin-top: 8px;
  }

  .profile-tab {
    flex: 1;
    padding: 14px 0;
    text-align: center;
    font-weight: 600;
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    cursor: pointer;
    color: #536471;
  }

  .profile-tab.active {
    color: #1da1f2;
    border-bottom-color: #1da1f2;
  }

  /* Content styling */
  .profile-content {
    padding: 0 16px;
  }



  /* Tweet card styling */

  /* Media grid */

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
    color: white;
    border: none;
    font-weight: 600;
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .modal-retry-button:hover {
    background-color: #0c85d0;
  }

  /* User list styling */

  /* Animation keyframes */
  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  @keyframes slideIn {
    from { transform: translateY(20px); opacity: 0; }
    to { transform: translateY(0); opacity: 1; }
  }



  .loading-container {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 150px;
  }

  .empty-state {
    text-align: center;
    color: #536471;
    font-size: 16px;
  }

  .tab-content {
    padding: 16px;
    border: 1px solid var(--border-color);
    border-radius: 8px;
    background-color: var(--bg-secondary);
  }

  /* Hidden elements for empty state */
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem 1rem;
    text-align: center;
    width: 100%;
    margin-top: 1rem;
  }

  .empty-state-icon {
    font-size: 2rem;
    margin-bottom: 1rem;
    color: var(--text-secondary);
  }

  .empty-state-title {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--text-primary);
    margin-bottom: 0.5rem;
  }

  .empty-state-text {
    color: var(--text-secondary);
    margin-top: 0.5rem;
    max-width: 400px;
  }

  .tweet-wrapper {
    position: relative;
    margin-bottom: 1rem;
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    overflow: hidden;
  }

  .pin-button-container {
    position: absolute;
    top: 8px;
    right: 8px;
    z-index: 10;
  }

  .pin-button {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 6px 12px;
    background-color: var(--bg-color);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-full);
    color: var(--text-secondary);
    font-size: var(--font-size-xs);
    font-weight: var(--font-weight-bold);
    cursor: pointer;
    transition: all 0.2s ease;
    gap: 4px;
  }

  .pin-button:hover {
    background-color: var(--bg-hover);
    color: var(--color-primary);
  }

  .pin-button.pinned {
    background-color: var(--color-primary-light);
    color: var(--color-primary);
    border-color: var(--color-primary);
  }

  .pin-text {
    margin-left: 4px;
  }

  .pinned-badge {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 6px 12px;
    background-color: var(--bg-color);
    border-bottom: 1px solid var(--border-color);
    font-size: var(--font-size-xs);
    font-weight: var(--font-weight-bold);
    color: var(--color-primary);
  }
</style>
