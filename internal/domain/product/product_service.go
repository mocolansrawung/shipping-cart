package product

import (
	"math"

	"github.com/evermos/boilerplate-go/configs"
	"github.com/evermos/boilerplate-go/shared/failure"
	"github.com/gofrs/uuid"
)

type ProductService interface {
	CreateProduct(requestFormat ProductRequestFormat, userID uuid.UUID) (product Product, err error)
	GetProducts(params ProductQueryParams) (products []Product, total int, err error)
	CreatePaginationResponse(products []Product, total int, limit int, page int) (productPagination ProductPagination, err error)
	ResolveByID(id uuid.UUID) (product Product, err error)
}

type ProductServiceImpl struct {
	ProductRepository ProductRepository
	Config            *configs.Config
}

func ProvideProductServiceImpl(productRepository ProductRepository, config *configs.Config) *ProductServiceImpl {
	s := new(ProductServiceImpl)
	s.ProductRepository = productRepository
	s.Config = config

	return s
}

func (s *ProductServiceImpl) CreateProduct(requestFormat ProductRequestFormat, userID uuid.UUID) (product Product, err error) {
	product, err = product.NewProductFromRequestFormat(requestFormat, userID)
	if err != nil {
		return
	}

	if err != nil {
		return product, failure.BadRequest(err)
	}

	err = s.ProductRepository.CreateProduct(product)
	if err != nil {
		return
	}

	return
}

func (s *ProductServiceImpl) GetProducts(params ProductQueryParams) (products []Product, total int, err error) {
	products, err = s.ProductRepository.ResolveProductsByQuery(params)
	if err != nil {
		return
	}

	total, err = s.ProductRepository.CountAllProducts()
	if err != nil {
		return
	}

	return
}
func (s *ProductServiceImpl) CreatePaginationResponse(products []Product, total int, limit int, page int) (productPagination ProductPagination, err error) {
	productPagination = ProductPagination{
		Data:        products,
		Total:       total,
		PerPage:     limit,
		CurrentPage: page,
		TotalPages:  int(math.Ceil(float64(total) / float64(limit))),
	}

	return
}
func (s *ProductServiceImpl) ResolveByID(id uuid.UUID) (product Product, err error) {
	product, err = s.ProductRepository.ResolveProductByID(id)
	if product.IsDeleted() {
		return product, failure.NotFound("product")
	}

	return
}
