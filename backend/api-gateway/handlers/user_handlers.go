package handlers

import (
	"context"
	"net/http"
	"time"

	userProto "github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/proto"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetUserProfile retrieves the user's profile
func GetUserProfile(c *gin.Context) {
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

	conn, err := userConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to user service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer userConnPool.Put(conn)

	client := userProto.NewUserServiceClient(conn)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	req := &userProto.GetUserRequest{
		Id: userID,
	}

	resp, err := client.GetUser(ctx, req)
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
				Message: "Failed to get user profile: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateUserProfile updates the user's profile
func UpdateUserProfile(c *gin.Context) {
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

	var request userProto.UpdateUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	request.Id = userID

	conn, err := userConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to user service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer userConnPool.Put(conn)

	client := userProto.NewUserServiceClient(conn)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := client.UpdateUser(ctx, &request)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			} else if st.Code() == codes.InvalidArgument {
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
				Message: "Failed to update user profile: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}
