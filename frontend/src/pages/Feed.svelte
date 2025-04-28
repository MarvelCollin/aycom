<script lang="ts">
  import { onMount } from 'svelte';
  import SidebarNav from '../components/navigation/SidebarNav.svelte';
  import ComposeTweet from '../components/social/ComposeTweet.svelte';
  import TweetCard from '../components/social/TweetCard.svelte';
  import TrendsList from '../components/social/TrendsList.svelte';
  import SuggestedFollows from '../components/social/SuggestedFollows.svelte';
  import SearchBar from '../components/social/SearchBar.svelte';
  import type { ITweet } from '../interfaces/ISocialMedia';
  import { useAuth } from '../hooks/useAuth';
  import { useProfile } from '../hooks/useProfile';
  import appConfig from '../config/appConfig';

  // Get profile data
  const { profile, fetchProfile } = useProfile();
  
  // Get auth state if enabled
  interface AuthState {
    userId: string | null;
    accessToken: string | null;
    isAuthenticated: boolean;
    // Add other properties as needed
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
  
  // Dummy data for tweets
  const tweets: ITweet[] = [
    {
      id: 1,
      username: 'elonmusk',
      displayName: 'Elon Musk',
      avatar: '👨‍🚀',
      content: 'Just launched another rocket! 🚀 #SpaceX',
      timestamp: '2h',
      likes: 3240,
      replies: 421,
      reposts: 892,
      views: '1.2M'
    },
    {
      id: 2,
      username: 'AYCOM',
      displayName: 'AYCOM Official',
      avatar: 'AY',
      content: 'Welcome to our new social media platform! #AYCOM #Launch',
      timestamp: '5h',
      likes: 1548,
      replies: 246,
      reposts: 567,
      views: '820K'
    },
    {
      id: 3,
      username: 'tech_news',
      displayName: 'Tech News',
      avatar: '📱',
      content: 'Breaking: New advancements in AI technology have researchers excited about future applications. #AI #Technology',
      timestamp: '12h',
      likes: 982,
      replies: 124,
      reposts: 325,
      views: '456K'
    }
  ];
  
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
    console.log('New tweet from Feed page:', content);
    // Here you would typically send the tweet to an API
    showComposeModal = false;
  }
  
  // For cleanup
  let unsubscribeProfile: Function | undefined;
  
  // Load user data on mount if auth is enabled
  onMount(() => {
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

<div class="grid grid-cols-[auto_1fr_auto] min-h-screen bg-black text-white">
  <!-- Left Sidebar - Navigation -->
  <SidebarNav 
    {username} 
    {displayName} 
    {avatar} 
    on:toggleComposeModal={handleToggleComposeModal}
  />
  
  <!-- Main Content - Feed -->
  <main class="border-x border-gray-800 max-w-xl">
    <!-- Header -->
    <header class="sticky top-0 bg-black bg-opacity-80 backdrop-blur-sm z-10 p-4 border-b border-gray-800">
      <div class="flex justify-between items-center">
        <h1 class="text-xl font-bold">Home</h1>
        <div class="flex space-x-4">
          <button class="p-2 hover:bg-gray-800 rounded-full">
            <span>For you</span>
          </button>
          <button class="p-2 hover:bg-gray-800 rounded-full">
            <span>Following</span>
          </button>
        </div>
      </div>
    </header>
    
    <!-- Compose Tweet Input -->
    <div class="p-4 border-b border-gray-800">
      <div class="flex">
        <div class="w-12 h-12 bg-gray-300 rounded-full flex items-center justify-center mr-4">
          <span>{avatar}</span>
        </div>
        <div class="flex-1">
          <input 
            type="text" 
            placeholder="What's happening?" 
            class="w-full bg-transparent text-xl p-2 border-none outline-none"
            on:click={toggleComposeModal}
            readonly
          />
          <div class="flex mt-2">
            <button class="text-blue-500 mr-3">🖼️</button>
            <button class="text-blue-500 mr-3">📊</button>
            <button class="text-blue-500 mr-3">😊</button>
            <button class="text-blue-500 mr-3">📍</button>
          </div>
        </div>
      </div>
      <button 
        class="mt-2 bg-blue-500 text-white px-4 py-2 rounded-full font-bold float-right disabled:opacity-50"
        disabled={true}
      >
        Post
      </button>
      <div class="clear-both"></div>
    </div>
    
    <!-- Tweet Feed -->
    <div>
      {#each tweets as tweet (tweet.id)}
        <TweetCard {tweet} />
      {/each}
    </div>
  </main>
  
  <!-- Right Sidebar - Search and Recommendations -->
  <div class="w-80 p-4 overflow-y-auto border-l border-gray-800 sticky top-0 h-screen">
    <!-- Search -->
    <SearchBar />
    
    <!-- Premium Subscription -->
    <div class="bg-gray-900 rounded-xl mb-6 p-4">
      <h2 class="text-xl font-bold mb-2">Subscribe to Premium</h2>
      <p class="text-sm mb-4">Subscribe to unlock new features and if eligible, receive a share of revenue.</p>
      <button class="bg-blue-500 text-white px-4 py-2 rounded-full font-bold w-full">
        Subscribe
      </button>
    </div>
    
    <!-- Trends -->
    <TrendsList />
    
    <!-- Who to follow -->
    <SuggestedFollows />
  </div>
</div>

<!-- Compose Tweet Modal -->
{#if showComposeModal}
  <div class="fixed inset-0 bg-black bg-opacity-70 flex items-center justify-center z-50">
    <div class="bg-black border border-gray-800 rounded-xl w-full max-w-xl">
      <div class="flex justify-between items-center p-4 border-b border-gray-800">
        <button 
          class="text-xl hover:bg-gray-800 p-2 rounded-full" 
          on:click={toggleComposeModal}
        >
          ✕
        </button>
        <span></span>
      </div>
      
      <ComposeTweet on:tweet={handleTweet} />
    </div>
  </div>
{/if} 