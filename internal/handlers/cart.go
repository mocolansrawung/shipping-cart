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
			// r.Use(h.AuthMiddleware.AdminRoleCheck)
			r.Use(h.AuthMiddleware.ValidateAuth)
			r.Post("/", h.AddToCart)
			r.Get("/", h.GetCartByUserID)
		})
	})
}

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

	cartID, err := h.CartService.EnsureCartExists(userID)
	if err != nil {
		response.WithError(w, failure.BadRequest(err))
	}

	requestFormat.CartID = cartID

	err = shared.GetValidator().Struct(requestFormat)
	if err != nil {
		response.WithError(w, failure.BadRequest(err))
		return
	}

	item, err := h.CartService.AddItemToCart(requestFormat, userID)
	if err != nil {
		response.WithError(w, err)
		return
	}

	response.WithJSON(w, http.StatusCreated, item)
}
