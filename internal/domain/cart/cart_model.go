package cart

import (
	"encoding/json"
	"math"
	"time"

	"github.com/evermos/boilerplate-go/shared"
	"github.com/evermos/boilerplate-go/shared/nuuid"
	"github.com/gofrs/uuid"
	"github.com/guregu/null"
)

// Cart
type Cart struct {
	ID        uuid.UUID   `db:"id" validate:"required"`
	UserID    uuid.UUID   `db:"user_id" valdiate:"required"`
	CreatedAt time.Time   `db:"created_at" validate:"required"`
	CreatedBy uuid.UUID   `db:"created_by" validate:"required"`
	UpdatedAt null.Time   `db:"updated_at"`
	UpdatedBy nuuid.NUUID `db:"updated_by"`
	DeletedAt null.Time   `db:"deleted_at"`
	DeletedBy nuuid.NUUID `db:"deleted_by"`
	Items     []CartItem  `db:"-" validate:"required,dive,required"`
}

func (c *Cart) AttachItems(items []CartItem) Cart {
	for _, item := range items {
		if item.CartID == c.ID {
			c.Items = append(c.Items, item)
		}
	}

	return *c
}

func (c Cart) IsDeleted() (deleted bool) {
	return c.DeletedAt.Valid && c.DeletedBy.Valid
}

func (c Cart) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.ToResponseFormat())
}

func (c Cart) ToResponseFormat() CartResponseFormat {
	resp := CartResponseFormat{
		ID:        c.ID,
		UserID:    c.UserID,
		CreatedAt: c.CreatedAt,
		CreatedBy: c.CreatedBy,
		UpdatedAt: c.UpdatedAt,
		UpdatedBy: c.UpdatedBy.Ptr(),
		DeletedAt: c.DeletedAt,
		DeletedBy: c.DeletedBy.Ptr(),
	}

	for _, item := range c.Items {
		resp.Items = append(resp.Items, item.ToResponseFormat())
	}

	return resp
}

func (c *Cart) Validate() (err error) {
	validator := shared.GetValidator()
	return validator.Struct(c)
}

type CartRequestFormat struct {
	ID    uuid.UUID               `json:"id" validate:"required"`
	Items []CartItemRequestFormat `json:"items" validate:"required,dive,required"`
}

type CartResponseFormat struct {
	ID        uuid.UUID                `json:"id"`
	UserID    uuid.UUID                `json:"userID"`
	CreatedAt time.Time                `json:"createdAt"`
	CreatedBy uuid.UUID                `json:"createdBy"`
	UpdatedAt null.Time                `json:"updatedAt"`
	UpdatedBy *uuid.UUID               `json:"updatedBy"`
	DeletedAt null.Time                `json:"deletedAt,omitempty"`
	DeletedBy *uuid.UUID               `json:"deletedBy,omitempty"`
	Items     []CartItemResponseFormat `json:"items"`
}

// Cart Item
type CartItem struct {
	CartID    uuid.UUID   `db:"cart_id" validate:"required"`
	ProductID uuid.UUID   `db:"product_id" validate:"required"`
	UnitPrice float64     `db:"unit_price" validate:"required"`
	Quantity  int         `db:"quantity" validate:"required,min=1"`
	Cost      float64     `db:"cost" validate:"required,min=0"`
	Stock     int         `db:"stock"`
	CreatedAt time.Time   `db:"created_at" validate:"required"`
	CreatedBy uuid.UUID   `db:"created_by" validate:"required"`
	UpdatedAt null.Time   `db:"updated_at"`
	UpdatedBy nuuid.NUUID `db:"updated_by"`
	DeletedAt null.Time   `db:"deleted_at"`
	DeletedBy nuuid.NUUID `db:"deleted_by"`
}

func (ci CartItem) MarshalJSON() ([]byte, error) {
	return json.Marshal(ci.ToResponseFormat())
}

func (ci CartItem) NewCartItemFromRequestFormat(req CartItemRequestFormat, userID uuid.UUID) (newCartItem CartItem, err error) {
	newCartItem = CartItem{
		CartID:    req.CartID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		CreatedAt: time.Now(),
		CreatedBy: userID,
	}

	return
}

func (ci *CartItem) Validate() (err error) {
	validator := shared.GetValidator()
	return validator.Struct(ci)
}

func (ci *CartItem) Recalculate() {
	ci.Cost = math.Round(float64(ci.Quantity)*ci.UnitPrice*100) / 100
}

func (ci *CartItem) ToResponseFormat() CartItemResponseFormat {
	return CartItemResponseFormat{
		CartID:    ci.CartID,
		ProductID: ci.ProductID,
		UnitPrice: ci.UnitPrice,
		Quantity:  ci.Quantity,
		Cost:      ci.Cost,
		CreatedBy: ci.CreatedBy,
		CreatedAt: ci.CreatedAt,
		UpdatedAt: ci.UpdatedAt,
		UpdatedBy: ci.UpdatedBy.Ptr(),
		DeletedAt: ci.DeletedAt,
		DeletedBy: ci.DeletedBy.Ptr(),
	}
}

type CartItemRequestFormat struct {
	CartID    uuid.UUID `json:"cartID" validate:"required"`
	ProductID uuid.UUID `json:"productID" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required,min=1"`
}

type CartItemResponseFormat struct {
	CartID    uuid.UUID  `json:"cartID"`
	ProductID uuid.UUID  `json:"productID"`
	UnitPrice float64    `json:"unitPrice"`
	Quantity  int        `json:"quantity"`
	Cost      float64    `json:"cost"`
	Stock     int        `json:"-"`
	CreatedAt time.Time  `json:"createdAt"`
	CreatedBy uuid.UUID  `json:"createdBy"`
	UpdatedAt null.Time  `json:"updatedAt"`
	UpdatedBy *uuid.UUID `json:"updatedBy"`
	DeletedAt null.Time  `json:"deletedAt,omitempty"`
	DeletedBy *uuid.UUID `json:"deletedBy,omitempty"`
}
