<script lang="ts">
  import { onMount } from 'svelte';
  import { useTheme } from '../hooks/useTheme';
  
  // Get the theme store and toggleTheme function from our hook
  const { theme, toggleTheme } = useTheme();
  
  let isDarkMode;
  
  // Reactive declaration to update isDarkMode when theme changes
  $: isDarkMode = $theme === 'dark';
  
  onMount(() => {
    // Preload logo images
    const logoImgDark = new Image();
    logoImgDark.src = '/src/assets/logo/dark-logo.jpeg';
    
    const logoImgLight = new Image();
    logoImgLight.src = '/src/assets/logo/light-logo.jpeg';
  });
</script>

<div class="{isDarkMode ? 'bg-black text-white' : 'bg-white text-black'} min-h-screen w-full pos-relative overflow-x-hidden">
  <button 
    class="pos-absolute top-4 right-4 p-2 rounded-full z-10 {isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}"
    on:click={toggleTheme}
    aria-label="Toggle theme"
  >
    {#if isDarkMode}
      <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-yellow-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
      </svg>
    {:else}
      <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-gray-800" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
      </svg>
    {/if}
  </button>
  
  <div class="w-full max-w-screen-xl mx-auto p-8 md:p-0 flex flex-col md:flex-row md:h-screen">
    <div class="flex justify-center items-center w-full md:w-55 mb-8 md:mb-0 md:h-full">
      <div class="w-full max-w-xs md:max-w-md flex justify-center items-center">
        <img 
          src={isDarkMode ? "/src/assets/logo/light-logo.jpeg" : "/src/assets/logo/dark-logo.jpeg"} 
          alt="AYCOM Logo" 
          class="w-full h-auto rounded-xl shadow-lg"
        />
      </div>
    </div>
    
    <div class="flex flex-col px-4 w-full md:w-45 md:justify-center">
      <h1 class="text-6xl md:text-7xl font-bold mb-6 leading-tight">Happening now</h1>
      <h2 class="text-3xl md:text-4xl font-bold mb-4">Join today.</h2>
      <p class="text-xl mb-8">Connect, share, engage.</p>
      
      <div class="w-full max-w-sm">
        <div class="mb-6">
          <p class="text-sm mb-2">Don't have an account?</p>
          <a href="/register" class="block w-full py-3 bg-blue-500 text-white text-center rounded-full font-semibold hover:bg-blue-600 transition-colors mb-2">
            Create account
          </a>
          <p class="text-xs mt-2 mb-6 text-gray-500">
            By signing up, you agree to the <a href="#" class="text-blue-500 hover:underline">Terms of Service</a> and 
            <a href="#" class="text-blue-500 hover:underline">Privacy Policy</a>, including <a href="#" class="text-blue-500 hover:underline">Cookie Use</a>.
          </p>
        </div>
        
        <div class="border-t border-gray-800 my-6"></div>
        
        <div class="mb-6">
          <p class="text-sm mb-2">Already have an account?</p>
          <a 
            href="/login" 
            class="block w-full py-3 text-center rounded-full font-semibold hover:bg-gray-900 transition-colors border border-gray-800 text-blue-500"
          >
            Sign in
          </a>
        </div>
      </div>
    </div>
  </div>
</div>