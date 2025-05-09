import { createClient } from '@supabase/supabase-js';
import { createLoggerWithPrefix } from './logger';

const logger = createLoggerWithPrefix('Supabase');

const supabaseUrl = 'https://sdhtnvlmuywinhcglfsu.supabase.co';
const supabaseAnonKey = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InNkaHRudmxtdXl3aW5oY2dsZnN1Iiwicm9sZSI6ImFub24iLCJpYXQiOjE3NDU5MDE4NzUsImV4cCI6MjA2MTQ3Nzg3NX0.Jknb2LNtRgma15sEX0sgLHMPegpCQ1f-05QbZEgHq8M';

export const supabase = createClient(supabaseUrl, supabaseAnonKey);

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
    
    const timestamp = new Date().getTime();
    const fileExtension = file.name.split('.').pop();
    const fileName = `${timestamp}_${Math.random().toString(36).substring(2, 10)}.${fileExtension}`;
    const filePath = `${folder}/${fileName}`;
    
    const { data, error } = await supabase.storage
      .from('media')
      .upload(filePath, file, {
        cacheControl: '3600',
        upsert: false
      });
    
    if (error) {
      logger.error('Supabase storage upload error:', error);
      throw error;
    }
    
    const { data: urlData } = supabase.storage
      .from('media')
      .getPublicUrl(filePath);
      
    if (!urlData.publicUrl) {
      throw new Error('Failed to get public URL');
    }
    
    return {
      url: urlData.publicUrl,
      mediaType: getMediaType(file.type)
    };
    
  } catch (error) {
    logger.error('Media upload failed:', error);
    return null;
  }
}

export async function deleteMedia(url: string): Promise<boolean> {
  try {
    const urlObj = new URL(url);
    const path = urlObj.pathname.split('/').slice(2).join('/');
    
    const { error } = await supabase.storage
      .from('media')
      .remove([path]);
    
    if (error) {
      logger.error('Supabase storage delete error:', error);
      throw error;
    }
    
    return true;
  } catch (error) {
    logger.error('Media deletion failed:', error);
    return false;
  }
}

export async function uploadFile(file: File, bucket: string, path: string): Promise<string | null> {
  console.log(`Attempting to upload file to bucket: ${bucket}, path: ${path}`);
  try {
    const fileExt = file.name.split('.').pop();
    const fileName = `${Date.now()}.${fileExt}`;
    const filePath = `${path}/${fileName}`;
    
    console.log(`Generated file path: ${filePath}`);
    
    const { data, error } = await supabase
      .storage
      .from(bucket)
      .upload(filePath, file, {
        cacheControl: '3600',
        upsert: false
      });
      
    if (error) {
      console.error('Error uploading file:', error);
      if (bucket === 'profile-pictures' || bucket === 'banners') {
        console.log(`Attempting upload to fallback bucket: tpaweb`);
        const fallbackResult = await supabase
          .storage
          .from('tpaweb')
          .upload(`${bucket}/${filePath}`, file, {
            cacheControl: '3600',
            upsert: false
          });
          
        if (fallbackResult.error) {
          console.error('Fallback upload also failed:', fallbackResult.error);
          return null;
        }
        
        const { data: fallbackUrlData } = supabase
          .storage
          .from('tpaweb')
          .getPublicUrl(`${bucket}/${filePath}`);
          
        console.log('Fallback upload successful, URL:', fallbackUrlData.publicUrl);
        return fallbackUrlData.publicUrl;
      }
      
      return null;
    }
    
    const { data: { publicUrl } } = supabase
      .storage
      .from(bucket)
      .getPublicUrl(filePath);
      
    console.log('Upload successful, URL:', publicUrl);
    return publicUrl;
  } catch (err) {
    console.error('Exception during file upload:', err);
    return null;
  }
}

export async function uploadProfilePicture(file: File, userId: string): Promise<string | null> {
  return uploadFile(file, 'profile-pictures', userId);
}

export async function uploadBanner(file: File, userId: string): Promise<string | null> {
  return uploadFile(file, 'banners', userId);
}

export async function deleteFile(bucket: string, path: string): Promise<boolean> {
  try {
    const { error } = await supabase
      .storage
      .from(bucket)
      .remove([path]);
      
    if (error) {
      console.error('Error deleting file:', error);
      return false;
    }
    
    return true;
  } catch (err) {
    console.error('Exception during file deletion:', err);
    return false;
  }
}