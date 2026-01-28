package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

const (
	envOIDCClientID     = "LAZYCAT_AUTH_OIDC_CLIENT_ID"
	envOIDCClientSecret = "LAZYCAT_AUTH_OIDC_CLIENT_SECRET"
	envOIDCAuthURI      = "LAZYCAT_AUTH_OIDC_AUTH_URI"
	envOIDCTokenURI     = "LAZYCAT_AUTH_OIDC_TOKEN_URI"
	envOIDCUserInfoURI  = "LAZYCAT_AUTH_OIDC_USERINFO_URI"
	envOIDCRedirectURL  = "LAZYCAT_AUTH_OIDC_REDIRECT_URL"
	envOIDCBasePath     = "LAZYCAT_AUTH_OIDC_BASE_PATH"
	envOIDCCallbackPath = "LAZYCAT_AUTH_OIDC_CALLBACK_PATH"
	envAppDomain        = "LAZYCAT_APP_DOMAIN"
)

const (
	defaultOIDCBasePath     = "/auth/oidc"
	defaultOIDCCallbackPath = "/auth/oidc/callback"
	defaultAppDomain        = "localhost:3000"
)

const (
	cookieUserID    = "user_id"
	cookieUserRole  = "user_role"
	cookieOIDCState = "oidc_state"
	cookieMaxAge    = 3600 * 24
	stateMaxAge     = 300
)

const (
	RoleUser  = "USER"
	RoleAdmin = "ADMIN"
)

type OIDCConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
}

type OIDCProvider struct {
	config      *oauth2.Config
	userInfoURL string
}

func NewOIDCProvider() (*OIDCProvider, error) {
	config, err := loadOIDCConfig()
	if err != nil {
		return nil, err
	}

	log.Info().Str("client_id", config.ClientID).Bool("secret_set", config.ClientSecret != "").Msg("OIDC configuration loaded")

	return createOIDCProvider(config)
}

func loadOIDCConfig() (*OIDCConfig, error) {
	clientID := os.Getenv(envOIDCClientID)
	clientSecret := os.Getenv(envOIDCClientSecret)
	authURL := os.Getenv(envOIDCAuthURI)
	tokenURL := os.Getenv(envOIDCTokenURI)
	userInfoURL := os.Getenv(envOIDCUserInfoURI)

	if clientID == "" || clientSecret == "" || authURL == "" || tokenURL == "" {
		return nil, errors.New("missing required OIDC configuration")
	}

	return &OIDCConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  buildRedirectURL(),
		AuthURL:      authURL,
		TokenURL:     tokenURL,
		UserInfoURL:  userInfoURL,
	}, nil
}

func createOIDCProvider(config *OIDCConfig) (*OIDCProvider, error) {
	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "groups"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.AuthURL,
			TokenURL: config.TokenURL,
		},
	}

	return &OIDCProvider{
		config:      oauth2Config,
		userInfoURL: config.UserInfoURL,
	}, nil
}

func buildRedirectURL() string {
	if redirectURL := os.Getenv(envOIDCRedirectURL); redirectURL != "" {
		return redirectURL
	}

	domain := getEnvOrDefault(envAppDomain, defaultAppDomain)
	callbackPath := getEnvOrDefault(envOIDCCallbackPath, defaultOIDCCallbackPath)

	return fmt.Sprintf("https://%s%s", domain, callbackPath)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetOIDCBasePath() string {
	return getEnvOrDefault(envOIDCBasePath, defaultOIDCBasePath)
}

func GetOIDCCallbackPath() string {
	return getEnvOrDefault(envOIDCCallbackPath, defaultOIDCCallbackPath)
}

func generateState() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func AuthMiddleware(oidcProvider *OIDCProvider) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check for authentication header (injected by system)
			if userID := c.Request().Header.Get("x-hc-user-id"); userID != "" {
				return next(c)
			}

			// Check for session-based authentication
			if userID := c.Get("user_id"); userID != nil && userID.(string) != "" {
				c.Response().Header().Set("x-hc-user-id", userID.(string))
				if userRole := c.Get("user_role"); userRole != nil {
					c.Response().Header().Set("x-hc-user-role", userRole.(string))
				}
				return next(c)
			}

			// Allow access to public paths
			if isPublicPath(c.Request().URL.Path) {
				return next(c)
			}

			// Redirect to login
			return redirectToLogin(c, oidcProvider)
		}
	}
}

