<script lang="ts">
  import { onMount } from 'svelte';
  import { page, updatePageStore } from '../stores/routeStore';
  import Login from '../pages/Login.svelte';
  import Register from '../pages/Register.svelte';
  import Feed from '../pages/Feed.svelte';
  import Landing from '../pages/Landing.svelte';
  import GoogleCallback from '../pages/GoogleCallback.svelte';
  import Explore from '../pages/Explore.svelte';
  import Message from '../pages/Message.svelte';
  import Notification from '../pages/Notification.svelte';
  import Bookmarks from '../pages/Bookmarks.svelte';
  import Communities from '../pages/Communities.svelte';
  import CommunityDetail from '../pages/CommunityDetail.svelte';
  import Admin from '../pages/Admin.svelte';
  import ForgotPassword from '../pages/ForgotPassword.svelte';
  import appConfig from '../config/appConfig';
  import OwnProfile from '../pages/OwnProfile.svelte';
  import UserProfile from '../pages/UserProfile.svelte';
  import Premium from '../pages/Premium.svelte';
  import Setting from '../pages/Setting.svelte';
  
  let route = '/';
  let isAuthenticated = false;
  let userProfileId = '';
  let communityId = '';
  
  function handleNavigation() {
    const fullPath = window.location.pathname;
    console.log('Handling navigation to:', fullPath);
    
    // Check for user profile route pattern
    const userProfileMatch = fullPath.match(/^\/user\/([^\/]+)$/);
    if (userProfileMatch) {
      userProfileId = userProfileMatch[1];
      route = '/user';
      console.log(`User profile route matched with userId: ${userProfileId}`);
      
      // Update the route store with the userId parameter
      updatePageStore();
      return;
    }
    
    // Check for community detail route pattern
    const communityDetailMatch = fullPath.match(/^\/communities\/([^\/]+)$/);
    if (communityDetailMatch) {
      communityId = communityDetailMatch[1];
      route = '/community-detail';
      console.log(`Community detail route matched with communityId: ${communityId}`);
      
      // Update the route store with the communityId parameter
      updatePageStore();
      return;
    }
    
    route = fullPath;
    
    // Update the route store for other routes too
    updatePageStore();
    
    isAuthenticated = localStorage.getItem('aycom_authenticated') === 'true';
    console.log(`Authentication status: ${isAuthenticated}`);
    
    // Skip auth checks if disabled in config
    if (!appConfig.auth.enabled) {
      return;
    }
    
    if (!isAuthenticated && 
        (route === '/feed' ||
         route === '/explore' || 
         route === '/notifications' || 
         route === '/messages' || 
         route === '/bookmarks' ||
         route === '/communities' ||
         route === '/community-detail' ||
         route === '/premium' ||
         route === '/profile' ||
         route === '/settings' ||
         route === '/admin' ||
         route === '/user')) {
      console.log('Unauthenticated access to protected route, redirecting to login');
      window.history.replaceState({}, '', '/login');
      route = '/login';
    }
    
    if (isAuthenticated && 
        (route === '/login' || 
         route === '/register')) {
      console.log('Authenticated access to auth route, redirecting to feed');
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
        if (href !== window.location.pathname) {
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
    {#if isAuthenticated}
      <Feed />
    {:else}
      <Landing />
    {/if}
  {:else if route === '/login'}
    <Login />
  {:else if route === '/register'}
    <Register />
  {:else if route === '/forgot-password'}
    <ForgotPassword />
  {:else if route === '/feed'}
    <Feed />
  {:else if route === '/google/' || route === '/google'}
    <GoogleCallback />
  {:else if route === '/profile'}
    <OwnProfile />
  {:else if route === '/user'}
    <UserProfile userId={userProfileId} />
  {:else if route === '/explore'}
    <Explore />
  {:else if route === '/notifications'}
    <Notification />
  {:else if route === '/messages'}
    <Message />
  {:else if route === '/bookmarks'}
    <Bookmarks />
  {:else if route === '/communities'}
    <Communities />
  {:else if route === '/community-detail'}
    <CommunityDetail />
  {:else if route === '/premium'}
    <Premium />
  {:else if route === '/settings'}
    <Setting />
  {:else if route === '/admin'}
    <Admin />
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
    background-color: var(--bg-primary);
    color: var(--text-primary);
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
  }
  
  .not-found h1 {
    font-size: var(--font-size-3xl);
    margin-bottom: var(--space-4);
  }
  
  .not-found a {
    display: inline-block;
    margin-top: var(--space-6);
    padding: var(--space-2) var(--space-4);
    background-color: var(--color-primary);
    color: white;
    text-decoration: none;
    border-radius: var(--radius-full);
    font-weight: var(--font-weight-bold);
  }
  
  .not-found a:hover {
    background-color: var(--color-primary-hover);
  }
</style>