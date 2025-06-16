/**
 * Media-related interfaces
 */

/**
 * Media object - the primary interface for media across the application
 */
export interface IMedia {
  id?: string;
  url: string;
  type: "image" | "video" | "gif";
  width?: number;
  height?: number;
  thumbnail_url?: string;
  alt_text?: string;
}

/**
 * Media upload response
 */
export interface IMediaUploadResponse {
  success: boolean;
  data: {
    id: string;
    type: "image" | "video" | "gif";
    url: string;
    thumbnail?: string;
  };
}

/**
 * Media search response
 */
export interface IMediaSearchResponse {
  success: boolean;
  data: {
    message: string;
    query: string;
    results: IMedia[];
  };
}

/**
 * Media update request
 */
export interface IMediaUpdateRequest {
  media_urls: string[];
}

/**
 * Media update response
 */
export interface IMediaUpdateResponse {
  success: boolean;
  data: {
    message: string;
    urls: string[];
  };
}