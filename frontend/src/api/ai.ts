import axios from 'axios';
import appConfig from '../config/appConfig';
import { getAuthToken } from '../utils/auth';
import { createLoggerWithPrefix } from '../utils/logger';

const logger = createLoggerWithPrefix('AI API');
const baseUrl = `${appConfig.api.baseUrl}/ai`;

const predictionCache = new Map<string, any>();

interface ICategoryPredictionResponse {
  success: boolean;
  category?: string;
  confidence?: number;
  all_categories?: Record<string, number>;
  error?: string;
}

export async function predictThreadCategory(content: string): Promise<ICategoryPredictionResponse> {
  try {
    if (!content || content.trim().length < 5) {
      return {
        success: false,
        error: "Content too short for prediction"
      };
    }

    const trimmed = content.trim();
    const cacheKey = trimmed.substring(0, 100);
    if (predictionCache.has(cacheKey)) {
      logger.debug('Using cached prediction result');
      return predictionCache.get(cacheKey);
    }
    
    logger.debug('Predicting category for thread content');
    
    const response = await axios.post(`${baseUrl}/predict-category`, {
      content: trimmed
    }, { 
      timeout: 8000
    });
    
    if (response.data && response.data.error) {
      logger.warn('AI service returned error:', response.data.error);
      return {
        success: false,
        error: response.data.error
      };
    }
    
    logger.debug('Category prediction result:', response.data);
    
    const result = {
      success: true,
      category: response.data.category,
      confidence: response.data.confidence,
      all_categories: response.data.all_categories
    };

    predictionCache.set(cacheKey, result);
    
    return result;
  } catch (error) {
    logger.error('Failed to predict category:', error);
    return {
      success: false,
      error: "Failed to predict category"
    };
  }
}