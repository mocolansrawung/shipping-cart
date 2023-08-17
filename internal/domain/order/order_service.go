package order

import (
	"errors"
	"fmt"
	"strings"

	"github.com/evermos/boilerplate-go/configs"
	"github.com/evermos/boilerplate-go/internal/domain/cart"
	"github.com/evermos/boilerplate-go/internal/domain/product"
	"github.com/evermos/boilerplate-go/shared/failure"
	"github.com/gofrs/uuid"
)

type OrderService interface {
	Checkout(requestFormat OrderRequestFormat, userID uuid.UUID) (order Order, err error)
}

type OrderServiceImpl struct {
	OrderRepository   OrderRepository
	CartService       cart.CartService
	ProductRepository product.ProductRepository
	Config            *configs.Config
}

func ProvideOrderServiceImpl(orderRepository OrderRepository, cartService cart.CartService, config *configs.Config) *OrderServiceImpl {
	s := new(OrderServiceImpl)
	s.OrderRepository = orderRepository
	s.CartService = cartService
	s.Config = config

	return s
}

func (s *OrderServiceImpl) Checkout(requestFormat OrderRequestFormat, userID uuid.UUID) (order Order, err error) {
	// retrieve cart details
	cart, err := s.CartService.ResolveDetailsByUserID(userID)
	if err != nil {
		return
	}

	// handling empty cart
	if len(cart.Items) == 0 {
		err = errors.New("cart is empty")
		return order, err
	}

	// handling insufficient product stock
	var insufficientStockProducts []string
	for _, item := range cart.Items {
		if item.Stock < item.Quantity {
			insufficientStockProducts = append(insufficientStockProducts, item.ProductID.String())
		}
	}
	if len(insufficientStockProducts) > 0 {
		err = errors.New(fmt.Sprintf("Insufficient stock for products with IDs: %s", strings.Join(insufficientStockProducts, ", ")))
		return order, err
	}

	// must utilized request format
	order, err = order.NewFromRequestFormat(requestFormat, userID)
	if err != nil {
		return
	}

	if err != nil {
		return order, failure.BadRequest(err)
	}

	err = s.OrderRepository.Checkout(order, cart.ID)

	return

	// populate items and totalcost
	// order ID must be created -> fix this later
	// var orderItems []OrderItem
	// var totalCost float64
	// for _, cartItem := range cart.Items {
	// 	orderItem := OrderItem{
	// 		ProductID: cartItem.ProductID,
	// 		UnitPrice: cartItem.UnitPrice,
	// 		Quantity:  cartItem.Quantity,
	// 		Cost:      cartItem.Cost,
	// 		CreatedAt: time.Now(),
	// 		CreatedBy: userID,
	// 	}
	// 	orderItems = append(orderItems, orderItem)
	// 	totalCost += cartItem.Cost
	// }

	// newOrder := Order{
	// 	UserID:    userID,
	// 	TotalCost: totalCost,
	// 	Status:    OrderStatusPending,
	// 	Items:     orderItems,
	// }
}
