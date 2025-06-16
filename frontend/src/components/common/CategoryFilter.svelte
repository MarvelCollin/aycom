<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  import { fade } from 'svelte/transition';

  export let categories: string[] = [];
  export let selected: string[] = [];
  export let label: string = 'Categories';

  let isOpen = false;
  let allSelected = false;

  const { theme } = useTheme();
  $: isDarkMode = $theme === 'dark';

  const dispatch = createEventDispatcher();

  function toggleDropdown() {
    isOpen = !isOpen;
  }

  function handleClickOutside(event: MouseEvent) {
    const target = event.target as HTMLElement;
    const dropdown = document.querySelector('.category-dropdown');
    if (dropdown && !dropdown.contains(target)) {
      isOpen = false;
    }
  }

  function toggleCategory(category: string) {
    if (selected.includes(category)) {
      selected = selected.filter(c => c !== category);
    } else {
      selected = [...selected, category];
    }

    allSelected = categories.length > 0 && selected.length === categories.length;

    dispatch('change', { categories: selected });
  }

  function toggleSelectAll() {
    if (allSelected) {

      selected = [];
      allSelected = false;
    } else {

      selected = [...categories];
      allSelected = true;
    }

    dispatch('change', { categories: selected });
  }

  function clearFilters() {
    selected = [];
    allSelected = false;
    dispatch('change', { categories: selected });
  }

  import { onMount, onDestroy } from 'svelte';

  onMount(() => {
    document.addEventListener('click', handleClickOutside);
  });

  onDestroy(() => {
    document.removeEventListener('click', handleClickOutside);
  });
</script>

<div class="category-filter">
  <button 
    class="filter-button {isDarkMode ? 'dark' : ''}" 
    on:click={toggleDropdown}
    aria-haspopup="true"
    aria-expanded={isOpen}
  >
    <span>{label}</span>
    <span class="filter-count">{selected.length > 0 ? `(${selected.length})` : ''}</span>
    <span class="arrow {isOpen ? 'open' : ''}">▼</span>
  </button>

  {#if isOpen}
    <div class="category-dropdown {isDarkMode ? 'dark' : ''}" transition:fade={{ duration: 100 }}>
      <div class="dropdown-header">
        <label class="checkbox-container">
          <input 
            type="checkbox" 
            checked={allSelected}
            on:change={toggleSelectAll}
          />
          <span class="checkmark"></span>
          <span class="label">All Categories</span>
        </label>

        {#if selected.length > 0}
          <button class="clear-button" on:click={clearFilters}>Clear</button>
        {/if}
      </div>

      <div class="dropdown-body">
        {#each categories as category}
          <label class="checkbox-container">
            <input 
              type="checkbox" 
              checked={selected.includes(category)}
              on:change={() => toggleCategory(category)}
            />
            <span class="checkmark"></span>
            <span class="label">{category}</span>
          </label>
        {/each}
      </div>
    </div>
  {/if}
</div>

<style>
  .category-filter {
    position: relative;
    display: inline-block;
  }

  .filter-button {
    display: flex;
    align-items: center;
    padding: 0.5rem 1rem;
    background-color: white;
    border: 1px solid #e0e0e0;
    border-radius: 0.25rem;
    font-size: 0.875rem;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .filter-button.dark {
    background-color: #2d3748;
    border-color: #4a5568;
    color: white;
  }

  .filter-button:hover {
    border-color: #cbd5e0;
    background-color: #f7fafc;
  }

  .filter-button.dark:hover {
    border-color: #4a5568;
    background-color: #2d3748;
  }

  .filter-count {
    margin-left: 0.25rem;
    font-weight: 500;
  }

  .arrow {
    margin-left: 0.5rem;
    font-size: 0.625rem;
    transition: transform 0.2s ease;
  }

  .arrow.open {
    transform: rotate(180deg);
  }

  .category-dropdown {
    position: absolute;
    top: 100%;
    left: 0;
    margin-top: 0.25rem;
    width: 230px;
    max-height: 300px;
    background-color: white;
    border: 1px solid #e0e0e0;
    border-radius: 0.25rem;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
    z-index: 50;
    overflow: hidden;
  }

  .category-dropdown.dark {
    background-color: #2d3748;
    border-color: #4a5568;
    color: white;
  }

  .dropdown-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.5rem 1rem;
    border-bottom: 1px solid #e0e0e0;
  }

  .category-dropdown.dark .dropdown-header {
    border-color: #4a5568;
  }

  .clear-button {
    font-size: 0.75rem;
    color: #3182ce;
    background: none;
    border: none;
    cursor: pointer;
    padding: 0;
  }

  .category-dropdown.dark .clear-button {
    color: #90cdf4;
  }

  .dropdown-body {
    max-height: 240px;
    overflow-y: auto;
    padding: 0.5rem 0;
  }

  .checkbox-container {
    display: block;
    position: relative;
    padding: 0.5rem 1rem 0.5rem 2rem;
    cursor: pointer;
    font-size: 0.875rem;
    user-select: none;
  }

  .checkbox-container:hover {
    background-color: #f7fafc;
  }

  .category-dropdown.dark .checkbox-container:hover {
    background-color: #4a5568;
  }

  .checkbox-container input {
    position: absolute;
    opacity: 0;
    height: 0;
    width: 0;
  }

  .checkmark {
    position: absolute;
    left: 1rem;
    top: 0.6rem;
    height: 16px;
    width: 16px;
    background-color: #eee;
    border-radius: 0.25rem;
  }

  .category-dropdown.dark .checkmark {
    background-color: #4a5568;
  }

  .checkbox-container:hover input ~ .checkmark {
    background-color: #ccc;
  }

  .category-dropdown.dark .checkbox-container:hover input ~ .checkmark {
    background-color: #718096;
  }

  .checkbox-container input:checked ~ .checkmark {
    background-color: #3182ce;
  }

  .checkmark:after {
    content: "";
    position: absolute;
    display: none;
  }

  .checkbox-container input:checked ~ .checkmark:after {
    display: block;
  }

  .checkbox-container .checkmark:after {
    left: 5px;
    top: 2px;
    width: 5px;
    height: 8px;
    border: solid white;
    border-width: 0 2px 2px 0;
    transform: rotate(45deg);
  }
</style>