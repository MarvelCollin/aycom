<script lang="ts">
  export let title: string;
</script>

<div class="w-full">
  <!-- Header -->
  <header class="sticky top-0 bg-opacity-95 backdrop-blur-sm z-10 pt-0 border-b border-gray-700">
    <h1 class="px-4 py-3 text-xl font-bold">{title}</h1>
  </header>
  
  <!-- Content -->
  <div>
    <slot></slot>
  </div>
</div> 