<script lang="ts">
  import { onMount } from 'svelte';
  import MainLayout from '../components/layout/MainLayout.svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import type { IAuthStore } from '../interfaces/IAuth';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { toastStore } from '../stores/toastStore';
  
  const logger = createLoggerWithPrefix('Communities');
  
  const { getAuthState } = useAuth();
  const { theme } = useTheme();
  
  $: authState = getAuthState ? (getAuthState() as IAuthStore) : { userId: null, isAuthenticated: false, accessToken: null, refreshToken: null };
  $: isDarkMode = $theme === 'dark';
  $: sidebarUsername = authState?.userId ? `User_${authState.userId.substring(0, 4)}` : '';
  $: sidebarDisplayName = authState?.userId ? `User ${authState.userId.substring(0, 4)}` : '';
  $: sidebarAvatar = 'https://secure.gravatar.com/avatar/0?d=mp';
  
  interface Community {
    id: number;
    name: string;
    description: string;
    logo: string | null;
    memberCount: number;
    categories: string[];
    status: 'joined' | 'pending' | 'available';
    isPrivate: boolean;
  }
  
  let isLoading = true;
  let joinedCommunities: Community[] = [];
  let pendingCommunities: Community[] = [];
  let availableCommunities: Community[] = [];
  
  let filteredJoinedCommunities: Community[] = [];
  let filteredPendingCommunities: Community[] = [];
  let filteredAvailableCommunities: Community[] = [];
  
  let searchQuery = '';
  let selectedCategories: string[] = [];
  
  const paginationOptions = [25, 30, 35];
  let joinedPerPage = 25;
  let pendingPerPage = 25;
  let availablePerPage = 25;
  
  let joinedCurrentPage = 1;
  let pendingCurrentPage = 1;
  let availableCurrentPage = 1;
  
  const categories = [
    'Gaming', 'Sports', 'Food', 'Technology', 'Art', 'Music', 
    'Movies', 'Books', 'Fitness', 'Travel', 'Fashion', 'Education'
  ];
  
  function checkAuth() {
    if (!authState.isAuthenticated) {
      toastStore.showToast('You need to log in to access communities', 'warning');
      window.location.href = '/login';
      return false;
    }
    return true;
  }
  
  async function fetchCommunities() {
    isLoading = true;
    
    try {
      setTimeout(() => {
        joinedCommunities = [
          {
            id: 1,
            name: 'Gaming Enthusiasts',
            description: 'A community for passionate gamers to discuss the latest games and gaming news.',
            logo: null,
            memberCount: 15420,
            categories: ['Gaming', 'Technology'],
            status: 'joined',
            isPrivate: false
          },
          {
            id: 2,
            name: 'Frontend Developers',
            description: 'Connect with other frontend developers to share knowledge and resources.',
            logo: null,
            memberCount: 8754,
            categories: ['Technology', 'Education'],
            status: 'joined',
            isPrivate: false
          },
          {
            id: 3,
            name: 'Book Club',
            description: 'Discuss your favorite books and discover new reads with fellow book lovers.',
            logo: null,
            memberCount: 5230,
            categories: ['Books', 'Education'],
            status: 'joined',
            isPrivate: true
          }
        ];
        
        pendingCommunities = [
          {
            id: 4,
            name: 'Fitness Freaks',
            description: 'Share workout routines, nutrition tips, and fitness progress with like-minded individuals.',
            logo: null,
            memberCount: 12380,
            categories: ['Fitness', 'Health'],
            status: 'pending',
            isPrivate: true
          },
          {
            id: 5,
            name: 'Movie Buffs',
            description: 'A community for cinema enthusiasts to discuss films, directors, and cinematic techniques.',
            logo: null,
            memberCount: 9840,
            categories: ['Movies', 'Art'],
            status: 'pending',
            isPrivate: false
          }
        ];
        
        availableCommunities = [
          {
            id: 6,
            name: 'Travel Adventures',
            description: 'Share your travel experiences, tips, and stories from around the world.',
            logo: null,
            memberCount: 18650,
            categories: ['Travel', 'Photography'],
            status: 'available',
            isPrivate: false
          },
          {
            id: 7,
            name: 'Music Lovers',
            description: 'Discover new music, share playlists, and discuss your favorite artists and bands.',
            logo: null,
            memberCount: 21470,
            categories: ['Music', 'Entertainment'],
            status: 'available',
            isPrivate: false
          },
          {
            id: 8,
            name: 'Foodies Unite',
            description: 'Share recipes, restaurant recommendations, and culinary experiences.',
            logo: null,
            memberCount: 14250,
            categories: ['Food', 'Lifestyle'],
            status: 'available',
            isPrivate: false
          },
          {
            id: 9,
            name: 'AI & Machine Learning',
            description: 'Discuss the latest advancements in AI, machine learning, and data science.',
            logo: null,
            memberCount: 7630,
            categories: ['Technology', 'Science', 'Education'],
            status: 'available',
            isPrivate: false
          },
          {
            id: 10,
            name: 'Digital Nomads',
            description: 'Connect with others who work remotely while traveling the world.',
            logo: null,
            memberCount: 9320,
            categories: ['Travel', 'Work', 'Lifestyle'],
            status: 'available',
            isPrivate: false
          }
        ];
        
        filteredJoinedCommunities = [...joinedCommunities];
        filteredPendingCommunities = [...pendingCommunities];
        filteredAvailableCommunities = [...availableCommunities];
        
        isLoading = false;
        logger.debug('Communities loaded', { 
          joined: joinedCommunities.length,
          pending: pendingCommunities.length,
          available: availableCommunities.length
        });
      }, 1000);
      
    } catch (error) {
      console.error('Error fetching communities:', error);
      toastStore.showToast('Failed to load communities. Please try again.', 'error');
      isLoading = false;
    }
  }
  
  function filterCommunities() {
    const query = searchQuery.toLowerCase();
    const hasCategories = selectedCategories.length > 0;
    
    filteredJoinedCommunities = joinedCommunities.filter(community => {
      const matchesQuery = query === '' || 
        community.name.toLowerCase().includes(query) || 
        community.description.toLowerCase().includes(query);
      
      const matchesCategories = !hasCategories || 
        community.categories.some(category => selectedCategories.includes(category));
      
      return matchesQuery && matchesCategories;
    });
    
    filteredPendingCommunities = pendingCommunities.filter(community => {
      const matchesQuery = query === '' || 
        community.name.toLowerCase().includes(query) || 
        community.description.toLowerCase().includes(query);
      
      const matchesCategories = !hasCategories || 
        community.categories.some(category => selectedCategories.includes(category));
      
      return matchesQuery && matchesCategories;
    });
    
    filteredAvailableCommunities = availableCommunities.filter(community => {
      const matchesQuery = query === '' || 
        community.name.toLowerCase().includes(query) || 
        community.description.toLowerCase().includes(query);
      
      const matchesCategories = !hasCategories || 
        community.categories.some(category => selectedCategories.includes(category));
      
      return matchesQuery && matchesCategories;
    });
    
    joinedCurrentPage = 1;
    pendingCurrentPage = 1;
    availableCurrentPage = 1;
    
    logger.debug('Communities filtered', { 
      query,
      categories: selectedCategories,
      filteredJoined: filteredJoinedCommunities.length,
      filteredPending: filteredPendingCommunities.length,
      filteredAvailable: filteredAvailableCommunities.length
    });
  }
  
  function toggleCategory(category: string) {
    if (selectedCategories.includes(category)) {
      selectedCategories = selectedCategories.filter(c => c !== category);
    } else {
      selectedCategories = [...selectedCategories, category];
    }
    filterCommunities();
  }
  
  async function requestToJoin(communityId: number) {
    try {
      const community = availableCommunities.find(c => c.id === communityId);
      if (community) {
        community.status = 'pending';
        availableCommunities = availableCommunities.filter(c => c.id !== communityId);
        pendingCommunities = [...pendingCommunities, community];
        
        filteredAvailableCommunities = filteredAvailableCommunities.filter(c => c.id !== communityId);
        filteredPendingCommunities = [...filteredPendingCommunities, community];
        
        logger.debug('Join request sent', { communityId });
        toastStore.showToast('Join request sent', 'success');
      }
    } catch (error) {
      console.error('Error requesting to join community:', error);
      toastStore.showToast('Failed to send join request. Please try again.', 'error');
    }
  }
  
  function navigateToCommunity(communityId: number) {
    window.location.href = `/communities/${communityId}`;
  }
  
  function navigateToCreateCommunity() {
    window.location.href = '/create-community';
  }
  
  function getPaginatedJoinedCommunities() {
    const start = (joinedCurrentPage - 1) * joinedPerPage;
    const end = start + joinedPerPage;
    return filteredJoinedCommunities.slice(start, end);
  }
  
  function getPaginatedPendingCommunities() {
    const start = (pendingCurrentPage - 1) * pendingPerPage;
    const end = start + pendingPerPage;
    return filteredPendingCommunities.slice(start, end);
  }
  
  function getPaginatedAvailableCommunities() {
    const start = (availableCurrentPage - 1) * availablePerPage;
    const end = start + availablePerPage;
    return filteredAvailableCommunities.slice(start, end);
  }
  
  $: if (searchQuery !== undefined) {
    filterCommunities();
  }
  
  onMount(() => {
    logger.debug('Communities page mounted', { authState });
    if (checkAuth()) {
      fetchCommunities();
    }
  });
