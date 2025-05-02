package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreatePayment creates a new payment
func CreatePayment(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "create payment endpoint",
	})
}

// GetPayment gets a payment by ID
func GetPayment(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "get payment endpoint",
	})
}

// GetPaymentHistory gets a user's payment history
func GetPaymentHistory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "get payment history endpoint",
	})
}
