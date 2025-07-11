<script lang="ts">
  import { createEventDispatcher, tick, onMount } from "svelte";
  import {
    ImageIcon,
    BarChart2Icon,
    SmileIcon,
    MapPinIcon,
    XIcon,
    AlertCircleIcon,
    UsersIcon
  } from "svelte-feather-icons";
  import { createThread, uploadThreadMedia, replyToThread } from "../../api/thread";
  import { suggestThreadCategory } from "../../api/thread";
  import { getCommunityCategories, getUserCommunities, getJoinedCommunities } from "../../api/community";
  import { getThreadCategories } from "../../api/categories";
  import { toastStore } from "../../stores/toastStore";
  import { useTheme } from "../../hooks/useTheme";
  import { debounce } from "../../utils/helpers";
  import { getAuthToken, getUserRole, getUserId } from "../../utils/auth";
  import { uploadMultipleThreadMedia } from "../../utils/supabase";
  import appConfig from "../../config/appConfig";
  import type { ICategory } from "../../interfaces/ICategory";
  import type { ITweet } from "../../interfaces/ISocialMedia";

  export let isOpen = false;
  export let avatar = "";
  export let replyTo: ITweet | null = null;

  const { theme } = useTheme();
  $: isDarkMode = $theme === "dark";

  let newTweet = "";
  let files: File[] = [];
  let fileInputRef: HTMLInputElement;
  let isPosting = false;
  let errorMessage = "";
  let previewImages: string[] = [];
  const maxChars = 280;

  let suggestedCategory = "";
  let suggestedCategoryConfidence = 0;
  let isSuggestingCategory = false;
  let categoryTouched = false; 
  let selectedCategory = "";
  let allCategories: Record<string, number> = {};
  const categorySuggestionDebounceTimeout: ReturnType<typeof setTimeout> | null = null;

  let replyPermission = "Everyone";
  let showScheduleOptions = false;
  let scheduledDate = "";
  let scheduledTime = "";
  let showCommunityOptions = false;
  let availableCommunities: Array<{id: string, name: string, description?: string, logo_url?: string}> = [];
  let userCommunities: Array<{id: string, name: string, description?: string, logo_url?: string}> = [];
  let selectedCommunityId = "";
  let isLoadingCommunities = false;
  let showPollOptions = false;
  let pollQuestion = "";
  let pollOptions = ["", ""];
  let pollExpiryHours = 24;
  let pollWhoCanVote = "everyone";
  let isAdmin = false;
  let isAdvertisement = false;

  const categoryOptions = [
    { value: "technology", label: "Technology", icon: "laptop" },
    { value: "entertainment", label: "Entertainment", icon: "film" },
    { value: "health", label: "Health", icon: "heart" },
    { value: "sports", label: "Sports", icon: "activity" },
    { value: "business", label: "Business", icon: "briefcase" },
    { value: "politics", label: "Politics", icon: "flag" },
    { value: "education", label: "Education", icon: "book" },
    { value: "gaming", label: "Gaming", icon: "controller" },
    { value: "food", label: "Food", icon: "coffee" },
    { value: "travel", label: "Travel", icon: "map" },
    { value: "general", label: "General", icon: "hash" }
  ];

  $: charsRemaining = maxChars - newTweet.length;
  $: isOverLimit = charsRemaining < 0;
  $: isNearLimit = charsRemaining <= 20 && charsRemaining > 0;
  $: wordCount = newTweet.trim().split(/\s+/).filter(Boolean).length;
  $: wordPercent = Math.min(100, Math.round((wordCount / 280) * 100));

  const dispatch = createEventDispatcher();

  async function checkUserRole() {
    try {
      const role = await getUserRole();
      isAdmin = role === "admin";
    } catch (error) {
      console.error("Failed to check user role", error);
      isAdmin = false;
    }
  }

  async function loadCommunities() {
    try {
      isLoadingCommunities = true;

      const categories = await getCommunityCategories();
      availableCommunities = categories;

      const currentUserId = getUserId();
      if (!currentUserId) {
        console.error("Failed to get current user ID");
        isLoadingCommunities = false;
        return;
      }

      const userCommunitiesResponse = await getJoinedCommunities(currentUserId);
      if (userCommunitiesResponse && userCommunitiesResponse.communities) {
        userCommunities = userCommunitiesResponse.communities.map(community => ({
          id: community.id,
          name: community.name,
          description: community.description,
          logo_url: community.logo_url || community.avatar
        }));
      }

      console.log("Loaded user communities:", userCommunities);
      console.log("Loaded all communities:", availableCommunities);
      isLoadingCommunities = false;
    } catch (error) {
      availableCommunities = [];
      userCommunities = [];
      console.error("Failed to load communities", error);
      isLoadingCommunities = false;
    }
  }

  function extractHashtags(content: string): string[] {
    if (!content) return [];

    const hashtagRegex = /#([a-zA-Z0-9_]+)/g;
    const hashtags: string[] = [];
    let match;

    while ((match = hashtagRegex.exec(content)) !== null) {
      if (match[1]) {
        hashtags.push(match[1]);
      }
    }

    return hashtags;
  }

  function handleClose() {
    resetForm();
    dispatch("close");
  }

  function resetForm() {
    newTweet = "";
    files = [];
    previewImages = [];
    isPosting = false;
    errorMessage = "";
    suggestedCategory = "";
    suggestedCategoryConfidence = 0;
    categoryTouched = false;
    selectedCategory = "";
    allCategories = {};
    showScheduleOptions = false;
    scheduledDate = "";
    scheduledTime = "";
    showCommunityOptions = false;
    selectedCommunityId = "";
    showPollOptions = false;
    pollQuestion = "";
    pollOptions = ["", ""];
    pollExpiryHours = 24;
    pollWhoCanVote = "everyone";
    replyPermission = "Everyone";
    isAdvertisement = false;
  }

  const getSuggestedCategory = debounce(async (content: string) => {

    if (categoryTouched) return;

    if (!content || content.trim().length < 10) {
      suggestedCategory = "";
      suggestedCategoryConfidence = 0;
      return;
    }

    try {
      isSuggestingCategory = true;

      const result = await suggestThreadCategory(content).catch(() => ({ category: "", confidence: 0 }));

      if (result) {
        suggestedCategory = result.category || "";
        suggestedCategoryConfidence = result.confidence || 0;

        if (suggestedCategory && suggestedCategoryConfidence > 0.7 && !categoryTouched) {
          selectedCategory = suggestedCategory;
        }
      }
    } catch (error) {
      console.error("Error getting category suggestion:", error);

      suggestedCategory = "";
      suggestedCategoryConfidence = 0;
    } finally {
      isSuggestingCategory = false;
    }
  }, 500);

  $: if (newTweet) {
    getSuggestedCategory(newTweet);
  }

  function handleCategorySelect(category: string) {
    selectedCategory = category;
    categoryTouched = true;
  }

  function handleFileSelect(e: Event) {
    const input = e.target as HTMLInputElement;
    if (input.files) {

      const fileArray = Array.from(input.files);

      const newFiles = fileArray.slice(0, 4 - files.length);

      if (files.length + newFiles.length > 4) {
        toastStore.showToast("Maximum 4 media files allowed", "warning");
      }

      files = [...files, ...newFiles];

      newFiles.forEach(file => {
        const reader = new FileReader();
        reader.onload = (e) => {
          const result = e.target?.result as string;
          previewImages = [...previewImages, result];
        };
        reader.readAsDataURL(file);
      });
    }
  }

  function removeImage(index: number) {
    files = files.filter((_, i) => i !== index);
    previewImages = previewImages.filter((_, i) => i !== index);
  }

  function toggleScheduleOptions() {
    showScheduleOptions = !showScheduleOptions;
    if (showScheduleOptions) {

      const now = new Date();
      const tomorrow = new Date(now.getTime() + 24 * 60 * 60 * 1000);
      scheduledDate = tomorrow.toISOString().split("T")[0];

      const hours = now.getHours().toString().padStart(2, "0");
      const minutes = now.getMinutes().toString().padStart(2, "0");
      scheduledTime = `${hours}:${minutes}`;
    }
  }

  function toggleCommunityOptions() {
    showCommunityOptions = !showCommunityOptions;
  }

  function togglePollOptions() {
    showPollOptions = !showPollOptions;
    if (!showPollOptions) {

      pollQuestion = "";
      pollOptions = ["", ""];
      pollExpiryHours = 24;
      pollWhoCanVote = "everyone";
    }
  }

  function addPollOption() {
    if (pollOptions.length < 4) {
      pollOptions = [...pollOptions, ""];
    }
  }

  function removePollOption(index: number) {
    if (pollOptions.length > 2) {
      pollOptions = pollOptions.filter((_, i) => i !== index);
    }
  }

  function toggleAdvertisement() {
    if (isAdmin) {
      isAdvertisement = !isAdvertisement;
    }
  }

  async function handlePost() {
    if (isPosting) return;

    try {
      isPosting = true;
      errorMessage = "";

      if (newTweet.trim().length === 0 && files.length === 0) {
        errorMessage = "Please enter some content or attach media to post.";
        return;
      }

      if (isOverLimit) {
        errorMessage = "Tweet exceeds maximum character limit.";
        return;
      }

      if (replyTo && !replyTo.id) {
        errorMessage = "Missing parent thread ID.";
        return;
      }

      const hashtags = extractHashtags(newTweet);
      console.log("Extracted hashtags:", hashtags);

      const threadData: Record<string, any> = {
        content: newTweet.trim(),
        who_can_reply: replyPermission.toLowerCase(),
        hashtags: hashtags
      };

      if (selectedCategory) {
        threadData.category = selectedCategory;
      }

      if (showScheduleOptions && scheduledDate && scheduledTime) {
        const scheduledDateTime = new Date(`${scheduledDate}T${scheduledTime}`);
        threadData.scheduled_at = scheduledDateTime.toISOString();
      }

      if (showCommunityOptions && selectedCommunityId) {
        threadData.community_id = selectedCommunityId;
      }

      if (showPollOptions) {

        if (!pollQuestion || pollQuestion.trim() === "") {
          errorMessage = "Poll question cannot be empty";
          isPosting = false;
          return;
        }

        const validOptions = pollOptions.filter(opt => opt.trim() !== "");
        if (validOptions.length < 2) {
          errorMessage = "Poll requires at least 2 valid options";
          isPosting = false;
          return;
        }

        const endTime = new Date();
        endTime.setHours(endTime.getHours() + Number(pollExpiryHours));

        threadData.poll = {
          question: pollQuestion.trim(),
          options: validOptions,
          end_time: endTime.toISOString(),
          is_anonymous: pollWhoCanVote === "anonymous"
        };
      }

      if (isAdmin && isAdvertisement) {
        threadData.is_advertisement = true;
      }

      console.log("Sending post data:", threadData);

      if (replyTo) {
        try {

          const authToken = getAuthToken();
          if (!authToken) {
            throw new Error("Authentication required to reply");
          }

          const replyHashtags = extractHashtags(newTweet);
          console.log("Extracted hashtags for reply:", replyHashtags);

          threadData.thread_id = replyTo.id;
          threadData.hashtags = replyHashtags;

          if (replyTo.parent_id) {
            threadData.parent_reply_id = replyTo.id;
          }

          console.log("Sending reply data:", JSON.stringify(threadData));

          const response = await fetch(`${appConfig.api.baseUrl}/threads/${replyTo.id}/replies`, {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              "Authorization": `Bearer ${authToken}`
            },
            body: JSON.stringify(threadData)
          });

          if (!response.ok) {
            const errorData = await response.json().catch(() => ({ message: `HTTP error ${response.status}` }));
            throw new Error(errorData.message || `Failed to post reply: ${response.status}`);
          }

          const data = await response.json();

          if (files.length > 0 && data && data.id) {
            try {

              const mediaUrls = await uploadMultipleThreadMedia(files, data.id);

              if (mediaUrls && mediaUrls.length > 0) {

                await uploadThreadMedia(data.id, files);
                console.log("Successfully uploaded media for reply:", data.id);
              }
            } catch (uploadError) {
              console.error("Error uploading media for reply:", uploadError);
              toastStore.showToast("Your reply was published but media upload failed", "warning");
            }
          }

          toastStore.showToast("Your reply was published successfully", "success");
          resetForm();
          dispatch("posted", data);
          dispatch("close");
        } catch (replyError) {
          console.error("Error posting reply:", replyError);
          errorMessage = "Failed to post your reply. Please try again.";

          if (replyError instanceof Error) {

            if (replyError.message.includes("CORS") || replyError.message.includes("cross-origin")) {
              errorMessage = "Browser security prevented the request. This might be a CORS issue.";
            } else if (replyError.message.includes("307") || replyError.message.includes("redirect")) {
              errorMessage = "Server redirected the request unexpectedly. Please try again later.";
            } else {
              errorMessage += " " + replyError.message;
            }
          }

          toastStore.showToast(errorMessage, "error");
        }
      }
      else { 
        try {
          console.log("Sending thread data:", JSON.stringify(threadData));

          const authToken = getAuthToken();
          if (!authToken) {
            throw new Error("Authentication required to post");
          }

          const response = await fetch(`${appConfig.api.baseUrl}/threads`, {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              "Authorization": `Bearer ${authToken}`
            },
            body: JSON.stringify(threadData)
          });

          if (!response.ok) {
            const errorData = await response.json().catch(() => ({ message: `HTTP error ${response.status}` }));
            throw new Error(errorData.message || `Failed to create thread: ${response.status}`);
          }

          const data = await response.json();

          if (files.length > 0 && data && data.id) {
            try {

              const mediaUrls = await uploadMultipleThreadMedia(files, data.id);

              if (mediaUrls && mediaUrls.length > 0) {

                await uploadThreadMedia(data.id, files);
                console.log("Successfully uploaded media for thread:", data.id);
              }
            } catch (uploadError) {
              console.error("Error uploading media:", uploadError);
              toastStore.showToast("Your post was published but media upload failed", "warning");
            }
          }

          toastStore.showToast("Your post was published successfully", "success");
          resetForm();
          dispatch("posted", data);
          dispatch("close");
        } catch (threadError) {
          console.error("Error posting new thread:", threadError);
          errorMessage = "Failed to publish your post. Please try again.";

          if (threadError instanceof Error) {

            if (threadError.message.includes("CORS") || threadError.message.includes("cross-origin")) {
              errorMessage = "Browser security prevented the request. This might be a CORS issue.";
            } else if (threadError.message.includes("307") || threadError.message.includes("redirect")) {
              errorMessage = "Server redirected the request unexpectedly. Please try again later.";
            } else if (threadError.message.includes("Network error")) {
              errorMessage = threadError.message;
            } else {
              errorMessage += " " + threadError.message;
            }
          }

          toastStore.showToast(errorMessage, "error");
        }
      }
    } catch (error) {
      console.error("Error posting thread:", error);
      errorMessage = "Failed to publish your post. Please try again.";
      if (error instanceof Error) {
        errorMessage += " " + error.message;
      }
      toastStore.showToast(errorMessage, "error");
    } finally {
      isPosting = false;
    }
  }

  onMount(() => {
    checkUserRole();
    loadCommunities();
  });
