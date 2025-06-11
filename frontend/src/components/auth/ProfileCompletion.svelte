<script lang="ts">
  import { onMount } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  import { useValidation } from '../../hooks/useValidation';
  import { useAuth } from '../../hooks/useAuth';
  import { toastStore } from '../../stores/toastStore';
  import type { IDateOfBirth } from '../../interfaces/IAuth';
  import appConfig from '../../config/appConfig';
  
  export let missingFields: string[] = [];
  export let onComplete: () => void;
  export let onSkip: () => void;
  
  // Form data
  let gender = "";
  let dateOfBirth: IDateOfBirth = { month: "", day: "", year: "" };
  let securityQuestion = "";
  let securityAnswer = "";
  let isLoading = false;
  let error = "";
  
  // Form options
  const months = [
    "January", "February", "March", "April", "May", "June", 
    "July", "August", "September", "October", "November", "December"
  ];
  
  const securityQuestions = [
    "What was the name of your first pet?",
    "What city were you born in?",
    "What is your favorite video game?",
    "What was the name of your first school?",
    "What was your childhood nickname?"
  ];
  
  // Get the current year
  const currentYear = new Date().getFullYear();
  // Generate days 1-31
  const days = Array.from({ length: 31 }, (_, i) => (i + 1).toString());
  // Generate years for the last 100 years
  const years = Array.from({ length: 100 }, (_, i) => (currentYear - i).toString());
  
  // Validation errors
  let genderError = "";
  let dateOfBirthError = "";
  let securityQuestionError = "";
  
  const { theme } = useTheme();
  const validation = useValidation();
  const { getProfile, updateProfile } = useAuth();
  
  $: isDarkMode = $theme === 'dark';
  $: showGenderField = missingFields.includes('gender');
  $: showDateOfBirthField = missingFields.includes('date_of_birth');
  $: showSecurityFields = missingFields.includes('security_question') || missingFields.includes('security_answer');
  
  // Load existing profile data if available
  onMount(async () => {
    try {
      const profileData = await getProfile();
      if (profileData?.user) {
        if (profileData.user.gender && profileData.user.gender !== 'unknown') {
          gender = profileData.user.gender;
        }
        
        if (profileData.user.date_of_birth) {
          const [month, day, year] = profileData.user.date_of_birth.split('-');
          if (month && day && year) {
            const monthIndex = parseInt(month);
            dateOfBirth = {
              month: months[monthIndex - 1] || "",
              day: day,
              year: year
            };
          }
        }
        
        if (profileData.user.security_question) {
          securityQuestion = profileData.user.security_question;
        }
      }
    } catch (err) {
      console.error("Failed to load profile data:", err);
    }
  });
  
  // Validation methods
  function validateField(field: string, value: any): boolean {
    let errorMessage = "";
    
    switch (field) {
      case 'gender':
        errorMessage = validation.validateGender(value);
        genderError = errorMessage;
        break;
      case 'dateOfBirth':
        errorMessage = validation.validateDateOfBirth(value, months);
        dateOfBirthError = errorMessage;
        break;
      case 'securityQuestion':
        errorMessage = validation.validateSecurityQuestion(securityQuestion, securityAnswer);
        securityQuestionError = errorMessage;
        break;
    }
    
    return !errorMessage;
  }
  
  function validateForm(): boolean {
    let isValid = true;
    
    if (showGenderField) {
      const isGenderValid = validateField('gender', gender);
      isValid = isValid && isGenderValid;
    }
    
    if (showDateOfBirthField) {
      const isDateValid = validateField('dateOfBirth', dateOfBirth);
      isValid = isValid && isDateValid;
    }
    
    if (showSecurityFields) {
      const isSecurityValid = validateField('securityQuestion', {
        question: securityQuestion,
        answer: securityAnswer
      });
      isValid = isValid && isSecurityValid;
    }
    
    return isValid;
  }
  
  async function handleSubmit() {
    if (!validateForm()) {
      error = "Please correct the errors in the form.";
      if (appConfig.ui.showErrorToasts) toastStore.showToast(error);
      return;
    }
    
    isLoading = true;
    error = "";
    
    try {
      const updateData: Record<string, any> = {};
      
      if (showGenderField) {
        updateData.gender = gender;
      }
      
      if (showDateOfBirthField && dateOfBirth.month && dateOfBirth.day && dateOfBirth.year) {
        const monthIndex = months.indexOf(dateOfBirth.month) + 1;
        updateData.date_of_birth = `${monthIndex}-${dateOfBirth.day}-${dateOfBirth.year}`;
      }
      
      if (showSecurityFields) {
        updateData.security_question = securityQuestion;
        updateData.security_answer = securityAnswer;
      }
      
      const result = await updateProfile(updateData);
      
      if (result.success) {
        toastStore.showToast("Profile updated successfully", "success");
        onComplete();
      } else {
        error = result.message || "Failed to update profile";
        if (appConfig.ui.showErrorToasts) toastStore.showToast(error, "error");
      }
    } catch (err) {
      console.error("Error updating profile:", err);
      error = err instanceof Error ? err.message : "An unexpected error occurred";
      if (appConfig.ui.showErrorToasts) toastStore.showToast(error, "error");
    } finally {
      isLoading = false;
    }
  }
