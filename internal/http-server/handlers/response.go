package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"
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
