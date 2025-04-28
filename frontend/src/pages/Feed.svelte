<script lang="ts">
  import Sidebar from '../components/navigation/Sidebar.svelte';
  import RightSidebar from '../components/navigation/RightSidebar.svelte';
  import TweetComposer from '../components/forms/TweetComposer.svelte';
  import Tweet from '../components/common/Tweet.svelte';
  import { useTweets } from '../hooks/useTweets';
  
  // Get tweets and tweet functions from our hook
  const { tweets, addTweet } = useTweets();
  
  // Mock user info (should come from auth store in a real app)
  const currentUser = {
    username: 'username',
    displayName: 'User Name',
    avatar: '👤'
  };
  
  // Define the type for the custom event
  interface TweetEvent {
    detail: {
      content: string;
    };
  }
  
  // Handle new tweet creation
  function handleNewTweet(event: TweetEvent) {
    const tweetContent = event.detail.content;
    addTweet(tweetContent, currentUser);
  }
</script>

<div class="min-h-screen bg-black text-white">
  <!-- Navigation Sidebar -->
  <Sidebar 
    username={currentUser.username} 
    displayName={currentUser.displayName} 
    avatar={currentUser.avatar} 
  />
  
  <!-- Main Content -->
  <div class="ml-16 md:ml-64">
    <!-- Header -->
    <header class="sticky top-0 bg-black bg-opacity-80 backdrop-blur-sm z-10 p-4 border-b border-gray-800">
      <h1 class="text-xl font-bold">Home</h1>
    </header>
    
    <!-- Tweet Composer -->
    <TweetComposer 
      avatar={currentUser.avatar}
      on:tweet={handleNewTweet} 
    />
    
    <!-- Tweet Feed -->
    <div>
      {#each $tweets as tweet (tweet.id)}
        <Tweet {tweet} />
      {/each}
    </div>
  </div>
  
  <!-- Right Sidebar -->
  <RightSidebar />
</div> 