</script>

<div class="profile-completion-container">
  <h2 class="text-xl font-semibold mb-4">Complete Your Profile</h2>
  <p class="text-gray-600 dark:text-gray-300 mb-6">
    Please provide the following information to complete your profile.
    This will help us personalize your experience and secure your account.
  </p>
  
  {#if error}
    <div class="bg-red-500 bg-opacity-10 border border-red-500 text-red-500 px-4 py-3 rounded mb-4">
      {error}
    </div>
  {/if}
  
  <form on:submit|preventDefault={handleSubmit}>
    {#if showGenderField}
      <div class="auth-input-group">
        <label class="auth-label">Gender</label>
        <div class="flex flex-wrap gap-4">
          <label class="flex items-center space-x-2 cursor-pointer">
            <input 
              type="radio" 
              name="gender" 
              value="male"
              bind:group={gender}
              on:change={() => validateField('gender', gender)}
              class="form-radio"
            />
            <span>Male</span>
          </label>
          
          <label class="flex items-center space-x-2 cursor-pointer">
            <input 
              type="radio" 
              name="gender" 
              value="female"
              bind:group={gender}
              on:change={() => validateField('gender', gender)}
              class="form-radio"
            />
            <span>Female</span>
          </label>
          
          <label class="flex items-center space-x-2 cursor-pointer">
            <input 
              type="radio" 
              name="gender" 
              value="other"
              bind:group={gender}
              on:change={() => validateField('gender', gender)}
              class="form-radio"
            />
            <span>Other</span>
          </label>
        </div>
        
        {#if genderError}
          <p class="auth-error-message">{genderError}</p>
        {/if}
      </div>
    {/if}
    
    {#if showDateOfBirthField}
      <div class="auth-input-group">
        <label class="auth-label">Date of Birth</label>
        <div class="flex space-x-2">
          <select 
            bind:value={dateOfBirth.month} 
            on:change={() => validateField('dateOfBirth', dateOfBirth)}
            class="auth-select {isDarkMode ? 'auth-select-dark' : ''} {dateOfBirthError ? 'auth-input-error' : ''}"
          >
            <option value="">Month</option>
            {#each months as month}
              <option value={month}>{month}</option>
            {/each}
          </select>
          
          <select 
            bind:value={dateOfBirth.day} 
            on:change={() => validateField('dateOfBirth', dateOfBirth)}
            class="auth-select {isDarkMode ? 'auth-select-dark' : ''} {dateOfBirthError ? 'auth-input-error' : ''}"
          >
            <option value="">Day</option>
            {#each days as day}
              <option value={day}>{day}</option>
            {/each}
          </select>
          
          <select 
            bind:value={dateOfBirth.year} 
            on:change={() => validateField('dateOfBirth', dateOfBirth)}
            class="auth-select {isDarkMode ? 'auth-select-dark' : ''} {dateOfBirthError ? 'auth-input-error' : ''}"
          >
            <option value="">Year</option>
            {#each years as year}
              <option value={year}>{year}</option>
            {/each}
          </select>
        </div>
        
        {#if dateOfBirthError}
          <p class="auth-error-message">{dateOfBirthError}</p>
        {/if}
      </div>
    {/if}
    
    {#if showSecurityFields}
      <div class="auth-input-group">
        <label class="auth-label">Security Question</label>
        <select 
          bind:value={securityQuestion} 
          on:change={() => validateField('securityQuestion', { question: securityQuestion, answer: securityAnswer })}
          class="auth-select {isDarkMode ? 'auth-select-dark' : ''} {securityQuestionError ? 'auth-input-error' : ''}"
        >
          <option value="">Select a security question</option>
          {#each securityQuestions as question}
            <option value={question}>{question}</option>
          {/each}
        </select>
        
        <label class="auth-label mt-4">Security Answer</label>
        <input 
          type="text" 
          bind:value={securityAnswer} 
          on:blur={() => validateField('securityQuestion', { question: securityQuestion, answer: securityAnswer })}
          class="auth-input {isDarkMode ? 'auth-input-dark' : ''} {securityQuestionError ? 'auth-input-error' : ''}"
          placeholder="Your answer"
        />
        
        {#if securityQuestionError}
          <p class="auth-error-message">{securityQuestionError}</p>
        {/if}
      </div>
    {/if}
    
    <div class="flex gap-4 mt-6">
      <button 
        type="button" 
        class="btn-secondary" 
        on:click={onSkip} 
        disabled={isLoading}
      >
        Skip for now
      </button>
      
      <button 
        type="submit" 
        class="btn-primary" 
        disabled={isLoading}
      >
        {#if isLoading}
          <span class="loading-spinner"></span>
        {/if}
        Complete Profile
      </button>
    </div>
  </form>
</div>

<style>
  .profile-completion-container {
    max-width: 100%;
    padding: 1.5rem;
    background-color: white;
    border-radius: 0.5rem;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  }
  
  :global(.dark) .profile-completion-container {
    background-color: #1f2937;
    color: #f3f4f6;
  }
  
  .loading-spinner {
    display: inline-block;
    width: 1rem;
    height: 1rem;
    border: 2px solid #ffffff;
    border-top-color: transparent;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-right: 0.5rem;
  }
  
  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style> 