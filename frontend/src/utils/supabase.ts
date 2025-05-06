import { createClient } from '@supabase/supabase-js';

// Get Supabase URL and anon key from environment variables
const supabaseUrl = import.meta.env.VITE_SUPABASE_URL || 'https://sdhtnvlmuywinhcglfsu.supabase.co';
const supabaseAnonKey = import.meta.env.VITE_SUPABASE_ANON_KEY || 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InNkaHRudmxtdXl3aW5oY2dsZnN1Iiwicm9sZSI6ImFub24iLCJpYXQiOjE3NDU5MDE4NzUsImV4cCI6MjA2MTQ3Nzg3NX0.Jknb2LNtRgma15sEX0sgLHMPegpCQ1f-05QbZEgHq8M';

// Initialize Supabase client
export const supabase = createClient(supabaseUrl, supabaseAnonKey);

/**
 * Upload a file to Supabase storage
 * @param file The file to upload
 * @param bucket The storage bucket name
 * @param path Path within the bucket
 * @returns URL of the uploaded file or null if failed
 */
export async function uploadFile(file: File, bucket: string, path: string): Promise<string | null> {
  try {
    // Generate a unique filename with timestamp and original extension
    const fileExt = file.name.split('.').pop();
    const fileName = `${Date.now()}.${fileExt}`;
    const filePath = `${path}/${fileName}`;
    
    // Upload the file
    const { data, error } = await supabase
      .storage
      .from(bucket)
      .upload(filePath, file, {
        cacheControl: '3600',
        upsert: false
      });
      
    if (error) {
      console.error('Error uploading file:', error);
      return null;
    }
    
    // Get the public URL of the file
    const { data: { publicUrl } } = supabase
      .storage
      .from(bucket)
      .getPublicUrl(filePath);
      
    return publicUrl;
  } catch (err) {
    console.error('Exception during file upload:', err);
    return null;
  }
}

/**
 * Upload a profile picture to Supabase storage
 * @param file The image file to upload
 * @param userId User ID to associate with the file
 * @returns URL of the uploaded profile picture or null if failed
 */
export async function uploadProfilePicture(file: File, userId: string): Promise<string | null> {
  return uploadFile(file, 'profiles', userId);
}

/**
 * Upload a banner image to Supabase storage
 * @param file The image file to upload
 * @param userId User ID to associate with the file
 * @returns URL of the uploaded banner or null if failed
 */
export async function uploadBanner(file: File, userId: string): Promise<string | null> {
  return uploadFile(file, 'banners', userId);
}

/**
 * Delete a file from Supabase storage
 * @param bucket The storage bucket name
 * @param path Full path to the file within the bucket
 * @returns Boolean indicating success or failure
 */
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