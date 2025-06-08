<script lang="ts">
  import { onMount } from 'svelte';
  import { toastStore } from '../stores/toastStore';
  import MainLayout from '../components/layout/MainLayout.svelte';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { 
    getCommunityById, 
    checkUserCommunityMembership, 
    requestToJoin,
    listMembers,
    listRules
  } from '../api/community';
  import { getUserThreads } from '../api/thread';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import { getPublicUrl, SUPABASE_BUCKETS } from '../utils/supabase';
  import type { IAuthStore } from '../interfaces/IAuth';
    // Import ITweet interface or create Thread interface that extends it
  interface ITweet {
    id: string;
    content?: string;
    timestamp?: Date;
    username?: string;
    display_name?: string;
    avatar?: string;
    likes?: number;
    comments?: number;
    is_liked?: boolean;
    is_reposted?: boolean;
    is_bookmarked?: boolean;
    replies?: number;
    reposts?: number;
    bookmarks?: number;
    views?: number;
  }
  
  // Import icons
  import UsersIcon from 'svelte-feather-icons/src/icons/UsersIcon.svelte';
  import InfoIcon from 'svelte-feather-icons/src/icons/InfoIcon.svelte';
  import BookmarkIcon from 'svelte-feather-icons/src/icons/BookmarkIcon.svelte';
  import AlertCircleIcon from 'svelte-feather-icons/src/icons/AlertCircleIcon.svelte';
  import MessageSquareIcon from 'svelte-feather-icons/src/icons/MessageSquareIcon.svelte';
  import LockIcon from 'svelte-feather-icons/src/icons/LockIcon.svelte';
  import LogOutIcon from 'svelte-feather-icons/src/icons/LogOutIcon.svelte';
  import UserPlusIcon from 'svelte-feather-icons/src/icons/UserPlusIcon.svelte';
  
  // Components
  import TweetCard from '../components/social/TweetCard.svelte';
  import Spinner from '../components/common/Spinner.svelte';
  import UserCard from '../components/social/UserCard.svelte';
  import TabButtons from '../components/common/TabButtons.svelte';
  import Button from '../components/common/Button.svelte';
    // Define types for our data
  interface Community {
    id: string;
    name: string;
    description: string;
    logo_url?: string;
    logo?: string;
    banner_url?: string;
    banner?: string;
    creator_id?: string;
    creatorId?: string;
    is_approved?: boolean;
    isApproved?: boolean;
    is_private?: boolean;
    isPrivate?: boolean;
    categories?: string[];
    created_at?: Date | string;
    createdAt?: Date | string;
    member_count?: number;
    memberCount?: number;
  }
  
  interface Member {
    id: string;
    user_id: string;
    username: string;
    name: string;
    role: string;
    avatar_url: string;
    joined_at?: Date;
  }
  
  interface Rule {
    id: string;
    communityId?: string;
    title: string;
    description: string;
    order: number;
  }
  
  // Make Thread compatible with ITweet
  interface Thread extends ITweet {
    authorId?: string;
    createdAt?: Date;
  }
  
  const logger = createLoggerWithPrefix('CommunityDetail');
  
  const { getAuthState } = useAuth();
  const { theme } = useTheme();
    $: authState = getAuthState ? (getAuthState() as IAuthStore) : { 
    user_id: null, 
    is_authenticated: false, 
    access_token: null, 
    refresh_token: null 
  };
  $: isDarkMode = $theme === 'dark';
  
  // Get community ID from URL
  let communityId = '';
  $: {
    if (typeof window !== 'undefined') {
      const urlParts = window.location.pathname.split('/');
      if (urlParts.length > 2 && urlParts[1] === 'communities') {
        communityId = urlParts[2];
        console.log(`Detected community ID from URL: ${communityId}`);
      }
    }
  }
  
  // Community data
  let community: Community | null = null;
  let isLoading = true;
  let isMember = false;
  let isPending = false;
  let members: Member[] = [];
  let rules: Rule[] = [];
  let threads: Thread[] = [];
  let activeTab = 'posts'; // 'posts', 'members', 'rules', 'about'
  let errorMessage = '';
  
  onMount(async () => {
    if (!communityId) {
      toastStore.showToast('Invalid community ID', 'error');
      window.location.href = '/communities';
      return;
    }
    
    try {
      console.log(`Loading data for community: ${communityId}`);
      await loadCommunityData();
    } catch (error) {
      logger.error('Failed to load community data', error);
      errorMessage = error instanceof Error ? error.message : 'Unknown error';
      toastStore.showToast('Failed to load community data', 'error');
    }
  });
  
  // Helper function to get the Supabase URL for community logos/banners
  function getImageUrl(url, type = 'logo') {
    if (!url) return null;
    
    // Check if the URL is already a Supabase URL
    if (url.includes('supabase')) {
      return url;
    }
    
    // If it's just a path, construct the Supabase URL
    if (url.startsWith('/')) {
      return getPublicUrl(SUPABASE_BUCKETS.MEDIA, `communities${url}`);
    }
    
    // Try to extract the path if it's in a special format
    const parts = url.split('/');
    if (parts.length > 0) {
      const filename = parts[parts.length - 1];
      const folder = type === 'logo' ? 'logos' : 'banners';
      return getPublicUrl(SUPABASE_BUCKETS.MEDIA, `communities/${folder}/${filename}`);
    }
    
    return url;
  }
  
  async function loadCommunityData() {
    try {
      isLoading = true;
      errorMessage = '';
      
      // Get community details
      console.log(`Calling getCommunityById for community: ${communityId}`);
      const communityResponse = await getCommunityById(communityId);
      console.log('Community response:', communityResponse);
      
      if (!communityResponse || !communityResponse.community) {
        errorMessage = 'Community not found';
        throw new Error('Community not found');
      }
      
      // Normalize community data fields
      const communityData = communityResponse.community;
      community = {
        id: communityData.id || communityId,
        name: communityData.name || 'Unnamed Community',
        description: communityData.description || '',
        logo_url: communityData.logo_url || communityData.logo || '',
        banner_url: communityData.banner_url || communityData.banner || '',
        creator_id: communityData.creator_id || communityData.creatorId || '',
        is_approved: communityData.is_approved != null ? communityData.is_approved : (communityData.isApproved || false),
        is_private: communityData.is_private || communityData.isPrivate || false,
        categories: communityData.categories || [],
        created_at: communityData.created_at || communityData.createdAt || new Date(),
        member_count: communityData.member_count || communityData.memberCount || 0
      };
      
      console.log('Normalized community data:', community);
      
      // Check membership status
      const membershipResponse = await checkUserCommunityMembership(communityId);
      console.log('Membership response:', membershipResponse);
      isMember = membershipResponse?.status === 'member';
      isPending = membershipResponse?.status === 'pending';
      
      // Load posts, members, and rules in parallel
      await Promise.all([
        loadThreads(),
        loadMembers(),
        loadRules()
      ]);
      
    } catch (error) {
      // This will rarely happen now since getCommunityById returns default data instead of throwing
      logger.error('Error loading community data:', error);
      errorMessage = error instanceof Error ? error.message : 'Unknown error';
      
      if (error instanceof SyntaxError && error.message.includes('Unexpected end of JSON')) {
        toastStore.showToast('Unable to load community data. The server returned an invalid response.', 'error');
      } else if (error instanceof Error && error.message.includes('Empty response from server')) {
        toastStore.showToast('Unable to load community data. The server returned an empty response.', 'error');
      } else if (error instanceof Error && error.message.includes('Community not found')) {
        toastStore.showToast('The community you are looking for does not exist or has been removed.', 'error');
      } else {
        toastStore.showToast('Failed to load community data: ' + (error instanceof Error ? error.message : 'Unknown error'), 'error');
      }
      
    } finally {
      isLoading = false;
    }
  }
  
  async function loadThreads() {
    try {
      // For community posts, we use the getUserThreads with community parameter
      // This doesn't match the function signature, but the implementation accepts more params
      const threadsResponse = await getUserThreads(communityId, 1, 10);
      // Alternatively use query string parameter:
      // const threadsResponse = await getUserThreads(`${communityId}?communityId=${communityId}`);
      
      threads = (threadsResponse?.threads || []) as Thread[];
      
    } catch (error) {
      logger.error('Error loading community threads:', error);
      threads = [];
    }
  }
  
  async function loadMembers() {
    try {
      const membersResponse = await listMembers(communityId);
      members = membersResponse?.members || [];
    } catch (error) {
      logger.error('Error loading community members:', error);
      members = [];
    }
  }
  
  async function loadRules() {
    try {
      const rulesResponse = await listRules(communityId);
      rules = rulesResponse?.rules || [];
    } catch (error) {
      logger.error('Error loading community rules:', error);
      rules = [];
    }
  }
  
  async function handleJoinRequest() {
    if (!authState.is_authenticated) {
      toastStore.showToast('You need to log in to join communities', 'warning');
      return;
    }
    
    try {
      await requestToJoin(communityId, {});
      isPending = true;
      toastStore.showToast('Join request sent successfully', 'success');
    } catch (error) {
      logger.error('Error joining community:', error);
      toastStore.showToast('Failed to join community. Please try again.', 'error');
    }
  }
  
  const tabItems = [
    { id: 'posts', label: 'Posts', icon: MessageSquareIcon },
    { id: 'members', label: 'Members', icon: UsersIcon },
    { id: 'rules', label: 'Rules', icon: AlertCircleIcon },
    { id: 'about', label: 'About', icon: InfoIcon }
  ];
