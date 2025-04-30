<script lang="ts">
  import MainLayout from '../components/layout/MainLayout.svelte';
  import ComposeTweet from '../components/social/ComposeTweet.svelte';
  import CreateThreadModal from '../components/feed/CreateThreadModal.svelte';
  import { onMount } from 'svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import type { ITweet, ITrend, ISuggestedFollow } from '../interfaces/ISocialMedia';
  import type { IAuthStore } from '../interfaces/IAuth';

  // Accept the route prop for conditionally rendering content
  export let route: string;

  // Get auth store methods
  const { getAuthState } = useAuth();
  // Get theme store
  const { theme } = useTheme();

  // Reactive declarations for auth and theme
  $: authState = getAuthState ? (getAuthState() as IAuthStore) : { userId: null, isAuthenticated: false, accessToken: null, refreshToken: null };
  $: isDarkMode = $theme === 'dark';
  // User information for sidebar
  $: sidebarUsername = authState?.userId ? `User_${authState.userId.substring(0, 4)}` : '';
  $: sidebarDisplayName = authState?.userId ? `User ${authState.userId.substring(0, 4)}` : '';
  $: sidebarAvatar = '👤'; // Placeholder avatar

  // State for tweets and compose modal
  let tweets: ITweet[] = [];
  let showComposeModal: boolean = false;
  let showThreadModal: boolean = false;
  let selectedTweet: ITweet | null = null;
  
  // Trends data
  let trends: ITrend[] = [
    { category: 'Politics', title: 'Election 2023', postCount: '235K' },
    { category: 'Sports', title: 'Champions League', postCount: '187K' },
    { category: 'Technology', title: 'AI Advancements', postCount: '142K' },
    { category: 'Entertainment', title: 'New Movie Release', postCount: '98K' },
    { category: 'Health', title: 'Fitness Trends', postCount: '75K' }
  ];
  
  // Suggested users to follow
  let suggestedUsers: ISuggestedFollow[] = [
    { username: 'techguru', displayName: 'Tech Guru', avatar: 'https://i.pravatar.cc/150?img=1', verified: true, followerCount: 12000 },
    { username: 'newsupdate', displayName: 'News Update', avatar: 'https://i.pravatar.cc/150?img=2', verified: true, followerCount: 9800 },
    { username: 'sportscentral', displayName: 'Sports Central', avatar: 'https://i.pravatar.cc/150?img=3', verified: false, followerCount: 500 }
  ];

  onMount(async () => {
    // Fetch tweets (simulated)
    setTimeout(() => {
      tweets = [
        { 
          id: 1, 
          username: 'devuser', 
          displayName: 'Developer', 
          avatar: '🧑‍💻',
          content: 'Just released a new feature!', 
          timestamp: new Date().toISOString(), 
          likes: 42, 
          replies: 5,
          reposts: 11,
          views: '1.1K'
        },
        { 
          id: 2, 
          username: 'naturelover', 
          displayName: 'Nature Enthusiast', 
          avatar: '🌳',
          content: 'What a beautiful day! #sunshine', 
          timestamp: new Date(Date.now() - 3600000).toISOString(), 
          likes: 18, 
          replies: 2,
          reposts: 3,
          views: '580'
        }
      ];
    }, 1000);
  });

  function toggleComposeModal() {
    showComposeModal = !showComposeModal;
  }
  
  function openThreadModal(tweet: ITweet) {
    selectedTweet = tweet;
    showThreadModal = true;
  }
  
  function closeThreadModal() {
    showThreadModal = false;
    selectedTweet = null;
  }
  
  // Format the timestamp to a relative time string
  function formatTimestamp(timestamp: string): string {
    const date = new Date(timestamp);
    const now = new Date();
    const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);
    
    let interval = seconds / 31536000; // seconds in a year
    if (interval > 1) {
      return Math.floor(interval) + 'y';
    }
    interval = seconds / 2592000; // seconds in a month
    if (interval > 1) {
      return Math.floor(interval) + 'mo';
    }
    interval = seconds / 86400; // seconds in a day
    if (interval > 1) {
      return Math.floor(interval) + 'd';
    }
    interval = seconds / 3600; // seconds in an hour
    if (interval > 1) {
      return Math.floor(interval) + 'h';
    }
    interval = seconds / 60; // seconds in a minute
    if (interval > 1) {
      return Math.floor(interval) + 'm';
    }
    return Math.floor(seconds) + 's';
  }
</script>

<MainLayout
  username={sidebarUsername}
  displayName={sidebarDisplayName}
  avatar={sidebarAvatar}
  trends={trends}
  suggestedFollows={suggestedUsers}
  on:toggleComposeModal={toggleComposeModal}
