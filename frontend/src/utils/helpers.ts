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

export function formatRelativeTime(date: string | Date): string {
  const now = new Date();
  const then = new Date(date);
  const diffMs = now.getTime() - then.getTime();

  const diffSec = Math.floor(diffMs / 1000);

  if (diffSec < 60) {
    return `${diffSec}s`;
  }

  const diffMin = Math.floor(diffSec / 60);

  if (diffMin < 60) {
    return `${diffMin}m`;
  }

  const diffHour = Math.floor(diffMin / 60);

  if (diffHour < 24) {
    return `${diffHour}h`;
  }

  const diffDay = Math.floor(diffHour / 24);

  if (diffDay < 7) {
    return `${diffDay}d`;
  }

  const diffWeek = Math.floor(diffDay / 7);

  if (diffWeek < 4) {
    return `${diffWeek}w`;
  }

  const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

  return `${monthNames[then.getMonth()]} ${then.getDate()}`;
}

export function truncateText(text: string, maxLength: number): string {
  if (text.length <= maxLength) {
    return text;
  }

  const lastSpaceIndex = text.substring(0, maxLength).lastIndexOf(' ');

  if (lastSpaceIndex <= 0) {
    return text.substring(0, maxLength) + '...';
  }

  return text.substring(0, lastSpaceIndex) + '...';
}

export function damerauLevenshteinDistance(a: string, b: string): number {
  if (a.length === 0) return b.length;
  if (b.length === 0) return a.length;

  const matrix: number[][] = [];

  for (let i = 0; i <= a.length; i++) {
    matrix[i] = [i];
  }

  for (let j = 0; j <= b.length; j++) {
    matrix[0][j] = j;
  }

  for (let i = 1; i <= a.length; i++) {
    for (let j = 1; j <= b.length; j++) {
      const cost = a[i - 1] === b[j - 1] ? 0 : 1;

      matrix[i][j] = Math.min(
        matrix[i - 1][j] + 1,
        matrix[i][j - 1] + 1,
        matrix[i - 1][j - 1] + cost
      );

      if (i > 1 && j > 1 && a[i - 1] === b[j - 2] && a[i - 2] === b[j - 1]) {
        matrix[i][j] = Math.min(matrix[i][j], matrix[i - 2][j - 2] + cost);
      }
    }
  }

  return matrix[a.length][b.length];
}

export function stringSimilarity(a: string, b: string): number {
  const distance = damerauLevenshteinDistance(a.toLowerCase(), b.toLowerCase());
  const maxLength = Math.max(a.length, b.length);

  if (maxLength === 0) return 1.0;

  return 1.0 - (distance / maxLength);
}

export function fuzzySearch<T>(
  needle: string,
  haystack: T[],
  key?: string,
  threshold: number = 0.6
): T[] {
  if (!needle || !haystack || !haystack.length) return [];

  const cleanNeedle = needle.toLowerCase().trim();
  if (!cleanNeedle) return [];

  const results = haystack
    .map(item => {

      let targetString = '';
      if (typeof item === 'string') {
        targetString = item;
      } 

      else if (key && typeof item === 'object' && item !== null) {
        const objItem = item as any;
        targetString = objItem[key]?.toString() || '';
      }

      const similarity = stringSimilarity(cleanNeedle, targetString.toLowerCase());
      return { item, similarity };
    })
    .filter(result => result.similarity >= threshold)
    .sort((a, b) => b.similarity - a.similarity)
    .map(result => result.item);

  return results;
}