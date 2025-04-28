<script lang="ts">
  import type { INavigationItem } from '../../interfaces/ISocialMedia';
  
  export let username = "username";
  export let displayName = "User Name";
  export let avatar = "👤";
  
  // Navigation items
  const navItems: INavigationItem[] = [
    { label: 'Home', icon: '🏠', path: '/home' },
    { label: 'Explore', icon: '🔍', path: '/explore' },
    { label: 'Notifications', icon: '🔔', path: '/notifications' },
    { label: 'Messages', icon: '✉️', path: '/messages' },
    { label: 'Profile', icon: '👤', path: '/profile' }
  ];
</script>

<div class="fixed top-0 bottom-0 left-0 w-16 md:w-64 border-r border-gray-800 p-2 md:p-4">
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
    
    <!-- Post Button -->
    <button class="mt-4 w-12 h-12 md:w-full md:py-3 bg-blue-500 text-white rounded-full font-bold hover:bg-blue-600">
      <span class="md:hidden">+</span>
      <span class="hidden md:inline">Post</span>
    </button>
    
    <!-- User Profile -->
    <div class="mt-auto flex items-center p-3 hover:bg-gray-900 rounded-full cursor-pointer">
      <div class="w-8 h-8 bg-gray-300 rounded-full flex items-center justify-center mr-2">
        <span>{avatar}</span>
      </div>
      <div class="hidden md:block">
        <p class="font-semibold">{displayName}</p>
        <p class="text-gray-500 text-sm">@{username}</p>
      </div>
    </div>
  </div>
</div> 