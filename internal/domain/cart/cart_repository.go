package cart

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/evermos/boilerplate-go/infras"
	"github.com/evermos/boilerplate-go/shared/failure"
	"github.com/evermos/boilerplate-go/shared/logger"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
)

type CartRepository interface {
	ResolveCartByUserID(id uuid.UUID) (cart Cart, err error)
	ResolveItemsByCartID(ids []uuid.UUID) (cartItems []CartItem, err error)
	ResolveOrCreateCartByUserID(userID uuid.UUID) (cart Cart, err error)
	CreateCart(cart Cart) (err error)
	ResolveCartItemByProductID(cartID, productID uuid.UUID) (cartItem CartItem, found bool, err error)
	UpdateItemQuantity(cartItem CartItem) (err error)
	CreateCartItem(cartItem CartItem, userID uuid.UUID) (err error)
	ResolveDetailedItemsByCartID(ids []uuid.UUID) (cartItems []CartItem, err error)
}

type CartRepositoryMySQL struct {
	DB *infras.MySQLConn
}

func ProvideCartRepositoryMySQL(db *infras.MySQLConn) *CartRepositoryMySQL {
	s := new(CartRepositoryMySQL)
	s.DB = db

	return s
}

func (r *CartRepositoryMySQL) ResolveCartByUserID(id uuid.UUID) (cart Cart, err error) {
	insertQuery := `SELECT id, user_id, created_at, created_by, updated_at, updated_by, deleted_at, deleted_by FROM cart`
	err = r.DB.Read.Get(
		&cart,
		insertQuery+" WHERE user_id = ?",
		id.String())

	if err != nil && err == sql.ErrNoRows {
		err = failure.NotFound("cart")
		logger.ErrorWithStack(err)
		return
	} else if err != nil {
		logger.ErrorWithStack(err)
		return
	}

	return
}
func (r *CartRepositoryMySQL) ResolveItemsByCartID(ids []uuid.UUID) (cartItems []CartItem, err error) {
	insertQuery := `SELECT id, cart_id, product_id, unit_price, quantity, cost, created_at, created_by, updated_at, updated_by, deleted_at, deleted_by FROM cart_item`

	if len(ids) == 0 {
		return
	}

	query, args, err := sqlx.In(insertQuery+" WHERE cart_id IN (?)", ids)
	if err != nil {
		logger.ErrorWithStack(err)
		return
	}

	err = r.DB.Read.Select(&cartItems, query, args...)
	if err != nil {
		logger.ErrorWithStack(err)
		return
	}

	return
}
func (r *CartRepositoryMySQL) ResolveOrCreateCartByUserID(userID uuid.UUID) (cart Cart, err error) {
	cart, err = r.ResolveCartByUserID(userID)
	if err == nil {
		return
	}

	cart, err = cart.NewFromRequestFormat(userID)
	if err != nil {
		return
	}

	err = r.CreateCart(cart)
	if err != nil {
		logger.ErrorWithStack(err)
		return
	}

	return
}
func (r *CartRepositoryMySQL) CreateCart(cart Cart) (err error) {
	return r.DB.WithTransaction(func(tx *sqlx.Tx, e chan error) {
		if err := r.txCreate(tx, cart); err != nil {
			e <- err
			return
		}

		if err := r.txCreateItems(tx, cart.Items); err != nil {
			e <- err
			return
		}

		e <- nil
	})
}
func (r *CartRepositoryMySQL) ResolveCartItemByProductID(cartID, productID uuid.UUID) (cartItem CartItem, found bool, err error) {
	selectQuery := `SELECT id, cart_id, product_id, unit_price, quantity, cost, created_at, created_by, updated_at, updated_by, deleted_at, deleted_by FROM cart_item WHERE cart_id = ? AND product_id = ?`

	err = r.DB.Read.Get(&cartItem, selectQuery, cartID, productID)
	if err != nil && err == sql.ErrNoRows {
		return cartItem, false, nil
	} else if err != nil {
		logger.ErrorWithStack(err)
		return cartItem, false, err
	}

	return cartItem, true, nil
}
func (r *CartRepositoryMySQL) UpdateItemQuantity(cartItem CartItem) (err error) {
	return r.DB.WithTransaction(func(tx *sqlx.Tx, e chan error) {
		if err := r.txUpdate(tx, cartItem); err != nil {
			e <- err
			return
		}

		e <- nil
	})
}
func (r *CartRepositoryMySQL) CreateCartItem(cartItem CartItem, userID uuid.UUID) (err error) {
	if cartItem.CartID == uuid.Nil {
		newCart := Cart{
			UserID:    userID,
			CreatedAt: time.Now(),
			CreatedBy: userID,
		}

		if err := r.CreateCart(newCart); err != nil {
			return err
		}

		cartItem.CartID = newCart.ID
	}

	return r.DB.WithTransaction(func(tx *sqlx.Tx, e chan error) {
		err := r.txCreateItems(tx, []CartItem{cartItem})
		if err != nil {
			e <- err
			return
		}

		e <- nil
	})
}
func (r *CartRepositoryMySQL) ResolveDetailedItemsByCartID(ids []uuid.UUID) (cartItems []CartItem, err error) {
	initialQuery := `SELECT cart_item.cart_id, cart_item.product_id, cart_item.unit_price, cart_item.quantity, cart_item.cost, cart_item.created_at, cart_item.created_by, cart_item.updated_at, cart_item.updated_by, cart_item.deleted_at, cart_item.deleted_by, product.stock FROM cart_item JOIN product ON cart_item.product_id = product.id;
	`
	if len(ids) == 0 {
		return
	}

	query, args, err := sqlx.In(initialQuery+" WHERE cart_id IN (?)", ids)
	if err != nil {
		logger.ErrorWithStack(err)
		return
	}

	err = r.DB.Read.Select(&cartItems, query, args...)
	if err != nil {
		logger.ErrorWithStack(err)
		return
	}

	return
}

