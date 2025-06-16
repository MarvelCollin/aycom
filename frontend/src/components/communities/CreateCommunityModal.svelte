<script lang="ts">
  import { createEventDispatcher, onMount } from "svelte";
  import { useTheme } from "../../hooks/useTheme";
  import { toastStore } from "../../stores/toastStore";
  import { createLoggerWithPrefix } from "../../utils/logger";
  import { createCommunity, getCategories } from "../../api/community";

  import Spinner from "../common/Spinner.svelte";

  import XIcon from "svelte-feather-icons/src/icons/XIcon.svelte";
  import ImageIcon from "svelte-feather-icons/src/icons/ImageIcon.svelte";
  import AlertCircleIcon from "svelte-feather-icons/src/icons/AlertCircleIcon.svelte";
  import CheckIcon from "svelte-feather-icons/src/icons/CheckIcon.svelte";

  const logger = createLoggerWithPrefix("CreateCommunityModal");
  const dispatch = createEventDispatcher();
  const { theme } = useTheme();

  export let isOpen = false;

  $: isDarkMode = $theme === "dark";

  let isLoading = false;
  let isSubmitting = false;
  let isSuccess = false;

  let communityName = "";
  let description = "";
  let icon: File | null = null;
  let iconPreview: string | null = null;
  let banner: File | null = null;
  let bannerPreview: string | null = null;
  let rules = "";
  let availableCategories: string[] = [];
  let selectedCategories: string[] = [];
  let errors: Record<string, string> = {};

  const defaultCategories = [
    "Art", "Business", "Education", "Entertainment", "Gaming",
    "Health", "Lifestyle", "Music", "News", "Politics",
    "Science", "Sports", "Technology", "Travel"
  ];

  const categoryTemplates = [
    {
      name: "Technology Community",
      categories: ["Technology", "Programming", "Science"]
    },
    {
      name: "Gaming Community",
      categories: ["Gaming", "Entertainment", "Technology"]
    },
    {
      name: "Creative Community",
      categories: ["Art", "Music", "Entertainment"]
    },
    {
      name: "Business Network",
      categories: ["Business", "News", "Education"]
    },
    {
      name: "Lifestyle Group",
      categories: ["Lifestyle", "Health", "Travel"]
    }
  ];

  onMount(async () => {
    await fetchCategories();
  });

  async function fetchCategories() {
    try {
      isLoading = true;
      const response = await getCategories();

      if (Array.isArray(response) && response.length > 0) {
        availableCategories = response.map(cat => cat.name);
      } else if (response && typeof response === "object" && "categories" in response) {
        const typedResponse = response as { categories: Array<{name: string}> };
        if (typedResponse.categories && typedResponse.categories.length > 0) {
          availableCategories = typedResponse.categories.map(cat => cat.name);
        } else {
          availableCategories = defaultCategories;
        }
      } else {
        availableCategories = defaultCategories;
      }
    } catch (error) {
      logger.error("Error fetching categories:", error);
      availableCategories = defaultCategories;
    } finally {
      isLoading = false;
    }
  }

  function validateForm(): boolean {
    errors = {};

    if (!communityName.trim()) {
      errors.communityName = "Community name is required";
    }

    if (!description.trim()) {
      errors.description = "Description is required";
    }

    if (!icon) {
      errors.icon = "Community icon is required";
    }

    if (selectedCategories.length === 0) {
      errors.categories = "At least one category is required";
    }

    if (!banner) {
      errors.banner = "Community banner is required";
    }

    if (!rules.trim()) {
      errors.rules = "Community rules are required";
    }

    return Object.keys(errors).length === 0;
  }

  function handleIconChange(event: Event) {
    const input = event.target as HTMLInputElement;
    if (!input.files || input.files.length === 0) {
      return;
    }

    icon = input.files[0];
    const reader = new FileReader();
    reader.onload = e => {
      iconPreview = e.target?.result as string;
    };
    reader.readAsDataURL(icon);
  }

  function handleBannerChange(event: Event) {
    const input = event.target as HTMLInputElement;
    if (!input.files || input.files.length === 0) {
      return;
    }

    banner = input.files[0];
    const reader = new FileReader();
    reader.onload = e => {
      bannerPreview = e.target?.result as string;
    };
    reader.readAsDataURL(banner);
  }

  function toggleCategory(category: string) {
    if (selectedCategories.includes(category)) {
      selectedCategories = selectedCategories.filter(c => c !== category);
    } else {
      if (selectedCategories.length >= 5) {
        toastStore.showToast("Maximum 5 categories allowed", "warning");
        return;
      }
      selectedCategories = [...selectedCategories, category];
    }
  }

  function applyTemplate(template: { name: string, categories: string[] }) {

    const validCategories = template.categories.filter(cat =>
      availableCategories.includes(cat)
    );

    const limitedCategories = validCategories.slice(0, 5);

    if (limitedCategories.length === 0) {
      toastStore.showToast("No valid categories in this template", "warning");
      return;
    }

    selectedCategories = [...limitedCategories];
    toastStore.showToast(`Applied "${template.name}" template`, "success");
  }

  async function handleSubmit() {
    if (!validateForm()) {

      const firstErrorKey = Object.keys(errors)[0];
      const errorElement = document.querySelector(`[data-error="${firstErrorKey}"]`);
      if (errorElement) {
        errorElement.scrollIntoView({ behavior: "smooth", block: "center" });
      }
      return;
    }

    isSubmitting = true;

    try {

      const communityData = {
        name: communityName,
        description: description,
        icon: icon,
        banner: banner,
        categories: selectedCategories,
        rules: rules
      };

      const result = await createCommunity(communityData);

      isSuccess = true;
      toastStore.showToast("Community creation request submitted for approval", "success");

      setTimeout(() => {
        handleClose();
        dispatch("success");
      }, 2000);

    } catch (error) {
      logger.error("Error creating community:", error);
      toastStore.showToast("Failed to create community. Please try again.", "error");
    } finally {
      isSubmitting = false;
    }
  }

  function handleClose() {
    isOpen = false;
    dispatch("close");

    setTimeout(() => {
      communityName = "";
      description = "";
      icon = null;
      iconPreview = null;
      banner = null;
      bannerPreview = null;
      rules = "";
      selectedCategories = [];
      errors = {};
      isSuccess = false;
    }, 300);
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === "Escape") {
      handleClose();
    }
  }
