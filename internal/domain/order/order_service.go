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
	// 1. fetch the cart using the userID
	// need: GetCartByID
	cart, err := s.CartService.ResolveDetailsByUserID(userID)
	if err != nil {
		return
	}

	// 2. validate the cart -> if the cart exist and not empty
	// -> check product availability
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

	// 3. create an order using relevant detail from cart

	newOrder := Order{
		UserID:    userID,
		TotalCost: 0,
		Status:    "pending",
	}

	orderID, err := s.OrderRepository.CreateOrder(newOrder)
	if err != nil {
		return order, failure.InternalError(err)
	}

	// 4. transfer cart items into the order items

	// 5. remove the checked out cart items from the cart item table entirely

	// return the order details and error
	return
}
