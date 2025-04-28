<script lang="ts">
  import { onMount } from 'svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import Logo from '../components/common/Logo.svelte';

  // Get auth functions
  const { handleGoogleAuth } = useAuth();
  // Get theme store
  const { theme } = useTheme();

  // Reactive declaration to update isDarkMode when theme changes
  $: isDarkMode = $theme === 'dark';

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
      handleGoogleAuth({ credential })
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
  <div class="max-w-md w-full p-6 bg-gray-900 rounded-lg shadow-lg border border-gray-800">
    <div class="text-center">
      <!-- AYCOM Logo/Branding -->
      <div class="mb-6 flex justify-center">
        <Logo size="medium" />
      </div>
      
      <h1 class="text-2xl font-semibold text-white mb-4">
        Google Authentication
      </h1>
      
      {#if loading}
        <div class="mb-6">
          <p class="text-gray-400 mb-4">
            Processing your login...
          </p>
          <div class="flex justify-center">
            <div class="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
          </div>
        </div>
      {:else if success}
        <div class="mb-6 text-center">
          <div class="mb-4 flex justify-center">
            <div class="bg-green-500 rounded-full w-16 h-16 flex items-center justify-center text-white text-2xl">
              ✓
            </div>
          </div>
          <p class="text-green-400 text-xl font-semibold mb-2">Authentication Successful!</p>
          <p class="text-gray-400">Redirecting to your feed in {redirectCountdown} second{redirectCountdown !== 1 ? 's' : ''}...</p>
        </div>
      {:else if error}
        <div class="mb-6 text-center">
          <div class="mb-4 flex justify-center">
            <div class="bg-red-500 rounded-full w-16 h-16 flex items-center justify-center text-white text-2xl">
              ✕
            </div>
          </div>
          <p class="text-red-500 mb-4">
            {error}
          </p>
          <a 
            href="/login" 
            class="mt-4 inline-block px-6 py-2 bg-blue-500 text-white rounded-full hover:bg-blue-600 transition"
          >
            Return to login
          </a>
        </div>
      {/if}
      
      <p class="text-gray-500 text-sm mt-8">
        &copy; {new Date().getFullYear()} AYCOM. All rights reserved.
      </p>
    </div>
  </div>
</div>