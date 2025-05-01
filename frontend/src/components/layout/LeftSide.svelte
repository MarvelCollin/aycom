<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import ThemeToggle from '../common/ThemeToggle.svelte';
  import { useTheme } from '../../hooks/useTheme';

  // Props
  export let username = "";
  export let displayName = "";
  export let avatar = "👤";
  
  // Get theme from the store
  const { theme } = useTheme();
  $: isDarkMode = $theme === 'dark';
  
  // Event dispatcher
  const dispatch = createEventDispatcher();
  
  // Navigation items - updated according to requirements
  const navigationItems = [
    { label: "Home", path: "/home", icon: "home" },
    { label: "Explore", path: "/explore", icon: "hash" },
    { label: "Notifications", path: "/notifications", icon: "bell" },
    { label: "Messages", path: "/messages", icon: "mail" },
    { label: "Bookmarks", path: "/bookmarks", icon: "bookmark" },
    { label: "Communities", path: "/communities", icon: "users" },
    { label: "Premium", path: "/premium", icon: "star" },
    { label: "Profile", path: "/profile", icon: "user" },
    { label: "Settings", path: "/settings", icon: "settings" }
  ];
  
  // Toggle user menu
  let showUserMenu = false;
  function toggleUserMenu() {
    showUserMenu = !showUserMenu;
  }
  
  // Handle compose tweet modal action
  function handleToggleComposeModal() {
    dispatch('toggleComposeModal');
  }
  
  // Get current path for active state
  let currentPath = window.location.pathname;
</script>

<div class="flex flex-col h-full py-2 px-2 {isDarkMode ? 'text-white' : 'text-black'}">
  <!-- Logo -->
  <div class="px-3 mb-4">
    <a href="/" class="flex items-center justify-center md:justify-start p-3 rounded-full hover:bg-gray-200 dark:hover:bg-gray-800">
      <div class="text-3xl font-bold text-blue-500">AY</div>
    </a>
  </div>

  <!-- Navigation Menu -->
  <nav class="flex-1">
    <ul class="space-y-1">
      {#each navigationItems as item}
        <li>
          <a 
            href={item.path} 
            class="flex items-center px-4 py-3 rounded-full {currentPath === item.path ? 'font-bold' : 'font-normal'} {isDarkMode ? 'text-white hover:bg-gray-800' : 'text-black hover:bg-gray-200'}"
          >
            <!-- Icon -->
            <div class="flex items-center justify-center w-6 h-6">
              {#if item.icon === 'home'}
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
                </svg>
              {:else if item.icon === 'hash'}
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" />
                </svg>
              {:else if item.icon === 'bell'}
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
                </svg>
              {:else if item.icon === 'mail'}
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                </svg>
              {:else if item.icon === 'bookmark'}
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
                </svg>
              {:else if item.icon === 'users'}
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                </svg>
              {:else if item.icon === 'star'}
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z" />
                </svg>
              {:else if item.icon === 'user'}
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                </svg>
              {:else if item.icon === 'settings'}
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
              {/if}
            </div>
            <!-- Label -->
            <span class="hidden md:block text-xl ml-4">{item.label}</span>
          </a>
        </li>
      {/each}
    </ul>

    <!-- Post Button -->
    <div class="mt-4 px-3">
      <button 
        class="w-full py-3 bg-blue-500 text-white rounded-full font-bold hover:bg-blue-600"
        on:click={handleToggleComposeModal}
      >
        <span class="md:hidden">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 mx-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
        </span>
        <span class="hidden md:block text-lg">Post</span>
      </button>
    </div>

    <div class="mt-4 px-3">
      <div class="flex justify-center md:justify-start p-2">
        <ThemeToggle size="md" />
      </div>
    </div>
  </nav>

  <div class="mt-4 px-3 mb-4">
    <button 
      class="flex items-center w-full p-3 rounded-full {isDarkMode ? 'hover:bg-gray-800' : 'hover:bg-gray-200'}"
      on:click={toggleUserMenu}
    >
      <div class="w-10 h-10 rounded-full {isDarkMode ? 'bg-gray-700' : 'bg-gray-300'} flex items-center justify-center overflow-hidden">
        {#if typeof avatar === 'string' && avatar.startsWith('http')}
          <img src={avatar} alt={username} class="w-full h-full object-cover" />
        {:else}
          <span class="text-lg">{avatar}</span>
        {/if}
      </div>
      <div class="hidden md:block ml-3 flex-1 text-left">
        <p class="font-bold text-sm {isDarkMode ? 'text-white' : 'text-black'}">{displayName || 'User'}</p>
        <p class="text-sm {isDarkMode ? 'text-gray-300' : 'text-gray-700'}">@{username}</p>
      </div>
      <div class="hidden md:flex">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 {isDarkMode ? 'text-gray-300' : 'text-gray-700'}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h.01M12 12h.01M19 12h.01M6 12a1 1 0 11-2 0 1 1 0 012 0zm7 0a1 1 0 11-2 0 1 1 0 012 0zm7 0a1 1 0 11-2 0 1 1 0 012 0z" />
        </svg>
      </div>
    </button>
    
    <!-- User Menu Dropdown -->
    {#if showUserMenu}
      <div 
        class="absolute bottom-20 left-2 w-60 rounded-lg shadow border {isDarkMode ? 'bg-gray-900 border-gray-800' : 'bg-white border-gray-200'} z-50"
      >
        <div class="py-2">
          <button
            class="flex items-center w-full px-4 py-3 {isDarkMode ? 'text-white hover:bg-gray-800' : 'text-black hover:bg-gray-100'}"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
            </svg>
            Log out @{username}
          </button>
        </div>
      </div>
    {/if}
  </div>
</div>