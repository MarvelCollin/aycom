<script lang="ts">
  import { onMount } from "svelte";
  import { toastStore } from "../stores/toastStore";
  import MainLayout from "../components/layout/MainLayout.svelte";
  import { createLoggerWithPrefix } from "../utils/logger";
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
  } from "../api/community";
  import { getUserThreads, getAllThreads } from "../api/thread";
  import { useAuth } from "../hooks/useAuth";
  import { useTheme } from "../hooks/useTheme";
  import { getPublicUrl, SUPABASE_BUCKETS } from "../utils/supabase";
  import type { IAuthStore } from "../interfaces/IAuth";
  import { formatStorageUrl } from "../utils/common";
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
  import UsersIcon from "svelte-feather-icons/src/icons/UsersIcon.svelte";
  import InfoIcon from "svelte-feather-icons/src/icons/InfoIcon.svelte";
  import BookmarkIcon from "svelte-feather-icons/src/icons/BookmarkIcon.svelte";
  import AlertCircleIcon from "svelte-feather-icons/src/icons/AlertCircleIcon.svelte";
  import MessageSquareIcon from "svelte-feather-icons/src/icons/MessageSquareIcon.svelte";
  import LockIcon from "svelte-feather-icons/src/icons/LockIcon.svelte";
  import LogOutIcon from "svelte-feather-icons/src/icons/LogOutIcon.svelte";
  import UserPlusIcon from "svelte-feather-icons/src/icons/UserPlusIcon.svelte";
  import TrendingUpIcon from "svelte-feather-icons/src/icons/TrendingUpIcon.svelte";
  import ClockIcon from "svelte-feather-icons/src/icons/ClockIcon.svelte";
  import ImageIcon from "svelte-feather-icons/src/icons/ImageIcon.svelte";
  import UserCheckIcon from "svelte-feather-icons/src/icons/UserCheckIcon.svelte";
    // Components
  import TweetCard from "../components/social/TweetCard.svelte";
  import Spinner from "../components/common/Spinner.svelte";
  import UserCard from "../components/social/UserCard.svelte";
  import TabButtons from "../components/common/TabButtons.svelte";
  import Button from "../components/common/Button.svelte";

  // Community-specific components
  import CommunityPosts from "../components/communities/CommunityPosts.svelte";
  import CommunityMembers from "../components/communities/CommunityMembers.svelte";
  import CommunityRules from "../components/communities/CommunityRules.svelte";
  import CommunityAbout from "../components/communities/CommunityAbout.svelte";
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
    order: number;  }

  // Make Thread compatible with ITweet
  interface Thread extends ITweet {
    authorId?: string;
    createdAt?: Date;
    media_url?: string;
    media_type?: 'image' | 'video';
  }

  const logger = createLoggerWithPrefix("CommunityDetail");

  const { getAuthState } = useAuth();
  const { theme } = useTheme();
    $: authState = getAuthState ? (getAuthState() as IAuthStore) : {
    user_id: null,
    is_authenticated: false,
    access_token: null,
    refresh_token: null
  };
  $: isDarkMode = $theme === "dark";

  // Get community ID from URL or props
  export let communityId = "";

  // If not provided directly, extract from URL
  $: {
    if (!communityId && typeof window !== "undefined") {
      const urlParts = window.location.pathname.split("/");
      if (urlParts.length > 2 && urlParts[1] === "communities") {
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
  let userRole = "member"; // Can be 'owner', 'admin', 'moderator', 'member'
  let rules: Rule[] = [];
  let threads: Thread[] = [];
  let pendingMembers: Member[] = [];
  let activeTab = "top"; // 'top', 'latest', 'media', 'about', 'manage'
  let errorMessage = "";
  // Add properties to store different thread types
  let topThreads: Thread[] = [];
  let latestThreads: Thread[] = [];
  let mediaThreads: Thread[] = [];
  let isLoadingTopThreads = false;
  let isLoadingLatestThreads = false;
  let isLoadingMediaThreads = false;
  let topMembers: Member[] = [];
  let isLoadingTopMembers = false;

  // Media pagination
  let mediaPage = 1;
  let hasMoreMedia = true;
  let isLoadingMoreMedia = false;

  onMount(async () => {
    if (!communityId) {
      console.error("No community ID available");
      toastStore.showToast("Invalid community ID", "error");
      window.location.href = "/communities";
      return;
    }

    console.log(`CommunityDetail component mounted with ID: ${communityId}`);

    try {
      console.log(`Loading data for community: ${communityId}`);
      await loadCommunityData();
    } catch (error) {
      logger.error("Failed to load community data", error);
      errorMessage = error instanceof Error ? error.message : "Unknown error";
      toastStore.showToast("Failed to load community data", "error");
    }
  });

  // Helper function to get the Supabase URL for community logos/banners
  function getImageUrl(url, type = "logo") {
    if (!url) return null;

    // Use the shared formatStorageUrl utility for consistent image handling
    return formatStorageUrl(url);
  }

  async function loadCommunityData() {
    try {
      isLoading = true;
      errorMessage = "";

      // Get community details
      console.log(`Calling getCommunityById for community: ${communityId}`);
      let communityResponse;

      try {
        // Try the main community API
        communityResponse = await getCommunityById(communityId);
        console.log("Community response from API:", communityResponse);

        if (!communityResponse || (!communityResponse.community && !communityResponse.data)) {
          throw new Error("Invalid API response format");
        }
      } catch (err) {
        console.warn("Primary API call failed:", err);
        errorMessage = "Community not found";
        throw new Error("Community not found or inaccessible");
      }

      // Extract community data handling different response formats
      let communityData;
      if (communityResponse.community) {
        communityData = communityResponse.community;
      } else if (communityResponse.data) {
        communityData = communityResponse.data;
      } else {
        errorMessage = "Invalid community data format";
        throw new Error("Invalid community data format");
      }

      // Normalize community data fields
      community = {
        id: communityData.id || communityId,
        name: communityData.name || "Unnamed Community",
        description: communityData.description || "",
        logo_url: communityData.logo_url || communityData.logo || "",
        banner_url: communityData.banner_url || communityData.banner || "",
        creator_id: communityData.creator_id || communityData.creatorId || "",
        is_approved: communityData.is_approved != null ? communityData.is_approved : (communityData.isApproved || false),
        is_private: communityData.is_private || communityData.isPrivate || false,
        categories: communityData.categories || [],
        created_at: communityData.created_at || communityData.createdAt || new Date(),
        member_count: communityData.member_count || communityData.memberCount || 0
      };

      console.log("Normalized community data:", community);

      // Check membership status
      try {        // Only check membership if user is logged in
        const authData = localStorage.getItem("auth");
        console.log("🔍 DEBUG: Auth data from localStorage:", authData ? "present" : "missing");
        if (authData) {
          const auth = JSON.parse(authData);
          console.log("🔍 DEBUG: Parsed auth data:", {
            hasToken: !!auth.access_token,
            expiresAt: auth.expires_at,
            isExpired: auth.expires_at ? new Date(auth.expires_at) <= new Date() : "no expiry set"
          });
          if (auth.access_token && (!auth.expires_at || new Date(auth.expires_at) > new Date())) {
            console.log("🔍 DEBUG: Calling checkUserCommunityMembership for community:", communityId);
            const membershipResponse = await checkUserCommunityMembership(communityId);
            console.log("🔍 DEBUG: Raw membership response:", membershipResponse);

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
            if (typedResponse?.status === "member" ||
                typedResponse?.is_member === true ||
                typedResponse?.data?.is_member === true ||
                typedResponse?.data?.status === "member") {
              isMember = true;
              console.log("✅ DEBUG: User is a member of this community");

              // Get the user's role in the community
              userRole = typedResponse?.user_role ||
                         typedResponse?.data?.user_role ||
                         "member";

              console.log(`✅ DEBUG: User has role "${userRole}" in this community`);
            } else if (typedResponse?.status === "pending" ||
                       typedResponse?.data?.status === "pending") {
              isPending = true;
              console.log("⏳ DEBUG: User has a pending join request for this community");
            } else {
              isMember = false;
              isPending = false;
              console.log("❌ DEBUG: User is not a member of this community");
              console.log("❌ DEBUG: Response details for debugging:", JSON.stringify(typedResponse, null, 2));
            }          }
        } else {
          console.log("🔍 DEBUG: User not logged in, skipping membership check");
          isMember = false;
          isPending = false;
        }
      } catch (membershipError) {
        console.error("❌ DEBUG: Error checking membership status:", membershipError);
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
          loadPendingMembers(),
          loadTopThreads(),
          loadLatestThreads(),
          loadMediaThreads(),
          loadTopMembers()
        ]);
      } catch (loadError) {
        console.warn("Error loading related community data:", loadError);
      }

    } catch (error) {
      logger.error("Error loading community data:", error);
      errorMessage = error instanceof Error ? error.message : "Unknown error";

      if (!community) {
        if (error instanceof SyntaxError && error.message.includes("Unexpected end of JSON")) {
          toastStore.showToast("Unable to load community data. The server returned an invalid response.", "error");
        } else if (error instanceof Error && error.message.includes("Empty response from server")) {
          toastStore.showToast("Unable to load community data. The server returned an empty response.", "error");
        } else if (error instanceof Error && error.message.includes("Community not found")) {
          toastStore.showToast("The community you are looking for does not exist or has been removed.", "error");
        } else {
          toastStore.showToast("Failed to load community data", "error");
        }
      }
    } finally {
      isLoading = false;
    }
  }  async function loadThreads() {
    try {
      // For now, use getAllThreads to avoid the 400 error
      // In a real implementation, you'd want a community-specific threads endpoint
      const threadsResponse = await getAllThreads(1, 10);

      if (threadsResponse && Array.isArray(threadsResponse.threads)) {
        // Filter threads by community_id if the field exists, otherwise show all threads
        threads = threadsResponse.threads.filter(thread => 
          !thread.community_id || thread.community_id === communityId
        ) as Thread[];
      } else if (threadsResponse && threadsResponse.data && Array.isArray(threadsResponse.data.threads)) {
        threads = threadsResponse.data.threads.filter(thread => 
          !thread.community_id || thread.community_id === communityId
        ) as Thread[];
      } else {
        threads = [];
      }

      // If no community-specific threads found, show some general threads for demo purposes
      if (threads.length === 0 && threadsResponse && Array.isArray(threadsResponse.threads)) {
        threads = threadsResponse.threads.slice(0, 5) as Thread[];
      }

    } catch (error) {
      logger.error("Error loading community threads:", error);
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
      logger.error("Error loading community members:", error);
      members = [];
    }
  }
  // Process member avatars to use Supabase URLs
  function processMembersAvatars(membersList) {
    return membersList.map((member, index) => {
      const processedMember = { ...member };

      // Ensure unique ID - use user_id as primary, fallback to index-based ID
      if (!processedMember.id && processedMember.user_id) {
        processedMember.id = processedMember.user_id;
      } else if (!processedMember.id) {
        processedMember.id = `member-${index}-${Date.now()}`;
      }

      // Handle different avatar field names
      const avatarUrl = member.avatar_url || member.profile_picture_url || member.avatar || "";

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
        processedMember.name = processedMember.username || "Unknown User";
      }

      console.log("Processed member data:", {
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
      console.log("Raw pending join requests response:", JSON.stringify(pendingResponse, null, 2));

      if (pendingResponse && Array.isArray(pendingResponse.join_requests)) {
        console.log(`Found ${pendingResponse.join_requests.length} pending join requests`);

        // Debug log to see the structure of the first join request (if available)
        if (pendingResponse.join_requests.length > 0) {
          const sampleRequest = pendingResponse.join_requests[0];
          console.log("Example join request structure:", sampleRequest);
          console.log("All fields in join request:", Object.keys(sampleRequest));
          console.log("Join request ID:", sampleRequest.id);
          console.log("Join request user_id:", sampleRequest.user_id);
          console.log("Join request username field value:", sampleRequest.username);
          console.log("Join request name field value:", sampleRequest.name || sampleRequest.display_name);

          // If the join request contains a user object, log its structure too
          if (sampleRequest.user) {
            console.log("User object in join request:", sampleRequest.user);
            console.log("User object fields:", Object.keys(sampleRequest.user));
            console.log("User object username:", sampleRequest.user.username);
          }
        }          // Format users from join requests to match Member structure
        pendingMembers = pendingResponse.join_requests.map((request, index) => {
          console.log(`Processing request for user_id: ${request.user_id}, found username: ${request.username || "MISSING"}`);

          // The backend now returns real user data, so prioritize that
          const member = {
            id: request.id || request.user_id || `pending-${index}-${Date.now()}`,
            user_id: request.user_id || "",
            // Prioritize real username from backend, fallback only if not available
            username: request.username || `user_${(request.user_id || "").substring(0, 8)}`,
            name: request.name || request.username || `User ${(request.user_id || "").substring(0, 8)}`,
            role: "pending",
            avatar_url: request.avatar_url || request.profile_picture_url || "",
            requested_at: request.created_at || new Date()
          };

          console.log("Created member object:", member);
          return member;
        });

        // Process avatars for pending members
        pendingMembers = processMembersAvatars(pendingMembers);
      } else if (pendingResponse && pendingResponse.data && Array.isArray(pendingResponse.data.join_requests)) {
        // Similar debugging for alternative response format
        console.log(`Found ${pendingResponse.data.join_requests.length} pending join requests (alt format)`);

        if (pendingResponse.data.join_requests.length > 0) {
          const sampleRequest = pendingResponse.data.join_requests[0];
          console.log("Example join request structure (alt format):", sampleRequest);
          console.log("All fields in join request (alt format):", Object.keys(sampleRequest));
          console.log("Join request ID (alt):", sampleRequest.id);
          console.log("Join request user_id (alt):", sampleRequest.user_id);
          console.log("Join request username field value (alt):", sampleRequest.username);
        }          // Format users from join requests (alternative response format)
        pendingMembers = pendingResponse.data.join_requests.map((request, index) => {
          console.log(`Processing alt request for user_id: ${request.user_id}, found username: ${request.username || "MISSING"}`);

          // The backend now returns real user data, so prioritize that
          const member = {
            id: request.id || request.user_id || `pending-alt-${index}-${Date.now()}`,
            user_id: request.user_id || "",
            // Prioritize real username from backend, fallback only if not available
            username: request.username || `user_${(request.user_id || "").substring(0, 8)}`,
            name: request.name || request.username || `User ${(request.user_id || "").substring(0, 8)}`,
            role: "pending",
            avatar_url: request.avatar_url || request.profile_picture_url || "",
            requested_at: request.created_at || new Date()
          };

          console.log("Created member object (alt):", member);
          return member;
        });

        // Process avatars for pending members
        pendingMembers = processMembersAvatars(pendingMembers);
      } else {
        console.log("No valid join requests found in the response");
        pendingMembers = [];
      }

      console.log("Final processed pending members:", pendingMembers);
    } catch (error) {
      logger.error("Error loading pending join requests:", error);
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
      logger.error("Error loading community rules:", error);
      rules = [];
    }
  }

  async function handleJoinRequest() {
    if (!authState.is_authenticated) {
      toastStore.showToast("You need to log in to join communities", "warning");
      return;
    }

    try {
      await requestToJoin(communityId, {});
      isPending = true;
      toastStore.showToast("Join request sent successfully", "success");
    } catch (error) {
      logger.error("Error joining community:", error);
      toastStore.showToast("Failed to join community. Please try again.", "error");
    }
  }

  // Update the tabItems array with the new tabs - make it reactive to ensure it's evaluated after authState
  $: {
    console.log("Re-evaluating tabItems with authState:",
      authState ? { isAuthenticated: authState.is_authenticated, userId: authState.user_id } : "undefined",
      "Community creator:", community?.creator_id);
  }

  $: tabItems = [
    { id: "top", label: "Top", icon: TrendingUpIcon },
    { id: "latest", label: "Latest", icon: ClockIcon },
    { id: "media", label: "Media", icon: ImageIcon },
    { id: "about", label: "About", icon: InfoIcon },
    {
      id: "manage",
      label: "Manage Members",
      icon: UserCheckIcon,
      condition: () => {
        const isCreator = community?.creator_id === authState?.user_id;
        const hasRole = ["owner", "admin", "moderator"].includes(userRole);
        console.log("Manage tab condition:", { isCreator, hasRole, userRole, canManage: canManageCommunity() });
        return authState?.is_authenticated && (isCreator || hasRole);
      }
    }
  ].filter(tab => !tab.condition || tab.condition());

  // Function to handle join/leave community toggle
  function toggleJoinCommunity() {
    if (isMember) {
      // Use removeMember function from API instead of leaveCommunity
      removeMember(communityId, authState.user_id || "")
        .then(() => {
          isMember = false;
          toastStore.showToast("Left community successfully", "success");
        })
        .catch(error => {
          logger.error("Error leaving community:", error);
          toastStore.showToast("Failed to leave community", "error");
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
      console.error("Invalid tweet data for navigation", tweet);
      return;
    }

    const threadId = tweet.id;
    console.log(`Navigating to thread detail: ${threadId}`);

    // Construct the URL for thread detail
    const href = `/thread/${threadId}`;

    // Use navigation approach
    try {
      // First try to use history API for SPA navigation
      window.history.pushState({threadId}, "", href);

      // Dispatch a custom navigation event to trigger router update
      const navEvent = new CustomEvent("navigate", {
        detail: { href, threadId }
      });
      window.dispatchEvent(navEvent);

      // Trigger popstate as a fallback
      window.dispatchEvent(new PopStateEvent("popstate", {}));

      // If nothing works, reload the page after a short delay
      setTimeout(() => {
        if (window.location.pathname !== href) {
          console.warn("Navigation did not update the URL, forcing page reload");
          window.location.href = href;
        }
      }, 300);
    } catch (error) {
      console.error("Error in navigation:", error);
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
      toastStore.showToast("This community is pending approval. You cannot create posts yet.", "warning");
      return;
    }

    // Navigate to create post page with community context
    const href = `/create-post?community=${communityId}`;
    window.location.href = href;
  }

  async function handleApproveJoinRequest(requestId: string) {
    try {
      if (!authState.is_authenticated) {
        toastStore.showToast("You need to log in to approve join requests", "warning");
        return;
      }

      // Call the API to approve the join request
      await approveJoinRequest(communityId, requestId);
      toastStore.showToast("Join request approved successfully", "success");

      // Reload members and pending members
      await Promise.all([
        loadMembers(),
        loadPendingMembers()
      ]);
    } catch (error) {
      logger.error("Error approving join request:", error);
      toastStore.showToast("Failed to approve join request. Please try again.", "error");
    }
  }

  async function handleRejectJoinRequest(requestId: string) {
    try {
      if (!authState.is_authenticated) {
        toastStore.showToast("You need to log in to reject join requests", "warning");
        return;
      }

      // Call the API to reject the join request
      await rejectJoinRequest(communityId, requestId);
      toastStore.showToast("Join request rejected", "success");

      // Reload pending members
      await loadPendingMembers();
    } catch (error) {
      logger.error("Error rejecting join request:", error);
      toastStore.showToast("Failed to reject join request. Please try again.", "error");
    }
  }

  // Check if user can manage the community (approve/reject join requests, etc.)
  function canManageCommunity(): boolean {
    // Safety check - if authState is undefined, return false
    if (!authState) return false;

    // Debug the owner check
    console.log("DEBUG canManageCommunity:", {
      userRole,
      authUserId: authState.user_id,
      communityCreatorId: community?.creator_id,
      isAuthUserCreator: community?.creator_id === authState.user_id
    });

    // User must be logged in and have appropriate role
    // Ownership is determined either by userRole or by being the creator
    const isOwner = userRole === "owner" ||
                    (community?.creator_id && community.creator_id === authState.user_id);
    const isAdmin = userRole === "admin";
    const isModerator = userRole === "moderator";

    const canManage = authState.is_authenticated && (isOwner || isAdmin || isModerator);
    console.log("Can manage community:", canManage);
    return canManage;
  }

  // Handle kicking a member from the community
  async function handleKickMember(userId: string) {
    if (!communityId || !authState.user_id) return;

    try {
      // Call the API to remove the member
      await removeMember(communityId, userId);
      toastStore.showToast("Member removed from community", "success");

      // Refresh the members list
      await loadMembers();
    } catch (error) {
      logger.error("Error kicking member:", error);
      toastStore.showToast("Failed to remove member. Please try again.", "error");
    }
  }

  async function loadTopThreads() {
    try {
      isLoadingTopThreads = true;

      // First try to fetch from backend API for top threads
      try {
        // Here we would call a specific API endpoint for top threads if available
        // For example: const response = await fetch(`${API_BASE_URL}/communities/${communityId}/top-threads`);

        // For now, we'll use the existing threads and sort them by likes
        console.log("Sorting threads by likes to get top threads");
        topThreads = [...threads].sort((a, b) => (b.likes || 0) - (a.likes || 0)).slice(0, 10);
      } catch (apiError) {
        console.warn("Error fetching top threads from API, falling back to client-side sorting", apiError);
        topThreads = [...threads].sort((a, b) => (b.likes || 0) - (a.likes || 0)).slice(0, 10);
      }

      isLoadingTopThreads = false;
    } catch (error) {
      logger.error("Error loading top threads:", error);
      isLoadingTopThreads = false;
    }
  }

  async function loadLatestThreads() {
    try {
      isLoadingLatestThreads = true;
      // Here you would call an API to get latest threads
      // For now, we'll just sort the existing threads by date
      latestThreads = [...threads].sort((a, b) => {
        const dateA = a.timestamp ? new Date(a.timestamp).getTime() : 0;
        const dateB = b.timestamp ? new Date(b.timestamp).getTime() : 0;
        return dateB - dateA;
      });
      isLoadingLatestThreads = false;
    } catch (error) {
      logger.error("Error loading latest threads:", error);
      isLoadingLatestThreads = false;
    }
  }  async function loadMediaThreads() {
    try {
      isLoadingMediaThreads = true;
      
      // Reset for initial load
      mediaPage = 1;
      hasMoreMedia = true;
        // Generate mock media content for demonstration
      const mockMediaItems: Thread[] = Array.from({ length: 6 }, (_, index) => ({
        id: `media-${index + 1}`,
        content: `Mock media post ${index + 1} with some interesting content about the community. This could be a photo from a recent event, artwork shared by members, or any visual content related to our community discussions.`,
        timestamp: new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000), // Random date within last 30 days
        username: `user${Math.floor(Math.random() * 100)}`,
        display_name: `Community User ${index + 1}`,
        avatar: `https://api.dicebear.com/7.x/avataaars/svg?seed=${index}`,
        likes: Math.floor(Math.random() * 100),
        comments: Math.floor(Math.random() * 50),
        reposts: Math.floor(Math.random() * 25),
        views: Math.floor(Math.random() * 500),
        media_url: `https://picsum.photos/400/300?random=${index + 1}`,
        media_type: (Math.random() > 0.7 ? 'video' : 'image') as 'image' | 'video'
      }));

      mediaThreads = mockMediaItems;
      isLoadingMediaThreads = false;
    } catch (error) {
      logger.error("Error loading media threads:", error);
      isLoadingMediaThreads = false;
    }
  }
  async function loadMoreMedia() {
    if (isLoadingMoreMedia || !hasMoreMedia) return;

    try {
      isLoadingMoreMedia = true;
      mediaPage += 1;      
      
      // Generate more mock media content
      const moreMediaItems: Thread[] = Array.from({ length: 6 }, (_, index) => {
        const globalIndex = (mediaPage - 1) * 6 + index + 1;
        return {
          id: `media-${globalIndex}`,
          content: `Mock media post ${globalIndex} with some interesting content about the community. This could be a photo from a recent event, artwork shared by members, or any visual content related to our community discussions.`,
          timestamp: new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000),
          username: `user${Math.floor(Math.random() * 100)}`,
          display_name: `Community User ${globalIndex}`,
          avatar: `https://api.dicebear.com/7.x/avataaars/svg?seed=${globalIndex}`,
          likes: Math.floor(Math.random() * 100),
          comments: Math.floor(Math.random() * 50),
          reposts: Math.floor(Math.random() * 25),
          views: Math.floor(Math.random() * 500),
          media_url: `https://picsum.photos/400/300?random=${globalIndex}`,
          media_type: (Math.random() > 0.7 ? 'video' : 'image') as 'image' | 'video'
        };
      });

      mediaThreads = [...mediaThreads, ...moreMediaItems];

      // Reset page counter after reaching 10 pages to simulate cycling/infinite scroll
      if (mediaPage >= 10) {
        mediaPage = 0; // Reset to create cycling effect
      }

      isLoadingMoreMedia = false;
    } catch (error) {
      logger.error("Error loading more media:", error);
      isLoadingMoreMedia = false;
    }
  }

  // Scroll handler for infinite scroll
  function handleMediaScroll(event: Event) {
    const target = event.target as Element;
    if (target.scrollTop + target.clientHeight >= target.scrollHeight - 100) {
      loadMoreMedia();
    }
  }

  async function loadTopMembers() {
    try {
      isLoadingTopMembers = true;

      // First try to fetch from backend API for top members
      try {
        // Here we would call a specific API endpoint for top members if available
        // For example: const response = await fetch(`${API_BASE_URL}/communities/${communityId}/top-members`);

        // For now, we'll check if we have member data and sort it by a proxy for popularity
        // In a real application, you would have a dedicated API endpoint for this
        if (members && members.length > 0) {
          console.log("Selecting top members from available members");

          // Since we don't have follower count in our current data model,
          // we'll just take the first 3 members, prioritizing admins and moderators
          const sortedMembers = [...members].sort((a, b) => {
            // First prioritize by role (admin > moderator > member)
            const roleValueMap = { owner: 3, admin: 2, moderator: 1, member: 0 };
            const roleValueA = roleValueMap[a.role] || 0;
            const roleValueB = roleValueMap[b.role] || 0;
            return roleValueB - roleValueA;
          });

          topMembers = sortedMembers.slice(0, 3);
        } else {
          topMembers = [];
        }
      } catch (apiError) {
        console.warn("Error fetching top members from API, falling back to first 3 members", apiError);
        topMembers = members.slice(0, 3);
      }

      isLoadingTopMembers = false;
    } catch (error) {
      logger.error("Error loading top members:", error);
      isLoadingTopMembers = false;
    }
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
        <div class="community-banner" style={community.banner_url ? `background-image: url(${getImageUrl(community.banner_url, "banner")})` : ""}>
          <div class="community-info-overlay">
            <div class="community-logo-container">
              {#if community.logo_url}
                <img src={getImageUrl(community.logo_url, "logo")} alt={community.name} class="community-logo" />
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
        {#if activeTab === "top"}
          <div class="top-content">
            <h2 class="section-title">Top Members</h2>
            <div class="top-members">
              {#if isLoadingTopMembers}
                <div class="loading-indicator"><Spinner size="medium" /></div>
              {:else if topMembers.length > 0}
                {#each topMembers as member}
                  <UserCard
                    user={{
                      id: member.user_id,
                      username: member.username,
                      name: member.name,
                      avatar_url: member.avatar_url,
                      role: member.role
                    }}
                    showFollowButton={false}
                  />
                {/each}
              {:else}
                <p class="empty-message">No members to display.</p>
              {/if}
            </div>

            <h2 class="section-title">Top Posts</h2>
            {#if isLoadingTopThreads}
              <div class="loading-indicator"><Spinner size="medium" /></div>
            {:else if topThreads.length > 0}
              <CommunityPosts
                threads={topThreads}
                {isMember}
                canPostInCommunity={canPostInCommunity()}
                communityIsApproved={community.is_approved}
                on:threadClick={handleThreadClick}
                on:createPost={handleCreatePost}
              />
            {:else}
              <p class="empty-message">No posts to display.</p>
            {/if}
          </div>

        {:else if activeTab === "latest"}
          <h2 class="section-title">Latest Posts</h2>
          {#if isLoadingLatestThreads}
            <div class="loading-indicator"><Spinner size="medium" /></div>
          {:else if latestThreads.length > 0}
            <CommunityPosts
              threads={latestThreads}
              {isMember}
              canPostInCommunity={canPostInCommunity()}
              communityIsApproved={community.is_approved}
              on:threadClick={handleThreadClick}
              on:createPost={handleCreatePost}
            />
          {:else}
            <p class="empty-message">No recent posts to display.</p>
          {/if}        {:else if activeTab === "media"}
          <h2 class="section-title">Media</h2>
          {#if isLoadingMediaThreads}
            <div class="loading-indicator"><Spinner size="medium" /></div>
          {:else if mediaThreads.length > 0}
            <div class="media-posts-container" on:scroll={handleMediaScroll}>
              {#each mediaThreads as mediaItem, index (mediaItem.id || `media-${index}`)}
                <div class="media-post-card">
                  <div class="media-post-header">
                    <div class="user-avatar">
                      <img src={mediaItem.avatar} alt={mediaItem.display_name} />
                    </div>
                    <div class="user-info">
                      <div class="user-name">{mediaItem.display_name}</div>
                      <div class="user-handle">@{mediaItem.username}</div>                      <div class="post-time">
                        {mediaItem.timestamp ? new Date(mediaItem.timestamp).toLocaleDateString() : 'Unknown date'}
                      </div>
                    </div>
                  </div>
                  
                  <div class="media-post-content">
                    <p>{mediaItem.content}</p>
                    <div class="media-container">
                      {#if mediaItem.media_type === 'video'}
                        <div class="video-placeholder">
                          <img src={mediaItem.media_url} alt="Video thumbnail" />
                          <div class="video-play-button">
                            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                              <circle cx="12" cy="12" r="12" fill="rgba(0,0,0,0.7)"/>
                              <polygon points="10,8 16,12 10,16" fill="white"/>
                            </svg>
                          </div>
                        </div>
                      {:else}
                        <img src={mediaItem.media_url} alt="Post media" class="media-image" />
                      {/if}
                    </div>
                  </div>
                  
                  <div class="media-post-actions">
                    <div class="action-button">
                      <MessageSquareIcon size="16" />
                      <span>{mediaItem.comments}</span>
                    </div>
                    <div class="action-button">
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <path d="M23 7L16 12L23 17V7Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                        <path d="M14 5L6 12L14 19V5Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                      </svg>
                      <span>{mediaItem.reposts}</span>
                    </div>
                    <div class="action-button">
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <path d="M20.84 4.61A5.5 5.5 0 0 0 16.5 2.03A5.44 5.44 0 0 0 12 4.17A5.44 5.44 0 0 0 7.5 2.03A5.5 5.5 0 0 0 3.16 4.61C1.8 5.95 1 7.78 1 9.72C1 13.91 8.5 20.5 12 22.39C15.5 20.5 23 13.91 23 9.72C23 7.78 22.2 5.95 20.84 4.61Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                      </svg>
                      <span>{mediaItem.likes}</span>
                    </div>
                    <div class="action-button">
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <circle cx="18" cy="5" r="3" stroke="currentColor" stroke-width="2"/>
                        <circle cx="6" cy="12" r="3" stroke="currentColor" stroke-width="2"/>
                        <circle cx="18" cy="19" r="3" stroke="currentColor" stroke-width="2"/>
                        <line x1="8.59" y1="13.51" x2="15.42" y2="17.49" stroke="currentColor" stroke-width="2"/>
                        <line x1="15.41" y1="6.51" x2="8.59" y2="10.49" stroke="currentColor" stroke-width="2"/>
                      </svg>
                      <span>{mediaItem.views}</span>
                    </div>
                  </div>
                </div>
              {/each}
              
              {#if isLoadingMoreMedia}
                <div class="loading-indicator">
                  <Spinner size="medium" />
                  <p>Loading more media...</p>
                </div>
              {/if}
              
              {#if hasMoreMedia && !isLoadingMoreMedia && mediaThreads.length > 0}
                <div class="load-more-container">
                  <Button variant="outlined" on:click={loadMoreMedia}>
                    Load More Media
                  </Button>
                </div>
              {/if}
              
              {#if !hasMoreMedia && mediaThreads.length > 0}
                <div class="end-of-content">
                  <p>You've reached the end of the media content!</p>
                </div>
              {/if}
            </div>
          {:else}
            <p class="empty-message">No media content to display.</p>
          {/if}

        {:else if activeTab === "about"}
          <CommunityAbout {community} />

        {:else if activeTab === "manage"}
          <h2 class="section-title">Manage Members</h2>
          <CommunityMembers
            {members}
            {pendingMembers}
            canManageCommunity={canManageCommunity()}
            currentUserId={authState.user_id || ""}
            on:approveJoinRequest={(e) => handleApproveJoinRequest(e.detail)}
            on:rejectJoinRequest={(e) => handleRejectJoinRequest(e.detail)}
            on:kickMember={(e) => handleKickMember(e.detail)}
          />
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

  .top-content {
    display: flex;
    flex-direction: column;
    gap: var(--space-6);
  }

  .section-title {
    font-size: var(--font-size-xl);
    font-weight: var(--font-weight-bold);
    margin-bottom: var(--space-4);
    padding-bottom: var(--space-2);
    border-bottom: 1px solid var(--border-color);
  }

  .top-members {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }  /* New media posts styles */
  .media-posts-container {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    max-height: 70vh;
    overflow-y: auto;
    padding-right: var(--space-2);
  }

  .media-post-card {
    background: var(--bg-primary);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-lg);
    padding: var(--space-4);
    transition: box-shadow 0.2s ease;
  }

  .media-post-card:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }

  :global(.dark) .media-post-card:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  }

  .media-post-header {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    margin-bottom: var(--space-3);
  }

  .media-post-header .user-avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    overflow: hidden;
  }

  .media-post-header .user-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .media-post-header .user-info {
    flex: 1;
  }

  .media-post-header .user-name {
    font-weight: var(--font-weight-bold);
    font-size: var(--font-size-sm);
    color: var(--text-primary);
  }

  .media-post-header .user-handle {
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
  }

  .media-post-header .post-time {
    font-size: var(--font-size-xs);
    color: var(--text-tertiary);
  }

  .media-post-content p {
    margin-bottom: var(--space-3);
    line-height: 1.5;
    color: var(--text-primary);
  }

  .media-container {
    position: relative;
    border-radius: var(--radius-md);
    overflow: hidden;
    margin-bottom: var(--space-3);
  }

  .media-image {
    width: 100%;
    height: auto;
    max-height: 400px;
    object-fit: cover;
    display: block;
  }

  .video-placeholder {
    position: relative;
    width: 100%;
  }

  .video-placeholder img {
    width: 100%;
    height: auto;
    max-height: 400px;
    object-fit: cover;
    display: block;
  }

  .video-play-button {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    cursor: pointer;
    transition: transform 0.2s ease;
  }

  .video-play-button:hover {
    transform: translate(-50%, -50%) scale(1.1);
  }

  .media-post-actions {
    display: flex;
    justify-content: space-around;
    padding-top: var(--space-2);
    border-top: 1px solid var(--border-color);
  }

  .action-button {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    padding: var(--space-2);
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: background-color 0.2s ease;
    color: var(--text-secondary);
    font-size: var(--font-size-sm);
  }

  .action-button:hover {
    background-color: var(--bg-secondary);
    color: var(--text-primary);
  }

  .loading-indicator {
    display: flex;
    justify-content: center;
    padding: var(--space-6);
  }

  .empty-message {
    text-align: center;
    color: var(--text-secondary);
    padding: var(--space-6);
  }  /* Responsive adjustments for the media posts */
  @media (max-width: 768px) {

    .media-post-card {
      padding: var(--space-3);
    }

    .media-post-header {
      gap: var(--space-2);
    }

    .media-post-actions {
      gap: var(--space-1);
    }

    .action-button {
      padding: var(--space-1);
      font-size: var(--font-size-xs);
    }
  }
  .action-button:hover {
    background-color: var(--bg-secondary);
    color: var(--text-primary);
  }

  .load-more-container {
    display: flex;
    justify-content: center;
    padding: var(--space-4);
  }

  .end-of-content {
    text-align: center;
    padding: var(--space-4);
    color: var(--text-secondary);
    font-style: italic;
  }

  .loading-indicator {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: var(--space-6);
    gap: var(--space-2);
  }

  .loading-indicator p {
    color: var(--text-secondary);
    font-size: var(--font-size-sm);
  }

  .empty-message {
    text-align: center;
    color: var(--text-secondary);
    padding: var(--space-6);
  }

  /* Responsive adjustments for the media posts */
  @media (max-width: 768px) {
    .media-post-card {
      padding: var(--space-3);
    }

    .media-post-header {
      gap: var(--space-2);
    }

    .media-post-actions {
      gap: var(--space-1);
    }

    .action-button {
      padding: var(--space-1);
      font-size: var(--font-size-xs);
    }

    .media-posts-container {
      max-height: 60vh;
    }
  }

  @media (max-width: 480px) {
    .media-post-header .user-avatar {
      width: 32px;
      height: 32px;
    }

    .media-post-actions {
      flex-wrap: wrap;
      gap: var(--space-1);
    }

    .action-button {
      flex: 1;
      justify-content: center;
      min-width: 0;
    }
  }
</style>
