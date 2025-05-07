<script lang="ts">
  import { onMount } from 'svelte';
  import MainLayout from '../components/layout/MainLayout.svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import { isAuthenticated, getUserId } from '../utils/auth';
  import { getProfile, updateProfile, pinThread, unpinThread, pinReply, unpinReply } from '../api/user';
  import { getUserThreads, getUserReplies, getUserLikedThreads, getUserMedia, getThreadReplies, likeThread, unlikeThread, bookmarkThread, removeBookmark } from '../api/thread';
  import { toastStore } from '../stores/toastStore';
  import ThreadCard from '../components/explore/ThreadCard.svelte';
  import TweetCard from '../components/social/TweetCard.svelte';
  import LoadingSkeleton from '../components/common/LoadingSkeleton.svelte';
  import ProfileEditModal from '../components/profile/ProfileEditModal.svelte';
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
    displayName: string;
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
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  $: authState = getAuthState();
  
  // Profile data
  let profileData = {
    id: '',
    username: '',
    displayName: '',
    bio: '',
    profilePicture: '',
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
    // Check if we have debugging enabled
    const debug = true;
    if (debug) {
      console.log('Converting thread to tweet:', thread);
      console.log('Thread is_pinned raw value:', thread.is_pinned);
      console.log('Thread is_pinned type:', typeof thread.is_pinned);
    }
    
    // Handle is_pinned value - be flexible with the format
    let isPinned = false;
    if (thread.is_pinned === true || thread.is_pinned === 'true' || thread.is_pinned === 1 || thread.is_pinned === '1') {
      isPinned = true;
      console.log(`Thread ${thread.id} IS PINNED`);
    } else if (thread.is_pinned === 't') {
      // Handle PostgreSQL's boolean format - 't' means true
      isPinned = true;
      console.log(`Thread ${thread.id} IS PINNED (PostgreSQL 't' format)`);
    } else if (thread.IsPinned === true || thread.IsPinned === 'true' || thread.IsPinned === 1 || thread.IsPinned === '1' || thread.IsPinned === 't') {
      // Try alternate capitalization
      isPinned = true;
      console.log(`Thread ${thread.id} IS PINNED (using IsPinned property)`);
    } else {
      console.log(`Thread ${thread.id} is NOT pinned, value was:`, thread.is_pinned);
    }
    
    // Default values
    let username = 'anonymous';
    let displayName = 'User';
    let profilePicture = 'https://secure.gravatar.com/avatar/0?d=mp'; // Default avatar
    let content = thread.content || '';
    
    // Get author data from all possible locations
    // First try direct author fields
    if (thread.author_username) {
      username = thread.author_username;
    } else if (thread.authorUsername) {
      username = thread.authorUsername;
    } else if (thread.username) {
      username = thread.username;
    }
    
    if (thread.author_name) {
      displayName = thread.author_name;
    } else if (thread.authorName) {
      displayName = thread.authorName;
    } else if (thread.display_name) {
      displayName = thread.display_name;
    } else if (thread.displayName) {
      displayName = thread.displayName;
    }
    
    if (thread.author_avatar) {
      profilePicture = thread.author_avatar;
    } else if (thread.authorAvatar) {
      profilePicture = thread.authorAvatar;
    } else if (thread.profile_picture_url) {
      profilePicture = thread.profile_picture_url;
    } else if (thread.profilePictureUrl) {
      profilePicture = thread.profilePictureUrl;
    } else if (thread.avatar) {
      profilePicture = thread.avatar;
    }
    
    // Use the created_at timestamp if available, fall back to UTC now
    let timestamp = new Date().toISOString();
    if (thread.created_at) {
      timestamp = new Date(thread.created_at).toISOString();
    } else if (thread.createdAt) {
      timestamp = new Date(thread.createdAt).toISOString();
    } else if (thread.timestamp) {
      timestamp = thread.timestamp;
    }
    
    return {
      id: thread.id,
      threadId: thread.thread_id || thread.id,
      username: username,
      displayName: displayName,
      content: content,
      timestamp: timestamp,
      avatar: profilePicture,
      likes: thread.likes_count || thread.like_count || thread.metrics?.likes || 0,
      replies: thread.replies_count || thread.reply_count || thread.metrics?.replies || 0,
      reposts: thread.reposts_count || thread.repost_count || thread.metrics?.reposts || 0,
      bookmarks: thread.bookmarks_count || thread.bookmark_count || (thread.view_count > 0 ? thread.view_count : 0) || thread.metrics?.bookmarks || 0,
      views: thread.views || String(thread.views_count || '0'),
      media: thread.media || [],
      isLiked: thread.is_liked || thread.isLiked || false,
      isReposted: thread.is_repost || thread.isReposted || false,
      isBookmarked: thread.is_bookmarked || thread.isBookmarked || false,
      is_pinned: isPinned,
      replyTo: null
    };
  }
  
  // Helper function to troubleshoot and manually repair pinned threads if needed
  async function troubleshootPinnedThreads() {
    console.log("Troubleshooting pinned threads...");
    
    try {
      // Step 1: Fetch recent threads to check if any should be pinned
      const postsData = await getUserThreads('me');
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
          const replyData = reply.reply || reply;
          
          const userData = reply.user || {};
          
          const enrichedReply = {
            id: replyData.id,
            thread_id: replyData.thread_id || threadId,
            content: replyData.content || '',
            created_at: replyData.created_at || new Date().toISOString(),
            author_id: userData.id || replyData.user_id,
            author_username: userData.username || reply.author_username,
            author_name: userData.name || reply.author_name,
            author_avatar: userData.profile_picture_url || reply.author_avatar,
            parent_id: replyData.parent_id,
            is_liked: reply.is_liked || false,
            is_bookmarked: reply.is_bookmarked || false,
            likes_count: reply.likes_count || 0,
            replies_count: 0 
          };
          
          const convertedReply = ensureTweetFormat(enrichedReply);
          
          convertedReply.replyTo = threadId as any;
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
      if (tab === 'posts') {
        // Load user's tweets
        const postsData = await getUserThreads('me');
        
        // Debug: Log the raw thread data to check for pinned status
        console.log('Raw thread data from API:', postsData.threads);
        console.log('Thread pinned status sample:', postsData.threads.map(t => ({id: t.id, is_pinned: t.is_pinned})));
        
        // Convert threads and ensure proper format
        let allPosts = (postsData.threads || []).map(thread => ensureTweetFormat(thread));
        
        // Check parsed pinned status again
        console.log('Pinned posts after processing:', allPosts.filter(p => p.is_pinned === true).length);
        console.log('Pinned post IDs after processing:', allPosts.filter(p => p.is_pinned === true).map(p => p.id));
        
        // Sort the posts: pinned posts first, then by creation date
        allPosts.sort((a, b) => {
          // Debug log pinned status during sort
          console.log(`Comparing posts - A(${a.id}): ${a.is_pinned}, B(${b.id}): ${b.is_pinned}`);
          
          // First sort by pinned status (pinned posts first)
          if (a.is_pinned && !b.is_pinned) return -1;
          if (!a.is_pinned && b.is_pinned) return 1;
          
          // Then sort by creation date (newest first)
          const dateA = new Date(a.timestamp);
          const dateB = new Date(b.timestamp);
          return dateB.getTime() - dateA.getTime();
        });
        
        posts = allPosts;
        console.log(`Loaded ${posts.length} posts (${posts.filter(p => p.is_pinned).length} pinned)`);
        console.log('Pinned posts after sorting:', posts.filter(p => p.is_pinned).map(p => ({id: p.id, is_pinned: p.is_pinned})));
        
        // Additional debug: check final order of posts
        console.log('Final post order (first 5):', posts.slice(0, 5).map(p => ({id: p.id, is_pinned: p.is_pinned})));
      } 
      else if (tab === 'replies') {
        // Load user's replies
        const repliesData = await getUserReplies('me');
        replies = (repliesData.replies || []).map(reply => ensureTweetFormat(reply));
      } 
      else if (tab === 'likes') {
        // Load user's liked tweets
        const likesData = await getUserLikedThreads('me');
        likes = (likesData.threads || []).map(thread => ensureTweetFormat(thread));
      } 
      else if (tab === 'media') {
        // Load user's media posts
        const mediaData = await getUserMedia('me');
        media = mediaData.media || [];
      }
    } catch (error) {
      console.error(`Error loading ${tab} tab:`, error);
      errorMessage = `Failed to load ${tab}. Please try again later.`;
      toastStore.showToast(`Failed to load ${tab}. Please try again later.`, 'error');
    } finally {
      isLoading = false;
    }
  }
  
  async function loadProfileData() {
    isLoading = true;
    try {
      const response = await getProfile();
      if (response && response.user) {
        profileData = {
          id: response.user.id || '',
          username: response.user.username || '',
          displayName: response.user.display_name || '',
          bio: response.user.bio || '',
          profilePicture: response.user.profile_picture_url || '',
          backgroundBanner: response.user.background_banner_url || '',
          followerCount: response.user.follower_count || 0,
          followingCount: response.user.following_count || 0,
          joinedDate: response.user.created_at ? new Date(response.user.created_at).toLocaleDateString('en-US', { month: 'long', year: 'numeric' }) : '',
          email: response.user.email || '',
          dateOfBirth: response.user.date_of_birth || '',
          gender: response.user.gender || ''
        };
      }
    } catch (error) {
      console.error('Error loading profile:', error);
      toastStore.showToast('Failed to load profile data. Please try again.', 'error');
    } finally {
      isLoading = false;
      loadTabContent('posts');
    }
  }
  
  async function handleProfileUpdate(event) {
    const updatedData = event.detail;
    isUpdatingProfile = true;
    
    try {
      const response = await updateProfile(updatedData);
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
  
  async function handleLike(event) {
    const threadId = event.detail;
    try {
      await likeThread(threadId);
      
      posts = posts.map(post => {
        if (post.id === threadId) {
          return { ...post, is_liked: true, likes_count: (post.likes_count || 0) + 1 };
        }
        return post;
      });
      
      likes = likes.map(like => {
        if (like.id === threadId) {
          return { ...like, is_liked: true, likes_count: (like.likes_count || 0) + 1 };
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
      
      posts = posts.map(post => {
        if (post.id === threadId) {
          return { ...post, is_liked: false, likes_count: Math.max(0, (post.likes_count || 0) - 1) };
        }
        return post;
      });
      
      likes = likes.map(like => {
        if (like.id === threadId) {
          return { ...like, is_liked: false, likes_count: Math.max(0, (like.likes_count || 0) - 1) };
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
      
      posts = posts.map(post => {
        if (post.id === threadId) {
          return { ...post, is_bookmarked: true, bookmarks_count: (post.bookmarks_count || 0) + 1 };
        }
        return post;
      });
      
      likes = likes.map(like => {
        if (like.id === threadId) {
          return { ...like, is_bookmarked: true, bookmarks_count: (like.bookmarks_count || 0) + 1 };
        }
        return like;
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
      
      posts = posts.map(post => {
        if (post.id === threadId) {
          return { ...post, is_bookmarked: false, bookmarks_count: Math.max(0, (post.bookmarks_count || 0) - 1) };
        }
        return post;
      });
      
      likes = likes.map(like => {
        if (like.id === threadId) {
          return { ...like, is_bookmarked: false, bookmarks_count: Math.max(0, (like.bookmarks_count || 0) - 1) };
        }
        return like;
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
        const dateA = new Date(a.created_at || a.timestamp);
        const dateB = new Date(b.created_at || b.timestamp);
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
      
      // Revert UI if there's an error
      loadTabContent('posts');
    }
  }
  
  async function handlePinReply(replyId, isPinned) {
    try {
      if (isPinned) {
        await unpinReply(replyId);
      } else {
        await pinReply(replyId);
      }
      loadTabContent('replies');
      toastStore.showToast(isPinned ? 'Reply unpinned' : 'Reply pinned', 'success');
    } catch (error) {
      console.error('Error pinning/unpinning reply:', error);
      toastStore.showToast('Failed to pin/unpin reply', 'error');
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
  
  onMount(async () => {
    try {
      await loadProfileData();
      // Call troubleshooting function
      await troubleshootPinnedThreads();
      await loadTabContent(activeTab);
      console.log('UserProfile component mounted successfully');
    } catch (error) {
      console.error('Failed to load user profile:', error);
      errorMessage = 'Failed to load profile. Please try again later.';
    }
  });
</script>

<MainLayout
  username={profileData.username}
  displayName={profileData.displayName}
  avatar={profileData.profilePicture}
>
  <div class="w-full min-h-screen border-x border-gray-200 dark:border-gray-800">
    {#if isLoading && !profileData.id}
      <LoadingSkeleton type="profile" />
    {:else}
      <div class="w-full h-48 overflow-hidden relative">
        {#if profileData.backgroundBanner}
          <img 
            src={profileData.backgroundBanner} 
            alt="Banner" 
            class="w-full h-full object-cover"
          />
        {:else}
          <div class="w-full h-full bg-blue-500"></div>
        {/if}
      </div>
      
      <div class="flex justify-between px-4 -mt-16 relative z-10">
        <div class="relative">
          <button 
            class="block border-4 border-white dark:border-black rounded-full overflow-hidden cursor-pointer"
            on:click={() => showPicturePreview = true}
          >
            {#if profileData.profilePicture}
              <img 
                src={profileData.profilePicture} 
                alt={profileData.displayName} 
                class="w-32 h-32 object-cover"
              />
            {:else}
              <div class="w-32 h-32 flex items-center justify-center bg-blue-200 dark:bg-blue-700 text-4xl font-bold">
                {profileData.displayName.charAt(0).toUpperCase()}
              </div>
            {/if}
          </button>
        </div>
        
        <div class="mt-16">
          <button 
            class="px-4 py-2 rounded-full dark:bg-black border border-gray-100 dark:border-black font-bold hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
            on:click={() => showEditModal = true}
          >
            Edit profile
          </button>
        </div>
      </div>
      
      <div class="p-4 mt-2">
        <h1 class="text-xl font-bold dark:text-white">{profileData.displayName}</h1>
        <p class="text-gray-500 dark:text-gray-400">@{profileData.username}</p>
        
        {#if profileData.bio}
          <p class="my-3 dark:text-white whitespace-pre-wrap">{profileData.bio}</p>
        {/if}
        
        <div class="flex items-center mt-3 text-gray-500 dark:text-gray-400 text-sm">
          <span class="flex items-center">
            <CalendarIcon size="16" class="mr-1" />
            {formatJoinDate(profileData.joinedDate)}
          </span>
        </div>
        
        <div class="flex mt-3 gap-5">
          <a href={`/following/${profileData.id}`} class="flex items-center gap-1 hover:underline text-gray-600 dark:text-gray-300">
            <span class="font-bold text-black dark:text-white">{profileData.followingCount}</span>
            <span>Following</span>
          </a>
          <a href={`/followers/${profileData.id}`} class="flex items-center gap-1 hover:underline text-gray-600 dark:text-gray-300">
            <span class="font-bold text-black dark:text-white">{profileData.followerCount}</span>
            <span>Followers</span>
          </a>
        </div>
      </div>
      
      <div class="flex border-b border-gray-200 dark:border-gray-800">
        <button 
          class="flex-1 py-4 text-gray-500 dark:bg-black dark:text-gray-400 font-medium hover:bg-gray-50 dark:hover:bg-gray-900 relative {activeTab === 'posts' ? 'text-blue-500 font-bold' : ''}"
          on:click={() => setActiveTab('posts')}
        >
          Posts
          {#if activeTab === 'posts'}
            <div class="absolute bottom-0 left-0 w-full h-1 bg-blue-500 rounded-t"></div>
          {/if}
        </button>
        <button 
          class="flex-1 py-4 text-gray-500 dark:bg-black dark:text-gray-400 dark:bg-black font-medium hover:bg-gray-50 dark:hover:bg-gray-900 relative {activeTab === 'replies' ? 'text-blue-500 font-bold' : ''}"
          on:click={() => setActiveTab('replies')}
        >
          Replies
          {#if activeTab === 'replies'}
            <div class="absolute bottom-0 left-0 w-full h-1 bg-blue-500 rounded-t"></div>
          {/if}
        </button>
        <button 
          class="flex-1 py-4 text-gray-500 dark:bg-black dark:text-gray-400 font-medium hover:bg-gray-50 dark:hover:bg-gray-900 relative {activeTab === 'likes' ? 'text-blue-500 font-bold' : ''}"
          on:click={() => setActiveTab('likes')}
        >
          Likes
          {#if activeTab === 'likes'}
            <div class="absolute bottom-0 left-0 w-full h-1 bg-blue-500 rounded-t"></div>
          {/if}
        </button>
        <button 
          class="flex-1 py-4 text-gray-500 dark:bg-black dark:text-gray-400 font-medium hover:bg-gray-50 dark:hover:bg-gray-900 relative {activeTab === 'media' ? 'text-blue-500 font-bold' : ''}"
          on:click={() => setActiveTab('media')}
        >
          Media
          {#if activeTab === 'media'}
            <div class="absolute bottom-0 left-0 w-full h-1 bg-blue-500 rounded-t"></div>
          {/if}
        </button>
      </div>
      
            <div class="p-4">
        {#if isLoading}
          <LoadingSkeleton type="threads" count={3} />
        {:else if activeTab === 'posts'}
          {#if posts.length === 0}
            <div class="flex flex-col items-center justify-center py-8">
              <p class="text-gray-500 dark:text-gray-400">No posts yet</p>
            </div>
          {:else}
            {#each posts as post (post.id)}
              <div class="mb-4 p-4 rounded-lg border border-gray-200 dark:border-gray-800 {post.is_pinned ? 'bg-gray-50 dark:bg-gray-900 border-l-4 border-l-blue-500' : ''}">
                {#if post.is_pinned}
                  <div class="flex items-center text-blue-500 text-xs font-bold mb-2">
                    <PinIcon size="14" class="mr-1" />
                    Pinned post
                  </div>
                {/if}
                <TweetCard 
                  tweet={ensureTweetFormat(post)} 
                  {isDarkMode} 
                  isAuthenticated={!!authState.isAuthenticated}
                  isLiked={post.is_liked || false}
                  isBookmarked={post.is_bookmarked || false}
                  isReposted={post.is_reposted || false}
                  replies={repliesMap.get(post.id) || []}
                  showReplies={false}
                  nestingLevel={0}
                  nestedRepliesMap={new Map()}
                  on:reply={handleReply}
                  on:loadReplies={handleLoadReplies}
                  on:click={handleThreadClick}
                />
                <button 
                  class="mt-2 text-xs dark:bg-black text-gray-500 dark:text-gray-400 hover:text-blue-500 dark:hover:text-blue-400 hover:underline"
                  on:click={() => handlePinThread(post.id, post.is_pinned)}
                >
                  {post.is_pinned ? 'Unpin from profile' : 'Pin to profile'}
                </button>
              </div>
            {/each}
          {/if}
        {:else if activeTab === 'replies'}
          {#if replies.length === 0}
            <div class="flex flex-col items-center justify-center py-8">
              <p class="text-gray-500 dark:text-gray-400">No replies yet</p>
            </div>
          {:else}
            {#each replies as reply (reply.id)}
              <div class="mb-4 p-4 rounded-lg border border-gray-200 dark:border-gray-800 {reply.is_pinned ? 'bg-gray-50 dark:bg-gray-900' : ''}">
                {#if reply.is_pinned}
                  <div class="flex items-center text-blue-500 text-xs font-bold mb-2">
                    <PinIcon size="14" class="mr-1" />
                    Pinned
                  </div>
                {/if}
                <div class="text-sm text-blue-500 mb-2">
                  Replying to <a href={`/thread/${reply.thread_id}`} class="hover:underline">thread</a>
                </div>
                <TweetCard 
                  tweet={ensureTweetFormat(reply)} 
                  {isDarkMode} 
                  isAuthenticated={!!authState.isAuthenticated}
                  isLiked={reply.is_liked || false}
                  isBookmarked={reply.is_bookmarked || false}
                  isReposted={reply.is_reposted || false}
                  replies={repliesMap.get(reply.id) || []}
                  showReplies={false}
                  nestingLevel={0}
                  nestedRepliesMap={new Map()}
                  on:reply={handleReply}
                  on:loadReplies={handleLoadReplies}
                  on:click={handleThreadClick}
                  on:like={handleLike}
                  on:unlike={handleUnlike}
                  on:bookmark={handleBookmark}
                  on:removeBookmark={handleRemoveBookmark}
                />
                <button 
                  class="mt-2 text-xs dark:bg-black text-gray-500 dark:text-gray-400 hover:text-blue-500 dark:hover:text-blue-400 hover:underline"
                  on:click={() => handlePinReply(reply.id, reply.is_pinned)}
                >
                  {reply.is_pinned ? 'Unpin from profile' : 'Pin to profile'}
                </button>
              </div>
            {/each}
          {/if}
        {:else if activeTab === 'likes'}
          {#if likes.length === 0}
            <div class="flex flex-col items-center justify-center py-8">
              <p class="text-gray-500 dark:text-gray-400">No likes yet</p>
            </div>
          {:else}
            {#each likes as like (like.id)}
              <div class="mb-4">
                <TweetCard 
                  tweet={ensureTweetFormat(like)} 
                  {isDarkMode} 
                  isAuthenticated={!!authState.isAuthenticated}
                  isLiked={like.is_liked || true}
                  isBookmarked={like.is_bookmarked || false}
                  isReposted={like.is_reposted || false}
                  replies={repliesMap.get(like.id) || []}
                  showReplies={false}
                  nestingLevel={0}
                  nestedRepliesMap={new Map()}
                  on:reply={handleReply}
                  on:loadReplies={handleLoadReplies}
                  on:click={handleThreadClick}
                />
              </div>
            {/each}
          {/if}
        {:else if activeTab === 'media'}
          {#if media.length === 0}
            <div class="flex flex-col items-center justify-center py-8">
              <p class="text-gray-500 dark:text-gray-400">No media yet</p>
            </div>
          {:else}
            <div class="grid grid-cols-3 gap-0.5">
              {#each media as item (item.id)}
                <a href={`/thread/${item.thread_id}`} class="aspect-square overflow-hidden relative rounded-lg">
                  {#if item.type === 'image'}
                    <img src={item.url} alt="Media content" class="w-full h-full object-cover" />
                  {:else if item.type === 'video'}
                    <div class="relative w-full h-full">
                      <video src={item.url} class="w-full h-full object-cover">
                        <track kind="captions" label="English" src="" default />
                      </video>
                      <div class="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-10 h-10 bg-black/60 rounded-full flex items-center justify-center text-white">
                        <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                          <path d="M8 5.14L19 12L8 18.86V5.14Z" fill="currentColor"/>
                        </svg>
                      </div>
                    </div>
                  {:else if item.type === 'gif'}
                    <div class="relative w-full h-full">
                      <img src={item.url} alt="GIF content" class="w-full h-full object-cover" />
                      <div class="absolute bottom-2 left-2 bg-black/60 text-white text-xs font-bold px-1.5 py-0.5 rounded">
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
  
  <!-- Profile edit modal -->
  {#if showEditModal}
    <ProfileEditModal 
      profile={{
        id: profileData.id,
        username: profileData.username,
        displayName: profileData.displayName,
        bio: profileData.bio,
        profile_picture: profileData.profilePicture,
        banner: profileData.backgroundBanner,
        email: profileData.email,
        dateOfBirth: profileData.dateOfBirth,
        gender: profileData.gender,
        verified: false,
        followerCount: profileData.followerCount,
        followingCount: profileData.followingCount,
        joinDate: profileData.joinedDate
      }}
      isOpen={showEditModal}
      on:close={() => showEditModal = false}
      on:updateProfile={handleProfileUpdate}
      on:profilePictureUpdated={(e) => {
        profileData = {...profileData, profilePicture: e.detail.url};
      }}
      on:bannerUpdated={(e) => {
        profileData = {...profileData, backgroundBanner: e.detail.url};
      }}
    />
  {/if}
  
  <!-- Profile picture preview modal -->
  {#if showPicturePreview && profileData.profilePicture}
    <div 
      class="fixed inset-0 bg-black/80 flex items-center justify-center z-50" 
      on:click={() => showPicturePreview = false}
      on:keydown={(e) => e.key === 'Escape' && (showPicturePreview = false)}
      role="dialog"
      aria-modal="true"
      aria-label="Profile picture preview"
      tabindex="0"
    >
      <div 
        class="relative max-w-[90%] max-h-[90%]" 
        on:click|stopPropagation
        on:keydown|stopPropagation
        role="document"
      >
        <button 
          class="absolute -top-10 right-0 text-white p-2" 
          on:click={() => showPicturePreview = false}
          aria-label="Close preview"
        >
          <XIcon size="24" />
        </button>
        <img 
          src={profileData.profilePicture} 
          alt={profileData.displayName} 
          class="max-w-full max-h-[80vh] rounded-lg"
        />
      </div>
    </div>
  {/if}
</MainLayout>

<style>
  /* Only keeping background-related native CSS as requested */
  :global(:root) {
    --bg-color: #ffffff;
    --bg-secondary: #f7f9fa;
    --bg-highlight: #f7f9fa;
  }

  :global([data-theme="dark"]) {
    --bg-color: #000000;
    --bg-secondary: #16181c;
    --bg-highlight: #080808;
  }
</style>
