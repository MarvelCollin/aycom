<script lang="ts">
  import { onMount } from 'svelte';
  import MainLayout from '../components/layout/MainLayout.svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { toastStore } from '../stores/toastStore';
  import { 
    getJoinedCommunities,
    getPendingCommunities, 
    getDiscoverCommunities,
    getCategories,
    requestToJoin, 
    checkUserCommunityMembership
  } from '../api/community';
  import type { ICategoriesResponse, ICategory } from '../interfaces/ICategory';
  import { getPublicUrl, SUPABASE_BUCKETS } from '../utils/supabase';
  
  // Import components
  import Pagination from '../components/common/Pagination.svelte';
  import CategoryFilter from '../components/common/CategoryFilter.svelte';
  import Spinner from '../components/common/Spinner.svelte';
  import CreateCommunityModal from '../components/communities/CreateCommunityModal.svelte';
  
  // Import icons
  import SearchIcon from 'svelte-feather-icons/src/icons/SearchIcon.svelte';
  import FilterIcon from 'svelte-feather-icons/src/icons/FilterIcon.svelte';
  import CheckIcon from 'svelte-feather-icons/src/icons/CheckIcon.svelte';
  import UsersIcon from 'svelte-feather-icons/src/icons/UsersIcon.svelte';
  import LockIcon from 'svelte-feather-icons/src/icons/LockIcon.svelte';
  import PlusIcon from 'svelte-feather-icons/src/icons/PlusIcon.svelte';
  import AlertCircleIcon from 'svelte-feather-icons/src/icons/AlertCircleIcon.svelte';
  
  const logger = createLoggerWithPrefix('Communities');
  
  // Auth setup
  const { getAuthState } = useAuth();
  const { theme } = useTheme();
  let authState = getAuthState();
  $: isDarkMode = $theme === 'dark';
  
  // Community creation modal state
  let isCreateModalOpen = false;
  
  // Pagination configuration
  let limitOptions = [25, 50, 100];
  let currentPage = 1;
  let limit = limitOptions[0];
  let totalCount = 0;  // Total number of communities
  let totalPages = 1;
  
  // Filter settings
  let activeTab = 'joined'; // Default to 'joined' tab
  let searchQuery = '';
  let selectedCategories: string[] = [];
  let availableCategories: string[] = [];
  
  // Community data
  let isLoading = false;
  let communities: any[] = []; // Communities for the current tab
  let error: string | null = null;
  
  // Map to track membership status for each community in discover tab
  let communityMembershipStatus = new Map();
  
  // Helper function to calculate total pages
  function calculateTotalPages(total: number, perPage: number): number {
    if (total <= 0 || perPage <= 0) return 1;
    return Math.ceil(total / perPage);
  }
  
  // Fetch communities based on active tab
  async function fetchCommunities() {
    isLoading = true;
    error = null;
    
    try {
      // Prepare parameters for API call
      const params = {
        page: currentPage,
        limit: limit,
        q: searchQuery,
        category: selectedCategories
      };
      
      let result;
      logger.info(`Fetching communities for tab: ${activeTab}`);
      
      try {
        // Use appropriate API call based on active tab
        switch (activeTab) {
          case 'joined':
            result = await getJoinedCommunities(authState.user_id || '', params);
            break;
          case 'pending':
            result = await getPendingCommunities(authState.user_id || '', params);
            break;
          case 'discover':
            result = await getDiscoverCommunities(authState.user_id || '', params);
            break;
          default:
            result = await getJoinedCommunities(authState.user_id || '', params);
        }
      } catch (apiError) {
        // Handle API call errors gracefully
        logger.error(`API error in ${activeTab} tab:`, apiError);
        result = {
          success: true,
          communities: [],
          total: 0,
          page: currentPage,
          limit
        };
      }
      
      // Add more detailed logging
      console.log(`[Communities] Raw API response for ${activeTab} tab:`, JSON.stringify(result));
      logger.info('[Communities] API response:', result);
      
      // Handle response data
      if (result && result.success !== false) {
        // Extract communities from the result
        communities = result.communities || [];
        totalCount = result.total || 0;
        
        console.log('[Communities] Extracted communities:', communities);
        console.log('[Communities] Total count:', totalCount);
        
        // Make sure communities is always an array
        if (!Array.isArray(communities)) {
          console.log('[Communities] Communities is not an array, resetting to empty array');
          communities = [];
        }
        
        // Set up pagination
        totalPages = calculateTotalPages(totalCount, limit);
        if (currentPage > totalPages && totalPages > 0) {
          currentPage = totalPages;
        }
        
        // Check membership status for communities in the discover tab
        if (activeTab === 'discover') {
          checkMembershipStatusForAll();
        }
        
        error = null;
      } else {
        // Handle API error by showing empty data
        logger.error('[Communities] Error fetching communities:', result?.error || 'Unknown error');
        communities = [];
        totalCount = 0;
        totalPages = 1;
        error = result?.error?.message || 'Failed to load communities';
      }
    } catch (err) {
      // Handle unexpected errors
      logger.error('[Communities] Exception fetching communities:', err);
      communities = [];
      totalCount = 0;
      totalPages = 1;
      error = 'An unexpected error occurred';
    } finally {
      isLoading = false;
    }
  }
  
  // Fetch available categories
  async function fetchCategories() {
    try {
      const response = await getCategories();
      
      // Handle both array response and object with categories property
      if (Array.isArray(response)) {
        availableCategories = response.map(cat => cat.name);
      } else if (response && typeof response === 'object' && 'categories' in response) {
        // Use type assertion to tell TypeScript this is a valid object with categories property
        const typedResponse = response as { categories: ICategory[] };
        availableCategories = typedResponse.categories.map(cat => cat.name);
      }
    } catch (error) {
      logger.error('Error fetching categories:', error);
    }
  }
  
  // Check membership status for all communities in discover tab
  async function checkMembershipStatusForAll() {
    if (!Array.isArray(communities) || communities.length === 0 || activeTab !== 'discover') return;
    
    for (const community of communities) {
      if (community && community.id) {
        try {
          const status = await checkMembershipStatus(community.id);
          communityMembershipStatus.set(community.id, status);
          // Force update for reactivity
          communityMembershipStatus = communityMembershipStatus;
        } catch (err) {
          logger.error(`Error checking membership for community ${community.id}:`, err);
        }
      }
    }
  }
  
  // Function to check membership status for a community
  async function checkMembershipStatus(communityId: string): Promise<string> {
    if (!authState.is_authenticated || !communityId) return 'none';
    
    try {
      const membershipResponse = await checkUserCommunityMembership(communityId);
      return membershipResponse.status || 'none';
    } catch (error) {
      logger.warn(`Error checking membership for community ${communityId}:`, error);
      return 'none'; // Default to 'none' on error
    }
  }
  
  // Handle page change
  function handlePageChange(event: CustomEvent) {
    currentPage = event.detail.page;
    fetchCommunities();
  }
  
  // Handle limit change
  function handleLimitChange(event: Event) {
    const target = event.target as HTMLSelectElement;
    limit = parseInt(target.value);
    currentPage = 1; // Reset to first page when changing limit
    fetchCommunities();
  }
  
  // Handle tab change
  function setActiveTab(tabName: string) {
    activeTab = tabName;
    currentPage = 1; // Reset to first page when changing tabs
    fetchCommunities();
  }
  
  // Handle search
  function handleSearch() {
    currentPage = 1; // Reset to first page when searching
    fetchCommunities();
  }
  
  // Handle category filter change
  function handleCategoryChange(event: CustomEvent) {
    selectedCategories = event.detail.categories;
    currentPage = 1; // Reset to first page when changing filters
    fetchCommunities();
  }
  
  // Clear filters
  function clearFilters() {
    searchQuery = '';
    selectedCategories = [];
    currentPage = 1;
    fetchCommunities();
  }
  
  // Handle join request
  async function joinCommunity(communityId: string, event?: Event) {
    // Prevent propagation to avoid navigation
    if (event) {
      event.stopPropagation();
      event.preventDefault();
    }
    
    if (!authState.is_authenticated) {
      toastStore.showToast('You must be logged in to join communities', 'warning');
      return;
    }
    
    try {
      const response = await requestToJoin(communityId, {});
      if (response.success) {
        toastStore.showToast('Join request sent successfully', 'success');
        // Update the local membership status immediately
        communityMembershipStatus.set(communityId, 'pending');
        communityMembershipStatus = communityMembershipStatus; // Force update
      } else {
        toastStore.showToast(response.message || 'Failed to request to join community', 'error');
      }
    } catch (error) {
      logger.error('Error requesting to join community:', error);
      toastStore.showToast('Failed to request to join community', 'error');
    }
  }
  
  // Handle community click - navigate to detail page
  function handleCommunityClick(communityId: string) {
    if (!communityId) {
      logger.error('Invalid community ID');
      return;
    }
    
    const href = `/communities/${communityId}`;
    window.location.href = href;
  }
  
  // Open community creation modal
  function openCreateModal() {
    if (!authState.is_authenticated) {
      toastStore.showToast('You must be logged in to create a community', 'warning');
      return;
    }
    isCreateModalOpen = true;
  }
  
  // Handle successful community creation
  function handleCommunityCreated() {
    // Refresh communities list
    toastStore.showToast('Community created successfully!', 'success');
    setActiveTab('joined');
  }
  
  // Helper function to get the Supabase URL for community logos
  function getLogoUrl(community: any): string|null {
    if (!community.logo_url) return null;
    
    // Check if the URL is already a complete URL
    if (community.logo_url.startsWith('http')) {
      return community.logo_url;
    }
    
    // If it's just a path, construct the Supabase URL
    if (community.logo_url.startsWith('/')) {
      return getPublicUrl(SUPABASE_BUCKETS.MEDIA, `communities${community.logo_url}`);
    }
    
    return community.logo_url;
  }
  
  // Initial data loading
  onMount(() => {
    authState = getAuthState();
    if (authState.is_authenticated) {
      fetchCategories();
      fetchCommunities();
    } else {
      logger.warn('User not authenticated');
      toastStore.showToast('Sign in to view your communities', 'info');
      // Still fetch discover communities for non-authenticated users
      setActiveTab('discover');
    }
  });
