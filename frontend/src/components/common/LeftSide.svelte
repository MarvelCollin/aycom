<script lang="ts">
  import SidebarNav from '../navigation/SidebarNav.svelte';
  import { createEventDispatcher } from 'svelte';
  
  export let username = "";
  export let displayName = "";
  export let avatar = "";
  export let isDarkMode = false;
  
  const dispatch = createEventDispatcher();
  
  function handleToggleComposeModal() {
    dispatch('toggleComposeModal');
  }
</script>

<div class="w-[68px] md:w-[275px] flex-shrink-0 h-screen">
  <SidebarNav 
    {username} 
    {displayName} 
    {avatar}
    {isDarkMode}
    on:toggleComposeModal={handleToggleComposeModal}
  />
</div>
