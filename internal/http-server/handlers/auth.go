package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	internalErr = "smth wrong"
)

type inputSignUp struct {
	FingerPrint string `json:"finger_print" binding:"required"`
}

func (h *Handler) signUp(c *gin.Context) {
	const op = "http-server.handlers.signUp"

	var input inputSignUp
	if err := c.BindJSON(&input); err != nil {
		zap.S().Warn(fmt.Errorf("%s:%w", op, err))
		newErrorResponse(c, http.StatusBadRequest, errBindingJSON.Error())
		return
	}

	timeCode, err := h.service.Auth.CreateRegSession(c, input.FingerPrint)
	if err != nil {
		zap.S().Warn(fmt.Errorf("%s:%w", op, err))
		newErrorResponse(c, http.StatusInternalServerError, internalErr)
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"time_code": timeCode,
	})
}

type inputSignIn struct {
	FingerPrint string `json:"finger_print" binding:"required"`
	TimeCode    string `json:"time_code" binding:"required"`
}

type outputSignIn struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) signIn(c *gin.Context) {
	const op = "http-server.handlers.signIn"

	var input inputSignIn
	if err := c.BindJSON(&input); err != nil {
		zap.S().Warn(fmt.Errorf("%s:%w", op, err))
		newErrorResponse(c, http.StatusBadRequest, errBindingJSON.Error())
		return
	}

	userID, err := h.service.Auth.GetRegSession(c, input.FingerPrint, input.TimeCode)
	if err != nil {
		zap.S().Warn(fmt.Errorf("%s:%w", op, err))
		newErrorResponse(c, http.StatusInternalServerError, internalErr)
		return
	}

	accessToken, err := h.jwtManager.GenerateToken(userID, "user")
	refreshToken, err := h.jwtManager.NewRefreshToken()

	err = h.service.InitUser(c, userID, refreshToken)
	if err != nil {
		zap.S().Warn(fmt.Errorf("%s:%w", op, err))
		newErrorResponse(c, http.StatusInternalServerError, internalErr)
		return
	}

	c.JSON(http.StatusOK, outputSignIn{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *Handler) refresh(c *gin.Context) {

	// todo implement me
}
