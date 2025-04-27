<script lang="ts">
  import './styles/magniview.css';
  import Router from './routes/Router.svelte';
  import ThemeProvider from './components/ThemeProvider.svelte';
  import { theme } from './hooks/useTheme';
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
