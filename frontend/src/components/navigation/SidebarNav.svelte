<script lang="ts">
  import { useTheme } from '../../hooks/useTheme';
  import { useAuth } from '../../hooks/useAuth';
  import type { INavigationItem } from '../../interfaces/ISocialMedia';
  import { createEventDispatcher } from 'svelte';
  
  export let username = "username";
  export let displayName = "User Name";
  export let avatar = "👤";
  
  const dispatch = createEventDispatcher();
  const { theme, toggleTheme } = useTheme();
  const { logout } = useAuth();
  
  let showUserMenu = false;
  
  // Toggle user menu
  function toggleUserMenu() {
    showUserMenu = !showUserMenu;
  }
  
  // Handle logout
  function handleLogout() {
    logout();
    window.location.href = '/login';
  }
  
  // Handle post button click
  function handlePostClick() {
    dispatch('toggleComposeModal');
  }
  
  // Navigation items
  const navItems: INavigationItem[] = [
    { label: 'Home', icon: '🏠', path: '/feed' },
    { label: 'Explore', icon: '🔍', path: '/explore' },
    { label: 'Notifications', icon: '🔔', path: '/notifications' },
    { label: 'Messages', icon: '✉️', path: '/messages' },
    { label: 'Grok', icon: '🤖', path: '/grok' },
    { label: 'Bookmarks', icon: '🔖', path: '/bookmarks' },
    { label: 'Communities', icon: '👥', path: '/communities' },
    { label: 'Premium', icon: '⭐', path: '/premium' },
    { label: 'Verified Orgs', icon: '✓', path: '/verified-orgs' },
    { label: 'Profile', icon: '👤', path: '/profile' },
    { label: 'More', icon: '•••', path: '/more' },
  ];
</script>

<div class="fixed top-0 bottom-0 left-0 w-16 md:w-64 border-r border-gray-800 p-2 md:p-4 overflow-y-auto">
  <div class="flex flex-col items-center md:items-start h-full">
    <!-- Logo -->
    <div class="p-2 mb-4 rounded-full hover:bg-gray-900">
      <span class="text-white text-2xl font-bold">AY</span>
    </div>
    
    <!-- Navigation Items -->
    <nav class="w-full">
      {#each navItems as item}
        <a 
          href={item.path} 
          class="flex items-center p-3 hover:bg-gray-900 rounded-full mb-1 font-medium"
        >
          <span class="mr-4">{item.icon}</span>
          <span class="hidden md:inline">{item.label}</span>
        </a>
      {/each}
    </nav>
    
    <!-- Theme Toggle -->
    <button 
      class="flex items-center p-3 hover:bg-gray-900 rounded-full mb-1 font-medium w-full"
      on:click={toggleTheme}
    >
      <span class="mr-4">{$theme === 'dark' ? '🌙' : '☀️'}</span>
      <span class="hidden md:inline">Theme</span>
    </button>
    
    <!-- Post Button -->
    <button 
      class="mt-4 w-12 h-12 md:w-full md:py-3 bg-blue-500 text-white rounded-full font-bold hover:bg-blue-600"
      on:click={handlePostClick}
    >
      <span class="md:hidden">+</span>
      <span class="hidden md:inline">Post</span>
    </button>
    
    <!-- User Profile -->
    <div class="relative mt-auto">
      <div 
        class="flex items-center p-3 hover:bg-gray-900 rounded-full cursor-pointer"
        on:click={toggleUserMenu}
      >
        <div class="w-8 h-8 bg-gray-300 rounded-full flex items-center justify-center mr-2">
          <span>{avatar}</span>
        </div>
        <div class="hidden md:block">
          <p class="font-semibold">{displayName}</p>
          <p class="text-gray-500 text-sm">@{username}</p>
        </div>
      </div>
      
      {#if showUserMenu}
        <div class="absolute bottom-full left-0 mb-2 w-48 bg-black border border-gray-800 rounded-lg shadow-lg">
          <button 
            class="w-full text-left p-3 hover:bg-gray-900"
            on:click={handleLogout}
          >
            Log out @{username}
          </button>
        </div>
      {/if}
    </div>
  </div>
</div> 