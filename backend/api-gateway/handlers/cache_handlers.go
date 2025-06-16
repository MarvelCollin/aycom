package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"aycom/backend/api-gateway/utils"
)


func ClearCache(c *gin.Context) {
	cacheKey := c.Query("key")
	pattern := c.Query("pattern")
	
	ctx := context.Background()
	
	if pattern != "" {
		
		if err := utils.DeleteCachePattern(ctx, pattern); err != nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "CACHE_ERROR", "Failed to clear cache pattern: "+err.Error())
			return
		}
		utils.SendSuccessResponse(c, http.StatusOK, gin.H{
			"message": "Cache pattern cleared successfully",
			"pattern": pattern,
		})
		return
	}
	
	if cacheKey != "" {
		
		if err := utils.DeleteCache(ctx, cacheKey); err != nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "CACHE_ERROR", "Failed to clear cache key: "+err.Error())
			return
		}
		utils.SendSuccessResponse(c, http.StatusOK, gin.H{
			"message": "Cache key cleared successfully",
			"key":     cacheKey,
		})
		return
	}
	
	utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Either 'key' or 'pattern' parameter is required")
}


func GetCacheStats(c *gin.Context) {
	redisClient := utils.GetRedisClient()
	if redisClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "REDIS_UNAVAILABLE", "Redis client not available")
		return
	}
	
	ctx := context.Background()
	
	
	info, err := redisClient.Info(ctx, "memory").Result()
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "REDIS_ERROR", "Failed to get Redis info: "+err.Error())
		return
	}
	
	
	threadCategories, _ := redisClient.Exists(ctx, "thread_categories").Result()
	communityCategories, _ := redisClient.Exists(ctx, "community_categories").Result()
	userProfileKeys, _ := redisClient.Keys(ctx, "user_profile:*").Result()
	
	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"redis_info": info,
		"cache_keys": gin.H{
			"thread_categories_exists":    threadCategories > 0,
			"community_categories_exists": communityCategories > 0,
			"user_profile_count":         len(userProfileKeys),
		},
	})
}
