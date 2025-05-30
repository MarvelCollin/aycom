import appConfig from '../config/appConfig';
import { writable } from 'svelte/store';

interface ServiceStatus {
  user_service: boolean;
  thread_service: boolean;
  community_service: boolean;
  last_checked: Date | null;
}

// Create a store to track service status
export const serviceStatus = writable<ServiceStatus>({
  user_service: true,  // Assume services are available initially
  thread_service: true,
  community_service: true,
  last_checked: null
});

// Check if user service is available
export async function checkUserServiceStatus(): Promise<boolean> {
  try {
    const API_BASE_URL = appConfig.api.baseUrl;
    const response = await fetch(`${API_BASE_URL}/health`, {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' },
    });
    
    const isAvailable = response.ok;
      // Update the store
    serviceStatus.update(status => ({
      ...status,
      user_service: isAvailable,
      last_checked: new Date()
    }));
    
    return isAvailable;
  } catch (error) {
    console.error('Error checking user service status:', error);
      // Update the store to indicate service is down
    serviceStatus.update(status => ({
      ...status,
      user_service: false,
      last_checked: new Date()
    }));
    
    return false;
  }
}

// Check if all services are available
export async function checkAllServices(): Promise<ServiceStatus> {
  const userServiceAvailable = await checkUserServiceStatus();
    // For now we're only checking user service, but we could add others
  const updatedStatus = {
    user_service: userServiceAvailable,
    thread_service: userServiceAvailable, // Assuming same availability for now
    community_service: userServiceAvailable, // Assuming same availability for now
    last_checked: new Date()
  };
  
  serviceStatus.set(updatedStatus);
  return updatedStatus;
}

// Schedule periodic service checks (every 30 seconds)
export function startServiceMonitoring(intervalMs = 30000): () => void {
  // Initial check
  checkAllServices();
  
  // Set up interval
  const intervalId = setInterval(checkAllServices, intervalMs);
  
  // Return function to stop monitoring
  return () => clearInterval(intervalId);
} 