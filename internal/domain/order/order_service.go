package order

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/evermos/boilerplate-go/configs"
	"github.com/evermos/boilerplate-go/internal/domain/cart"
	"github.com/evermos/boilerplate-go/internal/domain/product"
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
	cart, err := s.CartService.ResolveDetailsByUserID(userID)
	if err != nil {
		return
	}

	if len(cart.Items) == 0 {
		err = errors.New("cart is empty")
		return order, err
	}

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

	var orderItems []OrderItem
	for _, cartItem := range cart.Items {
		orderItem := OrderItem{
			ProductID: cartItem.ProductID,
			UnitPrice: cartItem.UnitPrice,
			Quantity:  cartItem.Quantity,
			Cost:      cartItem.Cost,
			CreatedAt: time.Now(),
			CreatedBy: userID,
		}
		orderItems = append(orderItems, orderItem)
	}

	newOrder := Order{
		UserID:    userID,
		TotalCost: 0,
		Status:    OrderStatusPending,
		Items:     orderItems,
	}

	// call repo service to implement checkout
	// fix later
	err = s.OrderRepository.Checkout(newOrder, cart.ID)

	return
}
