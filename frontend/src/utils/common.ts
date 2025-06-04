import { toastStore } from '../stores/toastStore';
import type { IAuthStore } from '../interfaces/IAuth';
import type { IMedia } from '../interfaces/IMedia';
import { createLoggerWithPrefix } from './logger';

const logger = createLoggerWithPrefix('common-utils');

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

    if (seconds < 60) {
      return `${seconds}s`;
    }

    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) {
      return `${minutes}m`;
    }

    const hours = Math.floor(minutes / 60);
    if (hours < 24) {
      return `${hours}h`;
    }

    const days = Math.floor(hours / 24);
    if (days < 7) {
      return `${days}d`;
    }

    const weeks = Math.floor(days / 7);
    if (weeks < 4) {
      return `${weeks}w`;
    }

    const months = Math.floor(days / 30);
    if (months < 12) {
      return `${months}mo`;
    }

    const years = Math.floor(days / 365);
    return `${years}y`;

  } catch (error) {
    console.error('Error formatting timestamp:', error);
    return 'now';
  }
}

export function checkAuth(authState: IAuthStore, featureName: string): boolean {
  if (!authState.is_authenticated) {
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

export function generateFilePreview(file: File): IMedia {
  const url = URL.createObjectURL(file);
  let type: 'image' | 'video' | 'gif';
  
  if (file.type.startsWith('image/')) {
    type = file.type === 'image/gif' ? 'gif' : 'image';
  } else if (file.type.startsWith('video/')) {
    type = 'video';
  } else {
    // Default to image for unsupported types
    type = 'image';
  }

  return {
    url,
    type,
    alt_text: file.name
  };
}

export function processUserMetadata(content: string): { username?: string, name?: string, content: string } {
  if (!content) return { content: '' };

  const userMetadataRegex = /^\[USER:([^@\]]+)(?:@([^\]]+))?\](.*)/;
  const match = content.match(userMetadataRegex);

  if (match) {
    return {
      username: match[1] || undefined,
      name: match[2] || undefined,
      content: match[3] || ''
    };
  }

  return { content };
}

export function handleApiError(error: unknown): { success: false, message: string } {
  if (error instanceof Error) {
    if (error.name === 'AbortError') {
      return { success: false, message: 'Request timed out. Please try again.' };
    }
    return { success: false, message: error.message };
  }
  return { success: false, message: 'An unexpected error occurred.' };
}

export function truncateText(text: string, maxLength: number = 100): string {
  if (!text || text.length <= maxLength) return text;
  return text.substring(0, maxLength) + '...';
}

export function isSupabaseStorageUrl(url: string): boolean {
  const supabaseUrlPattern = /supabase\.co\/storage\/v1\/object\/public\/.+/;
  return supabaseUrlPattern.test(url);
}

export function formatStorageUrl(url: string | null): string {
  if (!url) return '';

  if (url.startsWith('http://') || url.startsWith('https://')) {
    console.log('URL already complete:', url);
    return url;
  }

  const supabaseUrl = import.meta.env.VITE_SUPABASE_URL || 'https://sdhtnvlmuywinhcglfsu.supabase.co';

  if (!url.includes('/')) {

    if (url.match(/\.(jpg|jpeg|png|gif|webp|svg)$/i)) {
      const formatted = `${supabaseUrl}/storage/v1/object/public/profile-pictures/${url}`;
      console.log(`Formatted filename-only URL (profile): ${url} -> ${formatted}`);
      return formatted;
    }

    const formatted = `${supabaseUrl}/storage/v1/object/public/tpaweb/${url}`;
    console.log(`Formatted filename-only URL: ${url} -> ${formatted}`);
    return formatted;
  }

  if (url.startsWith('storage/v1/object/public/')) {
    const formatted = `${supabaseUrl}/${url}`;
    console.log(`Formatted storage path URL: ${url} -> ${formatted}`);
    return formatted;
  }

  if (url.startsWith('profile-pictures/')) {
    const formatted = `${supabaseUrl}/storage/v1/object/public/${url}`;
    console.log(`Formatted profile URL: ${url} -> ${formatted}`);
    return formatted;
  }

  if (url.startsWith('banners/')) {
    const formatted = `${supabaseUrl}/storage/v1/object/public/${url}`;
    console.log(`Formatted banner URL: ${url} -> ${formatted}`);
    return formatted;
  }

  const formatted = `${supabaseUrl}/storage/v1/object/public/${url}`;
  console.log(`Formatted default URL: ${url} -> ${formatted}`);
  return formatted;
}