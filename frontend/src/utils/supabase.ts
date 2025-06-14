import { createClient } from '@supabase/supabase-js';
import { createLoggerWithPrefix } from './logger';

const logger = createLoggerWithPrefix('Supabase');

// Base Supabase configuration
const supabaseUrl = import.meta.env.VITE_SUPABASE_URL || 'https://sdhtnvlmuywinhcglfsu.supabase.co';
const supabaseAnonKey = import.meta.env.VITE_SUPABASE_ANON_KEY || 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InNkaHRudmxtdXl3aW5oY2dsZnN1Iiwicm9sZSI6ImFub24iLCJpYXQiOjE3NDU5MDE4NzUsImV4cCI6MjA2MTQ3Nzg3NX0.Jknb2LNtRgma15sEX0sgLHMPegpCQ1f-05QbZEgHq8M';
const supabaseServiceRoleKey = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InNkaHRudmxtdXl3aW5oY2dsZnN1Iiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc0NTkwMTg3NSwiZXhwIjoyMDYxNDc3ODc1fQ.tWJLnDSGqG4SL1uPTt_g-4FYlxbzfDDUC8x4IFHwV3k';

// Create regular client with anon key
export const supabase = createClient(supabaseUrl, supabaseAnonKey, {
  auth: {
    persistSession: true,
    autoRefreshToken: true
  }
});

// Create admin client with service role for operations requiring higher privileges
export const supabaseAdmin = createClient(supabaseUrl, supabaseServiceRoleKey, {
  auth: {
    persistSession: false,
    autoRefreshToken: false
  }
});

// Storage endpoint
export const STORAGE_ENDPOINT = `${supabaseUrl}/storage/v1/s3`;

export const SUPABASE_BUCKETS = {
  MEDIA: 'media',
  PROFILES: 'profile-pictures',
  BANNERS: 'banners',
  THREAD_MEDIA: 'thread-media',
  USER_MEDIA: 'user-media'
};

export async function initializeSupabaseBuckets(): Promise<void> {
  logger.info('Initializing Supabase buckets...');
  
  try {
    // Get list of existing buckets
    const { data: existingBuckets, error } = await supabase.storage.listBuckets();
    
    if (error) {
      logger.error('Failed to list Supabase buckets:', error);
      return;
    }
    
    if (!existingBuckets) {
      logger.warn('No buckets returned from Supabase');
      return;
    }
    
    const existingBucketNames = existingBuckets.map(bucket => bucket.name);
    logger.debug(`Found existing buckets: ${existingBucketNames.join(', ')}`);
    
    // Check if our required buckets exist
    const missingBuckets: string[] = [];
    for (const [key, bucketName] of Object.entries(SUPABASE_BUCKETS)) {
      if (!existingBucketNames.includes(bucketName)) {
        missingBuckets.push(bucketName);
      }
    }
    
    if (missingBuckets.length > 0) {
      logger.warn(`The following buckets might be missing: ${missingBuckets.join(', ')}`);
      logger.warn('File uploads may fail if these buckets don\'t exist.');
    } else {
      logger.info('All required buckets found in predefined list.');
    }
    
    logger.info('Supabase buckets initialization complete');
    
    // Now try to actually verify with API call
    try {
      const { data: apiVerifiedBuckets, error } = await supabase.storage.listBuckets();
      if (!error && apiVerifiedBuckets) {
        const apiVerifiedNames = apiVerifiedBuckets.map(bucket => bucket.name);
        logger.info(`API verified buckets: ${apiVerifiedNames.join(', ')}`);
      }
    } catch (apiError) {
      logger.warn('Failed to verify buckets with API call:', apiError);
    }
    
  } catch (e) {
    logger.error('Error initializing Supabase buckets:', e);
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
      return null;
    }
    
    const fileName = generateUniqueFilename(file);
    const filePath = `${folder}/${fileName}`;
    let url = '';
    
    // Upload to the media bucket
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
      logger.error('Media upload failed:', uploadError);
      return null;
    }
    
    return {
      url,
      mediaType: getMediaType(file.type)
    };
  } catch (error) {
    logger.error('Error in uploadMedia:', error);
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
    // Maintain the original path structure
    const filePath = `${path}/${fileName}`;
    
    // Upload directly to the specified bucket
    const { data, error } = await supabase.storage
      .from(bucket)
      .upload(filePath, file, {
        cacheControl: '3600',
        upsert: false
      });

    if (error) {
      logger.error(`Upload failed for bucket ${bucket}:`, error);
      return null;
    }

    // Get the public URL
    const { data: urlData } = supabase.storage
      .from(bucket)
      .getPublicUrl(filePath);

    // Return the S3 endpoint URL
    return urlData.publicUrl.replace('/storage/v1/object/public/', '/storage/v1/s3/');
  } catch (error) {
    logger.error('File upload failed:', error);
    return null;
  }
}

