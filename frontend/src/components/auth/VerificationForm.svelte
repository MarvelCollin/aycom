<script lang="ts">
  import { useTheme } from '../../hooks/useTheme';
  
  export let verificationCode = "";
  export let showResendOption = false;
  export let timeLeft = "";
  export let onVerify: () => void;
  export let onResend: () => void;
  
  // Get theme
  const { theme } = useTheme();
  
  // Reactive declaration to update isDarkMode when theme changes
  $: isDarkMode = $theme === 'dark';
</script>

<div class="mb-6">
  <label for="verificationCode" class="block text-sm font-medium mb-1">Verification code</label>
  <input 
    type="text" 
    id="verificationCode" 
    bind:value={verificationCode} 
    class="w-full p-2 border {isDarkMode ? 'border-gray-700 bg-gray-800' : 'border-gray-300 bg-white'} rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
    placeholder="Verification code"
    data-cy="verification-code"
  />
</div>

{#if !showResendOption}
  <p class="text-sm text-center mb-4 text-gray-600 dark:text-gray-400" data-cy="verification-timer">Code expires in {timeLeft}</p>
{/if}

{#if showResendOption}
  <button 
    class="w-full text-blue-500 hover:underline mb-4 text-center"
    on:click={onResend}
    data-cy="resend-button"
  >
    Didn't receive email?
  </button>
{/if}

<button 
  class="w-full py-3 bg-blue-500 text-white text-center rounded-full font-semibold hover:bg-blue-600 transition-colors"
  on:click={onVerify}
  data-cy="verify-button"
>
  Next
</button> 