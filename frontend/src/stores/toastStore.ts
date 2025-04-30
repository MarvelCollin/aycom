import { writable } from 'svelte/store';

type ToastType = 'success' | 'error' | 'info' | 'warning';

interface Toast {
  id: string;
  message: string;
  type: ToastType;
  timeout: number;
}

function createToastStore() {
  const { subscribe, update } = writable<Toast[]>([]);

  function showToast(message: string, type: ToastType = 'info', timeout: number = 3000) {
    const id = Math.random().toString(36).substring(2, 9);
    
    update(toasts => [
      ...toasts,
      { id, message, type, timeout }
    ]);

    setTimeout(() => {
      removeToast(id);
    }, timeout);

    return id;
  }

  function removeToast(id: string) {
    update(toasts => toasts.filter(t => t.id !== id));
  }

  return {
    subscribe,
    showToast,
    removeToast
  };
}

export const toastStore = createToastStore(); 