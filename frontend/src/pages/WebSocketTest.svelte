<script>
import { onMount, onDestroy } from "svelte";
import { notificationWebsocketStore } from "../stores/notificationWebsocketStore";
import { authStore } from "../stores/authStore";
import { navigate } from "../utils/navigation";
import { toastStore } from "../stores/toastStore";

let connected = false;
let reconnecting = false;
let lastError = null;
let testResults = null;
let messages = [];

const unsubscribe = notificationWebsocketStore.subscribe(state => {
  connected = state.connected;
  reconnecting = state.reconnecting;
  lastError = state.lastError;
});

function handleMessage(message) {
  messages = [{
    timestamp: new Date().toISOString(),
    data: JSON.stringify(message, null, 2)
  }, ...messages].slice(0, 10); 
}

let unregisterHandler;

onMount(async () => {
  if (!authStore.isAuthenticated()) {

    toastStore.showToast("You need to log in to access this page", "warning");
    window.location.href = "/login";
    return;
  }

  unregisterHandler = notificationWebsocketStore.registerMessageHandler(handleMessage);

  notificationWebsocketStore.connect();
});

onDestroy(() => {
  unsubscribe();
  if (unregisterHandler) unregisterHandler();
  notificationWebsocketStore.disconnect();
});

async function testConnection() {
  testResults = { status: "Testing connection..." };
  testResults = await notificationWebsocketStore.testConnection();
}

function reconnect() {
  notificationWebsocketStore.connect();
}

function disconnect() {
  notificationWebsocketStore.disconnect();
}
</script>

<div class="container mx-auto p-4">
  <h1 class="text-2xl font-bold mb-6">WebSocket Connection Test</h1>

  <div class="bg-white rounded-lg shadow p-6 mb-6">
    <h2 class="text-xl font-semibold mb-4">Connection Status</h2>

    <div class="flex flex-col gap-2 mb-4">
      <div class="flex gap-2 items-center">
        <span class="font-bold">Connected:</span>
        {#if connected}
          <span class="text-green-600">Yes</span>
        {:else}
          <span class="text-red-600">No</span>
        {/if}
      </div>

      <div class="flex gap-2 items-center">
        <span class="font-bold">Reconnecting:</span>
        {#if reconnecting}
          <span class="text-yellow-600">Yes</span>
        {:else}
          <span>No</span>
        {/if}
      </div>

      <div class="flex gap-2 items-center">
        <span class="font-bold">Last Error:</span>
        {#if lastError}
          <span class="text-red-600">{lastError}</span>
        {:else}
          <span>None</span>
        {/if}
      </div>
    </div>

    <div class="flex gap-4">
      <button
        class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition"
        on:click={reconnect}
      >
        Connect/Reconnect
      </button>

      <button
        class="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 transition"
        on:click={disconnect}
      >
        Disconnect
      </button>

      <button
        class="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 transition"
        on:click={testConnection}
      >
        Test Connection
      </button>
    </div>
  </div>

  {#if testResults}
    <div class="bg-white rounded-lg shadow p-6 mb-6">
      <h2 class="text-xl font-semibold mb-4">Test Results</h2>

      <div class="flex flex-col gap-2">
        <div class="flex gap-2 items-center">
          <span class="font-bold">Status:</span>
          {#if testResults.success === true}
            <span class="text-green-600">Success</span>
          {:else if testResults.success === false}
            <span class="text-red-600">Failed</span>
          {:else}
            <span>{testResults.status}</span>
          {/if}
        </div>

        {#if testResults.message}
          <div class="flex gap-2">
            <span class="font-bold">Message:</span>
            <span>{testResults.message}</span>
          </div>
        {/if}

        {#if testResults.details}
          <div class="mt-2">
            <div class="font-bold">Details:</div>
            <pre class="bg-gray-100 p-2 rounded mt-1 overflow-x-auto text-sm">
              {testResults.details}
            </pre>
          </div>
        {/if}

        {#if testResults.error}
          <div class="mt-2">
            <div class="font-bold">Error:</div>
            <pre class="bg-gray-100 p-2 rounded mt-1 overflow-x-auto text-sm">
              {JSON.stringify(testResults.error, null, 2)}
            </pre>
          </div>
        {/if}
      </div>
    </div>
  {/if}

  <div class="bg-white rounded-lg shadow p-6">
    <h2 class="text-xl font-semibold mb-4">Received Messages</h2>

    {#if messages.length === 0}
      <p class="text-gray-500">No messages received yet.</p>
    {:else}
      <div class="overflow-y-auto max-h-96">
        {#each messages as message}
          <div class="border-b border-gray-200 py-2 last:border-b-0">
            <div class="text-xs text-gray-500 mb-1">{message.timestamp}</div>
            <pre class="bg-gray-100 p-2 rounded text-sm overflow-x-auto">
              {message.data}
            </pre>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>