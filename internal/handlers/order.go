package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/evermos/boilerplate-go/internal/domain/order"
	"github.com/evermos/boilerplate-go/shared"
	"github.com/evermos/boilerplate-go/shared/failure"
	"github.com/evermos/boilerplate-go/transport/http/middleware"
	"github.com/evermos/boilerplate-go/transport/http/response"
	"github.com/go-chi/chi"
)

type OrderHandler struct {
	OrderService   order.OrderService
	AuthMiddleware *middleware.Authentication
}

func ProvideOrderHandler(orderService order.OrderService, authMiddleware *middleware.Authentication) OrderHandler {
	return OrderHandler{
		OrderService:   orderService,
		AuthMiddleware: authMiddleware,
	}
}

func (h *OrderHandler) Router(r chi.Router) {
	r.Route("/orders", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(h.AuthMiddleware.ValidateAuth)
			r.Post("/checkout", h.CheckoutOrder)
		})
	})
}

func (h *OrderHandler) CheckoutOrder(w http.ResponseWriter, r *http.Request) {
	// extract the variables needed
	// from context
	claims, ok := r.Context().Value("claims").(shared.Claims)
	if !ok {
		response.WithError(w, failure.Unauthorized("Login needed"))
		return
	}
	userID := claims.UserID

	// from request body
	decoder := json.NewDecoder(r.Body)
	var requestFormat order.OrderRequestFormat
	err := decoder.Decode(&requestFormat)
	if err != nil {
		response.WithError(w, failure.BadRequest(err))
		return
	}

	// validator
	err = shared.GetValidator().Struct(requestFormat)
	if err != nil {
		response.WithError(w, failure.BadRequest(err))
		return
	}

	// call the service layer to create the order
	order, err := h.OrderService.Checkout(requestFormat, userID)
	if err != nil {
		response.WithError(w, err)
		return
	}

	response.WithJSON(w, http.StatusCreated, order)
}
