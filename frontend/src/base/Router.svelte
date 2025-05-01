<script lang="ts">
  import { onMount } from 'svelte';
  import Home from '../pages/Home.svelte';
  import Landing from '../pages/Landing.svelte';
  import Login from '../pages/Login.svelte';
  import Register from '../pages/Register.svelte';
  import Feed from '../pages/Feed.svelte';
  import Profile from '../pages/Profile.svelte';
  import GoogleCallback from '../pages/GoogleCallback.svelte';
  import appConfig from '../config/appConfig';
  
  let route = '/';
  let isAuthenticated = false; 
  
  function handleNavigation() {
    route = window.location.pathname;
    
    isAuthenticated = localStorage.getItem('aycom_authenticated') === 'true';
    
    // Skip auth checks if disabled in config
    if (!appConfig.auth.enabled) {
      return;
    }
    
    if (!isAuthenticated && 
        (route === '/home' || 
         route === '/feed' ||
         route === '/explore' || 
         route === '/notifications' || 
         route === '/messages' || 
         route === '/profile')) {
      window.history.replaceState({}, '', '/');
      route = '/';
    }
    
    if (isAuthenticated && 
        (route === '/login' || 
         route === '/register' || 
         route === '/')) {
      window.history.replaceState({}, '', '/feed');
      route = '/feed';
    }
  }
  
  function setAuthenticated(value: boolean): void {
    localStorage.setItem('aycom_authenticated', value.toString());
    isAuthenticated = value;
    handleNavigation();
  }
  
  onMount(() => {
    (window as any).login = () => setAuthenticated(true);
    (window as any).logout = () => setAuthenticated(false);
    
    window.addEventListener('popstate', handleNavigation);
    handleNavigation();
    
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
    <Landing />
  {:else if route === '/login'}
    <Login />
  {:else if route === '/register'}
    <Register />
  {:else if route === '/home'}
    <Feed {route} />
  {:else if route === '/feed'}
    <Feed {route} />
  {:else if route === '/google/' || route === '/google'}
    <GoogleCallback />
  {:else if route === '/profile'}
    <Profile />
  {:else if route === '/explore' || route === '/notifications' || route === '/messages'}
    <Feed {route} />
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
    background-color: #000;
    color: #fff;
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
  }
  
  .not-found h1 {
    font-size: 2.5rem;
    margin-bottom: 1rem;
  }
  
  .not-found a {
    display: inline-block;
    margin-top: 1.5rem;
    padding: 0.5rem 1rem;
    background-color: #1d9bf0;
    color: white;
    text-decoration: none;
    border-radius: 9999px;
    font-weight: bold;
  }
  
  .not-found a:hover {
    background-color: #1a8cd8;
  }
</style>