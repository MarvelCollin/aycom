import { createClient } from '@supabase/supabase-js';
import { createLoggerWithPrefix } from './logger';

const logger = createLoggerWithPrefix('Supabase');

const supabaseUrl = import.meta.env.VITE_SUPABASE_URL || 'https://sdhtnvlmuywinhcglfsu.supabase.co';
const supabaseAnonKey = import.meta.env.VITE_SUPABASE_ANON_KEY || 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InNkaHRudmxtdXl3aW5oY2dsZnN1Iiwicm9sZSI6ImFub24iLCJpYXQiOjE3NDU5MDE4NzUsImV4cCI6MjA2MTQ3Nzg3NX0.Jknb2LNtRgma15sEX0sgLHMPegpCQ1f-05QbZEgHq8M';

export const supabase = createClient(supabaseUrl, supabaseAnonKey);

export const SUPABASE_BUCKETS = {
  MEDIA: 'media',
  PROFILES: 'profile-pictures',
  BANNERS: 'banners',
  THREAD_MEDIA: 'thread-media',
  USER_MEDIA: 'user-media',
  FALLBACK: 'tpaweb'
};

export async function initializeSupabaseBuckets(): Promise<void> {
  logger.info('Initializing Supabase buckets...');
  
  try {
    const { data: existingBuckets, error } = await supabase.storage.listBuckets();
    
    if (error) {
      logger.error('Failed to list Supabase buckets:', error);
      logger.info('Will use fallback bucket for all operations');
      return;
    }
    
    const existingBucketNames = existingBuckets.map(bucket => bucket.name);
    logger.info(`Available buckets: ${existingBucketNames.join(', ')}`);
    
    if (!existingBucketNames.includes(SUPABASE_BUCKETS.FALLBACK)) {
      logger.warn(`Fallback bucket '${SUPABASE_BUCKETS.FALLBACK}' does not exist. Please create it in the Supabase dashboard.`);
    }
    
    logger.info('Supabase buckets initialization complete');
  } catch (err) {
    logger.error('Error initializing Supabase buckets:', err);
    logger.info('Will use fallback bucket for all operations');
  }
}

const ALLOWED_MIME_TYPES = {
  image: ['image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/svg+xml'],
  video: ['video/mp4', 'video/webm', 'video/ogg'],
  audio: ['audio/mpeg', 'audio/ogg', 'audio/wav']
};

const MAX_FILE_SIZE = 10 * 1024 * 1024;

export function validateFile(file: File): { valid: boolean; error?: string } {
  if (file.size > MAX_FILE_SIZE) {
    return {
      valid: false,
      error: `File size exceeds maximum allowed (${MAX_FILE_SIZE / (1024 * 1024)}MB)`
    };
  }

  const allowedTypes = [
    ...ALLOWED_MIME_TYPES.image,
    ...ALLOWED_MIME_TYPES.video,
    ...ALLOWED_MIME_TYPES.audio
  ];

  if (!allowedTypes.includes(file.type)) {
    return {
      valid: false,
      error: 'File type not supported'
    };
  }

  return { valid: true };
}

export function getMediaType(mimeType: string): 'image' | 'video' | 'audio' | 'unknown' {
  if (ALLOWED_MIME_TYPES.image.includes(mimeType)) return 'image';
  if (ALLOWED_MIME_TYPES.video.includes(mimeType)) return 'video';
  if (ALLOWED_MIME_TYPES.audio.includes(mimeType)) return 'audio';
  return 'unknown';
}

export function generateUniqueFilename(file: File): string {
  const timestamp = Date.now();
  const randomString = Math.random().toString(36).substring(2, 10);
  const fileExt = file.name.split('.').pop();
  return `${timestamp}_${randomString}.${fileExt}`;
}

export async function uploadMedia(
  file: File, 
  folder: string = 'chat'
): Promise<{ url: string; mediaType: string } | null> {
  try {
    const validation = validateFile(file);
    if (!validation.valid) {
      logger.error('File validation failed:', validation.error);
      throw new Error(validation.error);
    }

    const fileName = generateUniqueFilename(file);
    const filePath = `${folder}/${fileName}`;
    let url: string;

    try {
      const { data, error } = await supabase.storage
        .from(SUPABASE_BUCKETS.MEDIA)
        .upload(filePath, file, {
          cacheControl: '3600',
          upsert: false
        });

      if (error) {
        throw error;
      }

      const { data: urlData } = supabase.storage
        .from(SUPABASE_BUCKETS.MEDIA)
        .getPublicUrl(filePath);

      url = urlData.publicUrl;
    } catch (uploadError) {
      logger.error('Primary bucket upload failed, trying fallback:', uploadError);
      
      // Try fallback bucket
      const fallbackPath = `media/${folder}/${fileName}`;
      const { data, error } = await supabase.storage
        .from(SUPABASE_BUCKETS.FALLBACK)
        .upload(fallbackPath, file, {
          cacheControl: '3600',
          upsert: false
        });

      if (error) {
        logger.error('Fallback bucket upload also failed:', error);
        throw error;
      }

      const { data: urlData } = supabase.storage
        .from(SUPABASE_BUCKETS.FALLBACK)
        .getPublicUrl(fallbackPath);

      url = urlData.publicUrl;
    }

    return {
      url,
      mediaType: getMediaType(file.type)
    };

  } catch (error) {
    logger.error('Media upload failed:', error);
    return null;
  }
}

