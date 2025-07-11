<script lang="ts">
  import { onMount } from "svelte";
  import MainLayout from "../components/layout/MainLayout.svelte";
  import { useAuth } from "../hooks/useAuth";
  import { useTheme } from "../hooks/useTheme";
  import { createLoggerWithPrefix } from "../utils/logger";
  import { toastStore } from "../stores/toastStore";
  import {
    getJoinedCommunities,
    getPendingCommunities,
    getDiscoverCommunities,
    getCategories,
    requestToJoin,
    checkUserCommunityMembership,
    searchCommunities,
    getCommunities
  } from "../api/community";
  import type { ICategoriesResponse, ICategory } from "../interfaces/ICategory";
  import { getPublicUrl, SUPABASE_BUCKETS } from "../utils/supabase";
  import { formatStorageUrl } from "../utils/common";

  import Pagination from "../components/common/Pagination.svelte";
  import CategoryFilter from "../components/common/CategoryFilter.svelte";
  import Spinner from "../components/common/Spinner.svelte";
  import CreateCommunityModal from "../components/communities/CreateCommunityModal.svelte";

  import SearchIcon from "svelte-feather-icons/src/icons/SearchIcon.svelte";
  import FilterIcon from "svelte-feather-icons/src/icons/FilterIcon.svelte";
  import CheckIcon from "svelte-feather-icons/src/icons/CheckIcon.svelte";
  import UsersIcon from "svelte-feather-icons/src/icons/UsersIcon.svelte";
  import LockIcon from "svelte-feather-icons/src/icons/LockIcon.svelte";
  import PlusIcon from "svelte-feather-icons/src/icons/PlusIcon.svelte";
  import AlertCircleIcon from "svelte-feather-icons/src/icons/AlertCircleIcon.svelte";

  const logger = createLoggerWithPrefix("Communities");

  const { getAuthState } = useAuth();
  const { theme } = useTheme();
  let authState = getAuthState();
  $: isDarkMode = $theme === "dark";

  let isCreateModalOpen = false;

  const limitOptions = [25, 30, 35];
  let currentPage = 1;
  let limit = limitOptions[0];
  let totalCount = 0;  
  let totalPages = 1;

  let activeTab = "joined"; 
  let searchQuery = "";
  let selectedCategories: string[] = [];
  let availableCategories: string[] = [];

  let showFilters = false;
  let searchTimeout: ReturnType<typeof setTimeout>;

  let isLoading = false;
  let communities: any[] = []; 
  let error: string | null = null;

  let communityMembershipStatus = new Map();

  function calculateTotalPages(total: number, perPage: number): number {

    if (total <= 0 || perPage <= 0) return 1;
    return Math.ceil(total / perPage);
  }

  async function fetchCommunities() {
    isLoading = true;
    error = null;

    try {

      const params: any = {
        page: currentPage,
        limit: limit
      };

      if (searchQuery && searchQuery.trim()) {
        params.q = searchQuery.trim();
      }

      const validCategories = selectedCategories.filter(cat => cat && cat.trim());
      if (validCategories.length > 0) {
        params.category = validCategories;
      }

      let result;
      logger.info(`Fetching communities for tab: ${activeTab}`);

      try {

        switch (activeTab) {
          case "joined":
            result = await getJoinedCommunities(authState.user_id || "", params);
            break;
          case "pending":
            result = await getPendingCommunities(authState.user_id || "", params);
            break;
          case "discover":
            result = await getDiscoverCommunities(authState.user_id || "", params);
            break;
          case "all":

            if (searchQuery?.trim() || validCategories.length > 0) {
              const searchParams = {
                ...params,
                categories: validCategories
              };
              result = await searchCommunities(searchQuery?.trim() || "", params.page || 1, params.limit || 25, searchParams);
            } else {

              result = await getCommunities(params);
            }
            break;
          default:
            result = await getJoinedCommunities(authState.user_id || "", params);
        }
      } catch (apiError) {

        logger.error(`API error in ${activeTab} tab:`, apiError);
        result = {
          success: true,
          communities: [],
          total: 0,
          page: currentPage,
          limit
        };
      }

      console.log(`[Communities] Raw API response for ${activeTab} tab:`, result);
      logger.info("[Communities] API response:", result);

      if (result && result.success !== false) {

        if (result.data && result.data.communities) {

          communities = result.data.communities || [];
          totalCount = result.data.pagination?.total_count || 0;
          currentPage = result.data.pagination?.current_page || currentPage;
          limit = result.data.pagination?.per_page || limit;
        } else {

          communities = result.communities || [];
          totalCount = result.total || 0;
          currentPage = result.page || currentPage;
          limit = result.limit || limit;
        }

        console.log("[Communities] Extracted communities:", communities);
        console.log("[Communities] Total count:", totalCount);

        if (!Array.isArray(communities)) {
          console.log("[Communities] Communities is not an array, resetting to empty array");
          communities = [];
        }

        communities.forEach((community, index) => {
          console.log(`[Communities] Community ${index}:`, {
            id: community.id,
            name: community.name,
            description: community.description?.substring(0, 30) + "...",
            logo_url: community.logo_url || community.logoUrl,
            banner_url: community.banner_url || community.bannerUrl
          });
        });

        totalPages = calculateTotalPages(totalCount, limit);
        if (currentPage > totalPages && totalPages > 0) {
          currentPage = totalPages;
        }

        if (activeTab === "discover" || activeTab === "all") {
          await checkMembershipStatusForAll();
        }

        error = null;
      } else {

        logger.error("[Communities] Error fetching communities:", result?.error || "Unknown error");
        communities = [];
        totalCount = 0;
        totalPages = 1;
        error = result?.error?.message || "Failed to load communities";
      }
    } catch (err) {

      logger.error("[Communities] Exception fetching communities:", err);
      communities = [];
      totalCount = 0;
      totalPages = 1;
      error = "An unexpected error occurred";
    } finally {
      isLoading = false;
    }
  }

  async function checkMembershipStatusForAll() {
    if (!Array.isArray(communities) || communities.length === 0 || (activeTab !== "discover" && activeTab !== "all")) return;

    for (const community of communities) {
      if (community && community.id) {
        try {
          const status = await checkMembershipStatus(community.id);
          communityMembershipStatus.set(community.id, status);

          communityMembershipStatus = communityMembershipStatus;
        } catch (err) {
          logger.error(`Error checking membership for community ${community.id}:`, err);
        }
      }
    }
  }

  async function checkMembershipStatus(communityId: string): Promise<string> {
    if (!authState.is_authenticated || !communityId) return "none";

    try {
      const membershipResponse = await checkUserCommunityMembership(communityId);
      return membershipResponse.status || "none";
    } catch (error) {
      logger.warn(`Error checking membership for community ${communityId}:`, error);
      return "none"; 
    }
  }

  function handlePageChange(event: CustomEvent) {
    currentPage = event.detail.page;
    fetchCommunities();
  }

  function handleLimitChange(event: Event) {
    const target = event.target as HTMLSelectElement;
    limit = parseInt(target.value);
    currentPage = 1; 
    fetchCommunities();
  }

  function setActiveTab(tabName: string) {
    activeTab = tabName;
    currentPage = 1; 
    fetchCommunities();
  }

  function handleSearch() {
    currentPage = 1; 
    fetchCommunities();
  }

  function handleSearchInput() {

    if (activeTab === "all") {
      clearTimeout(searchTimeout);
      searchTimeout = setTimeout(() => {
        currentPage = 1;
        fetchCommunities();
      }, 300); 
    }
  }

  function handleCategoryChange(event: CustomEvent) {
    selectedCategories = event.detail.selected;
    currentPage = 1; 
    fetchCommunities();
  }

  function toggleFilters() {
    showFilters = !showFilters;
  }

  function clearFilters() {
    searchQuery = "";
    selectedCategories = [];
    currentPage = 1;
    fetchCommunities();
  }

  async function fetchCategories() {
    try {
      const categoriesResponse = await getCategories();
      console.log("[Communities] Categories API response:", categoriesResponse);

      if (categoriesResponse) {
        if (Array.isArray(categoriesResponse)) {

          availableCategories = categoriesResponse.map((cat: ICategory) => cat.name);
        } else if ((categoriesResponse as any).data && Array.isArray((categoriesResponse as any).data.categories)) {

          availableCategories = (categoriesResponse as any).data.categories.map((cat: ICategory) => cat.name);
        } else if (Array.isArray((categoriesResponse as any).categories)) {

          availableCategories = (categoriesResponse as any).categories.map((cat: ICategory) => cat.name);
        } else {
          availableCategories = [];
        }
      } else {
        logger.error("[Communities] Error fetching categories: No response");
        availableCategories = [];
      }

      console.log("[Communities] Available categories:", availableCategories);
    } catch (error) {
      logger.error("[Communities] Exception fetching categories:", error);
      availableCategories = [];
    }
  }

  function searchAllCommunities() {
    setActiveTab("all");
  }

  async function joinCommunity(communityId: string, event?: Event) {

    if (event) {
      event.stopPropagation();
      event.preventDefault();
    }

    if (!authState.is_authenticated) {
      toastStore.showToast("You must be logged in to join communities", "warning");
      return;
    }

    try {
      const response = await requestToJoin(communityId, {});
      if (response.success) {
        toastStore.showToast("Join request sent successfully", "success");

        communityMembershipStatus.set(communityId, "pending");
        communityMembershipStatus = communityMembershipStatus; 
      } else {
        toastStore.showToast(response.message || "Failed to request to join community", "error");
      }
    } catch (error) {
      logger.error("Error requesting to join community:", error);
      toastStore.showToast("Failed to request to join community", "error");
    }
  }

  function handleCommunityClick(communityId: string) {
    if (!communityId) {
      logger.error("Invalid community ID");
      return;
    }

    const href = `/communities/${communityId}`;
    window.location.href = href;
  }

  function openCreateModal() {
    if (!authState.is_authenticated) {
      toastStore.showToast("You must be logged in to create a community", "warning");
      return;
    }
    isCreateModalOpen = true;
  }

  function handleCommunityCreated() {

    toastStore.showToast("Community created successfully!", "success");
    setActiveTab("joined");
  }

  function getLogoUrl(community: any): string|null {
    if (!community) return null;

    const logoUrl = community.logo_url || community.logoUrl || community.logo;

    if (!logoUrl) return null;

    try {

      const fixedUrl = fixKnownProblematicUrl(logoUrl);
      if (fixedUrl !== logoUrl) {
        return fixedUrl;
      }

      return formatStorageUrl(logoUrl);
    } catch (error) {
      console.error("Error formatting logo URL:", error, logoUrl);
      return logoUrl; 
    }
  }

  function getBannerUrl(community: any): string|null {
    if (!community) return null;

    const bannerUrl = community.banner_url || community.bannerUrl || community.banner;

    if (!bannerUrl) return null;

    try {

      const fixedUrl = fixKnownProblematicUrl(bannerUrl);
      if (fixedUrl !== bannerUrl) {
        return fixedUrl;
      }

      return formatStorageUrl(bannerUrl);
    } catch (error) {
      console.error("Error formatting banner URL:", error, bannerUrl);
      return bannerUrl; 
    }
  }

  function fixKnownProblematicUrl(url: string): string {

    const knownUrlFixes: Record<string, string> = {
      "https://sdhtnvlmuywinhcglfsu.supabase.co/storage/v1/s3/tpaweb/1kolknj_1/1749614938807_vf09h7v5.jpg":
        "https://sdhtnvlmuywinhcglfsu.supabase.co/storage/v1/object/public/tpaweb/1kolknj_1/1749614938807_vf09h7v5.jpg",
      "https://sdhtnvlmuywinhcglfsu.supabase.co/storage/v1/s3/tpaweb/1kolknj_1/1749614937805_k3pne3t9.jpg":
        "https://sdhtnvlmuywinhcglfsu.supabase.co/storage/v1/object/public/tpaweb/1kolknj_1/1749614937805_k3pne3t9.jpg",
      "https://sdhtnvlmuywinhcglfsu.supabase.co/storage/v1/s3/uploads/community/community_banner_91df5727a9c5427e94cee0486e3bfdb7_1749202798.png":
        "https://sdhtnvlmuywinhcglfsu.supabase.co/storage/v1/object/public/uploads/community/community_banner_91df5727a9c5427e94cee0486e3bfdb7_1749202798.png",
      "https://sdhtnvlmuywinhcglfsu.supabase.co/storage/v1/s3/test/logo.jpg":
        "https://sdhtnvlmuywinhcglfsu.supabase.co/storage/v1/object/public/test/logo.jpg",
      "https://sdhtnvlmuywinhcglfsu.supabase.co/storage/v1/s3/test/banner.jpg":
        "https://sdhtnvlmuywinhcglfsu.supabase.co/storage/v1/object/public/test/banner.jpg",
      "https://sdhtnvlmuywinhcglfsu.supabase.co/storage/v1/s3/tpaweb/1kolknj_1/1749269410545_m10xabzo.png":
        "https://sdhtnvlmuywinhcglfsu.supabase.co/storage/v1/object/public/tpaweb/1kolknj_1/1749269410545_m10xabzo.png",
      "https://sdhtnvlmuywinhcglfsu.supabase.co/storage/v1/s3/uploads/community/community_logo_91df5727a9c5427e94cee0486e3bfdb7_1749202798.jpg":
        "https://sdhtnvlmuywinhcglfsu.supabase.co/storage/v1/object/public/uploads/community/community_logo_91df5727a9c5427e94cee0486e3bfdb7_1749202798.jpg",
      "https://sdhtnvlmuywinhcglfsu.supabase.co/storage/v1/s3/tpaweb/1kolknj_1/1749269411979_4070qdrp.png":
        "https://sdhtnvlmuywinhcglfsu.supabase.co/storage/v1/object/public/tpaweb/1kolknj_1/1749269411979_4070qdrp.png"
    };

    if (url in knownUrlFixes) {
      console.log("Fixed known problematic URL:", url, "to", knownUrlFixes[url]);
      return knownUrlFixes[url];
    }

    if (url.includes("/storage/v1/s3/")) {

      const fixedUrl = url.replace("/storage/v1/s3/", "/storage/v1/object/public/");
      console.log("Fixed pattern-matched URL:", url, "to", fixedUrl);
      return fixedUrl;
    }

    return url;
  }

  onMount(() => {
    authState = getAuthState();

    if (authState.is_authenticated) {
      activeTab = "joined"; 
      fetchCategories();
      fetchCommunities();
    } else {
      activeTab = "all"; 
      logger.info("User not authenticated, showing all communities");

      fetchCategories();
      fetchCommunities();
    }
  });
