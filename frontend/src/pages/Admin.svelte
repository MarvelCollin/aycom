<script lang="ts">
  import { onMount } from 'svelte';
  import { toastStore } from '../stores/toastStore';
  import MainLayout from '../components/layout/MainLayout.svelte';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import type { IAuthStore } from '../interfaces/IAuth';

  import * as adminAPI from '../api/admin';
  import { getAllUsers, checkAdminStatus } from '../api/user';

  import UsersIcon from 'svelte-feather-icons/src/icons/UsersIcon.svelte';
  import AlertCircleIcon from 'svelte-feather-icons/src/icons/AlertCircleIcon.svelte';
  import MessageSquareIcon from 'svelte-feather-icons/src/icons/MessageSquareIcon.svelte';
  import ShieldIcon from 'svelte-feather-icons/src/icons/ShieldIcon.svelte';
  import UserXIcon from 'svelte-feather-icons/src/icons/UserXIcon.svelte';
  import FlagIcon from 'svelte-feather-icons/src/icons/FlagIcon.svelte';
  import SettingsIcon from 'svelte-feather-icons/src/icons/SettingsIcon.svelte';
  import TrendingUpIcon from 'svelte-feather-icons/src/icons/TrendingUpIcon.svelte';
  import MailIcon from 'svelte-feather-icons/src/icons/MailIcon.svelte';
  import FolderIcon from 'svelte-feather-icons/src/icons/FolderIcon.svelte';

  import Spinner from '../components/common/Spinner.svelte';
  import TabButtons from '../components/common/TabButtons.svelte';
  import Button from '../components/common/Button.svelte';

  interface BaseResponse {
    success: boolean;
    message?: string;
  }

  interface RequestsResponse extends BaseResponse {
    requests?: any[];
    total_count?: number;
    page?: number;
    total_pages?: number;
  }

  interface CategoriesResponse extends BaseResponse {
    categories?: any[];
    total_count?: number;
    page?: number;
    total_pages?: number;
  }

  const logger = createLoggerWithPrefix('AdminDashboard');

  const { getAuthState } = useAuth();
  const { theme } = useTheme();

  $: authState = getAuthState ? (getAuthState() as IAuthStore) : { 
    user_id: null, 
    is_authenticated: false, 
    access_token: null, 
    refresh_token: null,
    is_admin: false
  };
  $: isDarkMode = $theme === 'dark';

  let isLoading = true;
  let isAdmin = false;
  let activeTab = 'overview'; 

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
    username: string;
    email: string;
    created_at: string;
    is_banned: boolean;
    is_admin: boolean;
    follower_count: number;
    profile_picture_url?: string;
  }

  interface CommunityRequest {
    id: string;
    name: string;
    description: string;
    status: string;
    created_at: string;
    requester: User;
    logo_url?: string;
    banner_url?: string;
    category_id?: string;
    user_id?: string;
  }

  interface PremiumRequest {
    id: string;
    reason: string;
    status: string;
    created_at: string;
    requester: User;
  }

  interface ReportRequest {
    id: string;
    reason: string;
    status: string;
    created_at: string;
    reporter: User;
    reported_user: User;
  }

  interface ThreadCategory {
    id: string;
    name: string;
    description: string;
    created_at: string;
  }

  interface CommunityCategory {
    id: string;
    name: string;
    description: string;
    created_at: string;
  }

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
  let communityRequests: CommunityRequest[] = [];
  let premiumRequests: PremiumRequest[] = [];
  let reportRequests: ReportRequest[] = [];
  let threadCategories: ThreadCategory[] = [];
  let communityCategories: CommunityCategory[] = [];

  let currentPage = 1;
  let totalCount = 0;
  let limit = 10;

  let newsletterSubject = '';
  let newsletterContent = '';
  let isSendingNewsletter = false;

  let newCategoryName = '';
  let newCategoryDescription = '';
  let editingCategory: ThreadCategory | CommunityCategory | null = null;
  let isProcessingRequest = false;
  let selectedRequestType = 'all';
  let requestStatusFilter = 'pending';
  let searchQuery = '';

  let showNewCategoryModal = false;
  let categoryType: 'thread' | 'community' = 'thread';

  onMount(async () => {
    try {
      logger.info("Admin.svelte mounted");
      logger.info("Auth state:", authState);

      isAdmin = await checkAdmin();

      if (isAdmin) {
        logger.info("User is admin, loading dashboard data");
        await loadDashboardData();
      } else {
        logger.warn("User is not an admin but reached admin page");
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      logger.error('Failed to load admin data', error);
      toastStore.showToast(`Failed to load admin data: ${message}`, 'error');
    }
  });

  async function checkAdmin() {
    try {

      logger.info('Checking admin status from auth state:', authState);
      if (authState.is_authenticated && authState.is_admin === true) {
        logger.info('Admin access granted based on auth state');
        return true;
      }

      logger.info('Auth state admin check failed, calling check-admin API endpoint');
      const adminCheckResult = await checkAdminStatus();

      if (adminCheckResult) {
        logger.info('Admin access granted based on API check');

        try {
          const authData = localStorage.getItem('auth');
          if (authData) {
            const auth = JSON.parse(authData);
            auth.is_admin = true;
            localStorage.setItem('auth', JSON.stringify(auth));
            logger.info('Updated localStorage with admin status');
          }
        } catch (e) {
          logger.error('Error updating auth state:', e);
        }

        return true;
      }

      logger.error('Non-admin user attempted to access admin page');
      toastStore.showToast('You do not have permission to access this page', 'error');
      window.location.href = '/feed';
      return false;
    } catch (error) {
      logger.error('Admin check failed with error:', error);
      toastStore.showToast('Failed to verify admin access', 'error');
      window.location.href = '/feed';
      return false;
    }
  }

  async function loadDashboardData() {
    try {
      isLoading = true;

      switch (activeTab) {
        case 'overview':
          await loadStatistics();
          break;
        case 'users':
          await loadUsers();
          break;
        case 'requests':
          await loadRequests();
          break;
        case 'categories':
          await loadCategories();
          break;
      }

    } catch (error) {
      logger.error('Error loading dashboard data:', error);
      toastStore.showToast('Failed to load dashboard data', 'error');
    } finally {
      isLoading = false;
    }
  }

  async function loadStatistics() {
    try {
      const response = await adminAPI.getDashboardStatistics();
      if (response.success) {
        statistics = {
          totalUsers: response.data?.total_users || 0,
          activeUsers: response.data?.active_users || 0,
          totalCommunities: response.data?.total_communities || 0,
          totalThreads: response.data?.total_threads || 0,
          pendingReports: response.data?.pending_reports || 0,
          newUsersToday: response.data?.new_users_today || 0,
          newPostsToday: response.data?.new_posts_today || 0
        };
        logger.info('Dashboard statistics loaded successfully');
      } else {
        logger.warn('Statistics API returned success: false');
        toastStore.showToast('Failed to load dashboard statistics', 'warning');
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      logger.error('Error loading statistics:', error);
      toastStore.showToast(`Statistics error: ${message}`, 'warning');

      statistics = {
        totalUsers: 0,
        activeUsers: 0,
        totalCommunities: 0,
        totalThreads: 0,
        pendingReports: 0,
        newUsersToday: 0,
        newPostsToday: 0
      };
    }
  }

  async function loadUsers() {
    try {
      logger.info(`Loading users with search: ${searchQuery}, page: ${currentPage}, limit: ${limit}`);
      const response = await getAllUsers(currentPage, limit, 'created_at', false, searchQuery);

      logger.info("Response from getAllUsers:", response);

      if (response && response.success) {
        users = response.users || [];
        totalCount = response.totalCount || 0;
        logger.info(`Loaded ${users.length} users (total: ${totalCount})`);

        if (users.length > 0) {
          logger.info(`First user data:`, users[0]);
        }
      } else {
        logger.warn('User response missing expected data structure or failed:', response.error);
        users = [];
        totalCount = 0;
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      logger.error('Error loading users:', error);
      toastStore.showToast(`Failed to load users: ${message}`, 'error');
      users = [];
      totalCount = 0;
    }
  }

  async function loadRequests() {
    try {

      communityRequests = [];
      premiumRequests = [];
      reportRequests = [];

      try {
        const communityResponse = await adminAPI.getCommunityRequests(currentPage, limit, requestStatusFilter !== 'all' ? requestStatusFilter : undefined);
        if (communityResponse.success) {
          communityRequests = communityResponse.requests || [];
          if (communityResponse.pagination && communityResponse.pagination.total_count) {
            totalCount = communityResponse.pagination.total_count;
          }
          logger.info(`Loaded ${communityRequests.length} community requests`);

          console.log('Community requests:', communityRequests);
          
          communityRequests.forEach((request, index) => {
            console.log(`Request #${index} - ${request.name}:`, { 
              fullRequest: request,
              requesterObject: request.requester,
              hasRequester: !!request.requester,
              requesterName: request.requester?.name,
              requesterUsername: request.requester?.username,
              userId: request.user_id
            });
          });
        }
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Unknown error';
        logger.error('Error loading community requests:', error);
        toastStore.showToast(`Community requests: ${message}`, 'warning');
      }

      try {
        const premiumResponse = await adminAPI.getPremiumRequests(currentPage, limit, requestStatusFilter !== 'all' ? requestStatusFilter : undefined);
        if (premiumResponse.success) {
          premiumRequests = premiumResponse.requests || [];
          logger.info(`Loaded ${premiumRequests.length} premium requests`);
        }
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Unknown error';
        logger.error('Error loading premium requests:', error);
        toastStore.showToast(`Premium requests: ${message}`, 'warning');
      }

      try {
        const reportResponse = await adminAPI.getReportRequests(currentPage, limit, requestStatusFilter !== 'all' ? requestStatusFilter : undefined);
        if (reportResponse.success) {
          reportRequests = reportResponse.requests || [];
          logger.info(`Loaded ${reportRequests.length} report requests`);
        }
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Unknown error';
        logger.error('Error loading report requests:', error);
        toastStore.showToast(`Report requests: ${message}`, 'warning');
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      logger.error('Error loading requests:', error);

      if (message.includes('permission')) {
        toastStore.showToast('You do not have sufficient permissions to view these requests', 'error');
      } else {
        toastStore.showToast('Failed to load requests', 'error');
      }
    }
  }

  async function loadCategories() {
    try {

      threadCategories = [];
      communityCategories = [];

      try {
        const threadResponse = await adminAPI.getThreadCategories(currentPage, limit);
        if (threadResponse.success) {
          threadCategories = threadResponse.categories || [];
          if (threadResponse.pagination && threadResponse.pagination.total_count) {
            totalCount = threadResponse.pagination.total_count;
          }
        }
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Unknown error';
        logger.error('Error loading thread categories:', error);
        toastStore.showToast(`Thread categories: ${message}`, 'warning');
      }

      try {
        const communityResponse = await adminAPI.getCommunityCategories(currentPage, limit);
        if (communityResponse.success) {
          communityCategories = communityResponse.categories || [];
        }
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Unknown error';
        logger.error('Error loading community categories:', error);
        toastStore.showToast(`Community categories: ${message}`, 'warning');
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      logger.error('Error loading categories:', error);

      if (message.includes('permission')) {
        toastStore.showToast('You do not have sufficient permissions to view these categories', 'error');
      } else {
        toastStore.showToast('Failed to load categories', 'error');
      }
    }
  }
    const tabItems = [
    { id: 'overview', label: 'Overview', icon: TrendingUpIcon },
    { id: 'users', label: 'Users', icon: UsersIcon },
    { id: 'requests', label: 'Requests', icon: AlertCircleIcon },
    { id: 'categories', label: 'Categories', icon: FolderIcon },
    { id: 'newsletter', label: 'Newsletter', icon: MailIcon }
  ];

  function handleTabChange(event) {
    activeTab = event.detail;
    currentPage = 1; 
    loadDashboardData();
  }

  function handlePrevPage() {
    if (currentPage > 1) {
      currentPage--;
      loadDashboardData();
    }
  }

  function handleNextPage() {
    if (currentPage < Math.ceil(totalCount / limit)) {
      currentPage++;
      loadDashboardData();
    }
  }

  async function handleBanUser(userId: string, isBanned: boolean) {
    try {
      const ban = !isBanned; 
      logger.info(`Processing ban for user ${userId} with ban=${ban}`);

      const response = await adminAPI.banUser(userId, ban, isBanned ? undefined : 'Admin action');
      if (response.success) {
        toastStore.showToast(`User ${isBanned ? 'unbanned' : 'banned'} successfully`, 'success');
        await loadUsers(); 
      } else {
        throw new Error(response.message || 'Failed to update user status');
      }
    } catch (error) {
      logger.error('Error updating user ban status:', error);
      toastStore.showToast('Failed to update user status', 'error');
    }
  }

  async function handleSendNewsletter() {
    if (!newsletterSubject.trim() || !newsletterContent.trim()) {
      toastStore.showToast('Please provide both subject and content', 'warning');
      return;
    }

    try {
      isSendingNewsletter = true;
      const response = await adminAPI.sendNewsletter(newsletterSubject, newsletterContent) as BaseResponse;

      if (response.success) {
        toastStore.showToast('Newsletter sent successfully', 'success');
        newsletterSubject = '';
        newsletterContent = '';
      } else {
        throw new Error(response.message || 'Failed to send newsletter');
      }
    } catch (error: any) {
      const newsletterError = error as Error;
      logger.error('Error sending newsletter:', newsletterError);
      if (newsletterError.message && newsletterError.message.includes('permission')) {
        toastStore.showToast('You do not have sufficient permissions to send newsletters', 'error');
      } else {
        toastStore.showToast(`Failed to send newsletter: ${newsletterError.message}`, 'error');
      }
    } finally {
      isSendingNewsletter = false;
    }
  }
  async function handleProcessCommunityRequest(requestId: string, approve: boolean) {
    try {
      isProcessingRequest = true;
      logger.info(`Processing community request ${requestId} with approve=${approve}`);

      // First check if the request still exists by refreshing data
      await loadRequests();
      
      // Check if the request still exists after refresh
      const requestExists = communityRequests.some(req => req.id === requestId);
      if (!requestExists) {
        toastStore.showToast('This request no longer exists. It may have already been processed.', 'warning');
        return;
      }

      const response = await adminAPI.processCommunityRequest(requestId, approve, approve ? 'Approved by admin' : 'Rejected by admin');
      if (response.success) {
        toastStore.showToast(`Community request ${approve ? 'approved' : 'rejected'}`, 'success');
        await loadRequests();
      } else {
        throw new Error(response.message || 'Failed to process request');
      }
    } catch (error) {
      logger.error('Error processing community request:', error);
      
      // Enhanced error handling for specific error cases
      if (error instanceof Error) {
        if (error.message.includes('404') || error.message.includes('not found')) {
          toastStore.showToast('This request no longer exists. Refreshing data...', 'warning');
          await loadRequests(); // Refresh the data to remove stale entries
        } else {
          toastStore.showToast(`Failed to process request: ${error.message}`, 'error');
        }
      } else {
        toastStore.showToast('Failed to process request', 'error');
      }
    } finally {
      isProcessingRequest = false;
    }
  }

  async function handleProcessPremiumRequest(requestId: string, approve: boolean) {
    try {
      isProcessingRequest = true;
      logger.info(`Processing premium request ${requestId} with approve=${approve}`);

      const response = await adminAPI.processPremiumRequest(requestId, approve, approve ? 'Approved by admin' : 'Rejected by admin');
      if (response.success) {
        toastStore.showToast(`Premium request ${approve ? 'approved' : 'rejected'}`, 'success');
        await loadRequests();
      } else {
        throw new Error(response.message || 'Failed to process request');
      }
    } catch (error) {
      logger.error('Error processing premium request:', error);
      toastStore.showToast('Failed to process request', 'error');
    } finally {
      isProcessingRequest = false;
    }
  }

  async function handleProcessReportRequest(requestId: string, approve: boolean) {
    try {
      isProcessingRequest = true;
      logger.info(`Processing report request ${requestId} with approve=${approve}`);

      const response = await adminAPI.processReportRequest(requestId, approve, approve ? 'Action taken by admin' : 'No action needed');
      if (response.success) {
        toastStore.showToast(`Report ${approve ? 'approved and user banned' : 'dismissed'}`, 'success');
        await loadRequests();
        await loadUsers(); 
      } else {
        throw new Error(response.message || 'Failed to process request');
      }
    } catch (error) {
      logger.error('Error processing report request:', error);
      toastStore.showToast('Failed to process request', 'error');
    } finally {
      isProcessingRequest = false;
    }
  }

  async function handleCreateCategory() {
    if (!newCategoryName.trim()) {
      toastStore.showToast('Please provide a category name', 'warning');
      return;
    }

    try {
      let response;
      if (categoryType === 'thread') {
        response = await adminAPI.createThreadCategory(newCategoryName, newCategoryDescription);
      } else {
        response = await adminAPI.createCommunityCategory(newCategoryName, newCategoryDescription);
      }

      if (response.success) {
        toastStore.showToast('Category created successfully', 'success');
        newCategoryName = '';
        newCategoryDescription = '';
        showNewCategoryModal = false;
        await loadCategories();
      } else {
        throw new Error((response as any).message || 'Failed to create category');
      }
    } catch (error) {
      logger.error('Error creating category:', error);
      toastStore.showToast('Failed to create category', 'error');
    }
  }

  async function handleUpdateCategory(categoryId: string, name: string, description: string, type: 'thread' | 'community') {
    try {
      let response;
      if (type === 'thread') {
        response = await adminAPI.updateThreadCategory(categoryId, name, description);
      } else {
        response = await adminAPI.updateCommunityCategory(categoryId, name, description);
      }

      if (response.success) {
        toastStore.showToast('Category updated successfully', 'success');
        editingCategory = null;
        await loadCategories();
      } else {
        throw new Error((response as any).message || 'Failed to update category');
      }
    } catch (error) {
      logger.error('Error updating category:', error);
      toastStore.showToast('Failed to update category', 'error');
    }
  }

  async function handleDeleteCategory(categoryId: string, type: 'thread' | 'community') {
    if (!confirm('Are you sure you want to delete this category? This action cannot be undone.')) {
      return;
    }

    try {
      let response;
      if (type === 'thread') {
        response = await adminAPI.deleteThreadCategory(categoryId);
      } else {
        response = await adminAPI.deleteCommunityCategory(categoryId);
      }

      if (response.success) {
        toastStore.showToast('Category deleted successfully', 'success');
        await loadCategories();
      } else {
        throw new Error((response as any).message || 'Failed to delete category');
      }
    } catch (error) {
      logger.error('Error deleting category:', error);
      toastStore.showToast('Failed to delete category', 'error');
    }
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

      <div class="admin-content">        {#if activeTab === 'overview'}
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
                  <span>{Math.round(statistics.activeUsers / statistics.totalUsers * 100) || 0}% of total</span>
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
                <h2>Recent Activity</h2>
              </div>

              <div class="activity-summary">
                <div class="activity-card">
                  <h4>Pending Community Requests</h4>
                  <div class="activity-value">{communityRequests.filter(r => r.status === 'pending').length}</div>
                </div>

                <div class="activity-card">
                  <h4>Pending Premium Requests</h4>
                  <div class="activity-value">{premiumRequests.filter(r => r.status === 'pending').length}</div>
                </div>

                <div class="activity-card">
                  <h4>Pending Reports</h4>
                  <div class="activity-value">{reportRequests.filter(r => r.status === 'pending').length}</div>
                </div>
              </div>
            </div>
          </div>

        {:else if activeTab === 'users'}
          <div class="users-section">
            <div class="section-header">
              <h2>Manage Users</h2>
              <div class="search-filter">
                <input 
                  type="text" 
                  placeholder="Search users..." 
                  bind:value={searchQuery}
                  on:input={() => loadUsers()}
                />
                <Button variant="outlined" on:click={() => loadUsers()}>Search</Button>
              </div>
            </div>

            <div class="users-table">
              {#if users.length === 0}
                <div class="no-results">
                  <p>No users found. {searchQuery ? 'Try a different search term.' : ''}</p>
                </div>
              {:else}
                <table>
                  <thead>
                    <tr>
                      <th>User</th>
                      <th>Email</th>
                      <th>Joined</th>
                      <th>Status</th>
                      <th>Followers</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {#each users as user}
                      <tr>
                        <td class="user-cell">
                          <div class="user-info">
                            <div class="avatar">
                              {#if user.profile_picture_url}
                                <img src={user.profile_picture_url} alt={user.name} />
                              {:else}
                                <div class="avatar-placeholder">
                                  {user.name?.charAt(0) || user.username?.charAt(0) || '?'}
                                </div>
                              {/if}
                            </div>
                            <div class="user-details">
                              <span class="name">{user.name || 'Unknown'}</span>
                              <span class="username">@{user.username || 'unknown'}</span>
                            </div>
                          </div>
                        </td>
                        <td>{user.email || 'No email'}</td>
                        <td>{formatDate(user.created_at)}</td>
                        <td>
                          {#if user.is_banned}
                            <span class="status-badge banned">Banned</span>
                          {:else if user.is_admin}
                            <span class="status-badge admin">Admin</span>
                          {:else}
                            <span class="status-badge active">Active</span>
                          {/if}
                        </td>
                        <td>{user.follower_count || 0}</td>
                        <td class="actions-cell">
                          <button 
                            class="action-btn {user.is_banned ? 'unban' : 'ban'}" 
                            on:click={() => handleBanUser(user.id, user.is_banned)}
                          >
                            {user.is_banned ? 'Unban' : 'Ban'}
                          </button>
                          <a href="/profile/{user.username}" class="view-link" target="_blank">View</a>
                        </td>
                      </tr>
                    {/each}
                  </tbody>
                </table>

                <div class="pagination">
                  <button 
                    class="pagination-btn" 
                    disabled={currentPage <= 1}
                    on:click={handlePrevPage}
                  >
                    Previous
                  </button>
                  <span class="page-info">Page {currentPage} of {Math.ceil(totalCount / limit)}</span>
                  <button 
                    class="pagination-btn" 
                    disabled={currentPage >= Math.ceil(totalCount / limit)}
                    on:click={handleNextPage}
                  >
                    Next
                  </button>
                </div>
              {/if}
            </div>
          </div>        {:else if activeTab === 'requests'}
          <div class="requests-section">
            <div class="section-header">
              <h2>Manage Requests</h2>
              <div class="section-controls">
                <div class="search-filter">
                  <select bind:value={selectedRequestType} on:change={() => loadRequests()}>
                    <option value="all">All Types</option>
                    <option value="community">Community Requests</option>
                    <option value="premium">Premium Requests</option>
                    <option value="report">Report Requests</option>
                  </select>
                  <select bind:value={requestStatusFilter} on:change={() => loadRequests()}>
                    <option value="all">All Status</option>
                    <option value="pending">Pending</option>
                    <option value="approved">Approved</option>
                    <option value="rejected">Rejected</option>
                  </select>
                </div>
                <Button 
                  variant="outlined" 
                  on:click={() => loadRequests()}
                  disabled={isLoading}
                >
                  {isLoading ? 'Refreshing...' : 'Refresh Data'}
                </Button>
              </div>
            </div>

            <!-- Community Requests -->
            {#if selectedRequestType === 'all' || selectedRequestType === 'community'}
              <div class="request-category">
                <h3>Community Creation Requests</h3>
                <div class="community-requests-grid">
                  {#each communityRequests.filter(r => requestStatusFilter === 'all' || r.status === requestStatusFilter) as request}
                    <div class="community-request-card">
                      <div class="card-header">
                        <div class="community-identity">
                          <div class="community-logo">
                            {#if request.logo_url}
                              <img src={request.logo_url} alt="Logo" />
                            {:else}
                              <div class="placeholder-logo">
                                <span>{request.name.substring(0, 1).toUpperCase()}</span>
                              </div>
                            {/if}
                          </div>
                          <div class="community-name-container">
                            <h3 class="community-name">{request.name}</h3>
                            <span class="status-badge {request.status}">{request.status}</span>
                          </div>
                        </div>
                        <div class="request-date">
                          <span class="date-label">Requested:</span>
                          <span class="date-value">{formatDate(request.created_at)}</span>
                        </div>
                      </div>
                      
                      <div class="card-body">
                        <div class="community-description">
                          <p>{request.description || 'No description provided'}</p>
                        </div>
                        
                        {#if request.banner_url}
                          <div class="community-banner">
                            <img src={request.banner_url} alt="Banner" />
                          </div>
                        {/if}
                        
                        <div class="requester-section">
                          <h4>Requested by:</h4>
                          <div class="requester-info">
                                                          <div class="requester-avatar">
                              {#if request.requester?.profile_picture_url}
                                <img src={request.requester.profile_picture_url} alt="Profile" />
                              {:else}
                                <div class="avatar-placeholder">
                                  <span>
                                    {#if request.requester?.name}
                                      {request.requester.name.substring(0, 1).toUpperCase()}
                                    {:else if request.requester?.username}
                                      {request.requester.username.substring(0, 1).toUpperCase()}
                                    {:else}
                                      U
                                    {/if}
                                  </span>
                                </div>
                              {/if}
                            </div>
                            <div class="requester-details">
                              <span class="requester-name">
                                {#if request.requester?.name}
                                  {request.requester.name}
                                {:else if request.requester?.username}
                                  {request.requester.username}
                                {:else}
                                  Unknown user
                                {/if}
                              </span>
                              {#if request.requester?.username && request.requester?.name !== request.requester?.username}
                                <span class="requester-username">@{request.requester.username}</span>
                              {/if}
                            </div>
                          </div>
                        </div>
                      </div>
                      
                      <div class="card-footer">
                        <button class="btn btn-approve" on:click={() => handleProcessCommunityRequest(request.id, true)}>
                          <span class="btn-icon">✓</span> Approve
                        </button>
                        <button class="btn btn-reject" on:click={() => handleProcessCommunityRequest(request.id, false)}>
                          <span class="btn-icon">✕</span> Reject
                        </button>
                      </div>
                    </div>
                  {:else}
                    <div class="no-requests">
                      <div class="empty-state">
                        <div class="empty-icon">📭</div>
                        <h3>No community requests</h3>
                        <p>There are no pending community requests to display.</p>
                      </div>
                    </div>
                  {/each}
                </div>
              </div>
            {/if}

            <!-- Premium Requests -->
            {#if selectedRequestType === 'all' || selectedRequestType === 'premium'}
              <div class="request-category">
                <h3>Premium User Requests</h3>
                <div class="requests-table">
                  <table>
                    <thead>
                      <tr>
                        <th>Requester</th>
                        <th>Reason</th>
                        <th>Date</th>
                        <th>Status</th>
                        <th>Actions</th>
                      </tr>
                    </thead>
                    <tbody>
                      {#each premiumRequests.filter(r => requestStatusFilter === 'all' || r.status === requestStatusFilter) as request}
                        <tr>
                          <td>{request.requester?.name || 'Unknown User'}</td>
                          <td class="description-cell">{request.reason}</td>
                          <td>{formatDate(request.created_at)}</td>
                          <td>
                            <span class="status-badge {request.status}">
                              {request.status}
                            </span>
                          </td>
                          <td>
                            <div class="action-buttons">
                              {#if request.status === 'pending'}
                                <Button 
                                  variant="primary" 
                                  size="small"
                                  disabled={isProcessingRequest}
                                  on:click={() => handleProcessPremiumRequest(request.id, true)}
                                >
                                  Approve
                                </Button>
                                <Button 
                                  variant="danger" 
                                  size="small"
                                  disabled={isProcessingRequest}
                                  on:click={() => handleProcessPremiumRequest(request.id, false)}
                                >
                                  Reject
                                </Button>
                              {/if}
                            </div>
                          </td>
                        </tr>
                      {/each}
                    </tbody>
                  </table>
                </div>
              </div>
            {/if}

            <!-- Report Requests -->
            {#if selectedRequestType === 'all' || selectedRequestType === 'report'}
              <div class="request-category">
                <h3>User Report Requests</h3>
                <div class="requests-table">
                  <table>
                    <thead>
                      <tr>
                        <th>Reporter</th>
                        <th>Reported User</th>
                        <th>Reason</th>
                        <th>Date</th>
                        <th>Status</th>
                        <th>Actions</th>
                      </tr>
                    </thead>
                    <tbody>
                      {#each reportRequests.filter(r => requestStatusFilter === 'all' || r.status === requestStatusFilter) as request}
                        <tr>
                          <td>{request.reporter?.name || 'Unknown User'}</td>
                          <td>{request.reported_user?.name || 'Unknown User'}</td>
                          <td class="description-cell">{request.reason}</td>
                          <td>{formatDate(request.created_at)}</td>
                          <td>
                            <span class="status-badge {request.status}">
                              {request.status}
                            </span>
                          </td>
                          <td>
                            <div class="action-buttons">
                              {#if request.status === 'pending'}
                                <Button 
                                  variant="danger" 
                                  size="small"
                                  disabled={isProcessingRequest}
                                  on:click={() => handleProcessReportRequest(request.id, true)}
                                >
                                  Ban User
                                </Button>
                                <Button 
                                  variant="secondary" 
                                  size="small"
                                  disabled={isProcessingRequest}
                                  on:click={() => handleProcessReportRequest(request.id, false)}
                                >
                                  Dismiss
                                </Button>
                              {/if}
                            </div>
                          </td>
                        </tr>
                      {/each}
                    </tbody>
                  </table>
                </div>
              </div>
            {/if}
          </div>

        {:else if activeTab === 'categories'}
          <div class="categories-section">
            <div class="section-header">
              <h2>Manage Categories</h2>
              <Button variant="primary" on:click={() => showNewCategoryModal = true}>
                Add Category
              </Button>
            </div>

            <!-- Thread Categories -->
            <div class="category-section">
              <h3>Thread Categories</h3>
              <div class="categories-table">
                <table>
                  <thead>
                    <tr>
                      <th>Name</th>
                      <th>Description</th>
                      <th>Created</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {#each threadCategories as category}
                      <tr>
                        <td>{category.name}</td>
                        <td class="description-cell">{category.description || '-'}</td>
                        <td>{formatDate(category.created_at)}</td>
                        <td>
                          <div class="action-buttons">
                            <Button 
                              variant="secondary" 
                              size="small"
                              on:click={() => {editingCategory = category; categoryType = 'thread'}}
                            >
                              Edit
                            </Button>
                            <Button 
                              variant="danger" 
                              size="small"
                              on:click={() => handleDeleteCategory(category.id, 'thread')}
                            >
                              Delete
                            </Button>
                          </div>
                        </td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
            </div>

            <!-- Community Categories -->
            <div class="category-section">
              <h3>Community Categories</h3>
              <div class="categories-table">
                <table>
                  <thead>
                    <tr>
                      <th>Name</th>
                      <th>Description</th>
                      <th>Created</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {#each communityCategories as category}
                      <tr>
                        <td>{category.name}</td>
                        <td class="description-cell">{category.description || '-'}</td>
                        <td>{formatDate(category.created_at)}</td>
                        <td>
                          <div class="action-buttons">
                            <Button 
                              variant="secondary" 
                              size="small"
                              on:click={() => {editingCategory = category; categoryType = 'community'}}
                            >
                              Edit
                            </Button>
                            <Button 
                              variant="danger" 
                              size="small"
                              on:click={() => handleDeleteCategory(category.id, 'community')}
                            >
                              Delete
                            </Button>
                          </div>
                        </td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
            </div>
          </div>

        {:else if activeTab === 'newsletter'}
          <div class="newsletter-section">
            <div class="section-header">
              <h2>Send Newsletter</h2>
            </div>

            <div class="newsletter-form">
              <div class="form-group">
                <label for="newsletter-subject">Subject</label>
                <input 
                  id="newsletter-subject"
                  type="text" 
                  bind:value={newsletterSubject}
                  placeholder="Enter newsletter subject..."
                  disabled={isSendingNewsletter}
                />
              </div>

              <div class="form-group">
                <label for="newsletter-content">Content</label>
                <textarea 
                  id="newsletter-content"
                  bind:value={newsletterContent}
                  placeholder="Enter newsletter content..."
                  rows="10"
                  disabled={isSendingNewsletter}
                ></textarea>
              </div>

              <div class="form-actions">
                <Button 
                  variant="primary"
                  disabled={isSendingNewsletter || !newsletterSubject.trim() || !newsletterContent.trim()}
                  on:click={handleSendNewsletter}
                >
                  {#if isSendingNewsletter}
                    <Spinner size="small" />
                    Sending...
                  {:else}
                    <MailIcon size="16" />
                    Send Newsletter
                  {/if}
                </Button>
              </div>

              <div class="newsletter-info">
                <p><strong>Note:</strong> This newsletter will be sent to all users who have subscribed to email notifications.</p>
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
      </div>    {/if}
  </div>
</MainLayout>

<!-- Category Creation Modal -->
{#if showNewCategoryModal}
  <div class="modal-overlay" on:click={() => showNewCategoryModal = false}>
    <div class="modal" on:click|stopPropagation>
      <div class="modal-header">
        <h3>Create New Category</h3>
        <button class="modal-close" on:click={() => showNewCategoryModal = false}>×</button>
      </div>

      <div class="modal-body">
        <div class="form-group">
          <label>Category Type</label>
          <select bind:value={categoryType}>
            <option value="thread">Thread Category</option>
            <option value="community">Community Category</option>
          </select>
        </div>

        <div class="form-group">
          <label for="category-name">Name</label>
          <input 
            id="category-name"
            type="text" 
            bind:value={newCategoryName}
            placeholder="Enter category name..."
          />
        </div>

        <div class="form-group">
          <label for="category-description">Description (Optional)</label>
          <textarea 
            id="category-description"
            bind:value={newCategoryDescription}
            placeholder="Enter category description..."
            rows="3"
          ></textarea>
        </div>
      </div>

      <div class="modal-footer">
        <Button variant="secondary" on:click={() => showNewCategoryModal = false}>
          Cancel
        </Button>
        <Button variant="primary" on:click={handleCreateCategory}>
          Create Category
        </Button>
      </div>
    </div>
  </div>
{/if}

<!-- Category Edit Modal -->
{#if editingCategory}
  <div class="modal-overlay" on:click={() => editingCategory = null}>
    <div class="modal" on:click|stopPropagation>
      <div class="modal-header">
        <h3>Edit Category</h3>
        <button class="modal-close" on:click={() => editingCategory = null}>×</button>
      </div>

      <div class="modal-body">
        <div class="form-group">
          <label for="edit-category-name">Name</label>
          <input 
            id="edit-category-name"
            type="text" 
            bind:value={editingCategory.name}
            placeholder="Enter category name..."
          />
        </div>

        <div class="form-group">
          <label for="edit-category-description">Description</label>
          <textarea 
            id="edit-category-description"
            bind:value={editingCategory.description}
            placeholder="Enter category description..."
            rows="3"
          ></textarea>
        </div>
      </div>

      <div class="modal-footer">
        <Button variant="secondary" on:click={() => editingCategory = null}>
          Cancel
        </Button>
        <Button 
          variant="primary" 
          on:click={() => {
            if (editingCategory) {
              handleUpdateCategory(editingCategory.id, editingCategory.name, editingCategory.description, categoryType);
            }
          }}
        >
          Update Category
        </Button>
      </div>
    </div>
  </div>
{/if}

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

  .section-controls {
    display: flex;
    align-items: center;
    gap: var(--space-3);
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

  .empty-state {
    text-align: center;
    padding: var(--space-6) 0;
    color: var(--text-secondary);
  }

  .view-more {
    text-align: center;
    margin-top: var(--space-4);
  }

  .activity-summary {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: var(--space-4);
    margin-top: var(--space-4);
  }

  .activity-card {
    background-color: var(--bg-secondary);
    border-radius: var(--radius-md);
    border: 1px solid var(--border-color);
    padding: var(--space-4);
    text-align: center;
  }

  .activity-card h4 {
    margin: 0 0 var(--space-2) 0;
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
  }

  .activity-value {
    font-size: var(--font-size-xl);
    font-weight: var(--font-weight-bold);
    color: var(--color-primary);
  }

  .request-category {
    margin-bottom: var(--space-8);
  }

  .request-category h3 {
    margin: 0 0 var(--space-4) 0;
    font-size: var(--font-size-lg);
    font-weight: var(--font-weight-medium);
    padding-bottom: var(--space-2);
    border-bottom: 1px solid var(--border-color);
  }

  .category-section {
    margin-bottom: var(--space-6);
  }

  .category-section h3 {
    margin: 0 0 var(--space-4) 0;
    font-size: var(--font-size-lg);
    font-weight: var(--font-weight-medium);
  }

  .newsletter-form {
    max-width: 800px;
    background-color: var(--bg-secondary);
    border-radius: var(--radius-md);
    border: 1px solid var(--border-color);
    padding: var(--space-6);
  }

  .form-group {
    margin-bottom: var(--space-4);
  }

  .form-group label {
    display: block;
    margin-bottom: var(--space-2);
    font-weight: var(--font-weight-medium);
    color: var(--text-primary);
  }

  .form-group input,
  .form-group textarea,
  .form-group select {
    width: 100%;
    padding: var(--space-3);
    border-radius: var(--radius-md);
    border: 1px solid var(--border-color);
    background-color: var(--bg-input);
    color: var(--text-primary);
    font-size: var(--font-size-base);
  }

  .form-group textarea {
    resize: vertical;
    min-height: 120px;
  }

  .form-actions {
    display: flex;
    gap: var(--space-3);
    margin-top: var(--space-6);
  }

  .newsletter-info {
    margin-top: var(--space-4);
    padding: var(--space-3);
    background-color: var(--bg-accent);
    border-radius: var(--radius-md);
    border-left: 4px solid var(--color-info);
  }

  .newsletter-info p {
    margin: 0;
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
  }

  .username {
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
    margin-top: var(--space-1);
  }

  .description-cell {
    max-width: 300px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .status-badge.admin {
    background-color: var(--color-primary-light);
    color: var(--color-primary);
    margin-left: var(--space-1);
  }

  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal {
    background-color: var(--bg-primary);
    border-radius: var(--radius-lg);
    border: 1px solid var(--border-color);
    width: 90%;
    max-width: 500px;
    max-height: 90vh;
    overflow-y: auto;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--space-4);
    border-bottom: 1px solid var(--border-color);
  }

  .modal-header h3 {
    margin: 0;
    font-size: var(--font-size-lg);
    font-weight: var(--font-weight-medium);
  }

  .modal-close {
    background: none;
    border: none;
    font-size: var(--font-size-xl);
    cursor: pointer;
    color: var(--text-secondary);
    padding: var(--space-1);
    border-radius: var(--radius-sm);
  }

  .modal-close:hover {
    background-color: var(--bg-accent);
    color: var(--text-primary);
  }

  .modal-body {
    padding: var(--space-4);
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-3);
    padding: var(--space-4);
    border-top: 1px solid var(--border-color);
  }

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

  .community-images {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
  }

  .community-image {
    width: 50px;
    height: 50px;
    object-fit: cover;
    border-radius: 4px;
    border: 1px solid var(--color-border);
  }

  .community-image.logo {
    border-radius: 50%;
  }

  .community-image.banner {
    width: 80px;
  }

  .community-image.placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: #f0f0f0;
    color: #888;
    font-size: 0.7rem;
    text-align: center;
    padding: 4px;
  }

  .users-table {
    width: 100%;
    overflow-x: auto;
    margin-top: var(--space-4);
  }

  .no-results {
    padding: var(--space-6);
    text-align: center;
    color: var(--text-secondary);
    background: var(--bg-secondary);
    border-radius: var(--radius-md);
    margin: var(--space-4) 0;
  }

  .users-table table {
    width: 100%;
    border-collapse: collapse;
    background-color: var(--bg-secondary);
    border-radius: var(--radius-md);
    overflow: hidden;
  }

  .users-table th {
    text-align: left;
    padding: var(--space-3) var(--space-4);
    border-bottom: 1px solid var(--border-color);
    font-weight: var(--font-weight-medium);
    color: var(--text-secondary);
    background-color: var(--bg-tertiary);
  }

  .users-table td {
    padding: var(--space-3) var(--space-4);
    border-bottom: 1px solid var(--border-color);
    vertical-align: middle;
  }

  .user-info {
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }

  .avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    overflow: hidden;
  }

  .avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .avatar-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: var(--color-primary);
    color: var(--color-primary);
    font-weight: var(--font-weight-bold);
  }

  .user-details {
    display: flex;
    flex-direction: column;
  }

  .user-details .name {
    font-weight: var(--font-weight-medium);
  }

  .user-details .username {
    color: var(--text-secondary);
    font-size: var(--font-size-sm);
  }

  .status-badge {
    display: inline-block;
    padding: var(--space-1) var(--space-2);
    border-radius: var(--radius-full);
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
    text-align: center;
  }

  .status-badge.active {
    background-color: var(--color-success-light);
    color: var(--color-success);
  }

  .status-badge.admin {
    background-color: var(--color-primary-light);
    color: var(--color-primary);
  }

  .status-badge.banned {
    background-color: var(--color-error-light);
    color: var(--color-error);
  }

  .actions-cell {
    display: flex;
    gap: var(--space-2);
  }

  .action-btn {
    padding: var(--space-1) var(--space-3);
    border-radius: var(--radius-sm);
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
    cursor: pointer;
    border: none;
    transition: all 0.2s ease;
  }

  .action-btn.ban {
    background-color: var(--color-error-light);
    color: var(--color-error);
  }

  .action-btn.ban:hover {
    background-color: var(--color-error);
    color: white;
  }

  .action-btn.unban {
    background-color: var(--color-success-light);
    color: var(--color-success);
  }

  .action-btn.unban:hover {
    background-color: var(--color-success);
    color: white;
  }

  .view-link {
    padding: var(--space-1) var(--space-3);
    background-color: var(--bg-tertiary);
    color: var(--text-primary);
    border-radius: var(--radius-sm);
    text-decoration: none;
    font-size: var(--font-size-sm);
    transition: all 0.2s ease;
  }

  .view-link:hover {
    background-color: var(--bg-accent);
  }

  .pagination {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-4);
    margin-top: var(--space-4);
  }

  .pagination-btn {
    padding: var(--space-2) var(--space-4);
    background-color: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .pagination-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .pagination-btn:not(:disabled):hover {
    background-color: var(--bg-accent);
  }

  .page-info {
    color: var(--text-secondary);
  }

  @media (max-width: 768px) {
    .users-table {
      font-size: var(--font-size-sm);
    }

    .actions-cell {
      flex-direction: column;
    }

    .user-info {
      gap: var(--space-2);
    }

    .avatar {
      width: 30px;
      height: 30px;
    }
  }

  .requester-info {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  
  .requester-avatar {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    object-fit: cover;
  }
  
  /* Community Requests Card Style */
  .community-requests-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 20px;
    margin-top: 16px;
  }
  
  .community-request-card {
    background-color: var(--bg-secondary);
    border-radius: 12px;
    border: 1px solid var(--border-color);
    overflow: hidden;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
    transition: transform 0.2s, box-shadow 0.2s;
  }
  
  .community-request-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
  }
  
  .card-header {
    padding: 16px;
    border-bottom: 1px solid var(--border-color);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  
  .community-identity {
    display: flex;
    align-items: center;
    gap: 12px;
    flex: 1;
  }
  
  .community-logo {
    width: 48px;
    height: 48px;
    border-radius: 12px;
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  .community-logo img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  
  .placeholder-logo {
    width: 100%;
    height: 100%;
    background-color: var(--color-primary);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 24px;
    font-weight: bold;
  }
  
  .community-name-container {
    display: flex;
    flex-direction: column;
  }
  
  .community-name {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
  }
  
  .status-badge {
    display: inline-block;
    padding: 2px 8px;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 500;
    margin-top: 4px;
    background-color: #f0f0f0;
  }
  
  .status-badge.pending {
    background-color: #fff8e1;
    color: #ffa000;
  }
  
  .status-badge.approved {
    background-color: #e8f5e9;
    color: #388e3c;
  }
  
  .status-badge.rejected {
    background-color: #ffebee;
    color: #d32f2f;
  }
  
  .request-date {
    font-size: 12px;
    color: var(--text-secondary);
    display: flex;
    flex-direction: column;
    align-items: flex-end;
  }
  
  .date-label {
    font-weight: 500;
  }
  
  .card-body {
    padding: 16px;
  }
  
  .community-description {
    margin-bottom: 16px;
  }
  
  .community-description p {
    margin: 0;
    color: var(--text-secondary);
    font-size: 14px;
    line-height: 1.5;
  }
  
  .community-banner {
    margin: 16px -16px;
    height: 120px;
    overflow: hidden;
  }
  
  .community-banner img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  
  .requester-section {
    margin-top: 16px;
    padding-top: 16px;
    border-top: 1px solid var(--border-color);
  }
  
  .requester-section h4 {
    margin: 0 0 8px 0;
    font-size: 14px;
    font-weight: 500;
    color: var(--text-secondary);
  }
  
  .requester-info {
    display: flex;
    align-items: center;
    gap: 12px;
  }
  
  .requester-avatar {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    overflow: hidden;
  }
  
  .requester-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  
  .avatar-placeholder {
    width: 100%;
    height: 100%;
    background-color: var(--color-primary);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 18px;
    font-weight: 500;
  }
  
  .requester-details {
    display: flex;
    flex-direction: column;
  }
  
  .requester-name {
    font-weight: 500;
    font-size: 14px;
  }
  
  .requester-username {
    font-size: 12px;
    color: var(--text-secondary);
  }
  
  .card-footer {
    display: flex;
    padding: 12px 16px;
    border-top: 1px solid var(--border-color);
    gap: 8px;
  }
  
  .btn-approve, .btn-reject {
    flex: 1;
    border: none;
    border-radius: 6px;
    padding: 8px 12px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    transition: background-color 0.2s;
  }
  
  .btn-approve {
    background-color: #e8f5e9;
    color: #388e3c;
  }
  
  .btn-approve:hover {
    background-color: #c8e6c9;
  }
  
  .btn-reject {
    background-color: #ffebee;
    color: #d32f2f;
  }
  
  .btn-reject:hover {
    background-color: #ffcdd2;
  }
  
  .btn-icon {
    font-size: 16px;
  }
  
  .no-requests {
    grid-column: 1 / -1;
  }
  
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 40px 20px;
    background-color: var(--bg-secondary);
    border-radius: 12px;
    border: 1px solid var(--border-color);
  }
  
  .empty-icon {
    font-size: 48px;
    margin-bottom: 16px;
  }
  
  .empty-state h3 {
    margin: 0 0 8px 0;
    font-size: 18px;
    font-weight: 600;
  }
  
  .empty-state p {
    margin: 0;
    color: var(--text-secondary);
    text-align: center;
  }
</style>