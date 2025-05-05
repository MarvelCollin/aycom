package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Product represents a product in the system
type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	ImageURL    string  `json:"image_url"`
	Stock       int     `json:"stock" binding:"required,gte=0"`
	CategoryID  string  `json:"category_id"`
	CreatedAt   string  `json:"created_at,omitempty"`
	UpdatedAt   string  `json:"updated_at,omitempty"`
}

// ProductsResponse represents a response containing multiple products
type ProductsResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Products []Product `json:"products"`
		Total    int       `json:"total"`
	} `json:"data"`
}

// ProductResponse represents a response containing a single product
type ProductResponse struct {
	Success bool    `json:"success"`
	Data    Product `json:"data"`
}

// ProductCreateResponse represents a response when creating a product
type ProductCreateResponse struct {
	Success bool    `json:"success"`
	Data    Product `json:"data"`
	Message string  `json:"message"`
}

// DeleteResponse represents a generic success response
type DeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ProductHandlers contains all product-related handlers

// @Summary List products
// @Description Returns a list of products
// @Tags Products
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param category query string false "Filter by category"
// @Success 200 {object} ProductsResponse
// @Router /api/v1/products [get]
func ListProducts(c *gin.Context) {
	// In a real implementation, you would call the product service via gRPC
	// For now, return sample data until the service is implemented
	products := []Product{
		{
			ID:          uuid.New().String(),
			Name:        "Smart Watch",
			Description: "Latest model smart watch with health tracking features",
			Price:       199.99,
			ImageURL:    "https://example.com/images/watch.jpg",
			Stock:       50,
			CategoryID:  "electronics",
			CreatedAt:   time.Now().Format(time.RFC3339),
		},
		{
			ID:          uuid.New().String(),
			Name:        "Wireless Earbuds",
			Description: "Noise cancelling wireless earbuds",
			Price:       129.99,
			ImageURL:    "https://example.com/images/earbuds.jpg",
			Stock:       120,
			CategoryID:  "electronics",
			CreatedAt:   time.Now().Format(time.RFC3339),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"products": products,
			"total":    len(products),
		},
	})
}

// @Summary Get product
// @Description Returns a product by ID
// @Tags Products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} ProductResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/products/{id} [get]
func GetProduct(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Product ID is required",
			Code:    "INVALID_PRODUCT_ID",
		})
		return
	}

	// In a real implementation, you would call the product service via gRPC
	// For now, return sample data until the service is implemented
	product := Product{
		ID:          productID,
		Name:        "Smart Watch",
		Description: "Latest model smart watch with health tracking features",
		Price:       199.99,
		ImageURL:    "https://example.com/images/watch.jpg",
		Stock:       50,
		CategoryID:  "electronics",
		CreatedAt:   time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    product,
	})
}

// @Summary Create product
// @Description Creates a new product
// @Tags Products
// @Accept json
// @Produce json
// @Param product body Product true "Product information"
// @Success 201 {object} ProductCreateResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/products [post]
func CreateProduct(c *gin.Context) {
	var product Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid product data: " + err.Error(),
			Code:    "INVALID_PRODUCT_DATA",
		})
		return
	}

	// Check required fields
	if product.Name == "" || product.Price <= 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Name and price are required",
			Code:    "MISSING_REQUIRED_FIELDS",
		})
		return
	}

	// Generate new ID
	product.ID = uuid.New().String()
	product.CreatedAt = time.Now().Format(time.RFC3339)
	product.UpdatedAt = product.CreatedAt

	// In a real implementation, you would call the product service via gRPC
	// For now, return the product as if it was created

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    product,
		"message": "Product created successfully",
	})
}

// @Summary Update product
// @Description Updates a product by ID
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param product body Product true "Product information"
// @Success 200 {object} ProductCreateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/products/{id} [put]
func UpdateProduct(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Product ID is required",
			Code:    "INVALID_PRODUCT_ID",
		})
		return
	}

	var product Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid product data: " + err.Error(),
			Code:    "INVALID_PRODUCT_DATA",
		})
		return
	}

	// Set the ID from the URL parameter
	product.ID = productID
	product.UpdatedAt = time.Now().Format(time.RFC3339)

	// In a real implementation, you would call the product service via gRPC
	// For now, return the product as if it was updated

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    product,
		"message": "Product updated successfully",
	})
}

// @Summary Delete product
// @Description Deletes a product by ID
// @Tags Products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} DeleteResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/products/{id} [delete]
func DeleteProduct(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Product ID is required",
			Code:    "INVALID_PRODUCT_ID",
		})
		return
	}

	// In a real implementation, you would call the product service via gRPC
	// For now, return success as if it was deleted

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Product deleted successfully",
	})
}
