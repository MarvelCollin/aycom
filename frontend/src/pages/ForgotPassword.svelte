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
  if (!recaptchaToken) {
    emailError = 'Please complete the reCAPTCHA verification.';
    return;
  }
  isLoading = true;
  try {
    const result = await getSecurityQuestion(email, recaptchaToken);
    securityQuestion = result.securityQuestion;
    oldPasswordHash = result.oldPasswordHash;
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
    await verifySecurityAnswer(email, securityAnswer);
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
    const result = await resetPassword(email, newPassword);
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

<AuthLayout title="Forgotten Account">
  {#if step === 1}
    <form on:submit|preventDefault={handleEmailSubmit} class="mb-4">
      <div class="mb-4">
        <label for="email" class="block text-sm font-medium mb-1">Email</label>
        <input type="email" id="email" bind:value={email} class="w-full p-2 border {isDarkMode ? 'border-gray-700 bg-gray-800' : 'border-gray-300 bg-white'} rounded focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="Email" required />
        {#if emailError}
          <p class="text-red-500 text-xs mt-1">{emailError}</p>
        {/if}
      </div>
      <div class="mb-6 hidden">
        <ReCaptchaWrapper
          bind:this={recaptchaWrapper}
          siteKey="6Ld6UysrAAAAAPW3XRLe-M9bGDgOPJ2kml1yCozA"
          theme={isDarkMode ? 'dark' : 'light'}
          on:success={handleRecaptchaSuccess}
          on:error={handleRecaptchaError}
          on:expired={handleRecaptchaExpired}
        />
      </div>
      <button type="submit" class="w-full py-3 bg-blue-500 text-white rounded-full font-semibold hover:bg-blue-600 transition-colors" disabled={isLoading}>Next</button>
    </form>
    <a href="/" class="text-xs text-blue-500 hover:underline">Back to Landing Page</a>
  {:else if step === 2}
    <form on:submit|preventDefault={handleAnswerSubmit} class="mb-4">
      <div class="mb-4">
        <label class="block text-sm font-medium mb-1">Security Question</label>
        <div class="mb-2">{securityQuestion}</div>
        <input type="text" bind:value={securityAnswer} class="w-full p-2 border {isDarkMode ? 'border-gray-700 bg-gray-800' : 'border-gray-300 bg-white'} rounded focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="Your answer" required />
        {#if answerError}
          <p class="text-red-500 text-xs mt-1">{answerError}</p>
        {/if}
      </div>
      <button type="submit" class="w-full py-3 bg-blue-500 text-white rounded-full font-semibold hover:bg-blue-600 transition-colors" disabled={isLoading}>Next</button>
    </form>
    <a href="/" class="text-xs text-blue-500 hover:underline">Back to Landing Page</a>
  {:else if step === 3}
    <form on:submit|preventDefault={handlePasswordSubmit} class="mb-4">
      <div class="mb-4">
        <label for="newPassword" class="block text-sm font-medium mb-1">New Password</label>
        <input type="password" id="newPassword" bind:value={newPassword} class="w-full p-2 border {isDarkMode ? 'border-gray-700 bg-gray-800' : 'border-gray-300 bg-white'} rounded focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="New Password" required />
        {#if newPasswordError}
          <p class="text-red-500 text-xs mt-1">{newPasswordError}</p>
        {/if}
      </div>
      <button type="submit" class="w-full py-3 bg-blue-500 text-white rounded-full font-semibold hover:bg-blue-600 transition-colors" disabled={isLoading}>Reset Password</button>
    </form>
    <a href="/" class="text-xs text-blue-500 hover:underline">Back to Landing Page</a>
  {/if}
</AuthLayout> 