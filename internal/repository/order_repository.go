package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/yourname/stockbit-appsec/internal/domain"
)

type OrderRepository struct {
	DB *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{DB: db}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *domain.Order) error {
	query := `INSERT INTO orders (user_id, stock_symbol, amount, type) VALUES ($1, $2, $3, $4) RETURNING id`
	err := r.DB.QueryRowContext(ctx, query, order.UserID, order.StockSymbol, order.Amount, order.Type).Scan(&order.ID)
	if err != nil {
		return err
	}
	return nil
}

func (r *OrderRepository) GetOrderById(ctx context.Context, id int) (*domain.Order, error) {
	query := `SELECT id, user_id, stock_symbol, amount, type FROM orders WHERE id = $1`
	row := r.DB.QueryRowContext(ctx, query, id)

	var order domain.Order
	err := row.Scan(&order.ID, &order.UserID, &order.StockSymbol, &order.Amount, &order.Type)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Order not found
		}
		return nil, err
	}
	return &order, nil
}
