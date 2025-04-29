import { writable } from 'svelte/store';

export type ToastType = 'info' | 'success' | 'warning' | 'error';

// Export the interface
export interface ToastState {
  message: string;
  type: ToastType;
  visible: boolean;
  duration: number;
  id: number; // Unique ID to trigger reactivity even if message/type are same
}

const defaultDuration = 4000; // Default duration in ms
let toastTimeoutId: NodeJS.Timeout | null = null;

// Initial state
const initialState: ToastState = {
  message: '',
  type: 'info',
  visible: false,
  duration: defaultDuration,
  id: 0
};

const { subscribe, set, update } = writable<ToastState>(initialState);

// Function to show a toast
function showToast(message: string, type: ToastType = 'error', duration: number = defaultDuration) {
  // Clear any existing timeout
  if (toastTimeoutId) {
    clearTimeout(toastTimeoutId);
  }

  // Set the new toast state
  set({
    message,
    type,
    visible: true,
    duration,
    id: Date.now() // Use timestamp as unique ID
  });

  // Set a timeout to hide the toast
  toastTimeoutId = setTimeout(() => {
    hideToast();
  }, duration);
}

// Function to hide the toast
function hideToast() {
  update(state => ({ ...state, visible: false }));
  if (toastTimeoutId) {
    clearTimeout(toastTimeoutId);
    toastTimeoutId = null;
  }
}

// Export the store and functions
export const toastStore = {
  subscribe,
  showToast,
  hideToast
}; 