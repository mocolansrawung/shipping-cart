package product

import (
	"database/sql"

	"github.com/evermos/boilerplate-go/infras"
	"github.com/evermos/boilerplate-go/shared/failure"
	"github.com/evermos/boilerplate-go/shared/logger"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	productQueries = struct {
		selectProducts string
		insertProduct  string
	}{
		selectProducts: `
			SELECT
				id,
				user_id,
				name,
				price,
				brand,
				category,
				stock,
				created_at,
				created_by,
				updated_at,
				updated_by,
				deleted_at,
				deleted_by
			FROM product
			WHERE 1=1
		`,

		insertProduct: `
			INSERT INTO product (
				id,
				user_id,
				name,
				price,
				brand,
				category,
				stock,
				created_at,
				created_by,
				updated_at,
				updated_by,
				deleted_at,
				deleted_by
			) VALUES (
				:id,
				:user_id,
				:name,
				:price,
				:brand,
				:category,
				:stock,
				:created_at,
				:created_by,
				:updated_at,
				:updated_by,
				:deleted_at,
				:deleted_by
			)
		`,
	}
)

type ProductRepository interface {
	CreateProduct(product Product) (err error)
	ResolveProductsByQuery(params ProductQueryParams) (products []Product, err error)
	CountAllProducts() (total int, err error)
}

type ProductRepositoryMySQL struct {
	DB *infras.MySQLConn
}

func ProvideProductRepositoryMySQL(db *infras.MySQLConn) *ProductRepositoryMySQL {
	s := new(ProductRepositoryMySQL)
	s.DB = db

	return s
}

func (r *ProductRepositoryMySQL) CreateProduct(product Product) (err error) {
	exists, err := r.ExistsByID(product.ID)
	if err != nil {
		logger.ErrorWithStack(err)
		return
	}

	if exists {
		err = failure.Conflict("create", "productID", "already exists")
		logger.ErrorWithStack(err)
		return
	}

	return r.DB.WithTransaction(func(tx *sqlx.Tx, e chan error) {
		if err := r.txCreate(tx, product); err != nil {
			e <- err
			return
		}

		e <- nil
	})
}

func (r *ProductRepositoryMySQL) ResolveProductsByQuery(params ProductQueryParams) (products []Product, err error) {
	query := productQueries.selectProducts

	var args []interface{}

	if params.Category != "" {
		query += " AND category = ?"
		args = append(args, params.Category)
	}

	if params.Brand != "" {
		query += " AND query = ?"
		args = append(args, params.Brand)
	}

	if params.Sort != "" && params.Order != "" {
		allowedSorts := map[string]bool{
			"name":  true,
			"price": true,
			"brand": true,
			"stock": true,
		}
		allowedOrders := map[string]bool{
			"asc":  true,
			"desc": true,
		}

		if allowedSorts[params.Sort] && allowedOrders[params.Order] {
			query += " ORDER BY " + params.Sort + " " + params.Order
		}
	}

	if params.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, params.Limit)

		if params.Page > 1 {
			offset := (params.Page - 1) * params.Limit
			query += " OFFSET ?"
			args = append(args, offset)
		}
	}

	err = r.DB.Read.Select(&products, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			err = failure.NotFound("courses")
			logger.ErrorWithStack(err)
			return nil, err
		}

		logger.ErrorWithStack(err)
		return nil, err
	}

	return products, nil
}

func (r *ProductRepositoryMySQL) CountAllProducts() (total int, err error) {
	query := `SELECT COUNT(*) FROM product`

	err = r.DB.Read.QueryRow(query).Scan(&total)
	if err != nil {
		logger.ErrorWithStack(err)
	}

	return total, nil
}

func (r *ProductRepositoryMySQL) ExistsByID(id uuid.UUID) (exists bool, err error) {
	err = r.DB.Read.Get(
		&exists,
		"SELECT COUNT(id) FROM product WHERE id = ?",
		id.String())

	if err != nil {
		logger.ErrorWithStack(err)
	}

	return
}

// Transactions
func (r *ProductRepositoryMySQL) txCreate(tx *sqlx.Tx, product Product) (err error) {
	stmt, err := tx.PrepareNamed(productQueries.insertProduct)
	if err != nil {
		logger.ErrorWithStack(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(product)
	if err != nil {
		logger.ErrorWithStack(err)
	}

	return
}
