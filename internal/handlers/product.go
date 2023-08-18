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
			r.Use(h.AuthMiddleware.ValidateAuth)
			r.Get("/", h.ResolveProducts)
			r.Post("/", h.CreateProduct)
		})
	})
}

// ResolveProducts retrieves a list of products with optional filtering and pagination.
// @Summary Retrieve a list of products.
// @Description This endpoint retrieves a list of products with optional filtering and pagination.
// @Tags products
// @Security EVMOauthToken
// @Param page query integer false "Page number for pagination (default 1)"
// @Param limit query integer false "Number of items per page (default 10)"
// @Param sort query string false "Sort field for ordering"
// @Param order query string false "Sort order ('asc' or 'desc')"
// @Param brand query string false "Filter by brand"
// @Param category query string false "Filter by category"
// @Produce json
// @Success 200 {object} response.Base{data=product.ProductResponseFormat}
// @Failure 400 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /v1/products [get]
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

// CreateProduct creates a new product.
// @Summary Create a new product.
// @Description This endpoint creates a new product.
// @Tags products
// @Security EVMOauthToken
// @Param product body product.ProductRequestFormat true "The product to be created."
// @Produce json
// @Success 201 {object} response.Base{data=product.ProductResponseFormat}
// @Failure 400 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /v1/products [post]
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
