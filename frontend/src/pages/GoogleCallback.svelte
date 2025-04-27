<script lang="ts">
  import { onMount } from 'svelte';
  import { useAuth } from '../hooks/useAuth';

  // Get auth functions
  const { handleGoogleAuth } = useAuth();

  let loading = true;
  let error = "";

  onMount(() => {
    // Get the credential from various potential sources
    // 1. From URL query parameter
    const params = new URLSearchParams(window.location.search);
    // 2. From URL hash (fragment)
    const hashParams = new URLSearchParams(window.location.hash.substring(1));
    // 3. From state object if redirected
    const state = window.history.state;
    
    // Try to get the credential from any available source
    const credential = 
      params.get('credential') || 
      hashParams.get('credential') || 
      (state && state.credential);
    
    if (credential) {
      // Process the Google credential
      handleGoogleAuth({ credential })
        .then(result => {
          if (result.success) {
            // Redirect to dashboard on success
            window.location.href = '/dashboard';
          } else {
            error = result.message || 'Google authentication failed';
            loading = false;
          }
        })
        .catch(err => {
          console.error('Error handling Google auth:', err);
          error = 'An unexpected error occurred';
          loading = false;
        });
    } else {
      error = 'No authentication credential found';
      loading = false;
    }
  });
</script>

<div class="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
  <div class="max-w-md w-full p-6 bg-white dark:bg-gray-800 rounded-lg shadow-lg">
    <div class="text-center">
      <h1 class="text-2xl font-semibold text-gray-900 dark:text-white mb-4">
        Google Authentication
      </h1>
      
      {#if loading}
        <div class="mb-4">
          <p class="text-gray-600 dark:text-gray-400">
            Processing your login...
          </p>
          <div class="mt-4 flex justify-center">
            <div class="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
          </div>
        </div>
      {:else if error}
        <div class="mb-4">
          <p class="text-red-500">
            {error}
          </p>
          <a href="/login" class="mt-4 inline-block text-blue-500 hover:underline">
            Return to login
          </a>
        </div>
      {/if}
    </div>
  </div>
</div> 