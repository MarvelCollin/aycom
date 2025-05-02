<script lang="ts">
  import { onMount } from 'svelte';
  import { useAuth } from '../hooks/useAuth';
  import AuthCallback from '../components/auth/AuthCallback.svelte';
  import type { IGoogleCredentialResponse } from '../interfaces/IAuth';
  import { toastStore } from '../stores/toastStore';

  // Get auth functions
  const { handleGoogleAuth } = useAuth();

  let loading = true;
  let error = "";
  let success = false;
  let redirectCountdown = 3;

  function startRedirectCountdown() {
    success = true;
    const interval = setInterval(() => {
      redirectCountdown--;
      if (redirectCountdown <= 0) {
        clearInterval(interval);
        window.location.href = '/feed';
      }
    }, 1000);
  }

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
      handleGoogleAuth({ credential } as IGoogleCredentialResponse)
        .then(result => {
          loading = false;
          if (result.success) {
            // Start countdown for redirect to feed page
            startRedirectCountdown();
          } else {
            error = result.message || 'Google authentication failed';
          }
        })
        .catch(err => {
          console.error('Error handling Google auth:', err);
          toastStore.showToast('Google authentication failed. Please try again.', 'error');
          error = 'An unexpected error occurred';
          loading = false;
        });
    } else {
      error = 'No authentication credential found';
      loading = false;
    }
  });
</script>

<div class="min-h-screen flex items-center justify-center bg-black text-white">
  <AuthCallback {loading} {error} {success} {redirectCountdown} />
</div>