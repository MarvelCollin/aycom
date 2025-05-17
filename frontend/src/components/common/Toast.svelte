<script lang="ts">
  import { toastStore } from '../../stores/toastStore';
  import type { Toast, ToastType, ToastPosition } from '../../interfaces/IToast';
  import { fly } from 'svelte/transition';
  import InfoIcon from 'svelte-feather-icons/src/icons/InfoIcon.svelte';
  import CheckCircleIcon from 'svelte-feather-icons/src/icons/CheckCircleIcon.svelte';
  import AlertTriangleIcon from 'svelte-feather-icons/src/icons/AlertTriangleIcon.svelte';
  import XCircleIcon from 'svelte-feather-icons/src/icons/XCircleIcon.svelte';
  import XIcon from 'svelte-feather-icons/src/icons/XIcon.svelte';

  let toasts: Toast[] = [];
  toastStore.subscribe((list) => {
    toasts = list;
  });

  const typeStyles = {
    info: {
      bg: 'bg-blue-500',
      icon: InfoIcon
    },
    success: {
      bg: 'bg-green-500',
      icon: CheckCircleIcon
    },
    warning: {
      bg: 'bg-yellow-500',
      icon: AlertTriangleIcon
    },
    error: {
      bg: 'bg-red-500',
      icon: XCircleIcon
    }
  };

  const positionClasses: Record<ToastPosition, string> = {
    'top-left': 'toast-top-left',
    'top-center': 'toast-top-center',
    'top-right': 'toast-top-right',
    'bottom-left': 'toast-bottom-left',
    'bottom-center': 'toast-bottom-center',
    'bottom-right': 'toast-bottom-right',
  };
</script>

{#each Object.keys(positionClasses) as pos}
  {#if toasts.filter(t => t.position === pos).length}
    <div class="toast-container {positionClasses[pos]}">
      {#each toasts.filter(t => t.position === pos) as toast (toast.id)}
        <div 
          class="toast-item {typeStyles[toast.type].bg}"
          transition:fly={{ y: pos.startsWith('top') ? -20 : 20, duration: 300 }}
        >
          <div class="toast-content">
            <div class="toast-icon">
              <svelte:component this={typeStyles[toast.type].icon} size="24" />
            </div>
            <div class="toast-message">
              <p>{toast.message}</p>
            </div>
            <div class="toast-close">
              <button 
                on:click={() => toastStore.removeToast(toast.id)} 
                class="toast-close-button"
              >
                <span class="sr-only">Close</span>
                <XIcon size="18" />
              </button>
            </div>
          </div>
        </div>
      {/each}
    </div>
  {/if}
{/each}

<style>
  .toast-container {
    position: fixed;
    z-index: 9999;
    pointer-events: none;
    width: 100%;
    max-width: 24rem;
  }
  
  /* Position classes */
  .toast-top-left {
    top: 1.25rem;
    left: 1.25rem;
  }
  
  .toast-top-center {
    top: 1.25rem;
    left: 50%;
    transform: translateX(-50%);
  }
  
  .toast-top-right {
    top: 1.25rem;
    right: 1.25rem;
  }
  
  .toast-bottom-left {
    bottom: 1.25rem;
    left: 1.25rem;
  }
  
  .toast-bottom-center {
    bottom: 1.25rem;
    left: 50%;
    transform: translateX(-50%);
  }
  
  .toast-bottom-right {
    bottom: 1.25rem;
    right: 1.25rem;
  }
  
  .toast-item {
    margin-bottom: 0.75rem;
    border-radius: 0.5rem;
    overflow: hidden;
    pointer-events: auto;
    box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
    color: white;
  }
  
  .toast-content {
    display: flex;
    align-items: center;
    padding: 1rem;
  }
  
  .toast-icon {
    flex-shrink: 0;
  }
  
  .toast-message {
    margin-left: 0.75rem;
    flex: 1;
    font-size: 0.875rem;
    font-weight: 500;
  }
  
  .toast-close {
    margin-left: 1rem;
    flex-shrink: 0;
  }
  
  .toast-close-button {
    display: inline-flex;
    color: white;
    border-radius: 0.375rem;
    padding: 0.25rem;
    transition: background-color 0.2s;
  }
  
  .toast-close-button:hover {
    background-color: rgba(255, 255, 255, 0.2);
  }
  
  .toast-close-button:focus {
    outline: none;
    box-shadow: 0 0 0 2px rgba(255, 255, 255, 0.5);
  }
</style>
