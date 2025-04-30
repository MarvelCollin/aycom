<script lang="ts">
  import LeftSide from './LeftSide.svelte';
  import RightSide from './RightSide.svelte';
  import Toast from '../common/Toast.svelte';
  import { useTheme } from '../../hooks/useTheme';

  // Props for passing data to sidebars
  export let username = "";
  export let displayName = "";
  export let avatar = "👤";
  export let trends = [];
  export let suggestedFollows = [];
  
  // Optional slot props
  export let showLeftSidebar = true;
  export let showRightSidebar = true;

  // Get theme state
  const { theme } = useTheme();
  $: isDarkMode = $theme === 'dark';

  // Function to handle tweet compose action
  function handleToggleComposeModal() {
    // Dispatch event to parent
    dispatch('toggleComposeModal');
  }

  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();
</script>

<div class="flex relative w-full h-screen {isDarkMode ? 'bg-black text-white' : 'bg-white text-black'}">
  <!-- Left Sidebar - Fixed position with independent scroll -->
  {#if showLeftSidebar}
    <div class="fixed left-0 top-0 z-40 h-screen border-r {isDarkMode ? 'border-gray-800' : 'border-gray-200'} overflow-y-auto {isDarkMode ? 'bg-black' : 'bg-white'}" style="width: 275px;">
      <LeftSide 
        {username}
        {displayName}
        {avatar}
        {isDarkMode}
        on:toggleComposeModal={handleToggleComposeModal}
      />
    </div>
    <div class="flex-shrink-0" style="width: 275px;"></div>
  {/if}
  
  <!-- Main Content Column - Scrollable -->
  <main class="flex-grow h-screen overflow-y-auto relative {isDarkMode ? 'bg-black' : 'bg-white'}">
    <slot></slot>
  </main>
  
  <!-- Right Sidebar - Fixed position with independent scroll -->
  {#if showRightSidebar}
    <div class="hidden md:block fixed right-0 top-0 z-40 h-screen {isDarkMode ? 'bg-black' : 'bg-white'} border-l {isDarkMode ? 'border-gray-800' : 'border-gray-200'} overflow-y-auto" style="width: 350px;">
      <div class="p-4">
        <RightSide 
          {isDarkMode}
          {trends}
          suggestedFollows={suggestedFollows}
        />
      </div>
    </div>
    <div class="hidden md:block flex-shrink-0" style="width: 350px;"></div>
  {/if}
  
  <!-- Toast Component -->
  <Toast />
</div>