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
	cart, err := s.CartService.ResolveDetailsByUserID(userID)
	if err != nil {
		return
	}

	if len(cart.Items) == 0 {
		err = errors.New("cart is empty")
		return order, err
	}

	var insufficientStockProducts []string
	for _, reqItem := range requestFormat.Items {
		var foundInCart bool
		for _, cartItem := range cart.Items {
			if reqItem.CartItemID == cartItem.ID {
				foundInCart = true
				if cartItem.Stock < cartItem.Quantity {
					insufficientStockProducts = append(insufficientStockProducts, cartItem.ProductID.String())
				}
				break
			}
		}
		if !foundInCart {
			return order, failure.BadRequest(errors.New("requested item not found in cart"))
		}
	}

	if len(insufficientStockProducts) > 0 {
		return order, errors.New(fmt.Sprintf("Insufficient stock for products with IDs: %s", strings.Join(insufficientStockProducts, ", ")))
	}

	order, err = order.NewFromRequestFormat(requestFormat, userID)
	if err != nil {
		return order, failure.BadRequest(err)
	}

	err = s.OrderRepository.Checkout(order, cart.ID)
	if err != nil {
		return
	}

	return
}
