package domain

type OrderType string

const (
	Buy  OrderType = "buy"
	Sell OrderType = "sell"
)

type Order struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	StockSymbol string    `json:"stock_symbol"`
	Amount      int       `json:"amount"`
	Type        OrderType `json:"type"`
}
