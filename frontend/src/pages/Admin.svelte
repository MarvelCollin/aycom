<script lang="ts">
  import { onMount } from 'svelte';
  import { toastStore } from '../stores/toastStore';
  import MainLayout from '../components/layout/MainLayout.svelte';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import type { IAuthStore } from '../interfaces/IAuth';
  
  // Import icons
  import UsersIcon from 'svelte-feather-icons/src/icons/UsersIcon.svelte';
  import AlertCircleIcon from 'svelte-feather-icons/src/icons/AlertCircleIcon.svelte';
  import MessageSquareIcon from 'svelte-feather-icons/src/icons/MessageSquareIcon.svelte';
  import ShieldIcon from 'svelte-feather-icons/src/icons/ShieldIcon.svelte';
  import UserXIcon from 'svelte-feather-icons/src/icons/UserXIcon.svelte';
  import FlagIcon from 'svelte-feather-icons/src/icons/FlagIcon.svelte';
  import SettingsIcon from 'svelte-feather-icons/src/icons/SettingsIcon.svelte';
  import TrendingUpIcon from 'svelte-feather-icons/src/icons/TrendingUpIcon.svelte';

  // Components
  import Spinner from '../components/common/Spinner.svelte';
  import TabButtons from '../components/common/TabButtons.svelte';
  import Button from '../components/common/Button.svelte';
  
  const logger = createLoggerWithPrefix('AdminDashboard');
  
  const { getAuthState } = useAuth();
  const { theme } = useTheme();
  
  $: authState = getAuthState ? (getAuthState() as IAuthStore) : { 
    userId: null, 
    isAuthenticated: false, 
    accessToken: null, 
    refreshToken: null 
  };
  $: isDarkMode = $theme === 'dark';
  
  let isLoading = true;
  let isAdmin = false;
  let activeTab = 'overview'; // overview, users, communities, reports, settings

  // Define interfaces for our data models
  interface AdminStatistics {
    totalUsers: number;
    activeUsers: number;
    totalCommunities: number;
    totalThreads: number;
    pendingReports: number;
    newUsersToday: number;
    newPostsToday: number;
  }

  interface User {
    id: string;
    name: string;
    email: string;
    createdAt: string;
    status: string;
    postCount: number;
    avatar: string | null;
  }

  interface Community {
    id: string;
    name: string;
    memberCount: number;
    postCount: number;
    status: string;
    createdAt: string;
  }

  interface Report {
    id: string;
    type: string;
    reason: string;
    reportedBy: string;
    targetId: string;
    status: string;
    createdAt: string;
  }

  // Define initial state with typed arrays
  let statistics: AdminStatistics = {
    totalUsers: 0,
    activeUsers: 0,
    totalCommunities: 0,
    totalThreads: 0,
    pendingReports: 0,
    newUsersToday: 0,
    newPostsToday: 0
  };
  
  let users: User[] = [];
  let communities: Community[] = [];
  let reports: Report[] = [];
  
  onMount(async () => {
    try {
      // Check if user is admin
      checkAdmin();
      
      // Load admin dashboard data
      if (isAdmin) {
        await loadDashboardData();
      }
    } catch (error) {
      logger.error('Failed to load admin data', error);
      toastStore.showToast('Failed to load admin data', 'error');
    }
  });
  
  function checkAdmin() {
    // Check if user is logged in and has admin role
    isAdmin = authState.isAuthenticated && authState.is_admin === true;
    
    if (!isAdmin) {
      toastStore.showToast('You do not have permission to access this page', 'error');
      window.location.href = '/feed';
    }
  }
  
  async function loadDashboardData() {
    try {
      isLoading = true;
      
      // In a real app, these would be API calls
      setTimeout(() => {
        // Load statistics
        statistics = {
          totalUsers: 4823,
          activeUsers: 1256,
          totalCommunities: 128,
          totalThreads: 18472,
          pendingReports: 17,
          newUsersToday: 86,
          newPostsToday: 342
        };
        
        // Load users
        users = Array(25).fill(null).map((_, i) => ({
          id: `user_${i+1}`,
          name: `User ${i+1}`,
          email: `user${i+1}@example.com`,
          createdAt: new Date(Date.now() - Math.random() * 10000000000).toISOString(),
          status: ['active', 'suspended', 'pending'][Math.floor(Math.random() * 3)],
          postCount: Math.floor(Math.random() * 500),
          avatar: null
        }));
        
        // Load communities
        communities = Array(15).fill(null).map((_, i) => ({
          id: `comm_${i+1}`,
          name: `Community ${i+1}`,
          memberCount: Math.floor(Math.random() * 10000),
          postCount: Math.floor(Math.random() * 5000),
          status: ['active', 'pending_approval', 'suspended'][Math.floor(Math.random() * 3)],
          createdAt: new Date(Date.now() - Math.random() * 10000000000).toISOString()
        }));
        
        // Load reports
        reports = Array(17).fill(null).map((_, i) => ({
          id: `report_${i+1}`,
          type: ['thread', 'user', 'comment', 'community'][Math.floor(Math.random() * 4)],
          reason: ['spam', 'harassment', 'misinformation', 'inappropriate', 'other'][Math.floor(Math.random() * 5)],
          reportedBy: `user_${Math.floor(Math.random() * 100)}`,
          targetId: `target_${Math.floor(Math.random() * 1000)}`,
          status: ['pending', 'reviewed', 'actioned', 'dismissed'][Math.floor(Math.random() * 4)],
          createdAt: new Date(Date.now() - Math.random() * 10000000).toISOString()
        }));
        
        isLoading = false;
      }, 1000);
      
    } catch (error) {
      logger.error('Error loading dashboard data:', error);
      toastStore.showToast('Failed to load dashboard data', 'error');
    } finally {
      isLoading = false;
    }
  }
  
  const tabItems = [
    { id: 'overview', label: 'Overview', icon: TrendingUpIcon },
    { id: 'users', label: 'Users', icon: UsersIcon },
    { id: 'communities', label: 'Communities', icon: MessageSquareIcon },
    { id: 'reports', label: 'Reports', icon: FlagIcon },
    { id: 'settings', label: 'Settings', icon: SettingsIcon }
  ];
  
  function handleTabChange(event) {
    activeTab = event.detail;
  }
  
  function formatDate(dateString) {
    if (!dateString) return '';
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  }
</script>

