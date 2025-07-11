import { writable } from "svelte/store";

type ThemeType = "light" | "dark";

const createThemeStore = () => {

  const getInitialTheme = (): ThemeType => {
    try {

      const storedTheme = localStorage.getItem("theme") as ThemeType;
      if (storedTheme && (storedTheme === "light" || storedTheme === "dark")) {
        return storedTheme;
      }

      if (window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches) {
        return "dark";
      }
    } catch (error) {
      console.error("Error getting initial theme:", error);
    }

    return "light";
  };

  const theme = writable<ThemeType>("light");

  const applyThemeToDOM = (themeValue: ThemeType) => {
    if (typeof document === "undefined") return;

    document.documentElement.classList.remove("light", "dark", "light-theme", "dark-theme", "light-mode", "dark-mode");
    document.documentElement.classList.add(themeValue, `${themeValue}-theme`, `${themeValue}-mode`);
    document.documentElement.setAttribute("data-theme", themeValue);

    document.body.setAttribute("data-theme", themeValue);
  };

  if (typeof window !== "undefined") {
    const initialTheme = getInitialTheme();
    theme.set(initialTheme);

    if (typeof document !== "undefined") {
      applyThemeToDOM(initialTheme);
    }
  }

  if (typeof window !== "undefined" && window.matchMedia) {
    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");

    const handleChange = (event: MediaQueryListEvent) => {
      const currentTheme = localStorage.getItem("theme");

      if (!currentTheme) {
        const newTheme = event.matches ? "dark" : "light";
        theme.set(newTheme);
        applyThemeToDOM(newTheme);
      }
    };

    if (mediaQuery.addEventListener) {
      mediaQuery.addEventListener("change", handleChange);
    } else {

      mediaQuery.addListener(handleChange);
    }
  }

  return {
    subscribe: theme.subscribe,
    set: (value: ThemeType) => {
      theme.set(value);
      try {
        localStorage.setItem("theme", value);

        if (typeof document !== "undefined") {
          applyThemeToDOM(value);
        }
      } catch (error) {
        console.error("Error setting theme:", error);
      }
    },

    update: (callback: (value: ThemeType) => ThemeType) => {

      let currentValue: ThemeType = "light";

      const unsubscribe = theme.subscribe((value) => {
        currentValue = value;
      });
      unsubscribe();

      const newValue = callback(currentValue);

      theme.set(newValue);

      try {
        localStorage.setItem("theme", newValue);

        if (typeof document !== "undefined") {
          applyThemeToDOM(newValue);
        }
      } catch (error) {
        console.error("Error updating theme:", error);
      }

      return newValue;
    }
  };
};

const themeStore = createThemeStore();

export function useTheme() {

  const toggleTheme = () => {
    themeStore.update(currentTheme => {
      const newTheme = currentTheme === "light" ? "dark" : "light";
      return newTheme;
    });
  };

  return {
    theme: themeStore,
    toggleTheme
  };
}