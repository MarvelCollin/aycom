import { writable } from "svelte/store";
import type { Toast, ToastType, ToastPosition } from "../interfaces/IToast";

export interface ToastOptions {
  message: string;
  type?: ToastType;
  timeout?: number;
  position?: ToastPosition;
}

function createToastStore() {
  const { subscribe, update } = writable<Toast[]>([]);

  function showToast(
    messageOrOptions: string | ToastOptions,
    type: ToastType = "info",
    timeout: number = 3000,
    position: ToastPosition = "top-right"
  ) {
    const id = Math.random().toString(36).substring(2, 9);

    let toast: Toast;

    if (typeof messageOrOptions === "string") {
      toast = {
        id,
        message: messageOrOptions,
        type,
        timeout,
        position,
      };
    } else {
      toast = {
        id,
        message: messageOrOptions.message,
        type: messageOrOptions.type || "info",
        timeout: messageOrOptions.timeout || 3000,
        position: messageOrOptions.position || "top-right",
      };
    }

    update((toasts) => [...toasts, toast]);

    setTimeout(() => {
      removeToast(id);
    }, toast.timeout);

    return id;
  }

  function removeToast(id: string) {
    update((toasts) => toasts.filter((t) => t.id !== id));
  }

  return {
    subscribe,
    showToast,
    removeToast,
  };
}

export const toastStore = createToastStore();
