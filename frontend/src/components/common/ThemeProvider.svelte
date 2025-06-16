<script lang="ts">
  import { onMount } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';

  const { theme } = useTheme();

  onMount(() => {

    try {

      const storedFontSize = localStorage.getItem('fontSize');
      if (storedFontSize) {
        document.documentElement.classList.remove('font-small', 'font-medium', 'font-large');
        document.documentElement.classList.add(`font-${storedFontSize}`);
      } else {

        document.documentElement.classList.add('font-medium');
      }

      const storedFontColor = localStorage.getItem('fontColor');
      if (storedFontColor) {
        document.documentElement.classList.remove('text-default', 'text-blue', 'text-green', 'text-purple');
        document.documentElement.classList.add(`text-${storedFontColor}`);

        const htmlElement = document.querySelector('html');
        if (htmlElement) {
          htmlElement.classList.remove('text-default', 'text-blue', 'text-green', 'text-purple');
          htmlElement.classList.add(`text-${storedFontColor}`);
        }
      } else {

        document.documentElement.classList.add('text-default');
        const htmlElement = document.querySelector('html');
        if (htmlElement) {
          htmlElement.classList.add('text-default');
        }
      }
    } catch (error) {
      console.error('Error initializing font preferences:', error);
    }

    const unsubscribe = theme.subscribe(currentTheme => {

      document.documentElement.setAttribute('data-theme', currentTheme);

      const htmlElement = document.querySelector('html');
      if (htmlElement) {
        htmlElement.setAttribute('data-theme', currentTheme);
      }

      if (currentTheme === 'dark') {
        document.documentElement.classList.add('dark-theme');
        document.documentElement.classList.remove('light-theme');

        if (htmlElement) {
          htmlElement.classList.add('dark-theme');
          htmlElement.classList.remove('light-theme');
        }
      } else {
        document.documentElement.classList.add('light-theme');
        document.documentElement.classList.remove('dark-theme');

        if (htmlElement) {
          htmlElement.classList.add('light-theme');
          htmlElement.classList.remove('dark-theme');
        }
      }

      document.body.setAttribute('data-theme', currentTheme);
    });

    return () => {
      unsubscribe();
    };
  });
</script>

<div class="theme-provider">
  <slot />
</div>

<style>
  .theme-provider {
    display: contents;
  }
</style>