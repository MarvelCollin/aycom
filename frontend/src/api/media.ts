import { uploadMedia as uploadMediaToStorage } from '../utils/supabase';
import appConfig from '../config/appConfig';
import { getAuthToken } from '../utils/auth';
import { createLoggerWithPrefix } from '../utils/logger';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('MediaAPI');

/**
 * Upload media file to the server
 * @param file File to upload
 * @param folder Optional folder destination
 * @returns Promise with URL and media type or null on failure
 */
export async function uploadMedia(file: File, folder: string = 'chat'): Promise<{url: string; mediaType: string} | null> {
  try {
    logger.debug(`Uploading media file ${file.name} to ${folder}`);
    
    // Create a FormData object to send the file
    const formData = new FormData();
    formData.append('file', file);
    formData.append('folder', folder);
    
    // Get auth token
    const token = getAuthToken();
    if (!token) {
      throw new Error('Authentication token not found');
    }
    
    // Call the backend API endpoint instead of Supabase directly
    const response = await fetch(`${API_BASE_URL}/media`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
      },
      body: formData,
      credentials: 'include',
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => null);
      throw new Error(errorData?.message || `Upload failed with status: ${response.status}`);
    }
    
    const data = await response.json();
    
    if (!data.success && !data.url) {
      throw new Error('Upload response missing url');
    }
    
    // Return the media info
    return {
      url: data.url || data.data?.url,
      mediaType: data.type || data.data?.type || getMediaTypeFromFile(file)
    };
  } catch (error) {
    logger.error(`Failed to upload media: ${error}`);
    
    // Fallback to direct Supabase upload (if configured)
    logger.debug('Attempting fallback to direct storage upload');
    try {
      return await uploadMediaToStorage(file, folder);
    } catch (fallbackError) {
      logger.error(`Fallback upload also failed: ${fallbackError}`);
      return null;
    }
  }
}

/**
 * Determine media type from file
 */
function getMediaTypeFromFile(file: File): string {
  const type = file.type.split('/')[0];
  if (type === 'image') {
    if (file.type === 'image/gif') {
      return 'gif';
    }
    return 'image';
  } else if (type === 'video') {
    return 'video';
  }
  return 'image'; // Default fallback
} 