package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	log_middlewares "github.com/UdinSemen/moscow-events-backend/internal/http-server/log-middlewares"
	"github.com/UdinSemen/moscow-events-backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	internalErr  = "smth wrong"
	timeOutWs    = 3
	errTimeout   = "timeout"
	opPrefixAuth = "http-server.handlers."
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
	const op = opPrefixAuth + "signUp"

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

	zap.S().Info(c.RemoteIP())
	var input inputSignIn

	if err := c.BindJSON(&input); err != nil {
		zap.S().Warn(fmt.Errorf("%s:%w", op, err))
		newErrorResponse(c, http.StatusBadRequest, errBindingJSON.Error())
		return
	}

	zap.S().Info(input)

	userID, err := h.service.Auth.GetRegSession(c, input.FingerPrint, input.TimeCode)
	if err != nil {
		zap.S().Warn(fmt.Errorf("%s:%w", op, err))
		newErrorResponse(c, http.StatusInternalServerError, internalErr)
		return
	}

	accessToken, err := h.jwtManager.GenerateToken(userID, "user")
	refreshToken, err := h.jwtManager.NewRefreshToken()

	err = h.service.InitUser(c, userID, refreshToken, c.RemoteIP(), input.FingerPrint)
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
	const op = opPrefixAuth + "signInWebSocket"

	reqId, ok := c.Get(log_middlewares.RequestIDCtx)
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

	userChanel := make(chan string)
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
			userID, err := h.service.Auth.GetRegSession(c, input.FingerPrint, input.TimeCode)
			if err != nil {
				counter++
				if errors.Is(err, services.ErrSessionNotConfirmed) {
					// todo
					if counter == timeOutWs {
						zap.L().Warn(op,
							zap.Error(err),
							zap.String("user_id", userID),
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
				if userID != "" {
					userChanel <- userID
				}
			}
			time.Sleep(time.Second)
		}
	}()

	userID := <-userChanel
	accessToken, err := h.jwtManager.GenerateToken(userID, "user")
	refreshToken, err := h.jwtManager.NewRefreshToken()

	err = h.service.InitUser(c, userID, refreshToken, c.RemoteIP(), input.FingerPrint)
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
			zap.S().Error(fmt.Errorf("%s:%w", op, err))
		}
	}
	zap.S().Error(fmt.Errorf("%s:%v", op, con.Close()))
}

type inputRefresh struct {
	RefreshToken string `json:"refresh_token"`
	FingerPrint  string `json:"finger_print"`
}

func (h *Handler) refresh(c *gin.Context) {
	const op = opPrefixAuth + "refresh"

	var input inputRefresh
	if err := c.BindJSON(&input); err != nil {
		zap.S().Warn(fmt.Errorf("%s:%w", op, err))
		newErrorResponse(c, http.StatusBadRequest, errBindingJSON.Error())
		return
	}

	// todo refresh token

}
