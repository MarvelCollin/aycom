import { toastStore } from '../stores/toastStore';
import type { IAuthStore } from '../interfaces/IAuth';
import type { IMedia } from '../interfaces/ISocialMedia';
import { createLoggerWithPrefix } from './logger';

const logger = createLoggerWithPrefix('common-utils');

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
  const supabaseUrlPattern = /supabase\.co\/storage\/v1\/object\/public\//;
  return supabaseUrlPattern.test(url);
}

export function formatStorageUrl(url: string | null): string {
  if (!url) return '';
  
  if (url.startsWith('http://') || url.startsWith('https://')) {
    return url;
  }
  
  const supabaseUrl = import.meta.env.VITE_SUPABASE_URL || 'https://sdhtnvlmuywinhcglfsu.supabase.co';
  return `${supabaseUrl}/storage/v1/object/public/${url}`;
} 