</script>

<MainLayout>
  <div class="communities-container {isDarkMode ? 'dark' : ''}">
    <div class="communities-header">
      <h1>Communities</h1>
      <button class="create-button" on:click={openCreateModal}>
        <PlusIcon size="18" />
        <span>Create Community</span>
      </button>
    </div>
      
    <div class="search-filter-container">
      <div class="search-box">
        <input 
          type="text" 
          placeholder="Search communities..."
          bind:value={searchQuery} 
          on:keydown={(e) => e.key === 'Enter' && handleSearch()}
        />
        <button on:click={handleSearch} aria-label="Search">
          <SearchIcon size="16" />
        </button>
      </div>
      
      {#if availableCategories.length > 0}
        <CategoryFilter 
          categories={availableCategories}
          selected={selectedCategories}
          on:change={handleCategoryChange}
        />
      {/if}
    </div>
    
    <div class="tab-container">
      <button 
        class="tab-button {activeTab === 'joined' ? 'active' : ''}"
        on:click={() => setActiveTab('joined')}
      >
        <CheckIcon size="16" />
        <span>Joined</span>
      </button>
      <button 
        class="tab-button {activeTab === 'pending' ? 'active' : ''}"
        on:click={() => setActiveTab('pending')}
      >
        <AlertCircleIcon size="16" />
        <span>Pending</span>
      </button>
      <button 
        class="tab-button {activeTab === 'discover' ? 'active' : ''}"
        on:click={() => setActiveTab('discover')}
      >
        <FilterIcon size="16" />
        <span>Discover</span>
      </button>
    </div>
    
    {#if error}
      <div class="error-message">
        <AlertCircleIcon size="18" />
        <span>{error}</span>
      </div>
    {/if}
    
    {#if isLoading}
      <div class="loading-container">
        <Spinner size="large" />
      </div>
    {:else}
      <div class="communities-list">
        {#if communities.length > 0}
          {#each communities as community (community.id)}
            <div class="community-card">
              <div class="community-card-content" on:click={() => handleCommunityClick(community.id)}>
                <div class="community-logo">
                  {#if community.logo_url}
                    <img src={getLogoUrl(community)} alt={community.name || 'Community'} />
                  {:else}
                    <div class="logo-placeholder">
                      {community.name && community.name.length > 0 ? community.name[0].toUpperCase() : 'C'}
                    </div>
                  {/if}
                </div>
                <div class="community-info">
                  <h3>{community.name || 'Unnamed Community'}</h3>
                  <p>{community.description || 'No description available'}</p>
                  
                  <div class="community-meta">
                    {#if community.categories && community.categories.length > 0}
                      <div class="community-categories">
                        {#each community.categories.slice(0, 3) as category}
                          <span class="category-tag">{category}</span>
                        {/each}
                        {#if community.categories.length > 3}
                          <span class="category-tag more">+{community.categories.length - 3}</span>
                        {/if}
                      </div>
                    {/if}
                    
                    <div class="community-stats">
                      <UsersIcon size="14" />
                      <span>{community.member_count || 0}</span>
                    </div>
                  </div>
                </div>
              </div>
              
              <!-- Show action buttons based on tab -->
              {#if activeTab === 'discover'}
                <div class="community-action">
                  {#if communityMembershipStatus.get(community.id) === 'member'}
                    <div class="membership-status joined">
                      <CheckIcon size="14" />
                      <span>Joined</span>
                    </div>
                  {:else if communityMembershipStatus.get(community.id) === 'pending'}
                    <div class="membership-status pending">
                      <AlertCircleIcon size="14" />
                      <span>Pending</span>
                    </div>
                  {:else}
                    <button 
                      class="join-button"
                      on:click={(e) => joinCommunity(community.id, e)}
                    >
                      {community.is_private ? 'Request to Join' : 'Join'}
                    </button>
                  {/if}
                </div>
              {:else if activeTab === 'pending'}
                <div class="community-action">
                  <div class="membership-status pending">
                    <AlertCircleIcon size="14" />
                    <span>Join Request Pending</span>
                  </div>
                </div>
              {:else if activeTab === 'joined'}
                <div class="community-action">
                  <button 
                    class="view-button"
                    on:click={() => handleCommunityClick(community.id)}
                  >
                    View
                  </button>
                </div>
              {/if}
            </div>
          {/each}
        {:else}
          <div class="empty-state">
            {#if activeTab === 'joined'}
              <p class="message">You haven't joined any communities yet</p>
              <p class="description">Discover and join communities that interest you</p>
              <button class="action-button" on:click={() => setActiveTab('discover')}>Explore Communities</button>
            {:else if activeTab === 'pending'}
              <p class="message">You don't have any pending join requests</p>
              <p class="description">Find communities to join</p>
              <button class="action-button" on:click={() => setActiveTab('discover')}>Discover Communities</button>
            {:else}
              <p class="message">No communities found matching your search</p>
              {#if searchQuery || selectedCategories.length > 0}
                <button class="action-button" on:click={clearFilters}>Clear Filters</button>
              {:else}
                <p class="description">Be the first to create a community</p>
                <button class="action-button" on:click={openCreateModal}>Create Community</button>
              {/if}
            {/if}
          </div>
        {/if}
      </div>
      
      <!-- Pagination controls -->
      {#if totalPages > 1}
        <div class="pagination-container">
          <div class="limit-selector">
            <label for="limit-select">Show:</label>
            <select id="limit-select" bind:value={limit} on:change={handleLimitChange}>
              {#each limitOptions as option}
                <option value={option}>{option}</option>
              {/each}
            </select>
          </div>
          
          <Pagination 
            totalPages={totalPages}
            currentPage={currentPage}
            maxDisplayPages={5}
            on:pageChange={handlePageChange}
          />
        </div>
      {/if}
    {/if}
  </div>
  
  <CreateCommunityModal 
    bind:isOpen={isCreateModalOpen}
    on:success={handleCommunityCreated}
  />
</MainLayout>

<style>
  .communities-container {
    padding: 1rem;
    max-width: 1200px;
    margin: 0 auto;
  }
  
  .communities-container.dark {
    color: #e2e8f0;
  }
  
  .communities-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
  }
  
  h1 {
    font-size: 1.5rem;
    font-weight: 700;
  }
  
  .create-button {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 1rem;
    background-color: var(--primary-color, #3182ce);
    color: white;
    border-radius: 9999px;
    text-decoration: none;
    transition: background-color 0.2s;
    border: none;
    cursor: pointer;
    font-weight: 500;
  }
  
  .create-button:hover {
    background-color: var(--primary-dark, #2c5282);
  }
  
  .search-filter-container {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 1rem;
    flex-wrap: wrap;
  }
  
  .search-box {
    flex: 1;
    position: relative;
    min-width: 250px;
  }
  
  .search-box input {
    width: 100%;
    padding: 0.5rem 2.5rem 0.5rem 1rem;
    border: 1px solid #e2e8f0;
    border-radius: 0.25rem;
    font-size: 0.875rem;
    outline: none;
    transition: border-color 0.2s;
  }
  
  .dark .search-box input {
    background-color: #2d3748;
    color: #e2e8f0;
    border-color: #4a5568;
  }
  
  .search-box input:focus {
    border-color: var(--primary-color, #3182ce);
  }
  
  .search-box button {
    position: absolute;
    right: 0.5rem;
    top: 50%;
    transform: translateY(-50%);
    background: none;
    border: none;
    color: #718096;
    cursor: pointer;
    padding: 0.25rem;
  }
  
  .dark .search-box button {
    color: #a0aec0;
  }
  
  .search-box button:hover {
    color: var(--primary-color, #3182ce);
  }
  
  .tab-container {
    display: flex;
    gap: 0.25rem;
    margin-bottom: 1rem;
    border-bottom: 1px solid #e2e8f0;
    overflow-x: auto;
    -ms-overflow-style: none; /* IE and Edge */
    scrollbar-width: none; /* Firefox */
  }
  
  .tab-container::-webkit-scrollbar {
    display: none; /* Chrome, Safari, Opera */
  }
  
  .dark .tab-container {
    border-color: #4a5568;
  }
  
  .tab-button {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    color: #718096;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    white-space: nowrap;
  }
  
  .dark .tab-button {
    color: #a0aec0;
  }
  
  .tab-button:hover {
    color: var(--primary-color, #3182ce);
  }
  
  .tab-button.active {
    color: var(--primary-color, #3182ce);
    border-bottom-color: var(--primary-color, #3182ce);
  }
  
  .error-message {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    background-color: #FEE2E2;
    color: #DC2626;
    border-radius: 0.25rem;
    margin-bottom: 1rem;
  }
  
  .dark .error-message {
    background-color: rgba(220, 38, 38, 0.2);
  }
  
  .loading-container {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 200px;
  }
  
  .communities-list {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 1rem;
    margin-bottom: 2rem;
  }
  
  @media (max-width: 640px) {
    .communities-list {
      grid-template-columns: 1fr;
    }
  }
  
  .community-card {
    display: flex;
    flex-direction: column;
    border: 1px solid #e2e8f0;
    border-radius: 0.5rem;
    overflow: hidden;
    transition: transform 0.2s, box-shadow 0.2s;
    background-color: white;
  }
  
  .dark .community-card {
    background-color: #2d3748;
    border-color: #4a5568;
  }
  
  .community-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
  }
  
  .dark .community-card:hover {
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.3), 0 2px 4px -1px rgba(0, 0, 0, 0.2);
  }
  
  .community-card-content {
    display: flex;
    padding: 1rem;
    flex: 1;
    cursor: pointer;
  }
  
  .community-logo {
    width: 60px;
    height: 60px;
    border-radius: 0.25rem;
    overflow: hidden;
    margin-right: 0.75rem;
    flex-shrink: 0;
    background-color: #f7fafc;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  .dark .community-logo {
    background-color: #4a5568;
  }
  
  .community-logo img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  
  .logo-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: var(--primary-color, #3182ce);
    color: white;
    font-size: 1.5rem;
    font-weight: bold;
  }
  
  .community-info {
    flex: 1;
    min-width: 0;
  }
  
  .community-info h3 {
    font-size: 1rem;
    font-weight: 600;
    margin: 0 0 0.25rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
    
  .community-info p {
    font-size: 0.875rem;
    color: #718096;
    margin: 0 0 0.5rem;
    overflow: hidden;
    text-overflow: ellipsis;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    /* Fallback for browsers that don't support line-clamp */
    line-height: 1.4;
    max-height: calc(1.4em * 2);
  }
  
  .dark .community-info p {
    color: #a0aec0;
  }
  
  .community-meta {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 0.75rem;
  }
  
  .community-categories {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem;
  }
  
  .category-tag {
    padding: 0.125rem 0.5rem;
    background-color: #edf2f7;
    color: #4a5568;
    border-radius: 9999px;
    font-size: 0.75rem;
    white-space: nowrap;
  }
  
  .category-tag.more {
    background-color: #e2e8f0;
    color: #4a5568;
  }
  
  .dark .category-tag {
    background-color: #4a5568;
    color: #e2e8f0;
  }
  
  .dark .category-tag.more {
    background-color: #2d3748;
  }
  
  .community-stats {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    color: #718096;
  }
  
  .dark .community-stats {
    color: #a0aec0;
  }
  
  .community-action {
    padding: 0.5rem 1rem;
    border-top: 1px solid #e2e8f0;
    display: flex;
    justify-content: center;
  }
  
  .dark .community-action {
    border-color: #4a5568;
  }
  
  .join-button, .view-button, .action-button {
    padding: 0.375rem 0.75rem;
    background-color: var(--primary-color, #3182ce);
    color: white;
    border: none;
    border-radius: 0.25rem;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.2s;
  }
  
  .view-button {
    background-color: #4a5568;
  }
  
  .join-button:hover, .action-button:hover {
    background-color: var(--primary-dark, #2c5282);
  }
  
  .view-button:hover {
    background-color: #2d3748;
  }
  
  .membership-status {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.375rem 0.75rem;
    border-radius: 0.25rem;
    font-size: 0.875rem;
    font-weight: 500;
  }
  
  .membership-status.joined {
    background-color: #48bb78;
    color: white;
  }
  
  .membership-status.pending {
    background-color: #ed8936;
    color: white;
  }
  
  .empty-state {
    grid-column: 1 / -1;
    padding: 3rem 1rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    color: #718096;
  }
  
  .dark .empty-state {
    color: #a0aec0;
  }
  
  .empty-state .message {
    font-size: 1.125rem;
    font-weight: 500;
    margin-bottom: 0.5rem;
  }
  
  .empty-state .description {
    margin-bottom: 1.5rem;
  }
  
  .pagination-container {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 1rem;
    padding-top: 1rem;
    border-top: 1px solid #e2e8f0;
  }
  
  .dark .pagination-container {
    border-color: #4a5568;
  }
  
  .limit-selector {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.875rem;
  }
  
  .limit-selector select {
    padding: 0.25rem 0.5rem;
    border: 1px solid #e2e8f0;
    border-radius: 0.25rem;
    background-color: white;
  }
  
  .dark .limit-selector select {
    background-color: #2d3748;
    color: #e2e8f0;
    border-color: #4a5568;
  }
  
  @media (max-width: 640px) {
    .communities-header {
      flex-direction: column;
      align-items: flex-start;
      gap: 1rem;
      margin-bottom: 1rem;
    }
    
    .create-button {
      width: 100%;
      justify-content: center;
    }
    
    .search-filter-container {
      flex-direction: column;
      width: 100%;
    }
    
    .search-box {
      width: 100%;
    }
    
    .pagination-container {
      flex-direction: column;
      gap: 1rem;
    }
  }
</style>
