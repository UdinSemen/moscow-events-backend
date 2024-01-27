package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	opPrefixAuthMiddleware = "http-server.handlers.auth-middleware."
	AuthHeader             = "Authorization"
	UserCtx                = "userCtx"
	okayAuth               = "user: %s success authed role: %s"
	invalidAuth            = "user: %s invalid authed role: %s"
	emptyAuthHeader        = "empty auth header"
	invalidAuthHeader      = "invalid auth header"
	invalidToken           = "invalid access token"
)

func (h *Handler) userIdentity(c *gin.Context) {
	const op = opPrefixAuthMiddleware + "userIdentity"

	header := c.GetHeader(AuthHeader)
	if header == "" {
		newErrorResponse(c, http.StatusUnauthorized, emptyAuthHeader)
		return
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		newErrorResponse(c, http.StatusUnauthorized, invalidAuthHeader)
		return
	}

	user, err := h.jwtManager.ParseToken(headerParts[1])
	if err != nil {
		zap.S().Infof(fmt.Sprintf(invalidAuth, user.Uuid, user.Role))
		zap.S().Warn(fmt.Errorf("%s:%w", op, err))
		newErrorResponse(c, http.StatusUnauthorized, invalidToken)
		return
	}

	zap.S().Infof(fmt.Sprintf(okayAuth, user.Uuid, user.Role))
	c.Set(UserCtx, user)
}
