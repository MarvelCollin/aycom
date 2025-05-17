/**
 * Format a timestamp into a relative time ago string
 * @param {string} timestamp - ISO timestamp string
 * @returns {string} - Formatted time ago string (e.g., "2h", "3d", "1w")
 */
export function formatTimeAgo(timestamp) {
  if (!timestamp) return 'now';
  
  try {
    const date = new Date(timestamp);
    
    if (isNaN(date.getTime())) {
      return 'now';
    }
    
    const now = new Date();
    const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);
    
    if (seconds < 0) return 'now';
    
    // Less than a minute
    if (seconds < 60) {
      return `${seconds}s`;
    }
    
    // Less than an hour
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) {
      return `${minutes}m`;
    }
    
    // Less than a day
    const hours = Math.floor(minutes / 60);
    if (hours < 24) {
      return `${hours}h`;
    }
    
    // Less than a week
    const days = Math.floor(hours / 24);
    if (days < 7) {
      return `${days}d`;
    }
    
    // Less than a month
    const weeks = Math.floor(days / 7);
    if (weeks < 4) {
      return `${weeks}w`;
    }
    
    // Less than a year
    const months = Math.floor(days / 30);
    if (months < 12) {
      return `${months}mo`;
    }
    
    // Years
    const years = Math.floor(days / 365);
    return `${years}y`;
    
  } catch (error) {
    console.error('Error formatting timestamp:', error);
    return 'now';
  }
}

/**
 * Extracts user metadata from content that might contain embedded information
 * @param {string} content - Content string that might contain embedded metadata
 * @returns {object} - Object with extracted username, displayName, and cleaned content
 */
export function processUserMetadata(content) {
  if (!content || typeof content !== 'string') {
    return { username: null, displayName: null, content: content || '' };
  }
  
  let username = null;
  let displayName = null;
  let cleanedContent = content;
  
  // Look for possible embedded usernames in format @username
  const usernameMatch = content.match(/@([a-zA-Z0-9_]+)/);
  if (usernameMatch && usernameMatch[1]) {
    username = usernameMatch[1];
  }
  
  // Look for possible embedded display names in format (Display Name)
  const displayNameMatch = content.match(/\(([^)]+)\)/);
  if (displayNameMatch && displayNameMatch[1]) {
    displayName = displayNameMatch[1];
  }
  
  // Clean the content if we found any metadata
  if (username || displayName) {
    cleanedContent = content
      .replace(/@([a-zA-Z0-9_]+)/, '')
      .replace(/\(([^)]+)\)/, '')
      .trim();
  }
  
  return {
    username,
    displayName,
    content: cleanedContent
  };
} 