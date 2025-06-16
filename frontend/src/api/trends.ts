import { getAuthToken } from "../utils/auth";
import appConfig from "../config/appConfig";
import type { ITrend } from "../interfaces/ITrend";
import { createLoggerWithPrefix } from "../utils/logger";
import { useAuth } from "../hooks/useAuth";

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix("trends-api");

export async function getTrends(limit: number = 5): Promise<ITrend[]> {
  try {
    const { checkAndRefreshTokenIfNeeded } = useAuth();
    await checkAndRefreshTokenIfNeeded();

    const token = getAuthToken();
    const url = `${API_BASE_URL}/trends?limit=${limit}`;

    logger.debug(`Fetching trends from API: ${url}`);

    const response = await fetch(url, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        "Authorization": token ? `Bearer ${token}` : ""
      },
      credentials: "include"
    });

    if (!response.ok && response.status === 401 && token) {
      logger.debug("Got 401, retrying without auth token");

      const publicResponse = await fetch(url, {
        method: "GET",
        headers: {
          "Content-Type": "application/json"
        },
        credentials: "include"
      });

      if (publicResponse.ok) {
        const data = await publicResponse.json();
        if (!data || !data.trends || !Array.isArray(data.trends)) {
          logger.warn("Invalid data format from trends API (public response)");
          return [];
        }

        logger.info(`Successfully fetched ${data.trends.length} trends from API`);
        return data.trends;
      }

      logger.error(`Failed to fetch trends without auth: ${publicResponse.status}`);
      return [];
    }

    if (!response.ok) {
      logger.error(`Failed to fetch trends: ${response.status}`);
      return [];
    }

    const data = await response.json();

    if (data && data.data && Array.isArray(data.data.trends)) {
      logger.info(`Successfully fetched ${data.data.trends.length} trends from API`);
      return data.data.trends;
    } else if (data && Array.isArray(data.trends)) {
      logger.info(`Successfully fetched ${data.trends.length} trends from API`);
      return data.trends;
    } else {
      logger.warn("Invalid data format from trends API", { data });
      return [];
    }
  } catch (error: any) {
    logger.error("Failed to fetch trends", { error: error.message });
    return [];
  }
}

