import axios from 'axios';
import appConfig from '../config/appConfig';
import { getAuthToken } from '../utils/auth';
import { createLoggerWithPrefix } from '../utils/logger';

const logger = createLoggerWithPrefix('AI API');
const baseUrl = `${appConfig.api.baseUrl}/ai`;

/**
 * Predicts the most suitable category for a thread based on its content
 * @param content The content of the thread
 * @returns Promise with the predicted category and confidence score
 */
export async function predictThreadCategory(content: string) {
  try {
    logger.debug('Predicting category for thread content');
    
    const response = await axios.post(`${baseUrl}/predict-category`, {
      content
    });
    
    logger.debug('Category prediction result:', response.data);
    
    return {
      success: true,
      category: response.data.category,
      confidence: response.data.confidence,
      all_categories: response.data.all_categories
    };
  } catch (error) {
    logger.error('Failed to predict category:', error);
    
    return {
      success: false,
      category: 'general',
      confidence: 0,
      all_categories: { 
        'general': 1.0,
        'technology': 0,
        'politics': 0,
        'entertainment': 0,
        'sports': 0,
        'business': 0
      }
    };
  }
} 