import { mount } from "svelte";
import "./styles/index.css";
import App from "./App.svelte";
import { initializeSupabaseBuckets } from "./utils/supabase";
import { createLoggerWithPrefix } from "./utils/logger";

const logger = createLoggerWithPrefix("Main");


async function initApp() {
  try {
    await initializeSupabaseBuckets();
    logger.info("Supabase buckets initialized");
  } catch (error) {
    logger.error("Failed to initialize Supabase buckets:", error);
    logger.info("Continuing app initialization despite bucket setup failure");
  }

  
  mount(App, {
    target: document.getElementById("app")!,
  });
}

initApp();
