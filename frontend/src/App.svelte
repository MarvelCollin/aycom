<script lang="ts">
  import './styles/magniview.css';
  
  // Import all CSS from styles/src
  import './styles/src/reset.css';
  import './styles/src/colors.css';
  import './styles/src/typography.css';
  import './styles/src/utilities.css';
  import './styles/src/components.css';
  import './styles/src/spacing-sizing.css';
  import './styles/src/flexbox.css';
  import './styles/src/grid.css';
  import './styles/src/background-border.css';
  import './styles/src/filter-effect.css';
  import './styles/src/transform.css';
  import './styles/src/transition.css';
  import './styles/src/keyframes.css';
  import './styles/src/interaction.css';
  import './styles/src/layout-flex.css';
  import './styles/src/responsive.css';
  import './styles/src/dark-mode.css';
  
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
