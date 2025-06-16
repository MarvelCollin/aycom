import appConfig from "../config/appConfig";
import { writable } from "svelte/store";

interface ServiceStatus {
  user_service: boolean;
  thread_service: boolean;
  community_service: boolean;
  last_checked: Date | null;
}

export const serviceStatus = writable<ServiceStatus>({
  user_service: true,
  thread_service: true,
  community_service: true,
  last_checked: null
});

export async function checkUserServiceStatus(): Promise<boolean> {
  try {
    const API_BASE_URL = appConfig.api.baseUrl;
    const response = await fetch(`${API_BASE_URL}/health`, {
      method: "GET",
      headers: { "Content-Type": "application/json" },
    });

    const isAvailable = response.ok;

    serviceStatus.update(status => ({
      ...status,
      user_service: isAvailable,
      last_checked: new Date()
    }));

    return isAvailable;
  } catch (error) {
    console.error("Error checking user service status:", error);

    serviceStatus.update(status => ({
      ...status,
      user_service: false,
      last_checked: new Date()
    }));

    return false;
  }
}

export async function checkAllServices(): Promise<ServiceStatus> {
  const userServiceAvailable = await checkUserServiceStatus();

  const updatedStatus = {
    user_service: userServiceAvailable,
    thread_service: userServiceAvailable,
    community_service: userServiceAvailable,
    last_checked: new Date()
  };

  serviceStatus.set(updatedStatus);
  return updatedStatus;
}

export function startServiceMonitoring(intervalMs = 30000): () => void {

  checkAllServices();

  const intervalId = setInterval(checkAllServices, intervalMs);

  return () => clearInterval(intervalId);
}