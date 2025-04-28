<script lang="ts">
  import { onMount } from 'svelte';
  import { useExternalServices } from '../../hooks/useExternalServices';
  import { useAuth } from '../../hooks/useAuth';
  import { useTheme } from '../../hooks/useTheme';
  import type { IGoogleCredentialResponse } from '../../interfaces/IAuth';

  export let onAuthSuccess: (result: any) => void = () => {};
  export let onAuthError: (error: string) => void = () => {};
  export let containerId = 'google-signin-button';
  
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
  
  onMount(() => {
    // Load Google Sign-In
    return loadGoogleAuth(
      containerId, 
      isDarkMode, 
      handleGoogleCredentialResponse
    );
  });
</script>

<div id={containerId} class="w-full mb-4 min-h-[40px]"></div> 