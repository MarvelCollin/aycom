package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ProductHandlers contains all product-related handlers

// ListProducts lists all products
func ListProducts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "list products endpoint",
	})
}

// GetProduct gets a product by ID
func GetProduct(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "get product endpoint",
	})
}

// CreateProduct creates a new product
func CreateProduct(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "create product endpoint",
	})
}

// UpdateProduct updates a product
func UpdateProduct(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "update product endpoint",
	})
}

// DeleteProduct deletes a product
func DeleteProduct(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "delete product endpoint",
	})
}
