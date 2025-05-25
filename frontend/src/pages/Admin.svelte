<script lang="ts">
  import { onMount } from 'svelte';
  import { toastStore } from '../stores/toastStore';
  import MainLayout from '../components/layout/MainLayout.svelte';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import type { IAuthStore } from '../interfaces/IAuth';
  
  // Import admin API functions
  import * as adminAPI from '../api/admin';
  import type { BaseResponse, RequestsResponse, CategoriesResponse } from '../api/admin';
  import { getAllUsers } from '../api/user';
  
  // Import icons
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

  // Components
  import Spinner from '../components/common/Spinner.svelte';
  import TabButtons from '../components/common/TabButtons.svelte';
  import Button from '../components/common/Button.svelte';
  
  // Define interfaces for API responses
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
    userId: null, 
    isAuthenticated: false, 
    accessToken: null, 
    refreshToken: null 
  };
  $: isDarkMode = $theme === 'dark';
  
  let isLoading = true;
  let isAdmin = false;
  let activeTab = 'overview'; // overview, users, requests, categories, newsletter

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
  let communityRequests: CommunityRequest[] = [];
  let premiumRequests: PremiumRequest[] = [];
  let reportRequests: ReportRequest[] = [];
  let threadCategories: ThreadCategory[] = [];
  let communityCategories: CommunityCategory[] = [];
  
  // Pagination state
  let currentPage = 1;
  let totalCount = 0;
  let limit = 10;
  
  // Newsletter state
  let newsletterSubject = '';
  let newsletterContent = '';
  let isSendingNewsletter = false;
  
  // Category management state
  let newCategoryName = '';
  let newCategoryDescription = '';
  let editingCategory: ThreadCategory | CommunityCategory | null = null;
  let isProcessingRequest = false;
  let selectedRequestType = 'all';
  let requestStatusFilter = 'pending';
  let searchQuery = '';
  
  // UI state
  let showNewCategoryModal = false;
  let categoryType: 'thread' | 'community' = 'thread';
  
  onMount(async () => {
    try {
      logger.info("Admin.svelte mounted");
      logger.info("Auth state:", authState);
      
      // Check if user is admin
      checkAdmin();
      
      // Load admin dashboard data
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
  
  function checkAdmin() {
    // Check if user is logged in and has admin role
    logger.info('Checking admin status:', authState);
    isAdmin = authState.isAuthenticated && authState.is_admin === true;
    
    if (!isAdmin) {
      logger.error('Non-admin user attempted to access admin page:', 
        { isAuthenticated: authState.isAuthenticated, is_admin: authState.is_admin });
      toastStore.showToast('You do not have permission to access this page', 'error');
      window.location.href = '/feed';
    } else {
      logger.info('Admin access granted');
    }
  }
  
  async function loadDashboardData() {
    try {
      isLoading = true;
      
      // Load initial data based on active tab
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
          totalUsers: response.totalUsers || 0,
          activeUsers: response.activeUsers || 0,
          totalCommunities: response.totalCommunities || 0,
          totalThreads: response.totalThreads || 0,
          pendingReports: response.pendingReports || 0,
          newUsersToday: response.newUsersToday || 0,
          newPostsToday: response.newPostsToday || 0
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
      
      // Initialize with zeros instead of mock data
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
      const response = await getAllUsers(limit, currentPage, 'created_at', false, searchQuery);
      
      if (response && response.success) {
        users = response.users || [];
        totalCount = response.totalCount || 0;
        logger.info(`Loaded ${users.length} users (total: ${totalCount})`);
      } else {
        logger.warn('User response missing expected data structure or failed');
        users = [];
        totalCount = 0;
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      logger.error('Error loading users:', error);
      toastStore.showToast(`Failed to load users: ${message}`, 'error');
      users = [];
    }
  }
  
  async function loadRequests() {
    try {
      // Reset request loading state
      communityRequests = [];
      premiumRequests = [];
      reportRequests = [];
      
      // Load community requests
      try {
        const communityResponse = await adminAPI.getCommunityRequests(currentPage, limit);
        if (communityResponse.success) {
          communityRequests = communityResponse.requests || [];
          if (communityResponse.total_count) {
            totalCount = communityResponse.total_count;
          }
        }
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Unknown error';
        logger.error('Error loading community requests:', error);
        toastStore.showToast(`Community requests: ${message}`, 'warning');
      }
      
      // Load premium requests
      try {
        const premiumResponse = await adminAPI.getPremiumRequests(currentPage, limit);
        if (premiumResponse.success) {
          premiumRequests = premiumResponse.requests || [];
        }
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Unknown error';
        logger.error('Error loading premium requests:', error);
        toastStore.showToast(`Premium requests: ${message}`, 'warning');
      }
      
      // Load report requests
      try {
        const reportResponse = await adminAPI.getReportRequests(currentPage, limit);
        if (reportResponse.success) {
          reportRequests = reportResponse.requests || [];
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
      // Reset category loading state
      threadCategories = [];
      communityCategories = [];
      
      // Load thread categories
      try {
        const threadResponse = await adminAPI.getThreadCategories(currentPage, limit);
        if (threadResponse.success) {
          threadCategories = threadResponse.categories || [];
          if (threadResponse.total_count) {
            totalCount = threadResponse.total_count;
          }
        }
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Unknown error';
        logger.error('Error loading thread categories:', error);
        toastStore.showToast(`Thread categories: ${message}`, 'warning');
      }
      
      // Load community categories
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
    currentPage = 1; // Reset pagination when changing tabs
    loadDashboardData();
  }
  
  // User Management Functions
  async function handleBanUser(userId: string, isBanned: boolean) {
    try {
      const response = await adminAPI.banUser(userId, !isBanned, isBanned ? undefined : 'Admin action');
      if (response.success) {
        toastStore.showToast(`User ${isBanned ? 'unbanned' : 'banned'} successfully`, 'success');
        await loadUsers(); // Reload users
      } else {
        throw new Error(response.message || 'Failed to update user status');
      }
    } catch (error) {
      logger.error('Error updating user ban status:', error);
      toastStore.showToast('Failed to update user status', 'error');
    }
  }
  
  // Newsletter Functions
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
  
  // Request Management Functions
  async function handleProcessCommunityRequest(requestId: string, approve: boolean) {
    try {
      isProcessingRequest = true;
      const response = await adminAPI.processCommunityRequest(requestId, approve, approve ? 'Approved by admin' : 'Rejected by admin');
      if (response.success) {
        toastStore.showToast(`Community request ${approve ? 'approved' : 'rejected'}`, 'success');
        await loadRequests();
      } else {
        throw new Error(response.message || 'Failed to process request');
      }
    } catch (error) {
      logger.error('Error processing community request:', error);
      toastStore.showToast('Failed to process request', 'error');
    } finally {
      isProcessingRequest = false;
    }
  }
  
  async function handleProcessPremiumRequest(requestId: string, approve: boolean) {
    try {
      isProcessingRequest = true;
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
      const response = await adminAPI.processReportRequest(requestId, approve, approve ? 'Action taken by admin' : 'No action needed');
      if (response.success) {
        toastStore.showToast(`Report ${approve ? 'approved and user banned' : 'dismissed'}`, 'success');
        await loadRequests();
        await loadUsers(); // Refresh users as someone might have been banned
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
  
  // Category Management Functions
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
        throw new Error(response.message || 'Failed to create category');
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
        throw new Error(response.message || 'Failed to update category');
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
        throw new Error(response.message || 'Failed to delete category');
      }
    } catch (error) {
      logger.error('Error deleting category:', error);
      toastStore.showToast('Failed to delete category', 'error');
    }
  }
  
  // Pagination Functions
  function handlePrevPage() {
    if (currentPage > 1) {
      currentPage--;
      loadDashboardData();
    }
  }
  
  function handleNextPage() {
    const totalPages = Math.ceil(totalCount / limit);
    if (currentPage < totalPages) {
      currentPage++;
      loadDashboardData();
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
                        {#if user.profile_picture_url}
                          <img src={user.profile_picture_url} alt={user.name || 'User'} class="user-avatar" />
                        {:else}
                          <div class="user-avatar-placeholder">
                            {(user.name || user.username || 'U').charAt(0).toUpperCase()}
                          </div>
                        {/if}
                        <div>
                          <div>{user.name || user.username || 'Unknown User'}</div>
                          <div class="username">@{user.username || 'unknown'}</div>
                        </div>
                      </td>
                      <td>{user.email}</td>
                      <td>{formatDate(user.created_at)}</td>
                      <td>
                        <span class="status-badge {user.is_banned ? 'suspended' : 'active'}">
                          {user.is_banned ? 'Banned' : 'Active'}
                        </span>
                        {#if user.is_admin}
                          <span class="status-badge admin">Admin</span>
                        {/if}
                      </td>
                      <td>{user.follower_count}</td>
                      <td>
                        <div class="action-buttons">
                          {#if !user.is_banned}
                            <Button 
                              variant="danger" 
                              size="small"
                              on:click={() => handleBanUser(user.id, user.is_banned)}
                            >
                              <UserXIcon size="14" />
                              <span>Ban</span>
                            </Button>
                          {:else}
                            <Button 
                              variant="secondary" 
                              size="small"
                              on:click={() => handleBanUser(user.id, user.is_banned)}
                            >
                              Unban
                            </Button>
                          {/if}
                        </div>
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
            
            <div class="pagination">
              <Button 
                variant="outlined" 
                size="small" 
                disabled={currentPage === 1}
                on:click={handlePrevPage}
              >
                Previous
              </Button>
              <span class="page-info">Page {currentPage} of {Math.ceil(totalCount / limit) || 1}</span>
              <Button 
                variant="outlined" 
                size="small"
                disabled={currentPage >= Math.ceil(totalCount / limit)}
                on:click={handleNextPage}
              >
                Next
              </Button>
            </div>
          </div>
          
        {:else if activeTab === 'requests'}
          <div class="requests-section">
            <div class="section-header">
              <h2>Manage Requests</h2>
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
            </div>
            
            <!-- Community Requests -->
            {#if selectedRequestType === 'all' || selectedRequestType === 'community'}
              <div class="request-category">
                <h3>Community Creation Requests</h3>
                <div class="requests-table">
                  <table>
                    <thead>
                      <tr>
                        <th>Community Name</th>
                        <th>Description</th>
                        <th>Requester</th>
                        <th>Date</th>
                        <th>Status</th>
                        <th>Actions</th>
                      </tr>
                    </thead>
                    <tbody>
                      {#each communityRequests.filter(r => requestStatusFilter === 'all' || r.status === requestStatusFilter) as request}
                        <tr>
                          <td>{request.name}</td>
                          <td class="description-cell">{request.description}</td>
                          <td>{request.requester?.name || 'Unknown User'}</td>
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
                                  on:click={() => handleProcessCommunityRequest(request.id, true)}
                                >
                                  Approve
                                </Button>
                                <Button 
                                  variant="danger" 
                                  size="small"
                                  disabled={isProcessingRequest}
                                  on:click={() => handleProcessCommunityRequest(request.id, false)}
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
  
  /* Activity Summary */
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
  
  /* Request Categories */
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
  
  /* Category Management */
  .category-section {
    margin-bottom: var(--space-6);
  }
  
  .category-section h3 {
    margin: 0 0 var(--space-4) 0;
    font-size: var(--font-size-lg);
    font-weight: var(--font-weight-medium);
  }
  
  /* Newsletter Form */
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
  
  /* Username styling */
  .username {
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
    margin-top: var(--space-1);
  }
  
  /* Description cell styling */
  .description-cell {
    max-width: 300px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  
  /* Admin badge */
  .status-badge.admin {
    background-color: var(--color-primary-light);
    color: var(--color-primary);
    margin-left: var(--space-1);
  }
  
  /* Modal Styles */
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
