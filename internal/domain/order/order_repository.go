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
	CreateOrder(order Order) (err error)
}

type OrderRepositoryMySQL struct {
	DB *infras.MySQLConn
}

func ProvideOrderRepositoryMySQL(db *infras.MySQLConn) *OrderRepositoryMySQL {
	s := new(OrderRepositoryMySQL)
	s.DB = db

	return s
}

func (r *OrderRepositoryMySQL) CreateOrder(order Order) (err error) {
	// Implement CreateOrder without RequestFormat or hit the endpoint, it's an automatated process that involved in the flow of Checkout endpoint.
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

	return r.DB.WithTransaction(func(tx *sqlx.Tx, e chan error) {
		if err := r.txCreate(tx, order); err != nil {
			e <- nil
			return
		}

		if err := r.txCreateItems(tx, order.Items); err != nil {
			e <- err
			return
		}

		e <- nil
	})
}
func (r *OrderRepositoryMySQL) ExistsByID(id uuid.UUID) (exists bool, err error) {
	err = r.DB.Read.Get(
		&exists,
		"SELECT COUNT(entity_id) FROM order WHERE id = ?",
		id.String())
	if err != nil {
		logger.ErrorWithStack(err)
	}

	return
}

// transactions
func (r *OrderRepositoryMySQL) composeBulkInsertItemQuery(orderItems []OrderItem) (query string, params []interface{}, err error) {
	bulkQuery := ``
	bulkPlaceholderQuery := ``

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
	query := ``

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
