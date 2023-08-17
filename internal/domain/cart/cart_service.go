package cart

import (
	"github.com/evermos/boilerplate-go/configs"
	"github.com/evermos/boilerplate-go/internal/domain/product"
	"github.com/evermos/boilerplate-go/shared/failure"
	"github.com/evermos/boilerplate-go/shared/logger"
	"github.com/gofrs/uuid"
)

type CartService interface {
	AddToCart(requestFormat CartItemRequestFormat, userID uuid.UUID) (cart Cart, err error)
	ResolveByUserID(id uuid.UUID) (cart Cart, err error)
}

type CartServiceImpl struct {
	CartRepository CartRepository
	ProductService product.ProductService
	Config         *configs.Config
}

func ProvideCartServiceImpl(cartRepository CartRepository, productService product.ProductService, config *configs.Config) *CartServiceImpl {
	s := new(CartServiceImpl)
	s.CartRepository = cartRepository
	s.ProductService = productService
	s.Config = config

	return s
}

func (s *CartServiceImpl) AddToCart(requestFormat CartItemRequestFormat, userID uuid.UUID) (cartItem CartItem, err error) {
	product, err := s.ProductService.ResolveByID(requestFormat.ProductID)
	if err != nil {
		return cartItem, failure.InternalError(err)
	}

	if product.Stock < requestFormat.Quantity {
		logger.ErrorWithStack(err)
		return cartItem, failure.BadRequestFromString("Not enough stock")
	}

	cart, err := s.CartRepository.ResolveOrCreateCartByUserID(userID)
	if err != nil {
		logger.ErrorWithStack(err)
		return cartItem, failure.InternalError(err)
	}

	cartItem, err = cartItem.NewFromRequestFormat(requestFormat, userID, cart.ID, product.Price)
	if err != nil {
		return cartItem, failure.InternalError(err)
	}

	existingCartItem, found, err := s.CartRepository.ResolveCartItemByProductID(cart.ID, product.ID)
	if err != nil {
		logger.ErrorWithStack(err)
		return cartItem, failure.InternalError(err)
	}

	if found {
		existingCartItem.Quantity += requestFormat.Quantity
		err = s.CartRepository.UpdateItemQuantity(existingCartItem)
		if err != nil {
			logger.ErrorWithStack(err)
			return cartItem, failure.InternalError(err)
		}
	} else {
		err = s.CartRepository.CreateCartItem(cartItem, userID)
		if err != nil {
			logger.ErrorWithStack(err)
			return cartItem, failure.InternalError(err)
		}
	}

	return cartItem, nil
}
func (s *CartServiceImpl) ResolveByUserID(id uuid.UUID) (cart Cart, err error) {
	cart, err = s.CartRepository.ResolveCartByUserID(id)
	if err != nil {
		return
	}

	if cart.IsDeleted() {
		return cart, failure.NotFound("cart")
	}

	var items []CartItem
	items, err = s.CartRepository.ResolveItemsByCartID([]uuid.UUID{cart.ID})
	if err != nil {
		return cart, err
	}

	cart.AttachItems(items)

	return
}
