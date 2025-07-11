import { writable } from "svelte/store";
import type { IDateOfBirth } from "../interfaces/IAuth";

export function useRegistrationForm() {

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
    verificationCode: "",
    recaptchaToken: ""
  });

  const errors = writable({
    name: "",
    username: "",
    email: "",
    password: [] as string[],
    confirmPassword: "",
    gender: "",
    dateOfBirth: "",
    securityQuestion: "",
    securityAnswer: "",
    profilePicture: "",
    banner: ""
  });

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

  const currentYear = new Date().getFullYear();

  const days = Array.from({ length: 31 }, (_, i) => (i + 1).toString());

  const years = Array.from({ length: 100 }, (_, i) => (currentYear - i).toString());

  const setFieldError = (field: string, errorMessage: string | string[]) => {
    errors.update(e => ({ ...e, [field]: errorMessage }));
  };

  const setServerErrors = (serverErrors: Record<string, string | string[]>) => {
    errors.update(e => ({ ...e, ...serverErrors }));
  };

  const clearErrors = () => {
    errors.update(() => ({
      name: "",
      username: "",
      email: "",
      password: [] as string[],
      confirmPassword: "",
      gender: "",
      dateOfBirth: "",
      securityQuestion: "",
      securityAnswer: "",
      profilePicture: "",
      banner: ""
    }));
  };

  const startTimer = () => {
    formState.update(state => {

      if (state.timerId) {
        clearInterval(state.timerId);
      }

      state.timeLeft = 300;
      state.showResendOption = false;

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
    return `${minutes}:${seconds < 10 ? "0" + seconds : seconds}`;
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
    setFieldError,
    setServerErrors,
    clearErrors,
    startTimer,
    formatTimeLeft,
    cleanupTimers
  };
}