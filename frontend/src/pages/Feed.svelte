<script lang="ts">
  import SidebarNav from '../components/navigation/SidebarNav.svelte';
  import ComposeTweet from '../components/social/ComposeTweet.svelte';
  import TweetCard from '../components/social/TweetCard.svelte';
  import TrendsList from '../components/social/TrendsList.svelte';
  import SuggestedFollows from '../components/social/SuggestedFollows.svelte';
  import SearchBar from '../components/social/SearchBar.svelte';
  import type { ITweet } from '../interfaces/ISocialMedia';

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
    }
  ];
  
  interface TweetEvent {
    detail: {
      content: string;
    }
  }
  
  function handleTweet(event: TweetEvent) {
    const { content } = event.detail;
    console.log('New tweet from Feed page:', content);
    // Here you would typically send the tweet to an API
  }
</script>

<div class="min-h-screen bg-black text-white">
  <!-- Navigation Sidebar -->
  <SidebarNav />
  
  <!-- Main Content -->
  <div class="ml-16 md:ml-64">
    <!-- Header -->
    <header class="sticky top-0 bg-black bg-opacity-80 backdrop-blur-sm z-10 p-4 border-b border-gray-800">
      <h1 class="text-xl font-bold">Home</h1>
    </header>
    
    <!-- Tweet Compose -->
    <ComposeTweet on:tweet={handleTweet} />
    
    <!-- Tweet Feed -->
    <div>
      {#each tweets as tweet (tweet.id)}
        <TweetCard {tweet} />
      {/each}
    </div>
  </div>
  
  <!-- Right Sidebar -->
  <div class="hidden lg:block fixed top-0 bottom-0 right-0 w-80 border-l border-gray-800 p-4 overflow-y-auto">
    <!-- Search -->
    <SearchBar />
    
    <!-- Trends -->
    <TrendsList />
    
    <!-- Who to follow -->
    <SuggestedFollows />
  </div>
</div> 