>
  <!-- Dynamic Content Area -->
  <div class="min-h-screen border-x border-gray-200 dark:border-gray-800">
    <!-- Dynamic Header based on route -->
    <div class="sticky top-0 z-10 bg-white/80 dark:bg-black/80 backdrop-blur-md border-b border-gray-200 dark:border-gray-800 px-4 py-3">
      {#if route === '/home' || route === '/feed'}
        <h1 class="text-xl font-bold">Home</h1>
      {:else if route === '/explore'}
        <h1 class="text-xl font-bold">Explore</h1>
      {:else if route === '/notifications'}
        <h1 class="text-xl font-bold">Notifications</h1>
      {:else if route === '/messages'}
        <h1 class="text-xl font-bold">Messages</h1>
      {:else if route === '/profile'}
        <h1 class="text-xl font-bold">Profile</h1>
      {/if}
    </div>
    
    <!-- Dynamic Content based on route -->
    {#if route === '/home' || route === '/feed'}
      <!-- Home Feed Content -->
      {#each tweets as tweet (tweet.id)}
        <!-- Replace TweetCard with custom post component that opens modal on click -->
        <div 
          class="post {isDarkMode ? 'post-dark' : ''} border-b {isDarkMode ? 'border-gray-800' : 'border-gray-200'} hover:bg-opacity-50 {isDarkMode ? 'hover:bg-gray-900 bg-black text-white' : 'hover:bg-gray-50 bg-white text-black'} transition-colors cursor-pointer"
          on:click={() => openThreadModal(tweet)}
        >
          <div class="post-header p-4">
            <div class="flex items-start">
              <div class="post-avatar-container w-12 h-12 rounded-full overflow-hidden {isDarkMode ? 'bg-gray-700' : 'bg-gray-200'} flex items-center justify-center mr-3">
                {#if typeof tweet.avatar === 'string' && tweet.avatar.startsWith('http')}
                  <img src={tweet.avatar} alt={tweet.username} class="w-full h-full object-cover" />
                {:else}
                  <div class="text-xl">{tweet.avatar}</div>
                {/if}
              </div>
              
              <div class="flex-1 min-w-0">
                <div class="flex items-center">
                  <span class="font-bold {isDarkMode ? 'text-white' : 'text-black'} mr-1.5">{tweet.displayName}</span>
                  <span class="text-gray-500 text-sm truncate">@{tweet.username}</span>
                  <span class="text-gray-500 mx-1.5">·</span>
                  <span class="text-gray-500 text-sm">{formatTimestamp(tweet.timestamp)}</span>
                </div>
                
                <div class="post-content my-2 {isDarkMode ? 'text-white' : 'text-black'}">
                  <p>{tweet.content}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      {/each}
    {:else if route === '/explore'}
      <!-- Explore Content -->
      <div class="p-4">
        <div class="bg-gray-100 dark:bg-gray-800 rounded-lg p-4">
          <h2 class="text-lg font-semibold mb-2">Trending Topics</h2>
          <p>Explore page content coming soon...</p>
        </div>
      </div>
    {:else if route === '/notifications'}
      <!-- Notifications Content -->
      <div class="p-4">
        <div class="bg-gray-100 dark:bg-gray-800 rounded-lg p-4">
          <h2 class="text-lg font-semibold mb-2">All Notifications</h2>
          <p>No notifications to display.</p>
        </div>
      </div>
    {:else if route === '/messages'}
      <!-- Messages Content -->
      <div class="p-4">
        <div class="bg-gray-100 dark:bg-gray-800 rounded-lg p-4">
          <h2 class="text-lg font-semibold mb-2">Messages</h2>
          <p>Your message inbox is empty.</p>
        </div>
      </div>
    {:else if route === '/profile'}
      <!-- Profile Content -->
      <div class="p-4">
        <div class="bg-gray-100 dark:bg-gray-800 rounded-lg p-4">
          <h2 class="text-lg font-semibold mb-2">Your Profile</h2>
          <p>Profile information will appear here.</p>
        </div>
      </div>
    {/if}
  </div>
</MainLayout>

{#if showComposeModal}
  <ComposeTweet 
    {isDarkMode}
    on:close={toggleComposeModal}
    avatar={sidebarAvatar} 
  />
{/if}

{#if showThreadModal && selectedTweet}
  <CreateThreadModal 
    username={selectedTweet.username}
    displayName={selectedTweet.displayName}
    avatar={selectedTweet.avatar}
    isAdmin={false}
    on:close={closeThreadModal}
  />
{/if}

<style>
  /* Additional custom styles for posts */
  .post {
    padding: 0.5rem 0;
  }

  .post-avatar-container {
    flex-shrink: 0;
  }
</style>