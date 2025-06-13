<script lang="ts">
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import AuthLayout from '../components/layout/AuthLayout.svelte';
  import GoogleSignInButton from '../components/auth/GoogleSignInButton.svelte';
  import { toastStore } from '../stores/toastStore';
  import appConfig from '../config/appConfig';
  import ReCaptchaWrapper from '../components/auth/ReCaptchaWrapper.svelte';
  import { getAuthToken, clearAuthData } from '../utils/auth';
  import DebugPanel from '../components/common/DebugPanel.svelte';
  import { onMount } from 'svelte';
  import Toast from '../components/common/Toast.svelte';
  import ProfileCompletion from '../components/auth/ProfileCompletion.svelte';

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
  
  // New state for Google auth profile completion
  let showProfileCompletion = false;
  let missingProfileFields: string[] = [];
  
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
    
    // Clear any existing auth data to ensure we get fresh tokens
    clearAuthData();
  });
  
  async function handleSubmit() {
    isLoading = true;
    error = "";
    
    try {
      console.log(`Submitting login form for email: ${email}`);
      
      const trimmedEmail = email.trim();
      
      const result = await login(trimmedEmail, password);
      isLoading = false;
      
      if (result.success) {
        toastStore.showToast('Login successful!', 'success');
        setTimeout(() => {
          const currentPath = window.location.pathname;
          if (currentPath !== '/feed') {
            console.log('Login successful, redirecting to feed');
            window.location.href = '/feed';
          }
        }, 100);
      } else {
        const errorMessage = result.message || "Login failed. Please check your credentials.";
        console.error('Login failed with message:', errorMessage);
        error = errorMessage; 
        toastStore.showToast(errorMessage, 'error');
      }
    } catch (err) {
      isLoading = false;
      const message = err instanceof Error ? err.message : 'Login failed. Please try again.';
      console.error("Login Exception:", err);
      error = message;
      toastStore.showToast(message, 'error');
    }
  }
  
  interface AuthResult {
    success: boolean;
    message?: string;
    [key: string]: any;
  }
  
  function handleGoogleAuthSuccess(result: AuthResult) {
    console.log('Google auth success in Login page');
    
    if (result.requires_profile_completion && result.missing_fields?.length > 0) {
      console.log('User needs to complete profile information:', result.missing_fields);
      missingProfileFields = result.missing_fields;
      showProfileCompletion = true;
    } else {
      toastStore.showToast('Google login successful', 'success');
      window.location.href = '/feed';
    }
  }
  
  function handleGoogleAuthError(message: string) {
    console.error('Google auth error in Login page:', message);
    error = message; 
    if (appConfig.ui.showErrorToasts) {
      toastStore.showToast(`Google Auth Error: ${message}`, 'error');
    }
  }

  function handleProfileCompleted() {
    console.log('Profile completion successful');
    toastStore.showToast('Profile updated successfully', 'success');
    window.location.href = '/feed';
  }

  function handleProfileSkipped() {
    console.log('Profile completion skipped');
    toastStore.showToast('You can complete your profile later in account settings', 'info');
    window.location.href = '/feed';
  }
</script>

<AuthLayout 
  title={showProfileCompletion ? "Complete Your Profile" : "Sign in to AYCOM"}
  showBackButton={showProfileCompletion}
  onBack={() => showProfileCompletion = false}
>
  {#if !showProfileCompletion}
    <div class="auth-social-btn-container" data-cy="google-login-button">
      <GoogleSignInButton 
        onAuthSuccess={handleGoogleAuthSuccess} 
        onAuthError={handleGoogleAuthError}
        class="auth-social-btn {isDarkMode ? 'auth-social-btn-dark' : ''}"
      />
    </div>
    
    <div class="auth-divider {isDarkMode ? 'auth-divider-dark' : ''}">
      <span class="auth-divider-text">or sign in with email</span>
    </div>
    
    {#if error}
      <div class="bg-red-500 bg-opacity-10 border border-red-500 text-red-500 px-4 py-3 rounded mb-6" data-cy="error-message">
        {error}
      </div>
    {/if}
    
    <form on:submit|preventDefault={handleSubmit} class="mb-6">
      <div class="auth-input-group">
        <label for="email" class="auth-label">Email Address</label>
        <input 
          type="email" 
          id="email" 
          bind:value={email} 
          class="auth-input {isDarkMode ? 'auth-input-dark' : ''}"
          placeholder="Enter your email address"
          data-cy="email-input"
        />
      </div>
      
      <div class="auth-input-group">
        <div class="flex justify-between items-center mb-2">
          <label for="password" class="auth-label">Password</label>
          <a href="/forgot-password" class="auth-forgot-password" data-cy="forgot-password">Forgot password?</a>
        </div>
        <input 
          type="password" 
          id="password" 
          bind:value={password} 
          class="auth-input {isDarkMode ? 'auth-input-dark' : ''}"
          placeholder="Enter your password"
          data-cy="password-input"
        />
      </div>
      
      <div class="auth-checkbox-group">
        <input 
          type="checkbox" 
          id="remember-me"
          bind:checked={rememberMe} 
          class="auth-checkbox"
          data-cy="remember-me"
        />
        <label for="remember-me" class="auth-checkbox-label">Keep me signed in</label>
      </div>
      
      <button 
        type="submit"
        class="auth-btn"
        disabled={isLoading}
        data-cy="login-button"
      >
        {#if isLoading}
          <svg class="animate-spin h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          Signing in...
        {:else}
          Sign In to Your Account
        {/if}
      </button>
    </form>
    
    <div class="auth-footer">
      Don't have an account? <a href="/register" class="auth-link" data-cy="register-link">Create one now</a>
    </div>
  {:else}
    <ProfileCompletion 
      missingFields={missingProfileFields} 
      onComplete={handleProfileCompleted}
      onSkip={handleProfileSkipped}
    />
  {/if}
</AuthLayout>

<DebugPanel />

<Toast />
