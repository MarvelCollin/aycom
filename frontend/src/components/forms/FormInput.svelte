<script lang="ts">
  export let id = '';
  export let type = 'text';
  export let value = '';
  export let label = '';
  export let placeholder = '';
  export let error = '';
  export let required = false;
  export let maxlength = undefined;
  export let dataCy = '';
  
  // Create an event dispatcher
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();
  
  // Function to handle input changes
  function handleInput(event: Event) {
    value = (event.target as HTMLInputElement | HTMLTextAreaElement).value;
    dispatch('input', { value });
  }
  
  // Function to handle blur events for validation
  function handleBlur() {
    dispatch('blur');
  }
  
  // Show character count for text inputs with maxlength
  $: showCharCount = !!maxlength && (type === 'text' || type === 'textarea');
</script>

<div class="mb-4">
  {#if label}
    <div class="flex justify-between">
      <label for={id} class="block text-sm font-medium mb-1">{label}</label>
      {#if showCharCount}
        <span class="text-xs text-gray-400" data-cy={`${dataCy}-char-count`}>{value.length} / {maxlength}</span>
      {/if}
    </div>
  {/if}
  
  {#if type === 'textarea'}
    <textarea 
      {id}
      bind:value
      class="w-full p-2 border border-gray-600 rounded bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500"
      {placeholder}
      on:blur={handleBlur}
      {maxlength}
      {required}
      data-cy={dataCy}
    ></textarea>
  {:else}
    <input 
      {id}
      {type}
      bind:value
      class="w-full p-2 border border-gray-600 rounded bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500"
      {placeholder}
      on:blur={handleBlur}
      {maxlength}
      {required}
      data-cy={dataCy}
    />
  {/if}
  
  {#if error}
    <p class="text-red-500 text-xs mt-1" data-cy={`${dataCy}-error`}>{error}</p>
  {/if}
</div> 