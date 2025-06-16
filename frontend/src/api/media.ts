import { uploadMedia as uploadMediaToStorage } from '../utils/supabase';

export async function uploadMedia(file: File, folder: string = 'general'): Promise<{url: string; mediaType: string} | null> {
  return uploadMediaToStorage(file, folder);
} 