<script lang="ts">
  export let isDarkMode = false;
</script>

<main class="w-full md:w-[600px] border-x {isDarkMode ? 'border-gray-800 border-opacity-80' : 'border-gray-200'} min-h-screen flex-shrink-0">
  <slot />
</main> 