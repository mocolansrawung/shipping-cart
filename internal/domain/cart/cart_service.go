package cart

import (
	"errors"
	"time"

	"github.com/evermos/boilerplate-go/configs"
	"github.com/evermos/boilerplate-go/shared/failure"
	"github.com/evermos/boilerplate-go/shared/logger"
	"github.com/gofrs/uuid"
)

type CartService interface {
	EnsureCartExists(userID uuid.UUID) (cartID uuid.UUID, err error)
	AddItemToCart(requestFormat CartItemRequestFormat, userID uuid.UUID) (cartItem CartItem, err error)
	ResolveByUserID(id uuid.UUID, withItems bool) (cart Cart, err error)
}

type CartServiceImpl struct {
	CartRepository CartRepository
	Config         *configs.Config
}

func ProvideCartServiceImpl(cartRepository CartRepository, config *configs.Config) *CartServiceImpl {
	s := new(CartServiceImpl)
	s.CartRepository = cartRepository
	s.Config = config

	return s
}

func (s *CartServiceImpl) EnsureCartExists(userID uuid.UUID) (cartID uuid.UUID, err error) {
	exists, err := s.CartRepository.ExistsByUserID(userID)
	if err != nil {
		// logger.Error("Failed to check if cart exists:", err)
		return
	}

	if !exists {
		newCart := Cart{
			UserID:    userID,
			CreatedAt: time.Now(),
			CreatedBy: userID,
		}
		err = s.CartRepository.CreateCart(newCart)
		if err != nil {
			return
		}
	}

	cart, err := s.CartRepository.ResolveByUserID(userID)
	if err != nil {
		return
	}

	cartID = cart.ID

	return
}

func (s *CartServiceImpl) AddItemToCart(requestFormat CartItemRequestFormat, userID uuid.UUID) (cartItem CartItem, err error) {
	cartItem, err = cartItem.NewCartItemFromRequestFormat(requestFormat, userID)
	if err != nil {
		logger.ErrorWithStack(err)
		return cartItem, failure.BadRequest(err)
	}

	cartItem.CartID = requestFormat.CartID

	price, stock, err := s.CartRepository.GetPriceAndStockByProductID(cartItem.ProductID)
	if err != nil || stock < cartItem.Quantity {
		logger.ErrorWithStack(err)
		return cartItem, failure.BadRequest(errors.New("insufficient stock"))
	}
	cartItem.UnitPrice = price
	cartItem.Recalculate()

	currentQuantity, err := s.CartRepository.GetCurrentItemQuantity(cartItem)
	if err != nil {
		return cartItem, failure.InternalError(err)
	}

	if currentQuantity > 0 {
		updatedQuantity := currentQuantity + cartItem.Quantity
		cartItem.Quantity = updatedQuantity

		cartItem.Recalculate()

		err = s.CartRepository.UpdateItemQuantity(cartItem)
		if err != nil {
			return cartItem, failure.InternalError(err)
		}

	} else {
		err := s.CartRepository.AddItemToCart(cartItem, userID)
		if err != nil {
			return cartItem, failure.InternalError(err)
		}
	}

	return
}

func (s *CartServiceImpl) ResolveByUserID(id uuid.UUID, withItems bool) (cart Cart, err error) {
	cart, err = s.CartRepository.ResolveByUserID(id)
	if err != nil {
		return
	}

	if cart.IsDeleted() {
		return cart, failure.NotFound("cart")
	}

	if withItems {
		var items []CartItem
		items, err = s.CartRepository.ResolveItemsByCartID([]uuid.UUID{cart.ID})
		if err != nil {
			return cart, err
		}

		cart.AttachItems(items)
	}

	return
}
