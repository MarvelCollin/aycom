<script lang="ts">  import { createEventDispatcher, onMount, onDestroy } from "svelte";
  import { createLoggerWithPrefix } from "../../utils/logger";
  import { isAuthenticated as checkAuth, getUserId } from "../../utils/auth";
  import { useAuth } from "../../hooks/useAuth";
  import { likeThread, unlikeThread, replyToThread, getReplyReplies, likeReply, unlikeReply, bookmarkThread, removeBookmark, deleteThread } from "../../api";
  import { tweetInteractionStore } from "../../stores/tweetInteractionStore";
  import { notificationStore } from "../../stores/notificationStore";
  import { toastStore } from "../../stores/toastStore";
  import { formatStorageUrl } from "../../utils/common";
  import type { ITweet } from "../../interfaces/ISocialMedia";
  import type { IMedia } from "../../interfaces/IMedia";
  import Linkify from "../common/Linkify.svelte";
  import MessageCircleIcon from "svelte-feather-icons/src/icons/MessageCircleIcon.svelte";
  import RefreshCwIcon from "svelte-feather-icons/src/icons/RefreshCwIcon.svelte";
  import HeartIcon from "svelte-feather-icons/src/icons/HeartIcon.svelte";
  import BookmarkIcon from "svelte-feather-icons/src/icons/BookmarkIcon.svelte";
  import MoreHorizontalIcon from "svelte-feather-icons/src/icons/MoreHorizontalIcon.svelte";
  import ArrowUpIcon from "svelte-feather-icons/src/icons/ArrowUpIcon.svelte";
  import AlertTriangleIcon from "svelte-feather-icons/src/icons/AlertTriangleIcon.svelte";
  import TrashIcon from "svelte-feather-icons/src/icons/Trash2Icon.svelte";
  import UserIcon from "svelte-feather-icons/src/icons/UserIcon.svelte";
  import CornerUpRightIcon from "svelte-feather-icons/src/icons/CornerUpRightIcon.svelte";
  import EyeIcon from "svelte-feather-icons/src/icons/EyeIcon.svelte";
  import ChevronUpIcon from "svelte-feather-icons/src/icons/ChevronUpIcon.svelte";
  import ChevronDownIcon from "svelte-feather-icons/src/icons/ChevronDownIcon.svelte";
  import CheckCircleIcon from "svelte-feather-icons/src/icons/CheckCircleIcon.svelte";
  import { authStore } from "../../stores/authStore";
  import { useTheme } from "../../hooks/useTheme";
  import XIcon from "svelte-feather-icons/src/icons/XIcon.svelte";
  import UsersIcon from "svelte-feather-icons/src/icons/UsersIcon.svelte";

  interface ExtendedTweet extends ITweet {
    retweet_id?: string;
    threadId?: string;
    thread_id?: string;
    tweetId?: string;
    userId?: string;
    authorId?: string;
    author_id?: string;
    author_username?: string;
    authorUsername?: string;
    authorName?: string;
    avatar?: string;
    displayName?: string;
    display_name?: string;
    timestamp?: string;

    likes?: number;
    replies?: number;
    reposts?: number;
    bookmarks?: number;
    views?: number;

    isLiked?: boolean;
    isReposted?: boolean;
    isBookmarked?: boolean;
    isPinned?: boolean;
    is_verified?: boolean; 
    is_admin?: boolean; 

    bookmarked_thread?: any;
    bookmarked_reply?: any;
    parent_reply?: any;
    parent_thread?: any;
    parent_content?: string | null;
    parent_user?: {
      id?: string;
      username?: string;
      name?: string;
      profile_picture_url?: string;
    } | null;
    isComment?: boolean;

    user?: {
      id?: string;
      username?: string;
      name?: string;
      profile_picture_url?: string;
      is_verified?: boolean;
      verified?: boolean;
    };
    thread?: {
      author?: {
        id?: string;
        username?: string;
        name?: string;
      };
    };
    user_data?: {
      id?: string;
      username?: string;
      name?: string;
    };
    author?: {
      id?: string;
      username?: string;
      name?: string;
      profile_picture_url?: string;
    };
  }

  const logger = createLoggerWithPrefix("TweetCard");

  const { getAuthState, getAuthToken, refreshToken, checkAndRefreshTokenIfNeeded } = useAuth();

  export let isAuth: boolean = false;

  let authState: any = getAuthState();
  let isAuthenticated = Boolean(authState?.is_authenticated || isAuth);

  const updateAuthState = () => {
    authState = getAuthState();
    isAuthenticated = Boolean(authState?.is_authenticated || isAuth);
  };

  $: {

    isAuthenticated = Boolean(authState?.is_authenticated || isAuth);
  }

  export let tweet: ITweet | ExtendedTweet;
  export let isDarkMode: boolean = false;

  export const isLiked: boolean = false;
  export const isReposted: boolean = false;
  export const isBookmarked: boolean = false;

  export const inReplyToTweet: ITweet | null = null;
  export let replies: (ITweet | ExtendedTweet)[] = [];
  export let showReplies: boolean = false;
  export let nestingLevel: number = 0;
  const MAX_NESTING_LEVEL = 3;
  export let nestedRepliesMap: Map<string, (ITweet | ExtendedTweet)[]> = new Map();

  const dispatch = createEventDispatcher();

  $: processedTweet = processTweetContent(tweet);

  let storeInteraction: any = undefined;

  $: storeInteraction = $tweetInteractionStore?.get(processedTweet.id);

  $: effectiveIsLiked = storeInteraction?.is_liked ?? processedTweet.is_liked;
  $: effectiveIsReposted = storeInteraction?.is_reposted ?? processedTweet.is_reposted;
  $: effectiveIsBookmarked = storeInteraction?.is_bookmarked ?? processedTweet.is_bookmarked;

  $: effectiveLikes = storeInteraction?.likes ?? parseCount(processedTweet.likes_count);
  $: effectiveReplies = storeInteraction?.replies ?? parseCount(processedTweet.replies_count);
  $: effectiveReposts = storeInteraction?.reposts ?? parseCount(processedTweet.reposts_count);
  $: effectiveBookmarks = storeInteraction?.bookmarks ?? parseCount(processedTweet.bookmark_count);

  $: if (processedTweet.id) {
    console.debug(`Tweet ${processedTweet.id} counts:`, {
      likes: effectiveLikes,
      replies: effectiveReplies,
      reposts: effectiveReposts,
      bookmarks: effectiveBookmarks,
      fromStore: !!storeInteraction,
      rawCounts: {
        likes_count: processedTweet.likes_count,
        replies_count: processedTweet.replies_count,
        bookmark_count: processedTweet.bookmark_count
      }
    });
  }

  let isLikeLoading = false;
  const isRepostLoading = false;
  let isBookmarkLoading = false;
  const isShowingReplies = showReplies;

  const isLoadingReplies = false;
  let repliesErrorState = false;

  let isLoading = true;

  const isLoadingNestedReplies = new Map<string, boolean>();
  const nestedRepliesErrorState = new Map<string, boolean>();

  let replyActionsLoading = new Map<string, any>();

  let currentLikeRequestId = 0;
  const currentRepostRequestId = 0;
  const currentBookmarkRequestId = 0;

  $: hasReplies = effectiveReplies > 0 || parseCount(processedTweet.replies_count) > 0;
    $: processedReplies = replies.map(reply => processTweetContent(reply));
  $: tweetId = typeof processedTweet.id === "number" ? String(processedTweet.id) : processedTweet.id;

  $: isReply = Boolean(processedTweet.parent_id || processedTweet.parent_content || processedTweet.parent_user);

  onMount(() => {
    if (processedTweet) {
      tweetInteractionStore.initTweet(processedTweet);
    }

    setTimeout(() => {
      isLoading = false;
    }, 500);
  });

  $: if (processedTweet && processedTweet.id) {
    tweetInteractionStore.initTweet(processedTweet);
  }

  $: {
    if (tweetId) {
      tweetInteractionStore.subscribe(store => {
        storeInteraction = store.get(tweetId);
      });
    }
  }

  function isValidUsername(username: string | undefined | null): boolean {
    return !!username &&
           username !== "anonymous" &&
           username !== "user" &&
           username !== "unknown" &&
           username !== "undefined";
  }

  function isValidDisplayName(name: string | undefined | null): boolean {
    return !!name &&
           name !== "User" &&
           name !== "Anonymous User" &&
           name !== "Anonymous" &&
           name !== "undefined";
  }

  function processTweetContent(rawTweet: any): ExtendedTweet {
    try {
    if (!rawTweet) {
        console.error("Null or undefined tweet data passed to TweetCard");
      return createPlaceholderTweet();
    }

      console.debug("Processing tweet data:", {
        id: rawTweet.id,
        replies_count: rawTweet.replies_count,
        likes_count: rawTweet.likes_count,
        bookmark_count: rawTweet.bookmark_count,
        is_liked: rawTweet.is_liked,
        is_bookmarked: rawTweet.is_bookmarked
      });

      const processed: ExtendedTweet = {
        id: rawTweet.id || rawTweet.thread_id || "",
        content: rawTweet.content !== undefined ? rawTweet.content : "",
        created_at: rawTweet.created_at || new Date().toISOString(),
        updated_at: rawTweet.updated_at,

        user_id: rawTweet.user_id || "",
        username: rawTweet.username || "anonymous",
        name: rawTweet.name || "User",
        profile_picture_url: rawTweet.profile_picture_url ? formatStorageUrl(rawTweet.profile_picture_url) : "",

        likes_count: typeof rawTweet.likes_count === "number" ? rawTweet.likes_count : 0,
        replies_count: typeof rawTweet.replies_count === "number" ? rawTweet.replies_count : 0,
        reposts_count: typeof rawTweet.reposts_count === "number" ? rawTweet.reposts_count : 0,
        bookmark_count: typeof rawTweet.bookmark_count === "number" ? rawTweet.bookmark_count : 0,
        views_count: typeof rawTweet.views_count === "number" ? rawTweet.views_count : 0,

        is_liked: Boolean(rawTweet.is_liked),
        is_reposted: Boolean(rawTweet.is_reposted),
        is_bookmarked: Boolean(rawTweet.is_bookmarked),
        is_pinned: Boolean(rawTweet.is_pinned),
        is_verified: isVerified(rawTweet),

        community_id: rawTweet.community_id || null,
        community_name: rawTweet.community_name || null,

        parent_id: rawTweet.parent_id || null,

        media: Array.isArray(rawTweet.media) ? rawTweet.media.map(m => ({
          id: m.id || "",
          url: m.url ? formatStorageUrl(m.url) : "",
          type: m.type || "image",
          thumbnail_url: m.thumbnail_url ? formatStorageUrl(m.thumbnail_url) :
                       (m.url ? formatStorageUrl(m.url) : "")
        })) : [],

        timestamp: String(new Date(rawTweet.created_at || Date.now()).getTime())
      };

      return processed;
    } catch (error) {
      console.error("Error processing tweet content:", error, rawTweet);
      return createPlaceholderTweet();
    }
  }

  function createPlaceholderTweet(): ExtendedTweet {
    return {
      id: `error-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`,
      content: "Error loading content",
      created_at: new Date().toISOString(),
      user_id: "",
      username: "error",
      name: "Error Loading Data",
      profile_picture_url: "https://secure.gravatar.com/avatar/0?d=mp",
      likes_count: 0,
      replies_count: 0,
      reposts_count: 0,
      bookmark_count: 0,
      is_liked: false,
      is_reposted: false,
      is_bookmarked: false,
      is_pinned: false,
      is_verified: false,
      is_admin: false,
      parent_id: null,
      media: []
    };
  }

  function extractUserId(rawTweet: any): string {
    return rawTweet.user_id ||
      rawTweet.UserID ||
      rawTweet.userId ||
      rawTweet.author_id ||
      rawTweet.authorId ||
      rawTweet.user?.id ||
      rawTweet.author?.id ||
      rawTweet.thread?.author?.id ||
      rawTweet.user_data?.id ||
      "";
  }

  function extractUsername(rawTweet: any): string {
    return rawTweet.username ||
      rawTweet.Username ||
      rawTweet.author_username ||
      rawTweet.authorUsername ||
      rawTweet.user?.username ||
      rawTweet.author?.username ||
      rawTweet.thread?.author?.username ||
      rawTweet.user_data?.username ||
      "anonymous";
  }

  function extractDisplayName(rawTweet: any): string {
    return rawTweet.name ||
      rawTweet.DisplayName ||
      rawTweet.display_name ||
      rawTweet.authorName ||
      rawTweet.author_name ||
      rawTweet.user?.name ||
      rawTweet.user?.display_name ||
      rawTweet.author?.name ||
      rawTweet.thread?.author?.name ||
      rawTweet.user_data?.name ||
      "User";
  }

  function extractProfilePicture(rawTweet: any): string {
    const picUrl = rawTweet.profile_picture_url ||
      rawTweet.ProfilePicture ||
      rawTweet.author_avatar ||
      rawTweet.avatar ||
      rawTweet.user?.profile_picture_url ||
      rawTweet.author?.profile_picture_url ||
      rawTweet.user_data?.profile_picture_url;

    if (!picUrl) return "";

    return formatStorageUrl(picUrl);
  }

  function safeParseNumber(value: any): number {
    if (value === undefined || value === null) return 0;
    if (typeof value === "number") return value;
    if (typeof value === "string") {
      const parsed = parseInt(value, 10);
      return isNaN(parsed) ? 0 : parsed;
    }
    return 0;
  }

  function validateMedia(media: any): IMedia[] {
    console.log("Validating media:", media);

    if (!media) {
      console.warn("Media is null or undefined");
      return [];
    }

    if (!Array.isArray(media)) {
      console.warn("Media is not an array, attempting to parse:", typeof media);
      try {

        if (typeof media === "string") {
          const parsed = JSON.parse(media);
          if (Array.isArray(parsed)) {
            console.log("Successfully parsed media string into array:", parsed);
            return validateMedia(parsed); 
          }
        }
        console.warn("Could not convert media to array");
        return [];
      } catch (e) {
        console.error("Error parsing media string:", e);
        return [];
      }
    }

    const validatedMedia = media
      .filter(item => {
        if (!item) {
          console.warn("Filtered out null/undefined media item");
          return false;
        }
        if (!item.url) {
          console.warn("Filtered out media item with no URL:", item);
          return false;
        }
        return true;
      })
              .map(item => {
        try {
          const formattedUrl = formatStorageUrl(item.url);

          return {
            id: item.id || `media-${Date.now()}-${Math.random().toString(36).substring(2, 15)}`,
            url: formattedUrl,
            type: item.type || "image",
            thumbnail: item.thumbnail ? formatStorageUrl(item.thumbnail) : formattedUrl,
            alt_text: item.alt_text || item.alt || "Media attachment"
          };
        } catch (error) {
          console.error("Error formatting URL for media item:", error, item);
          return null;
        }
      })
      .filter(item => item !== null);

    console.log("Validated media result:", validatedMedia);
    return validatedMedia;
  }

  function formatTimeAgo(timestamp: string | undefined): string {
    if (!timestamp) return "";

    try {
      const date = new Date(timestamp);
      const now = new Date();
      const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

      if (isNaN(date.getTime())) {
        return "";
      }

      if (seconds < 60) {
        return "just now";
      }

      if (seconds < 3600) {
        const minutes = Math.floor(seconds / 60);
        return `${minutes}m`;
      }

      if (seconds < 86400) {
        const hours = Math.floor(seconds / 3600);
        return `${hours}h`;
      }

      if (seconds < 604800) {
        const days = Math.floor(seconds / 86400);
        return `${days}d`;
      }

      const options: Intl.DateTimeFormatOptions = { month: "short", day: "numeric" };

      if (date.getFullYear() !== now.getFullYear()) {
        options.year = "numeric";
      }

      return date.toLocaleDateString(undefined, options);
    } catch (e) {
      console.error("Error formatting date:", e);
      return "";
    }
  }

  function parseCount(value: any): number {
    const result = (() => {
      if (value === undefined || value === null) return 0;
      if (typeof value === "number") return value;
      if (typeof value === "string") {
        const parsed = parseInt(value, 10);
        return isNaN(parsed) ? 0 : parsed;
      }
      return 0;
    })();

    if (value !== undefined && value !== null && value !== 0) {
      console.debug(`parseCount: ${value} (${typeof value}) -> ${result}`);
    }

    return result;
  }

  function safeToString(value: any): string {
    if (value === undefined || value === null) return "";
    return String(value);
  }

  function handleReply() {
    if (!checkAuth()) {
      toastStore.showToast("Please log in to reply to posts", "info");
      return;
    }

    console.log(`Triggering reply for tweet: ${processedTweet.id}`);

    const replyBtn = document.querySelector(".tweet-reply-btn");
    if (replyBtn) {
      replyBtn.classList.add("clicked");
      setTimeout(() => {
        replyBtn.classList.remove("clicked");
      }, 300);
    }

    dispatch("reply", safeToString(processedTweet.id));
  }
  async function handleRetweet() {
    if (!checkAuth()) {
      toastStore.showToast("Please log in to repost", "info");
      return;
    }

    tweetInteractionStore.updateTweetInteraction(tweetId, {
      is_reposted: !effectiveIsReposted,
      reposts: effectiveReposts + (!effectiveIsReposted ? 1 : -1),
      pending_repost: true
    });
    dispatch("repost", tweetId);
  }

  async function handleLikeClick() {

    updateAuthState();

    if (!isAuthenticated) {
      toastStore.showToast("Please log in to like posts", "info");
      return;
    }

    if (isLikeLoading) return;
    isLikeLoading = true;
    repliesErrorState = false;

    const requestId = ++currentLikeRequestId;

    try {

      try {
        if (typeof checkAndRefreshTokenIfNeeded === "function") {
          await checkAndRefreshTokenIfNeeded();
        }
      } catch (refreshError) {
        console.error("Error refreshing token:", refreshError);
      }

      const currentLikeStatus = storeInteraction?.is_liked ?? processedTweet.is_liked ?? false;

      const newLikeStatus = !currentLikeStatus;
      const newLikeCount = newLikeStatus ? effectiveLikes + 1 : Math.max(0, effectiveLikes - 1);

      console.log(`${newLikeStatus ? "Liking" : "Unliking"} tweet ${tweetId}`);
      console.log(`Current UI like status: ${currentLikeStatus}, new status: ${newLikeStatus}`);
      console.log(`Current UI like count: ${effectiveLikes}, optimistic new count: ${newLikeCount}`);

      tweetInteractionStore.updateTweetInteraction(tweetId, {
        is_liked: newLikeStatus,
        likes: newLikeCount,
        pending_like: true 
      });

      heartAnimating = true;
      setTimeout(() => {
        heartAnimating = false;
      }, 800); 

      try {

        const apiCall = newLikeStatus ? likeThread : unlikeThread;
        const response = await apiCall(tweetId);

        if (requestId === currentLikeRequestId) {
          console.log('API response:', response);

          let finalLikeCount = newLikeCount;
          if (response && typeof response.likes_count === 'number') {
            finalLikeCount = response.likes_count;
            console.log(`Updated like count from API: ${finalLikeCount}`);
          }

          tweetInteractionStore.updateTweetInteraction(tweetId, {
            is_liked: newLikeStatus,
            likes: finalLikeCount,
            pending_like: false
          });

          console.log(`Successfully ${newLikeStatus ? "liked" : "unliked"} tweet ${tweetId} on server`);
        }
      } catch (error) {

        const errorMsg = error instanceof Error ? error.message.toLowerCase() : "";
        const isAlreadyInState =
          (newLikeStatus && (errorMsg.includes("already liked") || errorMsg.includes("already exists"))) ||
          (!newLikeStatus && (errorMsg.includes("not liked") || errorMsg.includes("not found")));

        if (errorMsg.includes("401") || errorMsg.includes("unauthorized") ||
            errorMsg.includes("session") || errorMsg.includes("expired")) {

          toastStore.showToast("Your session has expired. Please log in again", "info");

          setTimeout(() => {
            window.location.href = "/login";
          }, 2000);

          tweetInteractionStore.updateTweetInteraction(tweetId, {
            is_liked: currentLikeStatus,
            likes: currentLikeStatus ? newLikeCount + 1 : Math.max(0, newLikeCount - 1),
            pending_like: false
          });
        } else if (isAlreadyInState) {

          console.log(`Tweet ${tweetId} is already in the ${newLikeStatus ? "liked" : "unliked"} state on server`);

          tweetInteractionStore.updateTweetInteraction(tweetId, {
            is_liked: newLikeStatus,
            pending_like: false
          });
        } else {

          console.error(`Error ${newLikeStatus ? "liking" : "unliking"} tweet:`, error);

          tweetInteractionStore.updateTweetInteraction(tweetId, {
            is_liked: currentLikeStatus,
            likes: currentLikeStatus ? newLikeCount + 1 : Math.max(0, newLikeCount - 1),
            pending_like: false
          });

          toastStore.showToast(
            `Failed to ${newLikeStatus ? "like" : "unlike"} post. Please try again.`,
            "error",
            3000
          );
        }
      }
    } catch (error) {
      console.error("Error toggling like:", error);

      if (requestId === currentLikeRequestId) {

        toastStore.showToast("Could not update like status. Please try again.", "error", 3000);

        const revertToLiked = storeInteraction?.is_liked ?? processedTweet.is_liked ?? false;
        tweetInteractionStore.updateTweetInteraction(tweetId, {
          is_liked: revertToLiked,
          likes: revertToLiked ? effectiveLikes + 1 : Math.max(0, effectiveLikes - 1),
          pending_like: false
        });

        repliesErrorState = true;
      }
    } finally {
      if (requestId === currentLikeRequestId) {
        isLikeLoading = false;
      }
    }
  }

  async function toggleBookmarkStatus(event: Event) {
    event.stopPropagation();

    updateAuthState();

    if (!isAuthenticated) {
      toastStore.showToast("Please log in to bookmark tweets", "info");
      return;
    }

    if (isBookmarkLoading) return;
    isBookmarkLoading = true;

    const status = storeInteraction?.is_bookmarked || processedTweet.is_bookmarked || false;

    tweetInteractionStore.updateTweetInteraction(String(processedTweet.id), {
      is_bookmarked: !status,
      bookmarks: !status ? effectiveBookmarks + 1 : effectiveBookmarks - 1,
      pending_bookmark: true
    });

    try {

      try {
        if (typeof checkAndRefreshTokenIfNeeded === "function") {
          await checkAndRefreshTokenIfNeeded();
        }
      } catch (refreshError) {
        console.error("Error refreshing token:", refreshError);
      }

      let response;
      if (!status) {
        response = await bookmarkThread(processedTweet.id);
        toastStore.showToast("Tweet bookmarked", "success");
      } else {
        response = await removeBookmark(processedTweet.id);
        toastStore.showToast("Bookmark removed", "success");
      }

      let finalBookmarkCount = !status ? effectiveBookmarks + 1 : effectiveBookmarks - 1;
      if (response && typeof response.bookmark_count === 'number') {
        finalBookmarkCount = response.bookmark_count;
        console.log(`Updated bookmark count from API: ${finalBookmarkCount}`);
      }

      tweetInteractionStore.updateTweetInteraction(String(processedTweet.id), {
        is_bookmarked: !status,
        bookmarks: finalBookmarkCount,
        pending_bookmark: false
      });
    } catch (error) {
      console.error("Error toggling bookmark:", error);

      const errorMsg = error instanceof Error ? error.message.toLowerCase() : "";

      if (errorMsg.includes("already bookmarked") || errorMsg.includes("already exists")) {

        tweetInteractionStore.updateTweetInteraction(String(processedTweet.id), {
          is_bookmarked: !status, 
          pending_bookmark: false
        });
        return;
      }

      if (errorMsg.includes("401") || errorMsg.includes("unauthorized") ||
          errorMsg.includes("session") || errorMsg.includes("expired")) {

        toastStore.showToast("Your session has expired. Please log in again", "info");

        setTimeout(() => {
          window.location.href = "/login";
        }, 2000);
      } else {
        toastStore.showToast("Failed to update bookmark", "error");
      }

      tweetInteractionStore.updateTweetInteraction(String(processedTweet.id), {
        is_bookmarked: status,
        bookmarks: status ? effectiveBookmarks + 1 : effectiveBookmarks - 1,
        pending_bookmark: false
      });
    } finally {
      isBookmarkLoading = false;
    }
  }

  function toggleReplies() {
    showReplies = !showReplies;

    if (showReplies && (!replies || replies.length === 0)) {
      console.log("Loading replies for tweet:", processedTweet.id);

      const replyContainer = document.getElementById(`replies-container-${tweetId}`);
      if (replyContainer) {
        replyContainer.classList.add("loading-replies");
      }

      dispatch("loadReplies", safeToString(processedTweet.id));

      if (replies && replies.length > 0 && nestingLevel === 0) {
        console.log("DEBUG: Found replies to process:", replies.length);
        replies.forEach((reply, index) => {
          console.log(`DEBUG: Reply ${index} structure:`, {
            id: reply.id,
            content: reply.content || "(empty)",
            nested_replies: reply.replies_count || 0,

            user_data: ((reply as ExtendedTweet).user) ? {
              username: ((reply as ExtendedTweet).user)?.username || "no username"
            } : "no user data"
          });

          if (reply && reply.replies_count > 0) {
            try {

              const replyId = safeToString(reply.id);
              getReplyReplies(replyId).then(nestedRepliesData => {
                if (nestedRepliesData && nestedRepliesData.replies) {
                  console.log(`DEBUG: Loaded ${nestedRepliesData.replies.length} nested replies for ${reply.id}`);
                  nestedRepliesMap.set(replyId, nestedRepliesData.replies.map(r => processTweetContent(r)));
                  nestedRepliesMap = new Map(nestedRepliesMap);
                }
              }).catch(error => {
                console.error(`Error pre-loading nested replies for ${reply.id}:`, error);
              });
            } catch (error) {
              console.error(`Error pre-loading nested replies for ${reply.id}:`, error);
            }
          }
        });
      }
    } else {
      console.log("Hiding replies for tweet:", processedTweet.id);
    }
  }

  function handleShare() {
    dispatch("share", processedTweet);
  }

  function handleClick() {
    dispatch("click", processedTweet);
  }

  function handleNestedReply(event) {

    const replyId = typeof event.detail === "string" ? event.detail : String(event.detail);
    dispatch("reply", replyId);
  }

  async function handleLoadNestedReplies(event) {

    const replyId = typeof event.detail === "string" ? event.detail : String(event.detail);

    if (!replyId) {
      console.error("Missing reply ID in handleLoadNestedReplies");
      return;
    }

    console.log(`Loading nested replies for reply: ${replyId}`);

    const replyContainer = document.querySelector(`#reply-${replyId}-container`);
    if (replyContainer) {
      replyContainer.classList.add("loading-nested-replies");
    }

    const loadingKey = `loading_${replyId}`;

    const isLoading = nestedRepliesMap.get(loadingKey) &&
                     (nestedRepliesMap.get(loadingKey) as any)?.loading === true;

    if (!isLoading) {

      const loadingState = { loading: true };
      nestedRepliesMap.set(loadingKey, loadingState as any);
      nestedRepliesMap = new Map(nestedRepliesMap);

      try {

        const response = await getReplyReplies(replyId, 1, 20);

        if (response && response.replies && response.replies.length > 0) {
          console.log(`Received ${response.replies.length} nested replies for reply ${replyId}`);

          const processedReplies = response.replies.map(reply => {
            return processTweetContent(reply);
          });

          nestedRepliesMap.set(replyId, processedReplies);

          const countKey = `total_count_${replyId}`;
          const countObject = { count: response.total_count || processedReplies.length };
          nestedRepliesMap.set(countKey, countObject as any);
        } else {
          console.warn(`No nested replies returned for reply ${replyId}`);
          nestedRepliesMap.set(replyId, []);
        }

        nestedRepliesMap.delete(loadingKey);

        nestedRepliesMap = new Map(nestedRepliesMap);

        if (replyContainer) {
          replyContainer.classList.remove("loading-nested-replies");
        }
      } catch (error) {
        console.error(`Error loading nested replies for reply ${replyId}:`, error);

        const retryKey = `retry_${replyId}`;
        const retryObject = { retry: true };
        nestedRepliesMap.set(retryKey, retryObject as any);

        nestedRepliesMap = new Map(nestedRepliesMap);
      }
    }
  }

  async function retryLoadNestedReplies(replyId) {

    const retryKey = `retry_${replyId}`;
    nestedRepliesMap.delete(retryKey);

    handleLoadNestedReplies({ detail: replyId });
  }

  function handleNestedLike(event) {
    if (event.type === "unlike") {
      dispatch("unlike", event.detail);
    } else {
      dispatch("like", event.detail);
    }
  }

  function handleNestedBookmark(event) {
    if (event.type === "removeBookmark") {
      dispatch("removeBookmark", event.detail);
    } else {
      dispatch("bookmark", event.detail);
    }
  }

  function handleNestedRepost(event) {
    dispatch("repost", event.detail);
  }

  function handleMediaClick(mediaIndex: number) {
    dispatch("mediaClick", {
      tweet: processedTweet,
      mediaIndex: mediaIndex
    });
  }

  async function handleLikeReply(replyId: any) {
    try {
      if (!checkAuth()) {
        showLoginModal();
        return;
      }

      const reply = replies.find(r => r.id === replyId);
      if (!reply) return;

      const loadingState = replyActionsLoading.get(String(replyId)) || {};
      if (loadingState.like) return;

      loadingState.like = true;
      replyActionsLoading.set(String(replyId), loadingState);
      replyActionsLoading = new Map(replyActionsLoading);

      replyHeartAnimations.set(String(replyId), true);
      setTimeout(() => {
        replyHeartAnimations.delete(String(replyId));
      }, 800);

      reply.is_liked = true;
      if (typeof reply.likes_count === "number") {
        reply.likes_count += 1;
      }

      if (window.navigator && window.navigator.vibrate) {
        try {
          window.navigator.vibrate(30); 
        } catch (e) {

        }
      }

      if (navigator && navigator.onLine) {
        try {
          await likeReply(String(replyId));
        } catch (error) {
          console.error("Error liking reply:", error);

          const errorMsg = error instanceof Error ? error.message.toLowerCase() : "";
          const isAlreadyLiked = errorMsg.includes("already liked");
          if (!isAlreadyLiked) {

            reply.is_liked = false;
            if (typeof reply.likes_count === "number" && reply.likes_count > 0) {
              reply.likes_count -= 1;
            }
            toastStore.showToast("Failed to like reply. Please try again.", "error");
          }
        }
      } else {

        try {
          const offlineReplyLikes = JSON.parse(localStorage.getItem("offlineReplyLikes") || "{}");
          offlineReplyLikes[replyId] = { action: "like", timestamp: Date.now() };
          localStorage.setItem("offlineReplyLikes", JSON.stringify(offlineReplyLikes));
          toastStore.showToast("Liked! Will be synced when you're back online.", "info", 2000);
        } catch (e) {
          console.error("Failed to save offline reply like", e);
        }
      }
    } catch (error) {
      console.error("Unhandled error in handleLikeReply:", error);
      toastStore.showToast("An unexpected error occurred. Please try again later.", "error");
    } finally {

      const loadingState = replyActionsLoading.get(String(replyId)) || {};
      loadingState.like = false;
      replyActionsLoading.set(String(replyId), loadingState);
      replyActionsLoading = new Map(replyActionsLoading);
    }
  }

  async function handleUnlikeReply(replyId: any) {
    try {
      if (!checkAuth()) {
        showLoginModal();
        return;
      }

      const reply = replies.find(r => r.id === replyId);
      if (!reply) return;

      const loadingState = replyActionsLoading.get(String(replyId)) || {};
      if (loadingState.like) return;

      loadingState.like = true;
      replyActionsLoading.set(String(replyId), loadingState);
      replyActionsLoading = new Map(replyActionsLoading);      
      reply.is_liked = false;
      if (typeof reply.likes_count === "number" && reply.likes_count > 0) {
        reply.likes_count -= 1;
      }

      try {
        await unlikeReply(String(replyId));
      } catch (error) {
        console.error("Error unliking reply:", error);

        const errorMsg = error instanceof Error ? error.message.toLowerCase() : "";
        const isNotLiked = errorMsg.includes("not liked") || errorMsg.includes("not found");
          if (!isNotLiked) {

          reply.is_liked = true;
          if (typeof reply.likes_count === "number") {
            reply.likes_count += 1;
          }
          toastStore.showToast("Failed to unlike reply. Please try again.", "error");
        }
      } finally {

        loadingState.like = false;
        replyActionsLoading.set(String(replyId), loadingState);
        replyActionsLoading = new Map(replyActionsLoading);
      }
    } catch (error) {
      console.error("Unhandled error in handleUnlikeReply:", error);
      toastStore.showToast("An unexpected error occurred. Please try again later.", "error");
    }
  }

  async function handleBookmarkReply(replyId: any) {
    try {
      if (!checkAuth()) {
        showLoginModal();
        return;
      }

      const reply = replies.find(r => r.id === replyId);
      if (!reply) return;

      const loadingState = replyActionsLoading.get(String(replyId)) || {};
      if (loadingState.bookmark) return;

      loadingState.bookmark = true;
      replyActionsLoading.set(String(replyId), loadingState);
      replyActionsLoading = new Map(replyActionsLoading);

      reply.is_bookmarked = true;
      if (typeof reply.bookmark_count === "number") {
        reply.bookmark_count += 1;
      }

      try {

        await bookmarkThread(String(replyId));
        toastStore.showToast("Reply bookmarked", "success");
      } catch (error) {
        console.error("Error bookmarking reply:", error);

        const errorMsg = error instanceof Error ? error.message.toLowerCase() : "";
        const isAlreadyBookmarked = errorMsg.includes("already bookmarked") || errorMsg.includes("already exists");

        if (!isAlreadyBookmarked) {

          reply.is_bookmarked = false;
          if (typeof reply.bookmark_count === "number" && reply.bookmark_count > 0) {
            reply.bookmark_count -= 1;
          }
          toastStore.showToast("Failed to bookmark reply. Please try again.", "error");
        } else {

          toastStore.showToast("Reply is already bookmarked", "info");
        }
      } finally {

        loadingState.bookmark = false;
        replyActionsLoading.set(String(replyId), loadingState);
        replyActionsLoading = new Map(replyActionsLoading);
      }
    } catch (error) {
      console.error("Unhandled error in handleBookmarkReply:", error);
      toastStore.showToast("An unexpected error occurred. Please try again later.", "error");
    }
  }

  async function handleUnbookmarkReply(replyId: any) {
    try {
      if (!checkAuth()) {
        showLoginModal();
        return;
      }

      const reply = replies.find(r => r.id === replyId);
      if (!reply) return;

      const loadingState = replyActionsLoading.get(String(replyId)) || {};
      if (loadingState.bookmark) return;

      loadingState.bookmark = true;
      replyActionsLoading.set(String(replyId), loadingState);
      replyActionsLoading = new Map(replyActionsLoading);

      reply.is_bookmarked = false;
      if (typeof reply.bookmark_count === "number" && reply.bookmark_count > 0) {
        reply.bookmark_count -= 1;
      }

      try {

        await removeBookmark(String(replyId));
        toastStore.showToast("Bookmark removed", "success");
      } catch (error) {
        console.error("Error unbookmarking reply:", error);

        const errorMsg = error instanceof Error ? error.message.toLowerCase() : "";
        const isNotBookmarked = errorMsg.includes("not bookmarked") || errorMsg.includes("not found");

        if (!isNotBookmarked) {

          reply.is_bookmarked = true;
          if (typeof reply.bookmark_count === "number") {
            reply.bookmark_count += 1;
          }
          toastStore.showToast("Failed to remove bookmark from reply. Please try again.", "error");
        } else {

          toastStore.showToast("Reply was not bookmarked", "info");
        }
      } finally {

        loadingState.bookmark = false;
        replyActionsLoading.set(String(replyId), loadingState);
        replyActionsLoading = new Map(replyActionsLoading);
      }
    } catch (error) {
      console.error("Unhandled error in handleUnbookmarkReply:", error);
      toastStore.showToast("An unexpected error occurred. Please try again later.", "error");
    }
  }

  function navigateToUserProfile(event: MouseEvent | KeyboardEvent, username: string, userId?: string | number | null) {
    event.stopPropagation(); 

    const effectiveUserId = userId ? String(userId) :
      processedTweet.userId ? String(processedTweet.userId) :
      processedTweet.authorId ? String(processedTweet.authorId) :
      processedTweet.author_id ? String(processedTweet.author_id) :
      processedTweet.user_id ? String(processedTweet.user_id) : null;

    console.log("🔍 User Navigation Debug:", {
      username,
      providedUserId: userId,
      effectiveUserId,
      availableFields: {
        userId: processedTweet.userId,
        authorId: processedTweet.authorId,
        author_id: processedTweet.author_id,
        user_id: processedTweet.user_id,
        displayName: processedTweet.displayName,
      },
      tweetInfo: {
        id: processedTweet.id,
        content: processedTweet.content?.substring(0, 30) + "...",
      }
    });

    if (effectiveUserId) {
      console.log(`✅ Using userId for navigation: ${effectiveUserId}`);
      window.location.href = `/user/${effectiveUserId}`;
      return;
    }

    if (username && username !== "anonymous" && username !== "user" && username !== "unknown") {
      console.log(`✅ Falling back to username for navigation: ${username}`);
      window.location.href = `/user/${username}`;
    } else {
      console.error("❌ Navigation failed: No valid ID or username available", {
        username, providedUserId: userId
      });
    }
  }

  function debugUserData() {

    const extTweet = tweet as ExtendedTweet;
    console.log("Main tweet:", {
      id: tweet.id,
      thread_id: extTweet.thread_id || extTweet.threadId,
      username: {
        processed: processedTweet.username,
        original: tweet.username,
        authorUsername: extTweet.authorUsername,
        author_username: extTweet.author_username,
        fromThread: extTweet.thread?.author?.username,
        fromUser: extTweet.user?.username,
        fromAuthor: extTweet.author?.username,
        fromUserData: extTweet.user_data?.username,
        isValid: isValidUsername(processedTweet.username),
      },
      displayName: {
        processed: processedTweet.displayName,
        original: extTweet.displayName,
        authorName: extTweet.authorName,
        name: tweet.name,
        display_name: extTweet.display_name,
        fromThread: extTweet.thread?.author?.name,
        fromUser: extTweet.user?.name,
        fromAuthor: extTweet.author?.name,
        fromUserData: extTweet.user_data?.name,
        isValid: isValidDisplayName(processedTweet.displayName),
      },
      userId: {
        processed: processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id,
        original: extTweet.userId,
        authorId: extTweet.authorId,
        author_id: extTweet.author_id,
        user_id: tweet.user_id,
        fromThread: extTweet.thread?.author?.id,
        fromUser: extTweet.user?.id,
        fromAuthor: extTweet.author?.id,
        fromUserData: extTweet.user_data?.id,
      },

      bookmarkedThread: extTweet.bookmarked_thread || null,
      contenWithUsers: tweet.content && tweet.content.includes("@") ?
        tweet.content.match(/@([a-zA-Z0-9_]+)/g) : null
    });
      if (replies.length > 0) {
      const extReply = replies[0] as ExtendedTweet;
      console.log("First reply:", {
        id: replies[0].id,
        thread_id: extReply.thread_id || extReply.threadId,
        username: {
          original: replies[0].username,
          authorUsername: extReply.authorUsername,
          author_username: extReply.author_username,
          fromThread: extReply.thread?.author?.username,
          fromUser: extReply.user?.username,
          fromAuthor: extReply.author?.username,
          fromUserData: extReply.user_data?.username        },
        displayName: {
          original: extReply.displayName,
          authorName: extReply.authorName,
          name: replies[0].name,
          display_name: extReply.display_name,
          fromThread: extReply.thread?.author?.name,
          fromUser: extReply.user?.name,
          fromAuthor: extReply.author?.name,
          fromUserData: extReply.user_data?.name
        }
      });
    }
    console.groupEnd();
  }

  $: {
    if (tweet) {
      setTimeout(debugUserData, 500); 
    }
  }

  export function handleRefreshReplies(event) {
    const { threadId, parentReplyId, newReply } = event.detail;

    const processedNewReply = processTweetContent(newReply);

    if (threadId === processedTweet.id && !parentReplyId) {

      replies = [processedNewReply, ...replies];

      showReplies = true;
    }

    else if (parentReplyId) {

      const currentNestedReplies = nestedRepliesMap.get(parentReplyId) || [];

      nestedRepliesMap.set(parentReplyId, [processedNewReply, ...currentNestedReplies]);

      nestedRepliesMap = new Map(nestedRepliesMap);

      const parentIndex = replies.findIndex(r => r.id === parentReplyId);
      if (parentIndex >= 0) {

        const newReplies = [...replies];

        const oldParent = replies[parentIndex];
        const replyCount = typeof (oldParent as ExtendedTweet).replies === "number" ?
          (oldParent as ExtendedTweet).replies! + 1 :
          Number((oldParent as ExtendedTweet).replies || 0) + 1;

        const newParent = { ...oldParent };

        (newParent as any).replies = replyCount;

        newReplies[parentIndex] = newParent;

        replies = newReplies;
      }
    }
  }

  async function handleRepostClick(event: Event) {
    event.stopPropagation();

    if (!checkAuth()) {
      toastStore.showToast("Please log in to repost", "error");
      return;
    }

    const status = storeInteraction?.is_reposted || processedTweet.is_reposted || false;

    tweetInteractionStore.updateTweetInteraction(String(processedTweet.id), {
      is_reposted: !status,
      reposts: !status ? effectiveReposts + 1 : effectiveReposts - 1,
      pending_repost: true
    });

    try {
      if (!status) {
        dispatch("repost");
      } else {
        dispatch("unrepost");
      }

      tweetInteractionStore.updateTweetInteraction(String(processedTweet.id), {
        pending_repost: false
      });
    } catch (error) {
      console.error("Error toggling repost:", error);
      toastStore.showToast("Failed to update repost", "error");

      tweetInteractionStore.updateTweetInteraction(String(processedTweet.id), {
        is_reposted: status,
        reposts: status ? effectiveReposts + 1 : effectiveReposts - 1,
        pending_repost: false
      });
    }
  }

  let heartAnimating = false;
  const replyHeartAnimations = new Map<string, boolean>();

  onMount(() => {

    if (tweet) {
      tweetInteractionStore.initTweet(tweet);

      syncInteractionWithServer();
    }

    const checkAuthStatus = () => {
      updateAuthState();
      console.log("Authentication status:", isAuthenticated);
    };

    checkAuthStatus();
    const authCheckInterval = setInterval(checkAuthStatus, 60000); 

    window.addEventListener("auth:changed", checkAuthStatus);

    document.addEventListener("visibilitychange", handleVisibilityChange);

    return () => {
      clearInterval(authCheckInterval);
      window.removeEventListener("auth:changed", checkAuthStatus);
      document.removeEventListener("visibilitychange", handleVisibilityChange);
    };
  });

  onDestroy(() => {
    document.removeEventListener("visibilitychange", handleVisibilityChange);
    window.removeEventListener("auth:changed", () => {});
  });

  function handleVisibilityChange() {
    if (document.visibilityState === "visible") {
      syncInteractionWithServer();
      syncOfflineLikes();
    }
  }

  function syncInteractionWithServer() {
    if (navigator && navigator.onLine) {
      tweetInteractionStore.syncWithServer();
    }
  }

  function syncOfflineLikes() {
    if (navigator && navigator.onLine) {
      try {

        const offlineLikes = JSON.parse(localStorage.getItem("offlineLikes") || "{}");
        if (Object.keys(offlineLikes).length > 0) {
          console.log(`Found ${Object.keys(offlineLikes).length} offline likes to sync`);

          Object.entries(offlineLikes).forEach(async ([id, data]: [string, any]) => {
            try {
              const action = data.action;
              const apiCall = action === "like" ? likeThread : unlikeThread;
              await apiCall(id);

              tweetInteractionStore.updateTweetInteraction(id, {
                is_liked: action === "like",
                pending_like: false
              });

              delete offlineLikes[id];
            } catch (error) {
              console.error(`Failed to sync offline ${data.action} for tweet ${id}:`, error);
            }
          });

          localStorage.setItem("offlineLikes", JSON.stringify(offlineLikes));
        }

        const offlineReplyLikes = JSON.parse(localStorage.getItem("offlineReplyLikes") || "{}");
        if (Object.keys(offlineReplyLikes).length > 0) {
          console.log(`Found ${Object.keys(offlineReplyLikes).length} offline reply likes to sync`);

          Object.entries(offlineReplyLikes).forEach(async ([id, data]: [string, any]) => {
            try {
              const action = data.action;
              const apiCall = action === "like" ? likeReply : unlikeReply;
              await apiCall(id);

              delete offlineReplyLikes[id];
            } catch (error) {
              console.error(`Failed to sync offline ${data.action} for reply ${id}:`, error);
            }
          });

          localStorage.setItem("offlineReplyLikes", JSON.stringify(offlineReplyLikes));
        }
      } catch (e) {
        console.error("Error syncing offline likes:", e);
      }
    }
  }

  function isVerified(rawTweet: any): boolean {
    return Boolean(rawTweet.is_verified || rawTweet.verified || rawTweet.user?.is_verified || rawTweet.user?.verified || false);
  }

  function navigateToThreadDetail(e: Event) {
    e.preventDefault();
    e.stopPropagation();

    if (!processedTweet || !processedTweet.id) {
      console.error("Cannot navigate to thread detail: missing tweet ID");
      return;
    }

    const threadId = String(processedTweet.id);

    dispatch("click", processedTweet);

    const threadDataObj = {
      id: threadId,
      thread_id: threadId,
      user_id: String(processedTweet.user_id || processedTweet.userId || processedTweet.author_id || processedTweet.authorId || ""),
      username: String(processedTweet.username || processedTweet.author_username || processedTweet.authorUsername || "user"),
      name: String(processedTweet.name || processedTweet.displayName || processedTweet.display_name || "User"),
      profile_picture_url: processedTweet.profile_picture_url || "",
      content: processedTweet.content || "",
      created_at: processedTweet.created_at || new Date().toISOString(),
      updated_at: processedTweet.updated_at || null,

      likes_count: effectiveLikes,
      replies_count: effectiveReplies,
      reposts_count: effectiveReposts,
      bookmark_count: effectiveBookmarks,
      views_count: Number(processedTweet.views_count || 0),
      is_liked: effectiveIsLiked,
      is_bookmarked: effectiveIsBookmarked,
      is_reposted: effectiveIsReposted,
      is_verified: Boolean(processedTweet.is_verified),
      media: Array.isArray(processedTweet.media) ? processedTweet.media.map(item => ({
        id: String(item.id || ""),
        url: String(item.url || ""),
        type: String(item.type || "image"),
        thumbnail_url: String(item.thumbnail_url || item.url || "")
      })) : []
    };

    console.log("Storing thread data in sessionStorage:", threadDataObj);

    try {
      const threadData = JSON.stringify(threadDataObj);
      sessionStorage.setItem("lastViewedThread", threadData);
    } catch (error) {
      console.error("Failed to store thread data in sessionStorage:", error);
    }

    window.location.href = `/thread/${threadId}`;
  }

  $: isCurrentUserAuthor = authState?.user_id === processedTweet?.user_id;
  let isSettingsDropdownOpen = false;

  function toggleSettingsDropdown(e) {
    e.stopPropagation();
    isSettingsDropdownOpen = !isSettingsDropdownOpen;
  }

  function handleClickOutside(e) {
    if (isSettingsDropdownOpen) {
      isSettingsDropdownOpen = false;
    }
  }

  let showDeleteConfirmationModal = false;
  let tweetToDelete: string | null = null;

  async function handleDeleteTweet(e) {
    e.stopPropagation();

    if (!isCurrentUserAuthor) {
      toastStore.showToast("You can only delete your own posts", "error");
      return;
    }

    tweetToDelete = String(processedTweet.id);
    showDeleteConfirmationModal = true;
    isSettingsDropdownOpen = false; 
  }

  async function confirmDeleteTweet() {
    if (tweetToDelete === null) {
      console.error("No tweet ID to delete");
      return;
    }

    try {
      await deleteThread(tweetToDelete);
      toastStore.showToast("Post deleted successfully", "success");

      dispatch("deleted", { id: tweetToDelete });

      const tweetElement = document.getElementById(`tweet-${tweetToDelete}`);
      if (tweetElement) {
        tweetElement.style.height = `${tweetElement.offsetHeight}px`;
        tweetElement.style.overflow = "hidden";

        setTimeout(() => {
          tweetElement.style.height = "0";
          tweetElement.style.opacity = "0";
          tweetElement.style.margin = "0";
          tweetElement.style.padding = "0";
          tweetElement.style.transition = "all 0.3s ease";
        }, 10);

        setTimeout(() => {
          tweetElement.remove();
        }, 300);
      }

    } catch (error) {
      console.error("Error deleting tweet:", error);
      toastStore.showToast("Failed to delete post. Please try again.", "error");
    } finally {

      showDeleteConfirmationModal = false;
      tweetToDelete = null;
    }
  }

  function cancelDeleteTweet() {
    showDeleteConfirmationModal = false;
    tweetToDelete = null;
  }

  onMount(() => {
    document.addEventListener("click", handleClickOutside);
  });

  onDestroy(() => {
    document.removeEventListener("click", handleClickOutside);
  });

  function isAdminTweet(tweet: any): boolean {

    if (tweet.is_admin === true) return true;

    if (tweet.user && tweet.user.is_admin === true) return true;

    if (tweet.author && tweet.author.is_admin === true) return true;

    return false;
  }
