package handlers

import (
	"context"
	"net/http"

	"apps-scheduler/internal/auth"

	gohelper "gitee.com/linakesi/lzc-sdk/lang/go"
	"gitee.com/linakesi/lzc-sdk/lang/go/common"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/metadata"
)

type UserInfoHandler struct{}

func NewUserInfoHandler() *UserInfoHandler {
	return &UserInfoHandler{}
}

type UserInfoResponse struct {
	UserID   string `json:"userId"`
	UserRole string `json:"userRole"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
}

func (h *UserInfoHandler) GetUserInfo(c echo.Context) error {
	userID := auth.GetUserID(c)
	userRole := auth.GetUserRole(c)

	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Not authenticated"})
	}

	resp := UserInfoResponse{
		UserID:   userID,
		UserRole: userRole,
		Name:     userID,
		Avatar:   "",
	}

	// Try to get detailed user info from gateway
	ctx := context.Background()
	ctx = metadata.AppendToOutgoingContext(ctx, "x-hc-user-id", userID)

	gw, err := gohelper.NewAPIGateway(ctx)
	if err == nil {
		defer gw.Close()
		userInfo, err := gw.Users.QueryUserInfo(ctx, &common.UserID{Uid: userID})
		if err == nil && userInfo != nil {
			if userInfo.Nickname != "" {
				resp.Name = userInfo.Nickname
			}
			if userInfo.Avatar != "" {
				resp.Avatar = userInfo.Avatar
			}
		} else {
			log.Debug().Err(err).Str("user_id", userID).Msg("Failed to query user info from gateway")
		}
	}

	return c.JSON(http.StatusOK, resp)
}
