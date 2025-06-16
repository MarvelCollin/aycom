<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  import { supabase } from '../../utils/supabase';
  import { formatStorageUrl } from '../../utils/common';
  import { toastStore } from '../../stores/toastStore';

  export let variant: 'primary' | 'secondary' | 'outlined' | 'text' | 'danger' | 'success' = 'primary';
  export let size: 'small' | 'medium' | 'large' = 'medium';
  export let type: 'button' | 'submit' | 'reset' = 'button';
  export let disabled: boolean = false;
  export let fullWidth: boolean = false;
  export let loading: boolean = false;
  export let icon: any = null;
  export let iconPosition: 'left' | 'right' = 'left';
  export let tooltip: string = '';

  const { theme } = useTheme();
  const dispatch = createEventDispatcher();

  async function testSupabaseConnection() {
    try {

      const { data: buckets, error } = await supabase.storage.listBuckets();

      if (error) {
        console.error("Supabase connection error:", error);
        toastStore.showToast(`Supabase error: ${error.message}`, 'error');
        return;
      }

      console.log("Supabase buckets:", buckets);

      const testUrls = [
        "profile-pictures/abc123.jpg",
        "banners/xyz789.png",
        "tpaweb/profile-pictures/test.jpg",
        "https://sdhtnvlmuywinhcglfsu.supabase.co/storage/v1/object/public/profile-pictures/test.jpg"
      ];

      testUrls.forEach(url => {
        const formatted = formatStorageUrl(url);
        console.log(`Original: ${url} -> Formatted: ${formatted}`);
      });

      toastStore.showToast(`Supabase connected. Found ${buckets.length} buckets.`, 'success');
    } catch (e: unknown) {
      console.error("Error testing Supabase:", e);
      const errorMessage = e instanceof Error ? e.message : 'Unknown error';
      toastStore.showToast(`Error: ${errorMessage}`, 'error');
    }
  }

  function handleClick(event) {
    if (disabled) {
      event.preventDefault();
      return;
    }

    if ($$props.class && $$props.class.includes('test-supabase')) {
      testSupabaseConnection();
    }

    dispatch('click', event);
  }

  $: buttonClass = `btn btn-${variant} btn-${size} ${fullWidth ? 'btn-full' : ''} ${loading ? 'btn-loading' : ''}`;
</script>

<button 
  {type}
  class={buttonClass}
  {disabled}
  on:click={handleClick}
  title={tooltip}
  {...$$restProps}
>
  {#if loading}
    <span class="loading-spinner"></span>
  {:else if icon}
    <span class="btn-icon">
      <svelte:component this={icon} size={size === 'small' ? '14' : size === 'large' ? '20' : '18'} />
    </span>
  {/if}

  <span class="btn-content">
    <slot></slot>
  </span>
</button>

<style>
  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-2);
    border-radius: var(--radius-md);
    font-weight: var(--font-weight-medium);
    cursor: pointer;
    transition: all var(--transition-fast);
    border: none;
    white-space: nowrap;
    text-decoration: none;
  }

  .btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .btn-small {
    padding: var(--space-1) var(--space-3);
    font-size: var(--font-size-xs);
    height: 32px;
  }

  .btn-medium {
    padding: var(--space-2) var(--space-4);
    font-size: var(--font-size-sm);
    height: 40px;
  }

  .btn-large {
    padding: var(--space-3) var(--space-6);
    font-size: var(--font-size-md);
    height: 48px;
  }

  .btn-primary {
    background-color: var(--color-primary);
    color: white;
  }

  .btn-primary:hover:not(:disabled) {
    background-color: var(--color-primary-hover);
  }

  .btn-secondary {
    background-color: var(--bg-secondary);
    color: var(--text-primary);
    border: 1px solid var(--border-color);
  }

  .btn-secondary:hover:not(:disabled) {
    background-color: var(--bg-hover);
  }

  .btn-outlined {
    background-color: transparent;
    border: 1px solid var(--color-primary);
    color: var(--color-primary);
  }

  .btn-outlined:hover:not(:disabled) {
    background-color: var(--color-primary-light);
  }

  .btn-text {
    background-color: transparent;
    color: var(--color-primary);
    padding-left: var(--space-2);
    padding-right: var(--space-2);
  }

  .btn-text:hover:not(:disabled) {
    background-color: var(--color-primary-light);
  }

  .btn-danger {
    background-color: var(--color-danger);
    color: white;
  }

  .btn-danger:hover:not(:disabled) {
    background-color: var(--color-danger-hover);
  }

  .btn-full {
    width: 100%;
  }

  .btn-loading {
    position: relative;
    color: transparent;
  }

  .loading-spinner {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 20px;
    height: 20px;
    border-radius: 50%;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-top-color: white;
    animation: spin 0.8s linear infinite;
  }

  .btn-secondary .loading-spinner,
  .btn-outlined .loading-spinner,
  .btn-text .loading-spinner {
    border: 2px solid rgba(0, 0, 0, 0.1);
    border-top-color: var(--color-primary);
  }

  @keyframes spin {
    to {
      transform: translate(-50%, -50%) rotate(360deg);
    }
  }

  .btn-icon {
    display: flex;
    align-items: center;
    justify-content: center;
  }
</style>