package mcpserver

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"apps-scheduler/internal/auth"
	"apps-scheduler/internal/biz"
	"apps-scheduler/internal/ent"
	"apps-scheduler/internal/ent/schedule"
	"apps-scheduler/internal/tokens"
	"apps-scheduler/internal/version"

	gohelper "gitee.com/linakesi/lzc-sdk/lang/go"
	"gitee.com/linakesi/lzc-sdk/lang/go/sys"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/metadata"
)

const (
	defaultCheckIntervalMinutes = 5
	minCheckIntervalMinutes     = 1
	maxCheckIntervalMinutes     = 1440
)

type principal struct {
	userID string
	role   string
	ticket string
}

type server struct {
	useCase *biz.UseCase
}

type principalContextKey struct{}

func NewHTTPHandler(useCase *biz.UseCase) http.Handler {
	s := &server{useCase: useCase}
	mcpServer := s.newMCPServer()
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return mcpServer
	}, &mcp.StreamableHTTPOptions{
		Stateless:                  true,
		JSONResponse:               true,
		DisableLocalhostProtection: true,
	})

	return s.authenticate(handler)
}

func (s *server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, err := s.principalFromRequest(r)
		if err != nil {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), principalContextKey{}, p)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *server) principalFromRequest(r *http.Request) (principal, error) {
	if token, ok := bearerToken(r.Header.Get("Authorization")); ok {
		item, err := s.useCase.GetActiveMCPTokenByHash(r.Context(), tokens.Hash(token))
		if err != nil {
			return principal{}, fmt.Errorf("invalid bearer token")
		}
		if err := s.useCase.TouchMCPToken(r.Context(), item.ID); err != nil {
			log.Warn().Err(err).Str("token_id", item.ID.String()).Msg("Failed to update MCP token last used time")
		}
		return principal{
			userID: item.UserID,
			role:   item.UserRole,
		}, nil
	}

	ticket := r.Header.Get("X-HC-USER-TICKET")
	userID := r.Header.Get("X-HC-USER-ID")
	if ticket == "" || userID == "" {
		return principal{}, fmt.Errorf("missing bearer token or LazyCat user ticket")
	}

	role := auth.RoleUser
	if isAdminHeader(r.Header.Get("X-HC-USER-IS-ADMIN")) || strings.EqualFold(r.Header.Get("X-HC-USER-ROLE"), auth.RoleAdmin) {
		role = auth.RoleAdmin
	}
	return principal{
		userID: userID,
		role:   role,
		ticket: ticket,
	}, nil
}

func (s *server) newMCPServer() *mcp.Server {
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "apps-scheduler",
		Title:   "App Scheduler",
		Version: version.Version,
	}, &mcp.ServerOptions{
		Instructions: "Manage LazyCat app schedules. Mutating tools operate as the authenticated LazyCat user.",
	})

	readOnly := true
	closedWorld := false
	destructive := true
	nonDestructive := false

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "list_apps",
		Title:       "List LazyCat apps",
		Description: "List installed non-built-in LazyCat apps that can be scheduled or operated.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, OpenWorldHint: &readOnly},
	}, s.listApps)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "list_schedules",
		Title:       "List schedules",
		Description: "List app automation schedules visible to the authenticated user.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, OpenWorldHint: &closedWorld},
	}, s.listSchedules)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "create_schedule",
		Title:       "Create schedule",
		Description: "Create a resume, pause, or keep-running schedule for an app.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &nonDestructive, OpenWorldHint: &closedWorld},
	}, s.createSchedule)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "update_schedule",
		Title:       "Update schedule",
		Description: "Replace the editable fields of an existing schedule.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive, OpenWorldHint: &closedWorld},
	}, s.updateSchedule)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "set_schedule_enabled",
		Title:       "Enable or disable schedule",
		Description: "Set whether a schedule is enabled.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true, OpenWorldHint: &closedWorld},
	}, s.setScheduleEnabled)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "delete_schedule",
		Title:       "Delete schedule",
		Description: "Delete an existing schedule owned by the token user, or any schedule for admin tokens.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive, OpenWorldHint: &closedWorld},
	}, s.deleteSchedule)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "resume_app",
		Title:       "Resume app",
		Description: "Resume an installed LazyCat app for the authenticated user.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &nonDestructive, OpenWorldHint: &readOnly},
	}, s.resumeApp)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "pause_app",
		Title:       "Pause app",
		Description: "Pause an installed LazyCat app for the authenticated user.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive, OpenWorldHint: &readOnly},
	}, s.pauseApp)

	return mcpServer
}