</script>

<MainLayout
  username={sidebarUsername}
  displayName={sidebarDisplayName}
  avatar={sidebarAvatar}
  on:toggleComposeModal={() => {}}
>
  <div class="min-h-screen border-x border-gray-200 dark:border-gray-800">
    <!-- Header -->
    <div class="sticky top-0 z-10 bg-white/80 dark:bg-black/80 backdrop-blur-md border-b border-gray-200 dark:border-gray-800 px-4 py-3">
      <div class="flex items-center justify-between">
        <h1 class="text-xl font-bold">Communities</h1>
        <button 
          class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-full text-sm font-bold"
          on:click={navigateToCreateCommunity}
        >
          Create Community
        </button>
      </div>
      
      <!-- Search -->
      <div class="relative mt-3">
        <input 
          type="text" 
          bind:value={searchQuery}
          placeholder="Search Communities" 
          class="w-full rounded-full pl-10 pr-4 py-2 {isDarkMode ? 'bg-gray-800 border-gray-700 text-white' : 'bg-gray-100 border-gray-200'}"
        />
        <div class="absolute left-3 top-2.5 text-gray-500">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </div>
      </div>
      
      <!-- Categories Filter -->
      <div class="mt-3 flex flex-wrap gap-2">
        {#each categories as category}
          <button 
            class="px-3 py-1 rounded-full text-sm {selectedCategories.includes(category) ? 'bg-blue-500 text-white' : 'bg-gray-200 dark:bg-gray-700 text-gray-800 dark:text-gray-200'}"
            on:click={() => toggleCategory(category)}
          >
            {category}
          </button>
        {/each}
      </div>
    </div>
    
    <!-- Content -->
    <div class="p-4 space-y-8">
      {#if isLoading}
        <div class="animate-pulse space-y-6">
          {#each Array(3) as _}
            <div>
              <div class="h-6 bg-gray-300 dark:bg-gray-700 rounded w-1/4 mb-4"></div>
              {#each Array(3) as _}
                <div class="rounded-lg bg-gray-300 dark:bg-gray-700 h-24 mb-3"></div>
              {/each}
            </div>
          {/each}
        </div>
      {:else}
        <!-- Joined Communities Section -->
        <div>
          <h2 class="text-lg font-bold mb-3">Your Communities</h2>
          
          {#if filteredJoinedCommunities.length === 0}
            <div class="bg-gray-100 dark:bg-gray-800 rounded-lg p-4 text-center">
              <p class="text-gray-500 dark:text-gray-400">
                {joinedCommunities.length === 0 
                  ? "You haven't joined any communities yet." 
                  : "No communities match your search criteria."}
              </p>
            </div>
          {:else}
            <!-- Pagination Controls -->
            <div class="flex justify-between items-center mb-3">
              <div class="flex items-center space-x-2">
                <span class="text-sm text-gray-500 dark:text-gray-400">Show:</span>
                <select 
                  bind:value={joinedPerPage}
                  class="bg-gray-100 dark:bg-gray-800 rounded px-2 py-1 text-sm"
                >
                  {#each paginationOptions as option}
                    <option value={option}>{option}</option>
                  {/each}
                </select>
              </div>
              
              <div class="flex items-center space-x-2">
                <button 
                  class="px-3 py-1 rounded bg-gray-200 dark:bg-gray-700 text-sm {joinedCurrentPage === 1 ? 'opacity-50 cursor-not-allowed' : 'hover:bg-gray-300 dark:hover:bg-gray-600'}"
                  disabled={joinedCurrentPage === 1}
                  on:click={() => joinedCurrentPage--}
                >
                  Previous
                </button>
                <span class="text-sm">
                  Page {joinedCurrentPage} of {Math.ceil(filteredJoinedCommunities.length / joinedPerPage) || 1}
                </span>
                <button 
                  class="px-3 py-1 rounded bg-gray-200 dark:bg-gray-700 text-sm {joinedCurrentPage >= Math.ceil(filteredJoinedCommunities.length / joinedPerPage) ? 'opacity-50 cursor-not-allowed' : 'hover:bg-gray-300 dark:hover:bg-gray-600'}"
                  disabled={joinedCurrentPage >= Math.ceil(filteredJoinedCommunities.length / joinedPerPage)}
                  on:click={() => joinedCurrentPage++}
                >
                  Next
                </button>
              </div>
            </div>
            
            <!-- Community List -->
            <div class="space-y-3">
              {#each getPaginatedJoinedCommunities() as community}
                <div 
                  class="relative p-4 border border-gray-200 dark:border-gray-800 rounded-xl hover:bg-gray-50 dark:hover:bg-gray-900 transition cursor-pointer"
                  role="button"
                  tabindex="0"
                  on:click={() => navigateToCommunity(community.id)}
                  on:keydown={(e) => e.key === 'Enter' && navigateToCommunity(community.id)}
                >
                  <div class="flex items-start">
                    <div class="w-12 h-12 rounded-full bg-gray-300 dark:bg-gray-600 flex items-center justify-center text-xl flex-shrink-0">
                      {#if community.logo}
                        <img src={community.logo} alt={community.name} class="w-full h-full rounded-full object-cover" />
                      {:else}
                        <span>{community.name.charAt(0)}</span>
                      {/if}
                    </div>
                    <div class="ml-3 flex-1">
                      <div class="flex items-center">
                        <h3 
                          class="text-lg font-semibold hover:underline cursor-pointer"
                          role="button"
                          tabindex="0"
                          on:click={() => navigateToCommunity(community.id)}
                          on:keydown={(e) => e.key === 'Enter' && navigateToCommunity(community.id)}
                        >
                          {community.name}
                        </h3>
                        {#if community.isPrivate}
                          <span class="ml-2 px-2 py-0.5 bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded text-xs">Private</span>
                        {/if}
                      </div>
                      <p class="text-gray-600 dark:text-gray-400 text-sm mt-1 line-clamp-2">{community.description}</p>
                      <div class="flex items-center mt-2">
                        <span class="text-sm text-gray-500 dark:text-gray-400">{community.memberCount.toLocaleString()} members</span>
                        <div class="mx-3 h-1 w-1 rounded-full bg-gray-300 dark:bg-gray-600"></div>
                        <div class="flex flex-wrap gap-1">
                          {#each community.categories as category}
                            <span class="text-xs px-2 py-0.5 bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded">{category}</span>
                          {/each}
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
        
        <!-- Pending Communities Section -->
        <div>
          <h2 class="text-lg font-bold mb-3">Pending Requests</h2>
          
          {#if filteredPendingCommunities.length === 0}
            <div class="bg-gray-100 dark:bg-gray-800 rounded-lg p-4 text-center">
              <p class="text-gray-500 dark:text-gray-400">
                {pendingCommunities.length === 0 
                  ? "You don't have any pending community requests." 
                  : "No pending requests match your search criteria."}
              </p>
            </div>
          {:else}
            <!-- Pagination Controls -->
            <div class="flex justify-between items-center mb-3">
              <div class="flex items-center space-x-2">
                <span class="text-sm text-gray-500 dark:text-gray-400">Show:</span>
                <select 
                  bind:value={pendingPerPage}
                  class="bg-gray-100 dark:bg-gray-800 rounded px-2 py-1 text-sm"
                >
                  {#each paginationOptions as option}
                    <option value={option}>{option}</option>
                  {/each}
                </select>
              </div>
              
              <div class="flex items-center space-x-2">
                <button 
                  class="px-3 py-1 rounded bg-gray-200 dark:bg-gray-700 text-sm {pendingCurrentPage === 1 ? 'opacity-50 cursor-not-allowed' : 'hover:bg-gray-300 dark:hover:bg-gray-600'}"
                  disabled={pendingCurrentPage === 1}
                  on:click={() => pendingCurrentPage--}
                >
                  Previous
                </button>
                <span class="text-sm">
                  Page {pendingCurrentPage} of {Math.ceil(filteredPendingCommunities.length / pendingPerPage) || 1}
                </span>
                <button 
                  class="px-3 py-1 rounded bg-gray-200 dark:bg-gray-700 text-sm {pendingCurrentPage >= Math.ceil(filteredPendingCommunities.length / pendingPerPage) ? 'opacity-50 cursor-not-allowed' : 'hover:bg-gray-300 dark:hover:bg-gray-600'}"
                  disabled={pendingCurrentPage >= Math.ceil(filteredPendingCommunities.length / pendingPerPage)}
                  on:click={() => pendingCurrentPage++}
                >
                  Next
                </button>
              </div>
            </div>
            
            <!-- Community List -->
            <div class="space-y-3">
              {#each getPaginatedPendingCommunities() as community}
                <div 
                  class="relative p-4 border border-gray-200 dark:border-gray-800 rounded-xl hover:bg-gray-50 dark:hover:bg-gray-900 transition cursor-pointer"
                  role="button"
                  tabindex="0"
                  on:click={() => navigateToCommunity(community.id)}
                  on:keydown={(e) => e.key === 'Enter' && navigateToCommunity(community.id)}
                >
                  <div class="flex items-start">
                    <div class="w-12 h-12 rounded-full bg-gray-300 dark:bg-gray-600 flex items-center justify-center text-xl flex-shrink-0">
                      {#if community.logo}
                        <img src={community.logo} alt={community.name} class="w-full h-full rounded-full object-cover" />
                      {:else}
                        <span>{community.name.charAt(0)}</span>
                      {/if}
                    </div>
                    <div class="ml-3 flex-1">
                      <div class="flex items-center">
                        <h3 
                          class="text-lg font-semibold hover:underline cursor-pointer"
                          role="button"
                          tabindex="0"
                          on:click={() => navigateToCommunity(community.id)}
                          on:keydown={(e) => e.key === 'Enter' && navigateToCommunity(community.id)}
                        >
                          {community.name}
                        </h3>
                        {#if community.isPrivate}
                          <span class="ml-2 px-2 py-0.5 bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded text-xs">Private</span>
                        {/if}
                        <span class="ml-2 px-2 py-0.5 bg-yellow-100 dark:bg-yellow-900 text-yellow-800 dark:text-yellow-200 rounded text-xs">Pending</span>
                      </div>
                      <p class="text-gray-600 dark:text-gray-400 text-sm mt-1 line-clamp-2">{community.description}</p>
                      <div class="flex items-center mt-2">
                        <span class="text-sm text-gray-500 dark:text-gray-400">{community.memberCount.toLocaleString()} members</span>
                        <div class="mx-3 h-1 w-1 rounded-full bg-gray-300 dark:bg-gray-600"></div>
                        <div class="flex flex-wrap gap-1">
                          {#each community.categories as category}
                            <span class="text-xs px-2 py-0.5 bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded">{category}</span>
                          {/each}
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
        
        <!-- Available Communities Section -->
        <div>
          <h2 class="text-lg font-bold mb-3">Discover Communities</h2>
          
          {#if filteredAvailableCommunities.length === 0}
            <div class="bg-gray-100 dark:bg-gray-800 rounded-lg p-4 text-center">
              <p class="text-gray-500 dark:text-gray-400">
                {availableCommunities.length === 0 
                  ? "There are no available communities at the moment." 
                  : "No available communities match your search criteria."}
              </p>
            </div>
          {:else}
            <!-- Pagination Controls -->
            <div class="flex justify-between items-center mb-3">
              <div class="flex items-center space-x-2">
                <span class="text-sm text-gray-500 dark:text-gray-400">Show:</span>
                <select 
                  bind:value={availablePerPage}
                  class="bg-gray-100 dark:bg-gray-800 rounded px-2 py-1 text-sm"
                >
                  {#each paginationOptions as option}
                    <option value={option}>{option}</option>
                  {/each}
                </select>
              </div>
              
              <div class="flex items-center space-x-2">
                <button 
                  class="px-3 py-1 rounded bg-gray-200 dark:bg-gray-700 text-sm {availableCurrentPage === 1 ? 'opacity-50 cursor-not-allowed' : 'hover:bg-gray-300 dark:hover:bg-gray-600'}"
                  disabled={availableCurrentPage === 1}
                  on:click={() => availableCurrentPage--}
                >
                  Previous
                </button>
                <span class="text-sm">
                  Page {availableCurrentPage} of {Math.ceil(filteredAvailableCommunities.length / availablePerPage) || 1}
                </span>
                <button 
                  class="px-3 py-1 rounded bg-gray-200 dark:bg-gray-700 text-sm {availableCurrentPage >= Math.ceil(filteredAvailableCommunities.length / availablePerPage) ? 'opacity-50 cursor-not-allowed' : 'hover:bg-gray-300 dark:hover:bg-gray-600'}"
                  disabled={availableCurrentPage >= Math.ceil(filteredAvailableCommunities.length / availablePerPage)}
                  on:click={() => availableCurrentPage++}
                >
                  Next
                </button>
              </div>
            </div>
            
            <!-- Community List -->
            <div class="space-y-3">
              {#each getPaginatedAvailableCommunities() as community}
                <div class="bg-white dark:bg-gray-800 rounded-lg p-4 border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700 transition">
                  <div class="flex items-start">
                    <div class="w-12 h-12 rounded-full bg-gray-300 dark:bg-gray-600 flex items-center justify-center text-xl flex-shrink-0">
                      {#if community.logo}
                        <img src={community.logo} alt={community.name} class="w-full h-full rounded-full object-cover" />
                      {:else}
                        <span>{community.name.charAt(0)}</span>
                      {/if}
                    </div>
                    <div class="ml-3 flex-1">
                      <div class="flex items-center">
                        <h3 
                          class="text-lg font-semibold hover:underline cursor-pointer"
                          role="button"
                          tabindex="0"
                          on:click={() => navigateToCommunity(community.id)}
                          on:keydown={(e) => e.key === 'Enter' && navigateToCommunity(community.id)}
                        >
                          {community.name}
                        </h3>
                        {#if community.isPrivate}
                          <span class="ml-2 px-2 py-0.5 bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded text-xs">Private</span>
                        {/if}
                      </div>
                      <p class="text-gray-600 dark:text-gray-400 text-sm mt-1 line-clamp-2">{community.description}</p>
                      <div class="flex items-center justify-between mt-2">
                        <div class="flex items-center">
                          <span class="text-sm text-gray-500 dark:text-gray-400">{community.memberCount.toLocaleString()} members</span>
                          <div class="mx-3 h-1 w-1 rounded-full bg-gray-300 dark:bg-gray-600"></div>
                          <div class="flex flex-wrap gap-1">
                            {#each community.categories as category}
                              <span class="text-xs px-2 py-0.5 bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded">{category}</span>
                            {/each}
                          </div>
                        </div>
                        <button 
                          class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-1 rounded-full text-sm"
                          on:click|stopPropagation={() => requestToJoin(community.id)}
                        >
                          Request to Join
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      {/if}
    </div>
  </div>
</MainLayout>

<style>
  /* Limit text to 2 lines */
  .line-clamp-2 {
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
  
  /* Skeleton loading animation */
  @keyframes pulse {
    0%, 100% { opacity: 0.5; }
    50% { opacity: 1; }
  }
  .animate-pulse {
    animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }
  
  .community-description {
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  /* Tabs */
  .tab-active {
    color: rgb(29, 155, 240);
    font-weight: 600;
  }
  
  .tab-active::after {
    content: '';
    position: absolute;
    bottom: -1px;
    left: 0;
    right: 0;
    height: 4px;
    background-color: rgb(29, 155, 240);
    border-radius: 9999px;
  }
</style>
