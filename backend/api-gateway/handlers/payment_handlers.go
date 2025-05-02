package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary Create payment
// @Description Creates a new payment
// @Tags Payments
// @Accept json
// @Produce json
// @Param request body models.CreatePaymentRequest true "Create payment request"
// @Success 201 {object} models.PaymentResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/payments [post]
func CreatePayment(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "create payment endpoint",
	})
}

// @Summary Get payment
// @Description Returns a payment by ID
// @Tags Payments
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} models.PaymentResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/payments/{id} [get]
func GetPayment(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "get payment endpoint",
	})
}

// @Summary Get payment history
// @Description Returns the payment history for the authenticated user
// @Tags Payments
// @Produce json
// @Success 200 {array} models.PaymentResponse
// @Router /api/v1/payments/history [get]
func GetPaymentHistory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "get payment history endpoint",
	})
}
