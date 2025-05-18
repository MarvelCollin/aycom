<script lang="ts">
  export let size: 'small' | 'medium' | 'large' = 'medium';
  export let color: 'primary' | 'secondary' | 'white' = 'primary';
  
  // Size mapping in pixels
  const sizeMap = {
    small: 16,
    medium: 24,
    large: 48
  };
  
  $: spinnerSize = sizeMap[size];
</script>

<div class="spinner-container" style="width: {spinnerSize}px; height: {spinnerSize}px;">
  <div class="spinner spinner-{color}"></div>
</div>

<style>
  .spinner-container {
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  .spinner {
    width: 100%;
    height: 100%;
    border-radius: 50%;
    position: relative;
    animation: rotate 1s linear infinite;
  }
  
  .spinner::before {
    content: "";
    box-sizing: border-box;
    position: absolute;
    inset: 0;
    border-radius: 50%;
    border: 3px solid transparent;
    animation: spinner-track 1.5s linear infinite;
  }
  
  .spinner-primary::before {
    border-top-color: var(--color-primary);
  }
  
  .spinner-secondary::before {
    border-top-color: var(--text-secondary);
  }
  
  .spinner-white::before {
    border-top-color: white;
  }
  
  @keyframes rotate {
    to {
      transform: rotate(360deg);
    }
  }
  
  @keyframes spinner-track {
    0% {
      transform: rotate(0deg);
    }
    100% {
      transform: rotate(360deg);
    }
  }
</style> 