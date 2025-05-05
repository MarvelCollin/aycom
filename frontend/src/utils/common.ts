import { toastStore } from '../stores/toastStore';
import type { IAuthStore } from '../interfaces/IAuth';
import type { IMedia } from '../interfaces/ISocialMedia';
import { createLoggerWithPrefix } from './logger';

const logger = createLoggerWithPrefix('common-utils');

/**
 * Format a timestamp as a relative time (e.g., "2h", "3d")
 * @param timestamp ISO string timestamp
 * @returns Formatted time string
 */
export function formatTimeAgo(timestamp: string): string {
  try {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffSec = Math.floor(diffMs / 1000);
    const diffMin = Math.floor(diffSec / 60);
    const diffHour = Math.floor(diffMin / 60);
    const diffDay = Math.floor(diffHour / 24);
    
    if (diffSec < 60) return `${diffSec}s`;
    if (diffMin < 60) return `${diffMin}m`;
    if (diffHour < 24) return `${diffHour}h`;
    if (diffDay < 7) return `${diffDay}d`;
    if (diffDay < 30) return `${Math.floor(diffDay / 7)}w`;
    if (diffDay < 365) return `${Math.floor(diffDay / 30)}mo`;
    
    return `${Math.floor(diffDay / 365)}y`;
  } catch (e) {
    return 'recently';
  }
}

/**
 * Check if the user is authenticated, redirect to login if not
 * @param authState Current authentication state
 * @param featureName Name of the feature requiring authentication
 * @returns Boolean indicating if authenticated
 */
export function checkAuth(authState: IAuthStore, featureName: string): boolean {
  if (!authState.isAuthenticated) {
    toastStore.showToast(`You need to log in to access ${featureName}`, 'warning');
    window.location.href = '/login';
    return false;
  }
  return true;
}

/**
 * Check if a timestamp is within the last X milliseconds
 * @param timestamp ISO string timestamp
 * @param withinMs Milliseconds to check against (default 60000 = 1 minute)
 * @returns Boolean indicating if timestamp is within specified time
 */
export function isWithinTime(timestamp: string, withinMs: number = 60000): boolean {
  try {
    const time = new Date(timestamp).getTime();
    const now = new Date().getTime();
    const diffMs = now - time;
    return diffMs < withinMs;
  } catch (e) {
    return false;
  }
}

/**
 * Generate a preview URL for a file
 * @param file File to generate preview for
 * @returns Object with URL and type
 */
export function generateFilePreview(file: File): IMedia {
  const url = URL.createObjectURL(file);
  const type = file.type.startsWith('image/') ? 'image' : 
               file.type.startsWith('video/') ? 'video' : 'document';
  
  return {
    url,
    type,
    alt: file.name
  };
}

/**
 * Process user metadata from tweet content field
 * Format: [USER:username@displayName]content
 * @param content Content string with potential user metadata
 * @returns Object with extracted username, displayName and cleaned content
 */
export function processUserMetadata(content: string): { username?: string, displayName?: string, content: string } {
  if (!content) return { content: '' };
  
  const userMetadataRegex = /^\[USER:([^@\]]+)(?:@([^\]]+))?\](.*)/;
  const match = content.match(userMetadataRegex);
  
  if (match) {
    return {
      username: match[1] || undefined,
      displayName: match[2] || undefined,
      content: match[3] || ''
    };
  }
  
  return { content };
}

/**
 * Safely handle API errors and return standardized error response
 * @param error Error object or unknown value
 * @returns Standardized error response
 */
export function handleApiError(error: unknown): { success: false, message: string } {
  if (error instanceof Error) {
    if (error.name === 'AbortError') {
      return { success: false, message: 'Request timed out. Please try again.' };
    }
    return { success: false, message: error.message };
  }
  return { success: false, message: 'An unexpected error occurred.' };
}

/**
 * Truncate text to specified length with ellipsis
 * @param text Text to truncate
 * @param maxLength Maximum length before truncation
 * @returns Truncated text
 */
export function truncateText(text: string, maxLength: number = 100): string {
  if (!text || text.length <= maxLength) return text;
  return text.substring(0, maxLength) + '...';
} 