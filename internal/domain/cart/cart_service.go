package cart

import (
	"time"

	"github.com/evermos/boilerplate-go/configs"
	"github.com/evermos/boilerplate-go/shared/failure"
	"github.com/gofrs/uuid"
)

type CartService interface {
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

func (s *CartServiceImpl) AddItemToCart(requestFormat CartItemRequestFormat, userID uuid.UUID) (cartItem CartItem, err error) {
	// 1. Check if cart exists for the user.
	exists, err := s.CartRepository.ExistsByUserID(userID)
	if err != nil {
		return cartItem, err
	}

	// 2. If the cart doesn't exist for this user, create one.
	if !exists {
		newCart := Cart{
			UserID:    userID,
			CreatedAt: time.Now(),
			CreatedBy: userID, // Assuming the user creating is the user itself. Change as needed.
		}

		err = s.CartRepository.CreateCart(newCart)
		if err != nil {
			return cartItem, err
		}
	}

	// 3. Proceed with adding the item to the cart.
	cartItem, err = cartItem.NewCartItemFromRequestFormat(requestFormat, userID)
	if err != nil {
		return cartItem, failure.BadRequest(err)
	}

	err = s.CartRepository.AddItemToCart(cartItem, userID)
	if err != nil {
		return cartItem, err
	}

	// Implement later:
	// Handle any business logic, e.g., checking if the item is already in the cart and updating the quantity.

	return cartItem, nil
}

// func (s *CartServiceImpl) AddItemToCart(requestFormat CartItemRequestFormat, userID uuid.UUID) (cartItem CartItem, err error) {
// 	// Call the repository's AddItemToCart method.
// 	cartItem, err = cartItem.NewCartItemFromRequestFormat(requestFormat, userID)
// 	if err != nil {
// 		return
// 	}

// 	if err != nil {
// 		return cartItem, failure.BadRequest(err)
// 	}

// 	err = s.CartRepository.AddItemToCart(cartItem, userID)
// 	if err != nil {
// 		return
// 	}

// 	// Implement later
// 	// Handle any business logic, e.g., checking if the item is already in the cart and updating the quantity.

// 	return
// }

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
