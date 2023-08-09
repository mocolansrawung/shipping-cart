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
	}
)

type CartRepository interface {
	CreateCart(cart Cart) (err error)
	AddItemToCart(cartItem CartItem, userID uuid.UUID) (err error)
	ResolveByUserID(id uuid.UUID) (cart Cart, err error)
	ResolveItemsByCartID(ids []uuid.UUID) (cartItems []CartItem, err error)
	ExistsByUserID(id uuid.UUID) (exists bool, err error)
	GetCartIDByUserID(userID uuid.UUID) (cartID uuid.UUID, err error)
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
	exists, err := r.ExistsByUserID(cart.UserID)
	if err != nil {
		logger.ErrorWithStack(err)
		return
	}

	if exists {
		err = failure.Conflict("create", "userID with cartID", "already exists")
		logger.ErrorWithStack(err)
		return
	}

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
	// 1. Check if the cart exists for the user.
	cartID, err := r.GetCartIDByUserID(userID)
	if err != nil {
		return
	}

	// 2. If not, create a new cart.
	if cartID == uuid.Nil {
		newCart := Cart{
			UserID:    userID,
			CreatedAt: time.Now(),
			CreatedBy: userID,
			// ... any other necessary fields
		}

		if err := r.CreateCart(newCart); err != nil {
			return err
		}

		// Set the CartID of the cartItem to the ID of the new cart.
		cartItem.CartID = newCart.ID
		cartID = newCart.ID
	}

	// 3. Add the item to the cart.
	return r.DB.WithTransaction(func(tx *sqlx.Tx, e chan error) {
		// Using your bulk insertion, although we're inserting only one item, we can make use of the bulk insert feature.
		err := r.txCreateItems(tx, []CartItem{cartItem})
		if err != nil {
			e <- err
			return
		}

		e <- nil
	})
}

// func (r *CartRepositoryMySQL) AddItemToCart(cartItem CartItem, userID uuid.UUID) (err error) {
// 	// 1. Check if the cart exists for the user.
// 	cartID, err := r.GetCartIDByUserID(userID)
// 	if err != nil {
// 		return
// 	}

// 	// 2. If not, create a new cart.

// 	if cartID == uuid.Nil {

// 	}

// 	// 3. Add the item to the cart.
// 	// 4. Return any errors encountered.

// 	return r.DB.WithTransaction(func(tx *sqlx.Tx, e chan error) {
// 		if err := r.txCreate(tx, cartItem); err != nil {
// 			e <- err
// 			return
// 		}

// 		e <- nil
// 	})

// }

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

// Exists Functions
func (r *CartRepositoryMySQL) ExistsByUserID(id uuid.UUID) (exists bool, err error) {
	err = r.DB.Read.Get(
		&exists,
		"SELECT COUNT(entity_id) FROM cart WHERE user_id = ?",
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
