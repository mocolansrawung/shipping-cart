package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/evermos/boilerplate-go/internal/domain/product"
	"github.com/evermos/boilerplate-go/shared"
	"github.com/evermos/boilerplate-go/shared/failure"
	"github.com/evermos/boilerplate-go/transport/http/middleware"
	"github.com/evermos/boilerplate-go/transport/http/response"
	"github.com/go-chi/chi"
)

type ProductHandler struct {
	ProductService product.ProductService
	AuthMiddleware *middleware.Authentication
}

func ProvideProductHandler(productService product.ProductService, authMiddleware *middleware.Authentication) ProductHandler {
	return ProductHandler{
		ProductService: productService,
		AuthMiddleware: authMiddleware,
	}
}

func (h *ProductHandler) Router(r chi.Router) {
	r.Route("/products", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			// r.Use(h.AuthMiddleware.UserRoleCheck)
			r.Use(h.AuthMiddleware.ValidateAuth)
			r.Get("/", h.ResolveProducts)
			r.Post("/", h.CreateProduct)
		})
	})
}

func (h *ProductHandler) ResolveProducts(w http.ResponseWriter, r *http.Request) {
	pageString := r.URL.Query().Get("page")
	page, err := shared.ConvertQueryParamsToInt(pageString)
	if err != nil || page < 0 {
		page = 1
	}

	limitString := r.URL.Query().Get("limit")
	limit, err := shared.ConvertQueryParamsToInt(limitString)
	if err != nil || limit <= 0 {
		limit = 10
	}

	params := product.ProductQueryParams{
		Page:     page,
		Limit:    limit,
		Sort:     r.URL.Query().Get("sort"),
		Order:    r.URL.Query().Get("order"),
		Brand:    r.URL.Query().Get("brand"),
		Category: r.URL.Query().Get("category"),
	}

	products, total, err := h.ProductService.GetProducts(params)
	if err != nil {
		response.WithError(w, err)
		return
	}

	resp, err := h.ProductService.CreatePaginationResponse(products, total, limit, page)

	response.WithJSON(w, http.StatusOK, resp)
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var requestFormat product.ProductRequestFormat
	err := decoder.Decode(&requestFormat)
	if err != nil {
		response.WithError(w, failure.BadRequest(err))
	}

	err = shared.GetValidator().Struct(requestFormat)
	if err != nil {
		response.WithError(w, failure.BadRequest(err))
		return
	}

	claims, ok := r.Context().Value("claims").(shared.Claims)
	if !ok {
		response.WithError(w, failure.Unauthorized("User not authorized"))
	}

	product, err := h.ProductService.CreateProduct(requestFormat, claims.UserID)
	if err != nil {
		response.WithError(w, err)
		return
	}

	response.WithJSON(w, http.StatusCreated, product)
}
