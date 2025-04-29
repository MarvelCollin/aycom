<script lang="ts">
  import MainColumn from '../components/common/MainColumn.svelte';
  import LeftSide from '../components/common/LeftSide.svelte';
  import RightSide from '../components/common/RightSide.svelte';
  import TweetCard from '../components/social/TweetCard.svelte';
  import ComposeTweet from '../components/social/ComposeTweet.svelte';
  import { onMount } from 'svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import type { ITweet, ITrend, ISuggestedFollow } from '../interfaces/ISocialMedia';
  import { get } from 'svelte/store'; // Import get for auth state

  // Get auth store methods
  const { subscribe, getAuthState } = useAuth();
  // Get theme store
  const { theme } = useTheme();

  // Reactive declarations for auth and theme
  $: authState = getAuthState ? getAuthState() : { userId: null, isAuthenticated: false, accessToken: null, refreshToken: null };
  $: isDarkMode = $theme === 'dark';
  // Placeholder user details - A real app would fetch these based on authState.userId
  $: sidebarUsername = authState?.userId ? `User_${authState.userId.substring(0, 4)}` : '';
  $: sidebarDisplayName = authState?.userId ? `User ${authState.userId.substring(0, 4)}` : '';
  $: sidebarAvatar = '👤'; // Placeholder avatar

  let tweets: ITweet[] = [];
  let showComposeModal: boolean = false;
  
  // Trends data - Updated to match ITrend interface
  let trends: ITrend[] = [
    { category: 'Politics', title: 'Election 2023', postCount: '235K' }, // Changed topic->category, tweets->postCount (made string)
    { category: 'Sports', title: 'Champions League', postCount: '187K' },
    { category: 'Technology', title: 'AI Advancements', postCount: '142K' },
    { category: 'Entertainment', title: 'New Movie Release', postCount: '98K' },
    { category: 'Health', title: 'Fitness Trends', postCount: '75K' }
  ];
  
  // Suggested users to follow - Updated to match ISuggestedFollow interface
  let suggestedUsers: ISuggestedFollow[] = [
    // Removed id, added followerCount
    { username: 'techguru', displayName: 'Tech Guru', avatar: 'https://i.pravatar.cc/150?img=1', verified: true, followerCount: 12000 }, 
    { username: 'newsupdate', displayName: 'News Update', avatar: 'https://i.pravatar.cc/150?img=2', verified: true, followerCount: 9800 },
    { username: 'sportscentral', displayName: 'Sports Central', avatar: 'https://i.pravatar.cc/150?img=3', verified: false, followerCount: 500 }
  ];

  onMount(async () => {
    // Fetch tweets (simulated) - Updated mock data to match ITweet interface
    setTimeout(() => {
      tweets = [
        { 
          id: 1, 
          username: 'devuser', 
          displayName: 'Developer', 
          avatar: '🧑‍💻', // Changed avatar source
          content: 'Just released a new feature!', 
          timestamp: new Date().toISOString(), 
          likes: 42, 
          replies: 5, // Changed comments->replies
          reposts: 11, // Changed retweets->reposts
          views: '1.1K' // Added views
        },
        { 
          id: 2, 
          username: 'naturelover', 
          displayName: 'Nature Enthusiast', 
          avatar: '🌳', // Changed avatar source
          content: 'What a beautiful day! #sunshine', 
          timestamp: new Date(Date.now() - 3600000).toISOString(), 
          likes: 18, 
          replies: 2, // Changed comments->replies
          reposts: 3, // Changed retweets->reposts
          views: '580' // Added views
        }
      ];
    }, 1000);
  });

  function toggleComposeModal() {
    showComposeModal = !showComposeModal;
  }
</script>

<div class="flex w-full h-screen {isDarkMode ? 'bg-black text-white' : 'bg-white text-gray-900'}">
  <!-- Left Sidebar -->
  <LeftSide 
    username={sidebarUsername}
    displayName={sidebarDisplayName}
    avatar={sidebarAvatar}
    {isDarkMode}
    on:toggleComposeModal={toggleComposeModal}
  />
  
  <!-- Main Content -->
  <MainColumn {isDarkMode}>
    <div class="min-h-screen border-x {isDarkMode ? 'border-gray-800' : 'border-gray-200'}">
      <div class="sticky top-0 z-10 {isDarkMode ? 'bg-black/80' : 'bg-white/80'} backdrop-blur-md border-b {isDarkMode ? 'border-gray-800' : 'border-gray-200'} px-4 py-3">
        <h1 class="text-xl font-bold">Home</h1>
      </div>
      
      {#each tweets as tweet (tweet.id)}
        <TweetCard {tweet} {isDarkMode} />
      {/each}
    </div>
  </MainColumn>
  
  <!-- Right Sidebar -->
  <RightSide {isDarkMode} {trends} suggestedUsers={suggestedUsers} />
  
  <!-- Compose Tweet Modal -->
  {#if showComposeModal}
    <ComposeTweet 
      {isDarkMode}
      on:close={toggleComposeModal}
      avatar={sidebarAvatar} 
    /> 
  {/if}
</div> 