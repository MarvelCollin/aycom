package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	threadProto "github.com/Acad600-Tpa/WEB-MV-242/backend/services/thread/proto"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateThread creates a new thread
func CreateThread(c *gin.Context) {
	// Get user ID from token
	userIDAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User ID not found in token",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Invalid User ID format in token",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	// Parse request
	var request threadProto.CreateThreadRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Set user ID from token
	request.UserId = userID

	// Get connection to thread service
	conn, err := threadConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to thread service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer threadConnPool.Put(conn)

	// Create thread service client
	client := threadProto.NewThreadServiceClient(conn)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Call thread service
	resp, err := client.CreateThread(ctx, &request)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.InvalidArgument {
				httpStatus = http.StatusBadRequest
			}
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to create thread: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetThread gets a thread by ID
func GetThread(c *gin.Context) {
	// Get thread ID from URL
	threadID := c.Param("id")
	if threadID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Thread ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Get connection to thread service
	conn, err := threadConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to thread service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer threadConnPool.Put(conn)

	// Create thread service client
	client := threadProto.NewThreadServiceClient(conn)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Call thread service
	resp, err := client.GetThreadById(ctx, &threadProto.GetThreadRequest{
		ThreadId: threadID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			}
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to get thread: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetThreadsByUser gets threads by user ID
func GetThreadsByUser(c *gin.Context) {
	// Get user ID from URL or from token
	userID := c.Param("userId")
	if userID == "" {
		userIDAny, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Success: false,
				Message: "User ID not found in token",
				Code:    "UNAUTHORIZED",
			})
			return
		}

		var ok bool
		userID, ok = userIDAny.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Invalid User ID format in token",
				Code:    "INTERNAL_ERROR",
			})
			return
		}
	}

	// Get pagination parameters
	page := 1
	limit := 10

	pageStr := c.Query("page")
	if pageStr != "" {
		pageInt, err := strconv.Atoi(pageStr)
		if err == nil && pageInt > 0 {
			page = pageInt
		}
	}

	limitStr := c.Query("limit")
	if limitStr != "" {
		limitInt, err := strconv.Atoi(limitStr)
		if err == nil && limitInt > 0 && limitInt <= 50 {
			limit = limitInt
		}
	}

	// Get connection to thread service
	conn, err := threadConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to thread service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer threadConnPool.Put(conn)

	// Create thread service client
	client := threadProto.NewThreadServiceClient(conn)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Call thread service
	resp, err := client.GetThreadsByUser(ctx, &threadProto.GetThreadsByUserRequest{
		UserId: userID,
		Page:   int32(page),
		Limit:  int32(limit),
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			}
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to get threads: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateThread updates a thread
func UpdateThread(c *gin.Context) {
	// Get user ID from token
	userIDAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User ID not found in token",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Invalid User ID format in token",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	// Get thread ID from URL
	threadID := c.Param("id")
	if threadID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Thread ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Parse request
	var request threadProto.UpdateThreadRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Set thread ID from URL
	request.ThreadId = threadID

	// Get connection to thread service
	conn, err := threadConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to thread service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer threadConnPool.Put(conn)

	// Create thread service client
	client := threadProto.NewThreadServiceClient(conn)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Call thread service
	resp, err := client.UpdateThread(ctx, &request)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			} else if st.Code() == codes.InvalidArgument {
				httpStatus = http.StatusBadRequest
			} else if st.Code() == codes.PermissionDenied {
				httpStatus = http.StatusForbidden
			}
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to update thread: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteThread deletes a thread
func DeleteThread(c *gin.Context) {
	// Get user ID from token
	userIDAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User ID not found in token",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Invalid User ID format in token",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	// Get thread ID from URL
	threadID := c.Param("id")
	if threadID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Thread ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Get connection to thread service
	conn, err := threadConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to thread service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer threadConnPool.Put(conn)

	// Create thread service client
	client := threadProto.NewThreadServiceClient(conn)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Call thread service
	_, err = client.DeleteThread(ctx, &threadProto.DeleteThreadRequest{
		ThreadId: threadID,
		UserId:   userID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			} else if st.Code() == codes.PermissionDenied {
				httpStatus = http.StatusForbidden
			}
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to delete thread: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Thread deleted successfully",
	})
}
