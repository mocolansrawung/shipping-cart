package order

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/evermos/boilerplate-go/shared"
	"github.com/evermos/boilerplate-go/shared/failure"
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
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
		CreatedBy: userID,
	}

	items := make([]OrderItem, 0)
	for _, requestItem := range req.Items {
		item := OrderItem{}
		item, err = item.NewFromRequestFormat(requestItem, orderID)
		if err != nil {
			return
		}
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
func (o *Order) UpdateStatus(newStatus OrderStatus) (err error) {
	stateChangeNotAllowedError := failure.Conflict(
		"stateChange",
		"order",
		fmt.Sprintf("cannot change from %s to %s", o.Status, newStatus))

	switch o.Status {
	case OrderStatusPending:
		if newStatus != OrderStatusProcessing {
			return stateChangeNotAllowedError
		}
	case OrderStatusProcessing:
		if newStatus != OrderStatusShipped {
			return stateChangeNotAllowedError
		}
	case OrderStatusShipped:
		if newStatus != OrderStatusDelivered {
			return stateChangeNotAllowedError
		}
	case OrderStatusDelivered, OrderStatusCanceled:
		return stateChangeNotAllowedError
	}

	o.Status = newStatus

	return nil
}
func (o *Order) Validate() (err error) {
	validator := shared.GetValidator()
	return validator.Struct(o)
}

type OrderRequestFormat struct {
	TotalCost float64                  `json:"totalCost" validate:"required"`
	Status    OrderStatus              `json:"orderStatus" validate:"required"`
	Items     []OrderItemRequestFormat `json:"items" validate:"required,dive,required"`
}

type OrderResponseFormat struct {
	ID        uuid.UUID                 `json:"ID"`
	UserID    uuid.UUID                 `json:"userID"`
	TotalCost float64                   `json:"totalCost"`
	Status    OrderStatus               `json:"status"`
	CreatedAt time.Time                 `json:"createdAt"`
	CreatedBy uuid.UUID                 `json:"createdBy"`
	UpdatedAt null.Time                 `json:"updatedAt"`
	UpdatedBy *uuid.UUID                `json:"updatedBy"`
	DeletedAt null.Time                 `json:"deletedAt,omitempty"`
	DeletedBy *uuid.UUID                `json:"deletedBy,omitempty"`
	Items     []OrderItemResponseFormat `json:"items"`
}

// Order Item
type OrderItem struct {
	OrderID   uuid.UUID   `db:"order_id" validate:"required"`
	ProductID uuid.UUID   `db:"product_id" validate:"required"`
	Quantity  int         `db:"quantity" validate:"required,min=1"`
	UnitPrice float64     `db:"unit_price" validate:"required"`
	Cost      float64     `db:"cost" validate:"required,min=0"`
	CreatedAt time.Time   `db:"created_at"`
	CreatedBy uuid.UUID   `db:"created_by"`
	UpdatedAt null.Time   `db:"deleted_at"`
	UpdatedBy nuuid.NUUID `db:"updated_by"`
	DeletedAt null.Time   `db:"deleted_at"`
	DeletedBy nuuid.NUUID `db:"deleted_by"`
}

func (oi OrderItem) MarshalJSON() ([]byte, error) {
	return json.Marshal(oi.ToResponseFormat())
}

func (oi OrderItem) NewFromRequestFormat(format OrderItemRequestFormat, orderID uuid.UUID) (newOrderItem OrderItem, err error) {
	newOrderItem = OrderItem{
		OrderID:   orderID,
		ProductID: format.ProductID,
	}

	return
}

func (oi *OrderItem) Recalculate() {
	oi.Cost = float64(oi.Quantity) * oi.UnitPrice
}

func (oi *OrderItem) ToResponseFormat() OrderItemResponseFormat {
	return OrderItemResponseFormat{
		OrderID:   oi.OrderID,
		ProductID: oi.ProductID,
		Quantity:  oi.Quantity,
		UnitPrice: oi.UnitPrice,
		Cost:      oi.Cost,
		CreatedAt: oi.CreatedAt,
		CreatedBy: oi.CreatedBy,
		UpdatedAt: oi.UpdatedAt,
		UpdatedBy: oi.UpdatedBy.Ptr(),
		DeletedAt: oi.DeletedAt,
		DeletedBy: oi.DeletedBy.Ptr(),
	}
}

type OrderItemRequestFormat struct {
	CartID    uuid.UUID `json:"cart_id" validate:"required"`
	ProductID uuid.UUID `json:"product_id" validate:"required"`
}

type OrderItemResponseFormat struct {
	OrderID   uuid.UUID  `json:"order_id"`
	ProductID uuid.UUID  `json:"product_id"`
	Quantity  int        `json:"quantity"`
	UnitPrice float64    `json:"unit_price"`
	Cost      float64    `json:"cost"`
	CreatedAt time.Time  `json:"createdAt"`
	CreatedBy uuid.UUID  `json:"createdBy"`
	UpdatedAt null.Time  `json:"updatedAt"`
	UpdatedBy *uuid.UUID `json:"updatedBy"`
	DeletedAt null.Time  `json:"deletedAt,omitempty"`
	DeletedBy *uuid.UUID `json:"deletedBy,omitempty"`
}
