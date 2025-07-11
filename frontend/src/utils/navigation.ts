import { toastStore } from "../stores/toastStore";

export function navigate(
  route: string,
  options: {
    replace?: boolean,
    showToast?: boolean,
    toastMessage?: string,
    toastType?: "info" | "success" | "warning" | "error"
  } = {}
) {
  const {
    replace = false,
    showToast = false,
    toastMessage,
    toastType = "info"
  } = options;

  if (showToast && toastMessage) {
    toastStore.showToast(toastMessage, toastType);
  }

  if (replace) {
    window.history.replaceState({}, "", route);
  } else {
    window.history.pushState({}, "", route);
  }

  window.dispatchEvent(new CustomEvent("navigate", { detail: { route } }));
}

export function navigateToLogin(
  options: {
    replace?: boolean,
    showToast?: boolean,
    toastMessage?: string
  } = {}
) {
  const {
    replace = true,
    showToast = true,
    toastMessage = "You need to log in to access this feature"
  } = options;

  navigate("/login", {
    replace,
    showToast,
    toastMessage,
    toastType: "warning"
  });
}

export function navigateToHome(
  options: {
    replace?: boolean,
    showToast?: boolean,
    toastMessage?: string,
    toastType?: "info" | "success" | "warning" | "error"
  } = {}
) {
  navigate("/feed", options);
}

export function navigateToUserProfile(
  userId: string,
  options: {
    replace?: boolean,
    showToast?: boolean,
    toastMessage?: string,
    toastType?: "info" | "success" | "warning" | "error"
  } = {}
) {
  navigate(`/user/${userId}`, options);
}

export function goBack(fallbackRoute = "/") {
  if (window.history.length > 1) {
    window.history.back();
  } else {
    navigate(fallbackRoute, { replace: true });
  }
}