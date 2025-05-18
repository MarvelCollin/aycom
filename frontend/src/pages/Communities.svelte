<script lang="ts">
  import { onMount } from 'svelte';
  import MainLayout from '../components/layout/MainLayout.svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import type { IAuthStore } from '../interfaces/IAuth';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { toastStore } from '../stores/toastStore';
  
  // Import icons
  import SearchIcon from 'svelte-feather-icons/src/icons/SearchIcon.svelte';
  import FilterIcon from 'svelte-feather-icons/src/icons/FilterIcon.svelte';
  import CheckIcon from 'svelte-feather-icons/src/icons/CheckIcon.svelte';
  import UsersIcon from 'svelte-feather-icons/src/icons/UsersIcon.svelte';
  import LockIcon from 'svelte-feather-icons/src/icons/LockIcon.svelte';
  import PlusIcon from 'svelte-feather-icons/src/icons/PlusIcon.svelte';
  import AlertCircleIcon from 'svelte-feather-icons/src/icons/AlertCircleIcon.svelte';
  
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
  
  // Current active tab
  let activeTab = 'discover'; // 'joined', 'pending', 'discover'
  
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
  }
  
  function toggleCategory(category: string) {
    if (selectedCategories.includes(category)) {
      selectedCategories = selectedCategories.filter(c => c !== category);
    } else {
      selectedCategories = [...selectedCategories, category];
    }
    
    filterCommunities();
  }
  
  function joinCommunity(communityId: number) {
    if (!authState.isAuthenticated) {
      toastStore.showToast('You need to log in to join communities', 'warning');
      return;
    }
    
    logger.debug('Join community', { communityId });
    
    // Find the community in the available list
    const communityIndex = availableCommunities.findIndex(c => c.id === communityId);
    
    if (communityIndex !== -1) {
      const community = availableCommunities[communityIndex];
      
      // If it's a private community, move to pending
      if (community.isPrivate) {
        // Update status
        community.status = 'pending';
        
        // Remove from available list
        availableCommunities = availableCommunities.filter(c => c.id !== communityId);
        
        // Add to pending list
        pendingCommunities = [community, ...pendingCommunities];
        
        // Update filtered lists
        filterCommunities();
        
        toastStore.showToast('Your request to join has been sent to the community moderators.', 'success');
      } else {
        // If it's public, move to joined
        community.status = 'joined';
        
        // Remove from available list
        availableCommunities = availableCommunities.filter(c => c.id !== communityId);
        
        // Add to joined list
        joinedCommunities = [community, ...joinedCommunities];
        
        // Update filtered lists
        filterCommunities();
        
        toastStore.showToast('You have joined the community!', 'success');
      }
    }
  }
  
  function leaveCommunity(communityId: number) {
    logger.debug('Leave community', { communityId });
    
    // Find the community in the joined list
    const communityIndex = joinedCommunities.findIndex(c => c.id === communityId);
    
    if (communityIndex !== -1) {
      const community = joinedCommunities[communityIndex];
      
      // Update status
      community.status = 'available';
      
      // Remove from joined list
      joinedCommunities = joinedCommunities.filter(c => c.id !== communityId);
      
      // Add to available list
      availableCommunities = [community, ...availableCommunities];
      
      // Update filtered lists
      filterCommunities();
      
      toastStore.showToast('You have left the community.', 'info');
    }
  }
  
  function cancelRequest(communityId: number) {
    logger.debug('Cancel request', { communityId });
    
    // Find the community in the pending list
    const communityIndex = pendingCommunities.findIndex(c => c.id === communityId);
    
    if (communityIndex !== -1) {
      const community = pendingCommunities[communityIndex];
      
      // Update status
      community.status = 'available';
      
      // Remove from pending list
      pendingCommunities = pendingCommunities.filter(c => c.id !== communityId);
      
      // Add to available list
      availableCommunities = [community, ...availableCommunities];
      
      // Update filtered lists
      filterCommunities();
      
      toastStore.showToast('Your request has been cancelled.', 'info');
    }
  }
  
  function clearFilters() {
    searchQuery = '';
    selectedCategories = [];
    filterCommunities();
  }
  
  function setActiveTab(tabName) {
    activeTab = tabName;
  }
  
  function handleCommunityClick(communityId) {
    window.location.href = `/communities/${communityId}`;
  }
  
  onMount(() => {
    if (checkAuth()) {
      fetchCommunities();
    }
  });
  
  $: {
    if (searchQuery !== undefined) {
      filterCommunities();
    }
  }
