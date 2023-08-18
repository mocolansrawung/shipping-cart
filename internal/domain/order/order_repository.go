package order

import (
	"fmt"
	"strings"

	"github.com/evermos/boilerplate-go/infras"
	"github.com/evermos/boilerplate-go/shared/failure"
	"github.com/evermos/boilerplate-go/shared/logger"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
)

type OrderRepository interface {
	Checkout(order Order, cartID uuid.UUID) (err error)
	ExistsByID(id uuid.UUID) (exists bool, err error)
}

type OrderRepositoryMySQL struct {
	DB *infras.MySQLConn
}

func ProvideOrderRepositoryMySQL(db *infras.MySQLConn) *OrderRepositoryMySQL {
	s := new(OrderRepositoryMySQL)
	s.DB = db

	return s
}

func (r *OrderRepositoryMySQL) Checkout(order Order, cartID uuid.UUID) (err error) {
	exists, err := r.ExistsByID(order.ID)
	if err != nil {
		logger.ErrorWithStack(err)
		return
	}

	if exists {
		err = failure.Conflict("create", "order", "already exists")
		logger.ErrorWithStack(err)
		return
	}

	var productIDs []uuid.UUID
	for _, item := range order.Items {
		productIDs = append(productIDs, item.ProductID)
	}

	return r.DB.WithTransaction(func(tx *sqlx.Tx, e chan error) {
		// Wrap the entire checkout process in a transaction
		txErr := func(err error) {
			if err != nil {
				e <- err
				tx.Rollback() // Rollback the transaction
			}
		}

		// Create the order
		if err := r.txCreate(tx, order); err != nil {
			txErr(err)
			return
		}

		// Transfer items to the order
		if err := r.txTransferItemsToOrder(tx, order.Items); err != nil {
			txErr(err)
			return
		}

		// Remove checked out cart items
		if err := r.txRemoveCheckedOutCartItems(tx, cartID, productIDs); err != nil {
			txErr(err)
			return
		}

		e <- nil
	})
}

func (r *OrderRepositoryMySQL) ExistsByID(id uuid.UUID) (exists bool, err error) {
	err = r.DB.Read.Get(
		&exists,
		"SELECT COUNT(id) FROM orders WHERE id = ?",
		id.String())
	if err != nil {
		logger.ErrorWithStack(err)
	}

	return
}

// Transactions
func (r *OrderRepositoryMySQL) composeBulkInsertItemQuery(orderItems []OrderItem) (query string, params []interface{}, err error) {
	bulkQuery := `INSERT INTO order_item (cart_id, product_id, unit_price, quantity, cost, created_at, created_by, updated_at, updated_by, deleted_at, deleted_by) VALUES `
	bulkPlaceholderQuery := `(:cart_id, :product_id, :unit_price, :quantity, :cost, :created_at, :created_by, :updated_at, :updated_by, :deleted_at, :deleted_by)`

	values := []string{}
	for _, oi := range orderItems {
		param := map[string]interface{}{
			"order_id": oi.OrderID,
		}
		q, args, err := sqlx.Named(bulkPlaceholderQuery, param)
		if err != nil {
			return query, params, err
		}
		values = append(values, q)
		params = append(params, args...)
	}
	query = fmt.Sprintf("%v %v", bulkQuery, strings.Join(values, ","))
	return
}
func (r *OrderRepositoryMySQL) txCreate(tx *sqlx.Tx, order Order) (err error) {
	query := `INSERT INTO orders (id, user_id, total_cost, status, created_at, created_by, updated_at, updated_by, deleted_at, deleted_by) VALUES (:id, :user_id, :total_cost, :status, :created_at, :created_by, :updated_at, :updated_by, :deleted_at, :deleted_by)`

	stmt, err := tx.PrepareNamed(query)
	if err != nil {
		logger.ErrorWithStack(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(order)
	if err != nil {
		logger.ErrorWithStack(err)
	}

	return
}
func (r *OrderRepositoryMySQL) txCreateItems(tx *sqlx.Tx, orderItems []OrderItem) (err error) {
	if len(orderItems) == 0 {
		return
	}

	query, args, err := r.composeBulkInsertItemQuery(orderItems)
	if err != nil {
		return
	}

	stmt, err := tx.Preparex(query)
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Stmt.Exec(args...)
	if err != nil {
		logger.ErrorWithStack(err)
	}

	return
}
func (r *OrderRepositoryMySQL) txTransferItemsToOrder(tx *sqlx.Tx, orderItems []OrderItem) error {
	return r.txCreateItems(tx, orderItems)
}
func (r *OrderRepositoryMySQL) txRemoveCheckedOutCartItems(tx *sqlx.Tx, cartID uuid.UUID, productIDs []uuid.UUID) error {
	if len(productIDs) == 0 {
		return nil
	}

	query := `DELETE FROM cart_item WHERE cart_id = ? AND product_id IN (?)`
	query, args, err := sqlx.In(query, cartID, productIDs)
	if err != nil {
		logger.ErrorWithStack(err)
		return err
	}

	query = tx.Rebind(query)

	stmt, err := tx.Prepare(query)
	if err != nil {
		logger.ErrorWithStack(err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		logger.ErrorWithStack(err)
		return err
	}

	return nil
}
