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

func newErrorWsResponse(con *websocket.Conn, statusCode int, message string) error {
	const op = "response.newErrorWsResponse"

	closeErr := websocket.FormatCloseMessage(statusCode, message)

	if err := con.WriteMessage(websocket.CloseMessage, closeErr); err != nil {
		return fmt.Errorf("%s:%w", op, fmt.Errorf("%w;%v", err, con.Close()))
	}

	return con.Close()
}