</script>

<MainLayout username={sidebarUsername} displayName={sidebarDisplayName} avatar={sidebarAvatar}>
  <div class="communities-container">
    <div class="communities-header">
      <h1 class="communities-title">Communities</h1>
      
      <div class="search-filter-container">
        <div class="search-input-wrapper">
          <div class="search-icon">
            <SearchIcon size="16" />
          </div>
          <input 
            type="text" 
            bind:value={searchQuery} 
            placeholder="Search communities" 
            class="search-input"
          />
        </div>
        
        <button class="filter-button" on:click={clearFilters}>
          <FilterIcon size="16" />
          <span>Clear Filters</span>
        </button>
      </div>
      
      <div class="categories-wrapper">
        {#each categories as category}
          <button 
            class="category-chip {selectedCategories.includes(category) ? 'selected' : ''}"
            on:click={() => toggleCategory(category)}
          >
            {category}
            {#if selectedCategories.includes(category)}
              <span><CheckIcon size="12" /></span>
            {/if}
          </button>
        {/each}
      </div>
    </div>
    
    <!-- Tabs Navigation -->
    <div class="communities-tabs">
      <button 
        class="communities-tab {activeTab === 'discover' ? 'active' : ''}" 
        on:click={() => setActiveTab('discover')}
      >
        Discover
      </button>
      <button 
        class="communities-tab {activeTab === 'joined' ? 'active' : ''}"
        on:click={() => setActiveTab('joined')}
      >
        Joined
        {#if filteredJoinedCommunities.length > 0}
          <span class="tab-count">{filteredJoinedCommunities.length}</span>
        {/if}
      </button>
      <button 
        class="communities-tab {activeTab === 'pending' ? 'active' : ''}"
        on:click={() => setActiveTab('pending')}
      >
        Pending
        {#if filteredPendingCommunities.length > 0}
          <span class="tab-count">{filteredPendingCommunities.length}</span>
        {/if}
      </button>
    </div>
    
    {#if isLoading}
      <div class="empty-state">
        <div class="empty-icon">⌛</div>
        <p class="empty-title">Loading communities...</p>
      </div>
    {:else}
      <!-- Discover Tab Content -->
      {#if activeTab === 'discover'}
        {#if filteredAvailableCommunities.length > 0}
          <div class="communities-grid">
            {#each filteredAvailableCommunities as community (community.id)}
              <div class="community-card" on:click={() => handleCommunityClick(community.id)}>
                <div class="community-banner">
                  <div class="community-avatar">
                    {community.name[0].toUpperCase()}
                  </div>
                </div>
                <div class="community-content">
                  <div class="community-header">
                    <h3 class="community-name">{community.name}</h3>
                    <p class="community-handle">#{community.categories[0]}</p>
                  </div>
                  
                  <p class="community-description">{community.description}</p>
                  
                  <div class="community-meta">
                    <div class="community-stats">
                      <div class="community-stat">
                        <UsersIcon size="14" />
                        <span>{community.memberCount.toLocaleString()}</span>
                      </div>
                    </div>
                    
                    <div class="community-type">
                      {#if community.isPrivate}
                        <LockIcon size="14" />
                        <span>Private</span>
                      {/if}
                    </div>
                  </div>
                  
                  <div class="community-action">
                    <button 
                      class="join-button"
                      on:click={() => joinCommunity(community.id)}
                    >
                      {community.isPrivate ? 'Request to Join' : 'Join'}
                    </button>
                  </div>
                </div>
              </div>
            {/each}
          </div>
        {:else}
          <div class="empty-state">
            <div class="empty-icon">🔍</div>
            <h3 class="empty-title">No communities found</h3>
            <p class="empty-text">We couldn't find any communities matching your search criteria.</p>
            <button class="create-community-button" on:click={clearFilters}>
              <FilterIcon size="16" />
              Clear Filters
            </button>
          </div>
        {/if}
      {/if}
      
      <!-- Joined Tab Content -->
      {#if activeTab === 'joined'}
        {#if filteredJoinedCommunities.length > 0}
          <div class="communities-grid">
            {#each filteredJoinedCommunities as community (community.id)}
              <div class="community-card" on:click={() => handleCommunityClick(community.id)}>
                <div class="community-banner">
                  <div class="community-avatar">
                    {community.name[0].toUpperCase()}
                  </div>
                </div>
                <div class="community-content">
                  <div class="community-header">
                    <h3 class="community-name">{community.name}</h3>
                    <p class="community-handle">#{community.categories[0]}</p>
                  </div>
                  
                  <p class="community-description">{community.description}</p>
                  
                  <div class="community-meta">
                    <div class="community-stats">
                      <div class="community-stat">
                        <UsersIcon size="14" />
                        <span>{community.memberCount.toLocaleString()}</span>
                      </div>
                    </div>
                    
                    <div class="community-type">
                      {#if community.isPrivate}
                        <LockIcon size="14" />
                        <span>Private</span>
                      {/if}
                    </div>
                  </div>
                  
                  <div class="community-action">
                    <button 
                      class="joined-button"
                      on:click={() => leaveCommunity(community.id)}
                    >
                      Leave
                    </button>
                  </div>
                </div>
              </div>
            {/each}
          </div>
        {:else}
          <div class="empty-state">
            <div class="empty-icon">👋</div>
            <h3 class="empty-title">You haven't joined any communities yet</h3>
            <p class="empty-text">Join some communities to see them here!</p>
            <button class="create-community-button" on:click={() => setActiveTab('discover')}>
              <PlusIcon size="16" />
              Discover Communities
            </button>
          </div>
        {/if}
      {/if}
      
      <!-- Pending Tab Content -->
      {#if activeTab === 'pending'}
        {#if filteredPendingCommunities.length > 0}
          <div class="communities-grid">
            {#each filteredPendingCommunities as community (community.id)}
              <div class="community-card" on:click={() => handleCommunityClick(community.id)}>
                <div class="community-banner">
                  <div class="community-avatar">
                    {community.name[0].toUpperCase()}
                  </div>
                </div>
                <div class="community-content">
                  <div class="community-header">
                    <h3 class="community-name">{community.name}</h3>
                    <p class="community-handle">#{community.categories[0]}</p>
                  </div>
                  
                  <p class="community-description">{community.description}</p>
                  
                  <div class="community-meta">
                    <div class="community-stats">
                      <div class="community-stat">
                        <UsersIcon size="14" />
                        <span>{community.memberCount.toLocaleString()}</span>
                      </div>
                    </div>
                    
                    <div class="community-type">
                      {#if community.isPrivate}
                        <LockIcon size="14" />
                        <span>Private</span>
                      {/if}
                    </div>
                  </div>
                  
                  <div class="community-action">
                    <button 
                      class="joined-button"
                      on:click={() => cancelRequest(community.id)}
                    >
                      Cancel Request
                    </button>
                  </div>
                </div>
              </div>
            {/each}
          </div>
        {:else}
          <div class="empty-state">
            <div class="empty-icon">✉️</div>
            <h3 class="empty-title">No pending requests</h3>
            <p class="empty-text">You don't have any pending join requests.</p>
            <button class="create-community-button" on:click={() => setActiveTab('discover')}>
              <PlusIcon size="16" />
              Discover Communities
            </button>
          </div>
        {/if}
      {/if}
    {/if}
  </div>
</MainLayout>

<style>
  .categories-wrapper {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    margin-top: var(--space-3);
  }
  
  .category-chip {
    background-color: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-full);
    padding: var(--space-1) var(--space-3);
    font-size: var(--font-size-xs);
    color: var(--text-secondary);
    cursor: pointer;
    transition: all var(--transition-fast);
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
  }
  
  .category-chip:hover {
    background-color: var(--bg-hover);
    color: var(--text-primary);
  }
  
  .category-chip.selected {
    background-color: var(--color-primary);
    color: white;
    border-color: var(--color-primary);
  }
  
  .tab-count {
    background-color: var(--bg-secondary);
    color: var(--text-secondary);
    border-radius: var(--radius-full);
    padding: 0 var(--space-1);
    font-size: var(--font-size-xs);
    margin-left: var(--space-1);
    min-width: 16px;
    height: 16px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }
  
  .communities-tab.active .tab-count {
    background-color: var(--color-primary);
    color: white;
  }
  
  .community-card {
    cursor: pointer;
    transition: transform 0.2s ease, box-shadow 0.2s ease;
  }
  
  .community-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }
</style>
