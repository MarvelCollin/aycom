<script lang="ts">
  import { onMount } from 'svelte';
  import { useTheme } from '../hooks/useTheme';
  import { useSocialMedia } from '../hooks/useSocialMedia';
  import SidebarNav from '../components/navigation/SidebarNav.svelte';
  import ComposeTweet from '../components/social/ComposeTweet.svelte';
  import TweetCard from '../components/social/TweetCard.svelte';
  import TrendsList from '../components/social/TrendsList.svelte';
  import SuggestedFollows from '../components/social/SuggestedFollows.svelte';
  import SearchBar from '../components/social/SearchBar.svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useProfile } from '../hooks/useProfile';
  import appConfig from '../config/appConfig';
  
  // Import Feather icons
  import { 
    XIcon
  } from 'svelte-feather-icons';

  // Get theme
  const { theme } = useTheme();
  
  // Reactive declaration to update isDarkMode when theme changes
  $: isDarkMode = $theme === 'dark';
  
  // Get social media data
  const { tweets, trends, suggestedUsers, postTweet, toggleBookmark, likeTweet, repostTweet } = useSocialMedia();
  
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
  
  // Set up compose tweet modal
  let showComposeModal = false;
  
  function toggleComposeModal() {
    showComposeModal = !showComposeModal;
  }
  
  // Handle the toggleComposeModal event from SidebarNav
  function handleToggleComposeModal() {
    toggleComposeModal();
  }
  
  interface TweetEvent {
    detail: {
      content: string;
    }
  }
  
  function handleTweet(event: TweetEvent) {
    const { content } = event.detail;
    postTweet(content);
    showComposeModal = false;
  }
  
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
      on:toggleComposeModal={handleToggleComposeModal}
    />
  </aside>
  
  <!-- Main Content - Feed -->
  <main class="col-start-2 border-x {isDarkMode ? 'border-gray-800' : 'border-gray-200'} min-h-screen">
    <!-- Header -->
    <header class="sticky top-0 {isDarkMode ? 'bg-gray-900 bg-opacity-90' : 'bg-white bg-opacity-90'} backdrop-blur-sm z-10 py-3 px-4 border-b {isDarkMode ? 'border-gray-800' : 'border-gray-200'}">
      <div class="flex justify-between items-center">
        <h1 class="text-xl font-bold">Home</h1>
        <div class="flex w-full justify-center">
          <div class="flex border-b-2 {isDarkMode ? 'border-blue-400' : 'border-blue-500'} w-1/2 justify-center">
            <button class="py-3 px-4 font-semibold">
              For you
            </button>
          </div>
          <div class="flex w-1/2 justify-center">
            <button class="py-3 px-4 text-gray-500 hover:bg-gray-200 hover:bg-opacity-20 rounded-full">
              Following
            </button>
          </div>
        </div>
      </div>
    </header>
    
    <!-- Compose Tweet Input -->
    <div class="p-4 border-b {isDarkMode ? 'border-gray-800' : 'border-gray-200'} transition-colors">
      <div class="flex">
        <div class="w-12 h-12 rounded-full flex items-center justify-center mr-4 overflow-hidden">
          <div class="w-full h-full {isDarkMode ? 'bg-gray-700' : 'bg-gray-300'} rounded-full flex items-center justify-center overflow-hidden">
            <span>{avatar}</span>
          </div>
        </div>
        <div class="flex-1">
          <input 
            type="text" 
            placeholder="What's happening?" 
            class="w-full {isDarkMode ? 'bg-transparent placeholder-gray-500' : 'bg-transparent placeholder-gray-400'} text-xl p-2 border-none outline-none transition-colors"
            on:click={toggleComposeModal}
            readonly
          />
          <div class="mt-2 border-t {isDarkMode ? 'border-gray-800' : 'border-gray-200'} pt-2 flex justify-between items-center">
            <div class="flex space-x-1 text-blue-500">
              <!-- Tweet tools like image upload would go here -->
            </div>
            <button 
              class="bg-blue-500 text-white px-4 py-1.5 rounded-full font-bold disabled:opacity-50 hover:bg-blue-600 transition-colors text-sm"
              disabled={true}
            >
              Post
            </button>
          </div>
        </div>
      </div>
    </div>
    
    <!-- Tweet Feed -->
    <div>
      {#each $tweets as tweet (tweet.id)}
        <TweetCard 
          {tweet} 
          {isDarkMode} 
          on:toggleBookmark={handleToggleBookmark}
          on:likeTweet={handleLikeTweet}
          on:repostTweet={handleRepostTweet}
        />
      {/each}
    </div>
  </main>
  
  <!-- Right Sidebar - Search and Recommendations (Hidden on mobile) -->
  <aside class="col-start-3 hidden md:block w-full overflow-y-auto border-l {isDarkMode ? 'border-gray-800' : 'border-gray-200'} sticky top-0 h-screen px-6 py-3 transition-colors">
    <!-- Search -->
    <div class="mb-4">
      <SearchBar {isDarkMode} />
    </div>
    
    <!-- Premium Subscription -->
    <div class="{isDarkMode ? 'bg-gray-800' : 'bg-gray-100'} rounded-xl mb-6 p-4 transition-colors">
      <h2 class="text-xl font-bold mb-2">Subscribe to Premium</h2>
      <p class="text-sm mb-4 {isDarkMode ? 'text-gray-300' : 'text-gray-600'}">Subscribe to unlock new features and if eligible, receive a share of revenue.</p>
      <button class="bg-blue-500 text-white px-4 py-2 rounded-full font-bold w-full hover:bg-blue-600 transition-colors">
        Subscribe
      </button>
    </div>
    
    <!-- Trends -->
    <TrendsList {isDarkMode} trends={$trends} />
    
    <!-- Who to follow -->
    <SuggestedFollows {isDarkMode} suggestedUsers={$suggestedUsers} />
  </aside>
</div>

<!-- Compose Tweet Modal -->
{#if showComposeModal}
  <div class="fixed inset-0 {isDarkMode ? 'bg-black bg-opacity-70' : 'bg-gray-500 bg-opacity-70'} flex items-center justify-center z-50 transition-colors">
    <div class="{isDarkMode ? 'bg-gray-900' : 'bg-white'} border {isDarkMode ? 'border-gray-800' : 'border-gray-300'} rounded-xl w-full max-w-xl shadow-xl transition-colors">
      <div class="flex justify-between items-center p-4 border-b {isDarkMode ? 'border-gray-800' : 'border-gray-200'} transition-colors">
        <button 
          class="text-xl {isDarkMode ? 'hover:bg-gray-800' : 'hover:bg-gray-100'} p-2 rounded-full transition-colors" 
          on:click={toggleComposeModal}
        >
          <XIcon size="20" />
        </button>
        <span></span>
      </div>
      
      <ComposeTweet on:tweet={handleTweet} {isDarkMode} {avatar} />
    </div>
  </div>
{/if} 