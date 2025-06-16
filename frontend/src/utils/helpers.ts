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

  const monthNames = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];

  return `${monthNames[then.getMonth()]} ${then.getDate()}`;
}

export function truncateText(text: string, maxLength: number): string {
  if (text.length <= maxLength) {
    return text;
  }

  const lastSpaceIndex = text.substring(0, maxLength).lastIndexOf(" ");

  if (lastSpaceIndex <= 0) {
    return text.substring(0, maxLength) + "...";
  }

  return text.substring(0, lastSpaceIndex) + "...";
}