<script lang="ts">
  import AlertCircleIcon from "svelte-feather-icons/src/icons/AlertCircleIcon.svelte";

  interface Rule {
    id: string;
    title: string;
    description: string;
    order?: number;
  }

  export let rules: Rule[] = [];
</script>

<div class="rules-container">
  <h2 class="section-title">Community Rules</h2>
  {#if rules.length > 0}
    <div class="rules-list">
      {#each rules as rule, i (rule.id)}
        <div class="rule-item">
          <div class="rule-number">{i + 1}</div>
          <div class="rule-content">
            <h3 class="rule-title">{rule.title}</h3>
            <p class="rule-description">{rule.description}</p>
          </div>
        </div>
      {/each}
    </div>
  {:else}
    <div class="empty-state">
      <AlertCircleIcon size="48" />
      <p>No rules have been set for this community</p>
    </div>
  {/if}
</div>

<style>
  .rules-container {
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

  .rules-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  .rule-item {
    display: flex;
    gap: var(--space-3);
    padding: var(--space-3);
    background-color: var(--bg-secondary);
    border-radius: var(--radius-md);
  }

  .rule-number {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    background-color: var(--color-primary);
    color: white;
    border-radius: 50%;
    font-weight: var(--font-weight-bold);
    flex-shrink: 0;
  }

  .rule-content {
    flex: 1;
  }

  .rule-title {
    font-size: var(--font-size-lg);
    font-weight: var(--font-weight-semibold);
    color: var(--text-primary);
    margin: 0 0 var(--space-2) 0;
  }

  .rule-description {
    font-size: var(--font-size-base);
    color: var(--text-secondary);
    line-height: 1.5;
    margin: 0;
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

  .empty-state p {
    color: var(--text-secondary);
    margin: 0;
  }
</style>