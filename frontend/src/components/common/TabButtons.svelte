<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  
  // Define tab item interface
  interface TabItem {
    id: string;
    label: string;
    icon?: any;
    disabled?: boolean;
  }
  
  // Component props
  export let items: TabItem[] = [];
  export let activeId: string = '';
  export let variant: 'default' | 'pills' | 'underline' = 'default';
  
  // Event dispatcher
  const dispatch = createEventDispatcher();
  
  // Handle tab click
  function handleTabClick(item: TabItem) {
    if (item.disabled) return;
    activeId = item.id;
    dispatch('tabChange', item.id);
  }
</script>

<div class="tab-buttons tab-buttons-{variant}">
  {#each items as item}
    <button 
      class="tab-button {activeId === item.id ? 'active' : ''} {item.disabled ? 'disabled' : ''}"
      on:click={() => handleTabClick(item)}
      disabled={item.disabled}
      type="button"
    >
      {#if item.icon}
        <div class="tab-icon">
          <svelte:component this={item.icon} size="18" />
        </div>
      {/if}
      <span>{item.label}</span>
    </button>
  {/each}
</div>

<style>
  .tab-buttons {
    display: flex;
    width: 100%;
    border-bottom: 1px solid var(--border-color);
    margin-bottom: var(--space-4);
  }
  
  .tab-button {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-2);
    padding: var(--space-3) var(--space-4);
    background: transparent;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
    transition: all var(--transition-fast);
    position: relative;
    white-space: nowrap;
  }
  
  .tab-button:hover:not(.active):not(.disabled) {
    color: var(--text-primary);
    background-color: var(--bg-hover);
  }
  
  .tab-button.active {
    color: var(--color-primary);
    font-weight: var(--font-weight-bold);
  }
  
  .tab-button.disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  
  /* Default style with underline indicator */
  .tab-buttons-default .tab-button.active::after {
    content: "";
    position: absolute;
    bottom: -1px;
    left: 0;
    right: 0;
    height: 3px;
    background-color: var(--color-primary);
    border-radius: 3px 3px 0 0;
  }
  
  /* Pills style */
  .tab-buttons-pills {
    border-bottom: none;
    background-color: var(--bg-secondary);
    border-radius: var(--radius-full);
    padding: var(--space-1);
  }
  
  .tab-buttons-pills .tab-button {
    border-radius: var(--radius-full);
    padding: var(--space-2) var(--space-4);
  }
  
  .tab-buttons-pills .tab-button.active {
    background-color: var(--color-primary);
    color: white;
  }
  
  /* Underline style */
  .tab-buttons-underline .tab-button {
    border-bottom: 2px solid transparent;
    padding-left: var(--space-2);
    padding-right: var(--space-2);
    margin-right: var(--space-4);
  }
  
  .tab-buttons-underline .tab-button.active {
    border-bottom-color: var(--color-primary);
  }
  
  .tab-icon {
    display: flex;
    align-items: center;
    justify-content: center;
  }
</style> 