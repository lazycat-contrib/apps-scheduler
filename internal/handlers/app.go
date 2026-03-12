package handlers

import (
	"context"
	"net/http"

	"apps-scheduler/internal/auth"

	gohelper "gitee.com/linakesi/lzc-sdk/lang/go"
	"gitee.com/linakesi/lzc-sdk/lang/go/sys"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/metadata"
)

type AppHandler struct{}

func NewAppHandler() *AppHandler {
	return &AppHandler{}
}

type AppInfo struct {
	AppID          string `json:"appId"`
	DeployID       string `json:"deployId"`
	Title          string `json:"title"`
	Icon           string `json:"icon"`
	Version        string `json:"version"`
	Status         string `json:"status"`
	InstanceStatus string `json:"instanceStatus"`
	MultiInstance  bool   `json:"multiInstance"`
}

func (h *AppHandler) ListApps(c echo.Context) error {
	ctx := context.Background()
	userID := auth.GetUserID(c)

	// Add user ID to context as required by SDK
	ctx = metadata.AppendToOutgoingContext(ctx, "x-hc-user-id", userID)

	gw, err := gohelper.NewAPIGateway(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create API gateway")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect to gateway"})
	}
	defer gw.Close()

	resp, err := gw.PkgManager.QueryApplication(ctx, &sys.QueryApplicationRequest{})
	if err != nil {
		log.Error().Err(err).Msg("Failed to query applications")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to query applications"})
	}

	apps := make([]AppInfo, 0, len(resp.InfoList))
	for _, info := range resp.InfoList {
		// Only include installed apps
		if info.Status != sys.AppStatus_Installed {
			continue
		}

		// Filter out preinstalled apps (Builtin field)
		if info.Builtin != nil && *info.Builtin {
			log.Debug().Str("app_id", info.Appid).Msg("Skipping builtin/preinstalled app")
			continue
		}

		// Log app info for debugging
		log.Debug().
			Str("app_id", info.Appid).
			Str("status", info.Status.String()).
			Str("instance_status", info.InstanceStatus.String()).
			Msg("Including app")

		app := AppInfo{
			AppID:          info.Appid,
			DeployID:       info.Appid, // Use AppID as DeployID since SDK doesn't provide separate DeployId
			Status:         info.Status.String(),
			InstanceStatus: info.InstanceStatus.String(),
			MultiInstance:  info.MultiInstance,
		}

		if info.Title != nil {
			app.Title = *info.Title
		} else {
			app.Title = info.Appid
		}

		if info.Icon != nil {
			app.Icon = *info.Icon
		}

		if info.Version != nil {
			app.Version = *info.Version
		}

		apps = append(apps, app)
	}

	log.Debug().Str("user_id", userID).Int("count", len(apps)).Msg("Listed applications")

	return c.JSON(http.StatusOK, apps)
}

func (h *AppHandler) ResumeApp(c echo.Context) error {
	appID := c.Param("appId")
	userID := auth.GetUserID(c)

	if appID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "App ID is required"})
	}

	ctx := context.Background()
	ctx = metadata.AppendToOutgoingContext(ctx, "x-hc-user-id", userID)

	gw, err := gohelper.NewAPIGateway(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create API gateway")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect to gateway"})
	}
	defer gw.Close()

	_, err = gw.PkgManager.Resume(ctx, &sys.AppInstance{
		Appid: appID,
		Uid:   userID,
	})
	if err != nil {
		log.Error().Err(err).Str("app_id", appID).Msg("Failed to resume application")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to resume application"})
	}

	log.Info().Str("app_id", appID).Str("user_id", userID).Msg("Application resumed")

	return c.JSON(http.StatusOK, map[string]string{"message": "Application resumed"})
}

func (h *AppHandler) PauseApp(c echo.Context) error {
	appID := c.Param("appId")
	userID := auth.GetUserID(c)

	if appID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "App ID is required"})
	}

	ctx := context.Background()
	ctx = metadata.AppendToOutgoingContext(ctx, "x-hc-user-id", userID)

	gw, err := gohelper.NewAPIGateway(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create API gateway")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect to gateway"})
	}
	defer gw.Close()

	_, err = gw.PkgManager.Pause(ctx, &sys.AppInstance{
		Appid: appID,
		Uid:   userID,
	})
	if err != nil {
		log.Error().Err(err).Str("app_id", appID).Msg("Failed to pause application")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to pause application"})
	}

	log.Info().Str("app_id", appID).Str("user_id", userID).Msg("Application paused")

	return c.JSON(http.StatusOK, map[string]string{"message": "Application paused"})
}
