<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import UserCard from '../social/UserCard.svelte';
  import Button from '../common/Button.svelte';
  import UsersIcon from 'svelte-feather-icons/src/icons/UsersIcon.svelte';

  // Types
  interface Member {
    id: string;
    user_id: string;
    username: string;
    name: string;
    role: string;
    avatar_url?: string;
    joined_at?: Date;
    requested_at?: Date | string;
  }

  // Props
  export let members: Member[] = [];
  export let pendingMembers: Member[] = [];
  export let canManageCommunity: boolean = false;

  // Event dispatcher
  const dispatch = createEventDispatcher();

  function handleApproveJoinRequest(requestId: string) {
    dispatch('approveJoinRequest', requestId);
  }

  function handleRejectJoinRequest(requestId: string) {
    dispatch('rejectJoinRequest', requestId);
  }
</script>

<div class="members-container">
  <h2 class="section-title">Members ({members.length})</h2>
  {#if members.length > 0}
    <div class="members-grid">
      {#each members as member (member.id)}
        <UserCard 
          user={{
            id: member.user_id || member.id,
            name: member.name || member.username || 'Unknown User',
            username: member.username || `user_${(member.user_id || '').substring(0, 8)}`,
            avatar_url: member.avatar_url || '',
            role: member.role || 'member'
          }} 
        />
      {/each}
    </div>
  {:else}
    <div class="empty-state">
      <UsersIcon size="48" />
      <p>No members found</p>
    </div>
  {/if}
  
  {#if pendingMembers.length > 0 && canManageCommunity}
    <div class="pending-members-section">
      <h2 class="section-title">Pending Join Requests ({pendingMembers.length})</h2>
      <div class="members-grid">
        {#each pendingMembers as member (member.id)}
          <div class="pending-member-card">
            <div class="pending-member-header">
              <div class="user-avatar">
                {#if member.avatar_url}
                  <img src={member.avatar_url} alt={member.username || member.name} />
                {:else}
                  <div class="user-avatar-placeholder">
                    {member.username ? member.username[0].toUpperCase() : "?"}
                  </div>
                {/if}
              </div>
              <div class="user-info">
                <h3 class="user-name">{member.name || member.username || 'Unknown User'}</h3>
                <p class="user-username">@{member.username || `user_${(member.user_id || '').substring(0, 8)}`}</p>
                <span class="user-role-badge pending">Pending</span>
              </div>
            </div>
            
            <div class="pending-member-info">
              <p><strong>Requested:</strong> {member.requested_at ? new Date(member.requested_at).toLocaleDateString() : 'Unknown'}</p>
            </div>
            
            <div class="pending-member-actions">
              <Button variant="success" size="small" on:click={() => handleApproveJoinRequest(member.id)}>
                Approve
              </Button>
              <Button variant="danger" size="small" on:click={() => handleRejectJoinRequest(member.id)}>
                Reject
              </Button>
            </div>
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>

<style>
  .members-container {
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

  .members-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
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

  .empty-state p {
    color: var(--text-secondary);
    margin: 0;
  }

  .pending-members-section {
    margin-top: var(--space-12);
    border-top: 1px solid var(--border-color);
    padding-top: var(--space-6);
  }

  .pending-member-card {
    position: relative;
    width: 100%;
    border: 1px solid var(--border-color);
    border-radius: var(--border-radius);
    overflow: hidden;
    background-color: var(--bg-primary);
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.05);
    transition: all 0.2s ease;
    display: flex;
    flex-direction: column;
  }

  .pending-member-card:hover {
    transform: translateY(-3px);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  }

  .pending-member-header {
    display: flex;
    padding: var(--space-3);
    gap: var(--space-3);
    align-items: center;
  }

  .user-avatar {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    overflow: hidden;
    flex-shrink: 0;
    border: 1px solid var(--border-color);
  }

  .user-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .user-avatar-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: var(--color-primary);
    color: white;
    font-weight: var(--font-weight-bold);
    font-size: var(--font-size-lg);
  }

  .user-info {
    flex: 1;
    min-width: 0;
  }

  .user-name {
    font-size: var(--font-size-base);
    font-weight: var(--font-weight-semibold);
    color: var(--text-primary);
    margin: 0 0 var(--space-1) 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .user-username {
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
    margin: var(--space-1) 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .user-role-badge {
    display: inline-block;
    padding: var(--space-1) var(--space-2);
    border-radius: var(--radius-sm);
    font-size: var(--font-size-xs);
    text-transform: capitalize;
    background-color: rgba(255, 193, 7, 0.2);
    color: #ff9800;
  }

  .pending-member-actions {
    display: flex;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-3);
    background-color: var(--background-secondary);
    justify-content: flex-end;
    border-top: 1px solid var(--border-color);
  }

  .pending-member-info {
    padding: var(--space-2) var(--space-3);
    font-size: var(--font-size-xs);
    color: var(--text-secondary);
    border-top: 1px solid var(--border-color);
    background-color: var(--background-secondary);
  }

  .pending-member-info p {
    margin: var(--space-1) 0;
  }

  :global(.dark) .pending-member-info {
    background-color: var(--background-secondary-dark);
  }

  :global(.dark) .pending-member-actions {
    background-color: var(--background-secondary-dark);
  }

  :global(.dark) .user-role-badge {
    background-color: rgba(255, 193, 7, 0.1);
    color: #ffb74d;
  }
</style>
