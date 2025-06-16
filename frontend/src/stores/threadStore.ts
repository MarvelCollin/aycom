import { writable } from "svelte/store";


export const currentThreadStore = writable(null);


export function setCurrentThread(threadData) {
  currentThreadStore.set(threadData);
}


export function clearCurrentThread() {
  currentThreadStore.set(null);
}