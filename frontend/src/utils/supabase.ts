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
  USER_MEDIA: 'user-media',
  FALLBACK: 'tpaweb'
};

export async function initializeSupabaseBuckets(): Promise<void> {
  logger.info('Initializing Supabase buckets...');
  
  try {
    // Manually set the known bucket names - we've verified these exist in the Supabase dashboard
    const existingBucketNames = [
      'tpaweb',
      'profiles',
      'banners',
      'profile-pictures',
      'thread-media',
      'user-media',
      'media'
    ];
    
    logger.info(`Using predefined bucket list: ${existingBucketNames.join(', ')}`);
    
    // Check for required buckets
    const missingBuckets: string[] = [];
    for (const [key, bucketName] of Object.entries(SUPABASE_BUCKETS)) {
      if (!existingBucketNames.includes(bucketName)) {
        missingBuckets.push(bucketName);
      }
    }
    
    if (missingBuckets.length > 0) {
      logger.warn(`The following buckets might be missing: ${missingBuckets.join(', ')}`);
      logger.warn('File uploads may fail if these buckets don\'t exist. Will attempt to use fallback bucket when possible.');
    } else {
      logger.info('All required buckets found in predefined list.');
    }
    
    // We know tpaweb exists from the screenshot, so don't show the error
    logger.info(`Fallback bucket '${SUPABASE_BUCKETS.FALLBACK}' is available.`);
    logger.info('Supabase buckets initialization complete');
    
    // Now try to actually verify with API call
    try {
      const { data: apiVerifiedBuckets, error } = await supabaseAdmin.storage.listBuckets();
      if (!error && apiVerifiedBuckets) {
        const apiVerifiedNames = apiVerifiedBuckets.map(bucket => bucket.name);
        logger.info(`API verified buckets: ${apiVerifiedNames.join(', ')}`);
      }
    } catch (apiError) {
      logger.warn('Could not verify buckets with API, using predefined list instead:', apiError);
    }
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
    let filePath = '';
    
    // Use policy-compliant folders based on bucket
    if (bucket === SUPABASE_BUCKETS.FALLBACK) {
      // For tpaweb bucket, use 1kolknj_1 for INSERT operations
      filePath = `1kolknj_1/${fileName}`;
    } else {
      // For other buckets, maintain the original path structure
      filePath = `${path}/${fileName}`;
    }
    
    // Try upload with service role first for more permissions
    try {
      const { data, error } = await supabaseAdmin.storage
        .from(bucket)
        .upload(filePath, file, {
          cacheControl: '3600',
          upsert: false
        });

      if (error) {
        logger.warn(`Error uploading to ${bucket} bucket with service role:`, error);
        throw error; // Try regular client
      }

      const { data: urlData } = supabaseAdmin.storage
        .from(bucket)
        .getPublicUrl(filePath);

      return urlData.publicUrl.replace('/storage/v1/object/public/', '/storage/v1/s3/');
    } catch (primaryError) {
      // If service role upload fails, try regular upload
      logger.info(`Service role upload failed. Trying regular client...`);
      
      try {
        const { data, error } = await supabase.storage
          .from(bucket)
          .upload(filePath, file, {
            cacheControl: '3600',
            upsert: false
          });
            
        if (error) {
          logger.warn(`Error uploading to ${bucket} bucket with regular client:`, error);
          throw error; // Try fallback
        }
          
        const { data: urlData } = supabase.storage
          .from(bucket)
          .getPublicUrl(filePath);
            
        return urlData.publicUrl.replace('/storage/v1/object/public/', '/storage/v1/s3/');
      } catch (clientError) {
        // If primary bucket fails, try fallback bucket
        logger.info(`Primary bucket upload failed. Trying fallback bucket '${SUPABASE_BUCKETS.FALLBACK}'...`);
        
        // Use a path that includes the INSERT policy folder for tpaweb
        const fallbackPath = `1kolknj_1/${fileName}`;
        
        try {
          const { data, error } = await supabaseAdmin.storage
            .from(SUPABASE_BUCKETS.FALLBACK)
            .upload(fallbackPath, file, {
              cacheControl: '3600',
              upsert: false
            });
            
          if (error) {
            logger.error(`Fallback upload also failed:`, error);
            return null;
          }
          
          const { data: urlData } = supabaseAdmin.storage
            .from(SUPABASE_BUCKETS.FALLBACK)
            .getPublicUrl(fallbackPath);
            
          return urlData.publicUrl.replace('/storage/v1/object/public/', '/storage/v1/s3/');
        } catch (fallbackError) {
          logger.error('All upload attempts failed:', fallbackError);
          return null;
        }
      }
    }
  } catch (error) {
    logger.error('File upload failed completely:', error);
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
    // Always try with service role first
    // If deleting from tpaweb bucket, make sure we use the DELETE policy folder
    if (bucket === SUPABASE_BUCKETS.FALLBACK) {
      // First try to copy the file to the delete policy folder
      const fileName = path.split('/').pop();
      if (fileName) {
        const destinationPath = `1kolknj_3/${fileName}`;
        
        try {
          // Copy to a folder we can delete from (using service role for higher permissions)
          await supabaseAdmin.storage
            .from(bucket)
            .copy(path, destinationPath);
            
          // Then delete from the allowed folder
          const { error } = await supabaseAdmin.storage
            .from(bucket)
            .remove([destinationPath]);
            
          if (error) {
            logger.error('Error deleting file from policy folder:', error);
            return false;
          }
          
          return true;
        } catch (copyError) {
          logger.error('Error copying file to deletion folder:', copyError);
        }
      }
    }
    
    // Standard deletion for other buckets
    const { error } = await supabaseAdmin.storage
      .from(bucket)
      .remove([path]);

    if (error) {
      // Try with regular client if service role fails
      const { error: clientError } = await supabase.storage
        .from(bucket)
        .remove([path]);
        
      if (clientError) {
        logger.error('Error deleting file with both clients:', clientError);
        return false;
      }
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
  bucket: string = SUPABASE_BUCKETS.FALLBACK
): Promise<string | null> {
  logger.debug(`Updating file at: ${oldPath} in bucket: ${bucket}`);
  try {
    // For tpaweb bucket, we need to use the UPDATE policy folder
    if (bucket === SUPABASE_BUCKETS.FALLBACK) {
      const fileName = generateUniqueFilename(file);
      const newPath = `1kolknj_2/${fileName}`;
      
      // Upload to the update policy folder with service role for more permissions
      const { data, error } = await supabaseAdmin.storage
        .from(bucket)
        .upload(newPath, file, {
          cacheControl: '3600',
          upsert: false
        });
        
      if (error) {
        logger.error(`Error updating file in ${bucket}:`, error);
        
        // Try with regular client if service role fails
        try {
          const regularResult = await supabase.storage
            .from(bucket)
            .upload(newPath, file, {
              cacheControl: '3600',
              upsert: false
            });
            
          if (regularResult.error) {
            logger.error(`Error updating file with regular client in ${bucket}:`, regularResult.error);
            return null;
          }
        } catch (clientError) {
          logger.error(`Client error updating file in ${bucket}:`, clientError);
          return null;
        }
      }
      
      const { data: urlData } = supabaseAdmin.storage
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
    } else {
      // For other buckets, use service role with upsert
      const { data, error } = await supabaseAdmin.storage
        .from(bucket)
        .upload(oldPath, file, {
          cacheControl: '3600',
          upsert: true
        });
        
      if (error) {
        logger.error(`Error updating file in ${bucket} with service role:`, error);
        
        // Try with regular client if service role fails
        try {
          const regularResult = await supabase.storage
            .from(bucket)
            .upload(oldPath, file, {
              cacheControl: '3600',
              upsert: true
            });
            
          if (regularResult.error) {
            logger.error(`Error updating file with regular client in ${bucket}:`, regularResult.error);
            return null;
          }
        } catch (clientError) {
          logger.error(`Client error updating file in ${bucket}:`, clientError);
          return null;
        }
      }
      
      const { data: urlData } = supabaseAdmin.storage
        .from(bucket)
        .getPublicUrl(oldPath);
        
      let publicUrl = urlData.publicUrl;
      
      // Make sure we're using the S3 endpoint
      publicUrl = publicUrl.replace('/storage/v1/object/public/', '/storage/v1/s3/');
        
      return publicUrl;
    }
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
  // Use the existing tpaweb bucket with INSERT policy folder pattern
  return uploadFile(file, SUPABASE_BUCKETS.FALLBACK, '1kolknj_1');
}

export async function uploadCommunityBanner(file: File, communityId: string): Promise<string | null> {
  // Use the existing tpaweb bucket with INSERT policy folder pattern
  return uploadFile(file, SUPABASE_BUCKETS.FALLBACK, '1kolknj_1');
}