/**
 * Formats a date to a human-readable relative time string (e.g. "2 hours ago")
 * @param dateString - ISO date string to format
 * @returns Formatted relative time string
 */
export function formatRelativeTime(dateString: string): string {
  if (!dateString) return "Unknown date";

  try {
    // Parse the date and ensure it's valid
    const date = new Date(dateString);

    // Check if the date is valid
    if (isNaN(date.getTime())) {
      console.warn("Invalid date string:", dateString);
      return "Invalid date";
    }

    const now = new Date();

    // Check if the date is in the future
    if (date > now) {
      console.warn("Date is in the future:", dateString);
      return "Just now";
    }

    const seconds = Math.round((now.getTime() - date.getTime()) / 1000);

    // Check if the time difference is unreasonably large (more than 10 years)
    if (seconds > 315360000) { // 10 years in seconds
      console.warn("Date is too far in the past:", dateString);
      return formatTimeForDisplay(date);
    }

    const minutes = Math.round(seconds / 60);
    const hours = Math.round(minutes / 60);
    const days = Math.round(hours / 24);
    const months = Math.round(days / 30);
    const years = Math.round(months / 12);

    if (seconds < 60) {
      return seconds <= 1 ? "just now" : `${seconds}s ago`;
    } else if (minutes < 60) {
      return minutes === 1 ? "1m ago" : `${minutes}m ago`;
    } else if (hours < 24) {
      return hours === 1 ? "1h ago" : `${hours}h ago`;
    } else if (days < 3) {
      // Show time for recent days
      return days === 1 ? "Yesterday" : `${days}d ago`;
    } else {
      // Use the actual date/time for older messages
      return formatTimeForDisplay(date);
    }
  } catch (error) {
    console.error("Error formatting date:", error);
    return "Date error";
  }
}

/**
 * Formats a date for display in a more readable format
 * @param date - Date to format
 * @returns Formatted date string
 */
export function formatTimeForDisplay(date: Date): string {
  try {
    // For dates within the current year, omit the year
    const now = new Date();
    const isCurrentYear = now.getFullYear() === date.getFullYear();

    // Create a unique timestamp with milliseconds for testing
    return date.toLocaleString("en-US", {
      month: "short",
      day: "numeric",
      year: isCurrentYear ? undefined : "numeric",
      hour: "numeric",
      minute: "2-digit",
      second: "2-digit",
      hour12: true
    }) + `:${date.getMilliseconds().toString().padStart(3, "0")}`;
  } catch (error) {
    console.error("Error formatting time for display:", error);
    return "Invalid date";
  }
}