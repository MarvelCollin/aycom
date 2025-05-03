package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ProductHandlers contains all product-related handlers

// @Summary List products
// @Description Returns a list of products
// @Tags Products
// @Produce json
// @Router /api/v1/products [get]
func ListProducts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "list products endpoint",
	})
}

// @Summary Get product
// @Description Returns a product by ID
// @Tags Products
// @Produce json
// @Param id path string true "Product ID"
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/products/{id} [get]
func GetProduct(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "get product endpoint",
	})
}

// @Summary Create product
// @Description Creates a new product
// @Tags Products
// @Accept json
// @Produce json
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/products [post]
func CreateProduct(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "create product endpoint",
	})
}

// @Summary Update product
// @Description Updates a product by ID
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/products/{id} [put]
func UpdateProduct(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "update product endpoint",
	})
}

// @Summary Delete product
// @Description Deletes a product by ID
// @Tags Products
// @Produce json
// @Param id path string true "Product ID"
// @Success 204 {object} nil
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/products/{id} [delete]
func DeleteProduct(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "delete product endpoint",
	})
}
