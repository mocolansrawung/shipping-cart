package order

import "github.com/evermos/boilerplate-go/infras"

type OrderRepository interface {
}

type OrderRepositoryMySQL struct {
	DB *infras.MySQLConn
}

func ProvideOrderRepositoryMySQL(db *infras.MySQLConn) *OrderRepositoryMySQL {
	s := new(OrderRepositoryMySQL)
	s.DB = db

	return s
}
