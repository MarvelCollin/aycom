<script lang="ts">
  import { useTheme } from "../../hooks/useTheme";

  export let verificationCode = "";
  export let showResendOption = false;
  export let timeLeft = "";
  export let onVerify: () => void;
  export let onResend: () => void;

  const { theme } = useTheme();

  $: isDarkMode = $theme === "dark";
</script>

<div class="auth-input-group">
  <label for="verificationCode" class="auth-label">Verification code</label>
  <input
    type="text"
    id="verificationCode"
    bind:value={verificationCode}
    class="auth-input {isDarkMode ? "auth-input-dark" : ""}"
    placeholder="Verification code"
    data-cy="verification-code-input"
  />
</div>

{#if !showResendOption}
  <p class="verification-timer" data-cy="resend-timer">Code expires in {timeLeft}</p>
{/if}

{#if showResendOption}
  <button
    class="auth-link block text-center mb-4 w-full"
    on:click={onResend}
    data-cy="resend-button"
  >
    Didn't receive email?
  </button>
{/if}

<button
  class="auth-btn"
  on:click={onVerify}
  data-cy="verify-button"
>
  Next
</button>