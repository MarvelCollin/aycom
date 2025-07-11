<script lang="ts">
  import { onMount } from "svelte";
  import { toastStore } from "../stores/toastStore";
  import MainLayout from "../components/layout/MainLayout.svelte";
  import { createLoggerWithPrefix } from "../utils/logger";
  import { useAuth } from "../hooks/useAuth";
  import { useTheme } from "../hooks/useTheme";
  import type { IAuthStore } from "../interfaces/IAuth";
  import { isUserAdmin } from "../utils/auth";
  import {
    standardizeCommunityRequest,
    standardizePremiumRequest,
    standardizeReportRequest,
    standardizeUser
  } from "../utils/standardizeApiData";
  import "../styles/pages/admin.css";

  import * as adminAPI from "../api/admin";
  import type { RequestsResponse, CategoriesResponse } from "../api/admin";
  import { getAllUsers, checkAdminStatus } from "../api/user";

  import UsersIcon from "svelte-feather-icons/src/icons/UsersIcon.svelte";
  import AlertCircleIcon from "svelte-feather-icons/src/icons/AlertCircleIcon.svelte";
  import MessageSquareIcon from "svelte-feather-icons/src/icons/MessageSquareIcon.svelte";
  import ShieldIcon from "svelte-feather-icons/src/icons/ShieldIcon.svelte";
  import UserXIcon from "svelte-feather-icons/src/icons/UserXIcon.svelte";
  import FlagIcon from "svelte-feather-icons/src/icons/FlagIcon.svelte";
  import SettingsIcon from "svelte-feather-icons/src/icons/SettingsIcon.svelte";
  import TrendingUpIcon from "svelte-feather-icons/src/icons/TrendingUpIcon.svelte";
  import MailIcon from "svelte-feather-icons/src/icons/MailIcon.svelte";
  import FolderIcon from "svelte-feather-icons/src/icons/FolderIcon.svelte";
  import RefreshCwIcon from "svelte-feather-icons/src/icons/RefreshCwIcon.svelte";

  import Spinner from "../components/common/Spinner.svelte";
  import TabButtons from "../components/common/TabButtons.svelte";
  import Button from "../components/common/Button.svelte";

  interface BaseResponse {
    success: boolean;
    message?: string;
  }

  const logger = createLoggerWithPrefix("AdminDashboard");

  const { getAuthState } = useAuth();
  const { theme } = useTheme();

  $: authState = getAuthState ? (getAuthState() as IAuthStore) : {
    user_id: null,
    is_authenticated: false,
    access_token: null,
    refresh_token: null,
    is_admin: false
  };
  $: isDarkMode = $theme === "dark";

  let isLoading = true;
  let isAdmin = false;
  let activeTab = "overview";

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
  const limit = 10;

  let newsletterSubject = "";
  let newsletterContent = "";
  let isSendingNewsletter = false;
  let newsletterSubscribers: User[] = [];
  let isLoadingSubscribers = false;
  let subscribersPage = 1;
  let subscribersPagination: any = null;

  let newCategoryName = "";
  let newCategoryDescription = "";
  let editingCategory: ThreadCategory | CommunityCategory | null = null;
  let isProcessingRequest = false;
  let selectedRequestType = "all";
  let requestStatusFilter = "pending";
  let searchQuery = "";

  let showNewCategoryModal = false;
  let categoryType: "thread" | "community" = "thread";

  let isLoadingCommunityRequests = false;
  let isLoadingPremiumRequests = false;
  let isLoadingReportRequests = false;
  let isLoadingCategories = false;

  let communityRequestsTotal = 0;
  const communityRequestsPagination: any = null;
  let premiumRequestsPagination: any = null;
  let reportRequestsPagination: any = null;
  let categoriesPagination: any = null;

  const communityRequestsPage = 1;
  const premiumRequestsPage = 1;
  const reportRequestsPage = 1;
  const categoriesPage = 1;

  let reportStatusFilter = "pending";

  let isSyncingCommunities = false;

  async function syncCommunityRequests() {
    try {
      isSyncingCommunities = true;
      logger.info("Syncing community requests between services");

      const result = await adminAPI.syncCommunityRequests();

      if (result && result.success) {
        const syncData = result.data;
        logger.info("Community sync completed successfully:", syncData);

        const message = `Sync completed: Found ${syncData.total_pending_communities} pending communities`;
        toastStore.showToast(message, "success");

        await loadCommunityRequests();

        logger.info("After sync, community requests:", communityRequests);
      } else {
        logger.error("Failed to sync community requests:", result);
        toastStore.showToast("Failed to sync community requests", "error");
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : "Unknown error";
      logger.error("Error syncing community requests:", error);
      toastStore.showToast(`Error syncing community requests: ${message}`, "error");
    } finally {
      isSyncingCommunities = false;
    }
  }

  onMount(async () => {
    try {
      logger.info("Admin.svelte mounted");
      logger.info("Auth state:", authState);

      isAdmin = await checkAdmin();

      if (isAdmin) {
        await loadDashboardData();

        if (activeTab !== "categories") {
          logger.info("Pre-loading categories data for better user experience");
          await loadCategories();
        }
      } else {

        logger.warn("User is not an admin");
        window.location.href = "/feed";
      }
    } catch (error) {
      logger.error("Error in onMount:", error);
      isLoading = false;
      toastStore.showToast("Error loading admin dashboard. Please try again later.", "error");
    } finally {

      isLoading = false;
    }
  });

  async function checkAdmin() {
    try {
      logger.info("Checking admin status with authState:", authState);

      if (authState && authState.is_admin === true) {
        logger.info("User is admin according to auth state");
        return true;
      }

      logger.info("Checking admin status via API call");
      const isAdmin = await checkAdminStatus();

      if (isAdmin) {
        logger.info("API confirmed user is admin");

        try {
          const authData = localStorage.getItem("auth");
          if (authData) {
            const auth = JSON.parse(authData);
            auth.is_admin = true;
            localStorage.setItem("auth", JSON.stringify(auth));
            logger.info("Updated localStorage with admin status");
          }
        } catch (e) {
          logger.error("Error updating auth state:", e);
        }

        return true;
      } else {
        logger.warn("User is not an admin according to API");
        return false;
      }
    } catch (error) {
      logger.error("Error checking admin status:", error);
      toastStore.showToast("Error checking admin status. Redirecting to feed.", "error");
      window.location.href = "/feed";
      return false;
    }
  }

  async function loadDashboardData() {
    try {
      isLoading = true;

      switch (activeTab) {
        case "overview":
          await loadStatistics();
          break;
        case "users":
          await loadUsers();
          break;
        case "requests":
          await loadAllRequests();
          break;
        case "categories":
          await loadCategories();
          break;
        case "newsletter":
          await loadNewsletterSubscribers();
          break;
      }

    } catch (error) {
      logger.error("Error loading dashboard data:", error);
      toastStore.showToast("Failed to load dashboard data", "error");
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
        logger.info("Dashboard statistics loaded successfully");
      } else {
        logger.warn("Statistics API returned success: false");
        toastStore.showToast("Failed to load dashboard statistics", "warning");
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : "Unknown error";
      logger.error("Error loading statistics:", error);
      toastStore.showToast(`Statistics error: ${message}`, "warning");

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
      const response = await getAllUsers(currentPage, limit, "created_at", false, searchQuery);

      logger.info("Response from getAllUsers:", response);

      if (response && response.success) {
        users = response.users || [];
        totalCount = response.totalCount || 0;
        logger.info(`Loaded ${users.length} users (total: ${totalCount})`);

        users = users.map(user => {

          const isBanned =
            user.is_banned === true ||
            String(user.is_banned) === "true" ||
            Number(user.is_banned) === 1;

          return {
            ...user,
            is_banned: isBanned
          };
        });

        if (users.length > 0) {
          logger.info("First user data:", users[0]);
          logger.info(`Ban status of first user: ${users[0].is_banned}`);
        }
      } else {
        logger.warn("User response missing expected data structure or failed:", response.error);
        users = [];
        totalCount = 0;
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : "Unknown error";
      logger.error("Error loading users:", error);
      toastStore.showToast(`Failed to load users: ${message}`, "error");
      users = [];
      totalCount = 0;
    }
  }

  async function loadCommunityRequests() {
    try {
      isLoadingCommunityRequests = true;
      logger.info(`Loading community requests (page: ${communityRequestsPage}, status: ${requestStatusFilter})`);

      const result = await adminAPI.getCommunityRequests(communityRequestsPage, 10, requestStatusFilter === "all" ? undefined : requestStatusFilter);
      logger.info("Community requests API response:", result);

      if (result && result.success) {

        if (result.requests && Array.isArray(result.requests)) {
          logger.info(`Found ${result.requests.length} community requests in 'requests' field`);
          communityRequests = result.requests.map(request => standardizeCommunityRequest(request));
          communityRequestsTotal = (result as any).total_count || 0;
        } else if (result.data && Array.isArray(result.data)) {
          logger.info(`Found ${result.data.length} community requests in 'data' field`);
          communityRequests = result.data.map(request => standardizeCommunityRequest(request));
          communityRequestsTotal = result.pagination?.total_count || 0;
        } else {
          logger.warn("Community requests response has unexpected format:", result);
          communityRequests = [];
          communityRequestsTotal = 0;
        }

        logger.info(`Processed ${communityRequests.length} community requests:`, communityRequests);
      } else {
        logger.error("Failed to load community requests:", result);
        toastStore.showToast("Failed to load community requests", "error");
        communityRequests = [];
        communityRequestsTotal = 0;
      }
    } catch (error) {
      logger.error("Error loading community requests:", error);
      toastStore.showToast("Error loading community requests", "error");
      communityRequests = [];
      communityRequestsTotal = 0;
    } finally {
      isLoadingCommunityRequests = false;
    }
  }

  async function loadPremiumRequests() {
    try {
      isLoadingPremiumRequests = true;
      const result = await adminAPI.getPremiumRequests(premiumRequestsPage, 10, requestStatusFilter === "all" ? undefined : requestStatusFilter);

      if (result && result.success) {

        if (result.requests && Array.isArray(result.requests)) {
          premiumRequests = result.requests.map(request => standardizePremiumRequest(request));
        } else if (result.data && Array.isArray(result.data)) {
          premiumRequests = result.data.map(request => standardizePremiumRequest(request));
        } else {
          premiumRequests = [];
          logger.warn("No valid premium requests data found in API response");
        }

        premiumRequestsPagination = result.pagination;
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : "Unknown error";
      logger.error("Failed to load premium requests:", error);
      toastStore.showToast(`Failed to load premium requests: ${message}`, "error");
    } finally {
      isLoadingPremiumRequests = false;
    }
  }

  async function loadReportRequests() {
    try {
      isLoadingReportRequests = true;
      const result = await adminAPI.getReportRequests(reportRequestsPage, 10, reportStatusFilter === "all" ? undefined : reportStatusFilter);

      if (result && result.success) {

        if (result.requests && Array.isArray(result.requests)) {
          reportRequests = result.requests.map(request => standardizeReportRequest(request));
        } else if (result.data && Array.isArray(result.data)) {
          reportRequests = result.data.map(request => standardizeReportRequest(request));
        } else {
          reportRequests = [];
          logger.warn("No valid report requests data found in API response");
        }

        reportRequestsPagination = result.pagination;
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : "Unknown error";
      logger.error("Failed to load report requests:", error);
      toastStore.showToast(`Failed to load report requests: ${message}`, "error");
    } finally {
      isLoadingReportRequests = false;
    }
  }

  async function loadThreadCategories() {
    try {
      isLoadingCategories = true;
      const result = await adminAPI.getThreadCategories(categoriesPage, 10);

      if (result && result.success) {
        threadCategories = result.data || [];
        categoriesPagination = result.pagination;
      }
    } catch (error) {
      console.error("Failed to load thread categories:", error);
      toastStore.showToast("Failed to load thread categories. Please try again.", "error");
    } finally {
      isLoadingCategories = false;
    }
  }

  async function loadCategories() {
    logger.info("Loading both thread and community categories");
    isLoadingCategories = true;

    try {

      const [threadResult, communityResult] = await Promise.allSettled([
        adminAPI.getThreadCategories(categoriesPage, 50),
        adminAPI.getCommunityCategories(categoriesPage, 50)
      ]);

      if (threadResult.status === "fulfilled" && threadResult.value && threadResult.value.success) {
        threadCategories = threadResult.value.data || [];
        logger.info(`Loaded ${threadCategories.length} thread categories`);
      } else {
        const error = threadResult.status === "rejected" ? threadResult.reason : "Unknown error";
        logger.warn("Failed to load thread categories:", error);
        threadCategories = [];
      }

      if (communityResult.status === "fulfilled" && communityResult.value && communityResult.value.success) {
        communityCategories = communityResult.value.data || [];
        logger.info(`Loaded ${communityCategories.length} community categories`);
      } else {
        const error = communityResult.status === "rejected" ? communityResult.reason : "Unknown error";
        logger.warn("Failed to load community categories:", error);
        communityCategories = [];
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : "Unknown error";
      logger.error("Error loading categories:", error);
      toastStore.showToast(`Failed to load categories: ${message}`, "error");

      threadCategories = [];
      communityCategories = [];
    } finally {
      isLoadingCategories = false;
    }
  }

  async function loadNewsletterSubscribers() {
    try {
      isLoadingSubscribers = true;
      logger.info("Loading newsletter subscribers");

      const response = await adminAPI.getNewsletterSubscribers(subscribersPage, 10);

      if (response && response.success) {
        newsletterSubscribers = response.data || [];
        subscribersPagination = response.pagination;
        logger.info(`Loaded ${newsletterSubscribers.length} newsletter subscribers`);
      } else {
        logger.warn("Failed to load newsletter subscribers");
        newsletterSubscribers = [];
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : "Unknown error";
      logger.error("Error loading newsletter subscribers:", error);
      toastStore.showToast(`Failed to load subscribers: ${message}`, "error");
      newsletterSubscribers = [];
    } finally {
      isLoadingSubscribers = false;
    }
  }

  function handleNextSubscribersPage() {
    if (subscribersPagination && subscribersPage < subscribersPagination.total_pages) {
      subscribersPage++;
      loadNewsletterSubscribers();
    }
  }

  function handlePrevSubscribersPage() {
    if (subscribersPage > 1) {
      subscribersPage--;
      loadNewsletterSubscribers();
    }
  }

  const tabItems = [
    { id: "overview", label: "Overview", icon: TrendingUpIcon },
    { id: "users", label: "Users", icon: UsersIcon },
    { id: "requests", label: "Requests", icon: AlertCircleIcon },
    { id: "categories", label: "Categories", icon: FolderIcon },
    { id: "newsletter", label: "Newsletter", icon: MailIcon }
  ];

  function handleTabChange(event) {
    activeTab = event.detail;
    currentPage = 1;

    logger.info(`Tab changed to: ${activeTab}`);

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
      logger.info(`Processing ${ban ? "ban" : "unban"} for user ${userId} with current ban status=${isBanned}`);

      const actionText = ban ? "Banning" : "Unbanning";
      toastStore.showToast(`${actionText} user...`, "info");

      const response = await adminAPI.banUser(userId, ban, ban ? "Admin ban action" : "Admin unban action");
      logger.info("Ban/unban API response:", response);

      if (response.success) {
        const actionCompleted = ban ? "banned" : "unbanned";
        toastStore.showToast(`User ${actionCompleted} successfully`, "success");

        logger.info("Reloading users list to reflect updated ban status");
        await loadUsers();
      } else {
        throw new Error(response.message || `Failed to ${ban ? "ban" : "unban"} user`);
      }
    } catch (error) {
      logger.error("Error updating user ban status:", error);
      toastStore.showToast(`Failed to update user status: ${error instanceof Error ? error.message : "Unknown error"}`, "error");

      await loadUsers();
    }
  }

  async function handleSendNewsletter() {
    if (!newsletterSubject.trim() || !newsletterContent.trim()) {
      toastStore.showToast("Please provide both subject and content", "warning");
      return;
    }

    try {
      isSendingNewsletter = true;
      const response = await adminAPI.sendNewsletter(newsletterSubject, newsletterContent) as BaseResponse;

      if (response.success) {
        toastStore.showToast("Newsletter sent successfully", "success");
        newsletterSubject = "";
        newsletterContent = "";
      } else {
        throw new Error(response.message || "Failed to send newsletter");
      }
    } catch (error: any) {
      const newsletterError = error as Error;
      logger.error("Error sending newsletter:", newsletterError);
      if (newsletterError.message && newsletterError.message.includes("permission")) {
        toastStore.showToast("You do not have sufficient permissions to send newsletters", "error");
      } else {
        toastStore.showToast(`Failed to send newsletter: ${newsletterError.message}`, "error");
      }
    } finally {
      isSendingNewsletter = false;
    }
  }
  async function handleProcessCommunityRequest(requestId: string, approve: boolean) {
    try {
      isProcessingRequest = true;
      logger.info(`Processing community request ${requestId} with approve=${approve}`);

      await loadAllRequests();

      const requestExists = communityRequests.some(req => req.id === requestId);
      if (!requestExists) {
        toastStore.showToast("This request no longer exists. It may have already been processed.", "warning");
        return;
      }

      const response = await adminAPI.processCommunityRequest(requestId, approve, approve ? "Approved by admin" : "Rejected by admin");
      if (response.success) {
        toastStore.showToast(`Community ${approve ? "approved" : "rejected"} successfully`, "success");

        await loadCommunityRequests();

        if (approve) {

          setTimeout(() => {
            loadAllRequests();
          }, 1000);
        }
      } else {
        logger.error("Failed to process community request:", response);
        toastStore.showToast(`Failed to ${approve ? "approve" : "reject"} community: ${response.message || "Unknown error"}`, "error");
      }
    } catch (error) {
      logger.error("Error processing community request:", error);
      toastStore.showToast(`Error ${approve ? "approving" : "rejecting"} community`, "error");
    } finally {
      isProcessingRequest = false;
    }
  }

  async function handleProcessPremiumRequest(requestId: string, approve: boolean) {
    try {
      isProcessingRequest = true;
      logger.info(`Processing premium request ${requestId} with approve=${approve}`);

      const response = await adminAPI.processPremiumRequest(requestId, approve, approve ? "Approved by admin" : "Rejected by admin");
      if (response.success) {
        toastStore.showToast(`Premium request ${approve ? "approved" : "rejected"}`, "success");
        await loadAllRequests();
      } else {
        throw new Error(response.message || "Failed to process request");
      }
    } catch (error) {
      logger.error("Error processing premium request:", error);
      toastStore.showToast("Failed to process request", "error");
    } finally {
      isProcessingRequest = false;
    }
  }

  async function handleProcessReportRequest(requestId: string, approve: boolean) {
    try {
      isProcessingRequest = true;
      logger.info(`Processing report request ${requestId} with approve=${approve}`);

      const response = await adminAPI.processReportRequest(requestId, approve, approve ? "Action taken by admin" : "No action needed");
      if (response.success) {
        toastStore.showToast(`Report ${approve ? "approved and user banned" : "dismissed"}`, "success");
        await loadAllRequests();
        await loadUsers();
      } else {
        throw new Error(response.message || "Failed to process request");
      }
    } catch (error) {
      logger.error("Error processing report request:", error);
      toastStore.showToast("Failed to process request", "error");
    } finally {
      isProcessingRequest = false;
    }
  }

  async function handleCreateCategory() {
    if (!newCategoryName.trim()) {
      toastStore.showToast("Please provide a category name", "warning");
      return;
    }

    try {
      let response;
      if (categoryType === "thread") {
        response = await adminAPI.createThreadCategory(newCategoryName, newCategoryDescription);
      } else {
        response = await adminAPI.createCommunityCategory(newCategoryName, newCategoryDescription);
      }

      if (response.success) {
        toastStore.showToast("Category created successfully", "success");
        newCategoryName = "";
        newCategoryDescription = "";
        showNewCategoryModal = false;
        await loadCategories();
      } else {
        throw new Error((response as any).message || "Failed to create category");
      }
    } catch (error) {
      logger.error("Error creating category:", error);
      toastStore.showToast("Failed to create category", "error");
    }
  }

  async function handleUpdateCategory(categoryId: string, name: string, description: string, type: "thread" | "community") {
    try {
      let response;
      if (type === "thread") {
        response = await adminAPI.updateThreadCategory(categoryId, name, description);
      } else {
        response = await adminAPI.updateCommunityCategory(categoryId, name, description);
      }

      if (response.success) {
        toastStore.showToast("Category updated successfully", "success");
        editingCategory = null;
        await loadCategories();
      } else {
        throw new Error((response as any).message || "Failed to update category");
      }
    } catch (error) {
      logger.error("Error updating category:", error);
      toastStore.showToast("Failed to update category", "error");
    }
  }

  async function handleDeleteCategory(categoryId: string, type: "thread" | "community") {
    if (!confirm("Are you sure you want to delete this category? This action cannot be undone.")) {
      return;
    }

    try {
      let response;
      if (type === "thread") {
        response = await adminAPI.deleteThreadCategory(categoryId);
      } else {
        response = await adminAPI.deleteCommunityCategory(categoryId);
      }

      if (response.success) {
        toastStore.showToast("Category deleted successfully", "success");
        await loadCategories();
      } else {
        throw new Error((response as any).message || "Failed to delete category");
      }
    } catch (error) {
      logger.error("Error deleting category:", error);
      toastStore.showToast("Failed to delete category", "error");
    }
  }

  function formatDate(dateString) {
    if (!dateString) return "";
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric"
    });
  }

  async function loadAllRequests() {
    loadCommunityRequests();
    loadPremiumRequests();
    loadReportRequests();
  }

  $: if (activeTab === "requests") {
    logger.info("Request tab activated, loading requests data");
    loadAllRequests();
  } else if (activeTab === "categories") {
    logger.info("Categories tab activated, loading categories data");
    loadCategories();
  } else if (activeTab === "newsletter") {
    logger.info("Newsletter tab activated, loading subscribers");
    loadNewsletterSubscribers();
  }

  function handleRequestStatusFilterChange(event) {
    requestStatusFilter = event.target.value;
    loadAllRequests();
  }

  function handleReportStatusFilterChange(event) {
    reportStatusFilter = event.target.value;
    loadReportRequests();
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

      <div class="admin-content">        {#if activeTab === "overview"}
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
                  <div class="activity-value">{communityRequests.filter(r => r.status === "pending").length}</div>
                </div>

                <div class="activity-card">
                  <h4>Pending Premium Requests</h4>
                  <div class="activity-value">{premiumRequests.filter(r => r.status === "pending").length}</div>
                </div>

                <div class="activity-card">
                  <h4>Pending Reports</h4>
                  <div class="activity-value">{reportRequests.filter(r => r.status === "pending").length}</div>
                </div>
              </div>
            </div>
          </div>

        {:else if activeTab === "users"}
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
                  <p>No users found. {searchQuery ? "Try a different search term." : ""}</p>
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
                                <img src={user.profile_picture_url} alt={user.username} />
                              {:else}
                                <div class="avatar-placeholder">
                                  {user.username?.charAt(0).toUpperCase() || "?"}
                                </div>
                              {/if}
                            </div>
                            <div class="user-details">
                              <span class="name">{user.name || user.username || "Unknown"}</span>
                              <span class="username">@{user.username || "unknown"}</span>
                            </div>
                          </div>
                        </td>
                        <td>{user.email || "No email"}</td>
                        <td>{formatDate(user.created_at)}</td>
                        <td>
                          {#if user.is_banned === true}
                            <span class="status-badge banned">Banned</span>
                          {:else if user.is_admin === true}
                            <span class="status-badge admin">Admin</span>
                          {:else}
                            <span class="status-badge active">Active</span>
                          {/if}
                        </td>
                        <td>{user.follower_count || 0}</td>
                        <td class="actions-cell">
                          <button
                            class="action-btn {user.is_banned === true ? "unban" : "ban"}"
                            on:click={() => handleBanUser(user.id, user.is_banned === true)}
                          >
                            {user.is_banned === true ? "Unban" : "Ban"}
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
          </div>        {:else if activeTab === "requests"}
          <div class="requests-section">
            <div class="section-header">
              <h2>Manage Requests</h2>
              <div class="section-controls">
                <div class="search-filter">
                  <select bind:value={selectedRequestType} on:change={() => loadAllRequests()}>
                    <option value="all">All Types</option>
                    <option value="community">Community Requests</option>
                    <option value="premium">Premium Requests</option>
                    <option value="report">Report Requests</option>
                  </select>
                  <select bind:value={requestStatusFilter} on:change={() => loadAllRequests()}>
                    <option value="all">All Status</option>
                    <option value="pending">Pending</option>
                    <option value="approved">Approved</option>
                    <option value="rejected">Rejected</option>
                  </select>
                </div>
                <Button
                  variant="outlined"
                  on:click={() => loadAllRequests()}
                  disabled={isLoading}
                >
                  {isLoading ? "Refreshing..." : "Refresh Data"}
                </Button>
              </div>
            </div>

            <!-- Community Requests -->
            {#if selectedRequestType === "all" || selectedRequestType === "community"}
              <div class="request-category">
                <div class="category-header-with-actions">
                  <h3>Community Creation Requests</h3>
                  <Button
                    variant="outlined"
                    on:click={syncCommunityRequests}
                    disabled={isSyncingCommunities}
                    color="primary"
                  >
                    {#if isSyncingCommunities}
                      <Spinner size="small" />
                    {:else}
                      <RefreshCwIcon size="16" />
                    {/if}
                    {isSyncingCommunities ? "Syncing..." : "Sync Communities"}
                  </Button>
                </div>
                <div class="community-requests-grid">
                  {#each communityRequests.filter(r => requestStatusFilter === "all" || r.status === requestStatusFilter) as request}
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
                          <p>{request.description || "No description provided"}</p>
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
            {#if selectedRequestType === "all" || selectedRequestType === "premium"}
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
                      {#each premiumRequests.filter(r => requestStatusFilter === "all" || r.status === requestStatusFilter) as request}
                        <tr>
                          <td>{request.requester?.name || "Unknown User"}</td>
                          <td class="description-cell">{request.reason}</td>
                          <td>{formatDate(request.created_at)}</td>
                          <td>
                            <span class="status-badge {request.status}">
                              {request.status}
                            </span>
                          </td>
                          <td>
                            <div class="action-buttons">
                              {#if request.status === "pending"}
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
            {#if selectedRequestType === "all" || selectedRequestType === "report"}
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
                      {#each reportRequests.filter(r => requestStatusFilter === "all" || r.status === requestStatusFilter) as request}
                        <tr>
                          <td>{request.reporter?.name || "Unknown User"}</td>
                          <td>{request.reported_user?.name || "Unknown User"}</td>
                          <td class="description-cell">{request.reason}</td>
                          <td>{formatDate(request.created_at)}</td>
                          <td>
                            <span class="status-badge {request.status}">
                              {request.status}
                            </span>
                          </td>
                          <td>
                            <div class="action-buttons">
                              {#if request.status === "pending"}
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

        {:else if activeTab === "categories"}
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
                        <td class="description-cell">{category.description || "-"}</td>
                        <td>{formatDate(category.created_at)}</td>
                        <td>
                          <div class="action-buttons">
                            <Button
                              variant="secondary"
                              size="small"
                              on:click={() => {editingCategory = category; categoryType = "thread";}}
                            >
                              Edit
                            </Button>
                            <Button
                              variant="danger"
                              size="small"
                              on:click={() => handleDeleteCategory(category.id, "thread")}
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
                        <td class="description-cell">{category.description || "-"}</td>
                        <td>{formatDate(category.created_at)}</td>
                        <td>
                          <div class="action-buttons">
                            <Button
                              variant="secondary"
                              size="small"
                              on:click={() => {editingCategory = category; categoryType = "community";}}
                            >
                              Edit
                            </Button>
                            <Button
                              variant="danger"
                              size="small"
                              on:click={() => handleDeleteCategory(category.id, "community")}
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

        {:else if activeTab === "newsletter"}
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

            <!-- Newsletter Subscribers -->
            <div class="subscribers-section">
              <div class="section-header">
                <h3>Newsletter Subscribers</h3>
              </div>

              {#if isLoadingSubscribers}
                <div class="loading-container">
                  <Spinner size="medium" />
                </div>
              {:else if newsletterSubscribers.length === 0}
                <div class="no-results">
                  <p>No users have subscribed to the newsletter.</p>
                </div>
              {:else}
                <div class="subscribers-table">
                  <table>
                    <thead>
                      <tr>
                        <th>User</th>
                        <th>Email</th>
                        <th>Joined</th>
                        <th>Status</th>
                      </tr>
                    </thead>
                    <tbody>
                      {#each newsletterSubscribers as user}
                        <tr>
                          <td class="user-cell">
                            <div class="user-info">
                              <div class="avatar">
                                {#if user.profile_picture_url}
                                  <img src={user.profile_picture_url} alt={user.name} />
                                {:else}
                                  <div class="avatar-placeholder">
                                    {user.name?.charAt(0) || user.username?.charAt(0) || "?"}
                                  </div>
                                {/if}
                              </div>
                              <div class="user-details">
                                <span class="name">{user.name || "Unknown"}</span>
                                <span class="username">@{user.username || "unknown"}</span>
                              </div>
                            </div>
                          </td>
                          <td>{user.email || "No email"}</td>
                          <td>{formatDate(user.created_at)}</td>
                          <td>
                            {#if user.is_banned}
                              <span class="status-badge banned">Banned</span>
                            {:else}
                              <span class="status-badge active">Active</span>
                            {/if}
                          </td>
                        </tr>
                      {/each}
                    </tbody>
                  </table>

                  {#if subscribersPagination && subscribersPagination.total_pages > 1}
                    <div class="pagination">
                      <button
                        class="pagination-btn"
                        disabled={subscribersPage <= 1}
                        on:click={handlePrevSubscribersPage}
                      >
                        Previous
                      </button>
                      <span class="page-info">Page {subscribersPage} of {subscribersPagination.total_pages}</span>
                      <button
                        class="pagination-btn"
                        disabled={subscribersPage >= subscribersPagination.total_pages}
                        on:click={handleNextSubscribersPage}
                      >
                        Next
                      </button>
                    </div>
                  {/if}
                </div>
              {/if}
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
          <label class="form-label">Category Type</label>
          <select
            bind:value={categoryType}
            class="theme-input"
          >
            <option value="thread">Thread Category</option>
            <option value="community">Community Category</option>
          </select>
        </div>

        <div class="form-group">
          <label for="category-name" class="form-label">Name</label>
          <input
            id="category-name"
            type="text"
            bind:value={newCategoryName}
            placeholder="Enter category name..."
            class="theme-input"
          />
        </div>

        <div class="form-group">
          <label for="category-description" class="form-label">Description (Optional)</label>
          <textarea
            id="category-description"
            bind:value={newCategoryDescription}
            placeholder="Enter category description..."
            rows="3"
            class="theme-input"
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
          <label for="edit-category-name" class="form-label">Name</label>
          <input
            id="edit-category-name"
            type="text"
            bind:value={editingCategory.name}
            placeholder="Enter category name..."
            class="theme-input"
          />
        </div>

        <div class="form-group">
          <label for="edit-category-description" class="form-label">Description</label>
          <textarea
            id="edit-category-description"
            bind:value={editingCategory.description}
            placeholder="Enter category description..."
            rows="3"
            class="theme-input"
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

            const category = editingCategory;
            if (category) {
              handleUpdateCategory(category.id, category.name, category.description, categoryType);
            }
          }}
        >
          Update Category
        </Button>
      </div>
    </div>
  </div>
{/if}