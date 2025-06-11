<script lang="ts">
  // Types
  interface Community {
    id: string;
    name: string;
    description: string;
    categories?: string[];
    created_at?: Date | string;
    [key: string]: any;
  }

  // Props
  export let community: Community | null = null;
</script>

<div class="about-container">
  <h2 class="section-title">About {community?.name || 'Community'}</h2>
  
  <div class="community-description">
    <p>{community?.description || 'No description provided'}</p>
  </div>
  
  {#if community?.categories && community.categories.length > 0}
    <div class="categories-section">
      <h3>Categories</h3>
      <div class="categories-list">
        {#each community.categories as category}
          <span class="category-tag">{category}</span>
        {/each}
      </div>
    </div>
  {/if}
  
  {#if community?.created_at}
    <div class="community-metadata">
      <p>Created: {new Date(community.created_at).toLocaleDateString()}</p>
    </div>
  {/if}
</div>

<style>
  .about-container {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  .section-title {
    font-size: var(--font-size-xl);
    font-weight: var(--font-weight-bold);
    margin-bottom: var(--space-4);
    padding-bottom: var(--space-2);
    border-bottom: 1px solid var(--border-color);
  }

  .community-description {
    padding: var(--space-4);
    background-color: var(--bg-secondary);
    border-radius: var(--radius-md);
  }

  .community-description p {
    font-size: var(--font-size-base);
    line-height: 1.6;
    color: var(--text-primary);
    margin: 0;
  }

  .categories-section {
    padding: var(--space-4);
    background-color: var(--bg-secondary);
    border-radius: var(--radius-md);
  }

  .categories-section h3 {
    font-size: var(--font-size-lg);
    font-weight: var(--font-weight-semibold);
    color: var(--text-primary);
    margin: 0 0 var(--space-3) 0;
  }

  .categories-list {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }

  .category-tag {
    display: inline-block;
    padding: var(--space-1) var(--space-2);
    background-color: var(--color-primary);
    color: white;
    border-radius: var(--radius-sm);
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
  }

  .community-metadata {
    padding: var(--space-4);
    background-color: var(--bg-secondary);
    border-radius: var(--radius-md);
  }

  .community-metadata p {
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
    margin: 0;
  }
</style>
