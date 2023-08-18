package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/evermos/boilerplate-go/internal/domain/cart"
	"github.com/evermos/boilerplate-go/shared"
	"github.com/evermos/boilerplate-go/shared/failure"
	"github.com/evermos/boilerplate-go/transport/http/middleware"
	"github.com/evermos/boilerplate-go/transport/http/response"
	"github.com/go-chi/chi"
)

type CartHandler struct {
	CartService    cart.CartService
	AuthMiddleware *middleware.Authentication
}

func ProvideCartHandler(cartService cart.CartService, authMiddleware *middleware.Authentication) CartHandler {
	return CartHandler{
		CartService:    cartService,
		AuthMiddleware: authMiddleware,
	}
}

func (h *CartHandler) Router(r chi.Router) {
	r.Route("/carts", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(h.AuthMiddleware.ValidateAuth)
			r.Post("/", h.AddToCart)
			r.Get("/", h.GetCartByUserID)
		})
	})
}

// GetCartByUserID retrieves the cart for the current user.
// @Summary Retrieve the cart for the current user.
// @Description This endpoint retrieves the cart for the current authenticated user.
// @Tags cart
// @Security JWTAuth
// @Produce json
// @Success 200 {object} response.Base{data=cart.CartResponseFormat}
// @Failure 401 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /v1/carts [get]
func (h *CartHandler) GetCartByUserID(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(shared.Claims)
	if !ok {
		response.WithError(w, failure.Unauthorized("Login needed"))
	}
	userID := claims.UserID

	cart, err := h.CartService.ResolveByUserID(userID)
	if err != nil {
		response.WithError(w, err)
		return
	}

	response.WithJSON(w, http.StatusOK, cart)
}

// AddToCart adds an item to the user's cart.
// @Summary Add an item to the cart.
// @Description This endpoint adds an item to the cart of the current authenticated user.
// @Tags cart
// @Security EVMOauthToken
// @Param item body cart.CartItemRequestFormat true "The item to be added to the cart."
// @Produce json
// @Success 201 {object} response.Base{data=cart.CartItemResponseFormat}
// @Failure 400 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /v1/carts [post]
func (h *CartHandler) AddToCart(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(shared.Claims)
	if !ok {
		response.WithError(w, failure.Unauthorized("Login needed"))
		return
	}
	userID := claims.UserID

	decoder := json.NewDecoder(r.Body)
	var requestFormat cart.CartItemRequestFormat
	err := decoder.Decode(&requestFormat)
	if err != nil {
		response.WithError(w, failure.BadRequest(err))
		return
	}

	err = shared.GetValidator().Struct(requestFormat)
	if err != nil {
		response.WithError(w, failure.BadRequest(err))
		return
	}

	item, err := h.CartService.AddToCart(requestFormat, userID)
	if err != nil {
		response.WithError(w, failure.InternalError(err))
		return
	}

	response.WithJSON(w, http.StatusCreated, item)
}
