<script lang="ts">
  import MainLayout from '../components/layout/MainLayout.svelte';
  import TweetCard from '../components/social/TweetCard.svelte';
  import { onMount } from 'svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import type { ITweet, ITrend, ISuggestedFollow } from '../interfaces/ISocialMedia';
  import type { IAuthStore } from '../interfaces/IAuth';

  const { getAuthState } = useAuth();
  const { theme } = useTheme();

  $: authState = getAuthState ? (getAuthState() as IAuthStore) : { userId: null, isAuthenticated: false, accessToken: null, refreshToken: null };
  $: isDarkMode = $theme === 'dark';
  $: sidebarUsername = authState?.userId ? `User_${authState.userId.substring(0, 4)}` : '';
  $: sidebarDisplayName = authState?.userId ? `User ${authState.userId.substring(0, 4)}` : '';
  $: sidebarAvatar = '👤'; 

  let tweets: ITweet[] = [];
  let showComposeModal: boolean = false;
  
  let trends: ITrend[] = [
    { category: 'Politics', title: 'Election 2023', postCount: '235K' },
    { category: 'Sports', title: 'Champions League', postCount: '187K' },
    { category: 'Technology', title: 'AI Advancements', postCount: '142K' },
    { category: 'Entertainment', title: 'New Movie Release', postCount: '98K' },
  ];
  let suggestedUsers: ISuggestedFollow[] = [
    { username: 'techguru', displayName: 'Tech Guru', avatar: 'https://i.pravatar.cc/150?img=1', verified: true, followerCount: 12000 },
    { username: 'newsupdate', displayName: 'News Update', avatar: 'https://i.pravatar.cc/150?img=2', verified: true, followerCount: 9800 },
    { username: 'sportscentral', displayName: 'Sports Central', avatar: 'https://i.pravatar.cc/150?img=3', verified: false, followerCount: 500 }
  ];

  onMount(async () => {
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
    console.log('Toggle compose modal called', showComposeModal);
    showComposeModal = !showComposeModal;
    console.log('New state:', showComposeModal);
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
  <div class="min-h-screen border-x {isDarkMode ? 'border-gray-800 bg-black' : 'border-gray-200 bg-white'}">
    <div class="sticky top-0 z-10 {isDarkMode ? 'bg-black/80 border-gray-800' : 'bg-white/80 border-gray-200'} backdrop-blur-md border-b px-4 py-3">
      <h1 class="text-xl font-bold {isDarkMode ? 'text-white' : 'text-black'}">Home</h1>
    </div>
    
    <!-- Tweet Composer -->
    <div class="p-4 border-b {isDarkMode ? 'border-gray-800' : 'border-gray-200'}">
      <div class="flex items-start space-x-3">
        <div class="w-10 h-10 rounded-full {isDarkMode ? 'bg-gray-700' : 'bg-gray-200'} flex items-center justify-center overflow-hidden">
          <span>{sidebarAvatar}</span>
        </div>
        <div class="flex-1">
          <button 
            class="w-full text-left px-4 py-2 {isDarkMode ? 'text-gray-400 bg-transparent hover:bg-gray-800 border-gray-700' : 'text-gray-500 bg-transparent hover:bg-gray-100 border-gray-300'} rounded-full border"
            on:click={toggleComposeModal}
          >
            What's happening?
          </button>
        </div>
      </div>
      <div class="flex justify-between mt-3 pl-12">
        <div class="flex space-x-4">
          <button 
            class="text-blue-500 hover:text-blue-600 transition-colors" 
            title="Media"
            aria-label="Add media"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          </button>
          <button 
            class="text-blue-500 hover:text-blue-600 transition-colors" 
            title="GIF"
            aria-label="Add GIF"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </button>
          <button 
            class="text-blue-500 hover:text-blue-600 transition-colors" 
            title="Poll"
            aria-label="Create poll"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
            </svg>
          </button>
        </div>
      </div>
    </div>
    
    <!-- Tweet Feed -->
    <div class="{isDarkMode ? 'bg-black' : 'bg-white'}">
      {#each tweets as tweet (tweet.id)}
        <TweetCard {tweet} {isDarkMode} />
      {:else}
        <div class="p-4 text-center {isDarkMode ? 'text-gray-400' : 'text-gray-500'}">
          <p>Loading tweets...</p>
        </div>
      {/each}
    </div>
  </div>
</MainLayout>