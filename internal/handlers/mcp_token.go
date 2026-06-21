package handlers

import (
	"net/http"
	"strings"
	"time"

	"apps-scheduler/internal/auth"
	"apps-scheduler/internal/biz"
	"apps-scheduler/internal/ent"
	"apps-scheduler/internal/tokens"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

const (
	defaultMCPTokenName = "MCP token"
	maxMCPTokenNameLen  = 80
)

type MCPTokenHandler struct {
	useCase *biz.UseCase
}

func NewMCPTokenHandler(useCase *biz.UseCase) *MCPTokenHandler {
	return &MCPTokenHandler{useCase: useCase}
}

type MCPTokenCreateRequest struct {
	Name string `json:"name"`
}

type MCPTokenResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"`
	UserID     string     `json:"userId"`
	UserRole   string     `json:"userRole"`
	CreatedAt  time.Time  `json:"createdAt"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
	RevokedAt  *time.Time `json:"revokedAt,omitempty"`
}

type MCPTokenCreateResponse struct {
	Token string           `json:"token"`
	Item  MCPTokenResponse `json:"item"`
}

func (h *MCPTokenHandler) ListTokens(c echo.Context) error {
	userID := auth.GetUserID(c)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Not authenticated"})
	}

	items, err := h.useCase.ListMCPTokens(c.Request().Context(), userID, true)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to list MCP tokens")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to list tokens"})
	}

	resp := make([]MCPTokenResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, mcpTokenResponse(item))
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *MCPTokenHandler) CreateToken(c echo.Context) error {
	userID := auth.GetUserID(c)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Not authenticated"})
	}

	var req MCPTokenCreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = defaultMCPTokenName
	}
	if len(name) > maxMCPTokenNameLen {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Token name is too long"})
	}

	plainToken, err := tokens.Generate()
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate MCP token")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	item, err := h.useCase.CreateMCPToken(
		c.Request().Context(),
		name,
		tokens.Hash(plainToken),
		tokens.DisplayPrefix(plainToken),
		userID,
		auth.GetUserRole(c),
	)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to save MCP token")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save token"})
	}

	return c.JSON(http.StatusCreated, MCPTokenCreateResponse{
		Token: plainToken,
		Item:  mcpTokenResponse(item),
	})
}

func (h *MCPTokenHandler) RevokeToken(c echo.Context) error {
	userID := auth.GetUserID(c)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Not authenticated"})
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid token ID"})
	}

	item, err := h.useCase.RevokeMCPToken(c.Request().Context(), id, userID, auth.GetUserRole(c))
	if err != nil {
		log.Error().Err(err).Str("token_id", id.String()).Str("user_id", userID).Msg("Failed to revoke MCP token")
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Failed to revoke token"})
	}

	return c.JSON(http.StatusOK, mcpTokenResponse(item))
}

func mcpTokenResponse(item *ent.MCPToken) MCPTokenResponse {
	return MCPTokenResponse{
		ID:         item.ID.String(),
		Name:       item.Name,
		Prefix:     item.TokenPrefix,
		UserID:     item.UserID,
		UserRole:   item.UserRole,
		CreatedAt:  item.CreatedAt,
		LastUsedAt: item.LastUsedAt,
		RevokedAt:  item.RevokedAt,
	}
}
