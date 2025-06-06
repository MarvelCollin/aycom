<script lang="ts">
  import { onMount } from 'svelte';
  import MainLayout from '../components/layout/MainLayout.svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { toastStore } from '../stores/toastStore';
  import { getCommunities, getCategories, requestToJoin } from '../api/community';
  
  // Import components
  import Pagination from '../components/common/Pagination.svelte';
  import CategoryFilter from '../components/common/CategoryFilter.svelte';
  import Spinner from '../components/common/Spinner.svelte';
  
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
  
  function getAuthFromStorage() {
    try {
      const authData = localStorage.getItem('auth');
      if (authData) {
        const auth = JSON.parse(authData);
        return {
          userId: auth.user_id,
          isAuthenticated: auth.is_authenticated && auth.access_token && 
            (!auth.expires_at || Date.now() < auth.expires_at),
          accessToken: auth.access_token,
          refreshToken: auth.refresh_token
        };
      }
    } catch (err) {
      logger.error("Error getting auth from storage:", err);
    }
    return { userId: null, isAuthenticated: false, accessToken: null, refreshToken: null };
  }
  
  $: authState = getAuthFromStorage();
  $: isDarkMode = $theme === 'dark';
  
  // Pagination settings
  let limitOptions = [25, 30, 35];
  let currentPage = 1;
  let limit = limitOptions[0];
  let totalItems = 0;
  let totalPages = 0;
  
  // Filter settings
  let activeTab = 'joined'; // 'joined', 'pending', 'discover'
  let searchQuery = '';
  let selectedCategories: string[] = [];
  let availableCategories: string[] = [];
  
  // Community data
  let isLoading = true;
  let joinedCommunities: any[] = [];
  let pendingCommunities: any[] = [];
  let availableCommunities: any[] = [];
  
  // Fetch communities based on active tab and filters
  async function fetchCommunities() {
    try {
      isLoading = true;
      
      const filter = activeTab === 'joined' ? 'joined' : 
                    activeTab === 'pending' ? 'pending' : 'all';
      
      const response = await getCommunities({
        page: currentPage,
        limit: limit,
        filter: filter,
        q: searchQuery,
        category: selectedCategories
      });
      
      if (response.success) {
        if (activeTab === 'joined') {
          joinedCommunities = response.communities;
        } else if (activeTab === 'pending') {
          pendingCommunities = response.communities;
        } else {
          availableCommunities = response.communities;
        }
        
        totalItems = response.pagination.total_count;
        totalPages = response.pagination.total_pages;
        
        // Update limit options if the API returns them
        if (response.limit_options && response.limit_options.length > 0) {
          limitOptions = response.limit_options;
        }
      } else {
        toastStore.showToast('Failed to fetch communities', 'error');
      }
    } catch (error) {
      logger.error('Error fetching communities:', error);
      toastStore.showToast('Failed to fetch communities', 'error');
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
      } else if (response && response.categories) {
        availableCategories = response.categories.map(cat => cat.name);
      }
    } catch (error) {
      logger.error('Error fetching categories:', error);
    }
  }
  
  // Handle page change
  function handlePageChange(event) {
    currentPage = event.detail.page;
    fetchCommunities();
  }
  
  // Handle limit change
  function handleLimitChange(event) {
    limit = parseInt(event.target.value);
    currentPage = 1; // Reset to first page when changing limit
    fetchCommunities();
  }
  
  // Handle tab change
  function setActiveTab(tabName) {
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
  function handleCategoryChange(event) {
    selectedCategories = event.detail.categories;
    currentPage = 1; // Reset to first page when changing filters
    fetchCommunities();
  }
  
  // Clear filters
  function clearFilters() {
    searchQuery = '';
    selectedCategories = [];
    fetchCommunities();
  }
  
  // Handle join request
  async function joinCommunity(communityId, event) {
    // Prevent propagation to avoid navigation
    if (event) {
      event.stopPropagation();
    }
    
    try {
      const response = await requestToJoin(communityId, {});
      if (response.success) {
        toastStore.showToast('Join request sent successfully', 'success');
        // Refresh communities to reflect the change
        fetchCommunities();
      } else {
        toastStore.showToast(response.message || 'Failed to request to join community', 'error');
      }
    } catch (error) {
      logger.error('Error requesting to join community:', error);
      toastStore.showToast('Failed to request to join community', 'error');
    }
  }
  
  // Handle community click
  function handleCommunityClick(communityId) {
    window.location.href = `/communities/${communityId}`;
  }
  
  // Initial data loading
  onMount(() => {
    try {
      const authData = localStorage.getItem('auth');
      if (authData) {
        const auth = JSON.parse(authData);
        if (auth.access_token && (!auth.expires_at || Date.now() < auth.expires_at)) {
          fetchCategories();
          fetchCommunities();
          return;
        }
      }
      // Only log the authentication issue, don't redirect
      logger.warn('User not authenticated or token expired');
    } catch (error) {
      logger.error('Error checking authentication:', error);
    }
  });
</script>

<MainLayout>
  <div class="communities-container {isDarkMode ? 'dark' : ''}">
    <div class="communities-header">
      <h1>Communities</h1>
      <a href="/communities/create" class="create-button">
        <PlusIcon size="18" />
        <span>Create Community</span>
      </a>
    </div>
      
      <div class="search-filter-container">
      <div class="search-box">
          <input 
            type="text" 
          placeholder="Search communities..."
            bind:value={searchQuery} 
          on:keydown={(e) => e.key === 'Enter' && handleSearch()}
        />
        <button on:click={handleSearch}>
          <SearchIcon size="16" />
        </button>
      </div>
      
      <CategoryFilter 
        categories={availableCategories}
        selected={selectedCategories}
        on:change={handleCategoryChange}
      />
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
    
    {#if isLoading}
      <div class="loading-skeleton">
        <Spinner size="large" />
      </div>
    {:else}
      <!-- Display appropriate community list based on active tab -->
      {#if activeTab === 'joined'}
        <div class="communities-list">
          {#if joinedCommunities.length > 0}
            {#each joinedCommunities as community (community.id)}
              <div class="community-card" on:click={() => handleCommunityClick(community.id)}>
                <div class="community-logo">
                  {#if community.logo}
                    <img src={community.logo} alt={community.name || 'Community'} />
                  {:else}
                    <div class="logo-placeholder">
                      {community.name && community.name.length > 0 ? community.name[0].toUpperCase() : 'C'}
                    </div>
                      {/if}
                    </div>
                <div class="community-info">
                  <h3>{community.name || 'Unnamed Community'}</h3>
                  <p>{community.description || 'No description available'}</p>
                  <div class="community-categories">
                    {#each community.categories || [] as category}
                      <span class="category-tag">{category}</span>
                    {/each}
                  </div>
                </div>
              </div>
            {/each}
        {:else}
          <div class="empty-state">
              <p>You haven't joined any communities yet.</p>
              <button on:click={() => setActiveTab('discover')}>Discover Communities</button>
          </div>
        {/if}
        </div>
      {:else if activeTab === 'pending'}
        <div class="communities-list">
          {#if pendingCommunities.length > 0}
            {#each pendingCommunities as community (community.id)}
              <div class="community-card" on:click={() => handleCommunityClick(community.id)}>
                <div class="community-logo">
                  {#if community.logo}
                    <img src={community.logo} alt={community.name || 'Community'} />
                  {:else}
                    <div class="logo-placeholder">
                      {community.name && community.name.length > 0 ? community.name[0].toUpperCase() : 'C'}
                    </div>
                      {/if}
                    </div>
                <div class="community-info">
                  <h3>{community.name || 'Unnamed Community'}</h3>
                  <p>{community.description || 'No description available'}</p>
                  <div class="community-categories">
                    {#each community.categories || [] as category}
                      <span class="category-tag">{category}</span>
                    {/each}
                  </div>
                  <div class="community-status">
                    <AlertCircleIcon size="16" />
                    <span>Join Request Pending</span>
                  </div>
                </div>
              </div>
            {/each}
        {:else}
          <div class="empty-state">
              <p>You don't have any pending join requests.</p>
          </div>
        {/if}
                  </div>
      {:else if activeTab === 'discover'}
        <div class="communities-list">
          {#if availableCommunities.length > 0}
            {#each availableCommunities as community (community.id)}
              <div class="community-card">
                <div 
                  class="community-card-content" 
                  on:click={() => handleCommunityClick(community.id)}
                >
                  <div class="community-logo">
                    {#if community.logo}
                      <img src={community.logo} alt={community.name || 'Community'} />
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
                      <div class="community-categories">
                        {#each community.categories || [] as category}
                          <span class="category-tag">{category}</span>
                        {/each}
                      </div>
                    <div class="community-stats">
                        <UsersIcon size="14" />
                        <span>{community.memberCount || 0}</span>
                      </div>
                    </div>
                  </div>
                </div>
                  <div class="community-action">
                    <button 
                    class="join-button"
                    on:click={(e) => joinCommunity(community.id, e)}
                    >
                    {community.isPrivate ? 'Request to Join' : 'Join'}
                    </button>
                </div>
              </div>
            {/each}
        {:else}
          <div class="empty-state">
              <p>No communities found matching your search.</p>
              <button on:click={clearFilters}>Clear Filters</button>
            </div>
          {/if}
        </div>
      {/if}
      
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
  }
  
  .create-button:hover {
    background-color: var(--primary-dark, #2c5282);
  }
  
  .search-filter-container {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 1rem;
  }
  
  .search-box {
    flex: 1;
    position: relative;
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
  
  .loading-skeleton {
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
  
  .community-card {
    display: flex;
    flex-direction: column;
    border: 1px solid #e2e8f0;
    border-radius: 0.5rem;
    overflow: hidden;
    transition: transform 0.2s, box-shadow 0.2s;
    background-color: white;
    cursor: pointer;
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
    .community-info p {    font-size: 0.875rem;
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
    margin-top: 0.5rem;
  }
  
  .category-tag {
    padding: 0.125rem 0.5rem;
    background-color: #edf2f7;
    color: #4a5568;
    border-radius: 9999px;
    font-size: 0.75rem;
    white-space: nowrap;
  }
  
  .dark .category-tag {
    background-color: #4a5568;
    color: #e2e8f0;
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
  
  .community-status {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    color: #dd6b20;
    font-size: 0.75rem;
    margin-top: 0.5rem;
  }
  
  .community-action {
    padding: 0.75rem 1rem;
    border-top: 1px solid #e2e8f0;
    display: flex;
    justify-content: flex-end;
  }
  
  .dark .community-action {
    border-color: #4a5568;
  }
  
  .join-button {
    padding: 0.375rem 0.75rem;
    background-color: var(--primary-color, #3182ce);
    color: white;
    border: none;
    border-radius: 0.25rem;
    font-size: 0.875rem;
    cursor: pointer;
    transition: background-color 0.2s;
  }
  
  .join-button:hover {
    background-color: var(--primary-dark, #2c5282);
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
  
  .empty-state button {
    margin-top: 1rem;
    padding: 0.5rem 1rem;
    background-color: var(--primary-color, #3182ce);
    color: white;
    border: none;
    border-radius: 0.25rem;
    cursor: pointer;
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
</style>
