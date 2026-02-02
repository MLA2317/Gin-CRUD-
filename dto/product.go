package dto

type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required,min=2"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Stock       int     `json:"stock" binding:"gte=0"`
}

type UpdateProductRequest struct {
	Name        *string  `json:"name" binding:"omitempty,min=2"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price" binding:"omitempty,gt=0"`
	Stock       *int     `json:"stock" binding:"omitempty,gte=0"`
}
