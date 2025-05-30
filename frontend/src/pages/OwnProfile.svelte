<script lang="ts">
  import { onMount } from 'svelte';
  import MainLayout from '../components/layout/MainLayout.svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import { isAuthenticated, getUserId } from '../utils/auth';
  import { getProfile, updateProfile, pinThread, unpinThread, pinReply, unpinReply, getUserById } from '../api/user';
  import { getUserThreads, getUserReplies, getUserLikedThreads, getUserMedia, getThreadReplies, likeThread, unlikeThread, bookmarkThread, removeBookmark } from '../api/thread';
  import { toastStore } from '../stores/toastStore';
  import ThreadCard from '../components/explore/ThreadCard.svelte';
  import TweetCard from '../components/social/TweetCard.svelte';
  import LoadingSkeleton from '../components/common/LoadingSkeleton.svelte';
  import ProfileEditModal from '../components/profile/ProfileEditModal.svelte';
  import { formatStorageUrl, isSupabaseStorageUrl } from '../utils/common';
  import type { ITweet } from '../interfaces/ISocialMedia';
  
  // Import Feather icons
  import CalendarIcon from 'svelte-feather-icons/src/icons/CalendarIcon.svelte';
  import XIcon from 'svelte-feather-icons/src/icons/XIcon.svelte';
  import PinIcon from 'svelte-feather-icons/src/icons/MapPinIcon.svelte';
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
    author_id?: string;
    author_username?: string;
    author_name?: string;
    author_avatar?: string;
    likes_count?: number;
    replies_count?: number;
    is_liked?: boolean;
    is_reposted?: boolean;
    is_bookmarked?: boolean;
    is_pinned?: boolean;
    media?: Array<{
      type: string;
      url: string;
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
    type: 'image' | 'video' | 'gif';
    thread_id: string;
    created_at?: string;
    [key: string]: any; // For any additional properties
  }
  
  // Auth and theme
  const { getAuthState } = useAuth();
  const { theme } = useTheme();
  
  // Get userId from URL parameter or use current user
  export let userId: string = '';
  
  // Determine if we're viewing the current user's profile or another user's profile
  $: isOwnProfile = !userId || userId === 'me' || userId === getUserId();
  $: profileUserId = isOwnProfile ? 'me' : userId;
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  $: authState = getAuthState();
  
  // Define default image URL for fallback
  const DEFAULT_AVATAR = "https://secure.gravatar.com/avatar/0?d=mp";
  
  // Profile data
  let profileData = {
    id: '',
    username: '',
    displayName: '',
    bio: '',
    profilePicture: DEFAULT_AVATAR,
    backgroundBanner: '',
    followerCount: 0,
    followingCount: 0,
    joinedDate: '',
    email: '',
    dateOfBirth: '',
    gender: ''
  };
  
  // Content data with types
  let posts: Thread[] = [];
  let replies: Reply[] = [];
  let likes: Thread[] = [];
  let media: ThreadMedia[] = [];
  
  // UI state
  let activeTab = 'posts';
  let isLoading = true;
  let showEditModal = false;
  let showPicturePreview = false;
  let isUpdatingProfile = false;
  let searchQuery = '';
  let showPinnedOnly = false;
  
  // Additional functions for thread interactions
  let repliesMap = new Map(); // Store replies for threads
  let nestedRepliesMap = new Map(); // Store nested replies
  
  // State variables
  let profile: any = null;
  let errorMessage = '';
  
  // Helper function to ensure an object has all ITweet properties
  function ensureTweetFormat(thread: any): ITweet {
    // Handle is_pinned value consistently
    let isPinned = false;
    if (thread.is_pinned === true || thread.is_pinned === 'true' || 
        thread.is_pinned === 1 || thread.is_pinned === '1' || 
        thread.is_pinned === 't') {
      isPinned = true;
      console.log(`Thread ${thread.id} IS PINNED`);
    }
    
    // Get username from all possible sources
    let username = thread.author_username || thread.username || 'anonymous';
    
    // Get display name from all possible sources
    let name = thread.author_name || thread.name || 'User';
    
    // Get profile picture from all possible sources
    let profile_picture_url = thread.author_profile_picture_url || thread.profile_picture_url || 
                       'https://secure.gravatar.com/avatar/0?d=mp';
    
    // Use the created_at timestamp if available, fall back to UTC now
    let created_at = thread.created_at || new Date().toISOString();
    if (typeof created_at === 'string' && !created_at.includes('T')) {
      // Convert to ISO format if it's not already
      created_at = new Date(created_at).toISOString();
    }
    
    // Normalize metrics
    const likes_count = thread.likes_count || 0;
    const replies_count = thread.replies_count || 0;
    const reposts_count = thread.reposts_count || 0;
    const bookmarks_count = thread.bookmarks_count || 0;
    const views_count = thread.views_count || 0;
    
    // Normalize interaction states
    const is_liked = thread.is_liked || false;
    const is_reposted = thread.is_reposted || false;
    const is_bookmarked = thread.is_bookmarked || false;

    return {
      id: thread.id,
      thread_id: thread.thread_id || thread.id,
      user_id: userId,
      username: username,
      name: name,
      content: thread.content || '',
      created_at: typeof created_at === 'string' ? created_at : new Date(created_at).toISOString(),
      profile_picture_url: profile_picture_url,
      likes_count: likes_count,
      replies_count: replies_count,
      reposts_count: reposts_count,
      bookmarks_count: bookmarks_count,
      views_count: views_count,
      media: thread.media || [],
      is_liked: is_liked,
      is_reposted: is_reposted,
      is_bookmarked: is_bookmarked,
      is_pinned: isPinned,
      reply_to: thread.parent_id ? { id: thread.parent_id } as any : null
    };
  }
  
  // Helper function to troubleshoot and manually repair pinned threads if needed
  async function troubleshootPinnedThreads() {
    console.log("Troubleshooting pinned threads...");
    
    try {
      // Step 1: Fetch recent threads to check if any should be pinned
      const postsData = await getUserThreads(profileUserId);
      const threads = postsData.threads || [];
      
      console.log("Threads from API:", threads.length);
      
      // Check different formats of is_pinned that might be present
      const pinned = {
        boolean: threads.filter(t => t.is_pinned === true).length,
        string: threads.filter(t => t.is_pinned === 'true').length,
        number: threads.filter(t => t.is_pinned === 1).length,
        stringNumber: threads.filter(t => t.is_pinned === '1').length,
        postgresTrue: threads.filter(t => t.is_pinned === 't').length,
        postgresFalse: threads.filter(t => t.is_pinned === 'f').length,
        altCase: threads.filter(t => t.IsPinned === true).length,
        nullOrUndefined: threads.filter(t => t.is_pinned === null || t.is_pinned === undefined).length
      };
      
      console.log("Pinned threads by format:", pinned);
      console.log("Possible pinned threads:", threads.filter(t => t.is_pinned === true || t.is_pinned === 'true' || 
                                                              t.is_pinned === 1 || t.is_pinned === '1' || 
                                                              t.is_pinned === 't' ||
                                                              t.IsPinned === true).map(t => t.id));
      
      // Check if the database shows it's pinned
      // This uses the raw database values we saw in the terminal output
      console.log("Database shows pinned threads - checking if they match frontend...");
      
      // Find threads that should be pinned based on our database check
      // These IDs should match what you see in the database
      const shouldBePinned = [
        "5a07bfd3-5e19-401c-89a6-c14da0e44da7", // Based on the db screenshot showing is_pinned = t
        "3d3ffe9e-3b9f-4fdd-a942-7eda514429ea"  // Based on the second row in the db showing is_pinned = t
      ];
      
      // Check if any of these threads aren't showing as pinned
      const unpinnedThreads = threads.filter(thread => 
        shouldBePinned.includes(thread.id) && 
        !(thread.is_pinned === true || thread.is_pinned === 'true' || 
          thread.is_pinned === 1 || thread.is_pinned === '1' || 
          thread.is_pinned === 't')
      );
      
      console.log("Threads that should be pinned according to database:", shouldBePinned);
      console.log("Threads that need pinning repair:", unpinnedThreads.map(t => t.id));
      
      // Check if there are any threads that have 't' as is_pinned value
      console.log("Threads with PostgreSQL 't' format:", threads.filter(t => t.is_pinned === 't').map(t => ({
        id: t.id, 
        is_pinned: t.is_pinned, 
        content: t.content?.substring(0, 30) + '...'
      })));
      
      // Fix any threads that should be pinned but aren't
      if (unpinnedThreads.length > 0) {
        console.log(`Attempting to repair pin status for ${unpinnedThreads.length} threads...`);
        
        for (const thread of unpinnedThreads) {
          try {
            console.log(`Repairing pin status for thread ${thread.id}`);
            await pinThread(thread.id);
          } catch (err) {
            console.error(`Failed to repair thread ${thread.id}:`, err);
          }
        }
        
        // Reload data after repairs
        toastStore.showToast("Repaired pinned thread status. Reloading...", "success");
        setTimeout(() => loadTabContent('posts'), 2000);
      }
    } catch (error) {
      console.error("Error troubleshooting pinned threads:", error);
    }
  }
  
  async function loadRepliesForThread(threadId) {
    try {
      const response = await getThreadReplies(threadId);
      if (response && response.replies) {
        console.log(`Loaded ${response.replies.length} replies for thread ${threadId}`);
        
        const convertedReplies = response.replies.map(reply => {
          // Standardize the reply data structure regardless of API format
          const replyData = reply.reply || reply;
          const userData = reply.user || {};
          
          const enrichedReply = {
            id: replyData.id,
            thread_id: replyData.thread_id || threadId,
            content: replyData.content || '',
            created_at: replyData.created_at || new Date().toISOString(),
            author_id: userData.id || replyData.user_id || replyData.author_id,
            author_username: userData.username || reply.author_username || reply.username,
            author_name: userData.name || userData.display_name || reply.author_name || reply.displayName,
            author_avatar: userData.profile_picture_url || reply.author_avatar || reply.avatar,
            parent_id: replyData.parent_id,
            is_liked: reply.is_liked || false,
            is_bookmarked: reply.is_bookmarked || false,
            likes_count: reply.likes_count || 0,
            replies_count: reply.replies_count || 0
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
    } catch (error) {
      console.error(`Error fetching replies for thread ${threadId}:`, error);
      toastStore.showToast('Failed to load replies. Please try again.', 'error');
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
    window.location.href = `/thread/${thread.id}`;
  }
  
  function setActiveTab(tab) {
    activeTab = tab;
    loadTabContent(tab);
  }
  
  async function loadTabContent(tab: string) {
    isLoading = true;
    try {
      // Get liked threads from localStorage for client-side verification
      let likedThreadIds: string[] = [];
      try {
        likedThreadIds = JSON.parse(localStorage.getItem('likedThreads') || '[]');
        console.log('Liked threads from localStorage:', likedThreadIds);
      } catch (err) {
        console.error('Error parsing liked threads from localStorage:', err);
      }
      
      if (tab === 'posts') {
        // Load user's threads
        const postsData = await getUserThreads(profileUserId);
        
        // Debug: Log the raw thread data
        console.log('Raw thread data from API:', postsData.threads);
        
        // Convert threads and ensure proper format
        let allPosts = (postsData.threads || []).map(thread => {
          // Check if this thread is in our liked threads
          if (likedThreadIds.includes(thread.id)) {
            thread.is_liked = true;
          }
          return ensureTweetFormat(thread);
        });
        
        // Sort: pinned posts first, then by creation date
        allPosts.sort((a, b) => {
          // First sort by pinned status
          if (a.is_pinned && !b.is_pinned) return -1;
          if (!a.is_pinned && b.is_pinned) return 1;
          
          // Then sort by creation date (newest first)
          const dateA = new Date(a.created_at);
          const dateB = new Date(b.created_at);
          return dateB.getTime() - dateA.getTime();
        });
        
        posts = allPosts;
        
        console.log('Posts processed:', posts.length);
        console.log('Pinned posts:', posts.filter(p => p.is_pinned).length);
        
      } else if (tab === 'replies') {
        // Load user's replies
        const repliesData = await getUserReplies(profileUserId);
        console.log('Raw replies data from API:', repliesData);
        
        // Ensure we have an array of replies
        replies = ((repliesData.replies || []).map(reply => {
          // Check if this reply is in our liked threads
          if (reply.id && likedThreadIds.includes(reply.id)) {
            reply.is_liked = true;
          }
          return ensureTweetFormat(reply);
        }));
        
      } else if (tab === 'likes') {
        // Load user's liked threads
        const likesData = await getUserLikedThreads(profileUserId);
        console.log('Raw likes data from API:', likesData);
        
        // Ensure we have an array of likes
        likes = ((likesData.threads || []).map(thread => {
          thread.is_liked = true; // These are liked by definition
          return ensureTweetFormat(thread);
        }));
        
      } else if (tab === 'media') {
        // Load user's media
        const mediaData = await getUserMedia(profileUserId);
        console.log('Raw media data from API:', mediaData);
        
        // Process media data
        media = (mediaData.media || []).map(item => {
          // Ensure all media items have consistent field names
          return {
            id: item.id || `media-${Math.random().toString(36).substr(2, 9)}`,
            url: item.url || item.media_url || '',
            type: item.type || item.media_type || 'image',
            thread_id: item.thread_id || '',
            created_at: item.created_at || new Date().toISOString()
          };
        });
      }
      
      // Update the active tab
      activeTab = tab;
      
      // Debug: log content counts
      console.log(`Loaded ${tab} content:`, {
        posts: tab === 'posts' ? posts.length : '(not loaded)',
        replies: tab === 'replies' ? replies.length : '(not loaded)',
        likes: tab === 'likes' ? likes.length : '(not loaded)',
        media: tab === 'media' ? media.length : '(not loaded)'
      });
      
    } catch (error: any) {
      console.error(`Error loading ${tab} content:`, error);
      toastStore.showToast(`Failed to load ${tab}. ${error.message || ''}`, 'error');
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
      } else {
        response = await getUserById(profileUserId);
      }
      
      if (response && response.user) {
        console.log('Raw profile data received:', response.user);
        
        // Extract user data exactly like in LeftSide.svelte
        const userData = response.user;
        
        profileData = {
          id: userData.id || '',
          username: userData.username || '',
          displayName: userData.name || userData.display_name || '',
          bio: userData.bio || '',
          profilePicture: userData.profile_picture_url || DEFAULT_AVATAR,
          backgroundBanner: userData.banner_url || userData.background_banner_url || '',
          followerCount: userData.follower_count || 0,
          followingCount: userData.following_count || 0,
          joinedDate: userData.created_at ? new Date(userData.created_at).toLocaleDateString('en-US', { month: 'long', year: 'numeric' }) : '',
          email: isOwnProfile ? (userData.email || '') : '',
          dateOfBirth: isOwnProfile ? (userData.date_of_birth || '') : '',
          gender: isOwnProfile ? (userData.gender || '') : ''
        };
        
        console.log('Profile loaded with avatar URL:', profileData.profilePicture);
      }
    } catch (error) {
      console.error('Error loading profile:', error);
      errorMessage = 'Failed to load profile. Please try again later.';
      toastStore.showToast('Failed to load profile. Please try again.', 'error');
    } finally {
      isLoading = false;
    }
  }
  
  async function handleProfileUpdate(event) {
    const updatedData = event.detail;
    isUpdatingProfile = true;
    
    try {
      // Map the form field names to what the backend API expects
      const apiData = {
        name: updatedData.displayName,      // Backend uses 'name' instead of 'displayName'
        bio: updatedData.bio,
        email: updatedData.email,
        date_of_birth: updatedData.dateOfBirth,  // Backend uses snake_case 'date_of_birth'
        gender: updatedData.gender
        // profile_picture_url and banner_url are handled separately via their own handlers
      };

      console.log('Sending profile update:', apiData);
      const response = await updateProfile(apiData);
      if (response && response.success) {
        toastStore.showToast('Profile updated successfully!', 'success');
        loadProfileData();
      } else {
        throw new Error(response.message || 'Failed to update profile');
      }
    } catch (error) {
      console.error('Error updating profile:', error);
      toastStore.showToast('Failed to update profile. Please try again.', 'error');
    } finally {
      isUpdatingProfile = false;
      showEditModal = false;
    }
  }
  
  // Handler for profile picture updated event
  function handleProfilePictureUpdated(e) {
    const url = e.detail.url;
    profileData = { ...profileData, profilePicture: url };
    console.log('Profile picture updated:', getProfileImageUrl(url));
  }
  
  // Handler for banner updated event
  function handleBannerUpdated(e) {
    const url = e.detail.url;
    profileData = { ...profileData, backgroundBanner: url };
    console.log('Banner updated:', getBannerImageUrl(url));
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
      
      toastStore.showToast('Post liked', 'success');
    } catch (error) {
      console.error('Error liking thread:', error);
      toastStore.showToast('Failed to like post. Please try again.', 'error');
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
      
      toastStore.showToast('Post unliked', 'success');
    } catch (error) {
      console.error('Error unliking thread:', error);
      toastStore.showToast('Failed to unlike post. Please try again.', 'error');
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
      
      toastStore.showToast('Post bookmarked', 'success');
    } catch (error) {
      console.error('Error bookmarking thread:', error);
      toastStore.showToast('Failed to bookmark post. Please try again.', 'error');
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
      
      toastStore.showToast('Post removed from bookmarks', 'success');
    } catch (error) {
      console.error('Error removing bookmark:', error);
      toastStore.showToast('Failed to remove bookmark. Please try again.', 'error');
    }
  }
  
  async function handlePinThread(threadId, isPinned) {
    try {
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
        if (a.is_pinned && !b.is_pinned) return -1;
        if (!a.is_pinned && b.is_pinned) return 1;
        
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
      toastStore.showToast(isPinned ? 'Thread unpinned' : 'Thread pinned', 'success');
    } catch (error) {
      console.error('Error pinning/unpinning thread:', error);
      toastStore.showToast('Failed to pin/unpin thread', 'error');
      
      // Revert UI if there's an error by reloading the tab
      loadTabContent('posts');
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
      toastStore.showToast(isPinned ? 'Reply unpinned' : 'Reply pinned', 'success');
    } catch (error) {
      console.error('Error pinning/unpinning reply:', error);
      toastStore.showToast('Failed to pin/unpin reply', 'error');
      
      // Revert UI if there's an error by reloading the tab
      loadTabContent('replies');
    }
  }
  
  function formatJoinDate(dateString) {
    if (!dateString) return '';
    
    const date = new Date(dateString);
    return `Joined ${date.toLocaleString('default', { month: 'long' })} ${date.getFullYear()}`;
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
    console.log('Repost thread:', threadId);
    toastStore.showToast('Repost functionality not implemented yet', 'info');
  }
  
  async function handleUnrepost(event) {
    const threadId = event.detail;
    // In a real implementation, this would call an API
    console.log('Unrepost thread:', threadId);
    toastStore.showToast('Unrepost functionality not implemented yet', 'info');
  }
  
  // Get profile picture URL ensuring it's a full URL
  function getProfileImageUrl(url) {
    if (!url || url === '') {
      console.log('No profile image URL provided, using default');
      return DEFAULT_AVATAR;
    }
    
    console.log('Processing profile image URL:', url);
    const formattedUrl = formatStorageUrl(url);
    console.log('Formatted profile image URL:', formattedUrl);
    return formattedUrl;
  }

  // Get banner image URL ensuring it's a full URL
  function getBannerImageUrl(url) {
    if (!url || url === '') {
      console.log('No banner image URL provided, using default');
      return DEFAULT_AVATAR;
    }
    
    console.log('Processing banner image URL:', url);
    const formattedUrl = formatStorageUrl(url);
    console.log('Formatted banner image URL:', formattedUrl);
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
        profileData = {
          ...profileData,
          username: userData.username || profileData.username,
          displayName: userData.name || userData.display_name || profileData.displayName,
          profilePicture: userData.profile_picture_url || DEFAULT_AVATAR,
          id: userData.id || profileData.id,
          bio: userData.bio || profileData.bio,
          backgroundBanner: userData.banner_url || userData.background_banner_url || profileData.backgroundBanner,
        };
      }
    } catch (err) {
      console.error('Failed to fetch user profile:', err);
      toastStore.showToast('Failed to load user profile. Please try again.', 'error');
    }
  }
  
  onMount(async () => {
    try {
      isLoading = true;
      
      // Load profile data
      await fetchProfile();
      
      // Load tab content
      await loadTabContent(activeTab);
      
      // Handle pinned posts if needed
      if (isOwnProfile) {
        await troubleshootPinnedThreads();
      }
    } catch (error) {
      console.error('Failed to load user profile:', error);
      errorMessage = 'Failed to load profile. Please try again later.';
    } finally {
      isLoading = false;
    }
  });
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
            src={profileData.backgroundBanner || '/assets/default-banner.jpg'} 
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
            on:click={() => showPicturePreview = true}
            aria-label="View profile picture"
          >
            <img 
              src={profileData.profilePicture || DEFAULT_AVATAR} 
              alt={profileData.displayName}
              class="profile-avatar"
            />
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
          <h1 class="profile-name">
            {profileData.displayName}
            <!-- Add verified badge if applicable -->
          </h1>
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
          <a href={`/following/${profileData.id}`} class="profile-stat">
            <span class="profile-stat-count">{profileData.followingCount}</span>
            <span class="profile-stat-label">Following</span>
          </a>
          <a href={`/followers/${profileData.id}`} class="profile-stat">
            <span class="profile-stat-count">{profileData.followerCount}</span>
            <span class="profile-stat-label">Followers</span>
          </a>
        </div>
      </div>
      
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
          class="profile-tab {activeTab === 'likes' ? 'active' : ''}"
          on:click={() => setActiveTab('likes')}
        >
          Likes
        </button>
        <button 
          class="profile-tab {activeTab === 'media' ? 'active' : ''}"
          on:click={() => setActiveTab('media')}
        >
          Media
        </button>
      </div>
      
      <div class="profile-content">
        {#if isLoading}
          <LoadingSkeleton type="threads" count={3} />
        {:else if activeTab === 'posts'}
          {#if posts.length === 0}
            <div class="profile-content-empty">
              <div class="profile-content-empty-icon">📝</div>
              <h3 class="profile-content-empty-title">No posts yet</h3>
              <p class="profile-content-empty-text">When you post, your posts will show up here.</p>
            </div>
          {:else}
            {#each posts as post (post.id)}
              <div class="tweet-card-container {post.is_pinned ? 'pinned' : ''}">
                {#if post.is_pinned}
                  <div class="pinned-indicator">
                    <PinIcon size="14" />
                    <span>Pinned post</span>
                  </div>
                {/if}
                <TweetCard 
                  tweet={ensureTweetFormat(post)} 
                  isDarkMode={isDarkMode} 
                  isAuthenticated={true}
                  isLiked={post.isLiked || post.is_liked}
                  isReposted={post.isReposted || post.is_repost}
                  isBookmarked={post.isBookmarked || post.is_bookmarked}
                  on:reply={handleReply}
                  on:repost={handleRepost}
                  on:unrepost={handleUnrepost}
                  on:like={handleLike}
                  on:unlike={handleUnlike}
                  on:bookmark={handleBookmark}
                  on:removeBookmark={handleRemoveBookmark}
                  on:loadReplies={handleLoadReplies}
                  replies={repliesMap.get(post.id) || []}
                  showReplies={false}
                  nestedRepliesMap={nestedRepliesMap}
                />
                {#if isOwnProfile}
                  <button 
                    class="pin-action-button"
                    on:click={() => handlePinThread(post.id, post.is_pinned)}
                  >
                    {post.is_pinned ? 'Unpin from profile' : 'Pin to profile'}
                  </button>
                {/if}
              </div>
            {/each}
          {/if}
        {:else if activeTab === 'replies'}
          {#if replies.length === 0}
            <div class="profile-content-empty">
              <div class="profile-content-empty-icon">💬</div>
              <h3 class="profile-content-empty-title">No replies yet</h3>
              <p class="profile-content-empty-text">When you reply to someone else's post, it will show up here.</p>
            </div>
          {:else}
            {#each replies as reply (reply.id)}
              <div class="tweet-card-container {reply.is_pinned ? 'pinned' : ''}">
                {#if reply.is_pinned}
                  <div class="pinned-indicator">
                    <PinIcon size="14" />
                    <span>Pinned reply</span>
                  </div>
                {/if}
                <div class="reply-indicator">
                  Replying to <a href={`/thread/${reply.thread_id}`}>thread</a>
                </div>
                <TweetCard 
                  tweet={ensureTweetFormat(reply)} 
                  isDarkMode={isDarkMode} 
                  isAuthenticated={true}
                  isLiked={reply.isLiked || reply.is_liked}
                  isReposted={reply.isReposted || reply.is_repost}
                  isBookmarked={reply.isBookmarked || reply.is_bookmarked}
                  on:reply={handleReply}
                  on:repost={handleRepost}
                  on:unrepost={handleUnrepost}
                  on:like={handleLike}
                  on:unlike={handleUnlike}
                  on:bookmark={handleBookmark}
                  on:removeBookmark={handleRemoveBookmark}
                  on:loadReplies={handleLoadReplies}
                  replies={repliesMap.get(reply.id) || []}
                  showReplies={false}
                  nestedRepliesMap={nestedRepliesMap}
                />
                {#if isOwnProfile}
                  <button 
                    class="pin-action-button"
                    on:click={() => handlePinReply(reply.id, reply.is_pinned)}
                  >
                    {reply.is_pinned ? 'Unpin from profile' : 'Pin to profile'}
                  </button>
                {/if}
              </div>
            {/each}
          {/if}
        {:else if activeTab === 'likes'}
          {#if likes.length === 0}
            <div class="profile-content-empty">
              <div class="profile-content-empty-icon">❤️</div>
              <h3 class="profile-content-empty-title">No likes yet</h3>
              <p class="profile-content-empty-text">When you like a post, it will show up here.</p>
            </div>
          {:else}
            {#each likes as like (like.id)}
              <div class="tweet-card-container">
                <TweetCard 
                  tweet={ensureTweetFormat(like)} 
                  isDarkMode={isDarkMode} 
                  isAuthenticated={true}
                  isLiked={like.isLiked || like.is_liked}
                  isReposted={like.isReposted || like.is_repost}
                  isBookmarked={like.isBookmarked || like.is_bookmarked}
                  on:reply={handleReply}
                  on:repost={handleRepost}
                  on:unrepost={handleUnrepost}
                  on:like={handleLike}
                  on:unlike={handleUnlike}
                  on:bookmark={handleBookmark}
                  on:removeBookmark={handleRemoveBookmark}
                  on:loadReplies={handleLoadReplies}
                  replies={repliesMap.get(like.id) || []}
                  showReplies={false}
                  nestedRepliesMap={nestedRepliesMap}
                />
              </div>
            {/each}
          {/if}
        {:else if activeTab === 'media'}
          {#if media.length === 0}
            <div class="profile-content-empty">
              <div class="profile-content-empty-icon">📷</div>
              <h3 class="profile-content-empty-title">No media yet</h3>
              <p class="profile-content-empty-text">When you post photos or videos, they will show up here.</p>
            </div>
          {:else}
            <div class="media-grid">
              {#each media as item (item.id)}
                <a href={`/thread/${item.thread_id}`} class="media-grid-item">
                  {#if item.type === 'image'}
                    <img src={item.url} alt="Media content" class="media-image" loading="lazy" />
                  {:else if item.type === 'video'}
                    <div class="media-video-container">
                      <video src={item.url} class="media-video">
                        <track kind="captions" label="English" src="" default />
                      </video>
                      <div class="media-video-play-button">
                        <svg class="media-video-play-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                          <path d="M8 5.14L19 12L8 18.86V5.14Z" fill="currentColor"/>
                        </svg>
                      </div>
                    </div>
                  {:else if item.type === 'gif'}
                    <div class="media-gif-container">
                      <img src={item.url} alt="GIF content" class="media-image" loading="lazy" />
                      <div class="media-gif-indicator">
                        GIF
                      </div>
                    </div>
                  {/if}
                </a>
              {/each}
            </div>
          {/if}
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
        bio: profileData.bio,
        profile_picture_url: profileData.profilePicture,
        banner_url: profileData.backgroundBanner,
        email: profileData.email,
        date_of_birth: profileData.dateOfBirth,
        gender: profileData.gender,
        is_verified: false,
        follower_count: profileData.followerCount,
        following_count: profileData.followingCount,
        created_at: profileData.joinedDate
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
  <dialog 
    open
    class="fixed inset-0 p-0 m-0 w-full h-full bg-transparent z-50 flex items-center justify-center" 
    aria-modal="true"
    aria-label="Profile picture preview"
  >
    <div class="fixed inset-0 bg-black/80" on:click={() => showPicturePreview = false} aria-hidden="true"></div>
    
    <div class="relative max-w-[90%] max-h-[90%] z-10">
      <button 
        class="absolute -top-10 right-0 text-white p-2 rounded hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-white" 
        on:click={() => showPicturePreview = false}
        aria-label="Close preview"
      >
        <XIcon size="24" />
      </button>
      
      <img 
        src={profileData.profilePicture || DEFAULT_AVATAR} 
        alt={profileData.displayName} 
        class="max-w-full max-h-[80vh] rounded-lg"
      />
    </div>
  </dialog>
  {/if}
</MainLayout>

<!-- Global keyboard handler for ESC key -->
<svelte:window on:keydown={(e) => {
  if (e.key === 'Escape') {
    if (showPicturePreview) {
      showPicturePreview = false;
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
  }

  .profile-avatar {    width: 100%;
    height: 100%;
    object-fit: cover;
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
    gap: 20px;
    margin: 8px 0 12px 0;
    color: #536471;
  }

  .profile-stat {
    display: flex;
    gap: 4px;
    text-decoration: none;
    color: inherit;
  }

  .profile-stat:hover {
    text-decoration: underline;
  }

  .profile-stat-count {
    font-weight: 700;
    color: var(--text-primary);
  }

  .profile-stat-label {
    color: #536471;
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

  .profile-content-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 40px 0;
    text-align: center;
  }

  .profile-content-empty-icon {
    font-size: 32px;
    margin-bottom: 16px;
  }

  .profile-content-empty-title {
    font-size: 20px;
    font-weight: 700;
    margin: 0 0 8px 0;
    color: var(--text-primary);
  }

  .profile-content-empty-text {
    font-size: 15px;
    color: #536471;
    max-width: 300px;
  }

  /* Tweet card styling */
  .tweet-card-container {
    border-bottom: 1px solid var(--border-color);
    padding: 12px 0;
  }

  .tweet-card-container.pinned {
    background-color: rgba(0, 0, 0, 0.02);
  }

  .pinned-indicator {
    display: flex;
    align-items: center;
    gap: 4px;
    color: #536471;
    font-size: 13px;
    margin-bottom: 4px;
  }

  .reply-indicator {
    font-size: 13px;
    color: #536471;
    margin-bottom: 4px;
  }

  .reply-indicator a {
    color: #1da1f2;
    text-decoration: none;
  }

  .pin-action-button {
    margin-top: 8px;
    padding: 6px 12px;
    background: none;
    border: none;
    font-size: 13px;
    color: #1da1f2;
    cursor: pointer;
  }

  .pin-action-button:hover {
    text-decoration: underline;
  }

  /* Media grid */
  .media-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
    gap: 8px;
  }

  .media-grid-item {
    aspect-ratio: 1/1;
    overflow: hidden;
    border-radius: 8px;
    display: block;
  }

  .media-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .media-video-container, 
  .media-gif-container {
    position: relative;
    width: 100%;
    height: 100%;
  }

  .media-video {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .media-video-play-button {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    background-color: rgba(0, 0, 0, 0.6);
    border-radius: 50%;
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .media-video-play-icon {
    width: 20px;
    height: 20px;
    color: white;
  }

  .media-gif-indicator {
    position: absolute;
    top: 8px;
    left: 8px;
    background-color: rgba(0, 0, 0, 0.6);
    color: white;
    font-size: 12px;
    font-weight: 700;
    padding: 2px 6px;
    border-radius: 4px;
  }
</style>