<MainLayout>
  <div class="admin-dashboard">
    <div class="admin-header">
      <h1>Admin Dashboard</h1>
      <div class="admin-badge">
        <ShieldIcon size="16" />
        <span>Admin</span>
      </div>
    </div>
    
    {#if isLoading}
      <div class="loading-container">
        <Spinner size="large" />
      </div>
    {:else if isAdmin}
      <TabButtons
        items={tabItems}
        activeId={activeTab}
        on:tabChange={handleTabChange}
      />
      
      <div class="admin-content">
        {#if activeTab === 'overview'}
          <div class="overview-section">
            <div class="stats-grid">
              <div class="stat-card">
                <h3>Total Users</h3>
                <div class="stat-value">{statistics.totalUsers.toLocaleString()}</div>
                <div class="stat-trend positive">
                  <span>+{statistics.newUsersToday} today</span>
                </div>
              </div>
              
              <div class="stat-card">
                <h3>Active Users</h3>
                <div class="stat-value">{statistics.activeUsers.toLocaleString()}</div>
                <div class="stat-percentage">
                  <span>{Math.round(statistics.activeUsers / statistics.totalUsers * 100)}% of total</span>
                </div>
              </div>
              
              <div class="stat-card">
                <h3>Total Communities</h3>
                <div class="stat-value">{statistics.totalCommunities.toLocaleString()}</div>
              </div>
              
              <div class="stat-card">
                <h3>Total Posts</h3>
                <div class="stat-value">{statistics.totalThreads.toLocaleString()}</div>
                <div class="stat-trend positive">
                  <span>+{statistics.newPostsToday} today</span>
                </div>
              </div>
            </div>
            
            <div class="reports-overview">
              <div class="section-header">
                <h2>Reports Needing Attention</h2>
                <span class="badge warning">{statistics.pendingReports}</span>
              </div>
              
              {#if reports.filter(r => r.status === 'pending').length > 0}
                <div class="reports-table">
                  <table>
                    <thead>
                      <tr>
                        <th>Type</th>
                        <th>Reason</th>
                        <th>Date</th>
                        <th>Actions</th>
                      </tr>
                    </thead>
                    <tbody>
                      {#each reports.filter(r => r.status === 'pending').slice(0, 5) as report}
                        <tr>
                          <td>
                            <span class="report-type {report.type}">
                              {report.type}
                            </span>
                          </td>
                          <td>{report.reason}</td>
                          <td>{formatDate(report.createdAt)}</td>
                          <td>
                            <div class="action-buttons">
                              <Button variant="secondary" size="small">Review</Button>
                            </div>
                          </td>
                        </tr>
                      {/each}
                    </tbody>
                  </table>
                </div>
                
                {#if statistics.pendingReports > 5}
                  <div class="view-more">
                    <Button variant="text">View all {statistics.pendingReports} pending reports</Button>
                  </div>
                {/if}
              {:else}
                <div class="empty-state">
                  <p>No pending reports. Looking good!</p>
                </div>
              {/if}
            </div>
          </div>
        
        {:else if activeTab === 'users'}
          <div class="users-section">
            <div class="section-header">
              <h2>Manage Users</h2>
              <div class="search-filter">
                <input type="text" placeholder="Search users..." />
                <Button variant="outlined">Filter</Button>
              </div>
            </div>
            
            <div class="users-table">
              <table>
                <thead>
                  <tr>
                    <th>User</th>
                    <th>Email</th>
                    <th>Joined</th>
                    <th>Status</th>
                    <th>Posts</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {#each users as user}
                    <tr>
                      <td class="user-cell">
                        {#if user.avatar}
                          <img src={user.avatar} alt={user.name} class="user-avatar" />
                        {:else}
                          <div class="user-avatar-placeholder">
                            {user.name.charAt(0).toUpperCase()}
                          </div>
                        {/if}
                        <span>{user.name}</span>
                      </td>
                      <td>{user.email}</td>
                      <td>{formatDate(user.createdAt)}</td>
                      <td>
                        <span class="status-badge {user.status}">
                          {user.status}
                        </span>
                      </td>
                      <td>{user.postCount}</td>
                      <td>
                        <div class="action-buttons">
                          <Button variant="secondary" size="small">Edit</Button>
                          {#if user.status !== 'suspended'}
                            <Button variant="danger" size="small">
                              <UserXIcon size="14" />
                              <span>Suspend</span>
                            </Button>
                          {:else}
                            <Button variant="secondary" size="small">Reactivate</Button>
                          {/if}
                        </div>
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
            
            <div class="pagination">
              <Button variant="outlined" size="small" disabled>Previous</Button>
              <span class="page-info">Page 1 of 10</span>
              <Button variant="outlined" size="small">Next</Button>
            </div>
          </div>
          
        {:else if activeTab === 'communities'}
          <div class="communities-section">
            <div class="section-header">
              <h2>Manage Communities</h2>
              <div class="search-filter">
                <input type="text" placeholder="Search communities..." />
                <Button variant="outlined">Filter</Button>
              </div>
            </div>
            
            <div class="communities-table">
              <table>
                <thead>
                  <tr>
                    <th>Community</th>
                    <th>Members</th>
                    <th>Posts</th>
                    <th>Created</th>
                    <th>Status</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {#each communities as community}
                    <tr>
                      <td>{community.name}</td>
                      <td>{community.memberCount}</td>
                      <td>{community.postCount}</td>
                      <td>{formatDate(community.createdAt)}</td>
                      <td>
                        <span class="status-badge {community.status}">
                          {community.status.replace('_', ' ')}
                        </span>
                      </td>
                      <td>
                        <div class="action-buttons">
                          <Button variant="secondary" size="small">View</Button>
                          {#if community.status === 'pending_approval'}
                            <Button variant="primary" size="small">Approve</Button>
                          {/if}
                          {#if community.status !== 'suspended'}
                            <Button variant="danger" size="small">Suspend</Button>
                          {:else}
                            <Button variant="secondary" size="small">Reactivate</Button>
                          {/if}
                        </div>
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
            
            <div class="pagination">
              <Button variant="outlined" size="small" disabled>Previous</Button>
              <span class="page-info">Page 1 of 2</span>
              <Button variant="outlined" size="small">Next</Button>
            </div>
          </div>
          
        {:else if activeTab === 'reports'}
          <div class="reports-section">
            <div class="section-header">
              <h2>Content Reports</h2>
              <div class="search-filter">
                <select>
                  <option value="all">All Types</option>
                  <option value="thread">Posts</option>
                  <option value="comment">Comments</option>
                  <option value="user">Users</option>
                  <option value="community">Communities</option>
                </select>
                <select>
                  <option value="all">All Status</option>
                  <option value="pending">Pending</option>
                  <option value="reviewed">Reviewed</option>
                  <option value="actioned">Actioned</option>
                  <option value="dismissed">Dismissed</option>
                </select>
              </div>
            </div>
            
            <div class="reports-table">
              <table>
                <thead>
                  <tr>
                    <th>Type</th>
                    <th>Reason</th>
                    <th>Reported By</th>
                    <th>Target ID</th>
                    <th>Date</th>
                    <th>Status</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {#each reports as report}
                    <tr>
                      <td>
                        <span class="report-type {report.type}">
                          {report.type}
                        </span>
                      </td>
                      <td>{report.reason}</td>
                      <td>{report.reportedBy}</td>
                      <td>{report.targetId}</td>
                      <td>{formatDate(report.createdAt)}</td>
                      <td>
                        <span class="status-badge {report.status}">
                          {report.status}
                        </span>
                      </td>
                      <td>
                        <div class="action-buttons">
                          <Button variant="secondary" size="small">View</Button>
                          {#if report.status === 'pending'}
                            <Button variant="danger" size="small">Action</Button>
                            <Button variant="outlined" size="small">Dismiss</Button>
                          {/if}
                        </div>
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
            
            <div class="pagination">
              <Button variant="outlined" size="small" disabled>Previous</Button>
              <span class="page-info">Page 1 of 2</span>
              <Button variant="outlined" size="small">Next</Button>
            </div>
          </div>
          
        {:else if activeTab === 'settings'}
          <div class="settings-section">
            <h2>Admin Settings</h2>
            
            <div class="setting-card">
              <h3>Content Moderation</h3>
              <div class="setting-row">
                <div class="setting-label">
                  <p>Auto-flag content with sensitive keywords</p>
                  <p class="setting-description">Automatically flag content that contains words from the sensitive words list</p>
                </div>
                <div class="setting-control">
                  <label class="toggle">
                    <input type="checkbox" checked />
                    <span class="toggle-slider"></span>
                  </label>
                </div>
              </div>
              
              <div class="setting-row">
                <div class="setting-label">
                  <p>Require approval for new communities</p>
                  <p class="setting-description">New communities must be manually approved by admins</p>
                </div>
                <div class="setting-control">
                  <label class="toggle">
                    <input type="checkbox" checked />
                    <span class="toggle-slider"></span>
                  </label>
                </div>
              </div>
            </div>
            
            <div class="setting-card">
              <h3>User Management</h3>
              <div class="setting-row">
                <div class="setting-label">
                  <p>Allow email registrations</p>
                  <p class="setting-description">Users can register with email and password</p>
                </div>
                <div class="setting-control">
                  <label class="toggle">
                    <input type="checkbox" checked />
                    <span class="toggle-slider"></span>
                  </label>
                </div>
              </div>
              
              <div class="setting-row">
                <div class="setting-label">
                  <p>Allow OAuth registrations</p>
                  <p class="setting-description">Users can register with Google, Twitter, etc.</p>
                </div>
                <div class="setting-control">
                  <label class="toggle">
                    <input type="checkbox" checked />
                    <span class="toggle-slider"></span>
                  </label>
                </div>
              </div>
              
              <div class="setting-row">
                <div class="setting-label">
                  <p>Auto-suspend users with 3+ reports</p>
                  <p class="setting-description">Users will be automatically suspended when reported 3 or more times</p>
                </div>
                <div class="setting-control">
                  <label class="toggle">
                    <input type="checkbox" />
                    <span class="toggle-slider"></span>
                  </label>
                </div>
              </div>
            </div>
          </div>
        {/if}
      </div>
    {:else}
      <div class="error-container">
        <AlertCircleIcon size="48" />
        <h2>Access Denied</h2>
        <p>You do not have permission to access this page.</p>
        <a href="/feed" class="back-link">Back to Feed</a>
      </div>
    {/if}
  </div>
</MainLayout>

<style>
  .admin-dashboard {
    width: 100%;
    max-width: 100%;
    min-height: 100vh;
    padding: var(--space-4);
  }
  
  .admin-header {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    margin-bottom: var(--space-6);
  }
  
  .admin-header h1 {
    font-size: var(--font-size-2xl);
    font-weight: var(--font-weight-bold);
    margin: 0;
  }
  
  .admin-badge {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    padding: var(--space-1) var(--space-3);
    background-color: var(--color-primary);
    color: white;
    border-radius: var(--radius-full);
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
  }
  
  .loading-container,
  .error-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: var(--space-10) var(--space-4);
    gap: var(--space-4);
    text-align: center;
  }
  
  .error-container h2 {
    font-size: var(--font-size-xl);
    font-weight: var(--font-weight-bold);
    margin: var(--space-2) 0;
  }
  
  .error-container p {
    color: var(--text-secondary);
    margin-bottom: var(--space-4);
  }
  
  .back-link {
    display: inline-block;
    padding: var(--space-2) var(--space-4);
    background-color: var(--color-primary);
    color: white;
    border-radius: var(--radius-full);
    text-decoration: none;
    font-weight: var(--font-weight-medium);
  }
  
  .admin-content {
    margin-top: var(--space-4);
  }
  
  /* Stats Grid */
  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    gap: var(--space-4);
    margin-bottom: var(--space-8);
  }
  
  .stat-card {
    padding: var(--space-4);
    background-color: var(--bg-secondary);
    border-radius: var(--radius-md);
    border: 1px solid var(--border-color);
  }
  
  .stat-card h3 {
    margin-top: 0;
    margin-bottom: var(--space-2);
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
    color: var(--text-secondary);
  }
  
  .stat-value {
    font-size: var(--font-size-2xl);
    font-weight: var(--font-weight-bold);
    margin-bottom: var(--space-2);
  }
  
  .stat-trend {
    font-size: var(--font-size-sm);
    display: flex;
    align-items: center;
    gap: var(--space-1);
  }
  
  .stat-trend.positive {
    color: var(--color-success);
  }
  
  .stat-trend.negative {
    color: var(--color-danger);
  }
  
  .stat-percentage {
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
  }
  
  /* Section Headers */
  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: var(--space-4);
  }
  
  .section-header h2 {
    font-size: var(--font-size-xl);
    font-weight: var(--font-weight-bold);
    margin: 0;
  }
  
  .badge {
    display: inline-block;
    padding: var(--space-1) var(--space-2);
    border-radius: var(--radius-full);
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
  }
  
  .badge.warning {
    background-color: var(--color-warning);
    color: var(--color-warning-contrast);
  }
  
  /* Tables */
  .reports-table,
  .users-table,
  .communities-table {
    width: 100%;
    margin-bottom: var(--space-4);
    overflow-x: auto;
  }
  
  table {
    width: 100%;
    border-collapse: collapse;
  }
  
  th {
    text-align: left;
    padding: var(--space-3);
    border-bottom: 2px solid var(--border-color);
    font-weight: var(--font-weight-medium);
    color: var(--text-secondary);
  }
  
  td {
    padding: var(--space-3);
    border-bottom: 1px solid var(--border-color);
    vertical-align: middle;
  }
  
  .user-cell {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  
  .user-avatar,
  .user-avatar-placeholder {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    background-color: var(--bg-accent);
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: var(--font-weight-bold);
  }
  
  .status-badge {
    display: inline-flex;
    padding: var(--space-1) var(--space-2);
    border-radius: var(--radius-full);
    font-size: var(--font-size-xs);
    text-transform: capitalize;
  }
  
  .status-badge.active,
  .status-badge.reviewed {
    background-color: var(--color-success-light);
    color: var(--color-success);
  }
  
  .status-badge.suspended,
  .status-badge.actioned {
    background-color: var(--color-danger-light);
    color: var(--color-danger);
  }
  
  .status-badge.pending,
  .status-badge.pending_approval {
    background-color: var(--color-warning-light);
    color: var(--color-warning);
  }
  
  .status-badge.dismissed {
    background-color: var(--bg-accent);
    color: var(--text-secondary);
  }
  
  .action-buttons {
    display: flex;
    gap: var(--space-2);
    flex-wrap: wrap;
  }
  
  /* Report Types */
  .report-type {
    display: inline-flex;
    padding: var(--space-1) var(--space-2);
    border-radius: var(--radius-full);
    font-size: var(--font-size-xs);
    text-transform: capitalize;
  }
  
  .report-type.thread {
    background-color: var(--color-primary-light);
    color: var(--color-primary);
  }
  
  .report-type.comment {
    background-color: var(--color-info-light);
    color: var(--color-info);
  }
  
  .report-type.user {
    background-color: var(--color-warning-light);
    color: var(--color-warning);
  }
  
  .report-type.community {
    background-color: var(--color-success-light);
    color: var(--color-success);
  }
  
  /* Search & Filter */
  .search-filter {
    display: flex;
    gap: var(--space-2);
  }
  
  .search-filter input,
  .search-filter select {
    padding: var(--space-2) var(--space-3);
    border-radius: var(--radius-md);
    border: 1px solid var(--border-color);
    background-color: var(--bg-input);
    color: var(--text-primary);
    font-size: var(--font-size-sm);
  }
  
  /* Pagination */
  .pagination {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: var(--space-4);
    margin-top: var(--space-6);
  }
  
  .page-info {
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
  }
  
  /* Empty State */
  .empty-state {
    text-align: center;
    padding: var(--space-6) 0;
    color: var(--text-secondary);
  }
  
  /* View More */
  .view-more {
    text-align: center;
    margin-top: var(--space-4);
  }
  
  /* Settings */
  .setting-card {
    background-color: var(--bg-secondary);
    border-radius: var(--radius-md);
    border: 1px solid var(--border-color);
    padding: var(--space-4);
    margin-bottom: var(--space-6);
  }
  
  .setting-card h3 {
    font-size: var(--font-size-lg);
    font-weight: var(--font-weight-medium);
    margin-top: 0;
    margin-bottom: var(--space-4);
    padding-bottom: var(--space-2);
    border-bottom: 1px solid var(--border-color);
  }
  
  .setting-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--space-3) 0;
    border-bottom: 1px solid var(--border-color-subtle);
  }
  
  .setting-row:last-child {
    border-bottom: none;
  }
  
  .setting-label p {
    margin: 0;
  }
  
  .setting-description {
    color: var(--text-secondary);
    font-size: var(--font-size-xs);
    margin-top: var(--space-1) !important;
  }
  
  /* Toggle Switch */
  .toggle {
    position: relative;
    display: inline-block;
    width: 50px;
    height: 26px;
  }
  
  .toggle input {
    opacity: 0;
    width: 0;
    height: 0;
  }
  
  .toggle-slider {
    position: absolute;
    cursor: pointer;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: var(--bg-accent);
    transition: .4s;
    border-radius: 34px;
  }
  
  .toggle-slider:before {
    position: absolute;
    content: "";
    height: 18px;
    width: 18px;
    left: 4px;
    bottom: 4px;
    background-color: white;
    transition: .4s;
    border-radius: 50%;
  }
  
  input:checked + .toggle-slider {
    background-color: var(--color-primary);
  }
  
  input:checked + .toggle-slider:before {
    transform: translateX(24px);
  }
  
  @media (max-width: 768px) {
    .stats-grid {
      grid-template-columns: 1fr;
    }
    
    .section-header {
      flex-direction: column;
      align-items: flex-start;
      gap: var(--space-3);
    }
    
    .search-filter {
      width: 100%;
    }
    
    .search-filter input,
    .search-filter select {
      width: 100%;
    }
    
    table {
      min-width: 600px;
    }
    
    .reports-table,
    .users-table,
    .communities-table {
      overflow-x: auto;
    }
  }
</style>