type listAppsInput struct{}

type appInfo struct {
	AppID          string `json:"appId" jsonschema:"LazyCat application package ID"`
	Title          string `json:"title" jsonschema:"display title"`
	Icon           string `json:"icon,omitempty" jsonschema:"icon URL"`
	Version        string `json:"version,omitempty" jsonschema:"installed version"`
	Status         string `json:"status" jsonschema:"install status"`
	InstanceStatus string `json:"instanceStatus" jsonschema:"runtime instance status"`
	MultiInstance  bool   `json:"multiInstance" jsonschema:"whether this is a multi-instance app"`
}

type listAppsOutput struct {
	Apps []appInfo `json:"apps" jsonschema:"installed non-built-in apps"`
}

func (s *server) listApps(ctx context.Context, _ *mcp.CallToolRequest, _ listAppsInput) (*mcp.CallToolResult, listAppsOutput, error) {
	p, err := principalFromContext(ctx)
	if err != nil {
		return nil, listAppsOutput{}, err
	}

	ctx, cancel := lazycatSDKContext(ctx, p, 10*time.Second)
	defer cancel()

	gw, err := gohelper.NewAPIGateway(ctx)
	if err != nil {
		return nil, listAppsOutput{}, fmt.Errorf("connect LazyCat gateway: %w", err)
	}
	defer gw.Close()

	resp, err := gw.PkgManager.QueryApplication(ctx, &sys.QueryApplicationRequest{})
	if err != nil {
		return nil, listAppsOutput{}, fmt.Errorf("query applications: %w", err)
	}

	apps := make([]appInfo, 0, len(resp.InfoList))
	for _, info := range resp.InfoList {
		if info.Status != sys.AppStatus_Installed {
			continue
		}
		if info.Builtin != nil && *info.Builtin {
			continue
		}

		app := appInfo{
			AppID:          info.Appid,
			Title:          info.Appid,
			Status:         info.Status.String(),
			InstanceStatus: info.InstanceStatus.String(),
			MultiInstance:  info.MultiInstance,
		}
		if info.Title != nil {
			app.Title = *info.Title
		}
		if info.Icon != nil {
			app.Icon = *info.Icon
		}
		if info.Version != nil {
			app.Version = *info.Version
		}
		apps = append(apps, app)
	}

	return nil, listAppsOutput{Apps: apps}, nil
}

type listSchedulesInput struct {
	IncludeAll bool `json:"includeAll,omitempty" jsonschema:"admin tokens only: include schedules from every user"`
}

