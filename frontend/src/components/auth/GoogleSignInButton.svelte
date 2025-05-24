<script lang="ts">
  import { onMount } from 'svelte';
  import { useExternalServices } from '../../hooks/useExternalServices';
  import { useAuth } from '../../hooks/useAuth';
  import { useTheme } from '../../hooks/useTheme';
  import type { IGoogleCredentialResponse } from '../../interfaces/IAuth';

  export let onAuthSuccess: (result: any) => void = () => {};
  export let onAuthError: (error: string) => void = () => {};
  export let containerId = 'google-signin-button';
  
  let buttonClass = '';
  export { buttonClass as class };
  
  let isLoading = true;
  let loadError = false;
  let errorMessage = '';
  
  const { loadGoogleAuth } = useExternalServices();
  
  const { handleGoogleAuth } = useAuth();
  
  const { theme } = useTheme();
  
  $: isDarkMode = $theme === 'dark';
  
  async function handleGoogleCredentialResponse(response: IGoogleCredentialResponse) {
    console.log('Handling Google credential response');
    isLoading = true;
    loadError = false;
    
    try {
      if (!response || !response.credential) {
        console.error('Invalid Google credential response:', response);
        errorMessage = 'Invalid response from Google';
        loadError = true;
        onAuthError('Invalid response from Google');
        return;
      }
      
      const result = await handleGoogleAuth(response);
      
      if (result.success) {
        console.log('Google auth successful, invoking success callback');
        onAuthSuccess(result);
      } else {
        console.error('Google auth failed:', result.message);
        errorMessage = result.message || 'Google authentication failed';
        loadError = true;
        onAuthError(result.message || 'Google authentication failed');
      }
    } catch (err) {
      console.error('Exception during Google auth:', err);
      errorMessage = err instanceof Error ? err.message : 'An unexpected error occurred';
      loadError = true;
      onAuthError('An unexpected error occurred');
    } finally {
      isLoading = false;
    }
  }
  
  onMount(() => {
    console.log('GoogleSignInButton mounted, loading Google Auth...');
    isLoading = true;
    loadError = false;
    
    const cleanup = loadGoogleAuth(
      containerId, 
      isDarkMode, 
      handleGoogleCredentialResponse
    );
    
    const timeoutId = setTimeout(() => {
      const buttonElement = document.getElementById(containerId);
      if (buttonElement && buttonElement.children.length === 0) {
        console.warn('Google button did not render within timeout period');
        loadError = true;
        errorMessage = 'Google Sign-In could not be loaded. Please try again later.';
      }
      isLoading = false;
    }, 5000);
    
    return () => {
      clearTimeout(timeoutId);
      cleanup();
    };
  });
</script>

<div class="w-full mb-4 relative">
  <div id={containerId} class="min-h-[40px] {buttonClass}"></div>
  
  {#if isLoading}
    <div class="absolute inset-0 flex items-center justify-center bg-gray-100 dark:bg-gray-800 bg-opacity-50 dark:bg-opacity-50 rounded">
      <div class="w-5 h-5 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
    </div>
  {/if}
  
  {#if loadError}
    <div class="mt-2 text-center text-sm text-red-500">
      {errorMessage}
    </div>
  {/if}
</div> 