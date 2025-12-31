package v1_vulnerable

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourname/stockbit-appsec/internal/usecase"
)

type OrderHandler struct {
	OrderUseCase *usecase.OrderUsecase
}

func NewOrderHandler(u *usecase.OrderUsecase) *OrderHandler {
	return &OrderHandler{OrderUseCase: u}
}

// GetOrder retrieves an order by ID.
// VULNERABILITY: It does NOT check if the authenticated user is the owner of the order.
// This allows any authenticated user to access any order (IDOR).
func (h *OrderHandler) GetOrder(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	// Unsafe: Directly getting order without ownership check
	order, err := h.OrderUseCase.GetOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}