// Transactions
func (r *CartRepositoryMySQL) txCreate(tx *sqlx.Tx, cart Cart) (err error) {
	insertQuery := `INSERT INTO cart (id, user_id, created_at, created_by, updated_at, updated_by, deleted_at, deleted_by) VALUES (:id, :user_id, :created_at, :created_by, :updated_at, :updated_by, :deleted_at, :deleted_by)`

	stmt, err := tx.PrepareNamed(insertQuery)
	if err != nil {
		logger.ErrorWithStack(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(cart)
	if err != nil {
		logger.ErrorWithStack(err)
	}

	return
}
func (r *CartRepositoryMySQL) composeBulkInsertItemQuery(cartItems []CartItem) (query string, params []interface{}, err error) {
	insertCartItemBulk := `INSERT INTO cart_item (id, cart_id, product_id, unit_price, quantity, cost, created_at, created_by, updated_at, updated_by, deleted_at, deleted_by) VALUES `
	insertCartItemBulkPlaceholder := `(:id, :cart_id, :product_id, :unit_price, :quantity, :cost, :created_at, :created_by, :updated_at, :updated_by, :deleted_at, :deleted_by)`

	values := []string{}
	for _, ci := range cartItems {
		param := map[string]interface{}{
			"id":         ci.ID,
			"cart_id":    ci.CartID,
			"product_id": ci.ProductID,
			"unit_price": ci.UnitPrice,
			"quantity":   ci.Quantity,
			"cost":       ci.Cost,
			"created_at": ci.CreatedAt,
			"created_by": ci.CreatedBy,
			"updated_at": ci.UpdatedAt,
			"updated_by": ci.UpdatedBy,
			"deleted_at": ci.DeletedAt,
			"deleted_by": ci.DeletedBy,
		}
		q, args, err := sqlx.Named(insertCartItemBulkPlaceholder, param)
		if err != nil {
			return query, params, err
		}
		values = append(values, q)
		params = append(params, args...)
	}
	query = fmt.Sprintf("%v %v", insertCartItemBulk, strings.Join(values, ","))
	return
}
func (r *CartRepositoryMySQL) txCreateItems(tx *sqlx.Tx, cartItems []CartItem) (err error) {
	if len(cartItems) == 0 {
		return
	}

	query, args, err := r.composeBulkInsertItemQuery(cartItems)
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
func (r *CartRepositoryMySQL) txUpdate(tx *sqlx.Tx, cartItem CartItem) (err error) {
	updateQuery := `UPDATE cart_item SET id = :id, cart_id = :cart_id, product_id = :product_id, unit_price = :unit_price, quantity = :quantity, cost = :cost, created_at = :created_at, created_by = :created_by, updated_at = :updated_at, updated_by = :updated_by, deleted_at = :deleted_at, deleted_by = :deleted_by WHERE id = :id`
	stmt, err := tx.PrepareNamed(updateQuery)
	if err != nil {
		logger.ErrorWithStack(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(cartItem)
	if err != nil {
		logger.ErrorWithStack(err)
	}

	return
}
