<script lang="ts">
  import { onMount, createEventDispatcher } from 'svelte';
  import { Recaptcha, recaptcha, observer } from 'svelte-recaptcha-v2';
  
  export let siteKey = import.meta.env.VITE_RECAPTCHA_SITE_KEY || '';
  export let theme: 'dark' | 'light' = 'light';
  
  const dispatch = createEventDispatcher();
  let recaptchaWidget: Recaptcha;
  let token: string | null = null;
  let recaptchaLoaded = false;
  
  onMount(() => {
    if (!siteKey) {
      console.error('VITE_RECAPTCHA_SITE_KEY is missing in environment variables');
      dispatch('error', { message: 'reCAPTCHA configuration missing' });
    }
  });
  
  function handleRecaptchaSuccess(event: CustomEvent<{ token: string }>) {
    token = event.detail.token;
    dispatch('success', { token });
  }
  
  function handleRecaptchaError() {
    console.error('reCAPTCHA error occurred');
    dispatch('error', { message: 'reCAPTCHA verification failed' });
  }
  
  function handleRecaptchaExpired() {
    token = null;
    dispatch('expired', { message: 'reCAPTCHA token expired' });
  }
  
  function handleRecaptchaReady() {
    recaptchaLoaded = true;
    dispatch('ready');
  }
  
  export function execute(): Promise<string> {
    if (!recaptchaLoaded) {
      return Promise.reject(new Error('reCAPTCHA not loaded yet'));
    }
    
    if (recaptcha) {
      recaptcha.execute();
      return new Promise<string>((resolve, reject) => {
        observer.then((event) => {
          const recaptchaToken = event.detail?.token;
          if (recaptchaToken) {
            resolve(recaptchaToken);
          } else {
            reject(new Error('Failed to get reCAPTCHA token'));
          }
        }).catch(reject);
      });
    } else {
      return Promise.reject(new Error('reCAPTCHA not initialized'));
    }
  }
  
  export function reset() {
    if (recaptchaWidget) {
      recaptchaWidget.reset();
      token = null;
    }
  }
</script>

{#if siteKey}
<Recaptcha
  bind:this={recaptchaWidget}
  sitekey={siteKey}
  size="invisible"
  badge="bottomright"
  {theme}
  on:success={handleRecaptchaSuccess}
  on:error={handleRecaptchaError}
  on:expired={handleRecaptchaExpired}
  on:ready={handleRecaptchaReady}
/>
{:else}
<div class="hidden">reCAPTCHA missing configuration</div>
{/if} 