</script>

<MainLayout>
  <div class="community-detail">
    {#if isLoading}
      <div class="loading-container">
        <Spinner size="large" />
      </div>
    {:else if community}
      <div class="community-header">
        <div class="community-banner" style={community.banner_url ? `background-image: url(${getImageUrl(community.banner_url, 'banner')})` : ''}>
          <div class="community-info-overlay">
            <div class="community-logo-container">
              {#if community.logo_url}
                <img src={getImageUrl(community.logo_url, 'logo')} alt={community.name} class="community-logo" />
              {:else}
                <div class="community-logo-placeholder">
                  {community.name.charAt(0).toUpperCase()}
                </div>
              {/if}
            </div>
            
            <div class="community-header-details">
              <div class="community-name-row">
                <h1 class="community-name">{community.name}</h1>
                {#if community.is_private}
                  <div class="community-badge private">
                    <LockIcon size="16" />
                    <span>Private</span>
                  </div>
                {/if}
              </div>
              
              <div class="community-stats">
                <div class="stat">
                  <UsersIcon size="16" />
                  <span>{community.member_count || members.length} Members</span>
                </div>
                <div class="stat">
                  <MessageSquareIcon size="16" />
                  <span>{threads.length} Posts</span>
                </div>
              </div>
              
              <div class="community-actions">
                {#if isMember}
                  <Button variant="outlined" icon={LogOutIcon}>
                    Leave Community
                  </Button>
                {:else if isPending}
                  <Button variant="outlined" disabled>
                    Join Request Pending
                  </Button>
                {:else}
                  <Button variant="primary" icon={UserPlusIcon} on:click={handleJoinRequest}>
                    Join Community
                  </Button>
                {/if}
              </div>
            </div>
          </div>
        </div>
        
        <TabButtons
          items={tabItems}
          activeId={activeTab}
          on:tabChange={(e) => activeTab = e.detail}
        />
      </div>
      
      <div class="community-content">
        {#if activeTab === 'posts'}
          {#if threads.length > 0}
            <div class="threads-container">
              {#each threads as thread (thread.id)}
                <TweetCard tweet={thread as any} />
              {/each}
            </div>
          {:else}
            <div class="empty-state">
              <MessageSquareIcon size="48" />
              <h2>No posts yet</h2>
              <p>Be the first to post in this community!</p>
              {#if isMember}
                <Button variant="primary">Create Post</Button>
              {/if}
            </div>
          {/if}
        
        {:else if activeTab === 'members'}
          <div class="members-container">
            <h2 class="section-title">Members ({members.length})</h2>
            {#if members.length > 0}
              <div class="members-grid">
                {#each members as member (member.id)}
                  <UserCard user={member} />
                {/each}
              </div>
            {:else}
              <div class="empty-state">
                <UsersIcon size="48" />
                <p>No members found</p>
              </div>
            {/if}
          </div>
          
        {:else if activeTab === 'rules'}
          <div class="rules-container">
            <h2 class="section-title">Community Rules</h2>
            {#if rules.length > 0}
              <div class="rules-list">
                {#each rules as rule, i (rule.id)}
                  <div class="rule-item">
                    <div class="rule-number">{i + 1}</div>
                    <div class="rule-content">
                      <h3 class="rule-title">{rule.title}</h3>
                      <p class="rule-description">{rule.description}</p>
                    </div>
                  </div>
                {/each}
              </div>
            {:else}
              <div class="empty-state">
                <AlertCircleIcon size="48" />
                <p>No rules have been set for this community</p>
              </div>
            {/if}
          </div>
          
        {:else if activeTab === 'about'}
          <div class="about-container">
            <h2 class="section-title">About {community.name}</h2>
            <div class="community-description">
              <p>{community.description || 'No description provided'}</p>
            </div>
            
            {#if community.categories && community.categories.length > 0}
              <div class="categories-section">
                <h3>Categories</h3>
                <div class="categories-list">
                  {#each community.categories as category}
                    <span class="category-tag">{category}</span>
                  {/each}
                </div>
              </div>
            {/if}
              {#if community.created_at}
              <div class="community-metadata">
                <p>Created: {new Date(community.created_at).toLocaleDateString()}</p>
              </div>
            {/if}
          </div>
        {/if}
      </div>
    {:else}
      <div class="error-container">
        <AlertCircleIcon size="48" />
        <h2>Community Not Found</h2>
        <p>The community you're looking for doesn't exist or you don't have permission to view it.</p>
        {#if errorMessage}
          <p class="error-details">Error: {errorMessage}</p>
        {/if}
        <a href="/communities" class="back-link">Back to Communities</a>
      </div>
    {/if}
  </div>
</MainLayout>

<style>
  .community-detail {
    width: 100%;
    max-width: 100%;
    min-height: 100vh;
  }
  
  .loading-container,
  .error-container,
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: var(--space-10) var(--space-4);
    gap: var(--space-4);
    text-align: center;
  }
  
  .error-container h2,
  .empty-state h2 {
    font-size: var(--font-size-xl);
    font-weight: var(--font-weight-bold);
    margin: var(--space-2) 0;
  }
  
  .error-container p,
  .empty-state p {
    color: var(--text-secondary);
    margin-bottom: var(--space-4);
  }
  
  .back-link {
    display: inline-block;
    padding: var(--space-2) var(--space-4);
    background-color: var(--color-primary);
    color: white;
    border-radius: var(--radius-full);
    text-decoration: none;
    font-weight: var(--font-weight-medium);
  }
  
  .community-header {
    width: 100%;
    border-bottom: 1px solid var(--border-color);
  }
  
  .community-banner {
    width: 100%;
    height: 200px;
    background-color: var(--color-primary);
    background-size: cover;
    background-position: center;
    position: relative;
  }
  
  .community-info-overlay {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    padding: var(--space-4);
    background: linear-gradient(0deg, rgba(0, 0, 0, 0.6) 0%, rgba(0, 0, 0, 0) 100%);
    display: flex;
    align-items: flex-end;
    color: white;
  }
  
  .community-logo-container {
    margin-right: var(--space-4);
    flex-shrink: 0;
  }
  
  .community-logo,
  .community-logo-placeholder {
    width: 80px;
    height: 80px;
    border-radius: 50%;
    border: 4px solid white;
    background-color: var(--color-primary);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: var(--font-size-3xl);
    font-weight: var(--font-weight-bold);
  }
  
  .community-header-details {
    flex: 1;
  }
  
  .community-name-row {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-bottom: var(--space-2);
  }
  
  .community-name {
    font-size: var(--font-size-2xl);
    font-weight: var(--font-weight-bold);
    margin: 0;
    text-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
  }
  
  .community-badge {
    display: flex;
    align-items: center;
    padding: var(--space-1) var(--space-2);
    border-radius: var(--radius-full);
    font-size: var(--font-size-sm);
    gap: var(--space-1);
  }
  
  .community-badge.private {
    background-color: rgba(0, 0, 0, 0.3);
    backdrop-filter: blur(5px);
  }
  
  .community-stats {
    display: flex;
    gap: var(--space-4);
    margin-bottom: var(--space-3);
  }
  
  .stat {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    font-size: var(--font-size-sm);
  }
  
  .community-content {
    padding: var(--space-4);
  }
  
  .section-title {
    font-size: var(--font-size-xl);
    font-weight: var(--font-weight-bold);
    margin-bottom: var(--space-4);
    padding-bottom: var(--space-2);
    border-bottom: 1px solid var(--border-color);
  }
  
  .threads-container {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }
  
  .members-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: var(--space-4);
  }
  
  .rules-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }
  
  .rule-item {
    display: flex;
    gap: var(--space-3);
    padding: var(--space-3);
    background-color: var(--bg-secondary);
    border-radius: var(--radius-md);
  }
  
  .rule-number {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    background-color: var(--color-primary);
    color: white;
    border-radius: 50%;
    font-weight: var(--font-weight-bold);
    flex-shrink: 0;
  }
  
  .rule-content {
    flex: 1;
  }
  
  .rule-title {
    font-weight: var(--font-weight-bold);
    margin-bottom: var(--space-1);
  }
  
  .rule-description {
    color: var(--text-secondary);
    font-size: var(--font-size-sm);
  }
  
  .community-description {
    margin-bottom: var(--space-6);
    line-height: 1.6;
  }
  
  .categories-section {
    margin-bottom: var(--space-6);
  }
  
  .categories-section h3 {
    font-size: var(--font-size-lg);
    font-weight: var(--font-weight-medium);
    margin-bottom: var(--space-3);
  }
  
  .categories-list {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }
  
  .category-tag {
    padding: var(--space-1) var(--space-3);
    background-color: var(--bg-accent);
    border-radius: var(--radius-full);
    font-size: var(--font-size-sm);
  }
  
  .community-metadata {
    color: var(--text-secondary);
    font-size: var(--font-size-sm);
  }
  
  @media (max-width: 768px) {
    .community-info-overlay {
      flex-direction: column;
      align-items: center;
      text-align: center;
    }
    
    .community-logo-container {
      margin: 0 0 var(--space-3) 0;
    }
    
    .community-name-row {
      justify-content: center;
    }
    
    .community-stats {
      justify-content: center;
    }
    
    .members-grid {
      grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
    }
  }
  
  .error-details {
    color: var(--color-danger, #e53e3e);
    margin-bottom: var(--space-4);
    font-size: var(--font-size-sm);
  }
</style>
