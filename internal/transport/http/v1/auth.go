package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horoshi10v/tires-shop/internal/domain"
)

type AuthHandler struct {
	service domain.AuthService
}

func NewAuthHandler(service domain.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// LoginTelegram handles Telegram Mini App authentication.
//
//	@Summary      Login via Telegram Mini App
//	@Description  Validates Telegram initData and returns a JWT token.
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Param        data  body      domain.AuthRequestDTO  true  "Telegram initData"
//	@Success      200   {object}  domain.AuthResponseDTO
//	@Failure      400   {object}  map[string]string
//	@Failure      401   {object}  map[string]string
//	@Router       /auth/telegram [post]
func (h *AuthHandler) LoginTelegram(c *gin.Context) {
	var req domain.AuthRequestDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	resp, err := h.service.LoginTelegram(c.Request.Context(), req.InitData)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
