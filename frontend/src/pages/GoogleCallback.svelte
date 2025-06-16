<script lang="ts">
  import { onMount } from "svelte";
  import { useAuth } from "../hooks/useAuth";
  import AuthCallback from "../components/auth/AuthCallback.svelte";
  import type { IGoogleCredentialResponse } from "../interfaces/IAuth";
  import { toastStore } from "../stores/toastStore";

  const { handleGoogleAuth } = useAuth();

  let loading = true;
  let error = "";
  let success = false;
  let redirectCountdown = 3;

  function startRedirectCountdown() {
    success = true;
    const interval = setInterval(() => {
      redirectCountdown--;
      if (redirectCountdown <= 0) {
        clearInterval(interval);
        window.location.href = "/feed";
      }
    }, 1000);
  }

  onMount(() => {

    const params = new URLSearchParams(window.location.search);

    const hashParams = new URLSearchParams(window.location.hash.substring(1));

    const state = window.history.state;

    const credential =
      params.get("credential") ||
      hashParams.get("credential") ||
      (state && state.credential);

    if (credential) {

      handleGoogleAuth({ credential } as IGoogleCredentialResponse)
        .then(result => {
          loading = false;
          if (result.success) {

            startRedirectCountdown();
          } else {
            error = result.message || "Google authentication failed";
          }
        })
        .catch(err => {
          console.error("Error handling Google auth:", err);
          toastStore.showToast("Google authentication failed. Please try again.", "error");
          error = "An unexpected error occurred";
          loading = false;
        });
    } else {
      error = "No authentication credential found";
      loading = false;
    }
  });
</script>

<div class="min-h-screen flex items-center justify-center bg-black text-white">
  <AuthCallback {loading} {error} {success} {redirectCountdown} />
</div>