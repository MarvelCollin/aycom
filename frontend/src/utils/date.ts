/**
 * Formats a date to a human-readable relative time string (e.g. "2 hours ago")
 * @param dateString - ISO date string to format
 * @returns Formatted relative time string
 */
export function formatRelativeTime(dateString: string): string {
  if (!dateString) return 'Unknown date';
  
  try {
    // Parse the date and ensure it's valid
    const date = new Date(dateString);
    
    // Check if the date is valid
    if (isNaN(date.getTime())) {
      console.warn('Invalid date string:', dateString);
      return 'Invalid date';
    }
    
    const now = new Date();
    
    // Check if the date is in the future
    if (date > now) {
      console.warn('Date is in the future:', dateString);
      return 'Just now';
    }
    
    const seconds = Math.round((now.getTime() - date.getTime()) / 1000);
    
    // Check if the time difference is unreasonably large (more than 10 years)
    if (seconds > 315360000) { // 10 years in seconds
      console.warn('Date is too far in the past:', dateString);
      return formatTimeForDisplay(date);
    }
    
    const minutes = Math.round(seconds / 60);
    const hours = Math.round(minutes / 60);
    const days = Math.round(hours / 24);
    const months = Math.round(days / 30);
    const years = Math.round(months / 12);

    if (seconds < 60) {
      return seconds <= 1 ? 'just now' : `${seconds} seconds ago`;
    } else if (minutes < 60) {
      return minutes === 1 ? '1 minute ago' : `${minutes} minutes ago`;
    } else if (hours < 24) {
      return hours === 1 ? '1 hour ago' : `${hours} hours ago`;
    } else if (days < 30) {
      return days === 1 ? '1 day ago' : `${days} days ago`;
    } else if (months < 12) {
      return months === 1 ? '1 month ago' : `${months} months ago`;
    } else {
      return years === 1 ? '1 year ago' : `${years} years ago`;
    }
  } catch (error) {
    console.error('Error formatting date:', error);
    return 'Date error';
  }
}

/**
 * Formats a date for display in a more readable format
 * @param date - Date to format
 * @returns Formatted date string
 */
export function formatTimeForDisplay(date: Date): string {
  try {
    return date.toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: 'numeric',
      minute: 'numeric',
      hour12: true
    });
  } catch (error) {
    console.error('Error formatting time for display:', error);
    return 'Invalid date';
  }
} 