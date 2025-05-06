<script lang="ts">
  import { onMount } from 'svelte';
  import MainLayout from '../components/layout/MainLayout.svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import { isAuthenticated, getUserId } from '../utils/auth';
  import { getProfile, updateProfile, pinThread, unpinThread, pinReply, unpinReply } from '../api/user';
  import { getUserThreads, getUserReplies, getUserLikedThreads, getUserMedia, getThreadReplies } from '../api/thread';
  import { toastStore } from '../stores/toastStore';
  import ThreadCard from '../components/explore/ThreadCard.svelte';
  import TweetCard from '../components/social/TweetCard.svelte';
  import LoadingSkeleton from '../components/common/LoadingSkeleton.svelte';
  import ProfileEditModal from '../components/profile/ProfileEditModal.svelte';
  import type { ITweet } from '../interfaces/ISocialMedia';
  
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
  
  // Additional functions for thread interactions
  let repliesMap = new Map(); // Store replies for threads
  
  // Helper function to ensure an object has all ITweet properties
  function ensureTweetFormat(thread: any): ITweet {
    return {
      id: thread.id,
      content: thread.content || '',
      username: thread.username || thread.author_username || 'user',
      displayName: thread.displayName || thread.author_name || 'User',
      timestamp: thread.timestamp || thread.created_at || new Date().toISOString(),
      likes: thread.likes || thread.likes_count || 0,
      replies: thread.replies || thread.replies_count || 0,
      reposts: thread.reposts || thread.reposts_count || 0,
      bookmarks: thread.bookmarks || thread.bookmarks_count || 0,
      views: thread.views || String(thread.views_count || '0'),
      avatar: thread.avatar || thread.author_avatar || null,
      isLiked: thread.isLiked || thread.is_liked || false,
      isReposted: thread.isReposted || thread.is_reposted || false,
      isBookmarked: thread.isBookmarked || thread.is_bookmarked || false,
      media: thread.media || []
    };
  }
  
  // Load replies for a thread
  async function loadRepliesForThread(threadId) {
    try {
      const response = await getThreadReplies(threadId);
      if (response && response.replies) {
        repliesMap.set(threadId, response.replies.map(reply => ensureTweetFormat(reply)));
        return repliesMap.get(threadId);
      }
      return [];
    } catch (error) {
      console.error('Error loading replies:', error);
      return [];
    }
  }
  
  // Event handlers for TweetCard
  function handleLoadReplies(event) {
    const threadId = event.detail;
    loadRepliesForThread(threadId);
  }
  
  function handleReply(event) {
    const threadId = event.detail;
    // Navigate to thread detail or open reply modal
    window.location.href = `/thread/${threadId}`;
  }
  
  function handleThreadClick(event) {
    const thread = event.detail;
    window.location.href = `/thread/${thread.id}`;
  }
  
  // Handle tab switching
  function setActiveTab(tab) {
    activeTab = tab;
    loadTabContent(tab);
  }
  
  // Load content for the active tab
  async function loadTabContent(tab) {
    isLoading = true;
    try {
      switch(tab) {
        case 'posts':
          console.log('Loading posts tab');
          try {
            const postsData = await getUserThreads('me');
            // Map API response to the Tweet format for TweetCard
            posts = (postsData.threads || []).map(thread => ensureTweetFormat(thread));
            console.log(`Loaded ${posts.length} posts`);
          } catch (error: any) {
            console.error('Error loading posts:', error);
            toastStore.showToast(`Failed to load posts: ${error.message}`, 'error');
            posts = [];
          }
          break;
        case 'replies':
          console.log('Loading replies tab');
          try {
            const repliesData = await getUserReplies('me');
            replies = repliesData.replies || [];
            console.log(`Loaded ${replies.length} replies`);
          } catch (error: any) {
            console.error('Error loading replies:', error);
            toastStore.showToast(`Failed to load replies: ${error.message}`, 'error');
            replies = [];
          }
          break;
        case 'likes':
          console.log('Loading likes tab');
          try {
            const likesData = await getUserLikedThreads('me');
            // Map API response to the Tweet format for TweetCard
            likes = (likesData.threads || []).map(thread => ensureTweetFormat(thread));
            console.log(`Loaded ${likes.length} liked threads`);
          } catch (error: any) {
            console.error('Error loading likes:', error);
            toastStore.showToast(`Failed to load likes: ${error.message}`, 'error');
            likes = [];
          }
          break;
        case 'media':
          console.log('Loading media tab');
          try {
            const mediaData = await getUserMedia('me');
            media = mediaData.media || [];
            console.log(`Loaded ${media.length} media items`);
          } catch (error: any) {
            console.error('Error loading media:', error);
            toastStore.showToast(`Failed to load media: ${error.message}`, 'error');
            media = [];
          }
          break;
      }
    } catch (error: any) {
      console.error(`General error loading ${tab}:`, error);
      toastStore.showToast(`Failed to load ${tab}. Please try again.`, 'error');
    } finally {
      isLoading = false;
    }
  }
  
  // Load profile data
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
      // Load posts by default
      loadTabContent('posts');
    }
  }
  
  // Handle profile update
  async function handleProfileUpdate(event) {
    const updatedData = event.detail;
    isUpdatingProfile = true;
    
    try {
      const response = await updateProfile(updatedData);
      if (response && response.success) {
        toastStore.showToast('Profile updated successfully!', 'success');
        loadProfileData(); // Reload profile data
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
  
  // Handle pin/unpin thread
  async function handlePinThread(threadId, isPinned) {
    try {
      if (isPinned) {
        await unpinThread(threadId);
      } else {
        await pinThread(threadId);
      }
      // Reload posts to reflect changes
      loadTabContent('posts');
      toastStore.showToast(isPinned ? 'Thread unpinned' : 'Thread pinned', 'success');
    } catch (error) {
      console.error('Error pinning/unpinning thread:', error);
      toastStore.showToast('Failed to pin/unpin thread', 'error');
    }
  }
  
  // Handle pin/unpin reply
  async function handlePinReply(replyId, isPinned) {
    try {
      if (isPinned) {
        await unpinReply(replyId);
      } else {
        await pinReply(replyId);
      }
      // Reload replies to reflect changes
      loadTabContent('replies');
      toastStore.showToast(isPinned ? 'Reply unpinned' : 'Reply pinned', 'success');
    } catch (error) {
      console.error('Error pinning/unpinning reply:', error);
      toastStore.showToast('Failed to pin/unpin reply', 'error');
    }
  }
  
  // Format date to display "Joined [Month] [Year]"
  function formatJoinDate(dateString) {
    if (!dateString) return '';
    
    const date = new Date(dateString);
    return `Joined ${date.toLocaleString('default', { month: 'long' })} ${date.getFullYear()}`;
  }
  
  // Initialize on component mount
  onMount(() => {
    // Check if user is authenticated
    if (!isAuthenticated()) {
      console.log('User not authenticated, redirecting to login');
      window.location.href = '/login';
      return;
    }
    
    // Get user ID
    const userId = getUserId();
    if (!userId) {
      console.error('No user ID available despite being authenticated');
      toastStore.showToast('Authentication error: User ID not found', 'error');
      // You may want to clear auth data and redirect to login
      window.location.href = '/login';
      return;
    }
    
    console.log(`User authenticated with ID: ${userId}`);
    loadProfileData();
  });
</script>

<MainLayout
  username={profileData.username}
  displayName={profileData.displayName}
  avatar={profileData.profilePicture}
>
  <!-- Profile container -->
  <div class="w-full min-h-screen border-x border-gray-200 dark:border-gray-800">
    {#if isLoading && !profileData.id}
      <LoadingSkeleton type="profile" />
    {:else}
      <!-- Banner image -->
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
      
      <!-- Profile header -->
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
            class="px-4 py-2 rounded-full border border-gray-300 dark:border-gray-700 font-bold hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
            on:click={() => showEditModal = true}
          >
            Edit profile
          </button>
        </div>
      </div>
      
      <!-- Profile info -->
      <div class="p-4 mt-2">
        <h1 class="text-xl font-bold dark:text-white">{profileData.displayName}</h1>
        <p class="text-gray-500 dark:text-gray-400">@{profileData.username}</p>
        
        {#if profileData.bio}
          <p class="my-3 dark:text-white whitespace-pre-wrap">{profileData.bio}</p>
        {/if}
        
        <div class="flex items-center mt-3 text-gray-500 dark:text-gray-400 text-sm">
          <span class="flex items-center">
            <svg class="w-4 h-4 mr-1" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M8 2V5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
              <path d="M16 2V5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
              <path d="M3 8H21" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
              <path d="M3 7.5C3 5.29086 4.79086 3.5 7 3.5H17C19.2091 3.5 21 5.29086 21 7.5V18C21 20.2091 19.2091 22 17 22H7C4.79086 22 3 20.2091 3 18V7.5Z" stroke="currentColor" stroke-width="1.5"/>
            </svg>
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
      
      <!-- Profile tabs -->
      <div class="flex border-b border-gray-200 dark:border-gray-800">
        <button 
          class="flex-1 py-4 text-gray-500 dark:text-gray-400 font-medium hover:bg-gray-50 dark:hover:bg-gray-900 relative {activeTab === 'posts' ? 'text-blue-500 font-bold' : ''}"
          on:click={() => setActiveTab('posts')}
        >
          Posts
          {#if activeTab === 'posts'}
            <div class="absolute bottom-0 left-0 w-full h-1 bg-blue-500 rounded-t"></div>
          {/if}
        </button>
        <button 
          class="flex-1 py-4 text-gray-500 dark:text-gray-400 font-medium hover:bg-gray-50 dark:hover:bg-gray-900 relative {activeTab === 'replies' ? 'text-blue-500 font-bold' : ''}"
          on:click={() => setActiveTab('replies')}
        >
          Replies
          {#if activeTab === 'replies'}
            <div class="absolute bottom-0 left-0 w-full h-1 bg-blue-500 rounded-t"></div>
          {/if}
        </button>
        <button 
          class="flex-1 py-4 text-gray-500 dark:text-gray-400 font-medium hover:bg-gray-50 dark:hover:bg-gray-900 relative {activeTab === 'likes' ? 'text-blue-500 font-bold' : ''}"
          on:click={() => setActiveTab('likes')}
        >
          Likes
          {#if activeTab === 'likes'}
            <div class="absolute bottom-0 left-0 w-full h-1 bg-blue-500 rounded-t"></div>
          {/if}
        </button>
        <button 
          class="flex-1 py-4 text-gray-500 dark:text-gray-400 font-medium hover:bg-gray-50 dark:hover:bg-gray-900 relative {activeTab === 'media' ? 'text-blue-500 font-bold' : ''}"
          on:click={() => setActiveTab('media')}
        >
          Media
          {#if activeTab === 'media'}
            <div class="absolute bottom-0 left-0 w-full h-1 bg-blue-500 rounded-t"></div>
          {/if}
        </button>
      </div>
      
      <!-- Tab content -->
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
              <div class="mb-4 p-4 rounded-lg border border-gray-200 dark:border-gray-800 {post.is_pinned ? 'bg-gray-50 dark:bg-gray-900' : ''}">
                {#if post.is_pinned}
                  <div class="flex items-center text-blue-500 text-xs font-bold mb-2">
                    <svg class="w-3.5 h-3.5 mr-1" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <path d="M12 21V12M12 12L18 8M12 12L6 8M6 4H18" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                    </svg>
                    Pinned
                  </div>
                {/if}
                <TweetCard 
                  tweet={ensureTweetFormat(post)} 
                  {isDarkMode} 
                  isAuthenticated={!!authState.isAuthenticated}
                  isLiked={post.isLiked || post.is_liked || false}
                  isBookmarked={post.isBookmarked || post.is_bookmarked || false}
                  isReposted={post.isReposted || post.is_reposted || false}
                  replies={repliesMap.get(post.id) || []}
                  showReplies={false}
                  nestingLevel={0}
                  nestedRepliesMap={new Map()}
                  on:reply={handleReply}
                  on:loadReplies={handleLoadReplies}
                  on:click={handleThreadClick}
                />
                <button 
                  class="mt-2 text-xs text-gray-500 dark:text-gray-400 hover:text-blue-500 dark:hover:text-blue-400 hover:underline"
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
                    <svg class="w-3.5 h-3.5 mr-1" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <path d="M12 21V12M12 12L18 8M12 12L6 8M6 4H18" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                    </svg>
                    Pinned
                  </div>
                {/if}
                <div class="p-3 bg-gray-50 dark:bg-gray-900 rounded-lg">
                  <p class="text-sm text-gray-500 dark:text-gray-400 mb-2">
                    Replying to <a href={`/thread/${reply.thread_id}`} class="text-blue-500 hover:underline">@{reply.thread_author}</a>
                  </p>
                  <p class="text-black dark:text-white">{reply.content}</p>
                  <div class="mt-2 text-xs text-gray-500 dark:text-gray-400">
                    <span>{new Date(reply.created_at).toLocaleDateString()}</span>
                  </div>
                </div>
                <button 
                  class="mt-2 text-xs text-gray-500 dark:text-gray-400 hover:text-blue-500 dark:hover:text-blue-400 hover:underline"
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
                  isLiked={like.isLiked || like.is_liked || true}
                  isBookmarked={like.isBookmarked || like.is_bookmarked || false}
                  isReposted={like.isReposted || like.is_reposted || false}
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
          <svg class="w-6 h-6" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M18 6L6 18" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M6 6L18 18" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
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