</script>

<MainLayout>
  <div class="communities-container {isDarkMode ? "dark" : ""}">
    <div class="communities-header">
      <h1>Communities</h1>
      <button class="create-button" on:click={openCreateModal}>
        <PlusIcon size="18" />
        <span>Create Community</span>
      </button>
    </div>

    <div class="search-filter-container">
      <div class="search-section">
        <div class="search-box">
          <input
            type="text"
            placeholder={activeTab === "all" ? "Search all communities..." : "Search communities..."}
            bind:value={searchQuery}
            on:input={handleSearchInput}
            on:keydown={(e) => e.key === "Enter" && handleSearch()}
          />
          <button on:click={handleSearch} aria-label="Search">
            <SearchIcon size="16" />
          </button>
        </div>

        <div class="search-actions">
          <button 
            class="filter-toggle-button {showFilters ? 'active' : ''}"
            on:click={toggleFilters}
            aria-label="Toggle filters"
          >
            <FilterIcon size="16" />
            <span>Filters</span>
          </button>

          {#if searchQuery || selectedCategories.length > 0}
            <button class="clear-filters-button" on:click={clearFilters}>
              Clear
            </button>
          {/if}
        </div>
      </div>

      {#if showFilters && availableCategories.length > 0}
        <div class="filters-section">
          <CategoryFilter
            categories={availableCategories}
            selected={selectedCategories}
            on:change={handleCategoryChange}
          />
        </div>
      {/if}
    </div>

    <div class="tab-container">
      <button
        class="tab-button {activeTab === "joined" ? "active" : ""}"
        on:click={() => setActiveTab("joined")}
      >
        <CheckIcon size="16" />
        <span>Joined</span>
      </button>
      <button
        class="tab-button {activeTab === "pending" ? "active" : ""}"
        on:click={() => setActiveTab("pending")}
      >
        <AlertCircleIcon size="16" />
        <span>Pending</span>
      </button>
      <button
        class="tab-button {activeTab === "discover" ? "active" : ""}"
        on:click={() => setActiveTab("discover")}
      >
        <FilterIcon size="16" />
        <span>Discover</span>
      </button>
      <button
        class="tab-button {activeTab === "all" ? "active" : ""}"
        on:click={() => setActiveTab("all")}
      >
        <SearchIcon size="16" />
        <span>All Communities</span>
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
              <div class="community-banner">
                {#if community.banner_url || community.bannerUrl || community.banner}
                  <img
                    src={getBannerUrl(community)}
                    alt={`${community.name || "Community"} banner`}
                    crossorigin="anonymous"
                    loading="lazy"
                    on:error={(e) => {
                      console.error("Community banner image failed to load:", getBannerUrl(community));

                      const imgElement = e.target as HTMLImageElement;
                      if (imgElement) {
                        imgElement.classList.add("image-error");

                        imgElement.setAttribute("data-original-url",
                          community.banner_url || community.bannerUrl || community.banner || "");

                        const parent = imgElement.parentElement;
                        if (parent) {
                          parent.classList.add("banner-placeholder");
                        }
                      }
                    }}
                  />
                {:else}
                  <div class="banner-placeholder"></div>
                {/if}
              </div>
              <div 
                class="community-card-content" 
                on:click={() => handleCommunityClick(community.id)}
                on:keydown={(e) => e.key === 'Enter' && handleCommunityClick(community.id)}
                role="button"
                tabindex="0"
              >
                <div class="community-logo">
                  {#if community.logo_url || community.logoUrl || community.logo}
                    <img
                      src={getLogoUrl(community)}
                      alt={community.name || "Community"}
                      crossorigin="anonymous"
                      loading="lazy"
                      on:error={(e) => {
                        console.error("Community logo image failed to load:", getLogoUrl(community));

                        const imgElement = e.target as HTMLImageElement;
                        if (imgElement) {
                          imgElement.classList.add("image-error");

                          imgElement.style.display = "none";
                          if (imgElement.parentElement) {
                            imgElement.parentElement.classList.add("logo-placeholder");

                            const letter = community.name && community.name.length > 0
                              ? community.name[0].toUpperCase()
                              : "C";
                            imgElement.parentElement.setAttribute("data-content", letter);
                          }

                          imgElement.setAttribute("data-original-url",
                            community.logo_url || community.logoUrl || community.logo || "");
                        }
                      }}
                    />
                    <!-- Fallback content that will be shown when image fails to load -->
                    <div class="logo-placeholder-fallback">
                      {community.name && community.name.length > 0 ? community.name[0].toUpperCase() : "C"}
                    </div>
                  {:else}
                    <div class="logo-placeholder">
                      {community.name && community.name.length > 0 ? community.name[0].toUpperCase() : "C"}
                    </div>
                  {/if}
                </div>
                <div class="community-info">
                  <h3 class="community-name">
                    {community.name || "Unnamed Community"}
                    {#if community.is_approved === false || community.isApproved === false}
                      <span class="pending-badge">
                        <AlertCircleIcon size="12" />
                        <span>Pending Admin Approval</span>
                      </span>
                    {/if}
                  </h3>
                  <p class="community-description">{community.description || "No description available"}</p>

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
                    {:else}
                      <div class="community-categories">
                        <span class="category-tag no-category">Uncategorized</span>
                      </div>
                    {/if}

                    <div class="community-stats">
                      <UsersIcon size="14" />
                      <span>{community.member_count || 0}</span>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Show action buttons based on tab - using a unified design -->
              <div class="community-action">
                {#if activeTab === "discover"}
                  {#if communityMembershipStatus.get(community.id) === "member"}
                    <div class="membership-status joined">
                      <CheckIcon size="14" />
                      <span>Joined</span>
                    </div>
                  {:else if communityMembershipStatus.get(community.id) === "pending"}
                    <div class="membership-status pending">
                      <AlertCircleIcon size="14" />
                      <span>Pending</span>
                    </div>
                  {:else}
                    <button
                      class="action-button join-button"
                      on:click={(e) => joinCommunity(community.id, e)}
                    >
                      {community.is_private ? "Request to Join" : "Join"}
                    </button>
                  {/if}
                {:else if activeTab === "pending"}
                  <div class="membership-status pending">
                    <AlertCircleIcon size="14" />
                    <span>Join Request Pending</span>
                  </div>
                {:else if activeTab === "joined"}
                  <button
                    class="action-button view-button"
                    on:click={() => handleCommunityClick(community.id)}
                  >
                    View
                  </button>
                {:else if activeTab === "all"}
                  {#if communityMembershipStatus.get(community.id) === "member"}
                    <div class="membership-status joined">
                      <CheckIcon size="14" />
                      <span>Joined</span>
                    </div>
                  {:else if communityMembershipStatus.get(community.id) === "pending"}
                    <div class="membership-status pending">
                      <AlertCircleIcon size="14" />
                      <span>Pending</span>
                    </div>
                  {:else}
                    <button
                      class="action-button join-button"
                      on:click={(e) => joinCommunity(community.id, e)}
                    >
                      {community.is_private ? "Request to Join" : "Join"}
                    </button>
                  {/if}
                {/if}
              </div>
            </div>
          {/each}
        {:else}
          <div class="empty-state">
            {#if activeTab === "joined"}
              <p class="message">You haven't joined any communities yet</p>
              <p class="description">Discover and join communities that interest you</p>
              <button class="action-button" on:click={() => setActiveTab("discover")}>Explore Communities</button>
            {:else if activeTab === "pending"}
              <p class="message">You don't have any pending join requests</p>
              <p class="description">Find communities to join</p>
              <button class="action-button" on:click={() => setActiveTab("discover")}>Discover Communities</button>
            {:else if activeTab === "all"}
              <p class="message">No communities found</p>
              {#if searchQuery || selectedCategories.length > 0}
                <p class="description">Try adjusting your search or filters</p>
                <button class="action-button" on:click={clearFilters}>Clear Filters</button>
              {:else}
                <p class="description">Be the first to create a community</p>
                <button class="action-button" on:click={openCreateModal}>Create Community</button>
              {/if}
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

      <!-- Pagination controls - always show even with fewer than limit items -->
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
          totalItems={totalCount}
          perPage={limit}
          {currentPage}
          maxDisplayPages={5}
          on:pageChange={handlePageChange}
        />
      </div>
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
    background: white;
    border-radius: 0.75rem;
    box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06);
    margin-bottom: 1.5rem;
    overflow: hidden;
    transition: all 0.2s ease-in-out;
  }

  .dark .search-filter-container {
    background: #1a202c;
    box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.3), 0 1px 2px 0 rgba(0, 0, 0, 0.2);
  }

  .search-section {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1rem;
    flex-wrap: wrap;
  }

  .search-box {
    flex: 1;
    min-width: 280px;
    position: relative;
    display: flex;
    align-items: center;
    background: #f7fafc;
    border: 1px solid #e2e8f0;
    border-radius: 0.5rem;
    overflow: hidden;
    transition: all 0.2s ease-in-out;
  }

  .search-box:focus-within {
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  .dark .search-box {
    background: #2d3748;
    border-color: #4a5568;
  }

  .dark .search-box:focus-within {
    border-color: #60a5fa;
    box-shadow: 0 0 0 3px rgba(96, 165, 250, 0.1);
  }

  .search-box input {
    flex: 1;
    padding: 0.75rem 1rem;
    border: none;
    background: transparent;
    outline: none;
    font-size: 0.875rem;
    color: #2d3748;
  }

  .dark .search-box input {
    color: #e2e8f0;
  }

  .search-box input::placeholder {
    color: #a0aec0;
  }

  .dark .search-box input::placeholder {
    color: #718096;
  }

  .search-box button {
    padding: 0.75rem;
    border: none;
    background: #3b82f6;
    color: white;
    cursor: pointer;
    transition: background-color 0.2s ease-in-out;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .search-box button:hover {
    background: #2563eb;
  }

  .search-actions {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex-shrink: 0;
  }

  .filter-toggle-button {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    border: 1px solid #e2e8f0;
    background: white;
    color: #374151;
    border-radius: 0.5rem;
    cursor: pointer;
    font-size: 0.875rem;
    font-weight: 500;
    transition: all 0.2s ease-in-out;
  }

  .filter-toggle-button:hover {
    background: #f9fafb;
    border-color: #d1d5db;
  }

  .filter-toggle-button.active {
    background: #3b82f6;
    border-color: #3b82f6;
    color: white;
  }

  .dark .filter-toggle-button {
    background: #2d3748;
    border-color: #4a5568;
    color: #e2e8f0;
  }

  .dark .filter-toggle-button:hover {
    background: #4a5568;
  }

  .clear-filters-button {
    padding: 0.75rem 1rem;
    border: 1px solid #ef4444;
    background: white;
    color: #ef4444;
    border-radius: 0.5rem;
    cursor: pointer;
    font-size: 0.875rem;
    font-weight: 500;
    transition: all 0.2s ease-in-out;
  }

  .clear-filters-button:hover {
    background: #ef4444;
    color: white;
  }

  .dark .clear-filters-button {
    background: #1a202c;
  }

  .filters-section {
    border-top: 1px solid #e2e8f0;
    padding: 1rem;
    background: #f8fafc;
    animation: slideDown 0.2s ease-out;
  }

  .dark .filters-section {
    border-color: #4a5568;
    background: #2d3748;
  }

  @keyframes slideDown {
    from {
      opacity: 0;
      transform: translateY(-10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  @media (max-width: 640px) {
    .search-section {
      flex-direction: column;
      align-items: stretch;
      gap: 0.75rem;
    }

    .search-box {
      min-width: 100%;
    }

    .search-actions {
      justify-content: center;
      flex-wrap: wrap;
    }

    .filter-toggle-button,
    .clear-filters-button {
      flex: 1;
      justify-content: center;
      min-width: 120px;
    }
  }

  .tab-container {
    display: flex;
    gap: 0.25rem;
    margin-bottom: 1rem;
    border-bottom: 1px solid #e2e8f0;
    overflow-x: auto;
    -ms-overflow-style: none; 
    scrollbar-width: none; 
  }

  .tab-container::-webkit-scrollbar {
    display: none; 
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
  }

  .community-card {
    display: flex;
    flex-direction: column;
    background-color: white;
    border-radius: 0.5rem;
    border: 1px solid #e2e8f0;
    overflow: hidden;
    transition: transform 0.2s ease, box-shadow 0.2s ease;
    height: 100%;
  }

  .community-banner {
    width: 100%;
    height: 80px;
    overflow: hidden;
    position: relative;
  }

  .community-banner img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .banner-placeholder {
    width: 100%;
    height: 100%;
    min-height: 120px;
    background-color: #e2e8f0;
    background-image: linear-gradient(45deg, #cbd5e1 25%, transparent 25%, transparent 75%, #cbd5e1 75%, #cbd5e1),
                      linear-gradient(45deg, #cbd5e1 25%, transparent 25%, transparent 75%, #cbd5e1 75%, #cbd5e1);
    background-size: 20px 20px;
    background-position: 0 0, 10px 10px;
  }

  .dark .community-banner .banner-placeholder {
    background-color: #334155;
    background-image: linear-gradient(45deg, #475569 25%, transparent 25%, transparent 75%, #475569 75%, #475569),
                      linear-gradient(45deg, #475569 25%, transparent 25%, transparent 75%, #475569 75%, #475569);
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
    display: flex;
    align-items: center;
    justify-content: center;
    width: 60px;
    height: 60px;
    border-radius: 50%;
    background-color: #3b82f6;
    color: white;
    font-weight: bold;
    font-size: 1.5rem;
    position: relative;
  }

  .logo-placeholder::after {
    content: attr(data-content);
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    color: white;
    font-weight: bold;
    font-size: 1.5rem;
  }

  .community-info {
    flex: 1;
    min-width: 0;
  }

  .community-name {
    margin: 0 0 0.25rem;
    font-size: 1.125rem;
    font-weight: 600;
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .pending-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    font-size: 0.7rem;
    font-weight: normal;
    color: #805600;
    background-color: #fff6e0;
    padding: 0.15rem 0.4rem;
    border-radius: 0.25rem;
    white-space: nowrap;
    vertical-align: middle;
  }

  .dark .pending-badge {
    background-color: #3c2f00;
    color: #ffdb7d;
  }

  .community-description {
    font-size: 0.875rem;
    color: #718096;
    margin: 0 0 0.5rem;
    overflow: hidden;
    text-overflow: ellipsis;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;

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

  .category-tag.no-category {
    background-color: #e2e8f0;
    color: #718096;
  }

  .category-tag.more {
    background-color: #e2e8f0;
    color: #4a5568;
  }

  .dark .category-tag {
    background-color: #4a5568;
    color: #e2e8f0;
  }

  .dark .category-tag.no-category {
    background-color: #2d3748;
    color: #a0aec0;
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

  .action-button {
    padding: 0.375rem 0.75rem;
    background-color: var(--primary-color, #3182ce);
    color: white;
    border: none;
    border-radius: 0.25rem;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.2s;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.25rem;
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
    justify-content: center;
    gap: 0.25rem;
    padding: 0.375rem 0.75rem;
    border-radius: 0.25rem;
    font-size: 0.875rem;
    font-weight: 500;
    min-width: 100px;
    text-align: center;
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

  @media (max-width: 768px) {
    .search-section {
      flex-direction: column;
      gap: 0.75rem;
    }

    .search-box {
      min-width: 100%;
    }

    .search-actions {
      justify-content: center;
      width: 100%;
    }

    .tab-container {
      justify-content: center;
      gap: 0.125rem;
    }

    .tab-button {
      flex: 1;
      min-width: 0;
    }

    .tab-button span {
      display: none;
    }

    .filters-section {
      padding: 0.75rem;
    }

    .communities-header {
      flex-direction: column;
      gap: 1rem;
      align-items: stretch;
    }

    .create-button {
      width: 100%;
      justify-content: center;
    }
  }

  @media (max-width: 480px) {
    .tab-button {
      padding: 0.5rem 0.25rem;
      font-size: 0.75rem;
    }

    .search-box input {
      font-size: 16px; 
    }
  }

  .banner-placeholder {
    height: 100%;
    width: 100%;
    min-height: 120px;
    background-color: #e2e8f0;
    background-image: linear-gradient(45deg, #cbd5e1 25%, transparent 25%, transparent 75%, #cbd5e1 75%, #cbd5e1),
                      linear-gradient(45deg, #cbd5e1 25%, transparent 25%, transparent 75%, #cbd5e1 75%, #cbd5e1);
    background-size: 20px 20px;
    background-position: 0 0, 10px 10px;
  }

  .dark .banner-placeholder {
    background-color: #334155;
    background-image: linear-gradient(45deg, #475569 25%, transparent 25%, transparent 75%, #475569 75%, #475569),
                      linear-gradient(45deg, #475569 25%, transparent 25%, transparent 75%, #475569 75%, #475569);
  }

  .logo-placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 60px;
    height: 60px;
    border-radius: 50%;
    background-color: #3b82f6;
    color: white;
    font-weight: bold;
    font-size: 1.5rem;
    position: relative;
  }

  .logo-placeholder::after {
    content: attr(data-content);
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    color: white;
    font-weight: bold;
    font-size: 1.5rem;
  }
</style>