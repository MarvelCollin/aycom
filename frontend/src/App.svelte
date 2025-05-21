<script lang="ts">
  import Router from './base/Router.svelte';
  import ThemeProvider from './components/common/ThemeProvider.svelte';
  import { useTheme } from './hooks/useTheme';
  
  const { theme } = useTheme();
</script>

<ThemeProvider>
  <div class="app">
    <main>
      <Router />
    </main>
  </div>
</ThemeProvider>

<style>
  .app {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
    background-color: var(--bg-primary);
    color: var(--text-primary);
  }

  main {
    flex: 1;
    width: 100%;
  }
</style>
