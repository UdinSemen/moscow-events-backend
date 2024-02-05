package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/UdinSemen/moscow-events-backend/internal/domain/models"
	logmiddlewares "github.com/UdinSemen/moscow-events-backend/internal/http-server/log-middlewares"
	"github.com/UdinSemen/moscow-events-backend/internal/services"
	storage "github.com/UdinSemen/moscow-events-backend/internal/storage/postgres"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	internalErr          = "something wrong, repeat after"
	invalidFingerprint   = "invalid fingerprint"
	differentFingerprint = "different fingerprint"
	refreshTokenExpired  = "refresh token expired"
	notExistSession      = "not exist session"
	timeOutWs            = 3
	errTimeout           = "timeout"
	opPrefixHandlers     = "http-server.handlers."
	nameFieldReqIDLog    = "req_id"
)

var (
	connUpgrade = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	ErrReqIdNotExist = errors.New("request id not exist")
)

type inputSignUp struct {
	FingerPrint string `json:"finger_print" binding:"required"`
}

func (h *Handler) signUp(c *gin.Context) {
	const op = opPrefixHandlers + "signUp"

	// todo req id
	zap.S().Info(c.RemoteIP())
	var input inputSignUp
	if err := c.BindJSON(&input); err != nil {
		zap.S().Warn(fmt.Errorf("%s:%w", op, err))
		newErrorResponse(c, http.StatusBadRequest, errBindingJSON.Error())
		return
	}
	zap.S().Info(input)

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

	reqId, ok := c.Get(logmiddlewares.RequestIDCtx)
	if !ok {
		zap.S().Errorf("%s:%v", op, ErrReqIdNotExist)
	}

	zap.S().Info(c.RemoteIP())
	var input inputSignIn

	if err := c.BindJSON(&input); err != nil {
		zap.S().Warn(fmt.Errorf("%s:%w", op, err))
		newErrorResponse(c, http.StatusBadRequest, errBindingJSON.Error())
		return
	}

	zap.S().Info(input)

	userTgId, err := h.service.Auth.GetRegSession(c, input.FingerPrint, input.TimeCode)
	if err != nil {
		zap.S().Warn(fmt.Errorf("%s:%w", op, err))
		newErrorResponse(c, http.StatusInternalServerError, internalErr)
		return
	}

	userDTO, err := h.service.Auth.GetUserDTOByTg(c, userTgId)
	if err != nil {
		zap.L().Error(op,
			zap.Error(err),
			zap.String("user_tg_id", userTgId),
			zap.Any("req_id", reqId),
		)
		newErrorResponse(c, http.StatusInternalServerError, internalErr)
		return
	}

	accessToken, refreshToken, err := h.service.Auth.GenerateTokens(c, userDTO.Uuid, userDTO.Role)
	if err != nil {
		zap.L().Error(op,
			zap.Error(err),
			zap.String("user_id", userDTO.Uuid),
			zap.Any("req_id", reqId),
		)
		newErrorResponse(c, http.StatusInternalServerError, internalErr)
		return
	}

	err = h.service.InitSession(c, userTgId, refreshToken, c.RemoteIP(), input.FingerPrint)
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

func (h *Handler) signInWebSocket(c *gin.Context) {
	const op = opPrefixHandlers + "signInWebSocket"

	reqId, ok := c.Get(logmiddlewares.RequestIDCtx)
	if !ok {
		zap.S().Errorf("%s:%v", op, ErrReqIdNotExist)
	}

	con, err := connUpgrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zap.S().Warn(fmt.Errorf("%s:%w", op, err))
		newErrorResponse(c, http.StatusBadRequest, errBindingJSON.Error())
		return
	}
	zap.S().Info(con.RemoteAddr())

	var input inputSignIn
	if err := con.ReadJSON(&input); err != nil {
		zap.S().Warn(fmt.Errorf("%s:%w", op, err))

		if websocket.IsCloseError(err, websocket.CloseNormalClosure) {

			zap.L().Warn("normal closure",
				zap.Error(err),
				zap.Any("req_id", reqId),
			)
		}
		if err := newErrorWsResponse(con, websocket.CloseInternalServerErr, internalErr); err != nil {
			zap.S().Error(fmt.Errorf("%s:%w", op, err))
		}
		return
	}

	zap.S().Info(input)

	userChanel := make(chan models.UserDTO)
	go func() {
		counter := 1
		for {
			zap.S().Debug(counter)
			if counter > timeOutWs {
				if err := newErrorWsResponse(con, websocket.CloseTryAgainLater, errTimeout); err != nil {
					zap.S().Error(fmt.Errorf("%s:%w", op, err))
				}
				return
			}
			userTgId, err := h.service.Auth.GetRegSession(c, input.FingerPrint, input.TimeCode)
			if err != nil {
				counter++
				if errors.Is(err, services.ErrSessionNotConfirmed) {
					if counter == timeOutWs {
						zap.L().Warn(op,
							zap.Error(err),
							zap.String("user_tg_id", userTgId),
							zap.Any("req_id", reqId),
						)
					}
				} else {
					zap.S().Warn(fmt.Errorf("%s:%w", op, err))
					if err := newErrorWsResponse(con, websocket.CloseTryAgainLater, internalErr); err != nil {
						zap.S().Error(fmt.Errorf("%s:%w", op, err))
					}
					return
				}
			} else {
				if userTgId != "" {
					userDTO, err := h.service.Auth.GetUserDTOByTg(c, userTgId)
					if err != nil {
						zap.L().Error(op,
							zap.Error(err),
							zap.String("user_tg_id", userTgId),
							zap.Any("req_id", reqId),
						)
					}
					userChanel <- userDTO
				}
			}
			time.Sleep(time.Second)
		}
	}()

	userDTO := <-userChanel

	accessToken, refreshToken, err := h.service.Auth.GenerateTokens(c, userDTO.Uuid, userDTO.Role)
	if err != nil {
		zap.L().Error(op,
			zap.Error(err),
			zap.String("user_id", userDTO.Uuid),
			zap.Any("req_id", reqId),
		)
	}

	err = h.service.InitSession(c, userDTO.Uuid, refreshToken, c.RemoteIP(), input.FingerPrint)
	if err != nil {
		zap.S().Warn(fmt.Errorf("%s:%w", op, err))
		if err := con.WriteMessage(websocket.CloseInternalServerErr, []byte(internalErr)); err != nil {
			zap.S().Error(fmt.Errorf("%s:%w", op, err))
		}
		return
	}

	if err := con.WriteJSON(outputSignIn{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}); err != nil {
		zap.S().Warn(fmt.Errorf("%s:%w", op, err))
		if err := con.WriteMessage(websocket.CloseInternalServerErr, []byte(internalErr)); err != nil {
			zap.L().Error(op,
				zap.Error(err),
				zap.Any("req_id", reqId),
			)
		}
	}

	zap.S().Error(fmt.Errorf("%s:%v", op, con.Close()))
}

