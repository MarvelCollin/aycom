<script lang="ts">
  import { onMount } from 'svelte';
  import appConfig from '../config/appConfig';
  import { isAuthenticated } from '../utils/auth';
  import { writable } from 'svelte/store';

  // Create page store
  export const currentPage = writable({ route: '/', userProfileId: '', communityId: '', threadId: '' });

  // Import pages
  import Landing from '../pages/Landing.svelte';
  import Login from '../pages/Login.svelte';
  import Register from '../pages/Register.svelte';
  import ForgotPassword from '../pages/ForgotPassword.svelte';
  import Feed from '../pages/Feed.svelte';
  import GoogleCallback from '../pages/GoogleCallback.svelte';
  import Explore from '../pages/Explore.svelte';
  import Notification from '../pages/Notification.svelte';
  import Message from '../pages/Message.svelte';
  import OwnProfile from '../pages/OwnProfile.svelte';
  import Bookmarks from '../pages/Bookmarks.svelte';
  import Communities from '../pages/Communities.svelte';
  import Admin from '../pages/Admin.svelte';
  import CommunityDetail from '../pages/CommunityDetail.svelte';
  import UserProfile from '../pages/UserProfile.svelte';
  import Premium from '../pages/Premium.svelte';
  import Setting from '../pages/Setting.svelte';
  import WebSocketTest from '../pages/WebSocketTest.svelte';
  import ThreadDetail from '../pages/ThreadDetail.svelte';

  let route = '/';
  let authStatus = false;
  let userProfileId = '';
  let communityId = '';
  let threadId = '';

  function handleNavigation() {
    const fullPath = window.location.pathname;
    console.log('Handling navigation to:', fullPath);

    // Extract user profile ID from URL
    const userProfileMatch = fullPath.match(/^\/user\/([^\/]+)$/);
    if (userProfileMatch) {
      userProfileId = userProfileMatch[1];
      route = '/user';
      console.log(`User profile route matched with userId: ${userProfileId}`);
      updatePageStore();
      return;
    }

    // Extract community ID from URL 
    const communityDetailMatch = fullPath.match(/^\/communities\/([^\/]+)$/);
    if (communityDetailMatch) {
      communityId = communityDetailMatch[1];
      route = '/community-detail';
      console.log(`Community detail route matched with communityId: ${communityId}`);
      updatePageStore();
      return;
    }

    // Extract thread ID from URL
    const threadDetailMatch = fullPath.match(/^\/thread\/([^\/]+)$/);
    if (threadDetailMatch) {
      threadId = threadDetailMatch[1];
      route = '/thread-detail';
      console.log(`Thread detail route matched with threadId: ${threadId}`);
      updatePageStore();
      return;
    }

    // Handle normal routes
    route = fullPath;
    updatePageStore();

    // Check authentication for protected routes
    authStatus = isAuthenticated();
    console.log(`Authentication status: ${authStatus}`);

    if (!appConfig.auth.enabled) {
      return;
    }

    if (!authStatus && 
        (route === '/feed' ||
         route === '/explore' || 
         route === '/notifications' || 
         route === '/messages' || 
         route === '/bookmarks' ||
         route === '/communities' ||
         route === '/community-detail' ||
         route === '/thread-detail' ||
         route === '/premium' ||
         route === '/profile' ||
         route === '/settings' ||
         route === '/admin' ||
         route === '/user' ||
         route === '/websocket-test')) {
      console.log('Unauthenticated access to protected route, redirecting to login');
      window.history.replaceState({}, '', '/login');
      route = '/login';
    }

    if (authStatus && 
        (route === '/login' || 
         route === '/register')) {
      console.log('Authenticated access to auth route, redirecting to feed');
      window.history.replaceState({}, '', '/feed');
      route = '/feed';
    }
  }

  // Update the page store
  function updatePageStore() {
    currentPage.set({ route, userProfileId, communityId, threadId });
  }

  // Type definition for the custom navigation event
  interface NavigateEvent {
    communityId?: string;
    threadId?: string;
    [key: string]: any;
  }

  onMount(() => {
    window.addEventListener('popstate', handleNavigation);
    
    // Define type-safe event handler for 'navigate' custom event
    function handleNavigateEvent(e: Event) {
      const event = e as CustomEvent<NavigateEvent>;
      console.log('Custom navigation event received:', event.detail);
      
      // If the event includes a communityId, set it directly
      if (event.detail && event.detail.communityId) {
        communityId = event.detail.communityId;
      }
      
      // If the event includes a threadId, set it directly
      if (event.detail && event.detail.threadId) {
        threadId = event.detail.threadId;
      }
      
      handleNavigation();
    }
    
    // Add the event listener
    window.addEventListener('navigate', handleNavigateEvent as EventListener);
    
    handleNavigation();

    document.body.addEventListener('click', (e) => {
      const target = e.target as HTMLElement;
      const anchor = target.closest('a');

      if (anchor && anchor.href.includes(window.location.origin) && !anchor.hasAttribute('target')) {
        e.preventDefault();
        const href = anchor.getAttribute('href') || '/';
        if (href !== window.location.pathname) {
          console.log(`Link click navigation to ${href}`);
          window.history.pushState({}, '', href);
          handleNavigation();
        }
      }
    });

    return () => {
      window.removeEventListener('popstate', handleNavigation);
      window.removeEventListener('navigate', handleNavigateEvent as EventListener);
    };
  });
</script>

<main>
  {#if route === '/'}
    {#if authStatus}
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
    <CommunityDetail communityId={communityId} />
  {:else if route === '/thread-detail'}
    <ThreadDetail threadId={threadId} />
  {:else if route === '/premium'}
    <Premium />
  {:else if route === '/settings'}
    <Setting />
  {:else if route === '/admin'}
    <Admin />
  {:else if route === '/websocket-test'}
    <WebSocketTest />
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