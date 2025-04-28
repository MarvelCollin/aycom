<script lang="ts">
  export let verificationCode = "";
  export let showResendOption = false;
  export let timeLeft = "";
  export let onVerify: () => void;
  export let onResend: () => void;
</script>

<div class="mb-6">
  <label for="verificationCode" class="block text-sm font-medium mb-1">Verification code</label>
  <input 
    type="text" 
    id="verificationCode" 
    bind:value={verificationCode} 
    class="w-full p-2 border border-gray-600 rounded bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500"
    placeholder="Verification code"
    data-cy="verification-code"
  />
</div>

{#if !showResendOption}
  <p class="text-sm text-center mb-4" data-cy="verification-timer">Code expires in {timeLeft}</p>
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