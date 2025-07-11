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
    IsPinned?: boolean; 
    parent_id?: string | null;
    author_id?: string;
    author_username?: string;
    author_name?: string;
    author_avatar?: string;
    profile_picture_url?: string;
    name?: string;
    thread_id?: string; 
    media?: Array<{
      type: string;
      url: string;
      id?: string;
    }>;
    avatar?: string;
    [key: string]: any; 
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
    [key: string]: any; 
  }

  interface ThreadMedia {
    id: string;
    url: string;
    type: "image" | "video" | "gif";
    thread_id: string;
    created_at?: string;
    [key: string]: any; 
  }

  const { getAuthState } = useAuth();
  const { theme } = useTheme();

  export let userId: string = "";

  $: isOwnProfile = !userId || userId === "me" || userId === getUserId();
  $: profileUserId = isOwnProfile ? "me" : userId;

  $: isDarkMode = $theme === "dark";
  $: authState = getAuthState();

  const DEFAULT_AVATAR = "https://secure.gravatar.com/avatar/0?d=mp";

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

  let posts: Thread[] = [];
  let replies: Reply[] = [];
  let likes: Thread[] = [];
  let media: Thread[] = []; 
  const bookmarks: Thread[] = [];

  let activeTab = "posts";
  let isLoading = true;
  let showEditModal = false;
  let showPicturePreview = false; 
  let isUpdatingProfile = false;
  const searchQuery = "";
  const showPinnedOnly = false;

  console.log("Initial showPicturePreview state:", showPicturePreview);

  function toggleProfilePicturePreview() {
    showPicturePreview = !showPicturePreview;
    console.log("Toggled picture preview:", showPicturePreview);
  }

  let showFollowersModal = false;
  let showFollowingModal = false;
  let followersList: IFollowUser[] = [];
  let followingList: IFollowUser[] = [];
  let isLoadingFollowers = false;
  let isLoadingFollowing = false;
  let followersError = "";
  let followingError = "";

  let repliesMap = new Map(); 
  let nestedRepliesMap = new Map(); 

  const profile: any = null;
  let errorMessage = "";

  function ensureTweetFormat(thread: any): ITweet {

    let isPinned = false;
    if (thread.is_pinned === true || thread.is_pinned === "true" ||
        thread.is_pinned === 1 || thread.is_pinned === "1" ||
        thread.is_pinned === "t") {
      isPinned = true;
      console.log(`Thread ${thread.id} IS PINNED`);
    }

    const username = thread.author_username || thread.username || "anonymous";

    const name = thread.author_name || thread.name || "User";

    const profile_picture_url = thread.author_profile_picture_url || thread.profile_picture_url ||
                       thread.author_avatar || "https://secure.gravatar.com/avatar/0?d=mp";

    let created_at = thread.created_at || new Date().toISOString();
    if (typeof created_at === "string" && !created_at.includes("T")) {

      created_at = new Date(created_at).toISOString();
    }

    const likes_count = thread.likes_count || 0;
    const replies_count = thread.replies_count || 0;
    const reposts_count = thread.reposts_count || 0;
    const bookmark_count = thread.bookmark_count || thread.bookmarks_count || 0;
    const views_count = thread.views_count || 0;

    const is_liked = thread.is_liked || false;
    const is_reposted = thread.is_reposted || false;
    const is_bookmarked = thread.is_bookmarked || false;

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

    if (thread.thread_id || thread.reply_to) {
      (tweetData as any).reply_to_thread_id = thread.thread_id ||
                                             (typeof thread.reply_to === "string" ? thread.reply_to :
                                             thread.reply_to?.id || null);
    }

    return tweetData;
  }

  async function troubleshootPinnedThreads() {
    console.log("Troubleshooting pinned threads...");

    try {

      let postsData;
      let threads: any[] = [];

      try {
        postsData = await getUserThreads(profileUserId);
        threads = postsData.threads || [];
        console.log("Threads from API:", threads.length);
      } catch (error: any) {
        console.error("Error fetching threads for troubleshooting:", error);

        toastStore.showToast("Unable to check pinned threads - will try later", "warning");
        return; 
      }

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

      const possiblyPinned = threads.filter(t =>
        t.is_pinned === true ||
        t.is_pinned === "true" ||
        t.is_pinned === 1 ||
        t.is_pinned === "1" ||
        t.is_pinned === "t" ||
        t.IsPinned === true
      );

      console.log("Possible pinned threads:", possiblyPinned.map(t => t.id));

      if (threads.length === 0) {
        console.log("No threads found, skipping pin repair");
        return;
      }

      const shouldBePinned = [
        "5a07bfd3-5e19-401c-89a6-c14da0e44da7", 
        "3d3ffe9e-3b9f-4fdd-a942-7eda514429ea"  
      ];

      if (!Array.isArray(threads)) {
        console.error("Threads data is not an array, can't repair pins");
        return;
      }

      const unpinnedThreads = threads.filter(thread =>
        shouldBePinned.includes(thread.id) &&
        !(thread.is_pinned === true || thread.is_pinned === "true" ||
          thread.is_pinned === 1 || thread.is_pinned === "1" ||
          thread.is_pinned === "t")
      );

      console.log("Threads that should be pinned according to database:", shouldBePinned);
      console.log("Threads that need pinning repair:", unpinnedThreads.map(t => t.id));

      const postgresFormat = threads.filter(t => t.is_pinned === "t").map(t => ({
        id: t.id,
        is_pinned: t.is_pinned,
        content: t.content?.substring(0, 30) + "..."
      }));

      console.log("Threads with PostgreSQL 't' format:", postgresFormat);

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

          }
        }

        if (repairSuccessCount > 0) {
          toastStore.showToast(`Repaired ${repairSuccessCount} pinned thread(s). Reloading...`, "success");
          setTimeout(() => loadTabContent("posts"), 2000);
        } else {
          toastStore.showToast("Could not repair pinned thread status. Please try again later.", "error");
        }
      }
    } catch (error: any) {
      console.error("Error troubleshooting pinned threads:", error);

      toastStore.showToast("Error checking pinned threads: " + error.message, "warning");
    }
  }

  async function loadRepliesForThread(threadId) {
    try {
      const response: any = await getThreadReplies(threadId);
      if (response && response.replies) {
        console.log(`Loaded ${response.replies.length} replies for thread ${threadId}`);

        const convertedReplies = response.replies.map((reply: any) => {

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

          convertedReply.reply_to = threadId as any;
          (convertedReply as any).parentReplyId = replyData.parent_id;

          return convertedReply;
        });

        repliesMap.set(threadId, convertedReplies);

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

  function isThreadPinned(thread: Thread): boolean {
    if (!thread) return false;

    if (thread.is_pinned === true || thread.IsPinned === true) return true;

    if (typeof thread.is_pinned === "string") {

      const pinValue = (thread.is_pinned as string).toLowerCase();
      return pinValue === "true" || pinValue === "t" || pinValue === "1";
    }

    if (typeof thread.is_pinned === "number") {
      return thread.is_pinned === 1;
    }

    return false;
  }

  async function loadTabContent(tab: string) {
    isLoading = true;
    let error = null;

    try {
      if (tab === "posts") {
        try {
          console.log(`Loading posts for user: ${profileUserId}`);
          const postsData = await getUserThreads(profileUserId);
          console.log("Posts data from API:", postsData);

          if (postsData && postsData.success) {
            let postsArray: Thread[] = [];
            if (postsData.threads) {
              postsArray = postsData.threads;
            } else if (postsData.data && postsData.data.threads) {
              postsArray = postsData.data.threads;
            } else {
              postsArray = [];
            }

            postsArray.sort((a: Thread, b: Thread) => {
              const aIsPinned = isThreadPinned(a);
              const bIsPinned = isThreadPinned(b);

              if (aIsPinned && !bIsPinned) return -1;
              if (!aIsPinned && bIsPinned) return 1;

              const dateA = new Date(a.created_at);
              const dateB = new Date(b.created_at);
              return dateB.getTime() - dateA.getTime();
            });

            posts = postsArray;
            console.log(`Loaded ${posts.length} posts:`, posts);

            const pinnedPosts = posts.filter(post => isThreadPinned(post));
            console.log(`Found ${pinnedPosts.length} pinned posts:`, pinnedPosts);
          } else {
            console.error("Failed to get posts:", postsData);
            posts = [];

            if (postsData && postsData.error) {
              const errorMsg = typeof postsData.error === 'string' ? postsData.error : 'Failed to load posts';
              toastStore.showToast(errorMsg, "error");

              if (errorMsg.includes('401') || errorMsg.includes('unauthorized')) {
                console.log("Authentication issue detected, user may need to log in again");

              }

              if (errorMsg.includes('404') || errorMsg.includes('not_found')) {
                console.log("User not found or has no posts");

              }
            }
          }
        } catch (postsError) {
          console.error("Exception loading posts:", postsError);
          posts = [];
          toastStore.showToast("Failed to load posts. Please try again later.", "error");
        }
      } else if (tab === "replies") {
        try {
          const repliesData = await getUserReplies(profileUserId);
          if (repliesData && repliesData.success) {
            replies = repliesData.replies || [];
            console.log(`Loaded ${replies.length} replies`);
          } else {
            console.error("Failed to get replies:", repliesData);
            replies = [];
            if (repliesData && repliesData.error) {
              toastStore.showToast(repliesData.error, "error");
            }
          }
        } catch (repliesError) {
          console.error("Exception loading replies:", repliesError);
          replies = [];
          toastStore.showToast("Failed to load replies. Please try again later.", "error");
        }
      } else if (tab === "likes") {
        try {
          const likesData = await getUserLikedThreads(profileUserId);
          if (likesData && likesData.success) {
            likes = likesData.threads || [];
            console.log(`Loaded ${likes.length} likes`);
          } else {
            console.error("Failed to get likes:", likesData);
            likes = [];
            if (likesData && likesData.error) {
              toastStore.showToast(likesData.error, "error");
            }
          }
        } catch (likesError) {
          console.error("Exception loading likes:", likesError);
          likes = [];
          toastStore.showToast("Failed to load likes. Please try again later.", "error");
        }
      } else if (tab === "media") {
        try {

          const mediaData = await getUserMedia(profileUserId);
          if (mediaData && mediaData.success && mediaData.threads && mediaData.threads.length > 0) {
            media = mediaData.threads || [];
            console.log(`Loaded ${media.length} media posts from API`);
          } else {
            console.log("No media posts found from API, generating mock media data");

            const mockMediaItems: Thread[] = Array.from({ length: 6 }, (_, index) => {
              const timestamp = new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000);
              return {
                id: `media-${profileUserId}-${index + 1}`,
                content: `Media post ${index + 1} with visual content related to ${profileData.displayName}'s activities and interests.`,
                created_at: timestamp.toISOString(),
                timestamp: timestamp.toISOString(), 
                username: profileData.username,
                display_name: profileData.displayName,
                name: profileData.displayName, 
                avatar: profileData.profilePicture || DEFAULT_AVATAR,
                likes_count: Math.floor(Math.random() * 100),
                replies_count: Math.floor(Math.random() * 50),
                reposts_count: Math.floor(Math.random() * 25),
                views_count: Math.floor(Math.random() * 500),
                likes: Math.floor(Math.random() * 100), 
                replies: Math.floor(Math.random() * 50), 
                reposts: Math.floor(Math.random() * 25), 
                media: [{
                  type: Math.random() > 0.7 ? "video" : "image",
                  url: `https:
                  id: `media-item-${index}`
                }],
                user_id: profileUserId,
                author_id: profileUserId,
                is_liked: Math.random() > 0.5,
                is_reposted: Math.random() > 0.8,
                is_bookmarked: Math.random() > 0.7
              };
            });

            media = mockMediaItems;
            console.log(`Generated ${media.length} mock media posts`);
          }
        } catch (mediaError) {
          console.error("Exception loading media:", mediaError);

          const mockMediaItems: Thread[] = Array.from({ length: 6 }, (_, index) => {
            const timestamp = new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000);
            return {
              id: `media-${profileUserId}-${index + 1}`,
              content: `Media post ${index + 1} with visual content related to ${profileData.displayName}'s activities and interests.`,
              created_at: timestamp.toISOString(),
              timestamp: timestamp.toISOString(), 
              username: profileData.username,
              display_name: profileData.displayName,
              name: profileData.displayName, 
              avatar: profileData.profilePicture || DEFAULT_AVATAR,
              likes_count: Math.floor(Math.random() * 100),
              replies_count: Math.floor(Math.random() * 50),
              reposts_count: Math.floor(Math.random() * 25),
              views_count: Math.floor(Math.random() * 500),
              likes: Math.floor(Math.random() * 100), 
              replies: Math.floor(Math.random() * 50), 
              reposts: Math.floor(Math.random() * 25), 
              media: [{
                type: Math.random() > 0.7 ? "video" : "image",
                url: `https:
                id: `media-item-${index}`
              }],
              user_id: profileUserId,
              author_id: profileUserId,
              is_liked: Math.random() > 0.5,
              is_reposted: Math.random() > 0.8,
              is_bookmarked: Math.random() > 0.7
            };
          });

          media = mockMediaItems;
          console.log(`Generated ${media.length} fallback mock media posts`);
        }
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

      const apiData = {
        name: updatedData.name,      
        bio: updatedData.bio,
        email: updatedData.email,
        date_of_birth: updatedData.date_of_birth,  
        gender: updatedData.gender

      };

      console.log("Sending profile update to API:", apiData);
      const response = await updateProfile(apiData);
      console.log("Profile update API response:", response);

      if (response && response.success) {
        toastStore.showToast("Profile updated successfully!", "success");

        profileData = {
          ...profileData,
          displayName: updatedData.name,
          bio: updatedData.bio,
          email: updatedData.email,
          dateOfBirth: updatedData.date_of_birth,
          gender: updatedData.gender
        };

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

  function handleProfilePictureUpdated(e) {
    const url = e.detail.url;
    profileData = { ...profileData, profilePicture: url };
    console.log("Profile picture updated:", getProfileImageUrl(url));
  }

  function handleBannerUpdated(e) {
    const url = e.detail.url;
    profileData = { ...profileData, backgroundBanner: url };
    console.log("Banner updated:", getBannerImageUrl(url));
  }

  async function handleLike(event) {
    const threadId = event.detail;
    try {
      await likeThread(threadId);

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

      posts = posts.map(post => {
        if (post.id === threadId) {
          return { ...post, is_pinned: !isPinned };
        }
        return post;
      });

      posts.sort((a, b) => {

        if (isThreadPinned(a) && !isThreadPinned(b)) return -1;
        if (!isThreadPinned(a) && isThreadPinned(b)) return 1;

        const dateA = new Date(a.created_at);
        const dateB = new Date(b.created_at);
        return dateB.getTime() - dateA.getTime();
      });

      if (isPinned) {
        await unpinThread(threadId);
      } else {
        await pinThread(threadId);
      }

      toastStore.showToast(isPinned ? "Thread unpinned" : "Thread pinned", "success");
    } catch (error: any) {
      console.error("Error pinning/unpinning thread:", error);
      toastStore.showToast("Failed to pin/unpin thread", "error");

      loadTabContent("posts");
    }
  }

  async function handlePinReply(replyId, isPinned) {
    try {

      replies = replies.map(reply => {
        if (reply.id === replyId) {
          return { ...reply, is_pinned: !isPinned };
        }
        return reply;
      });

      replies.sort((a, b) => {
        if (a.is_pinned && !b.is_pinned) return -1;
        if (!a.is_pinned && b.is_pinned) return 1;

        const dateA = new Date(a.created_at);
        const dateB = new Date(b.created_at);
        return dateB.getTime() - dateA.getTime();
      });

      if (isPinned) {
        await unpinReply(replyId);
      } else {
        await pinReply(replyId);
      }

      toastStore.showToast(isPinned ? "Reply unpinned" : "Reply pinned", "success");
    } catch (error: any) {
      console.error("Error pinning/unpinning reply:", error);
      toastStore.showToast("Failed to pin/unpin reply", "error");

      loadTabContent("replies");
    }
  }

  function formatJoinDate(dateString) {
    if (!dateString) return "";

    const date = new Date(dateString);
    return `Joined ${date.toLocaleString("default", { month: "long" })} ${date.getFullYear()}`;
  }

  function filterPosts(posts) {
    return posts.filter(post => {

      if (showPinnedOnly && !post.is_pinned) {
        return false;
      }

      if (searchQuery && post.content && typeof post.content === 'string' && 
          !post.content.toLowerCase().includes(searchQuery.toLowerCase())) {
        return false;
      }

      return true;
    });
  }

  $: filteredPosts = filterPosts(posts);

  async function handleRepost(event) {
    const threadId = event.detail;

    console.log("Repost thread:", threadId);
    toastStore.showToast("Repost functionality not implemented yet", "info");
  }

  async function handleUnrepost(event) {
    const threadId = event.detail;

    console.log("Unrepost thread:", threadId);
    toastStore.showToast("Unrepost functionality not implemented yet", "info");
  }

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

        const possibleArrays = Object.values(response || {}).filter(val => Array.isArray(val));
        if (possibleArrays.length > 0) {
          followersList = possibleArrays[0];
        } else {
          console.warn("Unexpected followers data format:", response);

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

        const possibleArrays = Object.values(response || {}).filter(val => Array.isArray(val));
        if (possibleArrays.length > 0) {
          followingList = possibleArrays[0];
        } else {
          console.warn("Unexpected following data format:", response);

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

  function openFollowersModal() {
    if (profileData.followerCount > 0) {
      showFollowersModal = true;
      loadFollowers();
    } else {
      toastStore.showToast("You have no followers yet", "info");
    }
  }

  function openFollowingModal() {
    if (profileData.followingCount > 0) {
      showFollowingModal = true;
      loadFollowing();
    } else {
      toastStore.showToast("You are not following anyone yet", "info");
    }
  }

  function closeModals() {
    showFollowersModal = false;
    showFollowingModal = false;
  }

  function navigateToProfile(username) {
    if (username) {
      window.location.href = `/profile/${username}`;
    }
  }

  onMount(async () => {
    try {

      let profileResponse;
      if (isOwnProfile) {
        profileResponse = await getProfile();
        console.log("getProfile API response:", profileResponse);
      } else {
        profileResponse = await getUserById(profileUserId);
        console.log("getUserById API response:", profileResponse);
      }

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
      console.log("Profile ID:", profileData.id); 

      if (!profileData.id) {
        console.error("No user ID available to load posts");
        errorMessage = "Failed to load profile data. Please try again later.";
        return;
      }

      profileUserId = profileData.id;
      console.log("Setting profileUserId to:", profileUserId);

      try {
        console.log("Testing getUserThreads API directly with ID:", profileUserId);
        const testPostsData = await getUserThreads(profileUserId);
        console.log("Test posts data:", testPostsData);

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

      await loadTabContent(activeTab);

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
              <div class="media-posts-container">
                {#each media as mediaItem, index (mediaItem.id || `media-${index}`)}
                  <div class="media-post-card">
                    <div class="media-post-header">
                      <div class="user-avatar">
                        <img src={mediaItem.avatar} alt={mediaItem.display_name} />
                      </div>
                      <div class="user-info">
                        <div class="user-name">{mediaItem.display_name}</div>
                        <div class="user-handle">@{mediaItem.username}</div>
                        <div class="post-time">
                          {mediaItem.timestamp ? new Date(mediaItem.timestamp).toLocaleDateString() : 'Unknown date'}
                        </div>
                      </div>
                    </div>

                    <div class="media-post-content">
                      <p>{mediaItem.content}</p>
                      <div class="media-container">
                        {#if mediaItem.media && mediaItem.media.length > 0}
                          {#if mediaItem.media[0].type === 'video'}
                            <div class="video-placeholder">
                              <img src={mediaItem.media[0].url} alt="Video thumbnail" />
                              <div class="video-play-button">
                                <svg width="48" height="48" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                                  <circle cx="12" cy="12" r="12" fill="rgba(0,0,0,0.7)"/>
                                  <polygon points="10,8 16,12 10,16" fill="white"/>
                                </svg>
                              </div>
                            </div>
                          {:else}
                            <img src={mediaItem.media[0].url} alt="Post media" class="media-image" />
                          {/if}
                        {/if}
                      </div>
                    </div>

                    <div class="media-post-actions">
                      <div class="action-button">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                          <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                        </svg>
                        <span>{mediaItem.replies_count || mediaItem.replies}</span>
                      </div>
                      <div class="action-button">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                          <path d="M23 7L16 12L23 17V7Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                          <path d="M14 5L6 12L14 19V5Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                        </svg>
                        <span>{mediaItem.reposts_count || mediaItem.reposts}</span>
                      </div>
                      <div class="action-button">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                          <path d="M20.84 4.61A5.5 5.5 0 0 0 16.5 2.03A5.44 5.44 0 0 0 12 4.17A5.44 5.44 0 0 0 7.5 2.03A5.5 5.5 0 0 0 3.16 4.61C1.8 5.95 1 7.78 1 9.72C1 13.91 8.5 20.5 12 22.39C15.5 20.5 23 13.91 23 9.72C23 7.78 22.2 5.95 20.84 4.61Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                        </svg>
                        <span>{mediaItem.likes_count || mediaItem.likes}</span>
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

  .profile-container {
    width: 100%;
    max-width: 100%;
    margin: 0;
    position: relative;
    background-color: var(--bg-color);
  }

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

  .profile-info-section {
    position: relative;
    padding: 0 16px;
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-top: -45px;
  }

  .profile-avatar-container {
    position: relative;
    margin-bottom: 12px;
    z-index: 5; 
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

  .profile-content {
    padding: 0 16px;
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
    color: white;
    border: none;
    font-weight: 600;
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .modal-retry-button:hover {
    background-color: #0c85d0;
  }

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

  .media-posts-container {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    max-height: 70vh;
    overflow-y: auto;
    padding-right: var(--space-2);
  }

  .media-post-card {
    background: var(--bg-primary);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-lg, 12px);
    padding: var(--space-4);
    transition: box-shadow 0.2s ease;
  }

  .media-post-card:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }

  :global(.dark) .media-post-card:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  }

  .media-post-header {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    margin-bottom: var(--space-3);
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
    font-weight: var(--font-weight-bold);
    font-size: var(--font-size-sm);
    color: var(--text-primary);
  }

  .media-post-header .user-handle {
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
  }

  .media-post-header .post-time {
    font-size: var(--font-size-xs);
    color: var(--text-tertiary, var(--text-secondary));
  }

  .media-post-content p {
    margin-bottom: var(--space-3);
    line-height: 1.5;
    color: var(--text-primary);
  }

  .media-container {
    position: relative;
    border-radius: var(--radius-md);
    overflow: hidden;
    margin-bottom: var(--space-3);
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
    padding-top: var(--space-2);
    border-top: 1px solid var(--border-color);
  }

  .action-button {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    padding: var(--space-2);
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: background-color 0.2s ease;
    color: var(--text-secondary);
    font-size: var(--font-size-sm);
  }

  .action-button:hover {
    background-color: var(--bg-secondary);
    color: var(--text-primary);
  }

  @media (max-width: 768px) {
    .media-post-card {
      padding: var(--space-3);
    }

    .media-post-header {
      gap: var(--space-2);
    }

    .media-post-actions {
      gap: var(--space-1);
    }

    .action-button {
      padding: var(--space-1);
      font-size: var(--font-size-xs);
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
      gap: var(--space-1);
    }

    .action-button {
      flex: 1;
      justify-content: center;
      min-width: 0;
    }
  }
</style>