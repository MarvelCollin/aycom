import { toastStore } from '../stores/toastStore';
import type { IAuthStore } from '../interfaces/IAuth';
import type { IMedia } from '../interfaces/ISocialMedia';
import { createLoggerWithPrefix } from './logger';

const logger = createLoggerWithPrefix('common-utils');

/**
 * Format a timestamp into a relative time ago string
 * @param {string} timestamp - ISO timestamp string
 * @returns {string} - Formatted time ago string (e.g., "2h", "3d", "1w")
 */
export function formatTimeAgo(timestamp: string): string {
  if (!timestamp) return 'now';
  
  try {
    const date = new Date(timestamp);
    
    if (isNaN(date.getTime())) {
      return 'now';
    }
    
    const now = new Date();
    const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);
    
    if (seconds < 0) return 'now';
    
    // Less than a minute
    if (seconds < 60) {
      return `${seconds}s`;
    }
    
    // Less than an hour
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) {
      return `${minutes}m`;
    }
    
    // Less than a day
    const hours = Math.floor(minutes / 60);
    if (hours < 24) {
      return `${hours}h`;
    }
    
    // Less than a week
    const days = Math.floor(hours / 24);
    if (days < 7) {
      return `${days}d`;
    }
    
    // Less than a month
    const weeks = Math.floor(days / 7);
    if (weeks < 4) {
      return `${weeks}w`;
    }
    
    // Less than a year
    const months = Math.floor(days / 30);
    if (months < 12) {
      return `${months}mo`;
    }
    
    // Years
    const years = Math.floor(days / 365);
    return `${years}y`;
    
  } catch (error) {
    console.error('Error formatting timestamp:', error);
    return 'now';
  }
}

export function checkAuth(authState: IAuthStore, featureName: string): boolean {
  if (!authState.isAuthenticated) {
    toastStore.showToast(`You need to log in to access ${featureName}`, 'warning');
    window.location.href = '/login';
    return false;
  }
  return true;
}

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
 * Generate a preview object for a file
 * @param {File} file - The file to generate a preview for
 * @returns {IMedia} - Media object with URL and type
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
 * Extracts user metadata from content that might contain embedded information
 * @param {string} content - Content string that might contain embedded metadata
 * @returns {object} - Object with extracted username, displayName, and cleaned content
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
 * Handle API errors in a consistent way
 * @param {unknown} error - The error to handle
 * @returns {object} - Standard error response object
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
 * Truncate text to a specified length and add ellipsis
 * @param {string} text - The text to truncate
 * @param {number} maxLength - Maximum length before truncation
 * @returns {string} - Truncated text with ellipsis if needed
 */
export function truncateText(text: string, maxLength: number = 100): string {
  if (!text || text.length <= maxLength) return text;
  return text.substring(0, maxLength) + '...';
}

/**
 * Check if a URL is a Supabase storage URL
 * @param {string} url - URL to check
 * @returns {boolean} - Whether the URL is a Supabase storage URL
 */
export function isSupabaseStorageUrl(url: string): boolean {
  const supabaseUrlPattern = /supabase\.co\/storage\/v1\/object\/public\//;
  return supabaseUrlPattern.test(url);
}

/**
 * Format a storage URL to ensure it includes the full Supabase path
 * @param {string|null} url - URL to format
 * @returns {string} - Formatted URL
 */
export function formatStorageUrl(url: string | null): string {
  if (!url) return '';
  
  // If the URL is already a complete URL, return it as is
  if (url.startsWith('http://') || url.startsWith('https://')) {
    console.log('URL already complete:', url);
    return url;
  }
  
  const supabaseUrl = import.meta.env.VITE_SUPABASE_URL || 'https://sdhtnvlmuywinhcglfsu.supabase.co';
  
  // Handle case where URL is just a filename without proper path structure
  if (!url.includes('/')) {
    // First try the profile-pictures bucket for profile images
    if (url.match(/\.(jpg|jpeg|png|gif|webp|svg)$/i)) {
      const formatted = `${supabaseUrl}/storage/v1/object/public/profile-pictures/${url}`;
      console.log(`Formatted filename-only URL (profile): ${url} -> ${formatted}`);
      return formatted;
    }
    
    // Default to tpaweb bucket
    const formatted = `${supabaseUrl}/storage/v1/object/public/tpaweb/${url}`;
    console.log(`Formatted filename-only URL: ${url} -> ${formatted}`);
    return formatted;
  }
  
  // Handle case where URL is storage path but missing base URL
  // Make sure we don't duplicate '/storage/v1/object/public/'
  if (url.startsWith('storage/v1/object/public/')) {
    const formatted = `${supabaseUrl}/${url}`;
    console.log(`Formatted storage path URL: ${url} -> ${formatted}`);
    return formatted;
  }
  
  // Handle profile-pictures/filename.jpg format
  if (url.startsWith('profile-pictures/')) {
    const formatted = `${supabaseUrl}/storage/v1/object/public/${url}`;
    console.log(`Formatted profile URL: ${url} -> ${formatted}`);
    return formatted;
  }
  
  // Handle banners/filename.jpg format
  if (url.startsWith('banners/')) {
    const formatted = `${supabaseUrl}/storage/v1/object/public/${url}`;
    console.log(`Formatted banner URL: ${url} -> ${formatted}`);
    return formatted;
  }
  
  // Default case - add the standard storage path prefix
  const formatted = `${supabaseUrl}/storage/v1/object/public/${url}`;
  console.log(`Formatted default URL: ${url} -> ${formatted}`);
  return formatted;
} 