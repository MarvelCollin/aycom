const { createClient } = require('@supabase/supabase-js');

const supabaseUrl = 'https://sdhtnvlmuywinhcglfsu.supabase.co';
const supabaseKey = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InNkaHRudmxtdXl3aW5oY2dsZnN1Iiwicm9sZSI6ImFub24iLCJpYXQiOjE3NDU5MDE4NzUsImV4cCI6MjA2MTQ3Nzg3NX0.Jknb2LNtRgma15sEX0sgLHMPegpCQ1f-05QbZEgHq8M';
const supabase = createClient(supabaseUrl, supabaseKey);

async function testSupabaseConnection() {
  try {
    console.log('Testing Supabase connection...');
    
    // List buckets
    const { data: buckets, error: bucketsError } = await supabase.storage.listBuckets();
    if (bucketsError) {
      console.error('Error listing buckets:', bucketsError);
    } else {
      console.log('Buckets:', buckets);
      
      // Test tpaweb bucket specifically
      if (buckets.some(b => b.name === 'tpaweb')) {
        console.log('tpaweb bucket exists, testing access...');
        
        // List files in tpaweb bucket
        const { data: files, error: filesError } = await supabase.storage
          .from('tpaweb')
          .list();
          
        if (filesError) {
          console.error('Error listing files in tpaweb bucket:', filesError);
        } else {
          console.log('Files in tpaweb bucket:', files);
        }
      }
    }
  } catch (error) {
    console.error('Unexpected error:', error);
  }
}

testSupabaseConnection(); 