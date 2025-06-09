import { uploadMedia as uploadMediaToStorage } from '../utils/supabase';

/**
 * Upload a media file to the storage and return the URL
 * @param file File to upload
 * @param folder Optional folder path within storage
 * @returns Object containing URL and media type, or null if upload failed
 */
export async function uploadMedia(file: File, folder: string = 'general'): Promise<{url: string; mediaType: string} | null> {
  return uploadMediaToStorage(file, folder);
} 