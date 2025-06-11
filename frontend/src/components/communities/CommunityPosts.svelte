<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import TweetCard from '../social/TweetCard.svelte';
  import Button from '../common/Button.svelte';
  import MessageSquareIcon from 'svelte-feather-icons/src/icons/MessageSquareIcon.svelte';

  // Types
  interface Thread {
    id: string;
    content?: string;
    timestamp?: Date;
    username?: string;
    display_name?: string;
    avatar?: string;
    [key: string]: any;
  }

  // Props
  export let threads: Thread[] = [];
  export let isMember: boolean = false;
  export let canPostInCommunity: boolean = false;
  export let communityIsApproved: boolean = true;

  // Event dispatcher
  const dispatch = createEventDispatcher();

  function handleThreadClick(event) {
    dispatch('threadClick', event.detail);
  }

  function handleCreatePost() {
    dispatch('createPost');
  }
</script>

<div class="posts-container">
  {#if threads.length > 0}
    <div class="threads-container">
      {#each threads as thread (thread.id)}        <TweetCard 
          tweet={thread as any} 
          on:click={handleThreadClick}
        />
      {/each}
    </div>
    
    {#if canPostInCommunity}
      <div class="create-post-floating">
        <Button variant="primary" icon={MessageSquareIcon} on:click={handleCreatePost}>
          Create Post
        </Button>
      </div>
    {/if}
  {:else}
    <div class="empty-state">
      <MessageSquareIcon size="48" />
      <h2>No posts yet</h2>
      <p>Be the first to post in this community!</p>
      {#if isMember && communityIsApproved}
        <Button variant="primary" on:click={handleCreatePost}>Create Post</Button>
      {:else if isMember && !communityIsApproved}
        <p class="approval-note">You can create posts once this community is approved by an admin.</p>
      {/if}
    </div>
  {/if}
</div>

<style>
  .posts-container {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  .threads-container {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-4);
    padding: var(--space-8);
    color: var(--text-secondary);
    text-align: center;
  }

  .empty-state h2 {
    font-size: var(--font-size-xl);
    font-weight: var(--font-weight-bold);
    color: var(--text-primary);
    margin: 0;
  }

  .empty-state p {
    color: var(--text-secondary);
    margin: 0;
  }

  .approval-note {
    color: var(--text-secondary);
    margin-top: var(--space-2);
    font-size: var(--font-size-sm);
    font-style: italic;
    padding: var(--space-2) var(--space-3);
    background-color: var(--bg-secondary);
    border-radius: var(--radius-md);
  }

  .create-post-floating {
    position: fixed;
    bottom: var(--space-6);
    right: var(--space-6);
    z-index: 100;
  }

  .create-post-floating :global(button) {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    border-radius: var(--radius-full);
  }

  .create-post-floating :global(button:hover) {
    box-shadow: 0 6px 16px rgba(0, 0, 0, 0.2);
    transform: translateY(-2px);
  }

  :global(.dark) .create-post-floating :global(button) {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  }

  :global(.dark) .create-post-floating :global(button:hover) {
    box-shadow: 0 6px 16px rgba(0, 0, 0, 0.4);
  }
</style>
