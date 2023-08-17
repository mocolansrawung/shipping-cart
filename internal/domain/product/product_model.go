package product

import (
	"encoding/json"
	"time"

	"github.com/evermos/boilerplate-go/shared"
	"github.com/evermos/boilerplate-go/shared/nuuid"
	"github.com/gofrs/uuid"
	"github.com/guregu/null"
)

type Product struct {
	ID        uuid.UUID   `db:"id" validate:"required"`
	UserID    uuid.UUID   `db:"user_id" validate:"required"`
	Name      string      `db:"name" validate:"required"`
	Price     float64     `db:"price" validate:"required,min=0"`
	Brand     string      `db:"brand" validate:"required"`
	Category  string      `db:"category" validate:"required"`
	Stock     int         `db:"stock" validate:"required,min=0"`
	CreatedAt time.Time   `db:"created_at" validate:"required"`
	CreatedBy uuid.UUID   `db:"created_by" validate:"required"`
	UpdatedAt null.Time   `db:"updated_at"`
	UpdatedBy nuuid.NUUID `db:"updated_by"`
	DeletedAt null.Time   `db:"deleted_at"`
	DeletedBy nuuid.NUUID `db:"deleted_by"`
}

type ProductPagination struct {
	Data        []Product `json:"data"`
	Total       int       `json:"total"`
	PerPage     int       `json:"perPage"`
	CurrentPage int       `json:"currentPage"`
	TotalPages  int       `json:"totalPages"`
}

type ProductQueryParams struct {
	Page     int
	Limit    int
	Sort     string
	Order    string
	Brand    string
	Category string
}

func (p *Product) IsDeleted() (deleted bool) {
	return p.DeletedAt.Valid && p.DeletedBy.Valid
}
func (p Product) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.ToResponseFormat())
}
func (p Product) NewProductFromRequestFormat(req ProductRequestFormat, userID uuid.UUID) (newProduct Product, err error) {
	productID, err := uuid.NewV4()

	newProduct = Product{
		ID:        productID,
		UserID:    userID,
		Name:      req.Name,
		Price:     req.Price,
		Brand:     req.Brand,
		Category:  req.Category,
		Stock:     req.Stock,
		CreatedAt: time.Now(),
		CreatedBy: userID,
	}

	err = newProduct.Validate()

	return
}
func (p *Product) Validate() (err error) {
	validator := shared.GetValidator()
	return validator.Struct(p)
}
func (p Product) ToResponseFormat() ProductResponseFormat {
	return ProductResponseFormat{
		ID:        p.ID,
		UserID:    p.UserID,
		Name:      p.Name,
		Price:     p.Price,
		Brand:     p.Brand,
		Category:  p.Category,
		Stock:     p.Stock,
		CreatedBy: p.CreatedBy,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		UpdatedBy: p.UpdatedBy.Ptr(),
		DeletedAt: p.DeletedAt,
		DeletedBy: p.DeletedBy.Ptr(),
	}
}

type ProductRequestFormat struct {
	Name     string  `json:"name" validate:"required"`
	Price    float64 `json:"price" validate:"required"`
	Brand    string  `json:"brand" validate:"required"`
	Category string  `json:"category" validate:"required"`
	Stock    int     `json:"stock" validate:"required"`
}

type ProductResponseFormat struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"userID"`
	Name      string     `json:"name"`
	Price     float64    `json:"price"`
	Brand     string     `json:"brand"`
	Category  string     `json:"category"`
	Stock     int        `json:"stock"`
	CreatedAt time.Time  `json:"createdAt"`
	CreatedBy uuid.UUID  `json:"createdBy"`
	UpdatedAt null.Time  `json:"updatedAt"`
	UpdatedBy *uuid.UUID `json:"updatedBy"`
	DeletedAt null.Time  `json:"deletedAt,omitempty"`
	DeletedBy *uuid.UUID `json:"deletedBy,omitempty"`
}
