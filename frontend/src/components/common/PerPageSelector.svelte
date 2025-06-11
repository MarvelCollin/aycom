<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  
  const dispatch = createEventDispatcher();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  
  // Props
  export let perPage: number = 20;
  export let options: number[] = [10, 20, 50, 100];
  export let label: string = "Show:";
  
  function handleChange(event: Event) {
    const target = event.target as HTMLSelectElement;
    const value = parseInt(target.value, 10);
    dispatch('perPageChange', value);
  }
</script>

<div class="per-page-selector {isDarkMode ? 'per-page-selector-dark' : ''}">
  <label for="perPage">{label}</label>
  <select
    id="perPage"
    class="selector-dropdown"
    value={perPage}
    on:change={handleChange}
  >
    {#each options as option}
      <option value={option}>{option} per page</option>
    {/each}
  </select>
</div>

<style>
  .per-page-selector {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  
  .per-page-selector-dark {
    color: var(--dark-text-secondary);
  }
  
  label {
    font-size: 14px;
    color: var(--text-secondary);
  }
  
  .selector-dropdown {
    appearance: none;
    padding: 6px 30px 6px 12px;
    border-radius: 9999px;
    border: 1px solid var(--border-color);
    background-color: transparent;
    font-size: 14px;
    color: var(--text-primary);
    cursor: pointer;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%23536471' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
    background-repeat: no-repeat;
    background-position: right 10px center;
    background-size: 12px;
  }
  
  .selector-dropdown:focus {
    border-color: var(--color-primary);
    outline: none;
  }
  
  .per-page-selector-dark .selector-dropdown {
    color: var(--dark-text-primary);
    border-color: var(--dark-border-color);
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%238b98a5' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
  }
  
  @media (max-width: 640px) {
    .per-page-selector {
      width: 100%;
      justify-content: center;
    }
    
    .selector-dropdown {
      padding: 5px 28px 5px 10px;
    }
  }
</style> 