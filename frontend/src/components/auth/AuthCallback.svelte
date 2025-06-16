<script lang="ts">
  import Logo from '../common/Logo.svelte';

  export let loading = false;
  export let success = false;
  export let error = "";
  export let redirectCountdown = 0;
</script>

<div class="max-w-md w-full p-6 bg-gray-900 rounded-lg shadow-lg border border-gray-800">
  <div class="text-center">
    <div class="mb-6 flex justify-center">
      <Logo size="large" />
    </div>

    <h1 class="text-2xl font-semibold text-white mb-4">
      Google Authentication
    </h1>

    {#if loading}
      <div class="mb-6">
        <p class="text-gray-400 mb-4">
          Processing your login...
        </p>
        <div class="flex justify-center">
          <div class="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
        </div>
      </div>
    {:else if success}
      <div class="mb-6 text-center">
        <div class="mb-4 flex justify-center">
          <div class="bg-green-500 rounded-full w-16 h-16 flex items-center justify-center text-white text-2xl">
            ✓
          </div>
        </div>
        <p class="text-green-400 text-xl font-semibold mb-2">Authentication Successful!</p>
        <p class="text-gray-400">Redirecting to your feed in {redirectCountdown} second{redirectCountdown !== 1 ? 's' : ''}...</p>
      </div>
    {:else if error}
      <div class="mb-6 text-center">
        <div class="mb-4 flex justify-center">
          <div class="bg-red-500 rounded-full w-16 h-16 flex items-center justify-center text-white text-2xl">
            ✕
          </div>
        </div>
        <p class="text-red-500 mb-4">
          {error}
        </p>
        <a 
          href="/login" 
          class="mt-4 inline-block px-6 py-2 bg-blue-500 text-white rounded-full hover:bg-blue-600 transition"
        >
          Return to login
        </a>
      </div>
    {/if}

    <p class="text-gray-500 text-sm mt-8">
      &copy; {new Date().getFullYear()} AYCOM. All rights reserved.
    </p>
  </div>
</div>