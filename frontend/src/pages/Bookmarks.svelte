<script lang="ts">
  import { useTheme } from '../hooks/useTheme';
  import { useSocialMedia } from '../hooks/useSocialMedia';
  import SidebarNav from '../components/navigation/SidebarNav.svelte';
  import TweetCard from '../components/social/TweetCard.svelte';
  import appConfig from '../config/appConfig';
  import { useAuth } from '../hooks/useAuth';
  import { onMount } from 'svelte';
  import { useProfile } from '../hooks/useProfile';
  
  // Get theme
  const { theme } = useTheme();
  
  // Reactive declaration to update isDarkMode when theme changes
  $: isDarkMode = $theme === 'dark';
  
  // Get social media data
  const { tweets, toggleBookmark, likeTweet, repostTweet } = useSocialMedia();
  
  // Reactive declaration to filter bookmarked tweets
  $: bookmarkedTweets = $tweets.filter(tweet => tweet.bookmarked);
  
  // Get profile data
  const { profile, fetchProfile } = useProfile();
  
  // Get auth state if enabled
  interface AuthState {
    userId: string | null;
    accessToken: string | null;
    isAuthenticated: boolean;
  }
  
  let authState: AuthState | null = null;
  if (appConfig.auth.enabled) {
    const { getAuthState } = useAuth();
    authState = getAuthState();
  }

  // User info for the sidebar
  let username = appConfig.auth.mockUser.username;
  let displayName = appConfig.auth.mockUser.displayName;
  let avatar = appConfig.auth.mockUser.avatar;
  
  function handleToggleBookmark(event: CustomEvent) {
    const { tweetId } = event.detail;
    toggleBookmark(tweetId);
  }
  
  function handleLikeTweet(event: CustomEvent) {
    const { tweetId } = event.detail;
    likeTweet(tweetId);
  }
  
  function handleRepostTweet(event: CustomEvent) {
    const { tweetId } = event.detail;
    repostTweet(tweetId);
  }
  
  // For cleanup
  let unsubscribeProfile: Function | undefined;
  
  // Load user data on mount if auth is enabled
  onMount(() => {
    // Apply theme class to document when component mounts
    document.documentElement.classList.add(isDarkMode ? 'dark' : 'light');
    
    async function loadUserData() {
      if (appConfig.auth.enabled && authState && authState.userId) {
        await fetchProfile(authState.userId);
        
        // Update profile information from the store
        unsubscribeProfile = profile.subscribe(profileData => {
          if (profileData) {
            username = profileData.username;
            displayName = profileData.name;
            avatar = profileData.profile_picture?.[0] || '👤';
          }
        });
      }
    }
    
    loadUserData();
    
    // Return cleanup function
    return () => {
      if (unsubscribeProfile) {
        unsubscribeProfile();
      }
    };
  });
</script>

<div class="grid grid-cols-[70px_1fr] md:grid-cols-[275px_600px_350px] xl:grid-cols-[275px_600px_350px] min-h-screen {isDarkMode ? 'bg-gray-900 text-white' : 'bg-white text-gray-900'} transition-colors duration-300">
  <!-- Left Sidebar - Navigation -->
  <aside class="fixed top-0 h-screen z-20">
    <SidebarNav 
      {username} 
      {displayName} 
      {avatar}
      {isDarkMode}
    />
  </aside>
  
  <!-- Main Content - Bookmarks -->
  <main class="col-start-2 border-x {isDarkMode ? 'border-gray-800' : 'border-gray-200'} min-h-screen">
    <!-- Header -->
    <header class="sticky top-0 {isDarkMode ? 'bg-gray-900 bg-opacity-90' : 'bg-white bg-opacity-90'} backdrop-blur-sm z-10 py-3 px-4 border-b {isDarkMode ? 'border-gray-800' : 'border-gray-200'}">
      <div class="flex flex-col">
        <h1 class="text-xl font-bold">Bookmarks</h1>
        <p class="text-sm {isDarkMode ? 'text-gray-400' : 'text-gray-500'}">@{username}</p>
      </div>
    </header>
    
    <!-- Bookmarked Tweets -->
    <div>
      {#if bookmarkedTweets.length === 0}
        <div class="flex flex-col items-center justify-center p-8 text-center h-[400px]">
          <h2 class="text-3xl font-bold mb-2">You haven't added any Tweets to your Bookmarks yet</h2>
          <p class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} mb-4 text-lg">When you do, they'll show up here.</p>
        </div>
      {:else}
        {#each bookmarkedTweets as tweet (tweet.id)}
          <TweetCard 
            {tweet} 
            {isDarkMode} 
            on:toggleBookmark={handleToggleBookmark}
            on:likeTweet={handleLikeTweet}
            on:repostTweet={handleRepostTweet}
          />
        {/each}
      {/if}
    </div>
  </main>
  
  <!-- Right Sidebar - Empty space for consistent layout -->
  <aside class="col-start-3 hidden md:block w-full overflow-y-auto border-l {isDarkMode ? 'border-gray-800' : 'border-gray-200'} sticky top-0 h-screen transition-colors">
    <!-- Empty space to maintain grid layout -->
  </aside>
</div> 