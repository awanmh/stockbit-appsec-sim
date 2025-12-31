package usecase

import (
	"context"

	"github.com/yourname/stockbit-appsec/internal/domain"
	"github.com/yourname/stockbit-appsec/internal/repository"
)

type OrderUsecase struct {
	OrderRepo *repository.OrderRepository
}

func NewOrderUsecase(orderRepo *repository.OrderRepository) *OrderUsecase {
	return &OrderUsecase{
		OrderRepo: orderRepo,
	}
}

func (u *OrderUsecase) PlaceOrder(ctx context.Context, userID int, symbol string, amount int, orderType domain.OrderType) error {
	order := &domain.Order{
		UserID:      userID,
		StockSymbol: symbol,
		Amount:      amount,
		Type:        orderType,
	}
	return u.OrderRepo.CreateOrder(ctx, order)
}

func (u *OrderUsecase) GetOrder(ctx context.Context, id int) (*domain.Order, error) {
	return u.OrderRepo.GetOrderById(ctx, id)
}
