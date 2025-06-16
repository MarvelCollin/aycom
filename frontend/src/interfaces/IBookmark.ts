/**
 * Bookmark-related interfaces
 */

import type { ITweet } from "./ISocialMedia";
import type { IPagination } from "./ICommon";

/**
 * Bookmarks response
 */
export interface IBookmarksResponse {
  success: boolean;
  data: {
    bookmarks: ITweet[];
    pagination: IPagination;
  };
}

/**
 * Search bookmarks request
 */
export interface ISearchBookmarksRequest {
  q: string;
  page?: number;
  limit?: number;
}

/**
 * Delete bookmark response
 */
export interface IDeleteBookmarkResponse {
  success: boolean;
  data: {
    message: string;
  };
}