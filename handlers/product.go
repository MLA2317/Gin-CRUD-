package handlers

import (
	"math"
	"net/http"
	"strconv"

	"gin-crud/database"
	"gin-crud/dto"
	"gin-crud/models"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct{}

func NewProductHandler() *ProductHandler {
	return &ProductHandler{}
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	userID := c.GetUint("user_id")

	product := models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		UserID:      userID,
	}

	if err := database.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "failed to create product",
		})
		return
	}

	c.JSON(http.StatusCreated, dto.Response{
		Success: true,
		Message: "product created successfully",
		Data:    product,
	})
}

func (h *ProductHandler) GetAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	var products []models.Product
	var total int64

	query := database.DB.Model(&models.Product{})

	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	query.Count(&total)

	if err := query.Preload("User").Limit(limit).Offset(offset).Order("created_at DESC").Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "failed to fetch products",
		})
		return
	}

	totalPage := int(math.Ceil(float64(total) / float64(limit)))

	c.JSON(http.StatusOK, dto.PaginatedResponse{
		Success: true,
		Message: "products retrieved",
		Data:    products,
		Meta: dto.Meta{
			Page:      page,
			Limit:     limit,
			Total:     total,
			TotalPage: totalPage,
		},
	})
}

func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "invalid product id",
		})
		return
	}

	var product models.Product
	if err := database.DB.Preload("User").First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.Response{
			Success: false,
			Message: "product not found",
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "product retrieved",
		Data:    product,
	})
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "invalid product id",
		})
		return
	}

	userID := c.GetUint("user_id")

	var product models.Product
	if err := database.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.Response{
			Success: false,
			Message: "product not found",
		})
		return
	}

	// Only the owner can update
	if product.UserID != userID {
		c.JSON(http.StatusForbidden, dto.Response{
			Success: false,
			Message: "you can only update your own products",
		})
		return
	}

	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.Stock != nil {
		updates["stock"] = *req.Stock
	}

	if err := database.DB.Model(&product).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "failed to update product",
		})
		return
	}

	database.DB.Preload("User").First(&product, id)

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "product updated successfully",
		Data:    product,
	})
}

func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "invalid product id",
		})
		return
	}

	userID := c.GetUint("user_id")

	var product models.Product
	if err := database.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.Response{
			Success: false,
			Message: "product not found",
		})
		return
	}

	// Only the owner can delete
	if product.UserID != userID {
		c.JSON(http.StatusForbidden, dto.Response{
			Success: false,
			Message: "you can only delete your own products",
		})
		return
	}

	if err := database.DB.Delete(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "failed to delete product",
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "product deleted successfully",
	})
}
