import { writable } from 'svelte/store';
import { useValidation } from './useValidation';
import type { IDateOfBirth } from '../interfaces/IAuth';

export function useRegistrationForm() {
  // Form data
  const formData = writable({
    name: "",
    email: "",
    username: "",
    password: "",
    confirmPassword: "",
    gender: "",
    dateOfBirth: {
      month: "",
      day: "",
      year: ""
    } as IDateOfBirth,
    profilePicture: null as File | string | null,
    banner: null as File | string | null,
    securityQuestion: "",
    securityAnswer: "",
    subscribeToNewsletter: false,
    verificationCode: ""
  });

  // Validation errors
  const errors = writable({
    name: "",
    username: "",
    email: "",
    password: [] as string[],
    confirmPassword: "",
    gender: "",
    dateOfBirth: "",
    securityQuestion: "",
    profilePicture: "",
    banner: ""
  });

  // Multi-step form state
  const formState = writable({
    step: 1,
    loading: false,
    success: false,
    error: "",
    showResendOption: false,
    timeLeft: 300,
    redirectCountdown: 3,
    timerId: undefined as number | undefined
  });

  // Constants for form
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

  // Get validation functions
  const validation = useValidation();

  // Validation methods
  const validateFormField = (field: string, value: any): boolean => {
    const updateErrors = (key: string, value: string | string[]) => {
      errors.update((e) => ({ ...e, [key]: value }));
    };

    switch (field) {
      case 'name':
        updateErrors('name', validation.validateName(value));
        break;
      case 'username':
        updateErrors('username', validation.validateUsername(value));
        break;
      case 'email':
        updateErrors('email', validation.validateEmail(value));
        break;
      case 'password':
        updateErrors('password', validation.validatePassword(value));
        // Also validate confirm password if it exists
        formData.update(data => {
          if (data.confirmPassword) {
            updateErrors('confirmPassword', validation.validateConfirmPassword(value, data.confirmPassword));
          }
          return data;
        });
        break;
      case 'confirmPassword':
        formData.update(data => {
          updateErrors('confirmPassword', validation.validateConfirmPassword(data.password, value));
          return data;
        });
        break;
      case 'gender':
        updateErrors('gender', validation.validateGender(value));
        break;
      case 'dateOfBirth':
        updateErrors('dateOfBirth', validation.validateDateOfBirth(value, months));
        break;
      case 'securityQuestion':
      case 'securityAnswer':
        formData.update(data => {
          updateErrors('securityQuestion', validation.validateSecurityQuestion(data.securityQuestion, data.securityAnswer));
          return data;
        });
        break;
      case 'profilePicture':
        updateErrors('profilePicture', value ? "" : "Profile picture is required");
        break;
      case 'banner':
        updateErrors('banner', value ? "" : "Banner image is required");
        break;
    }

    let isValid = false;
    errors.update(e => {
      if (field === 'password') {
        isValid = e[field].length === 0;
      } else if (field === 'securityQuestion' || field === 'securityAnswer') {
        isValid = !e.securityQuestion;
      } else {
        isValid = !e[field];
      }
      return e;
    });

    return isValid;
  };

  const validateStep1 = (): boolean => {
    let isFormValid = true;
    
    formData.update(data => {
      // Validate all form fields
      const fieldsToValidate = [
        'name', 'username', 'email', 'password', 'confirmPassword', 
        'gender', 'dateOfBirth', 'securityQuestion', 'profilePicture', 'banner'
      ];

      fieldsToValidate.forEach(field => {
        const fieldValue = field === 'dateOfBirth' ? data.dateOfBirth : data[field];
        const isFieldValid = validateFormField(field, fieldValue);
        if (!isFieldValid) {
          isFormValid = false;
        }
      });

      return data;
    });
    
    return isFormValid;
  };

  // Timer functions for verification code
  const startTimer = () => {
    formState.update(state => {
      // Clear any existing timer
      if (state.timerId) {
        clearInterval(state.timerId);
      }
      
      // Reset timer to 5 minutes (300 seconds)
      state.timeLeft = 300;
      state.showResendOption = false;
      
      // Start a new timer
      state.timerId = window.setInterval(() => {
        formState.update(s => {
          s.timeLeft -= 1;
          
          if (s.timeLeft <= 0) {
            clearInterval(s.timerId);
            s.showResendOption = true;
          }
          
          return s;
        });
      }, 1000);
      
      return state;
    });
  };

  const formatTimeLeft = () => {
    let timeLeft = 0;
    formState.update(state => {
      timeLeft = state.timeLeft;
      return state;
    });
    
    const minutes = Math.floor(timeLeft / 60);
    const seconds = timeLeft % 60;
    return `${minutes}:${seconds < 10 ? '0' + seconds : seconds}`;
  };

  const cleanupTimers = () => {
    formState.update(state => {
      if (state.timerId) {
        clearInterval(state.timerId);
      }
      return state;
    });
  };

  return {
    formData,
    errors,
    formState,
    months,
    days,
    years,
    securityQuestions,
    validateFormField,
    validateStep1,
    startTimer,
    formatTimeLeft,
    cleanupTimers
  };
} 