type scheduleInfo struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	AppID                string    `json:"appId"`
	AppTitle             string    `json:"appTitle"`
	Operation            string    `json:"operation"`
	WeekDays             []int     `json:"weekDays"`
	Hour                 int       `json:"hour"`
	Minute               int       `json:"minute"`
	CheckIntervalMinutes int       `json:"checkIntervalMinutes"`
	Enabled              bool      `json:"enabled"`
	Creator              string    `json:"creator"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

type schedulesOutput struct {
	Schedules []scheduleInfo `json:"schedules"`
}

func (s *server) listSchedules(ctx context.Context, _ *mcp.CallToolRequest, input listSchedulesInput) (*mcp.CallToolResult, schedulesOutput, error) {
	p, err := principalFromContext(ctx)
	if err != nil {
		return nil, schedulesOutput{}, err
	}

	var items []*ent.Schedule
	if input.IncludeAll && p.role == auth.RoleAdmin {
		items, err = s.useCase.ListSchedules(ctx)
	} else {
		items, err = s.useCase.ListSchedulesByUser(ctx, p.userID)
	}
	if err != nil {
		return nil, schedulesOutput{}, fmt.Errorf("list schedules: %w", err)
	}

	out := schedulesOutput{Schedules: make([]scheduleInfo, 0, len(items))}
	for _, item := range items {
		out.Schedules = append(out.Schedules, scheduleOutput(item))
	}
	return nil, out, nil
}

type scheduleInput struct {
	Name                 string `json:"name" jsonschema:"schedule name"`
	AppID                string `json:"appId" jsonschema:"LazyCat app package ID"`
	AppTitle             string `json:"appTitle,omitempty" jsonschema:"human-readable app title"`
	Operation            string `json:"operation" jsonschema:"one of: resume, pause, keep_running"`
	WeekDays             []int  `json:"weekDays,omitempty" jsonschema:"weekdays for timed tasks, 0=Sunday through 6=Saturday"`
	Hour                 int    `json:"hour,omitempty" jsonschema:"hour for timed tasks, 0-23"`
	Minute               int    `json:"minute,omitempty" jsonschema:"minute for timed tasks, 0-59"`
	CheckIntervalMinutes int    `json:"checkIntervalMinutes,omitempty" jsonschema:"keep_running check interval, 1-1440 minutes"`
}

type scheduleOutputWrapper struct {
	Schedule scheduleInfo `json:"schedule"`
}

func (s *server) createSchedule(ctx context.Context, _ *mcp.CallToolRequest, input scheduleInput) (*mcp.CallToolResult, scheduleOutputWrapper, error) {
	p, err := principalFromContext(ctx)
	if err != nil {
		return nil, scheduleOutputWrapper{}, err
	}
	normalized, err := normalizeScheduleInput(input)
	if err != nil {
		return nil, scheduleOutputWrapper{}, err
	}

	item, err := s.useCase.CreateSchedule(ctx, normalized.Name, normalized.AppID, normalized.AppTitle, normalized.Operation, p.userID, normalized.WeekDays, normalized.Hour, normalized.Minute, normalized.CheckIntervalMinutes)
	if err != nil {
		return nil, scheduleOutputWrapper{}, fmt.Errorf("create schedule: %w", err)
	}
	return nil, scheduleOutputWrapper{Schedule: scheduleOutput(item)}, nil
}

type updateScheduleInput struct {
	ID                   string `json:"id" jsonschema:"schedule UUID"`
	Name                 string `json:"name" jsonschema:"schedule name"`
	AppID                string `json:"appId" jsonschema:"LazyCat app package ID"`
	AppTitle             string `json:"appTitle,omitempty" jsonschema:"human-readable app title"`
	Operation            string `json:"operation" jsonschema:"one of: resume, pause, keep_running"`
	WeekDays             []int  `json:"weekDays,omitempty" jsonschema:"weekdays for timed tasks, 0=Sunday through 6=Saturday"`
	Hour                 int    `json:"hour,omitempty" jsonschema:"hour for timed tasks, 0-23"`
	Minute               int    `json:"minute,omitempty" jsonschema:"minute for timed tasks, 0-59"`
	CheckIntervalMinutes int    `json:"checkIntervalMinutes,omitempty" jsonschema:"keep_running check interval, 1-1440 minutes"`
	Enabled              bool   `json:"enabled" jsonschema:"whether the schedule is enabled"`
}

func (s *server) updateSchedule(ctx context.Context, _ *mcp.CallToolRequest, input updateScheduleInput) (*mcp.CallToolResult, scheduleOutputWrapper, error) {
	p, err := principalFromContext(ctx)
	if err != nil {
		return nil, scheduleOutputWrapper{}, err
	}
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, scheduleOutputWrapper{}, fmt.Errorf("invalid schedule id")
	}
	existing, err := s.useCase.GetSchedule(ctx, id)
	if err != nil {
		return nil, scheduleOutputWrapper{}, fmt.Errorf("schedule not found: %w", err)
	}
	if !canManageSchedule(existing, p) {
		return nil, scheduleOutputWrapper{}, fmt.Errorf("permission denied")
	}

	normalized, err := normalizeScheduleInput(scheduleInput{
		Name:                 input.Name,
		AppID:                input.AppID,
		AppTitle:             input.AppTitle,
		Operation:            input.Operation,
		WeekDays:             input.WeekDays,
		Hour:                 input.Hour,
		Minute:               input.Minute,
		CheckIntervalMinutes: input.CheckIntervalMinutes,
	})
	if err != nil {
		return nil, scheduleOutputWrapper{}, err
	}

	item, err := s.useCase.UpdateSchedule(ctx, id, normalized.Name, normalized.AppID, normalized.AppTitle, normalized.Operation, normalized.WeekDays, normalized.Hour, normalized.Minute, normalized.CheckIntervalMinutes, input.Enabled)
	if err != nil {
		return nil, scheduleOutputWrapper{}, fmt.Errorf("update schedule: %w", err)
	}
	return nil, scheduleOutputWrapper{Schedule: scheduleOutput(item)}, nil
}

type setScheduleEnabledInput struct {
	ID      string `json:"id" jsonschema:"schedule UUID"`
	Enabled bool   `json:"enabled" jsonschema:"desired enabled state"`
}

func (s *server) setScheduleEnabled(ctx context.Context, _ *mcp.CallToolRequest, input setScheduleEnabledInput) (*mcp.CallToolResult, scheduleOutputWrapper, error) {
	p, err := principalFromContext(ctx)
	if err != nil {
		return nil, scheduleOutputWrapper{}, err
	}
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, scheduleOutputWrapper{}, fmt.Errorf("invalid schedule id")
	}
	existing, err := s.useCase.GetSchedule(ctx, id)
	if err != nil {
		return nil, scheduleOutputWrapper{}, fmt.Errorf("schedule not found: %w", err)
	}
	if !canManageSchedule(existing, p) {
		return nil, scheduleOutputWrapper{}, fmt.Errorf("permission denied")
	}

	item, err := s.useCase.ToggleSchedule(ctx, id, input.Enabled)
	if err != nil {
		return nil, scheduleOutputWrapper{}, fmt.Errorf("set schedule enabled: %w", err)
	}
	return nil, scheduleOutputWrapper{Schedule: scheduleOutput(item)}, nil
}

type idInput struct {
	ID string `json:"id" jsonschema:"schedule UUID"`
}

type deleteScheduleOutput struct {
	Deleted bool   `json:"deleted"`
	ID      string `json:"id"`
}

func (s *server) deleteSchedule(ctx context.Context, _ *mcp.CallToolRequest, input idInput) (*mcp.CallToolResult, deleteScheduleOutput, error) {
	p, err := principalFromContext(ctx)
	if err != nil {
		return nil, deleteScheduleOutput{}, err
	}
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, deleteScheduleOutput{}, fmt.Errorf("invalid schedule id")
	}
	existing, err := s.useCase.GetSchedule(ctx, id)
	if err != nil {
		return nil, deleteScheduleOutput{}, fmt.Errorf("schedule not found: %w", err)
	}
	if !canManageSchedule(existing, p) {
		return nil, deleteScheduleOutput{}, fmt.Errorf("permission denied")
	}
	if err := s.useCase.DeleteSchedule(ctx, id); err != nil {
		return nil, deleteScheduleOutput{}, fmt.Errorf("delete schedule: %w", err)
	}
	return nil, deleteScheduleOutput{Deleted: true, ID: id.String()}, nil
}

type appOperationInput struct {
	AppID string `json:"appId" jsonschema:"LazyCat app package ID"`
}

type appOperationOutput struct {
	AppID   string `json:"appId"`
	Message string `json:"message"`
}

func (s *server) resumeApp(ctx context.Context, _ *mcp.CallToolRequest, input appOperationInput) (*mcp.CallToolResult, appOperationOutput, error) {
	p, err := principalFromContext(ctx)
	if err != nil {
		return nil, appOperationOutput{}, err
	}
	if input.AppID == "" {
		return nil, appOperationOutput{}, fmt.Errorf("appId is required")
	}

	ctx, cancel := lazycatSDKContext(ctx, p, 30*time.Second)
	defer cancel()

	gw, err := gohelper.NewAPIGateway(ctx)
	if err != nil {
		return nil, appOperationOutput{}, fmt.Errorf("connect LazyCat gateway: %w", err)
	}
	defer gw.Close()

	if _, err := gw.PkgManager.Resume(ctx, &sys.AppInstance{Appid: input.AppID, Uid: p.userID}); err != nil {
		return nil, appOperationOutput{}, fmt.Errorf("resume app: %w", err)
	}
	return nil, appOperationOutput{AppID: input.AppID, Message: "Application resume requested"}, nil
}

func (s *server) pauseApp(ctx context.Context, _ *mcp.CallToolRequest, input appOperationInput) (*mcp.CallToolResult, appOperationOutput, error) {
	p, err := principalFromContext(ctx)
	if err != nil {
		return nil, appOperationOutput{}, err
	}
	if input.AppID == "" {
		return nil, appOperationOutput{}, fmt.Errorf("appId is required")
	}

	ctx, cancel := lazycatSDKContext(ctx, p, 30*time.Second)
	defer cancel()

	gw, err := gohelper.NewAPIGateway(ctx)
	if err != nil {
		return nil, appOperationOutput{}, fmt.Errorf("connect LazyCat gateway: %w", err)
	}
	defer gw.Close()

	if _, err := gw.PkgManager.Pause(ctx, &sys.AppInstance{Appid: input.AppID, Uid: p.userID}); err != nil {
		return nil, appOperationOutput{}, fmt.Errorf("pause app: %w", err)
	}
	return nil, appOperationOutput{AppID: input.AppID, Message: "Application pause requested"}, nil
}

func principalFromContext(ctx context.Context) (principal, error) {
	p, ok := ctx.Value(principalContextKey{}).(principal)
	if !ok || p.userID == "" {
		return principal{}, fmt.Errorf("missing MCP identity")
	}
	return p, nil
}

func canManageSchedule(item *ent.Schedule, p principal) bool {
	return item.Creator == p.userID || p.role == auth.RoleAdmin
}

func bearerToken(header string) (string, bool) {
	fields := strings.Fields(header)
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") || fields[1] == "" {
		return "", false
	}
	return fields[1], true
}

func isAdminHeader(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y":
		return true
	default:
		return false
	}
}

func lazycatSDKContext(ctx context.Context, p principal, timeout time.Duration) (context.Context, context.CancelFunc) {
	pairs := []string{
		"x-hc-user-id", p.userID,
		"x-hc-user-role", p.role,
	}
	if p.ticket != "" {
		pairs = append(pairs, "x-hc-user-ticket", p.ticket)
	}
	return context.WithTimeout(metadata.AppendToOutgoingContext(ctx, pairs...), timeout)
}

func normalizeScheduleInput(input scheduleInput) (scheduleInput, error) {
	if input.Name == "" || input.AppID == "" || input.Operation == "" {
		return input, fmt.Errorf("name, appId, and operation are required")
	}
	if input.AppTitle == "" {
		input.AppTitle = input.AppID
	}

	switch input.Operation {
	case string(schedule.OperationResume), string(schedule.OperationPause):
		if len(input.WeekDays) == 0 {
			return input, fmt.Errorf("at least one weekday is required")
		}
		for _, day := range input.WeekDays {
			if day < 0 || day > 6 {
				return input, fmt.Errorf("weekDays values must be between 0 and 6")
			}
		}
		if input.Hour < 0 || input.Hour > 23 {
			return input, fmt.Errorf("hour must be between 0 and 23")
		}
		if input.Minute < 0 || input.Minute > 59 {
			return input, fmt.Errorf("minute must be between 0 and 59")
		}
		input.CheckIntervalMinutes = defaultCheckIntervalMinutes
	case string(schedule.OperationKeepRunning):
		input.WeekDays = []int{}
		input.Hour = 0
		input.Minute = 0
		if input.CheckIntervalMinutes == 0 {
			input.CheckIntervalMinutes = defaultCheckIntervalMinutes
		}
		if input.CheckIntervalMinutes < minCheckIntervalMinutes || input.CheckIntervalMinutes > maxCheckIntervalMinutes {
			return input, fmt.Errorf("checkIntervalMinutes must be between 1 and 1440")
		}
	default:
		return input, fmt.Errorf("operation must be 'resume', 'pause', or 'keep_running'")
	}
	return input, nil
}

func scheduleOutput(item *ent.Schedule) scheduleInfo {
	return scheduleInfo{
		ID:                   item.ID.String(),
		Name:                 item.Name,
		AppID:                item.AppID,
		AppTitle:             item.AppTitle,
		Operation:            string(item.Operation),
		WeekDays:             item.WeekDays,
		Hour:                 item.Hour,
		Minute:               item.Minute,
		CheckIntervalMinutes: item.CheckIntervalMinutes,
		Enabled:              item.Enabled,
		Creator:              item.Creator,
		CreatedAt:            item.CreatedAt,
		UpdatedAt:            item.UpdatedAt,
	}
}
