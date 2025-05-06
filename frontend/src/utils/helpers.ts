/**
 * Various utility helper functions
 */

/**
 * Creates a debounced function that delays invoking func until after wait milliseconds
 * have elapsed since the last time the debounced function was invoked.
 * 
 * @param func - The function to debounce
 * @param wait - The number of milliseconds to delay
 * @returns A debounced version of the function
 */
export function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: ReturnType<typeof setTimeout> | null = null;
  
  return function(...args: Parameters<T>): void {
    const later = () => {
      timeout = null;
      func(...args);
    };
    
    if (timeout !== null) {
      clearTimeout(timeout);
    }
    
    timeout = setTimeout(later, wait);
  };
}

/**
 * Format a date relative to now (e.g. "5m", "2h", "3d")
 */
export function formatRelativeTime(date: string | Date): string {
  const now = new Date();
  const then = new Date(date);
  const diffMs = now.getTime() - then.getTime();
  
  // Convert to seconds
  const diffSec = Math.floor(diffMs / 1000);
  
  if (diffSec < 60) {
    return `${diffSec}s`;
  }
  
  // Convert to minutes
  const diffMin = Math.floor(diffSec / 60);
  
  if (diffMin < 60) {
    return `${diffMin}m`;
  }
  
  // Convert to hours
  const diffHour = Math.floor(diffMin / 60);
  
  if (diffHour < 24) {
    return `${diffHour}h`;
  }
  
  // Convert to days
  const diffDay = Math.floor(diffHour / 24);
  
  if (diffDay < 7) {
    return `${diffDay}d`;
  }
  
  // Convert to weeks
  const diffWeek = Math.floor(diffDay / 7);
  
  if (diffWeek < 4) {
    return `${diffWeek}w`;
  }
  
  // Use the short month name and day for older dates
  const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
  
  // For older posts, show the date
  return `${monthNames[then.getMonth()]} ${then.getDate()}`;
}

/**
 * Truncate text to a maximum length with ellipsis
 */
export function truncateText(text: string, maxLength: number): string {
  if (text.length <= maxLength) {
    return text;
  }
  
  // Find the last space within the maxLength to avoid cutting words
  const lastSpaceIndex = text.substring(0, maxLength).lastIndexOf(' ');
  
  // If no space found or it's at the beginning, just cut at maxLength
  if (lastSpaceIndex <= 0) {
    return text.substring(0, maxLength) + '...';
  }
  
  // Otherwise cut at the last space
  return text.substring(0, lastSpaceIndex) + '...';
}

/**
 * Implement Damerau-Levenshtein distance for fuzzy matching
 * Returns a number indicating the "distance" between two strings
 * Smaller numbers indicate more similar strings
 */
export function damerauLevenshteinDistance(a: string, b: string): number {
  if (a.length === 0) return b.length;
  if (b.length === 0) return a.length;

  const matrix: number[][] = [];

  // Initialize the matrix
  for (let i = 0; i <= a.length; i++) {
    matrix[i] = [i];
  }

  for (let j = 0; j <= b.length; j++) {
    matrix[0][j] = j;
  }

  // Fill the matrix
  for (let i = 1; i <= a.length; i++) {
    for (let j = 1; j <= b.length; j++) {
      const cost = a[i - 1] === b[j - 1] ? 0 : 1;
      
      matrix[i][j] = Math.min(
        matrix[i - 1][j] + 1,      // Deletion
        matrix[i][j - 1] + 1,      // Insertion
        matrix[i - 1][j - 1] + cost // Substitution
      );
      
      // Check for transposition
      if (i > 1 && j > 1 && a[i - 1] === b[j - 2] && a[i - 2] === b[j - 1]) {
        matrix[i][j] = Math.min(matrix[i][j], matrix[i - 2][j - 2] + cost);
      }
    }
  }

  return matrix[a.length][b.length];
}

/**
 * Calculate similarity between two strings (0 to 1)
 * Uses Damerau-Levenshtein distance normalized to string length
 * Higher values indicate more similar strings
 */
export function stringSimilarity(a: string, b: string): number {
  const distance = damerauLevenshteinDistance(a.toLowerCase(), b.toLowerCase());
  const maxLength = Math.max(a.length, b.length);
  
  if (maxLength === 0) return 1.0; // Both strings are empty
  
  // Convert distance to similarity (0 to 1)
  return 1.0 - (distance / maxLength);
}