</script>

<!-- Add skeleton loading in the template -->
<div class="tweet-card {isDarkMode ? "tweet-card-dark" : ""}" id="tweet-{processedTweet.id}">
  <!-- Skeleton loader -->
  {#if isLoading}
    <div class="tweet-card-skeleton">
      <div class="tweet-card-container">
        <div class="tweet-card-content">
          <div class="tweet-card-header">
            <div class="tweet-avatar skeleton"></div>
            <div class="tweet-content-container">
              <div class="tweet-author-info">
                <div class="tweet-author-name-skeleton skeleton"></div>
                <div class="tweet-author-username-skeleton skeleton"></div>
              </div>
              <div class="tweet-text">
                <div class="tweet-text-skeleton skeleton"></div>
                <div class="tweet-text-skeleton skeleton"></div>
              </div>
            </div>
          </div>

          <!-- Media skeleton -->
          <div class="tweet-media-skeleton skeleton"></div>

          <!-- Action buttons skeleton -->
          <div class="tweet-actions">
            <div class="tweet-action-item-skeleton skeleton"></div>
            <div class="tweet-action-item-skeleton skeleton"></div>
            <div class="tweet-action-item-skeleton skeleton"></div>
            <div class="tweet-action-item-skeleton skeleton"></div>
          </div>
        </div>
      </div>
    </div>
  {:else}
    <!-- Regular tweet card content -->
    <div
      class="tweet-card-container"
      on:click|preventDefault={navigateToThreadDetail}
      on:keydown={(e) => e.key === "Enter" && navigateToThreadDetail(e)}
      role="button"
      tabindex="0"
      aria-label="View thread details">
      <div class="tweet-card-content">
        <div class="tweet-card-header">
          <a href={`/user/${processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id || processedTweet.username}`}
            class="tweet-avatar"
            on:click|preventDefault={(e) => navigateToUserProfile(e, processedTweet.username, processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id)}
            on:keydown={(e) => e.key === "Enter" && navigateToUserProfile(e, processedTweet.username, processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id)}>
            {#if processedTweet.profile_picture_url}
              <img src={processedTweet.profile_picture_url} alt={processedTweet.username} class="tweet-avatar-image" />
            {:else}
              <div class="tweet-avatar-placeholder">
                <div class="tweet-avatar-text">{processedTweet.username ? processedTweet.username[0].toUpperCase() : "U"}</div>
              </div>
            {/if}
          </a>
          <div class="tweet-content-container">
            <div class="tweet-author-info">
              <a href={`/user/${processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id || processedTweet.username}`}
                class="tweet-author-name {isDarkMode ? "tweet-author-name-dark" : ""}"
                on:click|preventDefault={(e) => navigateToUserProfile(e, processedTweet.username, processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id)}
                on:keydown={(e) => e.key === "Enter" && navigateToUserProfile(e, processedTweet.username, processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id)}>
                {#if !processedTweet.name && !processedTweet.displayName}
                  {console.warn("❌ MISSING DISPLAY NAME:", {id: processedTweet.id, tweetId: processedTweet.tweetId, username: processedTweet.username})}
                  <span class="tweet-error-text">User</span>
                {:else}
                  <span class="display-name-text">{processedTweet.name || processedTweet.displayName}</span>
                  {#if processedTweet.is_verified}
                    <span class="user-verified-badge">
                      <CheckCircleIcon size="14" />
                    </span>
                  {/if}
                {/if}
              </a>
              <a href={`/user/${processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id || processedTweet.username}`}
                class="tweet-author-username {isDarkMode ? "tweet-author-username-dark" : ""}"
                on:click|preventDefault={(e) => navigateToUserProfile(e, processedTweet.username, processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id)}
                on:keydown={(e) => e.key === "Enter" && navigateToUserProfile(e, processedTweet.username, processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id)}>
                {#if processedTweet.username === "user"}
                  {console.warn("❌ MISSING USERNAME:", {id: processedTweet.id, tweetId: processedTweet.tweetId, displayName: processedTweet.displayName})}
                  <span class="tweet-error-text">@user</span>
                {:else}
                  @{processedTweet.username}
                {/if}
              </a>
              <span class="tweet-dot-separator {isDarkMode ? "tweet-dot-separator-dark" : ""}">·</span>
              <span class="tweet-timestamp {isDarkMode ? "tweet-timestamp-dark" : ""}">{formatTimeAgo(processedTweet.timestamp)}</span>

              <!-- Settings button for tweet owner -->
              {#if isCurrentUserAuthor}
                <div class="tweet-settings-container">
                  <button
                    class="tweet-settings-btn {isDarkMode ? "tweet-settings-btn-dark" : ""}"
                    on:click|stopPropagation={toggleSettingsDropdown}
                    aria-label="Tweet settings"
                  >
                    <MoreHorizontalIcon size="18" />
                  </button>

                  {#if isSettingsDropdownOpen}
                    <div class="tweet-settings-dropdown {isDarkMode ? "tweet-settings-dropdown-dark" : ""}">
                      <button
                        class="tweet-settings-option tweet-delete-option {isDarkMode ? "tweet-settings-option-dark" : ""}"
                        on:click|stopPropagation={handleDeleteTweet}
                      >
                        <TrashIcon size="16" />
                        <span>Delete</span>
                      </button>
                    </div>
                  {/if}
                </div>
              {/if}
            </div>

            <div class="tweet-text {isDarkMode ? "tweet-text-dark" : ""}">
              {#if processedTweet.is_advertisement || isAdminTweet(processedTweet)}
                <div class="tweet-ad-badge {isDarkMode ? "tweet-ad-badge-dark" : ""}">
                  <span class="ad-label">AD</span>
                </div>
                {#if isAdminTweet(processedTweet)}
                  <div class="tweet-admin-indicator {isDarkMode ? "tweet-admin-indicator-dark" : ""}">
                    <span>This is admin</span>
                  </div>
                {/if}
              {/if}

              <!-- Display parent information if this is a reply -->
              {#if isReply && processedTweet.parent_user}
                <div class="tweet-parent-info {isDarkMode ? "tweet-parent-info-dark" : ""}">
                  <div class="tweet-parent-indicator">
                    <CornerUpRightIcon size="14" />
                    <span>Replying to</span>
                    <a
                      href={`/user/${processedTweet.parent_user.username || processedTweet.parent_user.id}`}
                      class="tweet-parent-username"
                      on:click|stopPropagation
                    >
                      @{processedTweet.parent_user.username || "user"}
                    </a>
                  </div>
                  {#if processedTweet.parent_content}
                    <div class="tweet-parent-content">
                      <div class="tweet-content-text">
                        <Linkify text={processedTweet.parent_content} />
                      </div>
                    </div>
                  {/if}
                </div>
              {/if}

              {#if processedTweet.content}
                <div class="tweet-content-text">
                  <Linkify text={processedTweet.content} />
                </div>
              {:else}
                <p class="tweet-empty-content">{processedTweet.is_reposted ? "Reposted" : "This post has no content"}</p>
              {/if}
            </div>

            <!-- Community information -->
            {#if processedTweet.community_id && processedTweet.community_name}
              <div class="tweet-community-info {isDarkMode ? "tweet-community-info-dark" : ""}">
                <UsersIcon size="16" />
                <span class="tweet-community-name">Posted in {processedTweet.community_name}</span>
              </div>
            {/if}

            {#if processedTweet.media && processedTweet.media.length > 0}
              <div class="tweet-media-container {isDarkMode ? "tweet-media-container-dark" : ""}">
                {#if processedTweet.media.length === 1}
                  <!-- Single Media Display -->
                  <div class="tweet-media-single">
                    {#if processedTweet.media[0].type === "image"}
                      <button
                        class="tweet-media-btn"
                        on:click|stopPropagation={() => handleMediaClick(0)}
                        aria-label="Open media in overlay"
                      >
                        <img
                          src={processedTweet.media[0].url}
                          alt={processedTweet.media[0].alt_text || "Media"}
                          class="tweet-media-img"
                        />
                      </button>
                    {:else if processedTweet.media[0].type === "video"}
                      <video
                        src={processedTweet.media[0].url}
                        controls
                        class="tweet-media-video"
                        on:click|stopPropagation={() => handleMediaClick(0)}
                      >
                        <track kind="captions" src="/captions/en.vtt" srclang="en" label="English" />
                      </video>
                    {:else}
                      <button
                        class="tweet-media-btn"
                        on:click|stopPropagation={() => handleMediaClick(0)}
                        aria-label="Open media in overlay"
                      >
                        <img
                          src={processedTweet.media[0].url}
                          alt="GIF"
                          class="tweet-media-img"
                        />
                      </button>
                    {/if}
                  </div>
                {:else if processedTweet.media.length > 1}
                  <!-- Multiple Media Grid -->
                  <div class="tweet-media-grid">
                    {#each processedTweet.media.slice(0, 4) as media, index (media.url || index)}
                      <div class="tweet-media-item">
                        {#if media.type === "image"}
                          <button
                            class="tweet-media-btn"
                            on:click|stopPropagation={() => handleMediaClick(index)}
                            aria-label="Open media in overlay"
                          >
                            <img
                              src={media.url}
                              alt={media.alt_text || "Media"}
                              class="tweet-media-img"
                            />
                          </button>
                        {:else if media.type === "video"}
                          <video
                            src={media.url}
                            controls
                            class="tweet-media-video"
                            on:click|stopPropagation={() => handleMediaClick(index)}
                          >
                            <track kind="captions" src="/captions/en.vtt" srclang="en" label="English" />
                          </video>
                        {:else}
                          <button
                            class="tweet-media-btn"
                            on:click|stopPropagation={() => handleMediaClick(index)}
                            aria-label="Open media in overlay"
                          >
                            <img
                              src={media.url}
                              alt="GIF"
                              class="tweet-media-img"
                              on:error={(e) => {
                                console.error("GIF failed to load:", media.url);
                              }}
                            />
                          </button>
                        {/if}
                      </div>
                    {/each}
                  </div>
                {/if}
              </div>
            {/if}

            <div class="tweet-actions {isDarkMode ? "tweet-actions-dark" : ""}">
              <div class="tweet-action-item">
                <button
                  class="tweet-action-btn tweet-reply-btn {hasReplies ? "has-replies" : ""} {isDarkMode ? "tweet-action-btn-dark" : ""}"
                  on:click|stopPropagation={handleReply}
                  aria-label="Reply to tweet"
                >
                  <MessageCircleIcon size="20" class="tweet-action-icon" />
                  <span class="tweet-action-count {hasReplies ? "tweet-reply-count-highlight" : ""}">{effectiveReplies}</span>
                </button>
                {#if !showReplies}
                  <button
                    class="view-replies-btn"
                    on:click|stopPropagation={toggleReplies}
                    aria-label="View all replies"
                  >
                    {#if hasReplies}
                      View {effectiveReplies} {effectiveReplies === 1 ? "reply" : "replies"}
                    {:else}
                      View replies
                    {/if}
                  </button>
                {/if}
              </div>
              <div class="tweet-action-item">
                <button
                  class="tweet-action-btn tweet-repost-btn {effectiveIsReposted ? "active" : ""} {isDarkMode ? "tweet-action-btn-dark" : ""}"
                  on:click|stopPropagation={handleRepostClick}
                  aria-label={effectiveIsReposted ? "Undo repost" : "Repost"}
                >
                  <RefreshCwIcon size="20" class="tweet-action-icon" />
                  <span class="tweet-action-count">{effectiveReposts}</span>
                </button>
              </div>
              <div class="tweet-action-item">
                <button
                  class="tweet-action-btn tweet-like-btn {effectiveIsLiked ? "active" : ""} {isLikeLoading ? "loading" : ""} {heartAnimating ? "animating" : ""} {isDarkMode ? "tweet-action-btn-dark" : ""}"
                  on:click|stopPropagation={handleLikeClick}
                  aria-label={effectiveIsLiked ? "Unlike this post" : "Like this post"}
                  aria-pressed={effectiveIsLiked}
                  disabled={isLikeLoading}
                  data-testid="like-button"
                  aria-live="polite"
                  tabindex="0"
                >
                  <div class="tweet-like-icon-wrapper">
                    {#if isLikeLoading}
                      <div class="tweet-action-loading"></div>
                      <HeartIcon size="20" fill={effectiveIsLiked ? "currentColor" : "none"} class="tweet-action-icon hidden" />
                    {:else}
                      <HeartIcon size="20" fill={effectiveIsLiked ? "currentColor" : "none"} class="tweet-action-icon {heartAnimating ? "heart-animation" : ""}" />
                    {/if}
                  </div>
                  <span class="tweet-action-count" aria-live="polite">{effectiveLikes}</span>
                  <span class="like-status-text">{effectiveIsLiked ? "Liked" : "Like"}</span>
                </button>
              </div>
              <div class="tweet-action-item">
                <button
                  class="tweet-action-btn tweet-bookmark-btn {effectiveIsBookmarked ? "active" : ""} {isBookmarkLoading ? "loading" : ""} {isDarkMode ? "tweet-action-btn-dark" : ""}"
                  on:click|stopPropagation={toggleBookmarkStatus}
                  aria-label={effectiveIsBookmarked ? "Remove bookmark" : "Bookmark"}
                  disabled={isBookmarkLoading}
                >
                  {#if isBookmarkLoading}
                    <div class="tweet-action-loading"></div>
                    <BookmarkIcon size="20" fill={effectiveIsBookmarked ? "currentColor" : "none"} class="tweet-action-icon hidden" />
                  {:else}
                    <BookmarkIcon size="20" fill={effectiveIsBookmarked ? "currentColor" : "none"} class="tweet-action-icon" />
                  {/if}
                  <span class="tweet-action-count">{effectiveBookmarks}</span>
                </button>
              </div>
              <div class="tweet-action-item">
                <button
                  class="tweet-action-btn tweet-views-btn {isDarkMode ? "tweet-action-btn-dark" : ""}"
                  on:click|stopPropagation={navigateToThreadDetail}
                  aria-label="View thread details"
                >
                  <EyeIcon size="20" class="tweet-action-icon" />
                  <span class="tweet-action-count">{processedTweet.views || "0"}</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  {/if}
</div>

{#if nestingLevel === 0}
  <div class="tweet-replies-toggle-container">
    <button
      class="tweet-replies-toggle {isDarkMode ? "tweet-replies-toggle-dark" : ""}"
      on:click|stopPropagation={toggleReplies}
      aria-expanded={showReplies}
      aria-controls="replies-container-{tweetId}"
    >
      {#if showReplies}
        <ChevronUpIcon size="16" class="tweet-replies-toggle-icon" />
        Hide replies
      {:else}
        <ChevronDownIcon size="16" class="tweet-replies-toggle-icon" />
        {#if hasReplies}
          View replies ({effectiveReplies})
        {:else}
          View replies
        {/if}
      {/if}
    </button>
  </div>
{/if}

{#if showReplies}
  <div id="replies-container-{tweetId}" class="tweet-replies-container {isDarkMode ? "tweet-replies-container-dark" : ""}">
    {#if replies.length === 0}
      <div class="tweet-replies-empty {isDarkMode ? "tweet-replies-empty-dark" : ""}">
        <div class="tweet-replies-empty-icon">
          <MessageCircleIcon size="20" />
        </div>
        <div class="tweet-replies-empty-text">
          No replies yet. Be the first to reply!
        </div>
        <button
          class="tweet-replies-empty-btn"
          on:click|stopPropagation={handleReply}
        >
          Reply
        </button>
      </div>
    {:else}
      {#each processedReplies as reply (reply.id || reply.tweetId)}
        {#if !reply.content && typeof reply.content !== "undefined"}
          {console.error("⚠️ EMPTY REPLY CONTENT:", reply)}
        {/if}
        <div id="reply-{reply.id}-container" class="nested-reply-container">
          <svelte:self
            tweet={reply}
            {isDarkMode}
            {isAuthenticated}
            isLiked={reply.isLiked || reply.is_liked || false}
            isReposted={reply.isReposted || false}
            isBookmarked={reply.isBookmarked || false}
            inReplyToTweet={null}
            replies={nestedRepliesMap.get(String(reply.id)) || []}
            showReplies={false}
            nestingLevel={nestingLevel + 1}
            {nestedRepliesMap}
            on:reply={handleNestedReply}
            on:like={handleNestedLike}
            on:unlike={handleNestedLike}
            on:repost={handleNestedRepost}
            on:bookmark={handleNestedBookmark}
            on:removeBookmark={handleNestedBookmark}
            on:loadReplies={handleLoadNestedReplies}
          />

          {#if (Number(reply.replies_count) > 0)}
            {#if nestedRepliesMap.has(`retry_${reply.id}`)}
              <!-- Show retry button when loading failed -->
              <div class="nested-replies-retry-container">
                <button
                  class="nested-replies-retry-btn"
                  on:click|stopPropagation={() => retryLoadNestedReplies(reply.id)}
                >
                  <RefreshCwIcon size="14" />
                  Failed to load replies. Retry?
                </button>
              </div>
            {:else if !nestedRepliesMap.has(String(reply.id))}
              <!-- Show view replies button when not loaded yet -->
              <div class="nested-replies-view-container">
                <button
                  class="nested-replies-view-btn"
                  on:click|stopPropagation={() => handleLoadNestedReplies({ detail: String(reply.id) })}
                >
                  <ChevronDownIcon size="14" />
                  View {Number(reply.replies_count) || 0} {(Number(reply.replies_count) || 0) === 1 ? "reply" : "replies"}
                </button>
              </div>
            {/if}
          {/if}
        </div>
      {/each}
    {/if}
  </div>
{/if}

<style>

  :root {

    --transition-fast: 0.2s ease;
    --color-primary: #1da1f2;
    --color-primary-hover: #1991da;
    --color-primary-light: #e8f5fe;
    --bg-primary: #ffffff;
    --bg-secondary: #f7f9fa;
    --text-primary: #0f1419;
    --text-secondary: #536471;
    --border-color: #eff3f4;
    --hover-primary: rgba(29, 161, 242, 0.1);
    --hover-light: rgba(29, 161, 242, 0.1);
    --hover-dark: rgba(255, 255, 255, 0.1);
    --bg-hover: rgba(0, 0, 0, 0.05);
    --bg-hover-dark: rgba(255, 255, 255, 0.1);
    --radius-md: 8px;
    --radius-full: 9999px;
    --color-danger: #e0245e;
    --color-danger-rgb: 224, 36, 94;
    --color-primary-rgb: 29, 161, 242;
  }

  :global(.dark-theme) {
    --bg-primary: #15202b;
    --bg-primary-dark: #15202b;
    --bg-secondary: #1e2732;
    --bg-secondary-dark: #1e2732;
    --text-primary: #ffffff;
    --text-primary-dark: #ffffff;
    --text-secondary: #8899a6;
    --text-secondary-dark: #8899a6;
    --border-color: #38444d;
    --border-color-dark: #38444d;
    --hover-light: rgba(29, 161, 242, 0.1);
    --hover-dark: rgba(255, 255, 255, 0.1);
    --bg-hover: rgba(255, 255, 255, 0.05);
    --bg-hover-dark: rgba(255, 255, 255, 0.1);
  }

  .tweet-card {
    width: 100%;
    margin: 0;
    padding: 0;
    border-bottom: 1px solid var(--border-color);
    background-color: var(--bg-primary);
  }

  .tweet-card-dark {
    background-color: var(--bg-primary-dark);
    border-bottom: 1px solid var(--border-color-dark);
  }

  .tweet-card-container {
    padding: 12px 16px;
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .tweet-card-container:hover {
    background-color: rgba(0, 0, 0, 0.03);
  }

  .tweet-card-dark .tweet-card-container:hover {
    background-color: rgba(255, 255, 255, 0.05);
  }

  .tweet-card-content {
    width: 100%;
  }

  .tweet-card {
    padding: 0.5rem 0;
    border-bottom: 1px solid var(--border-color);
    background-color: var(--bg-primary);
    color: var(--text-primary);
    transition: background-color var(--transition-fast);
  }

  .tweet-card-dark {
    background-color: var(--bg-primary-dark);
    color: var(--text-primary-dark);
    border-bottom: 1px solid var(--border-color-dark);
  }

  .tweet-actions {
    display: flex;
    justify-content: space-between;
    padding: 0.5rem 0;
    margin-top: 0.5rem;
  }

  .tweet-actions-dark {
    color: var(--text-secondary-dark);
  }

  .tweet-action-item {
    display: flex;
    align-items: center;
  }

  .tweet-action-btn {
    display: flex;
    align-items: center;
    padding: 0.5rem;
    border-radius: 9999px;
    transition: all var(--transition-fast);
    cursor: pointer;
    background-color: transparent;
    border: none;
    color: var(--text-secondary);
    position: relative;
    min-width: 65px;
  }

  .tweet-action-btn:hover {
    background-color: rgba(var(--color-primary-rgb), 0.1);
    color: var(--color-primary);
  }

  .tweet-action-btn.active {
    color: var(--color-primary);
  }

  .tweet-action-btn.loading {
    pointer-events: none;
  }

  .tweet-action-loading {
    position: absolute;
    left: 0.5rem;
    width: 20px;
    height: 20px;
    border: 2px solid rgba(var(--color-primary-rgb), 0.3);
    border-radius: 50%;
    border-top-color: var(--color-primary);    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .tweet-action-count {
    font-size: 0.875rem;
  }

  .like-status-text {
    font-size: 0;
    width: 0;
    height: 0;
    overflow: hidden;
    position: absolute;
    left: -9999px;
  }

  @media (max-width: 768px) {
    .like-status-text {
      font-size: 0.75rem;
      width: auto;
      height: auto;
      position: static;
      margin-left: 0.25rem;
      overflow: visible;
      display: none;
    }

    .tweet-like-btn.active .like-status-text {
      display: inline;
      color: var(--color-primary);
      font-weight: 500;
    }
  }

  .view-replies-btn {
    background-color: var(--hover-primary);
    border: 1px solid var(--color-primary-light);
    color: var(--color-primary);
    font-size: 0.9rem;
    margin-left: 0.75rem;
    cursor: pointer;
    padding: 0.5rem 0.75rem;
    border-radius: var(--radius-full);
    transition: all 0.2s;
    display: flex;
    align-items: center;
    gap: 0.25rem;
    font-weight: 600;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .view-replies-btn:hover {
    background-color: var(--color-primary-light);
    transform: translateY(-2px);
    box-shadow: 0 3px 6px rgba(0, 0, 0, 0.15);
  }

  .view-replies-btn:active {
    transform: translateY(0);
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  }

  .tweet-replies-toggle {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    background-color: var(--hover-primary);
    border: none;
    color: var(--color-primary);
    font-weight: 600;
    padding: 0.75rem;
    cursor: pointer;
    width: 100%;
    border-radius: var(--radius-md);
    margin-top: 0.5rem;
    transition: all 0.2s;
    border: 1px solid var(--border-color);
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .tweet-replies-toggle:hover {
    background-color: var(--color-primary-light);
    transform: translateY(-1px);
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.15);
  }

  .tweet-replies-toggle-dark {
    background-color: var(--bg-secondary-dark);
    border-color: var(--border-color-dark);
  }

  :global(.loading-replies) {
    position: relative;
  }

  :global(.loading-replies)::after {
    content: "";
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 5px;
    background: linear-gradient(
      90deg,
      transparent,
      var(--color-primary),
      transparent
    );
    animation: loading-animation 1.5s infinite;
  }

  :global(.loading-nested-replies) {
    position: relative;
    opacity: 0.8;
  }

  :global(.loading-nested-replies)::before {
    content: "";
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(255, 255, 255, 0.1);
    z-index: 1;
  }

  :global(.loading-nested-replies)::after {
    content: "";
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 3px;
    background: linear-gradient(
      90deg,
      transparent,
      var(--color-primary),
      transparent
    );
    animation: loading-animation 1.5s infinite;
    z-index: 2;
  }

  @keyframes loading-animation {
    0% { transform: translateX(-100%); }
    100% { transform: translateX(100%); }
  }

  .tweet-reply-count-highlight {
    color: var(--color-primary);
    font-weight: 600;
  }

  .tweet-views-btn {
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 4px;
    color: var(--text-secondary);
    background: transparent;
    border: none;
    padding: 0.5rem;
    border-radius: 9999px;
    transition: all 0.2s ease;
  }

  .tweet-views-btn:hover {
    color: var(--color-primary);
    background-color: var(--hover-light);
  }

  .tweet-action-btn-dark.tweet-views-btn {
    color: var(--text-secondary-dark);
  }

  .tweet-action-btn-dark.tweet-views-btn:hover {
    color: var(--color-primary);
    background-color: var(--hover-dark);
  }

  :global(.tweet-action-btn.clicked) {
    transform: scale(1.1);
    transition: transform 0.2s;
  }

  :global(.tweet-replies-toggle-icon) {
    transition: transform 0.2s;
  }

  .has-replies {
    font-weight: 500;
  }

  :global(.tweet-reply-action-btn) {
    background-color: var(--bg-secondary);
    color: var(--text-secondary);
    border: none;
    border-radius: var(--radius-full);
    padding: 0.375rem 0.625rem;
    display: flex;
    align-items: center;
    gap: 4px;
    cursor: pointer;
    transition: all 0.2s;
  }

  :global(.tweet-reply-action-btn-dark) {
    background-color: var(--bg-secondary-dark);
    color: var(--text-secondary-dark);
  }

  :global(.tweet-reply-action-btn:hover) {
    background-color: var(--hover-primary);
    color: var(--color-primary);
  }

  :global(.tweet-reply-action-btn-dark:hover) {
    background-color: var(--hover-primary);
    color: var(--color-primary);
  }

  :global(.tweet-reply-action-btn.active) {
    color: var(--color-primary);
  }

  :global(.tweet-reply-action-btn-dark.active) {
    color: var(--color-primary);
  }

  .tweet-replies-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 2rem 1rem;
    text-align: center;
    background-color: var(--bg-secondary);
    border-radius: var(--radius-md);
    margin: 1rem 0;
  }

  .tweet-replies-empty-dark {
    background-color: var(--bg-secondary-dark);
  }

  .tweet-replies-empty-icon {
    margin-bottom: 0.5rem;
    color: var(--text-secondary);
  }

  .tweet-replies-empty-text {
    margin-bottom: 1rem;
    color: var(--text-secondary);
    font-size: 0.95rem;
  }

  .tweet-replies-empty-btn {
    background-color: var(--color-primary);
    color: white;
    border: none;
    padding: 0.5rem 1.25rem;
    border-radius: var(--radius-full);
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .tweet-replies-empty-btn:hover {
    background-color: var(--color-primary-hover);
    transform: translateY(-1px);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .nested-reply-container {
    margin-left: 1rem;
    border-left: 2px solid var(--border-color);
    padding-left: 1rem;
    position: relative;
  }

  .nested-replies-retry-container,
  .nested-replies-view-container {
    margin-top: 0.5rem;
    margin-bottom: 1rem;
    padding-left: 3rem;
  }

  .nested-replies-retry-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background-color: rgba(255, 193, 7, 0.1);
    color: #e6a700;
    border: 1px solid rgba(255, 193, 7, 0.2);
    padding: 0.5rem 1rem;
    border-radius: 9999px;
    font-size: 0.875rem;
    cursor: pointer;
    transition: all 0.2s;
  }

  .nested-replies-retry-btn:hover {
    background-color: rgba(255, 193, 7, 0.2);
    transform: translateY(-1px);
  }

  .nested-replies-view-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background-color: rgba(29, 155, 240, 0.1);
    color: #1d9bf0;
    border: 1px solid rgba(29, 155, 240, 0.2);
    padding: 0.5rem 1rem;
    border-radius: 9999px;
    font-size: 0.875rem;
    cursor: pointer;
    transition: all 0.2s;
  }

  .nested-replies-view-btn:hover {
    background-color: rgba(29, 155, 240, 0.2);
    transform: translateY(-1px);
  }

  :global(.tweet-reply-action-btn.loading) {
    pointer-events: none;
    opacity: 0.8;
  }

  :global(.tweet-reply-action-loading) {
    width: 16px;
    height: 16px;
    border: 2px solid rgba(var(--color-danger-rgb), 0.3);
    border-radius: 50%;
    border-top-color: var(--color-danger);
    animation: spin 0.8s linear infinite;
  }

  :global(.tweet-reply-avatar-placeholder) {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    background-color: var(--color-primary-light);
    color: var(--color-primary);
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: bold;
    font-size: 14px;
  }

  .tweet-like-icon-wrapper {
    position: relative;
    display: flex;
    justify-content: center;
    align-items: center;
  }

  :global(.tweet-like-btn.animating .heart-animation) {
    animation: heartBeat 0.8s ease;
    transform-origin: center;
  }

  @keyframes heartBeat {
    0% {
      transform: scale(1);
    }
    15% {
      transform: scale(1.2);
    }
    30% {
      transform: scale(0.95);
    }
    45% {
      transform: scale(1.1);
    }
    60% {
      transform: scale(1);
    }
  }

  .tweet-action-loading {
    position: absolute;
    left: 0.5rem;
    width: 20px;
    height: 20px;
    border: 2px solid rgba(var(--color-danger-rgb), 0.3);
    border-radius: 50%;
    border-top-color: var(--color-danger);
    animation: spin 0.8s linear infinite;
  }

  @media (max-width: 768px) {
    .tweet-action-btn {
      padding: 0.75rem;
      min-height: 40px;
      min-width: 40px;
    }

    .like-status-text {
      font-size: 0.75rem;
      width: auto;
      height: auto;
      position: static;
      margin-left: 0.25rem;
      overflow: visible;
      transition: opacity 0.2s ease;
      opacity: 0;
    }

    .tweet-like-btn.active .like-status-text {
      display: inline;
      color: var(--color-danger);
      font-weight: 500;
      opacity: 1;
    }
  }

  .tweet-action-btn:focus {
    outline: 2px solid var(--color-primary);
    outline-offset: 2px;
  }

  :global(.heart-pulse) {
    animation: heartPulse 0.8s ease;
  }

  @keyframes heartPulse {
    0% {
      transform: scale(1);
    }
    25% {
      transform: scale(1.3);
    }
    50% {
      transform: scale(0.9);
    }
    75% {
      transform: scale(1.2);
    }
    100% {
      transform: scale(1);
    }
  }

  :global(.tweet-reply-action-loading-alt) {
    display: inline-block;
    width: 16px;
    height: 16px;
    border: 2px solid rgba(var(--color-danger-rgb), 0.3);
    border-radius: 50%;
    border-top-color: var(--color-danger);
    animation: spin 0.8s linear infinite;
    margin-right: 4px;
  }

  .user-verified-badge {
    color: #1DA1F2 !important;
    display: inline-flex;
    align-items: center;
    margin-left: 4px;
    filter: drop-shadow(0 0 1px rgba(29, 161, 242, 0.3));
  }

  .user-verified-badge :global(svg) {
    stroke-width: 2.5px;
    background-color: rgba(29, 161, 242, 0.1);
    border-radius: 50%;
  }

  .tweet-author-name {
    display: flex;
    align-items: center;
  }

  .display-name-text {
    margin-right: 2px;
    font-weight: 600;
  }

  .tweet-empty-content {
    font-style: italic;
    color: var(--text-secondary);
    opacity: 0.8;
  }

  :global([data-theme="dark"]) .tweet-empty-content {
    color: var(--dark-text-secondary);
  }

  .tweet-settings-container {
    position: relative;
    margin-left: auto;
  }

  .tweet-settings-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    background: transparent;
    border: none;
    border-radius: 50%;
    width: 30px;
    height: 30px;
    padding: 0;
    cursor: pointer;
    color: var(--text-secondary);
    transition: all 0.2s;
  }

  .tweet-settings-btn:hover {
    background-color: var(--hover-light);
    color: var(--color-primary);
  }

  .tweet-settings-btn-dark {
    color: var(--text-secondary-dark);
  }

  .tweet-settings-btn-dark:hover {
    background-color: var(--hover-dark);
    color: var(--color-primary);
  }

  .tweet-settings-dropdown {
    position: absolute;
    top: 100%;
    right: 0;
    background-color: var(--bg-primary);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
    z-index: 10;
    min-width: 150px;
    margin-top: 5px;
    overflow: hidden;
  }

  .tweet-settings-dropdown-dark {
    background-color: var(--bg-primary-dark);
    border-color: var(--border-color-dark);
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.3);
  }

  .tweet-settings-option {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 10px 15px;
    border: none;
    background-color: transparent;
    text-align: left;
    cursor: pointer;
    transition: all 0.2s;
    color: var(--text-primary);
  }

  .tweet-settings-option-dark {
    color: var(--text-primary-dark);
  }

  .tweet-delete-option {
    color: var(--color-danger);
  }

  .tweet-delete-option:hover {
    background-color: rgba(var(--color-danger-rgb), 0.1);
  }

  .tweet-settings-option-dark.tweet-delete-option:hover {
    background-color: rgba(var(--color-danger-rgb), 0.2);
  }

  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    backdrop-filter: blur(2px);
  }

  .modal-container {
    background-color: var(--bg-primary);
    border-radius: var(--radius-md);
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
    width: 90%;
    max-width: 400px;
    overflow: hidden;
    padding: 0;
    animation: modal-slide-in 0.3s ease;
  }

  @keyframes modal-slide-in {
    from {
      opacity: 0;
      transform: translateY(-20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .modal-header {
    padding: 16px 20px;
    border-bottom: 1px solid var(--border-color);
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .modal-title {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0;
  }

  .modal-close {
    background: transparent;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    padding: 5px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    transition: all 0.2s;
  }

  .modal-close:hover {
    background-color: var(--hover-light);
    color: var(--color-primary);
  }

  .modal-body {
    padding: 20px;
    color: var(--text-primary);
  }

  .modal-warning-text {
    color: var(--color-danger);
    font-weight: 500;
    margin-bottom: 8px;
  }

  .modal-footer {
    padding: 16px 20px;
    border-top: 1px solid var(--border-color);
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }

  .modal-btn {
    padding: 8px 16px;
    border-radius: var(--radius-full);
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .modal-btn-cancel {
    background-color: transparent;
    border: 1px solid var(--border-color);
    color: var(--text-primary);
  }

  .modal-btn-cancel:hover {
    background-color: var(--bg-hover);
  }

  .modal-btn-danger {
    background-color: var(--color-danger);
    border: 1px solid var(--color-danger);
    color: white;
  }

  .modal-btn-danger:hover {
    background-color: rgba(var(--color-danger-rgb), 0.9);
  }

  .modal-container-dark {
    background-color: var(--bg-primary-dark);
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
  }

  .modal-header-dark {
    border-bottom: 1px solid var(--border-color-dark);
  }

  .modal-title-dark {
    color: var(--text-primary-dark);
  }

  .modal-close-dark {
    color: var(--text-secondary-dark);
  }

  .modal-close-dark:hover {
    background-color: var(--hover-dark);
  }

  .modal-body-dark {
    color: var(--text-primary-dark);
  }

  .modal-footer-dark {
    border-top: 1px solid var(--border-color-dark);
  }

  .modal-btn-cancel-dark {
    border: 1px solid var(--border-color-dark);
    color: var(--text-primary-dark);
  }

  .modal-btn-cancel-dark:hover {
    background-color: var(--bg-hover-dark);
  }

  .tweet-community-info {
    display: flex;
    align-items: center;
    gap: 5px;
    margin-top: 6px;
    margin-bottom: 8px;
    padding: 4px 8px;
    border-radius: 16px;
    background-color: rgba(var(--color-primary-rgb), 0.08);
    color: var(--color-primary);
    width: fit-content;
    font-size: 0.85rem;
  }

  .tweet-community-info-dark {
    background-color: rgba(var(--color-primary-rgb), 0.15);
  }

  .tweet-community-name {
    font-weight: 500;
    }

  .tweet-ad-badge {
    background-color: #ffd700;
    color: #000;
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 0.8rem;
    margin-top: 4px;
    margin-bottom: 4px;
    display: inline-block;
  }

  .tweet-ad-badge-dark {
    background-color: #ffa500;
    color: #fff;
  }

  .ad-label {
    font-weight: bold;
  }

  .tweet-admin-indicator {
    color: #2f80ed;
    font-size: 0.9rem;
    font-weight: 600;
    margin-top: 6px;
    display: flex;
    align-items: center;
  }

  .tweet-admin-indicator::before {
    content: "⚙️";
    margin-right: 4px;
  }

  .tweet-admin-indicator-dark {
    color: #4da3ff;
  }

  .tweet-parent-info {
    margin-top: 0.5rem;
    padding-left: 1rem;
    border-left: 2px solid var(--border-color);
  }

  .tweet-parent-info-dark {
    border-left-color: var(--border-color-dark);
  }

  .tweet-parent-indicator {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    color: var(--text-secondary);
  }

  .tweet-parent-username {
    color: var(--color-primary);
    font-weight: 600;
  }

  .tweet-parent-content {
    margin-top: 0.25rem;
    color: var(--text-secondary);
  }

  .skeleton {
    background-color: var(--bg-secondary);
    background-image: linear-gradient(90deg,
      var(--bg-secondary) 25%,
      var(--bg-hover) 50%,
      var(--bg-secondary) 75%);
    background-size: 200% 100%;
    animation: loading-skeleton 1.5s ease-in-out infinite;
    border-radius: var(--radius-md);
  }

  @keyframes loading-skeleton {
    0% { background-position: 200% 0; }
    100% { background-position: -200% 0; }
  }

  .tweet-card-skeleton {
    padding: var(--space-3) var(--space-4);
    width: 100%;
  }

  .tweet-avatar.skeleton {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    margin-right: var(--space-3);
  }

  .tweet-author-name-skeleton {
    width: 120px;
    height: 18px;
    margin-right: var(--space-3);
    margin-bottom: var(--space-2);
  }

  .tweet-author-username-skeleton {
    width: 80px;
    height: 14px;
  }

  .tweet-text-skeleton {
    height: 16px;
    margin-bottom: var(--space-2);
  }

  .tweet-text-skeleton:first-child {
    width: 90%;
  }

  .tweet-text-skeleton:nth-child(2) {
    width: 70%;
  }

  .tweet-media-skeleton {
    height: 180px;
    width: 100%;
    margin-top: var(--space-3);
    border-radius: var(--radius-md);
  }

  .tweet-action-item-skeleton {
    height: 20px;
    width: 60px;
    border-radius: var(--radius-md);
  }

  .tweet-content-text {
    margin: var(--space-2) 0;
    word-break: break-word;
    line-height: 1.4;
  }

  .tweet-media-btn {
    background: none;
    border: none;
    padding: 0;
    cursor: pointer;
    display: block;
    width: 100%;
    height: 100%;
    border-radius: 12px;
    overflow: hidden;
    transition: transform 0.2s ease, box-shadow 0.2s ease;
  }

  .tweet-media-btn:hover {
    transform: scale(1.02);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  }

  .tweet-media-btn:focus {
    outline: 2px solid #1d9bf0;
    outline-offset: 2px;
  }

  .tweet-media-btn .tweet-media-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
</style>

<!-- Delete confirmation modal -->
{#if showDeleteConfirmationModal}
  <div
    class="modal-overlay"
    on:click|self={cancelDeleteTweet}
    on:keydown={(e) => e.key === "Escape" && cancelDeleteTweet()}
    role="dialog"
    aria-modal="true"
    tabindex="0">
    <div class="modal-container {isDarkMode ? "modal-container-dark" : ""}">
      <div class="modal-header {isDarkMode ? "modal-header-dark" : ""}">
        <h3 class="modal-title {isDarkMode ? "modal-title-dark" : ""}">Delete Post</h3>
        <button class="modal-close {isDarkMode ? "modal-close-dark" : ""}" on:click={cancelDeleteTweet}>
          <XIcon size="18" />
        </button>
      </div>
      <div class="modal-body {isDarkMode ? "modal-body-dark" : ""}">
        <p class="modal-warning-text">Are you sure you want to delete this post?</p>
        <p>This action cannot be undone. The post, all its likes, bookmarks, and replies will be permanently deleted.</p>
      </div>
      <div class="modal-footer {isDarkMode ? "modal-footer-dark" : ""}">
        <button
          class="modal-btn modal-btn-cancel {isDarkMode ? "modal-btn-cancel-dark" : ""}"
          on:click={cancelDeleteTweet}
        >
          Cancel
        </button>
        <button
          class="modal-btn modal-btn-danger"
          on:click={confirmDeleteTweet}
        >
          Delete
        </button>
      </div>
    </div>
  </div>
{/if}