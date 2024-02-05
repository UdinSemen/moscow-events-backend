package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/UdinSemen/moscow-events-backend/internal/domain/models"
	logmiddlewares "github.com/UdinSemen/moscow-events-backend/internal/http-server/log-middlewares"
	storage "github.com/UdinSemen/moscow-events-backend/internal/storage/postgres"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	NothingWasFound = "nothing was found"
)

type inputGetEvent struct {
	Category string      `json:"category"`
	Date     []time.Time `json:"date"`
}

type outputGetEvent struct {
	Events []models.Event `json:"events"`
}

func (h *Handler) getEvent(c *gin.Context) {
	const op = opPrefixHandlers + "getEvent"

	reqId, ok := c.Get(logmiddlewares.RequestIDCtx)
	if !ok {
		zap.S().Errorf("%s:%v", op, ErrReqIdNotExist)
	}

	userDTO, err := getUserDTOFromCtx(c)
	if err != nil {
		zap.L().Error(op,
			zap.Error(err),
			zap.Any(nameFieldReqIDLog, reqId),
		)
		return
	}

	var input inputGetEvent
	if err := c.BindJSON(&input); err != nil {
		zap.L().Warn(op,
			zap.Error(err),
			zap.Any(nameFieldReqIDLog, reqId),
		)
		newErrorResponse(c, http.StatusBadRequest, errBindingJSON.Error())
		return
	}

	events, err := h.service.Event.GetEvents(c, userDTO.Uuid, input.Category, input.Date)
	if err != nil {
		zap.L().Error(op,
			zap.Error(err),
			zap.Any(nameFieldReqIDLog, reqId),
		)
		if errors.Is(err, storage.ErrNoRows) {
			newErrorResponse(c, http.StatusOK, NothingWasFound)
			return
		}
		newErrorResponse(c, http.StatusInternalServerError, internalErr)
		return
	}

	c.JSON(http.StatusOK, outputGetEvent{
		Events: events,
	})
}