// Set this to true once all bucket-specific functions have been updated
export const allBucketFunctionsUpdated = true;

export async function uploadProfilePicture(file: File, userId: string): Promise<string | null> {
  // For profile pictures, we'll try the profile-pictures bucket first
  // but if that fails, use the tpaweb bucket with INSERT policy folder
  return uploadFile(file, SUPABASE_BUCKETS.PROFILES, userId);
}

export async function uploadBanner(file: File, userId: string): Promise<string | null> {
  // For banners, we'll try the banners bucket first
  // but if that fails, use the tpaweb bucket with INSERT policy folder
  return uploadFile(file, SUPABASE_BUCKETS.BANNERS, userId);
}

export async function uploadThreadMedia(file: File, threadId: string): Promise<string | null> {
  // For thread media, we'll try the thread-media bucket first
  // but if that fails, use the tpaweb bucket with INSERT policy folder
  return uploadFile(file, SUPABASE_BUCKETS.THREAD_MEDIA, threadId);
}

export async function uploadUserMedia(file: File, userId: string): Promise<string | null> {
  // For user media, we'll try the user-media bucket first
  // but if that fails, use the tpaweb bucket with INSERT policy folder
  return uploadFile(file, SUPABASE_BUCKETS.USER_MEDIA, userId);
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

export async function updateSupabaseFile(
  file: File, 
  oldPath: string, 
  bucket: string
): Promise<string | null> {
  logger.debug(`Updating file at: ${oldPath} in bucket: ${bucket}`);
  try {
    const fileName = generateUniqueFilename(file);
    const newPath = `${oldPath.split('/')[0]}/${fileName}`;
    
    // Upload to the bucket
    const { data, error } = await supabase.storage
      .from(bucket)
      .upload(newPath, file, {
        cacheControl: '3600',
        upsert: false
      });
      
    if (error) {
      logger.error(`Error updating file in ${bucket}:`, error);
      return null;
    }
    
    const { data: urlData } = supabase.storage
      .from(bucket)
      .getPublicUrl(newPath);
    
    let publicUrl = urlData.publicUrl;
    
    // Make sure we're using the S3 endpoint
    publicUrl = publicUrl.replace('/storage/v1/object/public/', '/storage/v1/s3/');
      
    // Try to delete the old file (but don't fail if it doesn't work)
    try {
      await deleteFile(bucket, oldPath);
    } catch (deleteError) {
      logger.warn(`Couldn't delete old file after update: ${oldPath}`, deleteError);
    }
      
    return publicUrl;
  } catch (error) {
    logger.error('File update failed:', error);
    return null;
  }
}

export async function uploadMultipleUserMedia(
  files: File[], 
  userId: string
): Promise<string[]> {
  return uploadMultipleFiles(files, SUPABASE_BUCKETS.USER_MEDIA, userId);
}

export async function uploadCommunityLogo(file: File, communityId: string): Promise<string | null> {
  return uploadFile(file, SUPABASE_BUCKETS.PROFILES, communityId);
}

export async function uploadCommunityBanner(file: File, communityId: string): Promise<string | null> {
  return uploadFile(file, SUPABASE_BUCKETS.BANNERS, communityId);
}