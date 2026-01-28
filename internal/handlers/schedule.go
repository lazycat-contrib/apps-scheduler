package handlers

import (
	"net/http"
	"time"

	"apps-scheduler/internal/auth"
	"apps-scheduler/internal/biz"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type ScheduleHandler struct {
	useCase *biz.UseCase
}

func NewScheduleHandler(useCase *biz.UseCase) *ScheduleHandler {
	return &ScheduleHandler{useCase: useCase}
}

type ScheduleRequest struct {
	Name      string `json:"name"`
	AppID     string `json:"appId"`
	AppTitle  string `json:"appTitle"`
	Operation string `json:"operation"`
	WeekDays  []int  `json:"weekDays"`
	Hour      int    `json:"hour"`
	Minute    int    `json:"minute"`
	Enabled   *bool  `json:"enabled,omitempty"`
}

type ScheduleResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	AppID     string    `json:"appId"`
	AppTitle  string    `json:"appTitle"`
	Operation string    `json:"operation"`
	WeekDays  []int     `json:"weekDays"`
	Hour      int       `json:"hour"`
	Minute    int       `json:"minute"`
	Enabled   bool      `json:"enabled"`
	Creator   string    `json:"creator"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (h *ScheduleHandler) ListSchedules(c echo.Context) error {
	ctx := c.Request().Context()

	schedules, err := h.useCase.ListSchedules(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list schedules")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to list schedules"})
	}

	response := make([]ScheduleResponse, 0, len(schedules))
	for _, sch := range schedules {
		response = append(response, ScheduleResponse{
			ID:        sch.ID.String(),
			Name:      sch.Name,
			AppID:     sch.AppID,
			AppTitle:  sch.AppTitle,
			Operation: string(sch.Operation),
			WeekDays:  sch.WeekDays,
			Hour:      sch.Hour,
			Minute:    sch.Minute,
			Enabled:   sch.Enabled,
			Creator:   sch.Creator,
			CreatedAt: sch.CreatedAt,
			UpdatedAt: sch.UpdatedAt,
		})
	}

	return c.JSON(http.StatusOK, response)
}

func (h *ScheduleHandler) CreateSchedule(c echo.Context) error {
	ctx := c.Request().Context()
	userID := auth.GetUserID(c)

	var req ScheduleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Name == "" || req.AppID == "" || req.Operation == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name, appId, and operation are required"})
	}

	if req.Operation != "resume" && req.Operation != "pause" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Operation must be 'resume' or 'pause'"})
	}

	if len(req.WeekDays) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "At least one weekday is required"})
	}

	sch, err := h.useCase.CreateSchedule(ctx, req.Name, req.AppID, req.AppTitle, req.Operation, userID, req.WeekDays, req.Hour, req.Minute)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create schedule")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create schedule"})
	}

	log.Info().Str("schedule_id", sch.ID.String()).Str("user_id", userID).Msg("Schedule created")

	return c.JSON(http.StatusCreated, ScheduleResponse{
		ID:        sch.ID.String(),
		Name:      sch.Name,
		AppID:     sch.AppID,
		AppTitle:  sch.AppTitle,
		Operation: string(sch.Operation),
		WeekDays:  sch.WeekDays,
		Hour:      sch.Hour,
		Minute:    sch.Minute,
		Enabled:   sch.Enabled,
		Creator:   sch.Creator,
		CreatedAt: sch.CreatedAt,
		UpdatedAt: sch.UpdatedAt,
	})
}

func (h *ScheduleHandler) UpdateSchedule(c echo.Context) error {
	ctx := c.Request().Context()
	userID := auth.GetUserID(c)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid schedule ID"})
	}

	// Check if schedule exists and user has permission
	existing, err := h.useCase.GetSchedule(ctx, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Schedule not found"})
	}

	// Only creator can update (or admin)
	if existing.Creator != userID && auth.GetUserRole(c) != auth.RoleAdmin {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Permission denied"})
	}

	var req ScheduleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Name == "" || req.AppID == "" || req.Operation == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name, appId, and operation are required"})
	}

	enabled := existing.Enabled
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	sch, err := h.useCase.UpdateSchedule(ctx, id, req.Name, req.AppID, req.AppTitle, req.Operation, req.WeekDays, req.Hour, req.Minute, enabled)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update schedule")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update schedule"})
	}

	log.Info().Str("schedule_id", sch.ID.String()).Str("user_id", userID).Msg("Schedule updated")

	return c.JSON(http.StatusOK, ScheduleResponse{
		ID:        sch.ID.String(),
		Name:      sch.Name,
		AppID:     sch.AppID,
		AppTitle:  sch.AppTitle,
		Operation: string(sch.Operation),
		WeekDays:  sch.WeekDays,
		Hour:      sch.Hour,
		Minute:    sch.Minute,
		Enabled:   sch.Enabled,
		Creator:   sch.Creator,
		CreatedAt: sch.CreatedAt,
		UpdatedAt: sch.UpdatedAt,
	})
}

func (h *ScheduleHandler) DeleteSchedule(c echo.Context) error {
	ctx := c.Request().Context()
	userID := auth.GetUserID(c)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid schedule ID"})
	}

	// Check if schedule exists and user has permission
	existing, err := h.useCase.GetSchedule(ctx, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Schedule not found"})
	}

	// Only creator can delete (or admin)
	if existing.Creator != userID && auth.GetUserRole(c) != auth.RoleAdmin {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Permission denied"})
	}

	if err := h.useCase.DeleteSchedule(ctx, id); err != nil {
		log.Error().Err(err).Msg("Failed to delete schedule")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete schedule"})
	}

	log.Info().Str("schedule_id", id.String()).Str("user_id", userID).Msg("Schedule deleted")

	return c.JSON(http.StatusOK, map[string]string{"message": "Schedule deleted"})
}

func (h *ScheduleHandler) ToggleSchedule(c echo.Context) error {
	ctx := c.Request().Context()
	userID := auth.GetUserID(c)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid schedule ID"})
	}

	// Check if schedule exists
	existing, err := h.useCase.GetSchedule(ctx, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Schedule not found"})
	}

	// Only creator can toggle (or admin)
	if existing.Creator != userID && auth.GetUserRole(c) != auth.RoleAdmin {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Permission denied"})
	}

	sch, err := h.useCase.ToggleSchedule(ctx, id, !existing.Enabled)
	if err != nil {
		log.Error().Err(err).Msg("Failed to toggle schedule")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to toggle schedule"})
	}

	log.Info().Str("schedule_id", sch.ID.String()).Bool("enabled", sch.Enabled).Msg("Schedule toggled")

	return c.JSON(http.StatusOK, ScheduleResponse{
		ID:        sch.ID.String(),
		Name:      sch.Name,
		AppID:     sch.AppID,
		AppTitle:  sch.AppTitle,
		Operation: string(sch.Operation),
		WeekDays:  sch.WeekDays,
		Hour:      sch.Hour,
		Minute:    sch.Minute,
		Enabled:   sch.Enabled,
		Creator:   sch.Creator,
		CreatedAt: sch.CreatedAt,
		UpdatedAt: sch.UpdatedAt,
	})
}
