<script lang="ts">
  import { toastStore } from '../../stores/toastStore';
  import type { Toast, ToastType, ToastPosition } from '@interfaces/IToast';
  import { fly } from 'svelte/transition';
  import { InfoIcon, CheckCircleIcon, AlertTriangleIcon, XCircleIcon, XIcon } from 'svelte-feather-icons';

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
    'top-left': 'top-5 left-5',
    'top-center': 'top-5 left-1/2 transform -translate-x-1/2',
    'top-right': 'top-5 right-5',
    'bottom-left': 'bottom-5 left-5',
    'bottom-center': 'bottom-5 left-1/2 transform -translate-x-1/2',
    'bottom-right': 'bottom-5 right-5',
  };
</script>

{#each Object.keys(positionClasses) as pos}
  {#if toasts.filter(t => t.position === pos).length}
    <div class="fixed z-50 pointer-events-none w-full max-w-sm" style="{positionClasses[pos]}" >
      {#each toasts.filter(t => t.position === pos) as toast (toast.id)}
        <div 
          class="mb-3 rounded-lg shadow-lg overflow-hidden {typeStyles[toast.type].bg} text-white pointer-events-auto"
          transition:fly={{ y: 20, duration: 300 }}
        >
          <div class="flex items-center p-4">
            <div class="flex-shrink-0">
              <svelte:component this={typeStyles[toast.type].icon} size="24" />
            </div>
            <div class="ml-3 flex-1">
              <p class="text-sm font-medium">{toast.message}</p>
            </div>
            <div class="ml-4 flex-shrink-0">
              <button 
                on:click={() => toastStore.removeToast(toast.id)} 
                class="inline-flex text-white rounded-md p-1 hover:bg-white/20 focus:outline-none focus:ring-2 focus:ring-white/50 transition-colors"
              >
                <span class="sr-only">Close</span>
                <X size="18" />
              </button>
            </div>
          </div>
        </div>
      {/each}
    </div>
  {/if}
{/each}
