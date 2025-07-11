import type { ITweet } from "./ISocialMedia";

export interface ExtendedTweet extends ITweet {

  retweet_id?: string;
  threadId?: string;
  thread_id?: string;
  tweetId?: string;
  userId?: string;
  authorId?: string;

  display_name?: string;
  avatar?: string;
  [key: string]: any;
}

interface IMedia {
  id: string;
  url: string;
  type: string;
  thumbnail?: string;
  alt_text?: string;
}

export function ensureTweetFormat(thread: any): ExtendedTweet {
  try {
    if (!thread || typeof thread !== "object") {
      return createEmptyTweet();
    }

    console.log("Raw thread data in ensureTweetFormat:", thread);

    let username = thread.username;

    if (!username && thread.author) {
      username = thread.author.username || thread.author_username;
    }

    if (!username) {
      username = thread.author_username ||
                thread.authorUsername ||
                thread.user?.username ||
                thread.author?.username ||
                thread.user_data?.username;
    }

    if (!username || username === "anonymous") {
      username = `user_${thread.user_id || thread.userId || thread.authorId || thread.id || Math.random().toString(36).substring(2, 9)}`;
    }

    const name = thread.name ||
               thread.author_name ||
               thread.authorName ||
               thread.display_name ||
               thread.displayName ||
               thread.author?.name ||
               thread.author?.display_name ||
               thread.user?.name ||
               thread.user_data?.name ||
               username;

    const profile_picture_url = thread.profile_picture_url ||
                        thread.profilePictureUrl ||
                        thread.author_avatar ||
                        thread.authorAvatar ||
                        thread.avatar ||
                        thread.author?.profile_picture_url ||
                        thread.user?.profile_picture_url ||
                        thread.user_data?.profile_picture_url ||
                        "https://secure.gravatar.com/avatar/0?d=mp";

    let created_at;
    try {
      if (!thread.created_at && !thread.createdAt && !thread.timestamp) {

        created_at = new Date().toISOString();
      } else if (typeof thread.created_at === "string") {

        if (thread.created_at.includes("T")) {
          created_at = thread.created_at;
        } else {

          const parsedDate = new Date(thread.created_at);

          if (isNaN(parsedDate.getTime())) {
            created_at = new Date().toISOString(); 
          } else {
            created_at = parsedDate.toISOString();
          }
        }
      } else if (thread.created_at instanceof Date) {

        created_at = thread.created_at.toISOString();
      } else if (typeof thread.createdAt === "string") {

        created_at = thread.createdAt;
      } else if (typeof thread.timestamp === "string" || typeof thread.timestamp === "number") {

        const date = new Date(thread.timestamp);
        if (isNaN(date.getTime())) {
          created_at = new Date().toISOString();
        } else {
          created_at = date.toISOString();
        }
      } else {

        created_at = new Date().toISOString();
      }
    } catch (e) {
      console.error("Error parsing date", e);
      created_at = new Date().toISOString(); 
    }

    let updated_at;
    try {
      if (!thread.updated_at && !thread.updatedAt) {

        updated_at = created_at;
      } else if (typeof thread.updated_at === "string") {

        if (thread.updated_at.includes("T")) {
          updated_at = thread.updated_at;
        } else {

          const parsedDate = new Date(thread.updated_at);

          if (isNaN(parsedDate.getTime())) {
            updated_at = created_at; 
          } else {
            updated_at = parsedDate.toISOString();
          }
        }
      } else if (thread.updated_at instanceof Date) {

        updated_at = thread.updated_at.toISOString();
      } else if (typeof thread.updatedAt === "string") {

        updated_at = thread.updatedAt;
      } else {

        updated_at = created_at;
      }
    } catch (e) {
      console.error("Error parsing updated_at date", e);
      updated_at = created_at; 
    }

    const likes_count = Number(thread.likes_count || thread.like_count || thread.metrics?.likes || 0);
    const replies_count = Number(thread.replies_count || thread.reply_count || thread.metrics?.replies || 0);
    const reposts_count = Number(thread.reposts_count || thread.repost_count || thread.metrics?.reposts || 0);
    const bookmark_count = Number(thread.bookmarks_count || thread.bookmark_count || thread.metrics?.bookmarks || 0);
    const views_count = Number(thread.views || thread.views_count || 0);

    const is_liked = Boolean(thread.is_liked || thread.isLiked || false);
    const is_reposted = Boolean(thread.is_repost || thread.isReposted || false);
    const is_bookmarked = Boolean(thread.is_bookmarked || thread.isBookmarked || false);
    const is_pinned = Boolean(
      thread.is_pinned === true ||
      thread.is_pinned === "true" ||
      thread.is_pinned === 1 ||
      thread.is_pinned === "1" ||
      thread.is_pinned === "t" ||
      thread.IsPinned === true ||
      false
    );

    const is_verified = Boolean(
      thread.is_verified ||
      thread.verified ||
      thread.user?.is_verified ||
      thread.author?.is_verified ||
      false
    );

    const media = Array.isArray(thread.media) ? thread.media : [];

    const id = thread.id || `thread-${Math.random().toString(36).substring(2, 9)}`;
    const user_id = thread.user_id || thread.userId || thread.author_id || thread.authorId || thread.author?.id || thread.user?.id || "";

    return {
      id,
      content: thread.content || "",
      created_at: created_at,
      updated_at: updated_at,

      user_id,
      username,
      name,
      profile_picture_url,

      likes_count,
      replies_count,
      reposts_count,
      bookmark_count,
      views_count,

      is_liked,
      is_reposted,
      is_bookmarked,
      is_pinned,
      is_verified,

      parent_id: thread.parent_id || thread.parentId || thread.parent_reply_id || thread.parentReplyId || null,

      media,

      thread_id: thread.thread_id || thread.threadId || id,

      ...thread
    };
  } catch (error) {
    console.error("Error formatting tweet:", error);
    return createEmptyTweet();
  }
}

function createEmptyTweet(): ExtendedTweet {
  const id = `error-${Math.random().toString(36).substring(2, 9)}`;
  return {
    id,
    content: "Error loading tweet",
    created_at: new Date().toISOString(),
    updated_at: undefined,

    user_id: "",
    username: "error",
    name: "Error",
    profile_picture_url: "",

    likes_count: 0,
    replies_count: 0,
    reposts_count: 0,
    bookmark_count: 0,
    views_count: 0,

    is_liked: false,
    is_reposted: false,
    is_bookmarked: false,
    is_pinned: false,
    is_verified: false,

    parent_id: null,

    media: [],

    thread_id: id
  };
}