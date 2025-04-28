<script lang="ts">
  import { useTheme } from '../../hooks/useTheme';
  import { useAuth } from '../../hooks/useAuth';
  import { useNavigation } from '../../hooks/useNavigation';
  import { createEventDispatcher } from 'svelte';
  import {
    LogOutIcon,
    MoonIcon,
    SunIcon,
    PlusIcon
  } from 'svelte-feather-icons';
  
  export let username = "username";
  export let displayName = "User Name";
  export let avatar = "👤";
  export let isDarkMode = false;
  
  const dispatch = createEventDispatcher();
  const { theme, toggleTheme } = useTheme();
  const { logout } = useAuth();
  const { navigationItems, activeItem, getIconComponent } = useNavigation();
  
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
  
  // Need to create icon components dynamically
  $: currentPath = $activeItem;
</script>

<div class="fixed top-0 bottom-0 left-0 w-16 md:w-[275px] border-r {isDarkMode ? 'border-gray-800' : 'border-gray-200'} p-2 md:p-4 overflow-y-auto {isDarkMode ? 'bg-gray-900' : 'bg-white'} transition-colors h-screen">
  <div class="flex flex-col items-center md:items-start h-full">
    <!-- Logo -->
    <div class="p-2 mb-4 rounded-full {isDarkMode ? 'hover:bg-gray-800' : 'hover:bg-gray-100'} transition-colors">
      <a href="/" class="flex items-center">
        <span class="{isDarkMode ? 'text-white' : 'text-gray-900'} text-2xl font-bold">AY</span>
      </a>
    </div>
    
    <!-- Navigation Items -->
    <nav class="w-full">
      {#each navigationItems as item}
        {@const IconComponent = getIconComponent(item.icon)}
        <a 
          href={item.path} 
          class="flex items-center p-3 {isDarkMode ? 'hover:bg-gray-800/70' : 'hover:bg-gray-100/70'} {currentPath === item.path ? 'font-bold' : ''} rounded-full mb-1 group transition-colors"
        >
          <div class="flex items-center justify-center w-8 h-8 mr-4">
            <svelte:component this={IconComponent} size="22" />
          </div>
          <span class="hidden md:inline text-xl {currentPath === item.path ? 'font-bold' : ''}">{item.label}</span>
        </a>
      {/each}
    </nav>
    
    <!-- Post Button -->
    <button 
      class="mt-4 w-12 h-12 md:w-full md:py-3 bg-blue-500 text-white rounded-full font-bold hover:bg-blue-600 transition-colors flex items-center justify-center md:text-lg"
      on:click={handlePostClick}
    >
      <span class="md:hidden"><PlusIcon size="24" /></span>
      <span class="hidden md:inline">Post</span>
    </button>
    
    <!-- Theme Toggle (moved up from the bottom) -->
    <button 
      class="flex items-center p-3 mt-4 {isDarkMode ? 'hover:bg-gray-800/70' : 'hover:bg-gray-100/70'} rounded-full mb-1 w-full transition-colors"
      on:click={toggleTheme}
    >
      <div class="flex items-center justify-center w-8 h-8 mr-4">
        {#if $theme === 'dark'}
          <MoonIcon size="22" />
        {:else}
          <SunIcon size="22" />
        {/if}
      </div>
      <span class="hidden md:inline text-xl">Theme</span>
    </button>
    
    <!-- User Profile -->
    <div class="mt-auto mb-4 w-full">
      <div 
        class="flex items-center p-3 {isDarkMode ? 'hover:bg-gray-800/70' : 'hover:bg-gray-100/70'} rounded-full cursor-pointer transition-colors"
        on:click={toggleUserMenu}
      >
        <div class="w-10 h-10 {isDarkMode ? 'bg-gray-700' : 'bg-gray-300'} rounded-full flex items-center justify-center mr-3 transition-colors overflow-hidden">
          <span>{avatar}</span>
        </div>
        <div class="hidden md:block flex-1">
          <p class="font-bold truncate max-w-[180px]">{displayName}</p>
          <p class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} text-sm truncate max-w-[180px]">@{username}</p>
        </div>
        <div class="hidden md:block">⋯</div>
      </div>
      
      {#if showUserMenu}
        <div class="absolute bottom-20 left-4 md:left-6 w-60 {isDarkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'} border rounded-2xl shadow-lg transition-colors z-10">
          <button 
            class="w-full text-left p-4 {isDarkMode ? 'hover:bg-gray-700' : 'hover:bg-gray-100'} transition-colors flex items-center rounded-2xl"
            on:click={handleLogout}
          >
            <LogOutIcon size="18" color="currentColor" />
            <span class="ml-3">Log out @{username}</span>
          </button>
        </div>
      {/if}
    </div>
  </div>
</div> 