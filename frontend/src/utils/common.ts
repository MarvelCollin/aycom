import { toastStore } from "../stores/toastStore";
import type { IAuthStore } from "../interfaces/IAuth";
import type { IMedia } from "../interfaces/IMedia";
import { createLoggerWithPrefix } from "./logger";

const logger = createLoggerWithPrefix("common-utils");

export function formatTimeAgo(timestamp: string): string {
  if (!timestamp) return "now";

  try {
    const date = new Date(timestamp);

    if (isNaN(date.getTime())) {
      return "now";
    }

    const now = new Date();
    const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

    if (seconds < 0) return "now";

    if (seconds < 60) {
      return `${seconds}s`;
    }

    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) {
      return `${minutes}m`;
    }

    const hours = Math.floor(minutes / 60);
    if (hours < 24) {
      return `${hours}h`;
    }

    const days = Math.floor(hours / 24);
    if (days < 7) {
      return `${days}d`;
    }

    const weeks = Math.floor(days / 7);
    if (weeks < 4) {
      return `${weeks}w`;
    }

    const months = Math.floor(days / 30);
    if (months < 12) {
      return `${months}mo`;
    }

    const years = Math.floor(days / 365);
    return `${years}y`;

  } catch (error) {
    console.error("Error formatting timestamp:", error);
    return "now";
  }
}

export function checkAuth(authState: IAuthStore, featureName: string): boolean {
  if (!authState.is_authenticated) {
    toastStore.showToast(`You need to log in to access ${featureName}`, "warning");
    window.location.href = "/login";
    return false;
  }
  return true;
}

export function isWithinTime(timestamp: string, withinMs: number = 60000): boolean {
  try {
    const time = new Date(timestamp).getTime();
    const now = new Date().getTime();
    const diffMs = now - time;
    return diffMs < withinMs;
  } catch (e) {
    return false;
  }
}

export function generateFilePreview(file: File): IMedia {
  const url = URL.createObjectURL(file);
  let type: "image" | "video" | "gif";

  if (file.type.startsWith("image/")) {
    type = file.type === "image/gif" ? "gif" : "image";
  } else if (file.type.startsWith("video/")) {
    type = "video";
  } else {

    type = "image";
  }

  return {
    url,
    type,
    alt_text: file.name
  };
}

export function processUserMetadata(content: string): { username?: string, name?: string, content: string } {
  if (!content) return { content: "" };

  const userMetadataRegex = /^\[USER:([^@\]]+)(?:@([^\]]+))?\](.*)/;
  const match = content.match(userMetadataRegex);

  if (match) {
    return {
      username: match[1] || undefined,
      name: match[2] || undefined,
      content: match[3] || ""
    };
  }

  return { content };
}

export function handleApiError(error: unknown): { success: false, message: string } {
  if (error instanceof Error) {
    if (error.name === "AbortError") {
      return { success: false, message: "Request timed out. Please try again." };
    }
    return { success: false, message: error.message };
  }
  return { success: false, message: "An unexpected error occurred." };
}

export function truncateText(text: string, maxLength: number = 100): string {
  if (!text || text.length <= maxLength) return text;
  return text.substring(0, maxLength) + "...";
}

export function isSupabaseStorageUrl(url: string): boolean {
  try {
    const urlObj = new URL(url);
    return urlObj.hostname.includes("supabase.co") &&
           (urlObj.pathname.includes("/storage/v1/object/public/") ||
            urlObj.pathname.includes("/storage/v1/s3/"));
  } catch (error) {

    return url.includes("supabase.co/storage/v1/object/public/") ||
           url.includes("supabase.co/storage/v1/s3/");
  }
}

export function formatStorageUrl(url: string | null): string {
  if (!url) return "";

  console.log("Original URL:", url);

  if (url.startsWith("http://") || url.startsWith("https://")) {

    if (url.includes("//storage/v1/s3/")) {
      return url.replace("//storage/v1/s3/", "/storage/v1/s3/");
    }

    try {
      const urlObj = new URL(url);

      if (url.includes("/storage/v1/object/public/") && !url.includes("/storage/v1/s3/")) {
        return url.replace("/storage/v1/object/public/", "/storage/v1/s3/");
      }

      console.log("URL hostname:", urlObj.hostname);
    } catch (e) {
      console.error("Error parsing URL:", e);
    }

    return url;
  }

  const supabaseUrl = import.meta.env.VITE_SUPABASE_URL || "https://sdhtnvlmuywinhcglfsu.supabase.co";

  if (url.startsWith("/")) {
    if (!url.includes("storage/")) {
      return `${supabaseUrl}/storage/v1/s3${url.slice(1)}`;
    }
  }

  if (url.includes("storage/v1/object/public/")) {
    const formatted = url.replace("storage/v1/object/public/", "storage/v1/s3/");

    if (formatted.startsWith("storage/")) {
      return `${supabaseUrl}/${formatted}`;
    }

    return formatted;
  }

  if (url.includes("storage/v1/s3/")) {
    if (url.startsWith("storage/")) {
      return `${supabaseUrl}/${url}`;
    }

    return url;
  }

  const knownBuckets = ["profile-pictures", "banners", "thread-media", "user-media", "media", "tpaweb", "test", "uploads"];
  for (const bucket of knownBuckets) {
    if (url.startsWith(`${bucket}/`)) {
      return `${supabaseUrl}/storage/v1/s3/${url}`;
    }
  }

  if (url.includes("community/community_")) {
    return `${supabaseUrl}/storage/v1/s3/uploads/${url}`;
  }

  if (url.includes("_1/")) {

    return `${supabaseUrl}/storage/v1/s3/${url}`;
  }

  const cleanPath = url.replace(/^\
  return `${supabaseUrl}/storage/v1/s3/${cleanPath}`;
}

export function formatNumber(num: number): string {
  if (num === undefined || num === null) return "0";

  if (num === 0) return "0";

  const absNum = Math.abs(num);
  const sign = num < 0 ? "-" : "";

  if (absNum < 1000) {
    return sign + absNum.toString();
  }

  const abbreviations = ["", "K", "M", "B", "T"];
  const tier = Math.floor(Math.log10(absNum) / 3);

  if (tier >= abbreviations.length) {
    return sign + absNum.toString(); 
  }

  const scale = Math.pow(10, tier * 3);
  const scaled = absNum / scale;

  const formatted = scaled.toFixed(1).replace(/\.0$/, "");

  return sign + formatted + abbreviations[tier];
}