</script>

{#if isOpen}
  <div class="modal-overlay {isDarkMode ? "dark" : ""}"
    on:click={handleClose}
    on:keydown={handleKeyDown}
    role="dialog"
    aria-modal="true"
    tabindex="-1">

    <div class="modal-container" on:click|stopPropagation>
      <div class="modal-header">
        <h2>Create Community</h2>
        <button class="close-button" on:click={handleClose} aria-label="Close modal">
          <XIcon size="20" />
        </button>
      </div>

      <div class="modal-body">
        {#if isLoading}
          <div class="loading-container">
            <Spinner size="large" />
          </div>
        {:else if isSuccess}
          <div class="success-container">
            <div class="success-icon">✓</div>
            <h3>Request Submitted!</h3>
            <p>Your community creation request has been submitted and is pending admin approval.</p>
          </div>
        {:else}
          <p class="modal-description">
            Create your own community to connect with people who share your interests.
            All communities are subject to admin approval before they are created.
          </p>

          <form class="community-form" on:submit|preventDefault={handleSubmit}>
            <div class="form-group" data-error="communityName">
              <label for="communityName">
                Community Name <span class="required">*</span>
              </label>
              <input
                id="communityName"
                type="text"
                class={errors.communityName ? "error" : ""}
                bind:value={communityName}
                placeholder="Enter community name"
                maxlength="50"
              />
              <div class="input-hint">
                <span>{communityName.length}/50 characters</span>
              </div>
              {#if errors.communityName}
                <div class="error-message">
                  <AlertCircleIcon size="14" />
                  <span>{errors.communityName}</span>
                </div>
              {/if}
            </div>

            <div class="form-group" data-error="description">
              <label for="description">
                Description <span class="required">*</span>
              </label>
              <textarea
                id="description"
                class={errors.description ? "error" : ""}
                bind:value={description}
                placeholder="Describe what your community is about"
                rows="4"
                maxlength="500"
              ></textarea>
              <div class="input-hint">
                <span>{description.length}/500 characters (minimum 30)</span>
              </div>
              {#if errors.description}
                <div class="error-message">
                  <AlertCircleIcon size="14" />
                  <span>{errors.description}</span>
                </div>
              {/if}
            </div>

            <div class="form-row">
              <div class="form-group media-upload" data-error="icon">
                <label>
                  Community Icon <span class="required">*</span>
                </label>
                <div class="media-preview {errors.icon ? "error" : ""}">
                  {#if iconPreview}
                    <img src={iconPreview} alt="Community icon preview" />
                  {:else}
                    <div class="upload-placeholder">
                      <ImageIcon size="24" />
                      <span>Upload Icon</span>
                    </div>
                  {/if}
                  <input
                    type="file"
                    accept="image/*"
                    on:change={handleIconChange}
                  />
                </div>
                {#if errors.icon}
                  <div class="error-message">
                    <AlertCircleIcon size="14" />
                    <span>{errors.icon}</span>
                  </div>
                {/if}
              </div>

              <div class="form-group media-upload" data-error="banner">
                <label>
                  Community Banner <span class="required">*</span>
                </label>
                <div class="media-preview banner-preview {errors.banner ? "error" : ""}">
                  {#if bannerPreview}
                    <img src={bannerPreview} alt="Community banner preview" />
                  {:else}
                    <div class="upload-placeholder">
                      <ImageIcon size="24" />
                      <span>Upload Banner</span>
                    </div>
                  {/if}
                  <input
                    type="file"
                    accept="image/*"
                    on:change={handleBannerChange}
                  />
                </div>
                {#if errors.banner}
                  <div class="error-message">
                    <AlertCircleIcon size="14" />
                    <span>{errors.banner}</span>
                  </div>
                {/if}
              </div>
            </div>

            <div class="form-group" data-error="categories">
              <label>
                Categories <span class="required">*</span>
                <span class="label-hint">(Select up to 5)</span>
              </label>

              <div class="templates-section">
                <h4>Quick Templates</h4>
                <div class="templates-grid">
                  {#each categoryTemplates as template}
                    <button
                      type="button"
                      class="template-button"
                      on:click={() => applyTemplate(template)}
                    >
                      {template.name}
                    </button>
                  {/each}
                </div>
              </div>

              <div class="categories-container {errors.categories ? "error" : ""}">
                <h4>Selected Categories: {selectedCategories.length}/5</h4>

                <div class="selected-categories">
                  {#if selectedCategories.length === 0}
                    <p class="no-selection">No categories selected</p>
                  {:else}
                    {#each selectedCategories as category}
                      <div class="selected-category">
                        <span>{category}</span>
                        <button
                          type="button"
                          class="remove-category"
                          on:click={() => toggleCategory(category)}
                          aria-label={`Remove ${category} category`}
                        >
                          <XIcon size="14" />
                        </button>
                      </div>
                    {/each}
                  {/if}
                </div>

                <h4>Available Categories</h4>
                <div class="categories-grid">
                  {#if availableCategories.length === 0}
                    <p class="no-categories">No categories available</p>
                  {:else}
                    {#each availableCategories as category}
                      <button
                        type="button"
                        class="category-chip {selectedCategories.includes(category) ? "selected" : ""}"
                        on:click={() => toggleCategory(category)}
                        disabled={selectedCategories.length >= 5 && !selectedCategories.includes(category)}
                      >
                        {#if selectedCategories.includes(category)}
                          <CheckIcon size="14" />
                        {/if}
                        <span>{category}</span>
                      </button>
                    {/each}
                  {/if}
                </div>

                {#if errors.categories}
                  <div class="error-message">
                    <AlertCircleIcon size="14" />
                    <span>{errors.categories}</span>
                  </div>
                {/if}
              </div>
            </div>

            <div class="form-group" data-error="rules">
              <label for="rules">
                Community Rules <span class="required">*</span>
              </label>
              <textarea
                id="rules"
                class={errors.rules ? "error" : ""}
                bind:value={rules}
                placeholder="Enter community rules (e.g., be respectful, no spamming, etc.)"
                rows="5"
              ></textarea>
              <div class="input-hint">
                <span>{rules.length} characters (minimum 50)</span>
              </div>
              {#if errors.rules}
                <div class="error-message">
                  <AlertCircleIcon size="14" />
                  <span>{errors.rules}</span>
                </div>
              {/if}
            </div>

            <div class="form-actions">
              <button
                type="button"
                class="cancel-button"
                on:click={handleClose}
                disabled={isSubmitting}
              >
                Cancel
              </button>
              <button
                type="submit"
                class="submit-button"
                disabled={isSubmitting}
              >
                {#if isSubmitting}
                  <Spinner size="small" />
                  <span>Submitting...</span>
                {:else}
                  <span>Create Community</span>
                {/if}
              </button>
            </div>
          </form>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000;
    padding: 20px;
    overflow-y: auto;
  }

  .modal-container {
    background-color: white;
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    width: 100%;
    max-width: 800px;
    max-height: 90vh;
    overflow-y: auto;
    animation: modalFadeIn 0.3s ease-out;
  }

  .dark .modal-container {
    background-color: #1a1a1a;
    color: #f0f0f0;
  }

  @keyframes modalFadeIn {
    from {
      opacity: 0;
      transform: translateY(-20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 24px;
    border-bottom: 1px solid #eaeaea;
  }

  .dark .modal-header {
    border-bottom-color: #333;
  }

  .modal-header h2 {
    font-size: 1.5rem;
    font-weight: 600;
    margin: 0;
  }

  .close-button {
    background: none;
    border: none;
    cursor: pointer;
    color: #666;
    border-radius: 50%;
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .close-button:hover {
    background-color: #f0f0f0;
  }

  .dark .close-button {
    color: #aaa;
  }

  .dark .close-button:hover {
    background-color: #333;
  }

  .modal-body {
    padding: 24px;
  }

  .modal-description {
    margin-bottom: 24px;
    color: #666;
    font-size: 0.95rem;
  }

  .dark .modal-description {
    color: #aaa;
  }

  .loading-container, .success-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 40px 0;
    text-align: center;
  }

  .success-icon {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    background-color: #4caf50;
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 32px;
    margin-bottom: 16px;
  }

  .success-container h3 {
    font-size: 1.5rem;
    margin-bottom: 8px;
  }

  .community-form {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .form-group {
    display: flex;
    flex-direction: column;
  }

  .form-row {
    display: flex;
    gap: 20px;
  }

  @media (max-width: 768px) {
    .form-row {
      flex-direction: column;
    }
  }

  label {
    font-weight: 500;
    margin-bottom: 8px;
    display: flex;
    align-items: center;
  }

  .label-hint {
    font-weight: normal;
    font-size: 0.85rem;
    color: #666;
    margin-left: 8px;
  }

  .dark .label-hint {
    color: #aaa;
  }

  .required {
    color: #f44336;
    margin-left: 4px;
  }

  input, textarea {
    padding: 10px 12px;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 0.95rem;
    transition: border-color 0.2s;
  }

  .dark input, .dark textarea {
    background-color: #2a2a2a;
    border-color: #444;
    color: #f0f0f0;
  }

  input:focus, textarea:focus {
    outline: none;
    border-color: #3498db;
  }

  .input-hint {
    font-size: 0.8rem;
    color: #666;
    margin-top: 4px;
    display: flex;
    justify-content: flex-end;
  }

  .dark .input-hint {
    color: #aaa;
  }

  input.error, textarea.error, select.error, .media-preview.error, .categories-container.error {
    border-color: #f44336;
  }

  .error-message {
    display: flex;
    align-items: center;
    gap: 6px;
    color: #f44336;
    font-size: 0.8rem;
    margin-top: 6px;
  }

  .media-upload {
    flex: 1;
  }

  .media-preview {
    height: 150px;
    border: 2px dashed #ddd;
    border-radius: 4px;
    overflow: hidden;
    position: relative;
    cursor: pointer;
    transition: all 0.2s;
  }

  .dark .media-preview {
    border-color: #444;
  }

  .media-preview:hover {
    border-color: #3498db;
  }

  .banner-preview {
    height: 100px;
  }

  .media-preview img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .upload-placeholder {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: #666;
    gap: 8px;
  }

  .dark .upload-placeholder {
    color: #aaa;
  }

  .media-preview input {
    position: absolute;
    width: 100%;
    height: 100%;
    top: 0;
    left: 0;
    opacity: 0;
    cursor: pointer;
  }

  .templates-section {
    margin-bottom: 16px;
  }

  .templates-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 10px;
    margin-top: 8px;
  }

  .template-button {
    padding: 8px 12px;
    background-color: #f5f5f5;
    border: 1px solid #ddd;
    border-radius: 4px;
    cursor: pointer;
    transition: all 0.2s;
    text-align: left;
    font-size: 0.9rem;
  }

  .template-button:hover {
    background-color: #e9e9e9;
    border-color: #ccc;
  }

  .dark .template-button {
    background-color: #333;
    border-color: #444;
    color: #f0f0f0;
  }

  .dark .template-button:hover {
    background-color: #3a3a3a;
  }

  .categories-container {
    border: 1px solid #ddd;
    border-radius: 4px;
    padding: 16px;
    margin-top: 8px;
  }

  .dark .categories-container {
    border-color: #444;
    background-color: #2a2a2a;
  }

  h4 {
    font-size: 0.95rem;
    font-weight: 500;
    margin: 0 0 10px 0;
  }

  .selected-categories {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 16px;
    min-height: 32px;
  }

  .no-selection {
    color: #888;
    font-style: italic;
    font-size: 0.9rem;
  }

  .dark .no-selection {
    color: #777;
  }

  .selected-category {
    display: flex;
    align-items: center;
    background-color: #e3f2fd;
    border-radius: 16px;
    padding: 4px 10px;
    font-size: 0.9rem;
    gap: 6px;
  }

  .dark .selected-category {
    background-color: #1e3a5f;
    color: #e3f2fd;
  }

  .remove-category {
    background: none;
    border: none;
    cursor: pointer;
    color: #666;
    padding: 2px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .dark .remove-category {
    color: #ccc;
  }

  .categories-grid {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    max-height: 200px;
    overflow-y: auto;
    padding: 4px;
    margin-top: 8px;
  }

  .category-chip {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    background-color: #f5f5f5;
    border: 1px solid #ddd;
    border-radius: 16px;
    cursor: pointer;
    transition: all 0.2s;
    font-size: 0.9rem;
  }

  .category-chip:hover:not(:disabled) {
    background-color: #e0e0e0;
  }

  .dark .category-chip {
    background-color: #333;
    border-color: #444;
    color: #f0f0f0;
  }

  .dark .category-chip:hover:not(:disabled) {
    background-color: #3a3a3a;
  }

  .category-chip.selected {
    background-color: #2196f3;
    color: white;
    border-color: #2196f3;
  }

  .dark .category-chip.selected {
    background-color: #1976d2;
    border-color: #1976d2;
  }

  .category-chip:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .no-categories {
    color: #888;
    font-style: italic;
    font-size: 0.9rem;
    margin: 0;
  }

  .dark .no-categories {
    color: #777;
  }

  .form-actions {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    margin-top: 16px;
  }

  .cancel-button, .submit-button {
    padding: 10px 20px;
    border-radius: 4px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .cancel-button {
    background: none;
    border: 1px solid #ddd;
    color: #666;
  }

  .cancel-button:hover:not(:disabled) {
    background-color: #f0f0f0;
  }

  .dark .cancel-button {
    border-color: #444;
    color: #ccc;
  }

  .dark .cancel-button:hover:not(:disabled) {
    background-color: #333;
  }

  .submit-button {
    background-color: #2196f3;
    border: none;
    color: white;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .submit-button:hover:not(:disabled) {
    background-color: #1e88e5;
  }

  .submit-button:disabled, .cancel-button:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }
</style>