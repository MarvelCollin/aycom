<script lang="ts">  import { onMount } from 'svelte';
  import MainLayout from '../components/layout/MainLayout.svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import { isAuthenticated, getUserId } from '../utils/auth';
  import { getUserById, followUser, unfollowUser, reportUser, blockUser, unblockUser } from '../api/user';
  import { getUserThreads, getUserReplies, getUserMedia } from '../api/thread';
  import { toastStore } from '../stores/toastStore';
  import ThreadCard from '../components/explore/ThreadCard.svelte';
  import TweetCard from '../components/social/TweetCard.svelte';
  import LoadingSkeleton from '../components/common/LoadingSkeleton.svelte';
  import type { ITweet } from '../interfaces/ISocialMedia';
  import { page } from '../stores/routeStore';
  import { createLoggerWithPrefix } from '../utils/logger';
  
  const logger = createLoggerWithPrefix('OtherProfile');

  // Import Feather icons
  import CalendarIcon from 'svelte-feather-icons/src/icons/CalendarIcon.svelte';
  import XIcon from 'svelte-feather-icons/src/icons/XIcon.svelte';
  import UserIcon from 'svelte-feather-icons/src/icons/UserIcon.svelte';
  import FlagIcon from 'svelte-feather-icons/src/icons/FlagIcon.svelte';
  import ShieldIcon from 'svelte-feather-icons/src/icons/ShieldIcon.svelte';
  import SlashIcon from 'svelte-feather-icons/src/icons/SlashIcon.svelte';
  import AlertCircleIcon from 'svelte-feather-icons/src/icons/AlertCircleIcon.svelte';
  
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
  
  // Get userId from URL parameter
  export let userId: string;
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  $: authState = getAuthState();
  $: currentUserId = getUserId();
  
  $: {
    // Log whenever userId changes
    logger.debug(`Profile rendering for userId: ${userId}, currentUserId: ${currentUserId}`);
  }
  
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
    isPrivate: false,
    isFollowing: false,
    isBlocked: false
  };
  
  // Content data with types
  let posts: Thread[] = [];
  let replies: Reply[] = [];
  let media: ThreadMedia[] = [];
    // UI state
  let activeTab = 'posts';
  let isLoading = true;
  let isFollowLoading = false;
  let isBlockLoading = false;
  let showReportModal = false;
  let reportReason = '';
  let showBlockConfirmModal = false;
  let showActionsDropdown = false;
  let errorMessage = '';
  
  
  // Additional functions for thread interactions
  let repliesMap = new Map(); // Store replies for threads
  let nestedRepliesMap = new Map(); // Store nested replies
  
  // Helper function to ensure an object has all ITweet properties
  function ensureTweetFormat(thread: any): ITweet {
    // Get username from all possible sources
    let username = thread.author_username || thread.authorUsername || thread.username || 'anonymous';
    
    // Get display name from all possible sources
    let displayName = thread.author_name || thread.authorName || thread.display_name || 
                      thread.displayName || 'User';
    
    // Get profile picture from all possible sources
    let profilePicture = thread.author_avatar || thread.authorAvatar || 
                        thread.profile_picture_url || thread.profilePictureUrl || 
                        thread.avatar || 'https://secure.gravatar.com/avatar/0?d=mp';
    
    // Use the created_at timestamp if available, fall back to UTC now
    let timestamp = thread.created_at || thread.createdAt || thread.timestamp || new Date().toISOString();
    if (typeof timestamp === 'string' && !timestamp.includes('T')) {
      // Convert to ISO format if it's not already
      timestamp = new Date(timestamp).toISOString();
    }
    
    // Normalize metrics
    const likes = thread.likes_count || thread.like_count || thread.metrics?.likes || 0;
    const replies = thread.replies_count || thread.reply_count || thread.metrics?.replies || 0;
    const reposts = thread.reposts_count || thread.repost_count || thread.metrics?.reposts || 0;
    const bookmarks = thread.bookmarks_count || thread.bookmark_count || thread.metrics?.bookmarks || 0;
    const views = Number(thread.views || thread.views_count || 0);
    
    // Normalize interaction states
    const isLiked = thread.is_liked || thread.isLiked || false;
    const isReposted = thread.is_repost || thread.isReposted || false;
    const isBookmarked = thread.is_bookmarked || thread.isBookmarked || false;
      
  return {
      id: thread.id,
      threadId: thread.thread_id || thread.id,
      userId: thread.user_id || thread.userId || thread.author_id || thread.authorId || thread.id,
      username: username,
      displayName: displayName,
      content: thread.content || '',
      timestamp: typeof timestamp === 'string' ? timestamp : new Date(timestamp).toISOString(),
      avatar: profilePicture,
      likes: likes,
      replies: replies,
      reposts: reposts,
      bookmarks: bookmarks,
      views: views,
      media: thread.media || [],
      isLiked: isLiked,
      isReposted: isReposted,
      isBookmarked: isBookmarked,
      isPinned: thread.is_pinned === true || thread.is_pinned === 'true' || thread.is_pinned === 1 || thread.is_pinned === '1' || thread.is_pinned === 't' || thread.IsPinned === true || false,
      replyTo: thread.parent_id ? { id: thread.parent_id } as any : null
    };
  }

  function handleReply(event) {
    const threadId = event.detail;
    window.location.href = `/thread/${threadId}`;
  }

  function setActiveTab(tab) {
    activeTab = tab;
    loadTabContent(tab);
  }
  
  async function loadTabContent(tab: string) {
    isLoading = true;
    try {
      // Check if user is private and not following
      if (profileData.isPrivate && !profileData.isFollowing) {
        isLoading = false;
        return;
      }
      
      if (tab === 'posts') {
        // Load user's threads
        logger.debug(`Loading posts for user ${userId}`);
        const postsData = await getUserThreads(userId);
        
        // Convert threads and ensure proper format
        posts = (postsData.threads || []).map(thread => ensureTweetFormat(thread));
        logger.debug(`Loaded ${posts.length} posts`);
        
        // Sort by creation date (newest first)
        posts.sort((a, b) => {
          const dateA = new Date(a.timestamp);
          const dateB = new Date(b.timestamp);
          return dateB.getTime() - dateA.getTime();
        });
      } 
      else if (tab === 'replies') {
        // Load user's replies
        logger.debug(`Loading replies for user ${userId}`);
        const repliesData = await getUserReplies(userId);
        replies = (repliesData.replies || []).map(reply => ensureTweetFormat(reply));
        logger.debug(`Loaded ${replies.length} replies`);
        
        // Sort replies by date (newest first)
        replies.sort((a, b) => {
          const dateA = new Date(a.timestamp);
          const dateB = new Date(b.timestamp);
          return dateB.getTime() - dateA.getTime();
        });
      } 
      else if (tab === 'media') {
        // Load user's media posts
        logger.debug(`Loading media for user ${userId}`);
        const mediaData = await getUserMedia(userId);
        
        // Ensure media items have all required fields
        media = (mediaData.media || []).map(item => ({
          id: item.id || `media-${Math.random().toString(36).substr(2, 9)}`,
          url: item.url || '',
          type: item.type || 'image',
          thread_id: item.thread_id || '',
          created_at: item.created_at || new Date().toISOString()
        }));
        logger.debug(`Loaded ${media.length} media items`);
        
        // Sort by date (newest first)
        media.sort((a, b) => {
          const dateA = new Date(a.created_at || '');
          const dateB = new Date(b.created_at || '');
          return dateB.getTime() - dateA.getTime();
        });
      }
    } catch (error) {
      logger.error(`Error loading ${tab} tab:`, error);
      errorMessage = `Failed to load ${tab}. Please try again later.`;
      toastStore.showToast(`Failed to load ${tab}. Please try again later.`, 'error');
    } finally {
      isLoading = false;
    }
  }
  
  async function loadProfileData() {
    isLoading = true;
    try {
      logger.debug(`Loading profile data for userId: ${userId}`);
      const response = await getUserById(userId);
      
      if (response && response.user) {
        profileData = {
          id: response.user.id || '',
          username: response.user.username || '',
          displayName: response.user.display_name || '',
          bio: response.user.bio || '',
          profilePicture: response.user.profile_picture_url || '',
          backgroundBanner: response.user.banner_url || response.user.background_banner_url || '',
          followerCount: response.user.follower_count || 0,
          followingCount: response.user.following_count || 0,
          joinedDate: response.user.created_at ? new Date(response.user.created_at).toLocaleDateString('en-US', { month: 'long', year: 'numeric' }) : '',
          isPrivate: response.user.is_private || false,
          isFollowing: response.user.is_following || false,
          isBlocked: response.user.is_blocked || false
        };
        logger.debug(`Profile data loaded: ${profileData.displayName} (@${profileData.username})`);
      } else {
        logger.error('User profile not found in response:', response);
        errorMessage = 'User not found';
      }
    } catch (error) {
      logger.error('Error loading profile:', error);
      errorMessage = 'Failed to load profile data.';
      toastStore.showToast('Failed to load profile data. Please try again.', 'error');
    } finally {
      isLoading = false;
      if (!errorMessage) {
        loadTabContent('posts');
      }
    }
  }
  
  async function handleFollow() {
    isFollowLoading = true;
    try {
      let success;
      if (profileData.isFollowing) {
        success = await unfollowUser(profileData.id);
      } else {
        success = await followUser(profileData.id);
      }
      
      if (success) {
        profileData = {
          ...profileData,
          isFollowing: !profileData.isFollowing,
          followerCount: profileData.isFollowing 
            ? Math.max(0, profileData.followerCount - 1) 
            : profileData.followerCount + 1
        };
        
        toastStore.showToast(
          profileData.isFollowing ? 'You are now following this user' : 'You have unfollowed this user',
          'success'
        );
      } else {
        throw new Error('Failed to update follow status');
      }
    } catch (error) {
      console.error('Error updating follow status:', error);
      toastStore.showToast('Failed to update follow status. Please try again.', 'error');
    } finally {
      isFollowLoading = false;
    }
  }
  
  async function submitReport() {
    if (!reportReason.trim()) {
      toastStore.showToast('Please provide a reason for the report', 'error');
      return;
    }
      try {
      const result = await reportUser(profileData.id, reportReason);
      
      if (result) {
        toastStore.showToast('User reported successfully. Our team will review this report.', 'success');
        showReportModal = false;
        reportReason = '';
      } else {
        throw new Error('Failed to report user');
      }
    } catch (error) {
      console.error('Error reporting user:', error);
      toastStore.showToast('Failed to report user. Please try again.', 'error');
    }
  }
  
  async function handleBlockUser() {
    isBlockLoading = true;
    try {
      let success;
      if (profileData.isBlocked) {
        success = await unblockUser(profileData.id);
      } else {
        success = await blockUser(profileData.id);
      }
      
      if (success) {
        profileData = {
          ...profileData,
          isBlocked: !profileData.isBlocked
        };
        
        toastStore.showToast(
          profileData.isBlocked ? 'User blocked successfully' : 'User unblocked successfully',
          'success'
        );
        
        // Close the modal
        showBlockConfirmModal = false;
        showActionsDropdown = false;
        
        // If we just blocked the user and they were being followed, simulate unfollow
        if (profileData.isBlocked && profileData.isFollowing) {
          profileData = {
            ...profileData,
            isFollowing: false,
            followerCount: Math.max(0, profileData.followerCount - 1)
          };
        }
      } else {
        throw new Error('Failed to update block status');
      }
    } catch (error) {
      console.error('Error updating block status:', error);
      toastStore.showToast('Failed to update block status. Please try again.', 'error');
    } finally {
      isBlockLoading = false;
    }
  }
  
  function formatJoinDate(dateString) {
    if (!dateString) return '';
    
    const date = new Date(dateString);
    return `Joined ${date.toLocaleString('default', { month: 'long' })} ${date.getFullYear()}`;
  }
  
  onMount(async () => {
    // Check if user is authenticated
    if (!isAuthenticated()) {
      logger.warn('User not authenticated, redirecting to login');
      window.location.href = '/login';
      return;
    }
    
    logger.debug('OtherProfile component mounted with userId:', userId);
    
    // Only load profile if we have a userId
    if (userId) {
      await loadProfileData();
    } else {
      logger.error('No userId provided');
      errorMessage = 'Invalid user ID';
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
          <div class="block border-4 border-white dark:border-black rounded-full overflow-hidden">
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
          </div>
        </div>
        
        <div class="mt-16 flex items-center space-x-2">
          <button 
            class="px-4 py-2 rounded-full font-bold transition-colors {profileData.isFollowing ? 'bg-transparent border border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-800' : 'bg-black dark:bg-white text-white dark:text-black hover:bg-gray-800 dark:hover:bg-gray-200'}"
            on:click={handleFollow}
            disabled={isFollowLoading}
          >
            {#if isFollowLoading}
              <span class="flex items-center justify-center">
                <svg class="animate-spin h-4 w-4 mr-1" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
              </span>
            {:else}
              {profileData.isFollowing ? 'Following' : 'Follow'}
            {/if}
          </button>
          
          <div class="relative">
            <button 
              class="p-2 rounded-full border border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
              on:click={() => showActionsDropdown = !showActionsDropdown}
              title="More actions"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width={2} d="M4 6h16M4 12h16m-7 6h7" />
              </svg>
            </button>
            
            {#if showActionsDropdown}
              <div class="absolute right-0 mt-2 w-48 bg-white dark:bg-gray-800 rounded-lg shadow-lg z-10">
                <button 
                  class="flex items-center w-full px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700"
                  on:click={() => {
                    showBlockConfirmModal = true;
                    showActionsDropdown = false;
                  }}
                >
                  <ShieldIcon class="w-5 h-5 mr-2" />
                  {profileData.isBlocked ? 'Unblock User' : 'Block User'}
                </button>
                
                <button 
                  class="flex items-center w-full px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700"
                  on:click={() => showReportModal = true}
                >
                  <FlagIcon class="w-5 h-5 mr-2" />
                  Report User
                </button>
              </div>
            {/if}
          </div>
        </div>
      </div>
      
      <div class="p-4 mt-2">
        <h1 class="text-xl font-bold dark:text-white">{profileData.displayName}</h1>
        <p class="text-gray-500 dark:text-gray-400">@{profileData.username}</p>
        
        {#if profileData.isPrivate && !profileData.isFollowing}
          <div class="mt-3 p-3 bg-gray-100 dark:bg-gray-800 rounded-lg">
            <p class="flex items-center">
              <UserIcon size="16" class="mr-2" />
              <span>This account is private. Follow this user to see their posts.</span>
            </p>
          </div>
        {:else if profileData.bio}
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
          class="flex-1 py-4 text-gray-500 dark:bg-black dark:text-gray-400 font-medium hover:bg-gray-50 dark:hover:bg-gray-900 relative {activeTab === 'replies' ? 'text-blue-500 font-bold' : ''}"
          on:click={() => setActiveTab('replies')}
        >
          Replies
          {#if activeTab === 'replies'}
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
        {#if profileData.isPrivate && !profileData.isFollowing}
          <div class="flex flex-col items-center justify-center py-8">
            <div class="bg-gray-100 dark:bg-gray-800 p-6 rounded-lg text-center">
              <UserIcon size="48" class="mx-auto mb-4 text-gray-500" />
              <p class="text-gray-700 dark:text-gray-300 font-medium">This account is private</p>
              <p class="text-gray-500 dark:text-gray-400 mt-2">
                Follow this user to see their posts, replies, and media.
              </p>
              <button 
                class="mt-4 px-4 py-2 bg-blue-500 text-white rounded-full hover:bg-blue-600 font-medium"
                on:click={handleFollow}
                disabled={isFollowLoading}
              >
                {isFollowLoading ? 'Processing...' : 'Follow'}
              </button>
            </div>
          </div>
        {:else if isLoading}
          <LoadingSkeleton type="threads" count={3} />
        {:else if activeTab === 'posts'}
          {#if posts.length === 0}
            <div class="flex flex-col items-center justify-center py-8">
              <p class="text-gray-500 dark:text-gray-400">No posts yet</p>
            </div>
          {:else}
            {#each posts as post (post.id)}
              <div class="mb-4 p-4 rounded-lg border border-gray-200 dark:border-gray-800">
                <TweetCard 
                  tweet={ensureTweetFormat(post)} 
                  isDarkMode={isDarkMode} 
                  isAuthenticated={true}
                  isLiked={post.isLiked || post.is_liked}
                  isReposted={post.isReposted || post.is_repost}
                  isBookmarked={post.isBookmarked || post.is_bookmarked}
                  on:reply={handleReply}
                />
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
              <div class="mb-4 p-4 rounded-lg border border-gray-200 dark:border-gray-800">
                <div class="text-sm text-blue-500 mb-2">
                  Replying to <a href={`/thread/${reply.thread_id}`} class="hover:underline">thread</a>
                </div>
                <TweetCard 
                  tweet={ensureTweetFormat(reply)} 
                  isDarkMode={isDarkMode} 
                  isAuthenticated={true}
                  isLiked={reply.isLiked || reply.is_liked}
                  isReposted={reply.isReposted || reply.is_repost}
                  isBookmarked={reply.isBookmarked || reply.is_bookmarked}
                  on:reply={handleReply}
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
            <div class="grid grid-cols-3 gap-1 sm:gap-2 md:gap-3">
              {#each media as item (item.id)}
                <a href={`/thread/${item.thread_id}`} class="aspect-square overflow-hidden relative rounded-lg border border-gray-200 dark:border-gray-800">
                  {#if item.type === 'image'}
                    <img src={item.url} alt="Media content" class="w-full h-full object-cover" loading="lazy" />
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
                      <img src={item.url} alt="GIF content" class="w-full h-full object-cover" loading="lazy" />
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
  
  <!-- Report Modal -->
  {#if showReportModal}
    <div 
      class="fixed inset-0 bg-black/60 flex items-center justify-center z-50" 
      on:click={() => showReportModal = false}
      on:keydown={(e) => e.key === 'Escape' && (showReportModal = false)}
      role="dialog"
      aria-modal="true"
      aria-labelledby="modal-title"
      tabindex="-1"
    >
      <div 
        class="bg-white dark:bg-gray-900 rounded-xl w-full max-w-lg mx-4 p-6"
        on:click|stopPropagation
        role="document"
        tabindex="0"
      >
        <h2 id="modal-title" class="text-xl font-bold mb-4 dark:text-white">Report @{profileData.username}</h2>
        <p class="mb-4 text-gray-600 dark:text-gray-300">Please tell us why you're reporting this account.</p>
        
        <textarea
          bind:value={reportReason}
          class="w-full p-3 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white mb-4"
          rows="4"
          placeholder="Describe the issue..."
        ></textarea>
        
        <div class="flex justify-end space-x-3">
          <button 
            class="px-4 py-2 border border-gray-300 dark:border-gray-700 rounded-full text-gray-700 dark:text-gray-300"
            on:click={() => showReportModal = false}
          >
            Cancel
          </button>
          <button 
            class="px-4 py-2 bg-red-500 text-white rounded-full hover:bg-red-600"
            on:click={submitReport}
          >
            Submit Report
          </button>
        </div>
      </div>
    </div>
  {/if}
  
  <!-- Block Confirmation Modal -->
  {#if showBlockConfirmModal}
    <div 
      class="fixed inset-0 bg-black/60 flex items-center justify-center z-50" 
      on:click={() => showBlockConfirmModal = false}
      on:keydown={(e) => e.key === 'Escape' && (showBlockConfirmModal = false)}
      role="dialog"
      aria-modal="true"
      aria-labelledby="block-modal-title"
      tabindex="-1"
    >
      <div 
        class="bg-white dark:bg-gray-900 rounded-xl w-full max-w-lg mx-4 p-6"
        on:click|stopPropagation
        role="document"
        tabindex="0"
      >
        <h2 id="block-modal-title" class="text-xl font-bold mb-4 dark:text-white">Block @{profileData.username}?</h2>
        <p class="mb-4 text-gray-600 dark:text-gray-300">
          They will not be able to follow you or view your posts, and you will not see their posts or notifications.
        </p>
        
        <div class="flex justify-end space-x-3">
          <button 
            class="px-4 py-2 border border-gray-300 dark:border-gray-700 rounded-full text-gray-700 dark:text-gray-300"
            on:click={() => showBlockConfirmModal = false}
          >
            Cancel
          </button>
          <button 
            class="px-4 py-2 bg-red-500 text-white rounded-full hover:bg-red-600"
            on:click={handleBlockUser}
          >
            {#if isBlockLoading}
              <span class="flex items-center justify-center">
                <svg class="animate-spin h-4 w-4 mr-1" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
              </span>
            {:else}
              Block
            {/if}
          </button>
        </div>
      </div>
    </div>
  {/if}
</MainLayout>

<style>
  .animate-spin {
    animation: spin 1s linear infinite;
  }
  
  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
</style>
