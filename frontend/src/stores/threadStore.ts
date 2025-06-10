import { writable } from 'svelte/store';

// Store for passing thread data between components
export const currentThreadStore = writable(null);

// Set the current thread data when navigating to thread detail
export function setCurrentThread(threadData) {
  currentThreadStore.set(threadData);
}

// Clear the thread data
export function clearCurrentThread() {
  currentThreadStore.set(null);
} 