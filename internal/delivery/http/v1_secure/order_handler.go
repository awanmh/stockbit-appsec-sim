package v1_secure

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
// SECURE: It checks if the authenticated user is the owner of the order.
func (h *OrderHandler) GetOrder(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	// 1. Get the authenticated User ID from context (set by Auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 2. Get the order
	order, err := h.OrderUseCase.GetOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// 3. Authorization Check: Ensure the order belongs to the user
	if order.UserID != userID.(int) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to access this order"})
		return
	}

	c.JSON(http.StatusOK, order)
}
