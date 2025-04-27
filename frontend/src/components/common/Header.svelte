<script>
  import { onMount } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  
  // Access the theme store
  const { theme } = useTheme();
  
  // Navbar properties
  export let sticky = true;
  export let transparent = false;
  
  // Handle scrolling effect
  let scrolled = false;
  let headerElement;

  onMount(() => {
    if (sticky) {
      const handleScroll = () => {
        scrolled = window.scrollY > 20;
      };
      
      window.addEventListener('scroll', handleScroll);
      
      return () => {
        window.removeEventListener('scroll', handleScroll);
      };
    }
  });
</script>

<header 
  bind:this={headerElement}
  class="header {sticky ? 'sticky top-0' : ''} {scrolled ? 'scrolled shadow-md' : ''} 
         {transparent && !scrolled ? 'bg-transparent' : $theme === 'dark' ? 'bg-gray-900' : 'bg-white'} 
         transition-all duration-300 z-40"
>
  <div class="container mx-auto px-4">
    <nav class="flex items-center justify-between h-16 md:h-20">
      <!-- Logo -->
      <div class="flex-shrink-0 flex items-center">
        <a href="/" class="flex items-center">
          <img 
            src={$theme === 'dark' ? "/src/assets/logo/dark-logo.jpeg" : "/src/assets/logo/light-logo.jpeg"} 
            alt="AYCOM Logo" 
            class="h-8 w-auto"
          />
          <span class="ml-2 text-xl font-bold {$theme === 'dark' ? 'text-white' : 'text-gray-900'}">AYCOM</span>
        </a>
      </div>
      
      <!-- Navigation for large screens -->
      <div class="hidden md:flex md:items-center md:space-x-8">
        <a href="/feed" class="nav-link {$theme === 'dark' ? 'text-gray-300 hover:text-white' : 'text-gray-700 hover:text-gray-900'}">Feed</a>
        <a href="/login" class="nav-link {$theme === 'dark' ? 'text-gray-300 hover:text-white' : 'text-gray-700 hover:text-gray-900'}">Login</a>
        <a href="/register" class="button-primary">Sign Up</a>
      </div>
      
      <!-- Mobile menu button -->
      <div class="md:hidden flex items-center">
        <button class="mobile-menu-button p-2 rounded-md {$theme === 'dark' ? 'text-gray-400 hover:text-white hover:bg-gray-800' : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'}">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
      </div>
    </nav>
  </div>
</header>

<style>
  .header {
    width: 100%;
  }
  
  .header.sticky {
    position: sticky;
    top: 0;
  }
  
  .header.scrolled {
    backdrop-filter: blur(8px);
  }
  
  .nav-link {
    font-weight: 500;
    transition: color 0.2s ease;
  }
  
  :global(.button-primary) {
    padding: 0.5rem 1.25rem;
    border-radius: 0.375rem;
    font-weight: 500;
    transition: all 0.2s ease;
    background-color: #3b82f6;
    color: white;
  }
  
  :global(.button-primary:hover) {
    background-color: #2563eb;
    transform: translateY(-1px);
  }
  
  :global(.button-primary:active) {
    transform: translateY(1px);
  }
  
  /* Dark mode adjustments */
  :global(.dark .button-primary) {
    background-color: #4f46e5;
  }
  
  :global(.dark .button-primary:hover) {
    background-color: #6366f1;
  }
</style>