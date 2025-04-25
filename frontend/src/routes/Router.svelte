<script lang="ts">
  import { onMount } from 'svelte';
  import Home from './Home.svelte';
  
  let route = '/';
  
  function handleNavigation() {
    route = window.location.pathname;
  }
  
  onMount(() => {
    window.addEventListener('popstate', handleNavigation);
    handleNavigation();
    
    // Add click handler for links
    document.body.addEventListener('click', (e) => {
      const target = e.target as HTMLElement;
      const anchor = target.closest('a');
      
      if (anchor && anchor.href.includes(window.location.origin) && !anchor.hasAttribute('target')) {
        e.preventDefault();
        const href = anchor.getAttribute('href') || '/';
        if (href !== route) {
          window.history.pushState({}, '', href);
          handleNavigation();
        }
      }
    });
    
    return () => {
      window.removeEventListener('popstate', handleNavigation);
    };
  });
</script>

<main>
  {#if route === '/'}
    <Home />
  {:else}
    <div class="not-found">
      <h1>404 - Page Not Found</h1>
      <p>The page you are looking for does not exist.</p>
      <a href="/">Go back home</a>
    </div>
  {/if}
</main>

<style>
  main {
    width: 100%;
  }
  
  .not-found {
    text-align: center;
    padding: 50px 20px;
  }
  
  .not-found h1 {
    font-size: 2.5rem;
    margin-bottom: 1rem;
  }
  
  .not-found a {
    display: inline-block;
    margin-top: 1.5rem;
    padding: 0.5rem 1rem;
    background-color: #4f46e5;
    color: white;
    text-decoration: none;
    border-radius: 4px;
  }
  
  .not-found a:hover {
    background-color: #4338ca;
  }
</style> 