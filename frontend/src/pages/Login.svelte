<script lang="ts">
  import { onMount } from 'svelte';
  import { useTheme } from '../hooks/useTheme';
  import { useAuth } from '../hooks/useAuth';
  import { useExternalServices } from '../hooks/useExternalServices';
  import type { GoogleCredentialResponse } from '../interfaces/auth';
  
  // Get the theme store and toggleTheme function from our hook
  const { theme } = useTheme();
  
  // Get auth functions from auth hook
  const { login, handleGoogleAuth } = useAuth();
  
  // Get external services functions
  const { loadGoogleAuth } = useExternalServices();
  
  // Reactive declaration to update isDarkMode when theme changes
  $: isDarkMode = $theme === 'dark';
  
  // Login state
  let email = "";
  let password = "";
  let rememberMe = false;
  let error = "";
  let isLoading = false;
  
  // Handle login form submission
  async function handleSubmit() {
    if (!email || !password) {
      error = "Please enter both email and password";
      return;
    }
    
    isLoading = true;
    error = "";
    
    const result = await login(email, password);
    
    isLoading = false;
    
    if (result.success) {
      // Redirect to dashboard
      window.location.href = '/dashboard';
    } else {
      error = result.message || "Login failed. Please check your credentials.";
    }
  }
  
  // Google authentication handler
  async function handleGoogleCredentialResponse(response: GoogleCredentialResponse) {
    isLoading = true;
    error = "";
    
    const result = await handleGoogleAuth(response);
    
    isLoading = false;
    
    if (result.success) {
      // Redirect to dashboard
      window.location.href = '/dashboard';
    } else {
      error = result.message || "Google authentication failed";
    }
  }
  
  onMount(() => {
    // Load Google Sign-In
    const googleCleanup = loadGoogleAuth(
      'google-signin-button', 
      isDarkMode, 
      handleGoogleCredentialResponse
    );
    
    // Return cleanup function
    return () => {
      googleCleanup();
    };
  });
</script>

<div class="{isDarkMode ? 'bg-black text-white' : 'bg-white text-black'} min-h-screen w-full flex justify-center items-center p-4">
  <div class="w-full max-w-md bg-dark-900 rounded-lg shadow-lg p-6">
    <div class="flex items-center justify-between mb-6">
      <a href="/" class="text-blue-500 hover:text-blue-600 transition-colors">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </a>
      
      <div class="mx-auto">
        <img 
          src={isDarkMode ? "/src/assets/logo/light-logo.jpeg" : "/src/assets/logo/dark-logo.jpeg"} 
          alt="AYCOM Logo" 
          class="h-8 w-auto"
        />
      </div>
    </div>
    
    <h1 class="text-2xl font-bold mb-6 text-center">Sign in to AYCOM</h1>
    
    <div id="google-signin-button" class="w-full mb-4"></div>
    
    <div class="flex items-center mb-4">
      <div class="flex-grow h-px bg-gray-600"></div>
      <span class="px-2 text-sm text-gray-400">or</span>
      <div class="flex-grow h-px bg-gray-600"></div>
    </div>
    
    {#if error}
      <div class="bg-red-500 bg-opacity-10 border border-red-500 text-red-500 px-4 py-3 rounded mb-4">
        {error}
      </div>
    {/if}
    
    <form on:submit|preventDefault={handleSubmit} class="mb-4">
      <!-- Email input -->
      <div class="mb-4">
        <label for="email" class="block text-sm font-medium mb-1">Email</label>
        <input 
          type="email" 
          id="email" 
          bind:value={email} 
          class="w-full p-2 border border-gray-600 rounded bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Email"
          required
        />
      </div>
      
      <!-- Password input -->
      <div class="mb-6">
        <div class="flex justify-between items-center mb-1">
          <label for="password" class="block text-sm font-medium">Password</label>
          <a href="/forgot-password" class="text-xs text-blue-500 hover:underline">Forgot password?</a>
        </div>
        <input 
          type="password" 
          id="password" 
          bind:value={password} 
          class="w-full p-2 border border-gray-600 rounded bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Password"
          required
        />
      </div>
      
      <!-- Remember me checkbox -->
      <div class="flex items-center justify-between mb-6">
        <label class="flex items-center">
          <input 
            type="checkbox" 
            bind:checked={rememberMe} 
            class="mr-2"
          />
          <span class="text-sm">Remember me</span>
        </label>
      </div>
      
      <!-- Submit button -->
      <button 
        type="submit"
        class="w-full py-3 bg-blue-500 text-white text-center rounded-full font-semibold hover:bg-blue-600 transition-colors flex justify-center items-center"
        disabled={isLoading}
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
    
    <!-- Register link -->
    <p class="text-sm mt-6 text-center">
      Don't have an account? <a href="/register" class="text-blue-500 hover:underline">Sign up</a>
    </p>
  </div>
</div>
