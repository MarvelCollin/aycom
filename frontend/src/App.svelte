<script lang="ts">
  import { onMount } from 'svelte';
  import './styles/global.scss';
  import Header from './components/Header.svelte';
  import Footer from './components/Footer.svelte';
  import Router from './routes/Router.svelte';

  let theme: 'light' | 'dark' = 'light';

  onMount(() => {
    // Check for saved theme preference
    const savedTheme = localStorage.getItem('theme');
    if (savedTheme) {
      theme = savedTheme as 'light' | 'dark';
      document.body.setAttribute('data-theme', theme);
    }
  });

  function toggleTheme() {
    theme = theme === 'light' ? 'dark' : 'light';
    document.body.setAttribute('data-theme', theme);
    localStorage.setItem('theme', theme);
  }
</script>

<div class="app">
  <Header {theme} {toggleTheme} />
<main>
    <Router />
  </main>
  <Footer />
  </div>

<style lang="scss">
  .app {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
  }

  main {
    flex: 1;
    padding: 20px;
    max-width: 1200px;
    margin: 0 auto;
    width: 100%;
  }

  @media (max-width: 768px) {
    main {
      padding: 15px;
  }
  }

  @media (max-width: 480px) {
    main {
      padding: 10px;
    }
  }
</style>