func isPublicPath(path string) bool {
	oidcBasePath := GetOIDCBasePath()
	return strings.HasPrefix(path, "/login") ||
		strings.HasPrefix(path, oidcBasePath) ||
		strings.HasPrefix(path, "/static")
}

func redirectToLogin(c echo.Context, oidcProvider *OIDCProvider) error {
	if oidcProvider != nil {
		return c.Redirect(http.StatusFound, "/login")
	}
	return c.Redirect(http.StatusFound, "/login?error=oidc_not_configured")
}

func (p *OIDCProvider) HandleLogin(c echo.Context) error {
	state := generateState()
	c.SetCookie(&http.Cookie{
		Name:     cookieOIDCState,
		Value:    state,
		MaxAge:   stateMaxAge,
		Path:     "/",
		HttpOnly: true,
	})

	authURL := p.config.AuthCodeURL(state)
	return c.Redirect(http.StatusFound, authURL)
}

func (p *OIDCProvider) HandleCallback(c echo.Context) error {
	if err := p.validateState(c); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Clear state cookie
	c.SetCookie(&http.Cookie{
		Name:   cookieOIDCState,
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})

	code := c.QueryParam("code")
	if code == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing authorization code"})
	}

	ctx := context.Background()
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to exchange token"})
	}

	userInfo, err := p.fetchUserInfo(ctx, token)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to fetch user info")
	}

	userID, userRole := extractUserIdentity(userInfo)
	if userID == "" {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get user ID"})
	}

	// Set session cookies
	c.SetCookie(&http.Cookie{
		Name:     cookieUserID,
		Value:    userID,
		MaxAge:   cookieMaxAge,
		Path:     "/",
		HttpOnly: true,
	})
	c.SetCookie(&http.Cookie{
		Name:     cookieUserRole,
		Value:    userRole,
		MaxAge:   cookieMaxAge,
		Path:     "/",
		HttpOnly: true,
	})

	return c.Redirect(http.StatusFound, "/")
}

func (p *OIDCProvider) validateState(c echo.Context) error {
	cookie, err := c.Cookie(cookieOIDCState)
	if err != nil {
		return errors.New("missing state cookie")
	}

	if c.QueryParam("state") != cookie.Value {
		return errors.New("invalid state")
	}

	return nil
}

func (p *OIDCProvider) fetchUserInfo(ctx context.Context, token *oauth2.Token) (map[string]interface{}, error) {
	if p.userInfoURL == "" {
		return nil, errors.New("user info URL not configured")
	}

	client := p.config.Client(ctx, token)
	resp, err := client.Get(p.userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %w", err)
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return userInfo, nil
}

func extractUserIdentity(userInfo map[string]interface{}) (userID, userRole string) {
	userRole = RoleUser

	if userInfo == nil {
		return "", userRole
	}

	if sub, ok := userInfo["sub"].(string); ok {
		userID = sub
	} else if preferredUsername, ok := userInfo["preferred_username"].(string); ok {
		userID = preferredUsername
	}

	if groups, ok := userInfo["groups"].([]interface{}); ok {
		for _, group := range groups {
			if groupStr, ok := group.(string); ok && groupStr == RoleAdmin {
				userRole = RoleAdmin
				break
			}
		}
	}

	return userID, userRole
}

func HandleLogout(c echo.Context) error {
	c.SetCookie(&http.Cookie{
		Name:   cookieUserID,
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})
	c.SetCookie(&http.Cookie{
		Name:   cookieUserRole,
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})
	return c.Redirect(http.StatusFound, "/login")
}

func SessionMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if cookie, err := c.Cookie(cookieUserID); err == nil && cookie.Value != "" {
				c.Set("user_id", cookie.Value)
				if roleCookie, err := c.Cookie(cookieUserRole); err == nil && roleCookie.Value != "" {
					c.Set("user_role", roleCookie.Value)
				}
			}
			return next(c)
		}
	}
}

func GetUserID(c echo.Context) string {
	if userID := c.Request().Header.Get("x-hc-user-id"); userID != "" {
		return userID
	}
	if userID := c.Get("user_id"); userID != nil {
		return userID.(string)
	}
	return ""
}

func GetUserRole(c echo.Context) string {
	if userRole := c.Request().Header.Get("x-hc-user-role"); userRole != "" {
		return userRole
	}
	if userRole := c.Get("user_role"); userRole != nil {
		return userRole.(string)
	}
	return RoleUser
}
