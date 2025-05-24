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
  import Toast from '../components/common/Toast.svelte';

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
    const token = getAuthToken();
    console.log('Login page - Auth token exists:', !!token, 'Current path:', window.location.pathname);
    
  });
  
  async function handleSubmit() {
    let errorMessage = "";
    if (!email || !password) {
      errorMessage = "Please enter both email and password";
      error = errorMessage;
      if (appConfig.ui.showErrorToasts) toastStore.showToast(errorMessage);
      return;
    }
    
    isLoading = true;
    error = "";
    
    try {
      const result = await login(email, password);
      isLoading = false;
      
      if (result.success) {
        setTimeout(() => {
          const currentPath = window.location.pathname;
          if (currentPath !== '/feed') {
            console.log('Login successful, redirecting to feed');
            window.location.href = '/feed';
          }
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
    console.log('Google auth success in Login page');
    toastStore.showToast('Google login successful', 'success');
    window.location.href = '/feed';
  }
  
  function handleGoogleAuthError(message: string) {
    console.error('Google auth error in Login page:', message);
    error = message; 
    if (appConfig.ui.showErrorToasts) {
      toastStore.showToast(`Google Auth Error: ${message}`, 'error');
    }
  }
</script>

<AuthLayout title="Sign in to AYCOM">
  <div class="auth-social-btn-container" data-cy="google-login-button">
    <GoogleSignInButton 
      onAuthSuccess={handleGoogleAuthSuccess} 
      onAuthError={handleGoogleAuthError}
      class="auth-social-btn {isDarkMode ? 'auth-social-btn-dark' : ''}"
    />
  </div>
  
  <div class="auth-divider {isDarkMode ? 'auth-divider-dark' : ''}">
    <span class="auth-divider-text">or</span>
  </div>
  
  {#if error}
    <div class="bg-red-500 bg-opacity-10 border border-red-500 text-red-500 px-4 py-3 rounded mb-4" data-cy="error-message">
      {error}
    </div>
  {/if}
  
  <form on:submit|preventDefault={handleSubmit} class="mb-4">
    <div class="auth-input-group">
      <label for="email" class="auth-label">Email</label>
      <input 
        type="email" 
        id="email" 
        bind:value={email} 
        class="auth-input {isDarkMode ? 'auth-input-dark' : ''} {error && !email ? 'auth-input-error' : ''}"
        placeholder="Email"
        required
        data-cy="email-input"
      />
      {#if error && !email}
        <p class="auth-error-message" data-cy="email-error">Email is required</p>
      {/if}
    </div>
    
    <div class="auth-input-group">
      <div class="flex justify-between items-center mb-1">
        <label for="password" class="auth-label">Password</label>
        <a href="/forgot-password" class="auth-forgot-password" data-cy="forgot-password">Forgot password?</a>
      </div>
      <input 
        type="password" 
        id="password" 
        bind:value={password} 
        class="auth-input {isDarkMode ? 'auth-input-dark' : ''} {error && !password ? 'auth-input-error' : ''}"
        placeholder="Password"
        required
        data-cy="password-input"
      />
      {#if error && !password}
        <p class="auth-error-message" data-cy="password-error">Password is required</p>
      {/if}
    </div>
    
    <div class="auth-checkbox-group">
      <input 
        type="checkbox" 
        id="remember-me"
        bind:checked={rememberMe} 
        class="auth-checkbox"
        data-cy="remember-me"
      />
      <label for="remember-me" class="auth-checkbox-label">Remember me</label>
    </div>
    
    
    <button 
      type="submit"
      class="auth-btn"
      disabled={isLoading}
      data-cy="login-button"
    >
      {#if isLoading}
        <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white inline-block" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        Signing in...
      {:else}
        Sign in
      {/if}
    </button>
  </form>
  
  <div class="auth-footer">
    Don't have an account? <a href="/register" class="auth-link" data-cy="register-link">Sign up</a>
  </div>
</AuthLayout>

<DebugPanel />

<Toast />
