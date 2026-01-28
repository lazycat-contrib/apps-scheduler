package handlers

import (
	"net/http"

	"apps-scheduler/internal/auth"
	"apps-scheduler/internal/biz"
	"apps-scheduler/internal/pkg/serverchan"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type NotifyHandler struct {
	useCase *biz.UseCase
}

func NewNotifyHandler(useCase *biz.UseCase) *NotifyHandler {
	return &NotifyHandler{useCase: useCase}
}

type NotifyConfigRequest struct {
	SendKey   string `json:"sendKey"`
	Enabled   bool   `json:"enabled"`
	OnSuccess bool   `json:"onSuccess"`
	OnFailure bool   `json:"onFailure"`
}

type NotifyConfigResponse struct {
	SendKey   string `json:"sendKey"`
	Enabled   bool   `json:"enabled"`
	OnSuccess bool   `json:"onSuccess"`
	OnFailure bool   `json:"onFailure"`
}

func (h *NotifyHandler) GetConfig(c echo.Context) error {
	ctx := c.Request().Context()
	userID := auth.GetUserID(c)

	config, err := h.useCase.GetNotifyConfig(ctx, userID)
	if err != nil {
		// Return empty config if not found
		return c.JSON(http.StatusOK, NotifyConfigResponse{
			SendKey:   "",
			Enabled:   false,
			OnSuccess: true,
			OnFailure: true,
		})
	}

	return c.JSON(http.StatusOK, NotifyConfigResponse{
		SendKey:   config.SendKey,
		Enabled:   config.Enabled,
		OnSuccess: config.OnSuccess,
		OnFailure: config.OnFailure,
	})
}

func (h *NotifyHandler) SaveConfig(c echo.Context) error {
	ctx := c.Request().Context()
	userID := auth.GetUserID(c)

	var req NotifyConfigRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	config, err := h.useCase.SaveNotifyConfig(ctx, userID, req.SendKey, req.Enabled, req.OnSuccess, req.OnFailure)
	if err != nil {
		log.Error().Err(err).Msg("Failed to save notify config")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save configuration"})
	}

	log.Info().Str("user_id", userID).Msg("Notify config saved")

	return c.JSON(http.StatusOK, NotifyConfigResponse{
		SendKey:   config.SendKey,
		Enabled:   config.Enabled,
		OnSuccess: config.OnSuccess,
		OnFailure: config.OnFailure,
	})
}

func (h *NotifyHandler) TestNotify(c echo.Context) error {
	ctx := c.Request().Context()
	userID := auth.GetUserID(c)

	config, err := h.useCase.GetNotifyConfig(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Please configure notification settings first"})
	}

	if config.SendKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "SendKey is not configured"})
	}

	client := serverchan.NewClient(config.SendKey)
	if err := client.Send("测试通知", "这是来自 **应用定时管家** 的测试通知"); err != nil {
		log.Error().Err(err).Msg("Failed to send test notification")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to send test notification: " + err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Test notification sent"})
}
