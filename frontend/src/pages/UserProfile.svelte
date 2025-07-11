<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "../stores/routeStore";
  import OwnProfile from "./OwnProfile.svelte";
  import OtherProfile from "./OtherProfile.svelte";
  import LoadingSkeleton from "../components/common/LoadingSkeleton.svelte";
  import MainLayout from "../components/layout/MainLayout.svelte";
  import { toastStore } from "../stores/toastStore";
  import { getUserById, getUserByUsername } from "../api/user";
  import { getUserId, isAuthenticated } from "../utils/auth";
  import { createLoggerWithPrefix } from "../utils/logger";

  const logger = createLoggerWithPrefix("UserProfile");

  export let userId: string = "";

  let isLoading = true;
  let error: string | null = null;
  let username = "";
  let name = "";
  let profile_picture_url = "";
  let isOwnProfile = false;

  const unsubscribe = page.subscribe(($page) => {
    logger.debug("Page store updated:", $page);

    if ($page.params.userId) {
      if (userId !== $page.params.userId) {
        logger.debug(`User ID changed from ${userId} to ${$page.params.userId}`);
        userId = $page.params.userId;

        const currentUserId = getUserId();
        isOwnProfile = userId === "me" || userId === currentUserId;
        logger.debug(`Is own profile: ${isOwnProfile}, currentUserId: ${currentUserId}`);

        loadUserBasicInfo(userId);
      }
    } else if (!userId) {

      parseUserIdFromUrl();
    }
  });

  function parseUserIdFromUrl() {
    const pathParts = window.location.pathname.split("/");
    const userIndex = pathParts.indexOf("user");

    if (userIndex >= 0 && userIndex + 1 < pathParts.length) {
      const urlUserId = pathParts[userIndex + 1];

      if (userId !== urlUserId) {
        logger.debug(`Parsed user ID from URL: ${urlUserId}`);
        userId = urlUserId;

        const currentUserId = getUserId();
        isOwnProfile = userId === "me" || userId === currentUserId;
        logger.debug(`Is own profile: ${isOwnProfile}, currentUserId: ${currentUserId}`);

        loadUserBasicInfo(userId);
      }
    } else {
      logger.error("Failed to parse user ID from URL");
      error = "Invalid user ID";
      isLoading = false;
    }
  }

  async function loadUserBasicInfo(id: string) {
    if (!id) {
      logger.error("Invalid user ID");
      error = "Invalid user ID";
      isLoading = false;
      return;
    }

    if (!isAuthenticated()) {
      logger.warn("User not authenticated, redirecting to login");
      window.location.href = "/login";
      return;
    }

    logger.debug(`Loading user info for ID: ${id}`);
    isLoading = true;
    error = null;

    try {

      const isUUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(id);

      let userData;
      if (isUUID) {

        userData = await getUserById(id);
      } else {

        userData = await getUserByUsername(id);
      }

      logger.debug("User data loaded:", userData);

      if (userData && userData.user) {
        username = userData.user.username || "";
        name = userData.user.name || "";
        profile_picture_url = userData.user.profile_picture_url || "";
        logger.debug(`User info: ${name} (@${username})`);
      } else {
        logger.error("User not found");
        error = "User not found";
      }
    } catch (err) {
      logger.error("Error loading user:", err);
      error = "Failed to load user profile";
      toastStore.showToast("Failed to load user profile. Please try again.", "error");
    } finally {
      isLoading = false;
    }
  }

  onMount(() => {
    logger.debug("UserProfile component mounted");

    if (!userId) {
      logger.debug("No userId from page store or props, parsing from URL");
      parseUserIdFromUrl();
    } else {
      logger.debug(`Using provided userId: ${userId}`);

      const currentUserId = getUserId();
      isOwnProfile = userId === "me" || userId === currentUserId;
      logger.debug(`Is own profile: ${isOwnProfile}, currentUserId: ${currentUserId}`);

      loadUserBasicInfo(userId);
    }

    const handlePopState = () => {
      logger.debug("PopState event triggered, parsing URL");
      parseUserIdFromUrl();
    };

    window.addEventListener("popstate", handlePopState);

    return () => {
      logger.debug("Cleaning up UserProfile component");
      unsubscribe();
      window.removeEventListener("popstate", handlePopState);
    };
  });
</script>

{#if isLoading}
  <MainLayout>
    <LoadingSkeleton type="profile" />
  </MainLayout>
{:else if error}
  <MainLayout>
    <div class="flex flex-col items-center justify-center p-8">
      <h2 class="text-xl font-bold text-gray-700 dark:text-gray-300 mb-4">{error}</h2>
      <a href="/" class="text-blue-500 hover:underline">Return to home</a>
    </div>
  </MainLayout>
{:else if isOwnProfile}
  <OwnProfile {userId} />
{:else}
  <OtherProfile {userId} />
{/if}