<script lang="ts">
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import AuthLayout from '../components/layout/AuthLayout.svelte';
  import GoogleSignInButton from '../components/auth/GoogleSignInButton.svelte';
  import { toastStore } from '../stores/toastStore';
  import appConfig from '../config/appConfig';
  import ReCaptchaWrapper from '../components/auth/ReCaptchaWrapper.svelte';
  import { getAuthToken } from '../utils/auth';
  import DebugPanel from '../components/common/DebugPanel.svelte';
  import { onMount } from 'svelte';

  const { login } = useAuth();
  
  const { theme } = useTheme();
  
  $: isDarkMode = $theme === 'dark';
  
  let email = "";
  let password = "";
  let rememberMe = false;
  let error = "";
  let isLoading = false;
  let recaptchaToken: string | null = null;
  let recaptchaWrapper: ReCaptchaWrapper;
  
  function handleRecaptchaSuccess(event: CustomEvent<{ token: string }>) {
    recaptchaToken = event.detail.token;
  }

  function handleRecaptchaError() {
    recaptchaToken = null;
  }

  function handleRecaptchaExpired() {
    recaptchaToken = null;
  }

  onMount(() => {
    if (getAuthToken()) {
      window.location.href = '/feed';
    }
  });
  
  async function handleSubmit() {
    let errorMessage = "";
    if (!email || !password) {
      errorMessage = "Please enter both email and password";
      error = errorMessage;
      if (appConfig.ui.showErrorToasts) toastStore.showToast(errorMessage);
      return;
    }
    // if (!recaptchaToken) {
    //   error = 'Please complete the reCAPTCHA verification.';
    //   return;
    // }
    isLoading = true;
    error = "";
    
    try {
      const result = await login(email, password);
      isLoading = false;
      
      if (result.success) {
        // Add a small delay to ensure auth state is fully updated before redirect
        setTimeout(() => {
          window.location.href = '/feed';
        }, 100);
      } else {
        errorMessage = result.message || "Login failed. Please check your credentials.";
        error = errorMessage; 
        toastStore.showToast(errorMessage, 'error');
      }
    } catch (err) {
      isLoading = false;
      console.error("Login Exception:", err);
      toastStore.showToast('Login failed. Please try again.', 'error');
    }
  }
  
  interface AuthResult {
    success: boolean;
    message?: string;
    [key: string]: any;
  }
  
  function handleGoogleAuthSuccess(result: AuthResult) {
    window.location.href = '/feed';
  }
  
  function handleGoogleAuthError(message: string) {
    error = message; 
    if (appConfig.ui.showErrorToasts) {
      toastStore.showToast(`Google Auth Error: ${message}`);
    }
  }
</script>

<AuthLayout title="Sign in to AYCOM">
  <div data-cy="google-login-button">
    <GoogleSignInButton 
      onAuthSuccess={handleGoogleAuthSuccess} 
      onAuthError={handleGoogleAuthError}
    />
  </div>
  
  <div class="flex items-center mb-4">
    <div class="flex-grow h-px bg-gray-300 dark:bg-gray-700"></div>
    <span class="px-2 text-sm text-gray-500 dark:text-gray-400">or</span>
    <div class="flex-grow h-px bg-gray-300 dark:bg-gray-700"></div>
  </div>
  
  {#if error}
    <div class="bg-red-500 bg-opacity-10 border border-red-500 text-red-500 px-4 py-3 rounded mb-4" data-cy="error-message">
      {error}
    </div>
  {/if}
  
  <form on:submit|preventDefault={handleSubmit} class="mb-4">
    <div class="mb-4">
      <label for="email" class="block text-sm font-medium mb-1">Email</label>
      <input 
        type="email" 
        id="email" 
        bind:value={email} 
        class="w-full p-2 border {isDarkMode ? 'border-gray-700 bg-gray-800' : 'border-gray-300 bg-white'} rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
        placeholder="Email"
        required
        data-cy="email-input"
      />
      {#if error && !email}
        <p class="text-red-500 text-sm mt-1" data-cy="email-error">Email is required</p>
      {/if}
    </div>
    
    <div class="mb-6">
      <div class="flex justify-between items-center mb-1">
        <label for="password" class="block text-sm font-medium">Password</label>
        <a href="/forgot-password" class="text-xs text-blue-500 hover:underline" data-cy="forgot-password">Forgot password?</a>
      </div>
      <input 
        type="password" 
        id="password" 
        bind:value={password} 
        class="w-full p-2 border {isDarkMode ? 'border-gray-700 bg-gray-800' : 'border-gray-300 bg-white'} rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
        placeholder="Password"
        required
        data-cy="password-input"
      />
      {#if error && !password}
        <p class="text-red-500 text-sm mt-1" data-cy="password-error">Password is required</p>
      {/if}
    </div>
    
    <div class="flex items-center justify-between mb-6">
      <label class="flex items-center">
        <input 
          type="checkbox" 
          bind:checked={rememberMe} 
          class="mr-2"
          data-cy="remember-me"
        />
        <span class="text-sm">Remember me</span>
      </label>
    </div>
    
    <!-- <div class="mb-6 hidden">
      <ReCaptchaWrapper
        bind:this={recaptchaWrapper}
        siteKey="6Ld6UysrAAAAAPW3XRLe-M9bGDgOPJ2kml1yCozA"
        theme={isDarkMode ? 'dark' : 'light'}
        on:success={handleRecaptchaSuccess}
        on:error={handleRecaptchaError}
        on:expired={handleRecaptchaExpired}
      />
    </div> -->
    <!-- {#if error === 'Please complete the reCAPTCHA verification.'}
      <p class="text-red-500 text-xs mt-1 text-center" data-cy="recaptcha-error">{error}</p>
    {/if} -->
    
    <div class="mb-6 flex justify-between">
      <a href="/" class="text-xs text-blue-500 hover:underline" data-cy="landing-link">Back to Landing Page</a>
      <a href="/forgot-password" class="text-xs text-blue-500 hover:underline" data-cy="forgot-password">Forgot password?</a>
    </div>
    
    <button 
      type="submit"
      class="w-full py-3 bg-blue-500 text-white text-center rounded-full font-semibold hover:bg-blue-600 transition-colors flex justify-center items-center"
      disabled={isLoading}
      data-cy="login-button"
    >
      {#if isLoading}
        <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        Signing in...
      {:else}
        Sign in
      {/if}
    </button>
  </form>
  
  <p class="text-sm mt-6 text-center">
    Don't have an account? <a href="/register" class="text-blue-500 hover:underline" data-cy="register-link">Sign up</a>
  </p>
</AuthLayout>

<DebugPanel />
