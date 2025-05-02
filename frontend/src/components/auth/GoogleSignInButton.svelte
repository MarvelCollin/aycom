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
  
  const { loadGoogleAuth } = useExternalServices();
  
  const { handleGoogleAuth } = useAuth();
  
  const { theme } = useTheme();
  
  $: isDarkMode = $theme === 'dark';
  
  async function handleGoogleCredentialResponse(response: IGoogleCredentialResponse) {
    try {
      const result = await handleGoogleAuth(response);
      
      if (result.success) {
        onAuthSuccess(result);
      } else {
        onAuthError(result.message || 'Google authentication failed');
      }
    } catch (err) {
      onAuthError('An unexpected error occurred');
    }
  }
  
  onMount(() => {
    return loadGoogleAuth(
      containerId, 
      isDarkMode, 
      handleGoogleCredentialResponse
    );
  });
</script>

<div class="w-full mb-4">
  <!-- Google Sign In Button  -->
  <div id={containerId} class="min-h-[40px] {buttonClass}"></div>
  
</div> 