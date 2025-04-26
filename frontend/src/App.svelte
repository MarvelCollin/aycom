<script lang="ts">
  import { onMount } from 'svelte';
  import './styles/magniview.css';
  import Router from './routes/Router.svelte';

  let theme: 'light' | 'dark' = 'dark';  // Default to dark theme for Twitter clone

  onMount(() => {
    // Check for saved theme preference
    const savedTheme = localStorage.getItem('theme');
    if (savedTheme) {
      theme = savedTheme as 'light' | 'dark';
      document.body.setAttribute('data-theme', theme);
    } else {
      // Set dark theme by default for Twitter
      document.body.setAttribute('data-theme', 'dark');
    }
  });

  function toggleTheme() {
    theme = theme === 'light' ? 'dark' : 'light';
    document.body.setAttribute('data-theme', theme);
    localStorage.setItem('theme', theme);
  }
</script>

<div class="app">
  <main>
    <Router />
  </main>
</div>

<style>
  .app {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
    background-color: black;
    color: white;
  }

  main {
    flex: 1;
    width: 100%;
  }
</style>
