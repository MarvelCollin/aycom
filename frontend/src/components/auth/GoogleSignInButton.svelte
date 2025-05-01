<script lang="ts">
  import { onMount } from 'svelte';
  import { useExternalServices } from '../../hooks/useExternalServices';
  import { useAuth } from '../../hooks/useAuth';
  import { useTheme } from '../../hooks/useTheme';
  import type { IGoogleCredentialResponse } from '../../interfaces/IAuth';

  export let onAuthSuccess: (result: any) => void = () => {};
  export let onAuthError: (error: string) => void = () => {};
  export let containerId = 'google-signin-button';
  
  // Forward additional HTML attributes for testing
  let buttonClass = '';
  export { buttonClass as class };
  
  // Get external services functions
  const { loadGoogleAuth } = useExternalServices();
  
  // Get auth functions
  const { handleGoogleAuth } = useAuth();
  
  // Get theme store
  const { theme } = useTheme();
  
  // Reactive declaration to update isDarkMode when theme changes
  $: isDarkMode = $theme === 'dark';
  
  // Google authentication handler
  async function handleGoogleCredentialResponse(response: IGoogleCredentialResponse) {
    try {
      const result = await handleGoogleAuth(response);
      
      if (result.success) {
        onAuthSuccess(result);
      } else {
        onAuthError(result.message || 'Google authentication failed');
      }
    } catch (err) {
      console.error('Error handling Google auth:', err);
      onAuthError('An unexpected error occurred');
    }
  }

  // For testing purposes, we need a manual trigger
  function handleManualSignIn() {
    // This is just for testing - in real use, the Google button will handle this
    console.log('Manual Google Sign-In clicked');
  }
  
  onMount(() => {
    // Load Google Sign-In
    return loadGoogleAuth(
      containerId, 
      isDarkMode, 
      handleGoogleCredentialResponse
    );
  });
</script>

<div class="w-full mb-4">
  <!-- Regular Google Sign-In Button Container -->
  <div id={containerId} class="min-h-[40px] {buttonClass}"></div>
  
  <!-- Fallback button for testing purposes -->
  <button
    class="w-full py-2 px-4 border border-gray-300 dark:border-gray-700 rounded-full flex items-center justify-center gap-2 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
    on:click={handleManualSignIn}
    data-cy="google-signin-button-test"
  >
    <svg width="18" height="18" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48">
      <path fill="#EA4335" d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z"/>
      <path fill="#4285F4" d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6c4.51-4.18 7.09-10.36 7.09-17.65z"/>
      <path fill="#FBBC05" d="M10.53 28.59c-.48-1.45-.76-2.99-.76-4.59s.27-3.14.76-4.59l-7.98-6.19C.92 16.46 0 20.12 0 24c0 3.88.92 7.54 2.56 10.78l7.97-6.19z"/>
      <path fill="#34A853" d="M24 48c6.48 0 11.93-2.13 15.89-5.81l-7.73-6c-2.15 1.45-4.92 2.3-8.16 2.3-6.26 0-11.57-4.22-13.47-9.91l-7.98 6.19C6.51 42.62 14.62 48 24 48z"/>
    </svg>
    <span>Sign in with Google</span>
  </button>
</div> 