package order

import (
	"encoding/json"
	"time"

	"github.com/evermos/boilerplate-go/shared"
	"github.com/evermos/boilerplate-go/shared/nuuid"
	"github.com/gofrs/uuid"
	"github.com/guregu/null"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCanceled   OrderStatus = "canceled"
)

type Order struct {
	ID        uuid.UUID   `db:"order_id" validate:"required"`
	UserID    uuid.UUID   `db:"user_id" validate:"required"`
	TotalCost float64     `db:"total_cost" validate:"required"`
	Status    OrderStatus `db:"status" validate:"required,oneof=pending processing shipped delivered canceled"`
	CreatedAt time.Time   `db:"created_at" validate:"required"`
	CreatedBy uuid.UUID   `db:"created_by" validate:"required"`
	UpdatedAt null.Time   `db:"updated_at"`
	UpdatedBy nuuid.NUUID `db:"updated_by"`
	DeletedAt null.Time   `db:"deleted_at"`
	DeletedBy nuuid.NUUID `db:"deleted_by"`
	Items     []OrderItem `db:"-" validate:"required,dive,required"`
}

func (o *Order) AttachItems(items []OrderItem) Order {
	for _, item := range items {
		if item.OrderID == o.ID {
			o.Items = append(o.Items, item)
		}
	}
	return *o
}

func (o *Order) IsDeleted() (deleted bool) {
	return o.DeletedAt.Valid && o.DeletedBy.Valid
}

func (o Order) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.ToResponseFormat())
}

func (o Order) NewOrderFromRequestFormat(req OrderRequestFormat, userID uuid.UUID) (newOrder Order, err error) {
	orderID, err := uuid.NewV4()
	if err != nil {
		return
	}

	newOrder = Order{
		ID:        orderID,
		UserID:    userID,
		TotalCost: req.TotalCost,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
		CreatedBy: userID,
	}

	items := make([]OrderItem, 0)
	for _, requestItem := range req.Items {
		item := OrderItem{}
		item = item.NewOrderItemFromRequestFormat(requestItem, userID)
		items = append(items, item)
	}
	newOrder.Items = items

	newOrder.Recalculate()
	err = newOrder.Validate()

	return
}

func (o *Order) Recalculate() {
	o.TotalCost = float64(0)
	recalculatedItems := make([]OrderItem, 0)
	for _, item := range o.Items {
		item.Recalculate()
		recalculatedItems = append(recalculatedItems, item)
		o.TotalCost += item.Cost
	}

	o.Items = recalculatedItems
}

func (o *Order) Validate() (err error) {
	validator := shared.GetValidator()
	return validator.Struct(o)
}

func (o Order) ToResponseFormat() OrderResponseFormat {
	resp := OrderResponseFormat{
		ID:        o.ID,
		UserID:    o.UserID,
		TotalCost: o.TotalCost,
		Status:    o.Status,
		CreatedAt: o.CreatedAt,
		CreatedBy: o.CreatedBy,
		UpdatedAt: o.UpdatedAt,
		UpdatedBy: o.UpdatedBy.Ptr(),
		DeletedAt: o.DeletedAt,
		DeletedBy: o.DeletedBy.Ptr(),
		Items:     make([]OrderItemResponseFormat, 0),
	}

	for _, item := range o.Items {
		resp.Items = append(resp.Items, item.ToResponseFormat())
	}

	return resp
}

type OrderRequestFormat struct {
	ID        uuid.UUID   `json:"ID" validate:"required"`
	TotalCost float64     `json:"totalCost" validate:"required"`
	Status    OrderStatus `json:"orderStatus" validate:"required"`
	Items     []OrderItem `json:"items" validate:"required,dive,required"`
}

type OrderResponseFormat struct {
	ID        uuid.UUID   `json:"ID"`
	UserID    uuid.UUID   `json:"userID"`
	TotalCost float64     `json:"totalCost"`
	Status    OrderStatus `json:"status"`
	CreatedAt time.Time   `json:"createdAt"`
	CreatedBy uuid.UUID   `json:"createdBy"`
	UpdatedAt null.Time   `json:"updatedAt"`
	UpdatedBy *uuid.UUID  `json:"updatedBy"`
	DeletedAt null.Time   `json:"deletedAt,omitempty"`
	DeletedBy *uuid.UUID  `json:"deletedBy,omitempty"`
	Items     []OrderItem `json:"items"`
}

// Order Item
type OrderItem struct {
	ID        uuid.UUID  `db:"id" validate:"required"`
	OrderID   uuid.UUID  `db:"order_id" validate:"required"`
	ProductID uuid.UUID  `db:"product_id" validate:"required"`
	UnitPrice uuid.UUID  `db:"unit_price" validate:"required"`
	Quantity  int        `db:"quantity" validate:"required"`
	Cost      float64    `db:"cost" validate:"required"`
	CreatedAt time.Time  `json:"createdAt"`
	CreatedBy uuid.UUID  `json:"createdBy"`
	UpdatedAt null.Time  `json:"updatedAt"`
	UpdatedBy *uuid.UUID `json:"updatedBy"`
	DeletedAt null.Time  `json:"deletedAt,omitempty"`
	DeletedBy *uuid.UUID `json:"deletedBy,omitempty"`
}

func (oi OrderItem) MarshalJSON() ([]byte, error) {
	return json.Marshal(oi.ToResponseFormat())
}

func (oi OrderItem) NewOrderItemFromRequestFormat(req OrderItemRequestFormat, userID uuid.UUID) (newOrderItem OrderItem, err error) {
	newOrderItem = OrderItem{
		OrderID:   req.OrderID,
		ProductID: req.ProductID,
	}

	return
}

type OrderItemRequestFormat struct {
	OrderID   uuid.UUID `json:"order_id" validate:"required"`
	CartID    uuid.UUID `json:"cart_id" validate:"required"`
	ProductID uuid.UUID `json:"product_id" validate:"required"`
}

type OrderItemResponseFormat struct {
	OrderID   uuid.UUID  `json:"order_id"`
	ProductID uuid.UUID  `json:"product_id"`
	UnitPrice uuid.UUID  `json:"unit_price"`
	Quantity  int        `json:"quantity"`
	Cost      float64    `json:"cost"`
	CreatedAt time.Time  `json:"createdAt"`
	CreatedBy uuid.UUID  `json:"createdBy"`
	UpdatedAt null.Time  `json:"updatedAt"`
	UpdatedBy *uuid.UUID `json:"updatedBy"`
	DeletedAt null.Time  `json:"deletedAt,omitempty"`
	DeletedBy *uuid.UUID `json:"deletedBy,omitempty"`
}
