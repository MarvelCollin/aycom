<script lang="ts">
import { onMount } from 'svelte';
import { useTheme } from '../hooks/useTheme';
import AuthLayout from '../components/layout/AuthLayout.svelte';
import ReCaptchaWrapper from '../components/auth/ReCaptchaWrapper.svelte';
import { toastStore } from '../stores/toastStore';
import { getAuthToken } from '../utils/auth';
import { getSecurityQuestion, verifySecurityAnswer, resetPassword } from '../api/passwordReset';
import { handleApiError } from '../utils/common';

let step = 1;
let email = '';
let emailError = '';
let securityQuestion = '';
let securityAnswer = '';
let answerError = '';
let newPassword = '';
let newPasswordError = '';
let oldPasswordHash = '';
let resetToken = '';
let recaptchaToken: string | null = null;
let recaptchaWrapper: ReCaptchaWrapper;
let isLoading = false;
const { theme } = useTheme();
$: isDarkMode = $theme === 'dark';

onMount(() => {
  if (getAuthToken()) {
    window.location.href = '/feed';
  }
});

async function handleEmailSubmit() {
  emailError = '';
  if (!email) {
    emailError = 'Email is required';
    return;
  }
  isLoading = true;
  try {
    const result = await getSecurityQuestion(email, recaptchaToken);
    securityQuestion = result.securityQuestion;
    oldPasswordHash = result.oldPasswordHash;
    email = result.email; // Update email from response in case it was normalized
    step = 2;
  } catch (error) {
    const errorResponse = handleApiError(error);
    emailError = errorResponse.message;
  } finally {
    isLoading = false;
  }
}

async function handleAnswerSubmit() {
  answerError = '';
  if (!securityAnswer) {
    answerError = 'Answer is required';
    return;
  }
  isLoading = true;
  try {
    const result = await verifySecurityAnswer(email, securityAnswer);
    resetToken = result.token;
    step = 3;
  } catch (error) {
    const errorResponse = handleApiError(error);
    answerError = errorResponse.message;
  } finally {
    isLoading = false;
  }
}

async function handlePasswordSubmit() {
  newPasswordError = '';
  if (!newPassword) {
    newPasswordError = 'Password is required';
    return;
  }
  if (newPassword === oldPasswordHash) {
    newPasswordError = 'New password cannot be the same as the old password.';
    return;
  }
  isLoading = true;
  try {
    const result = await resetPassword(email, newPassword, resetToken);
    toastStore.showToast(result.message || 'Password reset successful. Please login.', 'success');
    window.location.href = '/login';
  } catch (error) {
    const errorResponse = handleApiError(error);
    newPasswordError = errorResponse.message;
  } finally {
    isLoading = false;
  }
}

function handleRecaptchaSuccess(event: CustomEvent<{ token: string }>) {
  recaptchaToken = event.detail.token;
}
function handleRecaptchaError() {
  recaptchaToken = null;
}
function handleRecaptchaExpired() {
  recaptchaToken = null;
}
</script>

<AuthLayout title="Recover your password">
  {#if step === 1}
    <form on:submit|preventDefault={handleEmailSubmit} class="mb-4">
      <div class="auth-input-group">
        <label for="email" class="auth-label">Email</label>
        <input 
          type="email" 
          id="email" 
          bind:value={email} 
          class="auth-input {isDarkMode ? 'auth-input-dark' : ''} {emailError ? 'auth-input-error' : ''}" 
          placeholder="Enter your account email" 
          required 
        />
        {#if emailError}
          <p class="auth-error-message">{emailError}</p>
        {/if}
      </div>
      
      <div class="auth-input-group">
        <ReCaptchaWrapper
          bind:this={recaptchaWrapper}
          siteKey="6Ld6UysrAAAAAPW3XRLe-M9bGDgOPJ2kml1yCozA"
          theme={isDarkMode ? 'dark' : 'light'}
          on:success={handleRecaptchaSuccess}
          on:error={handleRecaptchaError}
          on:expired={handleRecaptchaExpired}
        />
      </div>
      
      <button 
        type="submit" 
        class="auth-btn"
        disabled={isLoading}
      >
        {isLoading ? 'Processing...' : 'Continue'}
      </button>
    </form>
    
    <div class="auth-footer">
      <a href="/login" class="auth-link">Back to login</a>
    </div>
  {:else if step === 2}
    <form on:submit|preventDefault={handleAnswerSubmit} class="mb-4">
      <div class="auth-input-group">
        <label class="auth-label">Security Question</label>
        <div class="mb-2 text-sm">{securityQuestion}</div>
        <input 
          type="text" 
          bind:value={securityAnswer} 
          class="auth-input {isDarkMode ? 'auth-input-dark' : ''} {answerError ? 'auth-input-error' : ''}" 
          placeholder="Your answer" 
          required 
        />
        {#if answerError}
          <p class="auth-error-message">{answerError}</p>
        {/if}
      </div>
      
      <button 
        type="submit" 
        class="auth-btn"
        disabled={isLoading}
      >
        {isLoading ? 'Verifying...' : 'Verify Answer'}
      </button>
    </form>
    
    <div class="auth-footer">
      <a href="/login" class="auth-link">Back to login</a>
    </div>
  {:else if step === 3}
    <form on:submit|preventDefault={handlePasswordSubmit} class="mb-4">
      <div class="auth-input-group">
        <label for="newPassword" class="auth-label">New Password</label>
        <input 
          type="password" 
          id="newPassword" 
          bind:value={newPassword} 
          class="auth-input {isDarkMode ? 'auth-input-dark' : ''} {newPasswordError ? 'auth-input-error' : ''}" 
          placeholder="Enter new password" 
          required 
        />
        {#if newPasswordError}
          <p class="auth-error-message">{newPasswordError}</p>
        {/if}
      </div>
      
      <button 
        type="submit" 
        class="auth-btn"
        disabled={isLoading}
      >
        {isLoading ? 'Resetting...' : 'Reset Password'}
      </button>
    </form>
    
    <div class="auth-footer">
      <a href="/login" class="auth-link">Back to login</a>
    </div>
  {/if}
</AuthLayout> 