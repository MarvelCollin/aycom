<script lang="ts">
  import { onMount } from 'svelte';
  import { useTheme } from '../hooks/useTheme';
  import LandingLayout from '../components/layout/LandingLayout.svelte';
  import darkLogo from '../assets/logo/dark-logo.jpeg';
  import lightLogo from '../assets/logo/light-logo.jpeg';
  
  // Get the theme store and toggleTheme function from our hook
  const { theme } = useTheme();
  
  // Reactive declaration to update isDarkMode when theme changes
  $: isDarkMode = $theme === 'dark';
  
  onMount(() => {
    // Preload logo images
    const logoImgDark = new Image();
    logoImgDark.src = darkLogo;
    
    const logoImgLight = new Image();
    logoImgLight.src = lightLogo;
  });
</script>

<LandingLayout>
  <div class="flex flex-col min-h-screen md:flex-row {isDarkMode ? 'dark-content' : 'light-content'}">
    <!-- Left side with logo -->
    <div class="w-full md:w-1/2 flex items-center justify-center md:justify-end p-4 md:pr-12 lg:pr-16">
      <div class="w-full max-w-md md:max-w-lg">
        <img 
          src={isDarkMode ? lightLogo : darkLogo} 
          alt="AYCOM Logo" 
          class="w-full h-auto rounded-xl {isDarkMode ? 'shadow-dark' : 'shadow-light'}"
        />
      </div>
    </div>
    
    <!-- Right side with content -->
    <div class="w-full md:w-1/2 flex flex-col justify-center p-6 md:pl-8 lg:pl-12 md:pr-16 lg:pr-24">
      <h1 class="text-5xl md:text-6xl font-bold mb-12 md:mb-14 leading-tight">Happening now</h1>
      
      <h2 class="text-3xl font-bold mb-8">Join AYCOM today</h2>
      
      <div class="w-full max-w-sm">
        <a href="/register" class="block w-full py-3 bg-blue-500 text-white text-center rounded-full font-semibold hover:bg-blue-600 transition-colors mb-4">
          Create account
        </a>
        
        <p class="text-xs text-gray-500 mb-8">
          By signing up, you agree to the <a href="#" class="text-blue-500 hover:underline">Terms of Service</a> and 
          <a href="#" class="text-blue-500 hover:underline">Privacy Policy</a>, including <a href="#" class="text-blue-500 hover:underline">Cookie Use</a>.
        </p>
        
        <div class="flex items-center my-6">
          <div class="flex-grow h-px bg-gray-300 dark:bg-gray-700"></div>
          <span class="px-4 text-sm text-gray-500">or</span>
          <div class="flex-grow h-px bg-gray-300 dark:bg-gray-700"></div>
        </div>
        
        <p class="text-base font-medium mb-3">Already have an account?</p>
        <a 
          href="/login" 
          class="block w-full py-3 text-center rounded-full font-semibold transition-colors border border-gray-300 dark:border-gray-700 text-blue-500 hover:bg-gray-100 dark:hover:bg-gray-800"
        >
          Sign in
        </a>
      </div>
    </div>
  </div>
</LandingLayout>

<style>
  .shadow-light {
    box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
  }
  
  .shadow-dark {
    box-shadow: 0 10px 15px -3px rgba(255, 255, 255, 0.1), 0 4px 6px -2px rgba(255, 255, 255, 0.05);
  }
  
  :global(.dark-mode) .dark-content {
    background-color: var(--bg-color);
    color: var(--text-color);
  }
  
  :global(.light-mode) .light-content {
    background-color: var(--bg-color);
    color: var(--text-color);
  }
</style>