type inputRefresh struct {
	RefreshToken string `json:"refresh_token"`
	FingerPrint  string `json:"finger_print"`
}

type outputRefresh struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) refresh(c *gin.Context) {
	const op = opPrefixHandlers + "refresh"

	reqId, ok := c.Get(logmiddlewares.RequestIDCtx)
	if !ok {
		zap.S().Errorf("%s:%v", op, ErrReqIdNotExist)
	}

	var input inputRefresh
	if err := c.BindJSON(&input); err != nil {
		zap.L().Warn(op,
			zap.Error(err),
			zap.Any("req_id", reqId),
		)
		newErrorResponse(c, http.StatusBadRequest, errBindingJSON.Error())
		return
	}

	zap.S().Debug(input)

	accessToken, refreshToken, err := h.service.Auth.RefreshToken(c, input.RefreshToken, input.FingerPrint)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidFingerPrint):
			zap.L().Warn(op,
				zap.Error(err),
				zap.Any("req_id", reqId),
			)
			newErrorResponse(c, http.StatusBadRequest, invalidFingerprint)
			return
		case errors.Is(err, services.ErrDifferentFingerPrint):
			zap.L().Warn(op,
				zap.Error(err),
				zap.Any("req_id", reqId),
			)
			newErrorResponse(c, http.StatusBadRequest, differentFingerprint)
			return
		case errors.Is(err, services.ErrRefreshTokenExp):
			zap.L().Warn(op,
				zap.Error(err),
				zap.Any("req_id", reqId),
			)
			newErrorResponse(c, http.StatusBadRequest, refreshTokenExpired)
			return
		case errors.Is(err, storage.ErrNoRows):
			zap.L().Warn(op,
				zap.Error(err),
				zap.Any("req_id", reqId),
			)
			newErrorResponse(c, http.StatusBadRequest, notExistSession)
			return
		default:
			zap.L().Error(op,
				zap.Error(err),
				zap.Any("req_id", reqId),
			)
			newErrorResponse(c, http.StatusInternalServerError, internalErr)
			return
		}
	}

	if err := h.service.Auth.RefreshSession(c, input.RefreshToken, refreshToken, c.RemoteIP()); err != nil {
		zap.L().Error(op,
			zap.Error(err),
			zap.Any("req_id", reqId),
		)
		newErrorResponse(c, http.StatusInternalServerError, internalErr)
		return
	}

	c.JSON(http.StatusOK, outputRefresh{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}
