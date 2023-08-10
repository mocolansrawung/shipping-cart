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

var (
	cartQueries = struct {
		selectCart                    string
		selectCartItem                string
		insertCart                    string
		insertCartItemBulk            string
		insertCartItemBulkPlaceholder string
		updateCartItem                string
	}{
		selectCart: `
			SELECT
				id,
				user_id,
				created_at,
				created_by,
				updated_at,
				updated_by,
				deleted_at,
				deleted_by
			FROM cart 
		`,

		selectCartItem: `
			SELECT
				cart_id,
				product_id,
				unit_price,
				quantity,
				cost,
				created_at,
				created_by,
				updated_at,
				updated_by,
				deleted_at,
				deleted_by
			FROM cart_item 
		`,

		insertCart: `
			INSERT INTO cart (
				id,
				user_id,
				created_at,
				created_by,
				updated_at,
				updated_by,
				deleted_at,
				deleted_by
			) VALUES (
				:id,
				:user_id,
				:created_at,
				:created_by,
				:updated_at,
				:updated_by,
				:deleted_at,
				:deleted_by
			)
		`,

		insertCartItemBulk: `
			INSERT INTO cart_item (
				cart_id,
				product_id,
				unit_price,
				quantity,
				cost,
				created_at,
				created_by,
				updated_at,
				updated_by,
				deleted_at,
				deleted_by
			) VALUES 
		`,

		insertCartItemBulkPlaceholder: `
			(
				:cart_id,
				:product_id,
				:unit_price,
				:quantity,
				:cost,
				:created_at,
				:created_by,
				:updated_at,
				:updated_by,
				:deleted_at,
				:deleted_by
			)
		`,

		updateCartItem: `
			UPDATE cart_item
			SET
				cart_id = :cart_id,
				product_id = :product_id,
				unit_price = :unit_price,
				quantity = :quantity,
				cost = :cost,
				created_at = :created_at,
				created_by = :created_by,
				updated_at = :updated_at,
				updated_by = :updated_by,
				deleted_at = :deleted_at
		`,
	}
)

type CartRepository interface {
	CreateCart(cart Cart) (err error)
	AddItemToCart(cartItem CartItem, userID uuid.UUID) (err error)
	ResolveByUserID(id uuid.UUID) (cart Cart, err error)
	ResolveItemsByCartID(ids []uuid.UUID) (cartItems []CartItem, err error)
	ExistsByUserID(id uuid.UUID) (exists bool, err error)
	GetCartIDByUserID(userID uuid.UUID) (cartID uuid.UUID, err error)
	GetPriceAndStockByProductID(productID uuid.UUID) (price float64, stock int, err error)
	ItemExistsInCart(cartItem CartItem) (exists bool, err error)
	GetCurrentItemQuantity(cartItem CartItem) (latestQuantity int, err error)
	UpdateItemQuantity(cartItem CartItem) (err error)
}

type CartRepositoryMySQL struct {
	DB *infras.MySQLConn
}

func ProvideCartRepositoryMySQL(db *infras.MySQLConn) *CartRepositoryMySQL {
	s := new(CartRepositoryMySQL)
	s.DB = db

	return s
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
func (r *CartRepositoryMySQL) AddItemToCart(cartItem CartItem, userID uuid.UUID) (err error) {
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
func (r *CartRepositoryMySQL) ResolveByUserID(id uuid.UUID) (cart Cart, err error) {
	err = r.DB.Read.Get(
		&cart,
		cartQueries.selectCart+" WHERE user_id = ?",
		id.String())

	if err != nil && err == sql.ErrNoRows {
		err = failure.NotFound("cart")
		logger.ErrorWithStack(err)
	} else if err != nil {
		logger.ErrorWithStack(err)
		return
	}

	return
}
func (r *CartRepositoryMySQL) ResolveItemsByCartID(ids []uuid.UUID) (cartItems []CartItem, err error) {
	if len(ids) == 0 {
		return
	}

	query, args, err := sqlx.In(cartQueries.selectCartItem+" WHERE cart_id IN (?)", ids)
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
func (r *CartRepositoryMySQL) ItemExistsInCart(cartItem CartItem) (exists bool, err error) {
	err = r.DB.Read.Get(
		&exists,
		"SELECT COUNT(produc_id) FROM cart WHERE cart_id = ? and product_id = ?",
		cartItem.CartID.String(), cartItem.ProductID.String())
	if err != nil {
		logger.ErrorWithStack(err)
	}

	return
}
func (r *CartRepositoryMySQL) GetCurrentItemQuantity(cartItem CartItem) (latestQuantity int, err error) {
	query := "SELECT quantity FROM cart_item WHERE cart_id = ? AND product_id = ?"
	err = r.DB.Read.QueryRow(query, cartItem.CartID, cartItem.ProductID).Scan(&latestQuantity)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	return
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
func (r *CartRepositoryMySQL) GetPriceAndStockByProductID(productID uuid.UUID) (price float64, stock int, err error) {
	query := "SELECT price, stock FROM product WHERE id = ?"
	err = r.DB.Read.QueryRow(query, productID).Scan(&price, &stock)
	if err != nil {
		return
	}

	return
}
func (r *CartRepositoryMySQL) ExistsByUserID(id uuid.UUID) (exists bool, err error) {
	err = r.DB.Read.Get(
		&exists,
		"SELECT COUNT(id) FROM cart WHERE user_id = ?",
		id.String())
	if err != nil {
		logger.ErrorWithStack(err)
	}

	return
}

func (r *CartRepositoryMySQL) GetCartIDByUserID(userID uuid.UUID) (cartID uuid.UUID, err error) {
	var id string
	err = r.DB.Read.Get(&id, "SELECT entity_id FROM cart WHERE user_id = ?", userID.String())
	if err == sql.ErrNoRows {
		return
	} else if err != nil {
		logger.ErrorWithStack(err)
		return
	}

	cartID, err = uuid.FromString(id)
	if err != nil {
		logger.ErrorWithStack(err)
		return
	}

	return
}

// Transactions
func (r *CartRepositoryMySQL) txCreate(tx *sqlx.Tx, cart Cart) (err error) {
	stmt, err := tx.PrepareNamed(cartQueries.insertCart)
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
	values := []string{}
	for _, ci := range cartItems {
		param := map[string]interface{}{
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
		q, args, err := sqlx.Named(cartQueries.insertCartItemBulkPlaceholder, param)
		if err != nil {
			return query, params, err
		}
		values = append(values, q)
		params = append(params, args...)
	}
	query = fmt.Sprintf("%v %v", cartQueries.insertCartItemBulk, strings.Join(values, ","))
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
	stmt, err := tx.PrepareNamed(cartQueries.updateCartItem)
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
