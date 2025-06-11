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
    listRules,
    removeMember,
    listJoinRequests,
    approveJoinRequest,
    rejectJoinRequest
  } from '../api/community';
  import { getUserThreads } from '../api/thread';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import { getPublicUrl, SUPABASE_BUCKETS } from '../utils/supabase';
  import type { IAuthStore } from '../interfaces/IAuth';
  import { formatStorageUrl } from '../utils/common';
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
  
  // Community-specific components
  import CommunityPosts from '../components/communities/CommunityPosts.svelte';
  import CommunityMembers from '../components/communities/CommunityMembers.svelte';
  import CommunityRules from '../components/communities/CommunityRules.svelte';
  import CommunityAbout from '../components/communities/CommunityAbout.svelte';
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
    requested_at?: Date | string;
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
  
  // Get community ID from URL or props
  export let communityId = '';
  
  // If not provided directly, extract from URL
  $: {
    if (!communityId && typeof window !== 'undefined') {
      const urlParts = window.location.pathname.split('/');
      if (urlParts.length > 2 && urlParts[1] === 'communities') {
        communityId = urlParts[2];
        console.log(`Extracted community ID from URL: ${communityId}`);
      }
    }
  }
  
  // Community data
  let community: Community | null = null;
  let isLoading = true;
  let isMember = false;
  let isPending = false;
  let members: Member[] = [];
  let userRole = 'member'; // Can be 'owner', 'admin', 'moderator', 'member'
  let rules: Rule[] = [];
  let threads: Thread[] = [];
  let pendingMembers: Member[] = [];
  let activeTab = 'posts'; // 'posts', 'members', 'rules', 'about'
  let errorMessage = '';
  
  onMount(async () => {
    if (!communityId) {
      console.error('No community ID available');
      toastStore.showToast('Invalid community ID', 'error');
      window.location.href = '/communities';
      return;
    }
    
    console.log(`CommunityDetail component mounted with ID: ${communityId}`);
    
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
    
    // Use the shared formatStorageUrl utility for consistent image handling
    return formatStorageUrl(url);
  }
  
  async function loadCommunityData() {
    try {
      isLoading = true;
      errorMessage = '';
      
      // Get community details
      console.log(`Calling getCommunityById for community: ${communityId}`);
      let communityResponse;
      
      try {
        // Try the main community API
        communityResponse = await getCommunityById(communityId);
        console.log('Community response from API:', communityResponse);
        
        if (!communityResponse || (!communityResponse.community && !communityResponse.data)) {
          throw new Error('Invalid API response format');
        }
      } catch (err) {
        console.warn('Primary API call failed:', err);
        errorMessage = 'Community not found';
        throw new Error('Community not found or inaccessible');
      }
      
      // Extract community data handling different response formats
      let communityData;
      if (communityResponse.community) {
        communityData = communityResponse.community;
      } else if (communityResponse.data) {
        communityData = communityResponse.data;
      } else {
        errorMessage = 'Invalid community data format';
        throw new Error('Invalid community data format');
      }
      
      // Normalize community data fields
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
      try {
        // Only check membership if user is logged in
        const authData = localStorage.getItem('auth');
        if (authData) {
          const auth = JSON.parse(authData);
          if (auth.access_token && (!auth.expires_at || new Date(auth.expires_at) > new Date())) {
            const membershipResponse = await checkUserCommunityMembership(communityId);
            console.log('Membership response:', membershipResponse);
            
            interface MembershipData {
              is_member?: boolean;
              status?: string;
              user_role?: string;
            }
            
            interface MembershipResponse {
              status?: string;
              is_member?: boolean;
              user_role?: string;
              data?: MembershipData;
            }
            
            // Cast the response to the proper type
            const typedResponse = membershipResponse as MembershipResponse;
            
            // Check various response formats for membership status
            if (typedResponse?.status === 'member' || 
                typedResponse?.is_member === true || 
                typedResponse?.data?.is_member === true ||
                typedResponse?.data?.status === 'member') {
              isMember = true;
              console.log('User is a member of this community');
              
              // Get the user's role in the community
              userRole = typedResponse?.user_role || 
                         typedResponse?.data?.user_role ||
                         'member';
                         
              console.log(`User has role "${userRole}" in this community`);
            } else if (typedResponse?.status === 'pending' || 
                       typedResponse?.data?.status === 'pending') {
              isPending = true;
              console.log('User has a pending join request for this community');
            } else {
              isMember = false;
              isPending = false;
              console.log('User is not a member of this community');
            }
          }
        } else {
          console.log('User not logged in, skipping membership check');
          isMember = false;
          isPending = false;
        }
      } catch (membershipError) {
        console.warn('Error checking membership status:', membershipError);
        // Default to non-member if check fails
        isMember = false;
        isPending = false;
      }
      
      // Load posts, members, and rules in parallel
      try {
        await Promise.allSettled([
          loadThreads(),
          loadMembers(),
          loadRules(),
          loadPendingMembers()
        ]);
      } catch (loadError) {
        console.warn('Error loading related community data:', loadError);
        // We can continue even if these fail
      }
      
    } catch (error) {
      logger.error('Error loading community data:', error);
      errorMessage = error instanceof Error ? error.message : 'Unknown error';
      
      if (!community) {
        // Only show a toast if we couldn't load the community at all
        if (error instanceof SyntaxError && error.message.includes('Unexpected end of JSON')) {
          toastStore.showToast('Unable to load community data. The server returned an invalid response.', 'error');
        } else if (error instanceof Error && error.message.includes('Empty response from server')) {
          toastStore.showToast('Unable to load community data. The server returned an empty response.', 'error');
        } else if (error instanceof Error && error.message.includes('Community not found')) {
          toastStore.showToast('The community you are looking for does not exist or has been removed.', 'error');
        } else {
          toastStore.showToast('Failed to load community data', 'error');
        }
      }
    } finally {
      isLoading = false;
    }
  }
  
  async function loadThreads() {
    try {
      // For community posts, we use the getUserThreads with community parameter
      const threadsResponse = await getUserThreads(communityId, 1, 10);
      
      if (threadsResponse && Array.isArray(threadsResponse.threads)) {
        threads = threadsResponse.threads as Thread[];
      } else if (threadsResponse && threadsResponse.data && Array.isArray(threadsResponse.data.threads)) {
        threads = threadsResponse.data.threads as Thread[];
      } else {
        threads = [];
      }
      
    } catch (error) {
      logger.error('Error loading community threads:', error);
      threads = [];
    }
  }
  
  async function loadMembers() {
    try {
      const membersResponse = await listMembers(communityId);
      
      if (membersResponse && Array.isArray(membersResponse.members)) {
        members = processMembersAvatars(membersResponse.members);
      } else if (membersResponse && membersResponse.data && Array.isArray(membersResponse.data.members)) {
        members = processMembersAvatars(membersResponse.data.members);
      } else {
        members = [];
      }
      
    } catch (error) {
      logger.error('Error loading community members:', error);
      members = [];
    }
  }
  
  // Process member avatars to use Supabase URLs
  function processMembersAvatars(membersList) {
    return membersList.map(member => {
      const processedMember = { ...member };
      
      // Handle different avatar field names
      const avatarUrl = member.avatar_url || member.profile_picture_url || member.avatar || '';
      
      if (avatarUrl) {
        // Process the avatar URL to use Supabase
        processedMember.avatar_url = getProfileImageUrl(avatarUrl);
      }
      
      // Ensure we have username and name even if they weren't in the original data
      // This handles both regular members and pending join requests
      if (!processedMember.username && processedMember.user_id) {
        // Generate a readable username from the user ID
        const userId = processedMember.user_id;
        processedMember.username = `user_${userId.substring(0, 8)}`;
      }
      
      if (!processedMember.name) {
        processedMember.name = processedMember.username || 'Unknown User';
      }
      
      console.log('Processed member data:', {
        id: processedMember.id,
        user_id: processedMember.user_id,
        username: processedMember.username,
        name: processedMember.name,
        role: processedMember.role
      });
      
      return processedMember;
    });
  }
  
  // Helper function to get Supabase URL for profile pictures
  function getProfileImageUrl(url) {
    if (!url) return null;
    
    // Use the shared formatStorageUrl utility for consistent image handling
    return formatStorageUrl(url);
  }
  
  async function loadPendingMembers() {
    try {
      // Only attempt to load pending members if user is authenticated and community exists
      if (!authState.is_authenticated || !community || !community.id) {
        pendingMembers = [];
        return;
      }
      
      const pendingResponse = await listJoinRequests(communityId);
      console.log('Raw pending join requests response:', JSON.stringify(pendingResponse, null, 2));
      
      if (pendingResponse && Array.isArray(pendingResponse.join_requests)) {
        console.log(`Found ${pendingResponse.join_requests.length} pending join requests`);
        
        // Debug log to see the structure of the first join request (if available)
        if (pendingResponse.join_requests.length > 0) {
          const sampleRequest = pendingResponse.join_requests[0];
          console.log('Example join request structure:', sampleRequest);
          console.log('All fields in join request:', Object.keys(sampleRequest));
          console.log('Join request ID:', sampleRequest.id);
          console.log('Join request user_id:', sampleRequest.user_id);
          console.log('Join request username field value:', sampleRequest.username);
          console.log('Join request name field value:', sampleRequest.name || sampleRequest.display_name);
          
          // If the join request contains a user object, log its structure too
          if (sampleRequest.user) {
            console.log('User object in join request:', sampleRequest.user);
            console.log('User object fields:', Object.keys(sampleRequest.user));
            console.log('User object username:', sampleRequest.user.username);
          }
        }
          // Format users from join requests to match Member structure
        pendingMembers = pendingResponse.join_requests.map(request => {
          console.log(`Processing request for user_id: ${request.user_id}, found username: ${request.username || 'MISSING'}`);
          
          // The backend now returns real user data, so prioritize that
          const member = {
            id: request.id || request.user_id || '',
            user_id: request.user_id || '',
            // Prioritize real username from backend, fallback only if not available
            username: request.username || `user_${(request.user_id || '').substring(0, 8)}`,
            name: request.name || request.username || `User ${(request.user_id || '').substring(0, 8)}`,
            role: 'pending',
            avatar_url: request.avatar_url || request.profile_picture_url || '',
            requested_at: request.created_at || new Date()
          };
          
          console.log('Created member object:', member);
          return member;
        });
        
        // Process avatars for pending members
        pendingMembers = processMembersAvatars(pendingMembers);
      } else if (pendingResponse && pendingResponse.data && Array.isArray(pendingResponse.data.join_requests)) {
        // Similar debugging for alternative response format
        console.log(`Found ${pendingResponse.data.join_requests.length} pending join requests (alt format)`);
        
        if (pendingResponse.data.join_requests.length > 0) {
          const sampleRequest = pendingResponse.data.join_requests[0];
          console.log('Example join request structure (alt format):', sampleRequest);
          console.log('All fields in join request (alt format):', Object.keys(sampleRequest));
          console.log('Join request ID (alt):', sampleRequest.id);
          console.log('Join request user_id (alt):', sampleRequest.user_id);
          console.log('Join request username field value (alt):', sampleRequest.username);
        }
          // Format users from join requests (alternative response format)
        pendingMembers = pendingResponse.data.join_requests.map(request => {
          console.log(`Processing alt request for user_id: ${request.user_id}, found username: ${request.username || 'MISSING'}`);
          
          // The backend now returns real user data, so prioritize that
          const member = {
            id: request.id || request.user_id || '',
            user_id: request.user_id || '',
            // Prioritize real username from backend, fallback only if not available
            username: request.username || `user_${(request.user_id || '').substring(0, 8)}`,
            name: request.name || request.username || `User ${(request.user_id || '').substring(0, 8)}`,
            role: 'pending',
            avatar_url: request.avatar_url || request.profile_picture_url || '',
            requested_at: request.created_at || new Date()
          };
          
          console.log('Created member object (alt):', member);
          return member;
        });
        
        // Process avatars for pending members
        pendingMembers = processMembersAvatars(pendingMembers);
      } else {
        console.log('No valid join requests found in the response');
        pendingMembers = [];
      }
      
      console.log('Final processed pending members:', pendingMembers);
    } catch (error) {
      logger.error('Error loading pending join requests:', error);
      pendingMembers = [];
    }
  }
  
  async function loadRules() {
    try {
      const rulesResponse = await listRules(communityId);
      
      if (rulesResponse && Array.isArray(rulesResponse.rules)) {
        rules = rulesResponse.rules;
      } else if (rulesResponse && rulesResponse.data && Array.isArray(rulesResponse.data.rules)) {
        rules = rulesResponse.data.rules;
      } else {
        rules = [];
      }
      
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
  
  // Function to handle join/leave community toggle
  function toggleJoinCommunity() {
    if (isMember) {
      // Use removeMember function from API instead of leaveCommunity
      removeMember(communityId, authState.user_id || '')
        .then(() => {
          isMember = false;
          toastStore.showToast('Left community successfully', 'success');
        })
        .catch(error => {
          logger.error('Error leaving community:', error);
          toastStore.showToast('Failed to leave community', 'error');
        });
    } else {
      // Use requestToJoin function instead of joinCommunity
      handleJoinRequest();
    }
  }
  
  // Handle thread click - navigate to thread detail
  function handleThreadClick(event) {
    const tweet = event.detail;
    if (!tweet || !tweet.id) {
      console.error('Invalid tweet data for navigation', tweet);
      return;
    }
    
    const threadId = tweet.id;
    console.log(`Navigating to thread detail: ${threadId}`);
    
    // Construct the URL for thread detail
    const href = `/thread/${threadId}`;
    
    // Use navigation approach
    try {
      // First try to use history API for SPA navigation
      window.history.pushState({threadId}, '', href);
      
      // Dispatch a custom navigation event to trigger router update
      const navEvent = new CustomEvent('navigate', { 
        detail: { href, threadId } 
      });
      window.dispatchEvent(navEvent);
      
      // Trigger popstate as a fallback
      window.dispatchEvent(new PopStateEvent('popstate', {}));
      
      // If nothing works, reload the page after a short delay
      setTimeout(() => {
        if (window.location.pathname !== href) {
          console.warn('Navigation did not update the URL, forcing page reload');
          window.location.href = href;
        }
      }, 300);
    } catch (error) {
      console.error('Error in navigation:', error);
      window.location.href = href; // Direct navigation as fallback
    }
  }
  
  // Check if the user can post in this community
  function canPostInCommunity(): boolean {
    // User must be logged in, a member, and community must be approved
    return authState.is_authenticated && isMember && community?.is_approved === true;
  }
  
  // Function to handle thread creation
  function handleCreatePost() {
    // Only allow post creation if community is approved
    if (!community?.is_approved) {
      toastStore.showToast('This community is pending approval. You cannot create posts yet.', 'warning');
      return;
    }
    
    // Navigate to create post page with community context
    const href = `/create-post?community=${communityId}`;
    window.location.href = href;
  }
  
  async function handleApproveJoinRequest(requestId: string) {
    try {
      if (!authState.is_authenticated) {
        toastStore.showToast('You need to log in to approve join requests', 'warning');
        return;
      }
      
      // Call the API to approve the join request
      await approveJoinRequest(communityId, requestId);
      toastStore.showToast('Join request approved successfully', 'success');
      
      // Reload members and pending members
      await Promise.all([
        loadMembers(),
        loadPendingMembers()
      ]);
    } catch (error) {
      logger.error('Error approving join request:', error);
      toastStore.showToast('Failed to approve join request. Please try again.', 'error');
    }
  }
  
  async function handleRejectJoinRequest(requestId: string) {
    try {
      if (!authState.is_authenticated) {
        toastStore.showToast('You need to log in to reject join requests', 'warning');
        return;
      }
      
      // Call the API to reject the join request
      await rejectJoinRequest(communityId, requestId);
      toastStore.showToast('Join request rejected', 'success');
      
      // Reload pending members
      await loadPendingMembers();
    } catch (error) {
      logger.error('Error rejecting join request:', error);
      toastStore.showToast('Failed to reject join request. Please try again.', 'error');
    }
  }
  
  // Check if user can manage the community (approve/reject join requests, etc.)
  function canManageCommunity(): boolean {
    // User must be logged in and have appropriate role
    // Ownership is determined either by userRole or by being the creator
    const isOwner = userRole === 'owner' || 
                    (community?.creator_id && authState.user_id === community.creator_id);
    const isAdmin = userRole === 'admin';
    const isModerator = userRole === 'moderator';
    
    return authState.is_authenticated && (isOwner || isAdmin || isModerator);
  }
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
                {#if !community.is_approved}
                  <div class="community-badge pending">
                    <AlertCircleIcon size="16" />
                    <span>Pending Admin Approval</span>
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
                {#if !community.is_approved}
                  <Button variant="outlined" disabled>
                    Community Awaiting Approval
                  </Button>
                {:else if isMember}
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
          <CommunityPosts 
            {threads}
            {isMember}
            canPostInCommunity={canPostInCommunity()}
            communityIsApproved={community.is_approved}
            on:threadClick={handleThreadClick}
            on:createPost={handleCreatePost}
          />
        
        {:else if activeTab === 'members'}
          <CommunityMembers 
            {members}
            {pendingMembers}
            canManageCommunity={canManageCommunity()}
            on:approveJoinRequest={(e) => handleApproveJoinRequest(e.detail)}
            on:rejectJoinRequest={(e) => handleRejectJoinRequest(e.detail)}
          />
          
        {:else if activeTab === 'rules'}
          <CommunityRules {rules} />
          
        {:else if activeTab === 'about'}
          <CommunityAbout {community} />
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
  .error-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: var(--space-10) var(--space-4);
    gap: var(--space-4);
    text-align: center;
  }
  
  .error-container h2 {
    font-size: var(--font-size-xl);
    font-weight: var(--font-weight-bold);
    margin: var(--space-2) 0;
  }
  
  .error-container p {
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
  
  .community-badge.pending {
    background-color: rgba(255, 193, 7, 0.2);
    color: #ff9800;
    border: 1px solid rgba(255, 152, 0, 0.3);
    backdrop-filter: blur(5px);
  }
  
  :global(.dark) .community-badge.pending {
    background-color: rgba(255, 193, 7, 0.1);
    color: #ffb74d;
    border: 1px solid rgba(255, 193, 7, 0.3);
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
  
  .error-details {
    color: var(--color-danger, #e53e3e);
    margin-bottom: var(--space-4);
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
  }
</style>
