<script lang="ts">
  import LeftSide from './LeftSide.svelte';
  import RightSide from './RightSide.svelte';
  import Toast from '../common/Toast.svelte';
  import { useTheme } from '../../hooks/useTheme';
  import type { ITrend, ISuggestedFollow } from '../../interfaces/ISocialMedia';

  export let username = "";
  export let displayName = "";
  export let avatar = "👤";
  export let trends: ITrend[] = [];
  export let suggestedFollows: ISuggestedFollow[] = [];
  
  export let showLeftSidebar = true;
  export let showRightSidebar = true;

  const { theme } = useTheme();
  $: isDarkMode = $theme === 'dark';

  function handleToggleComposeModal() {
    dispatch('toggleComposeModal');
  }

  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();
</script>

<div class="flex relative w-full h-screen {isDarkMode ? 'bg-black text-white' : 'bg-white text-black'}">
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
  
  <main class="flex-grow h-screen overflow-y-auto relative {isDarkMode ? 'bg-black' : 'bg-white'}">
    <slot></slot>
  </main>
  
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
    <div class="hidpden md:block flex-shrink-0" style="width: 350px;"></div>
  {/if}
  
  <Toast />
</div>