</script>

{#if isOpen}
  <div class="modal-overlay" on:click={handleClose}>
    <div
      class="modal-container {isDarkMode ? "modal-container-dark" : ""}"
      on:click|stopPropagation={() => {}}
    >
      <div class="modal-header {isDarkMode ? "modal-header-dark" : ""}">
        <button
          class="modal-close-button {isDarkMode ? "modal-close-button-dark" : ""}"
          on:click={handleClose}
          aria-label="Close"
        >
          <XIcon size="20" />
        </button>
        <span class="modal-title">{replyTo ? "Reply" : "Create a post"}</span>
      </div>

      <div class="compose-tweet-container">
        {#if replyTo}
          <div class="compose-tweet-reply-to">
            <div class="compose-tweet-reply-header">
              <span>Replying to</span>
              <a href={`/user/${replyTo.username}`} class="compose-tweet-reply-username">@{replyTo.username}</a>
            </div>
            <div class="compose-tweet-reply-content">
              <img src={replyTo.profile_picture_url || ""} alt={replyTo.username} class="compose-tweet-reply-avatar" />
              <div class="compose-tweet-reply-text">
                {replyTo.content || ""}
              </div>
            </div>
          </div>
        {/if}

        <div class="compose-tweet-header">
          <img src={avatar} alt="Your avatar" class="compose-tweet-avatar" />

          <div class="compose-tweet-input-area">
            <textarea
              class="compose-tweet-textarea"
              placeholder={replyTo ? "Post your reply" : "What's happening?"}
              bind:value={newTweet}
              autofocus
            ></textarea>

            <!-- Reply Permission Selector (if not replying) -->
            {#if !replyTo}
              <div class="reply-permission-container">
                <div class="reply-permission-selector">
                  <UsersIcon size="14" />
                  <select
                    bind:value={replyPermission}
                    class="reply-permission-select {isDarkMode ? "reply-permission-select-dark" : ""}"
                  >
                    <option value="Everyone">Everyone can reply</option>
                    <option value="Accounts You Follow">People you follow</option>
                    <option value="Verified Accounts">Only verified accounts</option>
                  </select>
                </div>
              </div>
            {/if}

            <!-- Community selector if showing community options -->
            {#if showCommunityOptions && !replyTo}
              <div class="compose-tweet-community-selection {isDarkMode ? "compose-tweet-community-selection-dark" : ""}">
                <h4>Post to Community</h4>
                {#if isLoadingCommunities}
                  <div class="community-loading">Loading your communities...</div>
                {:else if userCommunities.length === 0}
                  <div class="community-empty">
                    You're not a member of any communities yet.
                    <a href="/communities" class="community-link">Join a community</a>
                  </div>
                {:else}
                  <select
                    bind:value={selectedCommunityId}
                    class="compose-tweet-community-select {isDarkMode ? "compose-tweet-community-select-dark" : ""}"
                  >
                    <option value="">Select a community</option>
                    {#each userCommunities as community}
                      <option value={community.id}>{community.name}</option>
                    {/each}
                  </select>

                  {#if selectedCommunityId}
                    {#each userCommunities as community}
                      {#if community.id === selectedCommunityId}
                        <div class="selected-community-info">
                          <div class="selected-community-header">
                            {#if community.logo_url}
                              <img src={community.logo_url} alt={community.name} class="selected-community-logo" />
                            {:else}
                              <div class="selected-community-logo-placeholder">
                                {community.name.charAt(0).toUpperCase()}
                              </div>
                            {/if}
                            <span class="selected-community-name">{community.name}</span>
                          </div>
                          {#if community.description}
                            <p class="selected-community-description">{community.description}</p>
                          {/if}
                        </div>
                      {/if}
                    {/each}
                  {/if}
                {/if}
              </div>
            {/if}

            <!-- Schedule options -->
            {#if showScheduleOptions && !replyTo}
              <div class="compose-tweet-schedule {isDarkMode ? "compose-tweet-schedule-dark" : ""}">
                <h4>Schedule post</h4>
                <div class="compose-tweet-schedule-inputs">
                  <input
                    type="date"
                    bind:value={scheduledDate}
                    min={new Date().toISOString().split("T")[0]}
                    class="compose-tweet-schedule-date {isDarkMode ? "compose-tweet-schedule-date-dark" : ""}"
                  />
                  <input
                    type="time"
                    bind:value={scheduledTime}
                    class="compose-tweet-schedule-time {isDarkMode ? "compose-tweet-schedule-time-dark" : ""}"
                  />
                </div>
              </div>
            {/if}

            <!-- Poll options -->
            {#if showPollOptions}
              <div class="compose-tweet-poll-options {isDarkMode ? "compose-tweet-poll-options-dark" : ""}">
                <div class="compose-tweet-poll-question">
                  <input
                    type="text"
                    placeholder="Ask a question..."
                    bind:value={pollQuestion}
                    class="compose-tweet-poll-question-input {isDarkMode ? "compose-tweet-poll-question-input-dark" : ""}"
                  />
                </div>
                <div class="compose-tweet-poll-choices">
                  {#each pollOptions as option, index}
                    <div class="compose-tweet-poll-choice">
                      <input
                        type="text"
                        placeholder={`Option ${index + 1}`}
                        bind:value={pollOptions[index]}
                        class="compose-tweet-poll-choice-input {isDarkMode ? "compose-tweet-poll-choice-input-dark" : ""}"
                      />
                      {#if index > 1 && pollOptions.length > 2}
                        <button
                          class="compose-tweet-poll-choice-remove"
                          on:click={() => removePollOption(index)}
                        >×</button>
                      {/if}
                    </div>
                  {/each}
                </div>
                {#if pollOptions.length < 4}
                  <button
                    class="compose-tweet-poll-add-option {isDarkMode ? "compose-tweet-poll-add-option-dark" : ""}"
                    on:click={addPollOption}
                  >
                    + Add Option
                  </button>
                {/if}
                <div class="compose-tweet-poll-settings">
                  <div class="compose-tweet-poll-duration">
                    <label for="poll-duration">Poll duration:</label>
                    <select
                      id="poll-duration"
                      bind:value={pollExpiryHours}
                      class="compose-tweet-poll-select {isDarkMode ? "compose-tweet-poll-select-dark" : ""}"
                    >
                      <option value={1}>1 hour</option>
                      <option value={6}>6 hours</option>
                      <option value={12}>12 hours</option>
                      <option value={24}>24 hours</option>
                      <option value={72}>3 days</option>
                      <option value={168}>7 days</option>
                    </select>
                  </div>
                  <div class="compose-tweet-poll-who-can-vote">
                    <label for="poll-who-can-vote">Who can vote:</label>
                    <select
                      id="poll-who-can-vote"
                      bind:value={pollWhoCanVote}
                      class="compose-tweet-poll-select {isDarkMode ? "compose-tweet-poll-select-dark" : ""}"
                    >
                      <option value="everyone">Everyone</option>
                      <option value="following">Accounts you follow</option>
                      <option value="verified">Verified accounts</option>
                    </select>
                  </div>
                </div>
              </div>
            {/if}

            <!-- Category suggestion -->
            {#if newTweet && newTweet.length >= 10}
              <div class="category-suggestion-container">
                <div class="category-suggestion-header">
                  <span class="category-suggestion-title">
                    <span class="category-icon">#</span>
                    {isSuggestingCategory ? "Analyzing content..." : "Category"}
                  </span>

                  {#if suggestedCategory && !categoryTouched}
                    <span class="category-suggestion-info">
                      AI suggested: <strong>{suggestedCategory}</strong>
                      {#if suggestedCategoryConfidence > 0}
                        ({Math.round(suggestedCategoryConfidence * 100)}% confidence)
                      {/if}
                    </span>
                  {/if}
                </div>

                <div class="category-options">
                  {#each categoryOptions as option}
                    <button
                      class="category-option {selectedCategory === option.value ? "selected" : ""}
                             {!selectedCategory && suggestedCategory === option.value ? "suggested" : ""}"
                      on:click={() => handleCategorySelect(option.value)}
                    >
                      {option.label}
                    </button>
                  {/each}
                </div>
              </div>
            {/if}

            <!-- Admin advertisement option -->
            {#if isAdmin}
              <div class="compose-tweet-admin-options">
                <label class="compose-tweet-admin-option">
                  <input
                    type="checkbox"
                    bind:checked={isAdvertisement}
                  />
                  <span>Mark as advertisement</span>
                </label>
              </div>
            {/if}

            <!-- Media preview -->
            {#if previewImages.length > 0}
              <div class="compose-tweet-media-preview">
                <div class="compose-tweet-media-grid {previewImages.length === 1 ? "single" : ""}">
                  {#each previewImages as preview, i}
                    <div class="compose-tweet-media-item">
                      <img src={preview} alt="Preview" class="compose-tweet-media-img" />
                      <button
                        class="compose-tweet-media-remove"
                        on:click={() => removeImage(i)}
                        aria-label="Remove image"
                      >
                        <XIcon size="16" />
                      </button>
                    </div>
                  {/each}
                </div>
              </div>
            {/if}

            {#if errorMessage}
              <div class="compose-tweet-error">
                {errorMessage}
              </div>
            {/if}
          </div>
        </div>

        <div class="compose-tweet-actions">
          <div class="compose-tweet-tools">
            <label class="compose-tweet-tool" aria-label="Add images, GIFs or videos">
              <input
                type="file"
                class="compose-tweet-file-input"
                accept="image/*,video/*,.gif"
                multiple
                bind:this={fileInputRef}
                on:change={handleFileSelect}
                disabled={files.length >= 4 || isPosting}
              />
              <ImageIcon size="20" />
            </label>

            <button
              class="compose-tweet-tool"
              class:active-tool={showPollOptions}
              aria-label="Add poll"
              disabled={isPosting || files.length > 0}
              on:click={togglePollOptions}
            >
              <BarChart2Icon size="20" />
            </button>

            <button class="compose-tweet-tool" aria-label="Add emoji" disabled={isPosting}>
              <SmileIcon size="20" />
            </button>

            {#if !replyTo}
              <button
                class="compose-tweet-tool"
                class:active-tool={showScheduleOptions}
                aria-label="Schedule post"
                disabled={isPosting}
                on:click={toggleScheduleOptions}
              >
                <AlertCircleIcon size="20" />
              </button>

              <button
                class="compose-tweet-tool"
                class:active-tool={showCommunityOptions}
                aria-label="Post to community"
                disabled={isPosting}
                on:click={toggleCommunityOptions}
              >
                <UsersIcon size="20" />
              </button>
            {/if}
          </div>

          <div class="compose-tweet-submit-area">
            <!-- Word Counter with Circular Progress -->
            <div class="compose-tweet-word-count">
              <div class="compose-tweet-word-circle">
                <svg viewBox="0 0 36 36">
                  <path
                    d="M18 2a16 16 0 1 1 0 32 16 16 0 0 1 0-32"
                    fill="none"
                    stroke={isDarkMode ? "#374151" : "#e5e7eb"}
                    stroke-width="4"
                  />
                  <path
                    d="M18 2a16 16 0 1 1 0 32 16 16 0 0 1 0-32"
                    fill="none"
                    stroke={isNearLimit ? "#ef4444" : "#3b82f6"}
                    stroke-width="4"
                    stroke-dasharray="100, 100"
                    stroke-dashoffset={100 - wordPercent}
                    style="transition: stroke-dashoffset 0.2s ease;"
                  />
                </svg>
                <span class="compose-tweet-word-count-text {isNearLimit ? "near-limit" : ""}">{wordCount}</span>
              </div>
            </div>

            <button
              class="compose-tweet-submit"
              type="button"
              on:click={handlePost}
              disabled={isPosting || isOverLimit || (newTweet.trim() === "" && files.length === 0 && !showPollOptions)}
            >
              {isPosting ? "Publishing..." : "Post"}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
  .compose-tweet-container {
    padding: var(--space-3) var(--space-4);
  }

  .compose-tweet-header {
    display: flex;
    align-items: flex-start;
  }

  .compose-tweet-avatar {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    margin-right: var(--space-3);
    object-fit: cover;
  }

  .compose-tweet-input-area {
    flex: 1;
    min-width: 0;
  }

  .compose-tweet-textarea {
    width: 100%;
    min-height: 120px;
    padding: var(--space-3) 0;
    border: none;
    background-color: transparent;
    color: var(--text-primary);
    font-size: var(--font-size-lg);
    resize: none;
    overflow-y: auto;
  }

  .compose-tweet-textarea:focus {
    outline: none;
  }

  .compose-tweet-textarea::placeholder {
    color: var(--text-tertiary);
  }

  .compose-tweet-media-preview {
    margin-top: var(--space-3);
    border-radius: var(--radius-md);
    overflow: hidden;
  }

  .compose-tweet-media-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
    border-radius: var(--radius-md);
  }

  .compose-tweet-media-grid.single {
    grid-template-columns: 1fr;
  }

  .compose-tweet-media-item {
    position: relative;
    aspect-ratio: 16/9;
    background-color: var(--bg-tertiary);
    border-radius: var(--radius-md);
    overflow: hidden;
  }

  .compose-tweet-media-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .compose-tweet-media-remove {
    position: absolute;
    top: 8px;
    right: 8px;
    width: 24px;
    height: 24px;
    border-radius: 50%;
    background-color: rgba(0, 0, 0, 0.5);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    border: none;
  }

  .compose-tweet-actions {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: var(--space-3);
    padding-top: var(--space-3);
    border-top: 1px solid var(--border-color);
  }

  .compose-tweet-tools {
    display: flex;
    gap: var(--space-3);
  }

  .compose-tweet-tool {
    color: var(--color-primary);
    background: transparent;
    border: none;
    cursor: pointer;
    padding: var(--space-1);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background-color var(--transition-fast);
  }

  .compose-tweet-tool:hover {
    background-color: var(--hover-primary);
  }

  .compose-tweet-tool:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .compose-tweet-file-input {
    display: none;
  }

  .compose-tweet-submit-area {
    display: flex;
    align-items: center;
  }

  .compose-tweet-word-count {
    position: relative;
    display: flex;
    align-items: center;
    margin-right: 16px;
  }

  .compose-tweet-word-circle {
    position: relative;
    width: 32px;
    height: 32px;
  }

  .compose-tweet-word-count-text {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    font-size: 12px;
    font-weight: 500;
  }

  .compose-tweet-word-count-text.near-limit {
    color: #ef4444;
  }

  .active-tool {
    background-color: rgba(59, 130, 246, 0.1);
    color: #3b82f6;
  }

  .reply-permission-container {
    display: flex;
    margin: 8px 0;
  }

  .reply-permission-selector {
    display: flex;
    align-items: center;
    padding: 5px 10px;
    background-color: rgba(29, 155, 240, 0.1);
    color: #1d9bf0;
    border-radius: 16px;
    font-size: 13px;
    cursor: pointer;
  }

  .reply-permission-select {
    background: none;
    border: none;
    color: inherit;
    font-size: inherit;
    font-family: inherit;
    padding-left: 5px;
    cursor: pointer;
    outline: none;
    max-width: 200px;
  }

  .compose-tweet-reply-to {
    margin-bottom: 16px;
    padding: 16px;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    background-color: #f9fafb;
  }

  .compose-tweet-reply-to-dark {
    background-color: #1f2937;
    border-color: #384152;
  }

  .compose-tweet-reply-content {
    display: flex;
    gap: 12px;
  }

  .compose-tweet-reply-avatar-container {
    flex-shrink: 0;
  }

  .compose-tweet-reply-avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    object-fit: cover;
  }

  .compose-tweet-reply-info {
    flex: 1;
    overflow: hidden;
  }

  .compose-tweet-reply-author {
    margin-bottom: 4px;
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .compose-tweet-reply-name {
    font-weight: 600;
    font-size: 14px;
    color: #111827;
  }

  .compose-tweet-reply-to-dark .compose-tweet-reply-name {
    color: #f1f5f9;
  }

  .compose-tweet-reply-username {
    font-size: 14px;
    color: #6b7280;
  }

  .compose-tweet-reply-text {
    font-size: 14px;
    margin: 0;
    line-height: 1.5;
    overflow: hidden;
    display: -webkit-box;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 3; 
  }

  .compose-tweet-community-selection,
  .compose-tweet-schedule,
  .compose-tweet-poll-options {
    margin: 12px 0;
    padding: 12px;
    border-radius: 8px;
    background-color: #f8fafc;
    border: 1px solid #e5e7eb;
  }

  .compose-tweet-community-selection-dark,
  .compose-tweet-schedule-dark,
  .compose-tweet-poll-options-dark {
    background-color: #1e293b;
    border-color: #384152;
  }

  .compose-tweet-community-select,
  .compose-tweet-schedule-date,
  .compose-tweet-schedule-time,
  .compose-tweet-poll-select {
    width: 100%;
    padding: 8px;
    border-radius: 4px;
    border: 1px solid #e2e8f0;
    background: white;
    margin-top: 8px;
  }

  .compose-tweet-community-select-dark,
  .compose-tweet-schedule-date-dark,
  .compose-tweet-schedule-time-dark,
  .compose-tweet-poll-select-dark {
    border-color: #4b5563;
    background: #374151;
    color: #e5e7eb;
  }

  .compose-tweet-schedule-inputs {
    display: flex;
    gap: 8px;
  }

  .compose-tweet-poll-question,
  .compose-tweet-poll-choice {
    margin-bottom: 8px;
    position: relative;
  }

  .compose-tweet-poll-question-input,
  .compose-tweet-poll-choice-input {
    width: 100%;
    padding: 8px;
    border-radius: 4px;
    border: 1px solid #e2e8f0;
    background: white;
  }

  .compose-tweet-poll-question-input-dark,
  .compose-tweet-poll-choice-input-dark {
    border-color: #4b5563;
    background: #374151;
    color: #e5e7eb;
  }

  .compose-tweet-poll-choice-remove {
    position: absolute;
    right: 8px;
    top: 50%;
    transform: translateY(-50%);
    background: none;
    border: none;
    font-size: 18px;
    cursor: pointer;
    color: #ef4444;
  }

  .compose-tweet-poll-add-option {
    padding: 8px;
    background: none;
    border: 1px dashed #e2e8f0;
    border-radius: 4px;
    width: 100%;
    cursor: pointer;
    color: #3b82f6;
    margin-bottom: 12px;
  }

  .compose-tweet-poll-add-option-dark {
    border-color: #4b5563;
    color: #60a5fa;
  }

  .compose-tweet-poll-settings {
    display: flex;
    flex-wrap: wrap;
    gap: 16px;
  }

  .compose-tweet-poll-duration,
  .compose-tweet-poll-who-can-vote {
    flex: 1;
    min-width: 150px;
  }

  .compose-tweet-admin-options {
    margin: 12px 0;
    padding: 12px;
    border-radius: 8px;
    background-color: #fee2e2;
    border: 1px solid #fca5a5;
  }

  .compose-tweet-admin-option {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
  }

  .modal-container {
    position: relative;
    width: 600px;
    max-width: 95vw;
    max-height: 90vh;
    background-color: white;
    border-radius: 16px;
    box-shadow: 0 8px 30px rgba(0, 0, 0, 0.12);
    overflow-y: auto;
    z-index: 1001;
  }

  .modal-container-dark {
    background-color: #1e293b;
    color: #f1f5f9;
  }

  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(2px);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000;
  }

  .modal-header {
    display: flex;
    align-items: center;
    padding: 12px 16px;
    border-bottom: 1px solid #e5e7eb;
    position: sticky;
    top: 0;
    background-color: white;
    z-index: 10;
  }

  .modal-header-dark {
    background-color: #1e293b;
    border-bottom: 1px solid #384152;
  }

  .modal-close-button {
    background: none;
    border: none;
    cursor: pointer;
    color: #374151;
    padding: 8px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .modal-header-dark .modal-close-button {
    color: #e5e7eb;
  }

  .modal-title {
    margin-left: 16px;
    font-size: 18px;
    font-weight: 700;
  }

  .category-suggestion-container {
    margin: 10px 0;
    padding: 10px;
    border-radius: 8px;
    background-color: rgba(0, 0, 0, 0.02);
    border: 1px solid rgba(0, 0, 0, 0.05);
  }

  .category-suggestion-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
  }

  .category-suggestion-title {
    display: flex;
    align-items: center;
    gap: 5px;
    font-weight: 600;
    font-size: 14px;
  }

  .category-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 16px;
    font-weight: bold;
  }

  .category-suggestion-info {
    font-size: 12px;
    color: #555;
  }

  .category-options {
    display: flex;
    flex-wrap: wrap;
    gap: 5px;
  }

  .category-option {
    padding: 6px 12px;
    border-radius: 20px;
    font-size: 12px;
    background-color: #f1f1f1;
    border: 1px solid transparent;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .category-option:hover {
    background-color: #e0e0e0;
  }

  .category-option.selected {
    background-color: #1d9bf0;
    color: white;
  }

  .category-option.suggested {
    border: 1px dashed #1d9bf0;
    animation: pulse 2s infinite;
  }

  @keyframes pulse {
    0% {
      border-color: rgba(29, 155, 240, 0.5);
    }
    50% {
      border-color: rgba(29, 155, 240, 1);
    }
    100% {
      border-color: rgba(29, 155, 240, 0.5);
    }
  }

  :global(.dark) .category-suggestion-container {
    background-color: rgba(255, 255, 255, 0.05);
    border-color: rgba(255, 255, 255, 0.1);
  }

  :global(.dark) .category-suggestion-info {
    color: #aaa;
  }

  :global(.dark) .category-option {
    background-color: #2f3336;
    color: #e0e0e0;
  }

  :global(.dark) .category-option:hover {
    background-color: #3f4246;
  }

  .community-loading, .community-empty {
    padding: var(--space-2);
    margin: var(--space-2) 0;
    text-align: center;
    color: var(--text-secondary);
  }

  .community-link {
    color: var(--color-primary);
    text-decoration: none;
  }

  .community-link:hover {
    text-decoration: underline;
  }

  .selected-community-info {
    margin-top: var(--space-2);
    padding: var(--space-2);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    background-color: var(--bg-tertiary);
  }

  .selected-community-header {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-bottom: var(--space-1);
  }

  .selected-community-logo {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    object-fit: cover;
  }

  .selected-community-logo-placeholder {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    background-color: var(--color-primary);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: var(--font-weight-bold);
    font-size: var(--font-size-sm);
  }

  .selected-community-name {
    font-weight: var(--font-weight-bold);
  }

  .selected-community-description {
    font-size: var(--font-size-sm);
    margin: 0;
    color: var(--text-secondary);
  }
</style>