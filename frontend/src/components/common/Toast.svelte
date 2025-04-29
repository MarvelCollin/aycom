<script lang="ts">
  import { toastStore, type ToastState, type ToastType } from '../../stores/toastStore';
  import { fly } from 'svelte/transition';
  import {
    InfoIcon,
    CheckCircleIcon,
    AlertTriangleIcon,
    XCircleIcon,
    XIcon
  } from 'svelte-feather-icons';

  let message = '';
  let type: ToastType = 'info';
  let visible = false;
  let toastId = 0;

  // Subscribe to the toast store
  toastStore.subscribe((state: ToastState) => {
    // Only update if ID changes to allow re-triggering with same message
    if (state.id !== toastId) {
      message = state.message;
      type = state.type;
      visible = state.visible;
      toastId = state.id;
    }
  });

  // Map types to colors and icons
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

  $: currentStyle = typeStyles[type] || typeStyles.info;
</script>

{#if visible && message}
  <div 
    class="fixed bottom-5 right-5 z-50 max-w-sm rounded-lg shadow-lg overflow-hidden {currentStyle.bg} text-white"
    transition:fly={{ y: 20, duration: 300 }}
  >
    <div class="flex items-center p-4">
      <div class="flex-shrink-0">
        <svelte:component this={currentStyle.icon} size="24" />
      </div>
      <div class="ml-3 flex-1">
        <p class="text-sm font-medium">{message}</p>
      </div>
      <div class="ml-4 flex-shrink-0">
        <button 
          on:click={toastStore.hideToast} 
          class="inline-flex text-white rounded-md p-1 hover:bg-white/20 focus:outline-none focus:ring-2 focus:ring-white/50 transition-colors"
        >
          <span class="sr-only">Close</span>
          <XIcon size="18" />
        </button>
      </div>
    </div>
  </div>
{/if}