export async function deleteMedia(url: string): Promise<boolean> {
  try {
    if (!isSupabaseStorageUrl(url)) {
      logger.warn('Not a Supabase URL, cannot delete:', url);
      return false;
    }

    const { bucket, path } = extractBucketAndPathFromUrl(url);
    if (!bucket || !path) {
      logger.error('Failed to extract path from URL:', url);
      return false;
    }

    const { error } = await supabase.storage
      .from(bucket)
      .remove([path]);

    if (error) {
      logger.error(`Supabase storage delete error for ${bucket}/${path}:`, error);
      throw error;
    }

    return true;
  } catch (error) {
    logger.error('Media deletion failed:', error);
    return false;
  }
}

export function extractBucketAndPathFromUrl(url: string): { bucket: string | null; path: string | null } {
  try {
    const urlObj = new URL(url);
    const pathParts = urlObj.pathname.split('/');

    const publicIndex = pathParts.indexOf('public');
    if (publicIndex === -1 || publicIndex + 1 >= pathParts.length) {
      return { bucket: null, path: null };
    }

    const bucket = pathParts[publicIndex + 1];
    const path = pathParts.slice(publicIndex + 2).join('/');

    return { bucket, path };
  } catch (error) {
    logger.error('Failed to parse Supabase URL:', error);
    return { bucket: null, path: null };
  }
}

export function isSupabaseStorageUrl(url: string): boolean {
  try {
    const urlObj = new URL(url);
    return urlObj.hostname.includes('supabase.co') && 
           urlObj.pathname.includes('/storage/v1/object/public/');
  } catch (error) {
    return false;
  }
}

export async function uploadFile(file: File, bucket: string, path: string): Promise<string | null> {
  logger.debug(`Uploading to bucket: ${bucket}, path: ${path}`);
  try {
    const validation = validateFile(file);
    if (!validation.valid) {
      logger.error('File validation failed:', validation.error);
      return null;
    }

    const fileName = generateUniqueFilename(file);
    const filePath = `${path}/${fileName}`;

    const { data, error } = await supabase.storage
      .from(bucket)
      .upload(filePath, file, {
        cacheControl: '3600',
        upsert: false
      });

    if (error) {
      logger.error(`Error uploading to ${bucket}:`, error);

      // Use fallback bucket for any upload that fails
      logger.debug(`Attempting upload to fallback bucket: ${SUPABASE_BUCKETS.FALLBACK}`);
      
      const fallbackPath = `${bucket}/${path}/${fileName}`;
      const fallbackResult = await supabase.storage
        .from(SUPABASE_BUCKETS.FALLBACK)
        .upload(fallbackPath, file, {
          cacheControl: '3600',
          upsert: false
        });
        
      if (fallbackResult.error) {
        logger.error(`Error uploading to fallback bucket:`, fallbackResult.error);
        return null;
      }
      
      const { data: fallbackUrlData } = supabase.storage
        .from(SUPABASE_BUCKETS.FALLBACK)
        .getPublicUrl(fallbackPath);
        
      return fallbackUrlData.publicUrl;
    }

    const { data: urlData } = supabase.storage
      .from(bucket)
      .getPublicUrl(filePath);

    return urlData.publicUrl;
  } catch (error) {
    logger.error('File upload failed:', error);
    return null;
  }
}

// Set this to true once all bucket-specific functions have been updated
export const allBucketFunctionsUpdated = true;

export async function uploadProfilePicture(file: File, userId: string): Promise<string | null> {
  return uploadFile(file, SUPABASE_BUCKETS.PROFILES, userId);
}

export async function uploadBanner(file: File, userId: string): Promise<string | null> {
  return uploadFile(file, SUPABASE_BUCKETS.BANNERS, userId);
}

export async function uploadThreadMedia(file: File, threadId: string): Promise<string | null> {
  return uploadFile(file, SUPABASE_BUCKETS.THREAD_MEDIA, threadId);
}

export async function uploadMultipleFiles(
  files: File[], 
  bucket: string, 
  path: string
): Promise<string[]> {
  const urls: string[] = [];
  
  for (const file of files) {
    const url = await uploadFile(file, bucket, path);
    if (url) {
      urls.push(url);
    }
  }
  
  return urls;
}

export async function uploadMultipleThreadMedia(
  files: File[], 
  threadId: string
): Promise<string[]> {
  return uploadMultipleFiles(files, SUPABASE_BUCKETS.THREAD_MEDIA, threadId);
}

export async function deleteFile(bucket: string, path: string): Promise<boolean> {
  try {
    const { error } = await supabase.storage
      .from(bucket)
      .remove([path]);

    if (error) {
      logger.error('Error deleting file:', error);
      return false;
    }

    return true;
  } catch (err) {
    logger.error('Exception during file deletion:', err);
    return false;
  }
}

export function getPublicUrl(bucket: string, path: string): string | null {
  try {
    const { data } = supabase.storage
      .from(bucket)
      .getPublicUrl(path);

    return data.publicUrl;
  } catch (error) {
    logger.error('Failed to get public URL:', error);
    return null;
  }
}