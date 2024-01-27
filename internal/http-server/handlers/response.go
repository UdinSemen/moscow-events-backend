package handlers

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	errBindingJSON = errors.New("invalid JSON")
)

type errorResponse struct {
	Message string `json:"message"`
}

type statusResponse struct {
	Status string `json:"status"`
}

func newErrorResponse(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, errorResponse{message})
}

func newErrorWsResponse(con *websocket.Conn, messageType int, message string) error {
	const op = "response.newErrorWsResponse"

	if err := con.WriteMessage(messageType, []byte(message)); err != nil {
		return fmt.Errorf("%s:%s", op, fmt.Errorf("%w;%w", err, con.Close()))
	}

	return con.Close()
}
